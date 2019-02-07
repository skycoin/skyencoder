package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/skycoin/skyencoder"
)

/* TODO

TO DOCUMENT:

- (wiki) document encoding format
- (wiki) document usage in golang (maxlen, omitempty, "-", ignores unexported fields)
- omitempty can only be used in the last field of a struct and only at a toplevel struct
- decoding specifics (empty map/slice are always left nil)
- []struct{} cannot be properly decoded since it is empty

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
	structName     = flag.String("struct", "", "struct name, must be set")
	outputFilename = flag.String("output-file", "", "output file name; default <struct_name>_skyencoder.go")
	outputPath     = flag.String("output-path", "", "output path; defaults to the package's path, or the file's containing folder")
	buildTags      = flag.String("tags", "", "comma-separated list of build tags to apply")
	destPackage    = flag.String("package", "", "package name for the output; if not provided, defaults to the struct's package")
	silent         = flag.Bool("silent", false, "disable all non-error log output")
	noTest         = flag.Bool("no-test", false, "disable generating the _test.go file (test files require github.com/google/go-cmp/cmp and github.com/skycoin/encodertest)")
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

	program, err := skyencoder.LoadProgram(args, tags)
	if err != nil {
		log.Fatal("skyencoder.LoadProgram failed: ", err)
	}

	debugPrintln("args:", args)

	structInfo, err := skyencoder.FindStructInfoInProgram(program, *structName)
	if err != nil {
		log.Fatalf("Program did not contain valid struct for name %s: %v", *structName, err)
	}
	if structInfo == nil {
		log.Fatal("Program does not contain struct: ", *structName)
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
		destPath, err = skyencoder.FindDiskPathOfImport(structInfo.Package.Path())
		if err != nil {
			log.Fatal(err)
		}
		fmtFilename = filepath.Join(structInfo.Package.Path(), "foo123123123123999.go")
	} else if stat.IsDir() {
		destPath = args[0]
		fmtFilename = filepath.Join(args[0], "foo123123123123999.go")
	}

	src, err := skyencoder.BuildStructEncoder(structInfo, *destPackage, fmtFilename)
	if err != nil {
		log.Fatal("skyencoder.BuildStructEncoder failed: ", err)
	}

	var testSrc []byte
	if !*noTest {
		testSrc, err = skyencoder.BuildStructEncoderTest(structInfo, *destPackage, fmtFilename)
		if err != nil {
			log.Fatal("skyencoder.BuildStructEncoderTest failed: ", err)
		}
	}

	debugPrintln(string(src))

	outputFn := *outputFilename
	if outputFn == "" {
		// If the input is a filename, put next to the file
		// If the input is a package, put in the package
		outputFn = fmt.Sprintf("%s_skyencoder.go", skyencoder.ToSnakeCase(*structName))
	}

	outputPth := *outputPath
	if outputPth == "" {
		outputPth = destPath
	}
	outputFn = filepath.Join(outputPth, outputFn)

	if !*silent {
		log.Printf("Writing skyencoder for struct %q to file %q", *structName, outputFn)
	}

	if err := ioutil.WriteFile(outputFn, src, 0644); err != nil {
		log.Fatal("ioutil.WriteFile failed: ", err)
	}

	if !*noTest {
		outputExt := filepath.Ext(outputFn)
		base := outputFn[:len(outputFn)-len(outputExt)]
		testOutputFn := fmt.Sprintf("%s_test%s", base, outputExt)

		if !*silent {
			log.Printf("Writing skyencoder tests for struct %q to file %q", *structName, testOutputFn)
		}

		if err := ioutil.WriteFile(testOutputFn, testSrc, 0644); err != nil {
			log.Fatal("ioutil.WriteFile failed: ", err)
		}
	}
}
