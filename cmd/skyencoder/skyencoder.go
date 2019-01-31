package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/skycoin/skyencoder"
)

/* TODO

- benchmark results against the encoder
- add go:generate in skycoin
- add skycoin tests (verify entire db can be loaded)

TO DOCUMENT:

- (wiki) document encoding format
- (wiki) document usage in golang (maxlen, omitempty, "-", ignores unexported fields)

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
		log.Fatal("skyencoder.LoadProgram failed:", err)
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

	src, err := skyencoder.BuildStructEncoder(sInfo, *destPackage, fmtFilename)
	if err != nil {
		log.Fatal("skyencoder.BuildStructEncoder failed:", err)
	}

	debugPrintln(string(src))

	outputFn := *outputFilename
	if outputFn == "" {
		// If the input is a filename, put next to the file
		// If the input is a package, put in the package
		outputFn = fmt.Sprintf("%s_skyencoder.go", toSnakeCase(*structName))
	}

	outputPth := *outputPath
	if outputPth == "" {
		outputPth = destPath
	}
	outputFn = filepath.Join(outputPth, outputFn)

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
