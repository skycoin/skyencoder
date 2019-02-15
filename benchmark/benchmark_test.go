package benchmark

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/cipher/encoder"
	"github.com/skycoin/skycoin/src/coin"
)

func newBenchmarkStruct() *BenchmarkStruct {
	return &BenchmarkStruct{
		Int64:       12345678,
		String:      "foo",
		StringSlice: []string{"foo", "bar", "baz"},
		StaticStructArray: [3]StaticStruct{
			{
				A: 4,
				B: 16,
			},
			{
				A: 128,
				B: 12312312312,
			},
			{
				A: 196,
				B: 112313122222,
			},
		},
		DynamicStructSlice: []DynamicStruct{
			{
				C: "foobar",
			},
			{
				C: "foobarbaz",
			},
		},
		ByteArray:    [3]uint8{1, 2, 3},
		ByteSlice:    []uint8{1, 2, 3},
		StringMaxLen: "baz",
	}
}

func TestEncodeSizeEqual(t *testing.T) {
	bs := newBenchmarkStruct()
	n1 := EncodeSizeBenchmarkStruct(bs)
	n2 := encoder.Size(bs)
	if n1 != uint64(n2) {
		t.Fatalf("Encode size does not match (%d != %d)", n1, n2)
	}
}

func BenchmarkEncodeSize(b *testing.B) {
	bs := newBenchmarkStruct()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		EncodeSizeBenchmarkStruct(bs)
	}
}

func BenchmarkCipherEncodeSize(b *testing.B) {
	bs := newBenchmarkStruct()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		encoder.Size(bs)
	}
}

func TestEncodeEqual(t *testing.T) {
	bs := newBenchmarkStruct()

	data1 := make([]byte, EncodeSizeBenchmarkStruct(bs))

	if err := EncodeBenchmarkStruct(data1, bs); err != nil {
		t.Fatal(err)
	}

	data2 := encoder.Serialize(bs)

	if !bytes.Equal(data1, data2) {
		t.Fatal("EncodeBenchmarkStruct() != encoder.Serialize()")
	}
}

func BenchmarkEncode(b *testing.B) {
	bs := newBenchmarkStruct()

	n := EncodeSizeBenchmarkStruct(bs)

	buf := make([]byte, n)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		EncodeBenchmarkStruct(buf[:], bs)
	}
}

func BenchmarkEncodeSizePlusEncode(b *testing.B) {
	// Performs EncodeSize + Encode to better mimic encoder.Serialize which will do a size calculation internally
	bs := newBenchmarkStruct()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n := EncodeSizeBenchmarkStruct(bs)
		buf := make([]byte, n)
		EncodeBenchmarkStruct(buf[:], bs)
	}
}

func BenchmarkCipherEncode(b *testing.B) {
	bs := newBenchmarkStruct()

	b.ResetTimer()

	// Note: encoder.Serialize also calls datasizeWrite internally,
	// which has an extra reflect recursion over the struct,
	// so the comparison is not totally fair.
	// Compare to BenchmarkEncodeSizePlusEncode for a total comparison

	for i := 0; i < b.N; i++ {
		encoder.Serialize(bs)
	}
}

func TestDecodeEqual(t *testing.T) {
	bs := newBenchmarkStruct()

	data := encoder.Serialize(bs)

	data1 := make([]byte, len(data))
	copy(data1[:], data[:])

	data2 := make([]byte, len(data))
	copy(data2[:], data[:])

	var bs1 BenchmarkStruct
	if n, err := DecodeBenchmarkStruct(data1, &bs1); err != nil {
		t.Fatal(err)
	} else if n != len(data1) {
		t.Fatalf("DecodeBenchmarkStruct n should be %d, is %d", len(data1), n)
	}

	var bs2 BenchmarkStruct
	if n, err := encoder.DeserializeRaw(data2, &bs2); err != nil {
		t.Fatal(err)
	} else if n != len(data2) {
		t.Fatal(encoder.ErrRemainingBytes)
	}

	if !reflect.DeepEqual(*bs, bs2) {
		t.Fatal("encoder.DeserializeRaw incorrect result")
	}

	t.Logf("newBenchmarkStruct ByteSlice: %v", bs.ByteSlice)
	t.Logf("encoder.DeserializeRaw ByteSlice: %v", bs2.ByteSlice)
	t.Logf("DecodeBenchmarkStruct ByteSlice: %v", bs1.ByteSlice)

	if !reflect.DeepEqual(*bs, bs1) {
		t.Fatal("DecodeBenchmarkStruct incorrect result")
	}

	if !reflect.DeepEqual(bs1, bs2) {
		compareBenchmarkStruct(t, bs1, bs2)
		t.Fatal("DecodeBenchmarkStruct() != encoder.DeserializeRaw()")
	}
}

func compareBenchmarkStruct(t *testing.T, a, b BenchmarkStruct) {
	if a.Int64 != b.Int64 {
		t.Logf("Int64 mismatch")
	}

	if a.String != b.String {
		t.Logf("String mismatch")
	}

	if len(a.StringSlice) != len(b.StringSlice) {
		t.Logf("StringSlice len mismatch")
	}

	for i := range a.StringSlice {
		if a.StringSlice[i] != b.StringSlice[i] {
			t.Logf("StringSlice[%d] mismatch", i)
		}
	}

	for i := range a.StaticStructArray {
		if a.StaticStructArray[i].A != b.StaticStructArray[i].A {
			t.Logf("StaticStructArray[%d].A mismatch", i)
		}

		if a.StaticStructArray[i].B != b.StaticStructArray[i].B {
			t.Logf("StaticStructArray[%d].B mismatch", i)
		}
	}

	if len(a.DynamicStructSlice) != len(b.DynamicStructSlice) {
		t.Logf("DynamicStructSlice len mismatch")
	}

	for i := range a.DynamicStructSlice {
		if a.DynamicStructSlice[i].C != b.DynamicStructSlice[i].C {
			t.Logf("DynamicStructSlice[%d].C mismatch", i)
		}
	}

	if a.ByteArray != b.ByteArray {
		t.Logf("ByteArray mismatch")
	}

	if len(a.ByteSlice) != len(b.ByteSlice) {
		t.Logf("ByteSlice len mismatch")
	}

	for i := range a.ByteSlice {
		if a.ByteSlice[i] != b.ByteSlice[i] {
			t.Logf("ByteSlice[%d] mismatch", i)
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	bs := newBenchmarkStruct()
	data := encoder.Serialize(bs)
	var bs1 BenchmarkStruct

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data2 := data
		DecodeBenchmarkStruct(data2, &bs1)
	}
}

func BenchmarkCipherDecode(b *testing.B) {
	bs := newBenchmarkStruct()
	data := encoder.Serialize(bs)
	var bs1 BenchmarkStruct

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data2 := data
		encoder.DeserializeRaw(data2, &bs1)
	}
}

func newSignedBlock() *coin.SignedBlock {
	return &coin.SignedBlock{
		Block: coin.Block{
			Body: coin.BlockBody{
				Transactions: coin.Transactions{
					{
						Sigs: make([]cipher.Sig, 3),
						In:   make([]cipher.SHA256, 3),
						Out:  make([]coin.TransactionOutput, 3),
					},
					{
						Sigs: make([]cipher.Sig, 3),
						In:   make([]cipher.SHA256, 3),
						Out:  make([]coin.TransactionOutput, 3),
					},
					{
						Sigs: make([]cipher.Sig, 3),
						In:   make([]cipher.SHA256, 3),
						Out:  make([]coin.TransactionOutput, 3),
					},
				},
			},
		},
	}
}

func BenchmarkEncodeSizeSignedBlock(b *testing.B) {
	bs := newSignedBlock()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		EncodeSizeSignedBlock(bs)
	}
}

func BenchmarkCipherEncodeSizeSignedBlock(b *testing.B) {
	bs := newSignedBlock()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		encoder.Size(bs)
	}
}

func BenchmarkEncodeSignedBlock(b *testing.B) {
	bs := newSignedBlock()

	n := EncodeSizeSignedBlock(bs)

	buf := make([]byte, n)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		EncodeSignedBlock(buf[:], bs)
	}
}

func BenchmarkEncodeSizePlusEncodeSignedBlock(b *testing.B) {
	// Performs EncodeSize + Encode to better mimic encoder.Serialize which will do a size calculation internally
	bs := newSignedBlock()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n := EncodeSizeSignedBlock(bs)
		buf := make([]byte, n)
		EncodeSignedBlock(buf[:], bs)
	}
}

func BenchmarkCipherEncodeSignedBlock(b *testing.B) {
	bs := newSignedBlock()

	b.ResetTimer()

	// Note: encoder.Serialize also calls datasizeWrite internally,
	// which has an extra reflect recursion over the struct,
	// so the comparison is not totally fair.
	// Compare to BenchmarkEncodeSizePlusEncode for a total comparison

	for i := 0; i < b.N; i++ {
		encoder.Serialize(bs)
	}
}

func BenchmarkDecodeSignedBlock(b *testing.B) {
	bs := newSignedBlock()
	data := encoder.Serialize(bs)
	var bs1 coin.SignedBlock

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data2 := data
		DecodeSignedBlock(data2, &bs1)
	}
}

func BenchmarkCipherDecodeSignedBlock(b *testing.B) {
	bs := newSignedBlock()
	data := encoder.Serialize(bs)
	var bs1 coin.SignedBlock

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data2 := data
		encoder.DeserializeRaw(data2, &bs1)
	}
}
