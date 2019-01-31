package skyencoder

import (
	"go/build"
	"go/types"
	"io/ioutil"
	"os"
	"testing"

	"golang.org/x/tools/go/loader"
)

type Coins uint64

type Hash [20]byte

type DynamicStruct struct {
	Foo []string
	Bar int32
	Baz string
}

type StaticStruct struct {
	A    byte
	B    int32
	Hash Hash
}

type DemoStruct struct {
	Uint8                  uint8
	Uint16                 uint16
	Uint32                 uint32
	Uint64                 uint64
	Int8                   int8
	Int16                  int16
	Int32                  int32
	Int64                  int64
	Byte                   byte
	String                 string
	DynamicStruct          DynamicStruct
	StaticStruct           StaticStruct
	NamedByteArray         Hash
	NamedBasicType         Coins
	DynamicKeyMap          map[string]uint16
	DynamicElemMap         map[uint16]string
	DynamicMap             map[string]string
	DynamicNestedMap       map[string][10][]string
	DynamicArrayKeyMap     map[[10]string]uint32
	StaticByteArrayKeyMap  map[Hash]uint16
	StaticByteArrayElemMap map[uint16]Hash
	StaticStructMap        map[int32]StaticStruct
	SetMap                 map[int32]struct{}
	DynamicStringArray     [10]string
	StaticBasicArray       [10]int64
	StaticStructArray      [10]StaticStruct
	DynamicSlice           []string
	StaticSlice            []StaticStruct

	Uint8Slice                  []uint8
	Uint16Slice                 []uint16
	Uint32Slice                 []uint32
	Uint64Slice                 []uint64
	Int8Slice                   []int8
	Int16Slice                  []int16
	Int32Slice                  []int32
	Int64Slice                  []int64
	ByteSlice                   []byte
	StringSlice                 []string
	DynamicStructSlice          []DynamicStruct
	StaticStructSlice           []StaticStruct
	NamedByteArraySlice         []Hash
	NamedBasicTypeSlice         []Coins
	DynamicKeyMapSlice          []map[string]uint16
	DynamicElemMapSlice         []map[uint16]string
	DynamicMapSlice             []map[string]string
	DynamicNestedMapSlice       []map[string][10][]string
	DynamicArrayKeyMapSlice     []map[[10]string]uint32
	StaticByteArrayKeyMapSlice  []map[Hash]uint16
	StaticByteArrayElemMapSlice []map[uint16]Hash
	StaticStructMapSlice        []map[int32]StaticStruct
	SetMapSlice                 []map[int32]struct{}
	DynamicStringArraySlice     [][10]string
	StaticBasicArraySlice       [][10]int64
	StaticStructArraySlice      [][10]StaticStruct
	DynamicSliceSlice           [][]string
	StaticSliceSlice            [][]StaticStruct

	ignored    uint64 `enc:"-"`
	unexported uint64

	StringMaxLen    string          `enc:",maxlen=4"`
	MapMaxLen       map[int64]uint8 `enc:",maxlen=5"`
	ByteSliceMaxLen []byte          `enc:",maxlen=6"`
	SliceMaxLen     []int64         `enc:",maxlen=7"`
}

type DemoStructOmitEmpty struct {
	Int32     int32
	OmitEmpty []byte `enc:",omitempty"`
}

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

func testBuildCode(t *testing.T, structName, filename string) {
	program, err := LoadProgram([]string{"."}, nil)
	if err != nil {
		t.Fatal(err)
	}

	sInfo, err := FindStructInfoInProgram(program, structName)
	if err != nil {
		t.Fatal(err)
	}

	src, err := BuildStructEncoder(sInfo, "", filename)
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
}

func TestBuildDemoStruct(t *testing.T) {
	testBuildCode(t, "DemoStruct", "./demo_struct_skyencoder_test.go")
}

func TestBuildOmitEmptyStruct(t *testing.T) {
	testBuildCode(t, "DemoStructOmitEmpty", "./demo_struct_omit_empty_skyencoder_test.go")
}
