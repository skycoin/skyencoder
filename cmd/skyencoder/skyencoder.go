package main

import (
	"errors"
	"flag"
	"fmt"
	"go/build"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/structtag"
	"github.com/golang/tools/go/loader"
	"github.com/skycoin/skyencoder"
	"golang.org/x/tools/imports"
)

/* TODO

TO DOCUMENT:

* If -package flag is used, considers the generated code as a different package from the one in which the struct is defined (even if the name is the same),
if you want to generate the code in the same package as the struct, do not specify -package.

* Encoder details such as anonymous fields

*/

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
	structNames = flag.String("struct", "", "comma-separated list of struct names; must be set")
	output      = flag.String("output", "", "output file name; default srcdir/<type>_string.go")
	buildTags   = flag.String("tags", "", "comma-separated list of build tags to apply")
	destPackage = flag.String("package", "", "package name for the output; if not provided, defaults to the struct's package")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of skyencoder2:\n")
	fmt.Fprintf(os.Stderr, "\tskyencoder2 [flags] -struct T [directory]\n")
	fmt.Fprintf(os.Stderr, "\tskyencoder2 [flags] -struct T files... # Must be a single package\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("skyencoder2: ")

	flag.Usage = usage
	flag.Parse()

	if len(*structNames) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	structNames := strings.Split(*structNames, ",")
	var tags []string
	if len(*buildTags) > 0 {
		tags = strings.Split(*buildTags, ",")
	}

	// We accept either one directory or a list of files. Which do we have?
	args := flag.Args()
	if len(args) == 0 {
		// Default: process whole package in current directory.
		args = []string{"."}
	}

	// Load the package with the least restrictive parsing and type checking,
	// so that a package that doesn't compile can still have a struct declaration extracted
	buildContext := build.Default
	buildContext.BuildTags = append(buildContext.BuildTags, tags...)

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
		log.Fatal("loader.Config.FromArgs:", err)
	}

	if len(unused) != 0 {
		log.Fatal("Not all args consumed by loader.Config.FromArgs. Remaining args:", unused)
	}

	program, err := cfg.Load()
	if err != nil {
		log.Fatal("loader.Config.Load:", err)
	}

	debugPrintln("args:", args)

	structs := make([]*structInfo, len(structNames))
	for i, name := range structNames {
		s, err := findStructInfoInProgram(program, name)
		if err != nil {
			log.Fatalf("Program did not contain valid struct for name %s: %v", name, err)
		}
		if s == nil {
			log.Fatal("Program does not contain type:", name)
		}

		structs[i] = s
	}

	// Determine if the arg is a directory or multiple files
	// If it is a directory, construct an artificial filename in that directory for goimports formatting,
	// otherwise use the first filename specified (they must all be in the same package)
	fmtFilename := args[0]
	stat, err := os.Stat(args[0])
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
		// argument is a import path e.g. "github.com/skycoin/skycoin/src/coin"
		fmtFilename = filepath.Join(structs[0].Package.Path(), "foo123123123123999.go")
	} else if stat.IsDir() {
		fmtFilename = filepath.Join(args[0], "foo123123123123999.go")
	}

	for i, s := range structs {
		debugPrintln("Package path:", s.Package.Path())
		encodeSizeSrc, err := buildEncodeSize(structNames[i], s.Struct)
		if err != nil {
			log.Fatal("buildEncodeSize failed:", err)
		}

		encodeSrc, err := buildEncode(structNames[i], s.Struct)
		if err != nil {
			log.Fatal("buildEncode failed:", err)
		}

		// Use the struct's package for localizing type names to the package where application,
		// unless destPackage is specified, then treat all type names as non-local
		internalPackage := s.Package
		if *destPackage != "" {
			internalPackage = nil
		}

		decodeSrc, err := buildDecode(structNames[i], s.Struct, internalPackage)
		if err != nil {
			log.Fatal("buildDecode failed:", err)
		}

		src := append(encodeSizeSrc, append(encodeSrc, decodeSrc...)...)

		pkgName := *destPackage
		if pkgName == "" {
			pkgName = s.Package.Name()
		}

		pkgHeader := fmt.Sprintf("package %s\n\n", pkgName)
		src = append([]byte(pkgHeader), src...)

		// Format with imports
		src, err = imports.Process(fmtFilename, src, &imports.Options{
			Fragment:  true,
			Comments:  true,
			TabIndent: true,
			TabWidth:  8,
		})
		if err != nil {
			log.Fatal("imports.Process failed:", err)
		}

		fmt.Println(string(src))
	}
}

func buildEncodeSize(name string, s *types.Struct) ([]byte, error) {
	section, _, err := buildCodeSectionEncodeSize(s, "obj", nil)
	if err != nil {
		return nil, err
	}

	return wrapEncodeSizeFunc(name, section), nil
}

func buildEncode(name string, s *types.Struct) ([]byte, error) {
	section, err := buildCodeSectionEncode(s, "obj", false, nil)
	if err != nil {
		return nil, err
	}

	return wrapEncodeFunc(name, section), nil
}

func buildDecode(name string, s *types.Struct, p *types.Package) ([]byte, error) {
	section, err := buildCodeSectionDecode(s, p, "obj", false, "", nil)
	if err != nil {
		return nil, err
	}

	return wrapDecodeFunc(name, section), nil
}

func wrapEncodeFunc(structName, funcBody string) []byte {
	return []byte(fmt.Sprintf(`
// Encode%[1]s encodes an object of type %[1]s to the buffer in encoder.Encoder
func Encode%[1]s(e *encoder.Encoder, obj *%[1]s) error {
	%[2]s

	return nil
}
`, structName, funcBody))
}

func wrapEncodeSizeFunc(structName, funcBody string) []byte {
	return []byte(fmt.Sprintf(`
// EncodeSize%[1]s computes the size of an encoded object of type %[1]s
func EncodeSize%[1]s(obj *%[1]s) int {
	i := 0

	%[2]s

	return i
}
`, structName, funcBody))
}

func wrapDecodeFunc(structName, funcBody string) []byte {
	return []byte(fmt.Sprintf(`
// Decode%[1]s decodes an object of type %[1]s from the buffer in encoder.Decoder
func Decode%[1]s(e *encoder.Decoder, obj *%[1]s) error {
	%[2]s

	return nil
}
`, structName, funcBody))
}

func isDir(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}

type structInfo struct {
	Struct  *types.Struct
	Package *types.Package
}

func findStructInfoInProgram(p *loader.Program, name string) (*structInfo, error) {
	// For programs loaded by file, the package will be in p.Created. Look here first
	for _, pk := range p.Created {
		s, err := findStructTypeInPackage(pk, name)
		if err != nil {
			return nil, err
		}
		if s != nil {
			return &structInfo{
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
			return &structInfo{
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

func buildCodeSectionEncode(t types.Type, varName string, castType bool, options *skyencoder.Options) (string, error) {
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
			return skyencoder.BuildEncodeBool(varName, castType, options), nil
		case types.Int8:
			return skyencoder.BuildEncodeInt8(varName, castType, options), nil
		case types.Int16:
			return skyencoder.BuildEncodeInt16(varName, castType, options), nil
		case types.Int32:
			return skyencoder.BuildEncodeInt32(varName, castType, options), nil
		case types.Int64:
			return skyencoder.BuildEncodeInt64(varName, castType, options), nil
		case types.Uint8:
			return skyencoder.BuildEncodeUint8(varName, castType, options), nil
		case types.Uint16:
			return skyencoder.BuildEncodeUint16(varName, castType, options), nil
		case types.Uint32:
			return skyencoder.BuildEncodeUint32(varName, castType, options), nil
		case types.Uint64:
			return skyencoder.BuildEncodeUint64(varName, castType, options), nil
		case types.String:
			return skyencoder.BuildEncodeString(varName, options), nil
		default:
			return "", fmt.Errorf("Unhandled *types.Basic type %s for var %s", x.Name(), varName)
		}

	case *types.Map:
		keySection, err := buildCodeSectionEncode(x.Key(), "k", false, nil)
		if err != nil {
			return "", err
		}

		elemSection, err := buildCodeSectionEncode(x.Elem(), "v", false, nil)
		if err != nil {
			return "", err
		}

		return skyencoder.BuildEncodeMap(varName, "k", "v", keySection, elemSection, options), nil

	case *types.Struct:
		sections := make([]string, x.NumFields())
		for i := 0; i < x.NumFields(); i++ {
			f := x.Field(i)

			// TODO -- confirm that the original encoder ignores unexported fields
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

	case *types.Array:
		elem := x.Elem()

		if isByte(elem) {
			return skyencoder.BuildEncodeByteArray(varName, options), nil
		}

		elemSection, err := buildCodeSectionEncode(elem, "x", false, nil)
		if err != nil {
			return "", err
		}

		return skyencoder.BuildEncodeArray(varName, "x", elemSection, options), nil

	case *types.Slice:
		elem := x.Elem()

		if isByte(elem) {
			return skyencoder.BuildEncodeByteSlice(varName, options), nil
		}

		elemSection, err := buildCodeSectionEncode(elem, "x", false, nil)
		if err != nil {
			return "", err
		}

		return skyencoder.BuildEncodeSlice(varName, "x", elemSection, options), nil

	default:
		return "", fmt.Errorf("Unhandled type %T for var %s", x, varName)
	}
}

func buildCodeSectionEncodeSize(t types.Type, varName string, options *skyencoder.Options) (string, bool, error) {
	// castType applies to basic int types; if true, an additional cast will be made in the generated code.
	// This is to convert types like "type Foo int8" back to int8

	debugPrintf("buildCodeSectionEncodeSize type=%T varName=%s options=%+v\n", t, varName, options)

	if options != nil {
		if options.OmitEmpty && !omitEmptyIsValid(t) {
			return "", false, errors.New("omitempty is only valid for array, slice, map and string")
		}
	}

	switch x := t.(type) {
	case *types.Named:
		return buildCodeSectionEncodeSize(x.Underlying(), varName, options)

	case *types.Basic:
		switch x.Kind() {
		case types.Bool:
			return skyencoder.BuildEncodeSizeBool(varName, options), false, nil
		case types.Int8:
			return skyencoder.BuildEncodeSizeInt8(varName, options), false, nil
		case types.Int16:
			return skyencoder.BuildEncodeSizeInt16(varName, options), false, nil
		case types.Int32:
			return skyencoder.BuildEncodeSizeInt32(varName, options), false, nil
		case types.Int64:
			return skyencoder.BuildEncodeSizeInt64(varName, options), false, nil
		case types.Uint8:
			return skyencoder.BuildEncodeSizeUint8(varName, options), false, nil
		case types.Uint16:
			return skyencoder.BuildEncodeSizeUint16(varName, options), false, nil
		case types.Uint32:
			return skyencoder.BuildEncodeSizeUint32(varName, options), false, nil
		case types.Uint64:
			return skyencoder.BuildEncodeSizeUint64(varName, options), false, nil
		case types.String:
			return skyencoder.BuildEncodeSizeString(varName, options), true, nil
		default:
			return "", false, fmt.Errorf("Unhandled *types.Basic type %q for var %q", x.Name(), varName)
		}

	case *types.Array:
		elem := x.Elem()

		if isByte(elem) {
			return skyencoder.BuildEncodeSizeByteArray(varName, x.Len(), options), false, nil
		}

		elemSection, isDynamic, err := buildCodeSectionEncodeSize(elem, "x", nil)
		if err != nil {
			return "", false, err
		}

		return skyencoder.BuildEncodeSizeArray(varName, "x", elemSection, x.Len(), isDynamic, options), isDynamic, nil

	case *types.Slice:
		elem := x.Elem()

		if isByte(elem) {
			return skyencoder.BuildEncodeSizeByteSlice(varName, options), false, nil
		}

		elemSection, isDynamic, err := buildCodeSectionEncodeSize(elem, "x", nil)
		if err != nil {
			return "", false, err
		}

		return skyencoder.BuildEncodeSizeSlice(varName, "x", elemSection, isDynamic, options), true, nil

	case *types.Map:
		keySection, isDynamicKey, err := buildCodeSectionEncodeSize(x.Key(), "k", nil)
		if err != nil {
			return "", false, err
		}

		elemSection, isDynamicElem, err := buildCodeSectionEncodeSize(x.Elem(), "v", nil)
		if err != nil {
			return "", false, err
		}

		return skyencoder.BuildEncodeSizeMap(varName, "k", "v", keySection, elemSection, isDynamicKey, isDynamicElem, options), true, nil

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
			section, sectionIsDynamic, err := buildCodeSectionEncodeSize(f.Type(), nextVarName, options)
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

func buildCodeSectionDecode(t types.Type, p *types.Package, varName string, castType bool, typeName string, options *skyencoder.Options) (string, error) {
	// castType applies to basic int types; if true, an additional cast will be made in the generated code.
	// This is to convert types like "type Foo int8" back to int8

	pkgName := ""
	if p != nil {
		pkgName = p.String()
	}
	debugPrintf("buildCodeSectionDecode type=%T package=%s varName=%s castType=%v options=%+v\n", t, pkgName, varName, castType, options)

	if options != nil {
		if options.MaxLength != 0 && !maxLenIsValid(t) {
			return "", errors.New("maxlen is only valid for slice, string and map")
		}
	}

	switch x := t.(type) {
	case *types.Named:
		// TODO -- the typeName x.String() is used to cast or allocate values.
		// x.String() includes the package name. This is correct if we are generating code to a different package than the type is declared in,
		// but incorrect if we want it in the same package.
		// Need to detect if we are generating code to the same package that the type is declared in,
		// and if so, then use x.Obj().Name() which will return the type name without package
		return buildCodeSectionDecode(x.Underlying(), p, varName, true, x.String(), options)

	case *types.Basic:
		if typeName == "" {
			typeName = x.Name()
		}

		debugPrintf("types.Basic type name is %s\n", typeName)

		switch x.Kind() {
		case types.Bool:
			return skyencoder.BuildDecodeBool(varName, castType, typeName, options), nil
		case types.Int8:
			return skyencoder.BuildDecodeInt8(varName, castType, typeName, options), nil
		case types.Int16:
			return skyencoder.BuildDecodeInt16(varName, castType, typeName, options), nil
		case types.Int32:
			return skyencoder.BuildDecodeInt32(varName, castType, typeName, options), nil
		case types.Int64:
			return skyencoder.BuildDecodeInt64(varName, castType, typeName, options), nil
		case types.Uint8:
			return skyencoder.BuildDecodeUint8(varName, castType, typeName, options), nil
		case types.Uint16:
			return skyencoder.BuildDecodeUint16(varName, castType, typeName, options), nil
		case types.Uint32:
			return skyencoder.BuildDecodeUint32(varName, castType, typeName, options), nil
		case types.Uint64:
			return skyencoder.BuildDecodeUint64(varName, castType, typeName, options), nil
		case types.String:
			return skyencoder.BuildDecodeString(varName, options), nil
		default:
			return "", fmt.Errorf("Unhandled *types.Basic type %s for var %s", x.Name(), varName)
		}

	case *types.Map:
		keySection, err := buildCodeSectionDecode(x.Key(), p, "k", false, "", nil)
		if err != nil {
			return "", err
		}

		elemSection, err := buildCodeSectionDecode(x.Elem(), p, "v", false, "", nil)
		if err != nil {
			return "", err
		}

		return skyencoder.BuildDecodeMap(varName, "k", "v", keySection, elemSection, mapTypeName(x, p), options), nil

	case *types.Struct:
		sections := make([]string, x.NumFields())
		for i := 0; i < x.NumFields(); i++ {
			f := x.Field(i)

			// TODO -- confirm that the original encoder ignores unexported fields
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
			section, err := buildCodeSectionDecode(f.Type(), p, nextVarName, false, "", options)
			if err != nil {
				return "", err
			}

			sections[i] = section
		}

		return strings.Join(sections, "\n\n"), nil

	case *types.Array:
		elem := x.Elem()

		if isByte(elem) {
			return skyencoder.BuildDecodeByteArray(varName, options), nil
		}

		elemSection, err := buildCodeSectionDecode(elem, p, "x", false, "", nil)
		if err != nil {
			return "", err
		}

		return skyencoder.BuildDecodeArray(varName, "x", elemSection, options), nil

	case *types.Slice:
		elem := x.Elem()

		if isByte(elem) {
			return skyencoder.BuildDecodeByteSlice(varName, options), nil
		}

		elemSection, err := buildCodeSectionDecode(elem, p, "x", false, "", nil)
		if err != nil {
			return "", err
		}

		return skyencoder.BuildDecodeSlice(varName, "x", elemSection, sliceTypeName(x, p), options), nil

	default:
		return "", fmt.Errorf("Unhandled type %T for var %s", x, varName)
	}
}

func sliceTypeName(t *types.Slice, p *types.Package) string {
	elemType := typeNameOf(t.Elem(), p)
	return fmt.Sprintf("[]%s", elemType)
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
		return t.String()
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

func parseTag(tag string) (bool, *skyencoder.Options, error) {
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

	opts := &skyencoder.Options{}
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
		case types.Bool:
			// TODO -- determine if copying an array of bools is the same as encoding them separately (endianness could be problem)
			return true
		case types.Int8:
			// TODO -- determine if copying an array of bools is the same as encoding them separately (endianness could be problem)
			return true
		case types.Uint8:
			return true
		default:
			return false
		}
	default:
		return false
	}
}
