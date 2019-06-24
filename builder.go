package skyencoder

import (
	"errors"
	"fmt"
	"go/build"
	"go/types"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/structtag"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/imports"
)

const debug = false

func debugPrintln(args ...interface{}) {
	if debug {
		fmt.Println(args...)
	}
}

func debugPrintf(msg string, args ...interface{}) {
	if debug {
		fmt.Printf(msg, args...)
	}
}

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

// ToSnakeCase converts camel case to snake case
func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// FindDiskPathOfImport maps an import path (e.g. "github.com/skycoin/skycoin/src/coin") to a path on disk,
// searching GOPATH for the first matching directory
// TODO -- this might not work with go modules
func FindDiskPathOfImport(importPath string) (string, error) {
	gopath := os.Getenv("GOPATH")
	pts := strings.Split(gopath, ":")
	for _, pt := range pts {
		if pt == "" {
			continue
		}

		fullPath := filepath.Join(filepath.Join(pt, "src/"), importPath)

		stat, err := os.Stat(fullPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return "", err
			}
		} else if stat.IsDir() {
			return fullPath, nil
		}
	}

	return "", nil
}

// LoadProgram loads a program from args (which is a package or a set of files in a package) and build tags
func LoadProgram(args, buildTags []string) (*loader.Program, error) {
	buildContext := build.Default
	buildContext.BuildTags = append(buildContext.BuildTags, buildTags...)

	// Load the package with the least restrictive parsing and type checking,
	// so that a package that doesn't compile can still have a type declaration extracted
	cfg := loader.Config{
		Build:      &buildContext,
		ParserMode: 0,
		TypeChecker: types.Config{
			IgnoreFuncBodies:         true, // ignore functions
			FakeImportC:              true, // ignore import "C"
			DisableUnusedImportCheck: true, // ignore unused imports
		},
		TypeCheckFuncBodies: func(path string) bool {
			return false // ignore functions
		},
		AllowErrors: true,
	}

	loadTests := true
	unused, err := cfg.FromArgs(args, loadTests)
	if err != nil {
		return nil, fmt.Errorf("loader.Config.FromArgs failed: %v", err)
	}

	if len(unused) != 0 {
		return nil, fmt.Errorf("Not all args consumed by loader.Config.FromArgs. Remaining args: %v", unused)
	}

	program, err := cfg.Load()
	if err != nil {
		return nil, fmt.Errorf("loader.Config.Load: %v", err)
	}

	return program, nil
}

// StructInfo has metadata for a type loaded from source
type StructInfo struct {
	Name     string
	Type     *types.Struct
	Package  *types.Package
	Exported bool
}

// FindStructInfoInProgram finds a matching type by name from a `*loader.Program`.
func FindStructInfoInProgram(p *loader.Program, name string) (*StructInfo, error) {
	// For programs loaded by file, the package will be in p.Created. Look here first
	for _, pk := range p.Created {
		s, exported, err := findStructInPackage(pk, name)
		if err != nil {
			return nil, err
		}
		if s != nil {
			return &StructInfo{
				Name:     name,
				Type:     s,
				Package:  pk.Pkg,
				Exported: exported,
			}, nil
		}
	}

	// For programs loaded by import path, the package will be in imported
	for _, pk := range p.Imported {
		s, exported, err := findStructInPackage(pk, name)
		if err != nil {
			return nil, err
		}
		if s != nil {
			return &StructInfo{
				Name:     name,
				Type:     s,
				Package:  pk.Pkg,
				Exported: exported,
			}, nil
		}
	}

	return nil, nil
}

func findStructInPackage(p *loader.PackageInfo, name string) (*types.Struct, bool, error) {
	obj := p.Pkg.Scope().Lookup(name)
	if obj == nil {
		return nil, false, nil
	}

	t := obj.Type()
	switch x := t.(type) {
	case *types.Named:
		t = x.Underlying()
		switch y := t.(type) {
		case *types.Struct:
			return y, x.Obj().Exported(), nil
		default:
			return nil, false, fmt.Errorf("Found type with name %s but underlying type is %T, not struct", name, y)
		}
	case *types.Struct:
		return x, false, nil
	default:
		return nil, false, fmt.Errorf("Found type with name %s but underlying type is %T, not struct", name, x)
	}
}

// BuildStructEncoder builds formatted source code for encoding/decoding a type.
// If `destPackage` is empty, assumes the generated code will be in the same package as the type.
// Otherwise, the generated code will have this package in the package name declaration, and reference the type as an external type.
// `fmtFilename` is a somewhat arbitrary reference filename; when formatting the code with imports, the generated code is treated as
// being from this filename for the purpose of resolving the necessary import paths.
// If not using `destPackage`, `fmtFilename` should be an arbitrary filename in the same path as the file which contains the type.
// If using `destPackage`, `fmtFilename` should be an arbitrary filename in the path where the file is to be saved.
func BuildStructEncoder(s *StructInfo, destPackage, fmtFilename string, exported bool) ([]byte, error) {
	debugPrintln("Package path:", s.Package.Path())
	encodeSizeSrc, err := buildEncodeSize(s, destPackage != "", exported)
	if err != nil {
		return nil, fmt.Errorf("buildEncodeSize failed: %v", err)
	}

	encodeSrc, err := buildEncode(s, destPackage != "", exported)
	if err != nil {
		return nil, fmt.Errorf("buildEncode failed: %v", err)
	}

	// Use the type's package for localizing type names to the package,
	// unless destPackage is specified, then treat all type names as non-local
	internalPackage := s.Package
	if destPackage != "" {
		internalPackage = nil
	}

	decodeSrc, err := buildDecode(s, internalPackage, destPackage != "", exported)
	if err != nil {
		return nil, fmt.Errorf("buildDecode failed: %v", err)
	}

	src := append(encodeSizeSrc, append(encodeSrc, decodeSrc...)...)

	pkgName := destPackage
	if pkgName == "" {
		pkgName = s.Package.Name()
	}

	pkgHeader := fmt.Sprintf("// Code generated by github.com/skycoin/skyencoder. DO NOT EDIT.\n\npackage %s\n\n", pkgName)
	src = append([]byte(pkgHeader), src...)

	// Format with imports
	fmtSrc, err := imports.Process(fmtFilename, src, &imports.Options{
		Fragment:  false,
		Comments:  true,
		TabIndent: true,
		TabWidth:  8,
	})
	if err != nil {
		debugPrintln(string(src))
		return nil, fmt.Errorf("imports.Process failed: %v", err)
	}

	return fmtSrc, nil
}

// BuildStructEncoderTest builds the _test.go file that tests the code generated by BuildStructEncoder
func BuildStructEncoderTest(s *StructInfo, destPackage, fmtFilename string, exported bool) ([]byte, error) {
	pkgName := ""
	if destPackage != "" {
		pkgName = s.Package.Name()
	} else {
		destPackage = s.Package.Name()
	}

	hm, err := hasMap(s.Type)
	if err != nil {
		return nil, err
	}

	src := buildTest(s.Name, pkgName, destPackage, hm, exported)

	// Format with imports
	fmtSrc, err := imports.Process(fmtFilename, []byte(src), &imports.Options{
		Fragment:  false,
		Comments:  true,
		TabIndent: true,
		TabWidth:  8,
	})
	if err != nil {
		debugPrintln(string(src))
		return nil, fmt.Errorf("imports.Process failed: %v", err)
	}

	return fmtSrc, nil
}

func buildEncodeSize(s *StructInfo, externalPackage, exported bool) ([]byte, error) {
	section, _, err := buildCodeSectionEncodeSize(s.Type, "obj", "i", 0, nil)
	if err != nil {
		return nil, err
	}

	pkgName := ""
	if externalPackage {
		pkgName = s.Package.Name()
	}

	return wrapEncodeSizeFunc(s.Name, pkgName, "i0", section, exported), nil
}

func buildEncode(s *StructInfo, externalPackage, exported bool) ([]byte, error) {
	section, err := buildCodeSectionEncode(s.Type, "obj", true, true, nil)
	if err != nil {
		return nil, err
	}

	pkgName := ""
	if externalPackage {
		pkgName = s.Package.Name()
	}

	return wrapEncodeFunc(s.Name, pkgName, section, exported), nil
}

func buildDecode(s *StructInfo, p *types.Package, externalPackage, exported bool) ([]byte, error) {
	section, err := buildCodeSectionDecode(s.Type, p, "obj", true, s.Name, 0, nil)
	if err != nil {
		return nil, err
	}

	pkgName := ""
	if externalPackage {
		pkgName = s.Package.Name()
	}

	return wrapDecodeFunc(s.Name, pkgName, section, exported), nil
}

func buildCodeSectionEncode(t types.Type, varName string, castType, isTopLevel bool, options *Options) (string, error) {
	// castType applies to basic int types; if true, an additional cast will be made in the generated code.
	// This is to convert types like "type Foo int8" back to int8

	debugPrintf("buildCodeSectionEncode type=%T varName=%s castType=%v options=%+v\n", t, varName, castType, options)

	if options != nil {
		if options.OmitEmpty && !omitEmptyIsValid(t) {
			return "", errors.New("omitempty is only valid for array, slice, map and string")
		}
	}

	switch x := t.(type) {
	case *types.Named:
		return buildCodeSectionEncode(x.Underlying(), varName, true, false, options)

	case *types.Basic:
		switch x.Kind() {
		case types.Bool:
			return buildEncodeBool(varName, castType, options), nil
		case types.Int8:
			return buildEncodeInt8(varName, castType, options), nil
		case types.Int16:
			return buildEncodeInt16(varName, castType, options), nil
		case types.Int32:
			return buildEncodeInt32(varName, castType, options), nil
		case types.Int64:
			return buildEncodeInt64(varName, castType, options), nil
		case types.Uint8:
			return buildEncodeUint8(varName, castType, options), nil
		case types.Uint16:
			return buildEncodeUint16(varName, castType, options), nil
		case types.Uint32:
			return buildEncodeUint32(varName, castType, options), nil
		case types.Uint64:
			return buildEncodeUint64(varName, castType, options), nil
		case types.Float32:
			return buildEncodeFloat32(varName, castType, options), nil
		case types.Float64:
			return buildEncodeFloat64(varName, castType, options), nil
		case types.String:
			return buildEncodeString(varName, options), nil
		default:
			return "", fmt.Errorf("Unhandled *types.Basic type %s for var %s", x.Name(), varName)
		}

	case *types.Array:
		elem := x.Elem()

		if isByte(elem) {
			return buildEncodeByteArray(varName, options), nil
		}

		elemSection, err := buildCodeSectionEncode(elem, "x", false, false, nil)
		if err != nil {
			return "", err
		}

		return buildEncodeArray(varName, "x", elemSection, options), nil

	case *types.Slice:
		elem := x.Elem()

		if empty, err := isEmptyStruct(elem); err != nil {
			return "", err
		} else if empty {
			return "", fmt.Errorf("A slice of an empty encoded struct is not allowed (var=%q)", varName)
		}

		if isByte(elem) {
			return buildEncodeByteSlice(varName, options), nil
		}

		elemSection, err := buildCodeSectionEncode(elem, "x", false, false, nil)
		if err != nil {
			return "", err
		}

		return buildEncodeSlice(varName, "x", elemSection, options), nil

	case *types.Map:
		keySection, err := buildCodeSectionEncode(x.Key(), "k", false, false, nil)
		if err != nil {
			return "", err
		}

		elemSection, err := buildCodeSectionEncode(x.Elem(), "v", false, false, nil)
		if err != nil {
			return "", err
		}

		return buildEncodeMap(varName, "k", "v", keySection, elemSection, options), nil

	case *types.Struct:
		sections := make([]string, x.NumFields())
		for i := 0; i < x.NumFields(); i++ {
			f := x.Field(i)

			if !f.Exported() {
				continue
			}

			ignore, options, err := parseTag(x.Tag(i))
			if err != nil {
				return "", err
			}

			if ignore {
				continue
			}

			// NOTES ON OMITEMPTY
			// - Must be last field in struct
			// - Only applies to arrays, slices, maps and string
			if options != nil && options.OmitEmpty {
				if i != x.NumFields()-1 {
					return "", errors.New("omitempty option can only be used on the last field in a struct")
				}
				if !isTopLevel {
					return "", errors.New("omitempty option can only be used on a top-level struct")
				}
			}

			nextVarName := fmt.Sprintf("%s.%s", varName, f.Name())
			section, err := buildCodeSectionEncode(f.Type(), nextVarName, false, false, options)
			if err != nil {
				return "", err
			}

			sections[i] = section
		}

		return strings.Join(sections, "\n\n"), nil

	default:
		return "", fmt.Errorf("Unhandled type %T for var %s", x, varName)
	}
}

// buildCodeSectionEncodeSize returns the code section and whether or not the section has dynamic sizing (requiring a runtime len() check)
func buildCodeSectionEncodeSize(t types.Type, varName, baseCounterName string, depth int, options *Options) (string, bool, error) {
	debugPrintf("buildCodeSectionEncodeSize type=%T varName=%s baseCounterName=%s depth=%d options=%+v\n", t, varName, baseCounterName, depth, options)

	if options != nil {
		if options.OmitEmpty && !omitEmptyIsValid(t) {
			return "", false, errors.New("omitempty is only valid for array, slice, map and string")
		}
	}

	counterName := fmt.Sprintf("%s%d", baseCounterName, depth)

	switch x := t.(type) {
	case *types.Named:
		return buildCodeSectionEncodeSize(x.Underlying(), varName, baseCounterName, depth, options)

	case *types.Basic:
		switch x.Kind() {
		case types.Bool:
			return buildEncodeSizeBool(varName, counterName, options), false, nil
		case types.Int8:
			return buildEncodeSizeInt8(varName, counterName, options), false, nil
		case types.Int16:
			return buildEncodeSizeInt16(varName, counterName, options), false, nil
		case types.Int32:
			return buildEncodeSizeInt32(varName, counterName, options), false, nil
		case types.Int64:
			return buildEncodeSizeInt64(varName, counterName, options), false, nil
		case types.Uint8:
			return buildEncodeSizeUint8(varName, counterName, options), false, nil
		case types.Uint16:
			return buildEncodeSizeUint16(varName, counterName, options), false, nil
		case types.Uint32:
			return buildEncodeSizeUint32(varName, counterName, options), false, nil
		case types.Uint64:
			return buildEncodeSizeUint64(varName, counterName, options), false, nil
		case types.Float32:
			return buildEncodeSizeFloat32(varName, counterName, options), false, nil
		case types.Float64:
			return buildEncodeSizeFloat64(varName, counterName, options), false, nil
		case types.String:
			return buildEncodeSizeString(varName, counterName, options), true, nil
		default:
			return "", false, fmt.Errorf("Unhandled *types.Basic type %q for var %q", x.Name(), varName)
		}

	case *types.Array:
		elem := x.Elem()

		if isByte(elem) {
			return buildEncodeSizeByteArray(varName, counterName, x.Len(), options), false, nil
		}

		nextCounterName := fmt.Sprintf("%s%d", baseCounterName, depth+1)
		xVarName := fmt.Sprintf("x%d", depth+1)
		elemSection, isDynamic, err := buildCodeSectionEncodeSize(elem, xVarName, baseCounterName, depth+1, nil)
		if err != nil {
			return "", false, err
		}

		return buildEncodeSizeArray(varName, counterName, nextCounterName, xVarName, elemSection, x.Len(), isDynamic, options), isDynamic, nil

	case *types.Slice:
		elem := x.Elem()

		if empty, err := isEmptyStruct(elem); err != nil {
			return "", false, err
		} else if empty {
			return "", false, fmt.Errorf("A slice of an empty encoded struct is not allowed (var=%q)", varName)
		}

		if isByte(elem) {
			return buildEncodeSizeByteSlice(varName, counterName, options), true, nil
		}

		nextCounterName := fmt.Sprintf("%s%d", baseCounterName, depth+1)
		xVarName := fmt.Sprintf("x%d", depth+1)
		elemSection, isDynamic, err := buildCodeSectionEncodeSize(elem, xVarName, baseCounterName, depth+1, nil)
		if err != nil {
			return "", false, err
		}

		return buildEncodeSizeSlice(varName, counterName, nextCounterName, xVarName, elemSection, isDynamic, options), true, nil

	case *types.Map:
		nextCounterName := fmt.Sprintf("%s%d", baseCounterName, depth+1)
		kVarName := fmt.Sprintf("k%d", depth+1)
		vVarName := fmt.Sprintf("v%d", depth+1)

		keySection, isDynamicKey, err := buildCodeSectionEncodeSize(x.Key(), kVarName, baseCounterName, depth+1, nil)
		if err != nil {
			return "", false, err
		}

		elemSection, isDynamicElem, err := buildCodeSectionEncodeSize(x.Elem(), vVarName, baseCounterName, depth+1, nil)
		if err != nil {
			return "", false, err
		}

		return buildEncodeSizeMap(varName, counterName, nextCounterName, kVarName, vVarName, keySection, elemSection, isDynamicKey, isDynamicElem, options), true, nil

	case *types.Struct:
		isDynamic := false
		sections := make([]string, x.NumFields())
		for i := 0; i < x.NumFields(); i++ {
			f := x.Field(i)

			if !f.Exported() {
				continue
			}

			ignore, options, err := parseTag(x.Tag(i))
			if err != nil {
				return "", false, err
			}

			if ignore {
				continue
			}

			// NOTES ON OMITEMPTY
			// - Must be last field in struct
			// - Only applies to arrays, slices, maps and string
			if options != nil && options.OmitEmpty && i != x.NumFields()-1 {
				return "", false, errors.New("omitempty option can only be used on the last field in a struct")
			}

			nextVarName := fmt.Sprintf("%s.%s", varName, f.Name())
			section, sectionIsDynamic, err := buildCodeSectionEncodeSize(f.Type(), nextVarName, baseCounterName, depth, options)
			if err != nil {
				return "", false, err
			}

			if sectionIsDynamic {
				isDynamic = true
			}

			sections[i] = section
		}

		return strings.Join(sections, "\n\n"), isDynamic, nil

	default:
		return "", false, fmt.Errorf("Unhandled type %T for var %s", x, varName)
	}
}

func buildCodeSectionDecode(t types.Type, p *types.Package, varName string, castType bool, typeName string, depth int, options *Options) (string, error) {
	// castType applies to basic int types; if true, an additional cast will be made in the generated code.
	// This is to convert types like "type Foo int8" back to int8

	pkgName := ""
	if p != nil {
		pkgName = p.String()
	}
	debugPrintf("buildCodeSectionDecode type=%T package=%s varName=%s castType=%v typeName=%s depth=%d options=%+v\n", t, pkgName, varName, castType, typeName, depth, options)

	if options != nil {
		if options.MaxLength != 0 && !maxLenIsValid(t) {
			return "", errors.New("maxlen is only valid for slice, string and map")
		}
	}

	switch x := t.(type) {
	case *types.Named:
		return buildCodeSectionDecode(x.Underlying(), p, varName, true, typeNameOf(x, p), depth, options)

	case *types.Basic:
		if typeName == "" {
			typeName = typeNameOf(x, p)
		}

		debugPrintf("types.Basic type name is %s\n", typeName)

		switch x.Kind() {
		case types.Bool:
			return buildDecodeBool(varName, castType, typeName, options), nil
		case types.Int8:
			return buildDecodeInt8(varName, castType, typeName, options), nil
		case types.Int16:
			return buildDecodeInt16(varName, castType, typeName, options), nil
		case types.Int32:
			return buildDecodeInt32(varName, castType, typeName, options), nil
		case types.Int64:
			return buildDecodeInt64(varName, castType, typeName, options), nil
		case types.Uint8:
			return buildDecodeUint8(varName, castType, typeName, options), nil
		case types.Uint16:
			return buildDecodeUint16(varName, castType, typeName, options), nil
		case types.Uint32:
			return buildDecodeUint32(varName, castType, typeName, options), nil
		case types.Uint64:
			return buildDecodeUint64(varName, castType, typeName, options), nil
		case types.Float32:
			return buildDecodeFloat32(varName, castType, typeName, options), nil
		case types.Float64:
			return buildDecodeFloat64(varName, castType, typeName, options), nil
		case types.String:
			return buildDecodeString(varName, options), nil
		default:
			return "", fmt.Errorf("Unhandled *types.Basic type %s for var %s", x.Name(), varName)
		}

	case *types.Array:
		elem := x.Elem()

		if isByte(elem) {
			return buildDecodeByteArray(varName, options), nil
		}

		elemCounterName := fmt.Sprintf("z%d", depth)
		elemVarName := fmt.Sprintf("%s[%s]", varName, elemCounterName)
		elemSection, err := buildCodeSectionDecode(elem, p, elemVarName, false, "", depth+1, nil)
		if err != nil {
			return "", err
		}

		return buildDecodeArray(varName, elemCounterName, elemVarName, elemSection, options), nil

	case *types.Slice:
		elem := x.Elem()

		if empty, err := isEmptyStruct(elem); err != nil {
			return "", err
		} else if empty {
			return "", fmt.Errorf("A slice of an empty encoded struct is not allowed (var=%q)", varName)
		}

		if isByte(elem) {
			return buildDecodeByteSlice(varName, options), nil
		}

		elemCounterName := fmt.Sprintf("z%d", depth)
		elemVarName := fmt.Sprintf("%s[%s]", varName, elemCounterName)
		elemSection, err := buildCodeSectionDecode(elem, p, elemVarName, false, "", depth+1, nil)
		if err != nil {
			return "", err
		}

		return buildDecodeSlice(varName, elemCounterName, elemVarName, elemSection, sliceTypeName(x, p), options), nil

	case *types.Map:
		keyVarName := fmt.Sprintf("k%d", depth)
		keySection, err := buildCodeSectionDecode(x.Key(), p, keyVarName, false, "", depth+1, nil)
		if err != nil {
			return "", err
		}
		keyType := typeNameOf(x.Key(), p)

		elemVarName := fmt.Sprintf("v%d", depth)
		elemSection, err := buildCodeSectionDecode(x.Elem(), p, elemVarName, false, "", depth+1, nil)
		if err != nil {
			return "", err
		}
		elemType := typeNameOf(x.Elem(), p)

		return buildDecodeMap(varName, keyVarName, elemVarName, keyType, elemType, keySection, elemSection, mapTypeName(x, p), options), nil

	case *types.Struct:
		sections := make([]string, x.NumFields())
		for i := 0; i < x.NumFields(); i++ {
			f := x.Field(i)

			if !f.Exported() {
				continue
			}

			ignore, options, err := parseTag(x.Tag(i))
			if err != nil {
				return "", err
			}

			if ignore {
				continue
			}

			nextVarName := fmt.Sprintf("%s.%s", varName, f.Name())
			section, err := buildCodeSectionDecode(f.Type(), p, nextVarName, false, "", depth+1, options)
			if err != nil {
				return "", err
			}

			sections[i] = section
		}

		return strings.Join(sections, "\n\n"), nil

	default:
		return "", fmt.Errorf("Unhandled type %T for var %s", x, varName)
	}
}

func isEmptyStruct(t types.Type) (bool, error) {
	switch x := t.(type) {
	case *types.Named:
		return isEmptyStruct(x.Underlying())
	case *types.Struct:
		hasEncodable, err := structHasEncodableFields(x)
		if err != nil {
			return false, err
		}
		return !hasEncodable, nil
	default:
		return false, nil
	}
}

// structHasEncodableFields returns true if a struct has an exported field that is not disabled from encoding by a struct tag
func structHasEncodableFields(t *types.Struct) (bool, error) {
	has := false
	for i := 0; i < t.NumFields(); i++ {
		f := t.Field(i)
		if !f.Exported() {
			continue
		}

		ignore, _, err := parseTag(t.Tag(i))
		if err != nil {
			return false, err
		}

		if !ignore {
			has = true
		}
	}
	return has, nil
}

func sliceTypeName(t *types.Slice, p *types.Package) string {
	elemType := typeNameOf(t.Elem(), p)
	debugPrintf("sliceTypeName: elemType is %s\n", elemType)
	return fmt.Sprintf("[]%s", elemType)
}

func arrayTypeName(t *types.Array, p *types.Package) string {
	elemType := typeNameOf(t.Elem(), p)
	return fmt.Sprintf("[%d]%s", t.Len(), elemType)
}

func mapTypeName(t *types.Map, p *types.Package) string {
	// t.String() will return a type with fully qualified import paths, e.g.
	// map[int32]coin.UxOut will return "map[int32]github.com/skycoin/skycoin/src/coin.UxOut"
	// I can't find a way to get the type name without the import path, other than constructing it manually
	keyType := typeNameOf(t.Key(), p)
	elemType := typeNameOf(t.Elem(), p)
	return fmt.Sprintf("map[%s]%s", keyType, elemType)
}

func typeNameOf(t types.Type, p *types.Package) string {
	switch x := t.(type) {
	case *types.Named:
		obj := x.Obj()
		if p != nil && obj.Pkg().Path() == p.Path() {
			return obj.Name()
		}
		return fmt.Sprintf("%s.%s", obj.Pkg().Name(), obj.Name())
	case *types.Basic:
		return x.Name()
	case *types.Map:
		return mapTypeName(x, p)
	case *types.Slice:
		return sliceTypeName(x, p)
	case *types.Array:
		return arrayTypeName(x, p)
	case *types.Struct:
		return t.String()
	default:
		panic(fmt.Sprintf("typeNameOf unhandled type %T", x))
	}
}

func omitEmptyIsValid(t types.Type) bool {
	switch x := t.(type) {
	case *types.Named:
		return omitEmptyIsValid(x.Underlying())
	case *types.Basic:
		switch x.Kind() {
		case types.String:
			return true
		default:
			return false
		}
	case *types.Array, *types.Slice, *types.Map:
		return true
	default:
		return false
	}
}

func maxLenIsValid(t types.Type) bool {
	switch x := t.(type) {
	case *types.Named:
		return maxLenIsValid(x.Underlying())
	case *types.Basic:
		switch x.Kind() {
		case types.String:
			return true
		default:
			return false
		}
	case *types.Slice, *types.Map:
		return true
	default:
		return false
	}
}

func parseTag(tag string) (bool, *Options, error) {
	tags, err := structtag.Parse(tag)
	if err != nil {
		return false, nil, err
	}

	encTag, err := tags.Get("enc")
	if err != nil { // returns error when and only when tag not found
		return false, nil, nil
	}

	if encTag.Name != "" && encTag.Name != "-" {
		return false, nil, fmt.Errorf("Invalid struct tag name %q (must be empty or \"-\")", encTag.Name)
	}

	if encTag.Name == "-" {
		if len(encTag.Options) != 0 {
			return false, nil, fmt.Errorf("Invalid struct tag %q (is ignored with \"-\" but has options)", tag)
		}

		return true, nil, nil
	}

	opts := &Options{}
	for _, o := range encTag.Options {
		if o == "omitempty" {
			opts.OmitEmpty = true
		} else if strings.HasPrefix(o, "maxlen=") {
			numStr := o[len("maxlen="):]
			n, err := strconv.ParseUint(numStr, 10, 64)
			if err != nil {
				return false, nil, fmt.Errorf("Invalid maxlen option %q", o)
			}
			opts.MaxLength = n
		} else {
			return false, nil, fmt.Errorf("Invalid struct tag option %q", o)
		}
	}

	return false, opts, nil
}

func isByte(t types.Type) bool {
	switch x := t.(type) {
	case *types.Named:
		return isByte(x.Underlying())
	case *types.Basic:
		switch x.Kind() {
		case types.Uint8: // catches uint8, byte. int8 and bool, while only using 1 byte, cannot be used in a copy([]byte) call
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func hasMap(t types.Type) (bool, error) {
	switch x := t.(type) {
	case *types.Named:
		return hasMap(x.Underlying())

	case *types.Array:
		return hasMap(x.Elem())

	case *types.Slice:
		return hasMap(x.Elem())

	case *types.Map:
		return true, nil

	case *types.Struct:
		for i := 0; i < x.NumFields(); i++ {
			f := x.Field(i)

			if !f.Exported() {
				continue
			}

			ignore, _, err := parseTag(x.Tag(i))
			if err != nil {
				return false, err
			}

			if ignore {
				continue
			}

			has, err := hasMap(f.Type())
			if err != nil {
				return false, err
			}

			if has {
				return true, nil
			}
		}

		return false, nil

	default:
		return false, nil
	}
}
