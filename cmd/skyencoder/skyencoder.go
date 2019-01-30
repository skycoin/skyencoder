package main

import (
	"flag"
	"fmt"
	"go/build"
	"go/types"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/golang/tools/go/loader"
	"github.com/skycoin/skyencoder"
)

/* TODO

- determine if copying an array of bools and int8s is the same as encoding them separately (endianness could be problem)
- test cases

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
	structName  = flag.String("struct", "", "struct name, must be set")
	output      = flag.String("output", "", "output file name; default srcdir/<struct_name>_skyencoder.go")
	buildTags   = flag.String("tags", "", "comma-separated list of build tags to apply")
	destPackage = flag.String("package", "", "package name for the output; if not provided, defaults to the struct's package")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of skyencoder:\n")
	fmt.Fprintf(os.Stderr, "\tskyencoder [flags] -struct T [go import path e.g. github.com/skycoin/skycoin/src/coin]\n")
	fmt.Fprintf(os.Stderr, "\tskyencoder [flags] -struct T files... # Must be a single package\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("skyencoder: ")

	flag.Usage = usage
	flag.Parse()

	if *structName == "" {
		flag.Usage()
		os.Exit(2)
	}

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

	sInfo, err := skyencoder.FindStructInfoInProgram(program, *structName)
	if err != nil {
		log.Fatalf("Program did not contain valid struct for name %s: %v", *structName, err)
	}
	if sInfo == nil {
		log.Fatal("Program does not contain type:", *structName)
	}

	// Determine if the arg is a directory or multiple files
	// If it is a directory, construct an artificial filename in that directory for goimports formatting,
	// otherwise use the first filename specified (they must all be in the same package)
	fmtFilename := args[0]
	destPath := filepath.Dir(args[0])

	stat, err := os.Stat(args[0])
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		}
		// argument is a import path e.g. "github.com/skycoin/skycoin/src/coin"
		destPath, err = skyencoder.FindDiskPathOfImport(sInfo.Package.Path())
		if err != nil {
			log.Fatal(err)
		}
		fmtFilename = filepath.Join(sInfo.Package.Path(), "foo123123123123999.go")
	} else if stat.IsDir() {
		destPath = args[0]
		fmtFilename = filepath.Join(args[0], "foo123123123123999.go")
	}

	src, err := skyencoder.BuildStructEncoder(*structName, sInfo, *destPackage, fmtFilename)
	if err != nil {
		log.Fatal("skyencoder.BuildStructEncoder failed:", err)
	}

	debugPrintln(string(src))

	outputFn := *output
	if outputFn == "" {
		// If the input is a filename, put next to the file
		// If the input is a package, put in the package
		fn := fmt.Sprintf("%s_skyencoder.go", toSnakeCase(*structName))
		outputFn = filepath.Join(destPath, fn)
	}

	if err := ioutil.WriteFile(outputFn, src, 0644); err != nil {
		log.Fatal("ioutil.WriteFile failed:", err)
	}
}

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
