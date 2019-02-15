package skyencoder

import (
	"fmt"
	"go/build"
	"go/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/loader"

	// needed to verify test output
	_ "github.com/skycoin/skycoin/src/coin" // needed to verify test output
)

func removeFile(fn string) {
	os.Remove(fn)
}

func verifyProgramCompiles(t *testing.T, dir string) {
	// Load the package with the least restrictive parsing and type checking,
	// so that a package that doesn't compile can still have a struct declaration extracted
	cfg := loader.Config{
		Build:      &build.Default,
		ParserMode: 0,
		TypeChecker: types.Config{
			IgnoreFuncBodies:         false, // ignore functions
			FakeImportC:              false, // ignore import "C"
			DisableUnusedImportCheck: false, // ignore unused imports
		},
		AllowErrors: false,
	}

	loadTests := true
	unused, err := cfg.FromArgs([]string{dir}, loadTests)
	if err != nil {
		t.Fatal(err)
	}
	if len(unused) != 0 {
		t.Fatalf("Had unused args to cfg.FromArgs: %v", unused)
	}

	_, err = cfg.Load()
	if err != nil {
		t.Fatal(err)
	}
}

func testBuildCode(t *testing.T, structName, filename string) []byte {
	program, err := LoadProgram([]string{"."}, nil)
	if err != nil {
		t.Fatal(err)
	}

	sInfo, err := FindStructInfoInProgram(program, structName)
	if err != nil {
		t.Fatal(err)
	}

	src, err := BuildStructEncoder(sInfo, "", filename, true)
	if err != nil {
		t.Fatal(err)
	}

	// Go's parser and loader packages do not accept []byte, only filenames, so save the result to disk
	// and clean it up after the test
	defer removeFile(filename)
	err = ioutil.WriteFile(filename, src, 0644)
	if err != nil {
		t.Fatal(err)
	}

	verifyProgramCompiles(t, ".")

	return src
}

func TestBuildSkycoinSignedBlock(t *testing.T) {
	importPath := "github.com/skycoin/skycoin/src/coin"
	structName := "SignedBlock"

	fullPath, err := FindDiskPathOfImport(importPath)
	if err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(fullPath, "signed_block_skyencoder_xxxyyy.go")

	program, err := LoadProgram([]string{importPath}, nil)
	if err != nil {
		t.Fatal(err)
	}

	sInfo, err := FindStructInfoInProgram(program, structName)
	if err != nil {
		t.Fatal(err)
	}

	src, err := BuildStructEncoder(sInfo, "", filename, true)
	if err != nil {
		t.Fatal(err)
	}

	// Go's parser and loader packages do not accept []byte, only filenames, so save the result to disk
	// and clean it up after the test
	defer removeFile(filename)
	err = ioutil.WriteFile(filename, src, 0644)
	if err != nil {
		t.Fatal(err)
	}

	verifyProgramCompiles(t, importPath)
}

func testBuildCodeFails(t *testing.T, structName, filename string) {
	program, err := LoadProgram([]string{"."}, nil)
	if err != nil {
		t.Fatal(err)
	}

	sInfo, err := FindStructInfoInProgram(program, structName)
	if err != nil {
		t.Fatal(err)
	}

	_, err = BuildStructEncoder(sInfo, "", filename, true)
	if err == nil {
		t.Fatal("Expected BuildStructEncoder error")
	}
}

/* Invalid structs */

type MaxLenInt struct {
	Int64 int64 `enc:",maxlen=4"`
}

type MaxLenInvalid struct {
	String string `enc:",maxlen=foo"`
}

type OmitEmptyInt struct {
	Int64 int64 `enc:',omitempty"`
}

type OmitEmptyNotFinal struct {
	Int64  int64
	Extra  []byte `enc:",omitempty"`
	String string
}

type EmptyStructSlice1 struct {
	Foo []struct{}
}

type EmptyStructSlice2 struct {
	Foo []struct {
		unexported int64
	}
}

type EmptyStructSlice3 struct {
	Foo []struct {
		Ignored int64 `enc:"-"`
	}
}

func TestBuildFails(t *testing.T) {
	cases := []struct {
		name string
	}{
		{
			name: "MaxLenInt",
		},
		{
			name: "MaxLenInvalid",
		},
		{
			name: "OmitEmptyInt",
		},
		{
			name: "OmitEmptyNotFinal",
		},
		{
			name: "EmptyStructSlice1",
		},
		{
			name: "EmptyStructSlice2",
		},
		{
			name: "EmptyStructSlice3",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testBuildCodeFails(t, tc.name, fmt.Sprintf("./%s_skyencoder_test.go", ToSnakeCase(tc.name)))
		})
	}
}
