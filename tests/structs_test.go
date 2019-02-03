package tests

import (
	"testing"

	"github.com/skycoin/skycoin/src/cipher/encoder"
)

func TestMaxLenStringStructExceeded(t *testing.T) {
	obj2 := MaxLenStringStruct2{
		Foo: "1234",
	}

	n := EncodeSizeMaxLenStringStruct2(&obj2)

	data := make([]byte, n)

	err := EncodeMaxLenStringStruct1(data[:], &MaxLenStringStruct1{
		Foo: "1234",
	})
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("EncodeMaxLenStringStruct1 expected encoder.ErrMaxLenExceeded")
	}

	err = EncodeMaxLenStringStruct2(data[:], &MaxLenStringStruct2{
		Foo: "1234",
	})
	if err != nil {
		t.Fatalf("EncodeMaxLenStringStruct2 unexpected error: %v", err)
	}

	var obj1 MaxLenStringStruct1
	err = DecodeMaxLenStringStruct1(data[:], &obj1)
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("DecodeMaxLenStringStruct1 expected encoder.ErrMaxLenExceeded")
	}
}

func testMaxLenAllStructExceeded(t *testing.T, obj1 MaxLenAllStruct1, obj2 MaxLenAllStruct2) {
	n := EncodeSizeMaxLenAllStruct2(&obj2)

	data := make([]byte, n)

	obj1Bad := MaxLenAllStruct1(obj2)
	err := EncodeMaxLenAllStruct1(data[:], &obj1Bad)
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("EncodeMaxLenAllStruct1 expected encoder.ErrMaxLenExceeded")
	}

	err = EncodeMaxLenAllStruct2(data[:], &obj2)
	if err != nil {
		t.Fatalf("EncodeMaxLenAllStruct2 unexpected error: %v", err)
	}

	var obj1Empty MaxLenAllStruct1
	err = DecodeMaxLenAllStruct1(data[:], &obj1Empty)
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

	err := EncodeMaxLenNestedSliceStruct1(data[:], &MaxLenNestedSliceStruct1{
		Foo: []MaxLenStringStruct1{{
			Foo: "1234",
		}},
	})
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("EncodeMaxLenNestedSliceStruct1 expected encoder.ErrMaxLenExceeded")
	}

	err = EncodeMaxLenNestedSliceStruct2(data[:], &MaxLenNestedSliceStruct2{
		Foo: []MaxLenStringStruct2{{
			Foo: "1234",
		}},
	})
	if err != nil {
		t.Fatalf("EncodeMaxLenStringStruct2 unexpected error: %v", err)
	}

	var obj1 MaxLenNestedSliceStruct1
	err = DecodeMaxLenNestedSliceStruct1(data[:], &obj1)
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

	err := EncodeMaxLenNestedMapKeyStruct1(data[:], &MaxLenNestedMapKeyStruct1{
		Foo: map[MaxLenStringStruct1]int64{
			{Foo: "1234"}: 1,
		},
	})
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("EncodeMaxLenNestedMapKeyStruct1 expected encoder.ErrMaxLenExceeded")
	}

	err = EncodeMaxLenNestedMapKeyStruct2(data[:], &MaxLenNestedMapKeyStruct2{
		Foo: map[MaxLenStringStruct2]int64{
			{Foo: "1234"}: 1,
		},
	})
	if err != nil {
		t.Fatalf("EncodeMaxLenStringStruct2 unexpected error: %v", err)
	}

	var obj1 MaxLenNestedMapKeyStruct1
	err = DecodeMaxLenNestedMapKeyStruct1(data[:], &obj1)
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

	err := EncodeMaxLenNestedMapValueStruct1(data[:], &MaxLenNestedMapValueStruct1{
		Foo: map[int64]MaxLenStringStruct1{
			1: {Foo: "1234"},
		},
	})
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("EncodeMaxLenNestedMapValueStruct1 expected encoder.ErrMaxLenExceeded")
	}

	err = EncodeMaxLenNestedMapValueStruct2(data[:], &MaxLenNestedMapValueStruct2{
		Foo: map[int64]MaxLenStringStruct2{
			1: {Foo: "1234"},
		},
	})
	if err != nil {
		t.Fatalf("EncodeMaxLenStringStruct2 unexpected error: %v", err)
	}

	var obj1 MaxLenNestedMapValueStruct1
	err = DecodeMaxLenNestedMapValueStruct1(data[:], &obj1)
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

	err := EncodeOmitEmptyMaxLenStruct1(data[:], &OmitEmptyMaxLenStruct1{
		Extra: []byte{1, 2, 3, 4},
	})
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("EncodeOmitEmptyMaxLenStruct1 expected encoder.ErrMaxLenExceeded")
	}

	err = EncodeOmitEmptyMaxLenStruct2(data[:], &OmitEmptyMaxLenStruct2{
		Extra: []byte{1, 2, 3, 4},
	})
	if err != nil {
		t.Fatalf("EncodeOmitEmptyMaxLenStruct2 unexpected error: %v", err)
	}

	var obj1 OmitEmptyMaxLenStruct1
	err = DecodeOmitEmptyMaxLenStruct1(data[:], &obj1)
	if err != encoder.ErrMaxLenExceeded {
		t.Fatal("DecodeOmitEmptyMaxLenStruct1 expected encoder.ErrMaxLenExceeded")
	}
}
