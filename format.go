package skyencoder

import (
	"fmt"
)

func cast(typ, name string) string {
	return fmt.Sprintf("%s(%s)", typ, name)
}

// Options is parsed encoder struct tag options
type Options struct {
	OmitEmpty bool
	MaxLength uint64
}

/* Encode size */

func wrapEncodeSizeFunc(typeName, typePackageName, counterName, funcBody string, usePointerArg bool) []byte {
	fullTypeName := typeName
	if typePackageName != "" {
		fullTypeName = fmt.Sprintf("%s.%s", typePackageName, typeName)
	}

	pointerArg := ""
	if usePointerArg {
		pointerArg = "*"
	}

	return []byte(fmt.Sprintf(`
// EncodeSize%[1]s computes the size of an encoded object of type %[1]s
func EncodeSize%[1]s(obj %[5]s%[4]s) int {
	%[2]s := 0

	%[3]s

	return %[2]s
}
`, typeName, counterName, funcBody, fullTypeName, pointerArg))
}

func buildEncodeSizeBool(name, counterName string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		%[2]s++
	`, name, counterName)
}

func buildEncodeSizeUint8(name, counterName string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		%[2]s++
	`, name, counterName)
}

func buildEncodeSizeUint16(name, counterName string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		%[2]s += 2
	`, name, counterName)
}

func buildEncodeSizeUint32(name, counterName string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		%[2]s += 4
	`, name, counterName)
}

func buildEncodeSizeUint64(name, counterName string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		%[2]s += 8
	`, name, counterName)
}

func buildEncodeSizeInt8(name, counterName string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		%[2]s++
	`, name, counterName)
}

func buildEncodeSizeInt16(name, counterName string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		%[2]s += 2
	`, name, counterName)
}

func buildEncodeSizeInt32(name, counterName string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		%[2]s += 4
	`, name, counterName)
}

func buildEncodeSizeInt64(name, counterName string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		%[2]s += 8
	`, name, counterName)
}

func buildEncodeSizeFloat32(name, counterName string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		%[2]s += 4
	`, name, counterName)
}

func buildEncodeSizeFloat64(name, counterName string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		%[2]s += 8
	`, name, counterName)
}

func buildEncodeSizeString(name, counterName string, options *Options) string {
	body := fmt.Sprintf(`
	// %[1]s
	%[2]s += 4 + len(%[1]s)
	`, name, counterName)

	if options != nil && options.OmitEmpty {
		return fmt.Sprintf(`
		// omitempty
		if len(%[1]s) != 0 {
			%[2]s
		}
		`, name, body)
	}

	return body
}

func buildEncodeSizeByteArray(name, counterName string, length int64, options *Options) string {
	return fmt.Sprintf(`
	// %[1]s
	%[2]s += %[3]d
	`, name, counterName, length)
}

func buildEncodeSizeArray(name, counterName, nextCounterName, elemVarName, elemSection string, length int64, isDynamic bool, options *Options) string {
	if isDynamic {
		return fmt.Sprintf(`
		// %[1]s
		for _, %[2]s := range %[1]s {
			%[4]s := 0

			%[3]s

			%[5]s += %[4]s
		}
		`, name, elemVarName, elemSection, nextCounterName, counterName)
	}

	return fmt.Sprintf(`
	// %[1]s
	{
		%[5]s := 0

		%[4]s

		%[2]s += %[3]d * %[5]s
	}
	`, name, counterName, length, elemSection, nextCounterName)

	// return fmt.Sprintf(`
	// // %[1]s
	// i += %[3]d * func() int {
	// 	i := 0

	// 	%[2]s

	// 	return i
	// }()
	// `, name, elemSection, length)
}

func buildEncodeSizeByteSlice(name, counterName string, options *Options) string {
	body := fmt.Sprintf(`
	// %[1]s
	%[2]s += 4 + len(%[1]s)
	`, name, counterName)

	if options != nil && options.OmitEmpty {
		return fmt.Sprintf(`
		// omitempty
		if len(%[1]s) != 0 {
			%[2]s
		}
		`, name, body)
	}

	return body
}

func buildEncodeSizeSlice(name, counterName, nextCounterName, elemVarName, elemSection string, isDynamic bool, options *Options) string {
	var body string

	debugPrintf("BuildEncodeSizeSlice: counterName=%s\n", counterName)

	if isDynamic {
		body = fmt.Sprintf(`
		// %[1]s
		%[2]s += 4
		for _, %[3]s := range %[1]s {
			%[5]s := 0

			%[4]s

			%[2]s += %[5]s
		}
		`, name, counterName, elemVarName, elemSection, nextCounterName)
	} else {
		body = fmt.Sprintf(`
		// %[1]s
		%[2]s += 4
		{
			%[4]s := 0

			%[3]s

			%[2]s += len(%[1]s) * %[4]s
		}
		`, name, counterName, elemSection, nextCounterName)

		// body = fmt.Sprintf(`
		// // %[1]s
		// i += 4
		// i += len(%[1]s) * func() int {
		// 	i := 0

		// 	%[2]s

		// 	return i
		// }()`, name, elemSection)
	}

	if options != nil && options.OmitEmpty {
		return fmt.Sprintf(`
		// omitempty
		if len(%[1]s) != 0 {
			%[2]s
		}
		`, name, body)
	}

	return body
}

func buildEncodeSizeMap(name, counterName, nextCounterName, keyVarName, elemVarName, keySection, elemSection string, isDynamicKey, isDynamicElem bool, options *Options) string {
	var body string

	if isDynamicKey || isDynamicElem {
		if !isDynamicKey {
			keyVarName = "_"
		}

		if !isDynamicElem {
			elemVarName = "_"
		}

		body = fmt.Sprintf(`
		// %[1]s
		%[2]s += 4
		for %[3]s, %[4]s := range %[1]s {
			%[7]s := 0

			%[5]s

			%[6]s

			%[2]s += %[7]s
		}
		`, name, counterName, keyVarName, elemVarName, keySection, elemSection, nextCounterName)
	} else {
		body = fmt.Sprintf(`
		// %[1]s
		%[2]s += 4
		{
			%[5]s := 0

			%[3]s

			%[4]s

			%[2]s += len(%[1]s) * %[5]s
		}
		`, name, counterName, keySection, elemSection, nextCounterName)

		// body = fmt.Sprintf(`
		// // %[1]s
		// i += 4
		// i += len(%[1]s) * func() int {
		// 	i := 0

		// 	%[2]s

		// 	%[3]s

		// 	return i
		// }()
		// `, name, keySection, elemSection)
	}

	if options != nil && options.OmitEmpty {
		return fmt.Sprintf(`
		// omitempty
		if len(%[1]s) != 0 {
			%[2]s
		}
		`, name, body)
	}

	return body
}

/* Encode */

func wrapEncodeFunc(typeName, typePackageName, funcBody string, usePointerArg bool) []byte {
	fullTypeName := typeName
	if typePackageName != "" {
		fullTypeName = fmt.Sprintf("%s.%s", typePackageName, fullTypeName)
	}

	pointerArg := ""
	if usePointerArg {
		pointerArg = "*"
	}

	return []byte(fmt.Sprintf(`
// Encode%[1]s encodes an object of type %[1]s to the buffer in encoder.Encoder
func Encode%[1]s(e *encoder.Encoder, obj %[4]s%[3]s) error {
	%[2]s

	return nil
}
`, typeName, funcBody, fullTypeName, pointerArg))
}

func buildEncodeBool(name string, castType bool, options *Options) string {
	castName := name
	if castType {
		castName = cast("bool", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Bool(%[2]s)
	`, name, castName)
}

func buildEncodeUint8(name string, castType bool, options *Options) string {
	castName := name
	if castType {
		castName = cast("uint8", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Uint8(%[2]s)
	`, name, castName)
}

func buildEncodeUint16(name string, castType bool, options *Options) string {
	castName := name
	if castType {
		castName = cast("uint16", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Uint16(%[2]s)
	`, name, castName)
}

func buildEncodeUint32(name string, castType bool, options *Options) string {
	castName := name
	if castType {
		castName = cast("uint32", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Uint32(%[2]s)
	`, name, castName)
}

func buildEncodeUint64(name string, castType bool, options *Options) string {
	castName := name
	if castType {
		castName = cast("uint64", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Uint64(%[2]s)
	`, name, castName)
}

func buildEncodeInt8(name string, castType bool, options *Options) string {
	castName := name
	if castType {
		castName = cast("int8", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Int8(%[2]s)
	`, name, castName)
}

func buildEncodeInt16(name string, castType bool, options *Options) string {
	castName := name
	if castType {
		castName = cast("int16", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Int16(%[2]s)
	`, name, castName)
}

func buildEncodeInt32(name string, castType bool, options *Options) string {
	castName := name
	if castType {
		castName = cast("int32", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Int32(%[2]s)
	`, name, castName)
}

func buildEncodeInt64(name string, castType bool, options *Options) string {
	castName := name
	if castType {
		castName = cast("int64", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Int64(%[2]s)
	`, name, castName)
}

func buildEncodeFloat32(name string, castType bool, options *Options) string {
	castName := name
	if castType {
		castName = cast("float32", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Uint32(math.Float32bits(%[2]s))
	`, name, castName)
}

func buildEncodeFloat64(name string, castType bool, options *Options) string {
	castName := name
	if castType {
		castName = cast("float64", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Uint64(math.Float64bits(%[2]s))
	`, name, castName)
}

func buildEncodeString(name string, options *Options) string {
	body := fmt.Sprintf(`
	%[2]s

	// %[1]s length check
	if len(%[1]s) > math.MaxUint32 {
		return errors.New("%[1]s length exceeds math.MaxUint32")
	}

	// %[1]s
	e.ByteSlice([]byte(%[1]s))
	`, name, encodeMaxLengthCheck(name, options))

	if options != nil && options.OmitEmpty {
		return fmt.Sprintf(`
			// omitempty
			if len(%[1]s) != 0 {
				%[2]s
			}
		`, name, body)
	}

	return body
}

func buildEncodeByteArray(name string, options *Options) string {
	return fmt.Sprintf(`
	// %[1]s
	e.CopyBytes(%[1]s[:])
	`, name)
}

func buildEncodeArray(name, elemVarName, elemSection string, options *Options) string {
	return fmt.Sprintf(`
	// %[1]s
	for _, %[2]s := range %[1]s {
		%[3]s
	}
	`, name, elemVarName, elemSection)
}

func buildEncodeByteSlice(name string, options *Options) string {
	body := fmt.Sprintf(`
	%[2]s

	// %[1]s length check
	if len(%[1]s) > math.MaxUint32 {
		return errors.New("%[1]s length exceeds math.MaxUint32")
	}

	// %[1]s length
	e.Uint32(uint32(len(%[1]s)))

	// %[1]s copy
	e.CopyBytes(%[1]s)
	`, name, encodeMaxLengthCheck(name, options))

	if options != nil && options.OmitEmpty {
		return fmt.Sprintf(`
			// omitempty
			if len(%[1]s) != 0 {
				%[2]s
			}
		`, name, body)
	}

	return body
}

func buildEncodeSlice(name, elemVarName, elemSection string, options *Options) string {
	body := fmt.Sprintf(`
	%[4]s

	// %[1]s length check
	if len(%[1]s) > math.MaxUint32 {
		return errors.New("%[1]s length exceeds math.MaxUint32")
	}

	// %[1]s length
	e.Uint32(uint32(len(%[1]s)))

	// %[1]s
	for _, %[2]s := range %[1]s {
		%[3]s
	}
	`, name, elemVarName, elemSection, encodeMaxLengthCheck(name, options))

	if options != nil && options.OmitEmpty {
		return fmt.Sprintf(`
			// omitempty
			if len(%[1]s) != 0 {
				%[2]s
			}
		`, name, body)
	}

	return body
}

func buildEncodeMap(name, keyVarName, elemVarName, keySection, elemSection string, options *Options) string {
	if keySection == "" {
		keyVarName = "_"
	}
	if elemSection == "" {
		elemVarName = "_"
	}

	body := fmt.Sprintf(`
	// %[1]s

	%[6]s

	// %[1]s length check
	if len(%[1]s) > math.MaxUint32 {
		return errors.New("%[1]s length exceeds math.MaxUint32")
	}

	// %[1]s length
	e.Uint32(uint32(len(%[1]s)))

	for %[2]s, %[3]s := range %[1]s {
		%[4]s

		%[5]s
	}
	`, name, keyVarName, elemVarName, keySection, elemSection, encodeMaxLengthCheck(name, options))

	if options != nil && options.OmitEmpty {
		return fmt.Sprintf(`
			// omitempty
			if len(%[1]s) != 0 {
				%[2]s
			}
		`, name, body)
	}

	return body
}

func encodeMaxLengthCheck(name string, options *Options) string {
	if options != nil && options.MaxLength > 0 {
		return fmt.Sprintf(`
		// %[1]s maxlen check
		if len(%[1]s) > %d {
			return encoder.ErrMaxLenExceeded
		}
		`, name, options.MaxLength)
	}

	return ""
}

/* Decode */

func wrapDecodeFunc(typeName, typePackageName, funcBody string, usePointerArg bool) []byte {
	fullTypeName := typeName
	if typePackageName != "" {
		fullTypeName = fmt.Sprintf("%s.%s", typePackageName, typeName)
	}

	pointerArg := ""
	if usePointerArg {
		pointerArg = "*"
	}

	return []byte(fmt.Sprintf(`
// Decode%[1]s decodes an object of type %[1]s from the buffer in encoder.Decoder
func Decode%[1]s(d *encoder.Decoder, obj %[4]s%[3]s) error {
	%[2]s

	if len(d.Buffer) != 0 {
		return encoder.ErrRemainingBytes
	}

	return nil
}
`, typeName, funcBody, fullTypeName, pointerArg))
}

func buildDecodeBool(name string, castType bool, typeName string, options *Options) string {
	assign := "i"
	if castType {
		assign = cast(typeName, assign)
	}
	return fmt.Sprintf(`{
	// %[1]s
	i, err := d.Bool()
	if err != nil {
		return err
	}
	%[1]s = %[2]s
	}
	`, name, assign)
}

func buildDecodeUint8(name string, castType bool, typeName string, options *Options) string {
	assign := "i"
	if castType {
		assign = cast(typeName, assign)
	}
	return fmt.Sprintf(`{
	// %[1]s
	i, err := d.Uint8()
	if err != nil {
		return err
	}
	%[1]s = %[2]s
	}`, name, assign)
}

func buildDecodeUint16(name string, castType bool, typeName string, options *Options) string {
	assign := "i"
	if castType {
		assign = cast(typeName, assign)
	}
	return fmt.Sprintf(`{
	// %[1]s
	i, err := d.Uint16()
	if err != nil {
		return err
	}
	%[1]s = %[2]s
	}
	`, name, assign)
}

func buildDecodeUint32(name string, castType bool, typeName string, options *Options) string {
	assign := "i"
	if castType {
		assign = cast(typeName, assign)
	}
	return fmt.Sprintf(`{
	// %[1]s
	i, err := d.Uint32()
	if err != nil {
		return err
	}
	%[1]s = %[2]s
	}
	`, name, assign)
}

func buildDecodeUint64(name string, castType bool, typeName string, options *Options) string {
	assign := "i"
	if castType {
		assign = cast(typeName, assign)
	}
	return fmt.Sprintf(`{
	// %[1]s
	i, err := d.Uint64()
	if err != nil {
		return err
	}
	%[1]s = %[2]s
	}
	`, name, assign)
}

func buildDecodeInt8(name string, castType bool, typeName string, options *Options) string {
	assign := "i"
	if castType {
		assign = cast(typeName, assign)
	}
	return fmt.Sprintf(`{
	// %[1]s
	i, err := d.Int8()
	if err != nil {
		return err
	}
	%[1]s = %[2]s
	}
	`, name, assign)
}

func buildDecodeInt16(name string, castType bool, typeName string, options *Options) string {
	assign := "i"
	if castType {
		assign = cast(typeName, assign)
	}
	return fmt.Sprintf(`{
	// %[1]s
	i, err := d.Int16()
	if err != nil {
		return err
	}
	%[1]s = %[2]s
	}
	`, name, assign)
}

func buildDecodeInt32(name string, castType bool, typeName string, options *Options) string {
	assign := "i"
	if castType {
		assign = cast(typeName, assign)
	}
	return fmt.Sprintf(`{
	// %[1]s
	i, err := d.Int32()
	if err != nil {
		return err
	}
	%[1]s = %[2]s
	}
	`, name, assign)
}

func buildDecodeInt64(name string, castType bool, typeName string, options *Options) string {
	assign := "i"
	if castType {
		assign = cast(typeName, assign)
	}
	return fmt.Sprintf(`{
	// %[1]s
	i, err := d.Int64()
	if err != nil {
		return err
	}
	%[1]s = %[2]s
	}
	`, name, assign)
}

func buildDecodeFloat32(name string, castType bool, typeName string, options *Options) string {
	assign := "i"
	if castType {
		assign = cast(typeName, assign)
	}
	return fmt.Sprintf(`{
	// %[1]s
	i, err := d.Uint32()
	if err != nil {
		return err
	}
	%[1]s = math.Float32frombits(%[2]s)
	}
	`, name, assign)
}

func buildDecodeFloat64(name string, castType bool, typeName string, options *Options) string {
	assign := "i"
	if castType {
		assign = cast(typeName, assign)
	}
	return fmt.Sprintf(`{
	// %[1]s
	i, err := d.Uint64()
	if err != nil {
		return err
	}
	%[1]s = math.Float64frombits(%[2]s)
	}
	`, name, assign)
}

func buildDecodeString(name string, options *Options) string {
	return fmt.Sprintf(`{
	// %[1]s

	%[3]s

	ul, err := d.Uint32()
	if err != nil {
		return err
	}

	length := int(ul)
	if length < 0 || length > len(d.Buffer) {
		return encoder.ErrBufferUnderflow
	}

	%[2]s

	%[1]s = string(d.Buffer[:length])
	d.Buffer = d.Buffer[length:]
	}`, name, decodeMaxLengthCheck(options), decodeOmitEmptyCheck(options))
}

func buildDecodeByteArray(name string, options *Options) string {
	return fmt.Sprintf(`{
	// %[1]s
	if len(d.Buffer) < len(%[1]s) {
		return encoder.ErrBufferUnderflow
	}
	copy(%[1]s[:], d.Buffer[:len(%[1]s)])
	d.Buffer = d.Buffer[len(%[1]s):]
	}
	`, name)
}

func buildDecodeArray(name, elemCounterName, elemVarName, elemSection string, options *Options) string {
	return fmt.Sprintf(`{
	// %[1]s
	for %[2]s := range %[1]s {
		%[4]s
	}
	}
	`, name, elemCounterName, elemVarName, elemSection)
}

func buildDecodeByteSlice(name string, options *Options) string {
	return fmt.Sprintf(`{
	// %[1]s

	%[3]s

	ul, err := d.Uint32()
	if err != nil {
		return err
	}

	length := int(ul)
	if length < 0 || length > len(d.Buffer) {
		return encoder.ErrBufferUnderflow
	}

	%[2]s

	if length != 0 {
		%[1]s = make([]byte, length)

		copy(%[1]s[:], d.Buffer[:length])
		d.Buffer = d.Buffer[length:]
	}
	}`, name, decodeMaxLengthCheck(options), decodeOmitEmptyCheck(options))
}

func buildDecodeSlice(name, elemCounterName, elemVarName, elemSection, typeName string, options *Options) string {
	return fmt.Sprintf(`{
	// %[1]s

	%[7]s

	ul, err := d.Uint32()
	if err != nil {
		return err
	}

	length := int(ul)
	if length < 0 || length > len(d.Buffer) {
		return encoder.ErrBufferUnderflow
	}

	%[6]s

	if length != 0 {
		%[1]s = make(%[5]s, length)

		for %[2]s := range %[1]s {
			%[4]s
		}
	}
	}`, name, elemCounterName, elemVarName, elemSection, typeName, decodeMaxLengthCheck(options), decodeOmitEmptyCheck(options))
}

func buildDecodeMap(name, keyVarName, elemVarName, keySection, elemSection, typeName string, options *Options) string {
	return fmt.Sprintf(`{
	// %[1]s

	%[8]s

	ul, err := d.Uint32()
	if err != nil {
		return err
	}

	length := int(ul)
	if length < 0 || length > len(d.Buffer) {
		return encoder.ErrBufferUnderflow
	}

	%[7]s

	if length != 0 {
		%[1]s = make(%[6]s)

		for %[2]s, %[3]s := range %[1]s {
			%[4]s

			%[5]s

			%[1]s[%[2]s] = %[3]s
		}
	}
	}`, name, keyVarName, elemVarName, keySection, elemSection, typeName, decodeMaxLengthCheck(options), decodeOmitEmptyCheck(options))
}

func decodeMaxLengthCheck(options *Options) string {
	if options != nil && options.MaxLength > 0 {
		return fmt.Sprintf(`if length > %d {
			return encoder.ErrMaxLenExceeded
		}`, options.MaxLength)
	}

	return ""
}

func decodeOmitEmptyCheck(options *Options) string {
	if options != nil && options.OmitEmpty {
		return `if len(d.Buffer) == 0 {
			return nil
		}`
	}

	return ""
}

/* Test */

func buildTest(typeName, typePackageName, packageName string, usePointerArg bool) string {
	fullTypeName := typeName
	if typePackageName != "" {
		fullTypeName = fmt.Sprintf("%s.%s", typePackageName, typeName)
	}

	pointerArg := ""
	if usePointerArg {
		pointerArg = "*"
	}

	addressOp := ""
	if usePointerArg {
		addressOp = "&"
	}

	// TODO -- add a 2nd test for omitempty structs

	return fmt.Sprintf(`// Code generated by github.com/skycoin/skyencoder. DO NOT EDIT.
package %[5]s

import (
	"bytes"
	"fmt"
	mathrand "math/rand"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func newEmpty%[1]sForEncodeTest() %[3]s%[2]s {
	var obj %[2]s
	return %[4]sobj
}

func newRandom%[1]sForEncodeTest(t *testing.T, rand *mathrand.Rand) %[3]s%[2]s {
	var obj %[2]s
	err := encoder.PopulateRandom(&obj, rand, encoder.PopulateRandomOptions{
		MaxRandLen: 16,
		MinRandLen: 1,
	})
	if err != nil {
		t.Fatalf("encoder.PopulateRandom failed: %%v", err)
	}
	return %[4]sobj
}

func newRandomZeroLen%[1]sForEncodeTest(t *testing.T, rand *mathrand.Rand) %[3]s%[2]s {
	var obj %[2]s
	err := encoder.PopulateRandom(&obj, rand, encoder.PopulateRandomOptions{
		MaxRandLen:    0,
		MinRandLen:    0,
		EmptySliceNil: false,
		EmptyMapNil:   false,
	})
	if err != nil {
		t.Fatalf("encoder.PopulateRandom failed: %%v", err)
	}
	return %[4]sobj
}

func newRandomZeroLenNil%[1]sForEncodeTest(t *testing.T, rand *mathrand.Rand) %[3]s%[2]s {
	var obj %[2]s
	err := encoder.PopulateRandom(&obj, rand, encoder.PopulateRandomOptions{
		MaxRandLen:    0,
		MinRandLen:    0,
		EmptySliceNil: true,
		EmptyMapNil:   true,
	})
	if err != nil {
		t.Fatalf("encoder.PopulateRandom failed: %%v", err)
	}
	return %[4]sobj
}


func testSkyencoder%[1]s(t *testing.T, obj %[3]s%[2]s) {
	// EncodeSize

	n1, err := encoder.Size(obj)
	if err != nil {
		t.Fatalf("encoder.Size failed: %%v", err)
	}

	n2 := EncodeSize%[1]s(obj)

	if n1 != n2 {
		t.Fatalf("encoder.Size() != EncodeSize%[1]s() (%%d != %%d)", n1, n2)
	}

	// Encode

	data1 := encoder.Serialize(obj)

	data2 := make([]byte, n2)
	e := &encoder.Encoder{
		Buffer: data2[:],
	}

	err = Encode%[1]s(e, obj)
	if err != nil {
		t.Fatalf("Encode%[1]s failed: %%v", err)
	}

	if !bytes.Equal(data1, data2) {
		t.Fatal("encoder.Serialize() != Encode%[1]s()")
	}

	// Decode

	var obj2 %[2]s
	err = encoder.DeserializeRaw(data1, &obj2)
	if err != nil {
		t.Fatalf("encoder.DeserializeRaw failed: %%v", err)
	}


	if !cmp.Equal(%[3]sobj, obj2, cmpopts.EquateEmpty()) {
		t.Fatal("encoder.DeserializeRaw result wrong")
	}

	var obj3 %[2]s
	err = Decode%[1]s(&encoder.Decoder{
		Buffer: data2[:],
	}, &obj3)
	if err != nil {
		t.Fatalf("Decode%[1]s failed: %%v", err)
	}

	if !cmp.Equal(obj2, obj3, cmpopts.EquateEmpty()) {
		t.Fatal("encoder.DeserializeRaw() != Decode%[1]s()")
	}
}

func TestSkyencoder%[1]s(t *testing.T) {
	rand := mathrand.New(mathrand.NewSource(time.Now().Unix()))

	type testCase struct {
		name string
		obj  %[3]s%[2]s
	}

	cases := []testCase{
		{
			name: "empty object",
			obj:  newEmpty%[1]sForEncodeTest(),
		},
	}

	nRandom := 10

	for i := 0; i < nRandom; i++ {
		cases = append(cases, testCase{
			name: fmt.Sprintf("randomly populated object %%d", i),
			obj:  newRandom%[1]sForEncodeTest(t, rand),
		})
		cases = append(cases, testCase{
			name: fmt.Sprintf("randomly populated object %%d with zero length variable length contents", i),
			obj:  newRandomZeroLen%[1]sForEncodeTest(t, rand),
		})
		cases = append(cases, testCase{
			name: fmt.Sprintf("randomly populated object %%d with zero length variable length contents set to nil", i),
			obj:  newRandomZeroLenNil%[1]sForEncodeTest(t, rand),
		})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testSkyencoder%[1]s(t, tc.obj)
		})
	}
}

func decode%[1]sExpectError(t *testing.T, buf []byte, expectedErr error) {
	var obj %[2]s
	err := Decode%[1]s(&encoder.Decoder{
		Buffer: buf,
	}, &obj)

	if err == nil {
		// TODO -- how about omitempty?
		t.Fatal("Decode%[1]s: expected error, got nil")
	}

	if err != expectedErr {
		t.Fatalf("Decode%[1]s: expected encoder.ErrBufferUnderflow, got %%v", err)
	}
}

func testSkyencoder%[1]sDecodeErrors(t *testing.T, obj %[3]s%[2]s) {
	n := EncodeSize%[1]s(obj)
	buf := make([]byte, n)
	e := &encoder.Encoder{
		Buffer: buf[:],
	}

	err := Encode%[1]s(e, obj)
	if err != nil {
		t.Fatalf("Encode%[1]s failed: %%v", err)
	}

	t.Run("buffer underflow nil", func(t *testing.T) {
		decode%[1]sExpectError(t, nil, encoder.ErrBufferUnderflow)
	})

	for i := 0; i < n; i++ {

		t.Run(fmt.Sprintf("buffer underflow %%d", i), func(t *testing.T) {
			decode%[1]sExpectError(t, buf[:i], encoder.ErrBufferUnderflow)
		})
	}

	// Buffer too long (TODO: How about omitempty?)
	buf = append(buf, 0)
	t.Run("remaining bytes", func(t *testing.T) {
		decode%[1]sExpectError(t, buf[:], encoder.ErrRemainingBytes)
	})
}

func TestSkyencoder%[1]sDecodeErrors(t *testing.T) {
	rand := mathrand.New(mathrand.NewSource(time.Now().Unix()))

	emptyObj := newEmpty%[1]sForEncodeTest()
	fullObj := newRandom%[1]sForEncodeTest(t, rand)

	testSkyencoder%[1]sDecodeErrors(t, emptyObj)
	testSkyencoder%[1]sDecodeErrors(t, fullObj)
}

`, typeName, fullTypeName, pointerArg, addressOp, packageName)
}
