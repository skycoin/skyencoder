package skyencoder

import (
	"errors"
	"fmt"
	"go/build"
	"go/types"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/structtag"
	"github.com/golang/tools/go/loader"
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
	// so that a package that doesn't compile can still have a struct declaration extracted
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

// StructInfo has metadata for a struct type loaded from source
type StructInfo struct {
	Name    string
	Struct  *types.Struct
	Package *types.Package
}

// FindStructInfoInProgram finds a matching struct by name from a `*loader.Program`.
func FindStructInfoInProgram(p *loader.Program, name string) (*StructInfo, error) {
	// For programs loaded by file, the package will be in p.Created. Look here first
	for _, pk := range p.Created {
		s, err := findStructTypeInPackage(pk, name)
		if err != nil {
			return nil, err
		}
		if s != nil {
			return &StructInfo{
				Name:    name,
				Struct:  s,
				Package: pk.Pkg,
			}, nil
		}
	}

	// For programs loaded by import path, the package will be in imported
	for _, pk := range p.Imported {
		s, err := findStructTypeInPackage(pk, name)
		if err != nil {
			return nil, err
		}
		if s != nil {
			return &StructInfo{
				Name:    name,
				Struct:  s,
				Package: pk.Pkg,
			}, nil
		}
	}

	return nil, nil
}

func findStructTypeInPackage(p *loader.PackageInfo, name string) (*types.Struct, error) {
loop:
	for _, v := range p.Defs {
		if v == nil {
			continue
		}

		t := v.Type()
		switch x := t.(type) {
		case *types.Named:
			obj := x.Obj()
			if obj.Name() != name {
				continue loop
			}
			st := x.Underlying()
			switch y := st.(type) {
			case *types.Struct:
				return y, nil
			default:
				return nil, fmt.Errorf("Found type with name %s but underlying type is %T, not struct", name, y)
			}
		}
	}

	return nil, nil
}

// BuildStructEncoder builds formatted source code for encoding/decoding a struct.
// If `destPackage` is empty, assumes the generated code will be in the same package as the struct.
// Otherwise, the generated code will have this package in the package name declaration, and reference the struct as an external type.
// `fmtFilename` is a somewhat arbitrary reference filename; when formatting the code with imports, the generated code is treated as
// being from this filename for the purpose of resolving the necessary import paths.
// If not using `destPackage`, `fmtFilename` should be an arbitrary filename in the same path as the file which contains the struct.
// If using `destPackage`, `fmtFilename` should be an arbitrary filename in the path where the file is to be saved.
func BuildStructEncoder(s *StructInfo, destPackage, fmtFilename string) ([]byte, error) {
	debugPrintln("Package path:", s.Package.Path())
	encodeSizeSrc, err := buildEncodeSize(s.Name, s.Struct)
	if err != nil {
		return nil, fmt.Errorf("buildEncodeSize failed: %v", err)
	}

	encodeSrc, err := buildEncode(s.Name, s.Struct)
	if err != nil {
		return nil, fmt.Errorf("buildEncode failed: %v", err)
	}

	// Use the struct's package for localizing type names to the package where application,
	// unless destPackage is specified, then treat all type names as non-local
	internalPackage := s.Package
	if destPackage != "" {
		internalPackage = nil
	}

	decodeSrc, err := buildDecode(s.Name, s.Struct, internalPackage)
	if err != nil {
		return nil, fmt.Errorf("buildDecode failed: %v", err)
	}

	src := append(encodeSizeSrc, append(encodeSrc, decodeSrc...)...)

	pkgName := destPackage
	if pkgName == "" {
		pkgName = s.Package.Name()
	}

	pkgHeader := fmt.Sprintf("package %s\n\n", pkgName)
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

func buildEncodeSize(name string, s *types.Struct) ([]byte, error) {
	section, _, err := buildCodeSectionEncodeSize(s, "obj", "i", 0, nil)
	if err != nil {
		return nil, err
	}

	return WrapEncodeSizeFunc(name, "i0", section), nil
}

func buildEncode(name string, s *types.Struct) ([]byte, error) {
	section, err := buildCodeSectionEncode(s, "obj", false, nil)
	if err != nil {
		return nil, err
	}

	return WrapEncodeFunc(name, section), nil
}

func buildDecode(name string, s *types.Struct, p *types.Package) ([]byte, error) {
	section, err := buildCodeSectionDecode(s, p, "obj", false, "", 0, nil)
	if err != nil {
		return nil, err
	}

	return WrapDecodeFunc(name, section), nil
}

func buildCodeSectionEncode(t types.Type, varName string, castType bool, options *Options) (string, error) {
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
		return buildCodeSectionEncode(x.Underlying(), varName, true, options)

	case *types.Basic:
		switch x.Kind() {
		case types.Bool:
			return BuildEncodeBool(varName, castType, options), nil
		case types.Int8:
			return BuildEncodeInt8(varName, castType, options), nil
		case types.Int16:
			return BuildEncodeInt16(varName, castType, options), nil
		case types.Int32:
			return BuildEncodeInt32(varName, castType, options), nil
		case types.Int64:
			return BuildEncodeInt64(varName, castType, options), nil
		case types.Uint8:
			return BuildEncodeUint8(varName, castType, options), nil
		case types.Uint16:
			return BuildEncodeUint16(varName, castType, options), nil
		case types.Uint32:
			return BuildEncodeUint32(varName, castType, options), nil
		case types.Uint64:
			return BuildEncodeUint64(varName, castType, options), nil
		case types.String:
			return BuildEncodeString(varName, options), nil
		default:
			return "", fmt.Errorf("Unhandled *types.Basic type %s for var %s", x.Name(), varName)
		}

	case *types.Array:
		elem := x.Elem()

		if isByte(elem) {
			return BuildEncodeByteArray(varName, options), nil
		}

		elemSection, err := buildCodeSectionEncode(elem, "x", false, nil)
		if err != nil {
			return "", err
		}

		return BuildEncodeArray(varName, "x", elemSection, options), nil

	case *types.Slice:
		elem := x.Elem()

		if isByte(elem) {
			return BuildEncodeByteSlice(varName, options), nil
		}

		elemSection, err := buildCodeSectionEncode(elem, "x", false, nil)
		if err != nil {
			return "", err
		}

		return BuildEncodeSlice(varName, "x", elemSection, options), nil

	case *types.Map:
		keySection, err := buildCodeSectionEncode(x.Key(), "k", false, nil)
		if err != nil {
			return "", err
		}

		elemSection, err := buildCodeSectionEncode(x.Elem(), "v", false, nil)
		if err != nil {
			return "", err
		}

		return BuildEncodeMap(varName, "k", "v", keySection, elemSection, options), nil

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
			if options != nil && options.OmitEmpty && i != x.NumFields()-1 {
				return "", errors.New("omitempty option can only be used on the last field in a struct")
			}

			nextVarName := fmt.Sprintf("%s.%s", varName, f.Name())
			section, err := buildCodeSectionEncode(f.Type(), nextVarName, false, options)
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
			return BuildEncodeSizeBool(varName, counterName, options), false, nil
		case types.Int8:
			return BuildEncodeSizeInt8(varName, counterName, options), false, nil
		case types.Int16:
			return BuildEncodeSizeInt16(varName, counterName, options), false, nil
		case types.Int32:
			return BuildEncodeSizeInt32(varName, counterName, options), false, nil
		case types.Int64:
			return BuildEncodeSizeInt64(varName, counterName, options), false, nil
		case types.Uint8:
			return BuildEncodeSizeUint8(varName, counterName, options), false, nil
		case types.Uint16:
			return BuildEncodeSizeUint16(varName, counterName, options), false, nil
		case types.Uint32:
			return BuildEncodeSizeUint32(varName, counterName, options), false, nil
		case types.Uint64:
			return BuildEncodeSizeUint64(varName, counterName, options), false, nil
		case types.String:
			return BuildEncodeSizeString(varName, counterName, options), true, nil
		default:
			return "", false, fmt.Errorf("Unhandled *types.Basic type %q for var %q", x.Name(), varName)
		}

	case *types.Array:
		elem := x.Elem()

		if isByte(elem) {
			return BuildEncodeSizeByteArray(varName, counterName, x.Len(), options), false, nil
		}

		nextCounterName := fmt.Sprintf("%s%d", baseCounterName, depth+1)
		elemSection, isDynamic, err := buildCodeSectionEncodeSize(elem, "x", baseCounterName, depth+1, nil)
		if err != nil {
			return "", false, err
		}

		return BuildEncodeSizeArray(varName, counterName, nextCounterName, "x", elemSection, x.Len(), isDynamic, options), isDynamic, nil

	case *types.Slice:
		elem := x.Elem()

		if isByte(elem) {
			return BuildEncodeSizeByteSlice(varName, counterName, options), false, nil
		}

		nextCounterName := fmt.Sprintf("%s%d", baseCounterName, depth+1)
		elemSection, isDynamic, err := buildCodeSectionEncodeSize(elem, "x", baseCounterName, depth+1, nil)
		if err != nil {
			return "", false, err
		}

		return BuildEncodeSizeSlice(varName, counterName, nextCounterName, "x", elemSection, isDynamic, options), true, nil

	case *types.Map:
		nextCounterName := fmt.Sprintf("%s%d", baseCounterName, depth+1)

		keySection, isDynamicKey, err := buildCodeSectionEncodeSize(x.Key(), "k", baseCounterName, depth+1, nil)
		if err != nil {
			return "", false, err
		}

		elemSection, isDynamicElem, err := buildCodeSectionEncodeSize(x.Elem(), "v", baseCounterName, depth+1, nil)
		if err != nil {
			return "", false, err
		}

		return BuildEncodeSizeMap(varName, counterName, nextCounterName, "k", "v", keySection, elemSection, isDynamicKey, isDynamicElem, options), true, nil

	case *types.Struct:
		isDynamic := false
		sections := make([]string, x.NumFields())
		for i := 0; i < x.NumFields(); i++ {
			f := x.Field(i)

			// TODO -- confirm that the original encoder ignores unexported fields
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
			return BuildDecodeBool(varName, castType, typeName, options), nil
		case types.Int8:
			return BuildDecodeInt8(varName, castType, typeName, options), nil
		case types.Int16:
			return BuildDecodeInt16(varName, castType, typeName, options), nil
		case types.Int32:
			return BuildDecodeInt32(varName, castType, typeName, options), nil
		case types.Int64:
			return BuildDecodeInt64(varName, castType, typeName, options), nil
		case types.Uint8:
			return BuildDecodeUint8(varName, castType, typeName, options), nil
		case types.Uint16:
			return BuildDecodeUint16(varName, castType, typeName, options), nil
		case types.Uint32:
			return BuildDecodeUint32(varName, castType, typeName, options), nil
		case types.Uint64:
			return BuildDecodeUint64(varName, castType, typeName, options), nil
		case types.String:
			return BuildDecodeString(varName, options), nil
		default:
			return "", fmt.Errorf("Unhandled *types.Basic type %s for var %s", x.Name(), varName)
		}

	case *types.Array:
		elem := x.Elem()

		if isByte(elem) {
			return BuildDecodeByteArray(varName, options), nil
		}

		elemCounterName := fmt.Sprintf("zz%d", depth)
		elemVarName := fmt.Sprintf("%s[%s]", varName, elemCounterName)
		elemSection, err := buildCodeSectionDecode(elem, p, elemVarName, false, "", depth+1, nil)
		if err != nil {
			return "", err
		}

		return BuildDecodeArray(varName, elemCounterName, elemVarName, elemSection, options), nil

	case *types.Slice:
		elem := x.Elem()

		if isByte(elem) {
			return BuildDecodeByteSlice(varName, options), nil
		}

		elemCounterName := fmt.Sprintf("zz%d", depth)
		elemVarName := fmt.Sprintf("%s[%s]", varName, elemCounterName)
		elemSection, err := buildCodeSectionDecode(elem, p, elemVarName, false, "", depth+1, nil)
		if err != nil {
			return "", err
		}

		return BuildDecodeSlice(varName, elemCounterName, elemVarName, elemSection, sliceTypeName(x, p), options), nil

	case *types.Map:
		keyVarName := fmt.Sprintf("k%d", depth)
		keySection, err := buildCodeSectionDecode(x.Key(), p, keyVarName, false, "", depth+1, nil)
		if err != nil {
			return "", err
		}

		elemVarName := fmt.Sprintf("v%d", depth)
		elemSection, err := buildCodeSectionDecode(x.Elem(), p, elemVarName, false, "", depth+1, nil)
		if err != nil {
			return "", err
		}

		return BuildDecodeMap(varName, keyVarName, elemVarName, keySection, elemSection, mapTypeName(x, p), options), nil

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
