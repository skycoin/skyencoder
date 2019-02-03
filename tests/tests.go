package tests

import (
	"testing"

	"github.com/skycoin/skycoin/src/cipher/encoder"
)

/* Demo structs for test generation */

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
	Float32                float32
	Float64                float64
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

/* maxlen tag tests */

type MaxLenStringStruct1 struct {
	Foo string `enc:",maxlen=3"`
}

type MaxLenStringStruct2 struct {
	Foo string `enc:",maxlen=4"`
}

type MaxLenAllStruct1 struct {
	Foo string           `enc:",maxlen=3"`
	Bar []int64          `enc:",maxlen=3"`
	Baz map[uint64]int64 `enc:",maxlen=3"`
}

type MaxLenAllStruct2 struct {
	Foo string           `enc:",maxlen=4"`
	Bar []int64          `enc:",maxlen=4"`
	Baz map[uint64]int64 `enc:",maxlen=4"`
}

type MaxLenNestedSliceStruct1 struct {
	Foo []MaxLenStringStruct1
}

type MaxLenNestedSliceStruct2 struct {
	Foo []MaxLenStringStruct2
}

type MaxLenNestedMapKeyStruct1 struct {
	Foo map[MaxLenStringStruct1]int64
}

type MaxLenNestedMapKeyStruct2 struct {
	Foo map[MaxLenStringStruct2]int64
}

type MaxLenNestedMapValueStruct1 struct {
	Foo map[int64]MaxLenStringStruct1
}

type MaxLenNestedMapValueStruct2 struct {
	Foo map[int64]MaxLenStringStruct2
}

type OnlyOmitEmptyStruct struct {
	Extra []byte `enc:",omitempty"`
}

type OmitEmptyStruct struct {
	Foo   string
	Extra []byte `enc:",omitempty"`
}

type OmitEmptyMaxLenStruct1 struct {
	Foo   string
	Extra []byte `enc:",omitempty,maxlen=3"`
}

type OmitEmptyMaxLenStruct2 struct {
	Foo   string
	Extra []byte `enc:",maxlen=4,omitempty"`
}

func TestMaxLenStringStructExceeded(t *testing.T) {
	obj2 := MaxLenStringStruct2{
		Foo: "1234",
	}

	n := EncodeSizeMaxLenStringStruct2(&obj2)

	data := make([]byte, n)
	e := &encoder.Encoder{
		Buffer: data[:],
	}

	err := EncodeMaxLenStringStruct1(e, &MaxLenStringStruct1{
		Foo: "1234",
	})
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("EncodeMaxLenStringStruct1 expected encoder.ErrMaxLenExceeded")
	}

	err = EncodeMaxLenStringStruct2(&encoder.Encoder{
		Buffer: data[:],
	}, &MaxLenStringStruct2{
		Foo: "1234",
	})
	if err != nil {
		t.Fatalf("EncodeMaxLenStringStruct2 unexpected error: %v", err)
	}

	var obj1 MaxLenStringStruct1
	err = DecodeMaxLenStringStruct1(&encoder.Decoder{
		Buffer: data[:],
	}, &obj1)
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("DecodeMaxLenStringStruct1 expected encoder.ErrMaxLenExceeded")
	}
}

func testMaxLenAllStructExceeded(t *testing.T, obj1 MaxLenAllStruct1, obj2 MaxLenAllStruct2) {
	n := EncodeSizeMaxLenAllStruct2(&obj2)

	data := make([]byte, n)
	e := &encoder.Encoder{
		Buffer: data[:],
	}

	obj1Bad := MaxLenAllStruct1(obj2)
	err := EncodeMaxLenAllStruct1(e, &obj1Bad)
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("EncodeMaxLenAllStruct1 expected encoder.ErrMaxLenExceeded")
	}

	err = EncodeMaxLenAllStruct2(&encoder.Encoder{
		Buffer: data[:],
	}, &obj2)
	if err != nil {
		t.Fatalf("EncodeMaxLenAllStruct2 unexpected error: %v", err)
	}

	var obj1Empty MaxLenAllStruct1
	err = DecodeMaxLenAllStruct1(&encoder.Decoder{
		Buffer: data[:],
	}, &obj1Empty)
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("DecodeMaxLenAllStruct1 expected encoder.ErrMaxLenExceeded")
	}
}

func TestMaxLenAllStructExceeded(t *testing.T) {
	cases := []struct {
		name    string
		obj1    MaxLenAllStruct1
		obj1Bad MaxLenAllStruct1
		obj2    MaxLenAllStruct2
	}{
		{
			name: "string exceeds",
			obj1: MaxLenAllStruct1{
				Foo: "123",
			},
			obj1Bad: MaxLenAllStruct1{
				Foo: "123",
			},
			obj2: MaxLenAllStruct2{
				Foo: "1234",
			},
		},

		{
			name: "slice exceeds",
			obj1: MaxLenAllStruct1{
				Bar: []int64{1, 2, 3},
			},
			obj2: MaxLenAllStruct2{
				Bar: []int64{1, 2, 3, 4},
			},
		},

		{
			name: "map exceeds",
			obj1: MaxLenAllStruct1{
				Baz: map[uint64]int64{
					1: 2,
					3: 4,
					5: 6,
				},
			},
			obj2: MaxLenAllStruct2{
				Baz: map[uint64]int64{
					1: 2,
					3: 4,
					5: 6,
					7: 8,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testMaxLenAllStructExceeded(t, tc.obj1, tc.obj2)
		})
	}
}

func TestNestedMaxLenNestedSliceStruct(t *testing.T) {
	obj2 := MaxLenNestedSliceStruct2{
		Foo: []MaxLenStringStruct2{{
			Foo: "1234",
		}},
	}

	n := EncodeSizeMaxLenNestedSliceStruct2(&obj2)

	data := make([]byte, n)
	e := &encoder.Encoder{
		Buffer: data[:],
	}

	err := EncodeMaxLenNestedSliceStruct1(e, &MaxLenNestedSliceStruct1{
		Foo: []MaxLenStringStruct1{{
			Foo: "1234",
		}},
	})
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("EncodeMaxLenNestedSliceStruct1 expected encoder.ErrMaxLenExceeded")
	}

	err = EncodeMaxLenNestedSliceStruct2(&encoder.Encoder{
		Buffer: data[:],
	}, &MaxLenNestedSliceStruct2{
		Foo: []MaxLenStringStruct2{{
			Foo: "1234",
		}},
	})
	if err != nil {
		t.Fatalf("EncodeMaxLenStringStruct2 unexpected error: %v", err)
	}

	var obj1 MaxLenNestedSliceStruct1
	err = DecodeMaxLenNestedSliceStruct1(&encoder.Decoder{
		Buffer: data[:],
	}, &obj1)
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("DecodeMaxLenNestedSliceStruct1 expected encoder.ErrMaxLenExceeded")
	}
}

func TestNestedMaxLenNestedMapKeyStruct(t *testing.T) {
	obj2 := MaxLenNestedMapKeyStruct2{
		Foo: map[MaxLenStringStruct2]int64{
			{Foo: "1234"}: 1,
		},
	}

	n := EncodeSizeMaxLenNestedMapKeyStruct2(&obj2)

	data := make([]byte, n)
	e := &encoder.Encoder{
		Buffer: data[:],
	}

	err := EncodeMaxLenNestedMapKeyStruct1(e, &MaxLenNestedMapKeyStruct1{
		Foo: map[MaxLenStringStruct1]int64{
			{Foo: "1234"}: 1,
		},
	})
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("EncodeMaxLenNestedMapKeyStruct1 expected encoder.ErrMaxLenExceeded")
	}

	err = EncodeMaxLenNestedMapKeyStruct2(&encoder.Encoder{
		Buffer: data[:],
	}, &MaxLenNestedMapKeyStruct2{
		Foo: map[MaxLenStringStruct2]int64{
			{Foo: "1234"}: 1,
		},
	})
	if err != nil {
		t.Fatalf("EncodeMaxLenStringStruct2 unexpected error: %v", err)
	}

	var obj1 MaxLenNestedMapKeyStruct1
	err = DecodeMaxLenNestedMapKeyStruct1(&encoder.Decoder{
		Buffer: data[:],
	}, &obj1)
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("DecodeMaxLenNestedMapKeyStruct1 expected encoder.ErrMaxLenExceeded")
	}
}

func TestNestedMaxLenNestedMapValueStruct(t *testing.T) {
	obj2 := MaxLenNestedMapValueStruct2{
		Foo: map[int64]MaxLenStringStruct2{
			1: {Foo: "1234"},
		},
	}

	n := EncodeSizeMaxLenNestedMapValueStruct2(&obj2)

	data := make([]byte, n)
	e := &encoder.Encoder{
		Buffer: data[:],
	}

	err := EncodeMaxLenNestedMapValueStruct1(e, &MaxLenNestedMapValueStruct1{
		Foo: map[int64]MaxLenStringStruct1{
			1: {Foo: "1234"},
		},
	})
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("EncodeMaxLenNestedMapValueStruct1 expected encoder.ErrMaxLenExceeded")
	}

	err = EncodeMaxLenNestedMapValueStruct2(&encoder.Encoder{
		Buffer: data[:],
	}, &MaxLenNestedMapValueStruct2{
		Foo: map[int64]MaxLenStringStruct2{
			1: {Foo: "1234"},
		},
	})
	if err != nil {
		t.Fatalf("EncodeMaxLenStringStruct2 unexpected error: %v", err)
	}

	var obj1 MaxLenNestedMapValueStruct1
	err = DecodeMaxLenNestedMapValueStruct1(&encoder.Decoder{
		Buffer: data[:],
	}, &obj1)
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("DecodeMaxLenNestedMapValueStruct1 expected encoder.ErrMaxLenExceeded")
	}
}

func TestOmitEmptyMaxLenStructExceeded(t *testing.T) {
	obj2 := OmitEmptyMaxLenStruct2{
		Extra: []byte{1, 2, 3, 4},
	}

	n := EncodeSizeOmitEmptyMaxLenStruct2(&obj2)

	data := make([]byte, n)
	e := &encoder.Encoder{
		Buffer: data[:],
	}

	err := EncodeOmitEmptyMaxLenStruct1(e, &OmitEmptyMaxLenStruct1{
		Extra: []byte{1, 2, 3, 4},
	})
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("EncodeOmitEmptyMaxLenStruct1 expected encoder.ErrMaxLenExceeded")
	}

	err = EncodeOmitEmptyMaxLenStruct2(&encoder.Encoder{
		Buffer: data[:],
	}, &OmitEmptyMaxLenStruct2{
		Extra: []byte{1, 2, 3, 4},
	})
	if err != nil {
		t.Fatalf("EncodeOmitEmptyMaxLenStruct2 unexpected error: %v", err)
	}

	var obj1 OmitEmptyMaxLenStruct1
	err = DecodeOmitEmptyMaxLenStruct1(&encoder.Decoder{
		Buffer: data[:],
	}, &obj1)
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("DecodeOmitEmptyMaxLenStruct1 expected encoder.ErrMaxLenExceeded")
	}
}
