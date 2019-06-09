package tests

import (
	"errors"
	"math"

	"github.com/skycoin/skycoin/src/cipher/encoder"
)

// Code generated by github.com/skycoin/skyencoder. DO NOT EDIT.

// EncodeSizeDemoStructNestedBytes computes the size of an encoded object of type DemoStructNestedBytes
func EncodeSizeDemoStructNestedBytes(obj *DemoStructNestedBytes) uint64 {
	i0 := uint64(0)

	// obj.Objects
	i0 += 4
	for _, x1 := range obj.Objects {
		i1 := uint64(0)

		// x1.Data
		i1 += 4 + uint64(len(x1.Data))

		i0 += i1
	}

	return i0
}

// EncodeDemoStructNestedBytes encodes an object of type DemoStructNestedBytes to a buffer allocated to the exact size
// required to encode the object.
func EncodeDemoStructNestedBytes(obj *DemoStructNestedBytes) ([]byte, error) {
	n := EncodeSizeDemoStructNestedBytes(obj)
	buf := make([]byte, n)

	if err := EncodeDemoStructNestedBytesToBuffer(buf, obj); err != nil {
		return nil, err
	}

	return buf, nil
}

// EncodeDemoStructNestedBytesToBuffer encodes an object of type DemoStructNestedBytes to a []byte buffer.
// The buffer must be large enough to encode the object, otherwise an error is returned.
func EncodeDemoStructNestedBytesToBuffer(buf []byte, obj *DemoStructNestedBytes) error {
	if uint64(len(buf)) < EncodeSizeDemoStructNestedBytes(obj) {
		return encoder.ErrBufferUnderflow
	}

	e := &encoder.Encoder{
		Buffer: buf[:],
	}

	// obj.Objects length check
	if uint64(len(obj.Objects)) > math.MaxUint32 {
		return errors.New("obj.Objects length exceeds math.MaxUint32")
	}

	// obj.Objects length
	e.Uint32(uint32(len(obj.Objects)))

	// obj.Objects
	for _, x := range obj.Objects {

		// x.Data length check
		if uint64(len(x.Data)) > math.MaxUint32 {
			return errors.New("x.Data length exceeds math.MaxUint32")
		}

		// x.Data length
		e.Uint32(uint32(len(x.Data)))

		// x.Data copy
		e.CopyBytes(x.Data)

	}

	return nil
}

// DecodeDemoStructNestedBytes decodes an object of type DemoStructNestedBytes from a buffer.
// Returns the number of bytes used from the buffer to decode the object.
// If the buffer not long enough to decode the object, returns encoder.ErrBufferUnderflow.
func DecodeDemoStructNestedBytes(buf []byte, obj *DemoStructNestedBytes) (uint64, error) {
	d := &encoder.Decoder{
		Buffer: buf[:],
	}

	{
		// obj.Objects

		ul, err := d.Uint32()
		if err != nil {
			return 0, err
		}

		length := int(ul)
		if length < 0 || length > len(d.Buffer) {
			return 0, encoder.ErrBufferUnderflow
		}

		if length != 0 {
			obj.Objects = make([]DemoStructNestedBytesInner, length)

			for z1 := range obj.Objects {
				{
					// obj.Objects[z1].Data

					ul, err := d.Uint32()
					if err != nil {
						return 0, err
					}

					length := int(ul)
					if length < 0 || length > len(d.Buffer) {
						return 0, encoder.ErrBufferUnderflow
					}

					if length != 0 {
						obj.Objects[z1].Data = make([]byte, length)

						copy(obj.Objects[z1].Data[:], d.Buffer[:length])
						d.Buffer = d.Buffer[length:]
					}
				}
			}
		}
	}

	return uint64(len(buf) - len(d.Buffer)), nil
}

// DecodeDemoStructNestedBytesExact decodes an object of type DemoStructNestedBytes from a buffer.
// If the buffer not long enough to decode the object, returns encoder.ErrBufferUnderflow.
// If the buffer is longer than required to decode the object, returns encoder.ErrRemainingBytes.
func DecodeDemoStructNestedBytesExact(buf []byte, obj *DemoStructNestedBytes) error {
	if n, err := DecodeDemoStructNestedBytes(buf, obj); err != nil {
		return err
	} else if n != uint64(len(buf)) {
		return encoder.ErrRemainingBytes
	}

	return nil
}