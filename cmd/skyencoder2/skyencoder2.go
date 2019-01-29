package main

import (
	"errors"
	"flag"
	"fmt"
	"go/build"
	"go/format"
	"go/types"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/faith/structtag"
	"github.com/golang/tools/go/loader"
	"github.com/skycoin/skyencoder"
)

/* TODO

- Add option to specify the package name of the generated file? Otherwise generate in the package where the struct is found
- Use package of the struct and the destination package to include/exclude the package qualifier on named types
- Format with imports automatically (can use goimports? or collect external package names used?)
- Size calculation for buffer: precompute size instead of separate additions, allowing multiplication of fixed size element slices to avoid an extra function call
-	OR - allow the var counter name to be variable, and use a different one (e.g. "j") at different layers
- Size calculation for buffer: code to check if length of slice/string/map would exceed math.MaxUint32
- Size calculation for buffer, and/or encoding: check maxlen property and return error
- Support embedded fields? (required for coin.SignedBlock)

*/

var (
	structNames = flag.String("struct", "", "comma-separated list of struct names; must be set")
	output      = flag.String("output", "", "output file name; default srcdir/<type>_string.go")
	buildTags   = flag.String("tags", "", "comma-separated list of build tags to apply")
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

	fmt.Println("args:", args)

	structs := make([]*types.Struct, len(structNames))
	for i, name := range structNames {
		s, err := findStructTypeInProgram(program, name)
		if err != nil {
			log.Fatalf("Program did not contain valid struct for name %s: %v", name, err)
		}
		if s == nil {
			log.Fatal("Program does not contain type:", name)
		}

		structs[i] = s
	}

	fmt.Println()
	fmt.Println("---- ENCODE ----")
	fmt.Println()

	for i, s := range structs {
		fmt.Println(s.String())
		section, err := buildCodeSectionEncode(s, "obj", false, nil)
		if err != nil {
			log.Fatal("buildCodeSectionEncode failed:", err)
		}

		f := wrapEncodeFunc(structNames[i], section)

		formattedBytes, err := format.Source([]byte(f))
		if err != nil {
			log.Fatal("format.Source failed:", err)
		}

		fmt.Println(string(formattedBytes))
	}

	fmt.Println()
	fmt.Println("---- DECODE ----")
	fmt.Println()

	for i, s := range structs {
		fmt.Println(s.String())
		section, err := buildCodeSectionDecode(s, "obj", false, "", nil)
		if err != nil {
			log.Fatal("buildCodeSectionDecode failed:", err)
		}

		f := wrapDecodeFunc(structNames[i], section)

		// fmt.Println(f)

		formattedBytes, err := format.Source([]byte(f))
		if err != nil {
			log.Fatal("format.Source failed:", err)
		}

		fmt.Println(string(formattedBytes))
	}

	fmt.Println()
	fmt.Println("---- ENCODE SIZE ----")
	fmt.Println()

	for i, s := range structs {
		fmt.Println(s.String())
		section, _, err := buildCodeSectionEncodeSize(s, "obj", "i", nil)
		if err != nil {
			log.Fatal("buildCodeSectionEncodeSize failed:", err)
		}

		f := wrapEncodeSizeFunc(structNames[i], section, "i")

		fmt.Println(f)

		formattedBytes, err := format.Source([]byte(f))
		if err != nil {
			log.Fatal("format.Source failed:", err)
		}

		fmt.Println(string(formattedBytes))
	}
}

func wrapEncodeFunc(structName, funcBody string) string {
	return fmt.Sprintf(`
func Encode%[1]s(e *encoder.Encoder, obj *%[1]s) {
		%[2]s
	}
	`, structName, funcBody)
}

func wrapDecodeFunc(structName, funcBody string) string {
	return fmt.Sprintf(`
func Decode%[1]s(e *encoder.Decoder, obj *%[1]s) error {
		%[2]s
	}
	`, structName, funcBody)
}

func wrapEncodeSizeFunc(structName, funcBody, counterName string) string {
	return fmt.Sprintf(`
func EncodeSize%[1]s(obj *%[1]s) int {
		%[3]s := 0

		%[2]s

		return %[3]s
	}
	`, structName, funcBody, counterName)
}

func isDir(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}

func findStructTypeInProgram(p *loader.Program, name string) (*types.Struct, error) {
	for _, pk := range p.Created {
		s, err := findStructTypeInPackage(pk, name)
		if err != nil {
			return nil, err
		}
		if s != nil {
			return s, nil
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

	fmt.Printf("buildCodeSectionEncode type=%T varName=%s castType=%v options=%+v\n", t, varName, castType, options)

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

			// TODO -- determine if the original encoder handles embedded fields or not
			if f.Embedded() {
				return "", fmt.Errorf("struct referenced as %q contains embedded field", varName)
			}

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

func buildCodeSectionEncodeSize(t types.Type, varName, counterName string, options *skyencoder.Options) (string, bool, error) {
	// castType applies to basic int types; if true, an additional cast will be made in the generated code.
	// This is to convert types like "type Foo int8" back to int8

	fmt.Printf("buildCodeSectionEncodeSize type=%T varName=%s counterName=%s options=%+v\n", t, varName, counterName, options)

	if options != nil {
		if options.OmitEmpty && !omitEmptyIsValid(t) {
			return "", false, errors.New("omitempty is only valid for array, slice, map and string")
		}
	}

	switch x := t.(type) {
	case *types.Named:
		return buildCodeSectionEncodeSize(x.Underlying(), varName, counterName, options)

	case *types.Basic:
		switch x.Kind() {
		case types.Bool:
			return skyencoder.BuildEncodeSizeBool(varName, counterName, options), false, nil
		case types.Int8:
			return skyencoder.BuildEncodeSizeInt8(varName, counterName, options), false, nil
		case types.Int16:
			return skyencoder.BuildEncodeSizeInt16(varName, counterName, options), false, nil
		case types.Int32:
			return skyencoder.BuildEncodeSizeInt32(varName, counterName, options), false, nil
		case types.Int64:
			return skyencoder.BuildEncodeSizeInt64(varName, counterName, options), false, nil
		case types.Uint8:
			return skyencoder.BuildEncodeSizeUint8(varName, counterName, options), false, nil
		case types.Uint16:
			return skyencoder.BuildEncodeSizeUint16(varName, counterName, options), false, nil
		case types.Uint32:
			return skyencoder.BuildEncodeSizeUint32(varName, counterName, options), false, nil
		case types.Uint64:
			return skyencoder.BuildEncodeSizeUint64(varName, counterName, options), false, nil
		case types.String:
			return skyencoder.BuildEncodeSizeString(varName, counterName, options), true, nil
		default:
			return "", false, fmt.Errorf("Unhandled *types.Basic type %s for var %s", x.Name(), varName)
		}

	case *types.Array:
		elem := x.Elem()

		if isByte(elem) {
			return skyencoder.BuildEncodeSizeByteArray(varName, counterName, x.Len(), options), false, nil
		}

		elemSection, isDynamic, err := buildCodeSectionEncodeSize(elem, "x", counterName, nil)
		if err != nil {
			return "", false, err
		}

		elemSectionCounterName := counterName
		if !isDynamic {
			elemSectionCounterName = counterName + counterName
		}
		elemSection = fmt.Sprintf(elemSection, counterName)

		return skyencoder.BuildEncodeSizeArray(varName, counterName, "x", elemSectionCounterName, elemSection, x.Len(), isDynamic, options), isDynamic, nil

	case *types.Slice:
		elem := x.Elem()

		if isByte(elem) {
			return skyencoder.BuildEncodeSizeByteSlice(varName, counterName, options), false, nil
		}

		elemSection, isDynamic, err := buildCodeSectionEncodeSize(elem, "x", counterName, nil)
		if err != nil {
			return "", false, err
		}

		elemSectionCounterName := counterName
		if !isDynamic {
			elemSectionCounterName = counterName + counterName
		}
		elemSection = fmt.Sprintf(elemSection, counterName)

		return skyencoder.BuildEncodeSizeSlice(varName, counterName, "x", elemSectionCounterName, elemSection, isDynamic, options), true, nil

	case *types.Map:
		keySection, isDynamicKey, err := buildCodeSectionEncodeSize(x.Key(), "k", counterName, nil)
		if err != nil {
			return "", false, err
		}

		elemSection, isDynamicElem, err := buildCodeSectionEncodeSize(x.Elem(), "v", counterName, nil)
		if err != nil {
			return "", false, err
		}

		innerCounterName := counterName
		if !isDynamicKey && !isDynamicKey {
			innerCounterName = counterName + counterName
		}

		return skyencoder.BuildEncodeSizeMap(varName, counterName, "k", "v", innerCounterName, keySection, elemSection, isDynamicKey, isDynamicElem, options), true, nil

	case *types.Struct:
		isDynamic := false
		sections := make([]string, x.NumFields())
		for i := 0; i < x.NumFields(); i++ {
			f := x.Field(i)

			// TODO -- determine if the original encoder handles embedded fields or not
			if f.Embedded() {
				return "", false, fmt.Errorf("struct referenced as %q contains embedded field", varName)
			}

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
			section, sectionIsDynamic, err := buildCodeSectionEncodeSize(f.Type(), nextVarName, counterName, options)
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

func buildCodeSectionDecode(t types.Type, varName string, castType bool, typeName string, options *skyencoder.Options) (string, error) {
	// castType applies to basic int types; if true, an additional cast will be made in the generated code.
	// This is to convert types like "type Foo int8" back to int8

	fmt.Printf("buildCodeSectionDecode type=%T varName=%s castType=%v options=%+v\n", t, varName, castType, options)

	if options != nil {
		if options.MaxLength != 0 && !maxLenIsValid(t) {
			return "", errors.New("maxlen is only valid for slice and string")
		}
	}

	switch x := t.(type) {
	case *types.Named:
		// TODO -- the typeName x.String() is used to cast or allocate values.
		// x.String() includes the package name. This is correct if we are generating code to a different package than the type is declared in,
		// but incorrect if we want it in the same package.
		// Need to detect if we are generating code to the same package that the type is declared in,
		// and if so, then use x.Obj().Name() which will return the type name without package
		return buildCodeSectionDecode(x.Underlying(), varName, true, x.String(), options)

	case *types.Basic:
		if typeName == "" {
			typeName = x.Name()
		}

		fmt.Printf("types.Basic type name is %s\n", typeName)

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
		keySection, err := buildCodeSectionDecode(x.Key(), "k", false, "", nil)
		if err != nil {
			return "", err
		}

		elemSection, err := buildCodeSectionDecode(x.Elem(), "v", false, "", nil)
		if err != nil {
			return "", err
		}

		return skyencoder.BuildDecodeMap(varName, "k", "v", keySection, elemSection, mapTypeName(x), options), nil

	case *types.Struct:
		sections := make([]string, x.NumFields())
		for i := 0; i < x.NumFields(); i++ {
			f := x.Field(i)

			// TODO -- determine if the original encoder handles embedded fields or not
			if f.Embedded() {
				return "", fmt.Errorf("struct referenced as %q contains embedded field", varName)
			}

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
			section, err := buildCodeSectionDecode(f.Type(), nextVarName, false, "", options)
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

		elemSection, err := buildCodeSectionDecode(elem, "x", false, "", nil)
		if err != nil {
			return "", err
		}

		return skyencoder.BuildDecodeArray(varName, "x", elemSection, options), nil

	case *types.Slice:
		elem := x.Elem()

		if isByte(elem) {
			return skyencoder.BuildDecodeByteSlice(varName, options), nil
		}

		elemSection, err := buildCodeSectionDecode(elem, "x", false, "", nil)
		if err != nil {
			return "", err
		}

		return skyencoder.BuildDecodeSlice(varName, "x", elemSection, sliceTypeName(x), options), nil

	default:
		return "", fmt.Errorf("Unhandled type %T for var %s", x, varName)
	}
}

func sliceTypeName(t *types.Slice) string {
	elemType := typeNameOf(t.Elem())
	return fmt.Sprintf("[]%s", elemType)
}

func mapTypeName(t *types.Map) string {
	// t.String() will return a type with fully qualified import paths, e.g.
	// map[int32]coin.UxOut will return "map[int32]github.com/skycoin/skycoin/src/coin.UxOut"
	// I can't find a way to get the type name without the import path, other than constructing it manually
	// TODO - this will include the package name regardless of the relative package
	keyType := typeNameOf(t.Key())
	elemType := typeNameOf(t.Elem())
	return fmt.Sprintf("map[%s]%s", keyType, elemType)
}

func typeNameOf(t types.Type) string {
	switch x := t.(type) {
	case *types.Named:
		obj := x.Obj()
		return fmt.Sprintf("%s.%s", obj.Pkg().Name(), obj.Name())
	case *types.Basic:
		return x.Name()
	case *types.Map:
		return mapTypeName(x)
	case *types.Slice:
		return sliceTypeName(x)
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
	case *types.Slice:
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

func walkStructFields(s *types.Struct, prefix string) {
	for i := 0; i < s.NumFields(); i++ {
		f := s.Field(i)
		tag := s.Tag(i)

		fmt.Printf("%sStruct field i=%d name=%s type=%T, anon=%v embed=%v export=%v tag=%s\n", prefix, i, f.Name(), f.Type(), f.Anonymous(), f.Embedded(), f.Exported(), tag)

		ft := f.Type()
		switch x := ft.(type) {
		case *types.Named:
			handleType(x.Underlying(), prefix)
		default:
			handleType(ft, prefix)
		}
	}
}

func handleType(t types.Type, prefix string) {
	switch x := t.(type) {
	case *types.Basic:
		switch x.Kind() {
		case types.Bool:
			// TODO -- here is where we write out generated code into the buffer
			fmt.Printf("%s^^^BOOL^^^\n", prefix)
		case types.Int8:
			fmt.Printf("%s^^^INT8^^^\n", prefix)
		case types.Int16:
			fmt.Printf("%s^^^INT16^^^\n", prefix)
		case types.Int32:
			fmt.Printf("%s^^^INT32^^^\n", prefix)
		case types.Int64:
			fmt.Printf("%s^^^INT64^^^\n", prefix)
		case types.Uint8:
			fmt.Printf("%s^^^UINT8^^^\n", prefix)
		case types.Uint16:
			fmt.Printf("%s^^^UINT16^^^\n", prefix)
		case types.Uint32:
			fmt.Printf("%s^^^UINT32^^^\n", prefix)
		case types.Uint64:
			fmt.Printf("%s^^^UINT64^^^\n", prefix)
		case types.String:
			fmt.Printf("%s^^^STRING^^^\n", prefix)
		default:
			fmt.Printf("%sUNHANDLED BASIC TYPE %s\n", prefix, x.Name())
		}

	case *types.Map:
		fmt.Printf("%s^^^MAP^^^ keyType=%T elemType=%T\n", prefix, x.Key(), x.Elem())
	case *types.Struct:
		walkStructFields(x, prefix+"\t")
	case *types.Array:
		fmt.Printf("%s^^^ARRAY^^^ len=%d elemType=%T\n", prefix, x.Len(), x.Elem())
	case *types.Slice:
		fmt.Printf("%s^^^SLICE^^^ elemType=%T\n", prefix, x.Elem())
	default:
		fmt.Printf("%sUNHANDLED TYPE %T\n", prefix, x)
	}
}

func fullTypeName(x *types.Named) string {
	fTypeName := x.Obj().Name()
	if pkg := x.Obj().Pkg(); pkg != nil {
		fTypeName = fmt.Sprintf("%s.%s", pkg.Name(), fTypeName)
	}
	return fTypeName
}
