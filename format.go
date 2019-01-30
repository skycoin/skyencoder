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

func WrapEncodeSizeFunc(structName, funcBody string) []byte {
	return []byte(fmt.Sprintf(`
// EncodeSize%[1]s computes the size of an encoded object of type %[1]s
func EncodeSize%[1]s(obj *%[1]s) int {
	i := 0

	%[2]s

	return i
}
`, structName, funcBody))
}

func BuildEncodeSizeBool(name string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		i++
	`, name)
}

func BuildEncodeSizeUint8(name string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		i++
	`, name)
}

func BuildEncodeSizeUint16(name string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		i += 2
	`, name)
}

func BuildEncodeSizeUint32(name string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		i += 4
	`, name)
}

func BuildEncodeSizeUint64(name string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		i += 8
	`, name)
}

func BuildEncodeSizeInt8(name string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		i++
	`, name)
}

func BuildEncodeSizeInt16(name string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		i += 2
	`, name)
}

func BuildEncodeSizeInt32(name string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		i += 4
	`, name)
}

func BuildEncodeSizeInt64(name string, options *Options) string {
	return fmt.Sprintf(`
		// %[1]s
		i += 8
	`, name)
}

func BuildEncodeSizeString(name string, options *Options) string {
	body := fmt.Sprintf(`
	// %[1]s
	i += 4 + len(%[1]s)
	`, name)

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

func BuildEncodeSizeByteArray(name string, length int64, options *Options) string {
	return fmt.Sprintf(`
	// %[1]s
	i += %[2]d
	`, name, length)
}

func BuildEncodeSizeArray(name, elemVarName, elemSection string, length int64, isDynamic bool, options *Options) string {
	if isDynamic {
		return fmt.Sprintf(`
		// %[1]s
		for _, %[2]s = range %[1]s {
			%[3]s
		}
		`, name, elemVarName, elemSection)
	}

	return fmt.Sprintf(`
	// %[1]s
	i += %[3]d * func() int {
		i := 0

		%[2]s

		return i
	}()
	`, name, elemSection, length)
}

func BuildEncodeSizeByteSlice(name string, options *Options) string {
	body := fmt.Sprintf(`
	// %[1]s
	i += 4 + len(%[1]s)
	`, name)

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

func BuildEncodeSizeSlice(name, elemVarName, elemSection string, isDynamic bool, options *Options) string {
	var body string

	if isDynamic {
		body = fmt.Sprintf(`
		// %[1]s
		i += 4
		for _, %[2]s := range %[1]s {
			%[3]s
		}
		`, name, elemVarName, elemSection)
	} else {
		body = fmt.Sprintf(`
		// %[1]s
		i += 4
		i += len(%[1]s) * func() int {
			i := 0

			%[2]s

			return i
		}()`, name, elemSection)
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

func BuildEncodeSizeMap(name, keyVarName, elemVarName, keySection, elemSection string, isDynamicKey, isDynamicElem bool, options *Options) string {
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
		i += 4
		for %[2]s, %[3]s := range %[1]s {
			%[4]s

			%[5]s
		}
		`, name, keyVarName, elemVarName, keySection, elemSection)
	} else {
		body = fmt.Sprintf(`
		// %[1]s
		i += 4
		i += len(%[1]s) * func() int {
			i := 0

			%[2]s

			%[3]s

			return i
		}()
		`, name, keySection, elemSection)
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

func WrapEncodeFunc(structName, funcBody string) []byte {
	return []byte(fmt.Sprintf(`
// Encode%[1]s encodes an object of type %[1]s to the buffer in encoder.Encoder
func Encode%[1]s(e *encoder.Encoder, obj *%[1]s) error {
	%[2]s

	return nil
}
`, structName, funcBody))
}

func BuildEncodeBool(name string, castType bool, options *Options) string {
	if castType {
		name = cast("bool", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Bool(%[1]s)
	`, name)
}

func BuildEncodeUint8(name string, castType bool, options *Options) string {
	if castType {
		name = cast("uint8", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Uint8(%[1]s)
	`, name)
}

func BuildEncodeUint16(name string, castType bool, options *Options) string {
	if castType {
		name = cast("uint16", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Uint16(%[1]s)
	`, name)
}

func BuildEncodeUint32(name string, castType bool, options *Options) string {
	if castType {
		name = cast("uint32", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Uint32(%[1]s)
	`, name)
}

func BuildEncodeUint64(name string, castType bool, options *Options) string {
	if castType {
		name = cast("uint64", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Uint64(%[1]s)
	`, name)
}

func BuildEncodeInt8(name string, castType bool, options *Options) string {
	if castType {
		name = cast("int8", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Int8(%[1]s)
	`, name)
}

func BuildEncodeInt16(name string, castType bool, options *Options) string {
	if castType {
		name = cast("int16", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Int16(%[1]s)
	`, name)
}

func BuildEncodeInt32(name string, castType bool, options *Options) string {
	if castType {
		name = cast("int32", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Int32(%[1]s)
	`, name)
}

func BuildEncodeInt64(name string, castType bool, options *Options) string {
	if castType {
		name = cast("int64", name)
	}
	return fmt.Sprintf(`
	// %[1]s
	e.Int64(%[1]s)
	`, name)
}

func BuildEncodeString(name string, options *Options) string {
	body := fmt.Sprintf(`
	%[2]s

	// %[1]s
	e.Bytes([]byte(%[1]s))
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

func BuildEncodeByteArray(name string, options *Options) string {
	return fmt.Sprintf(`
	// %[1]s
	copy(e.Buffer[:], %[1]s[:])
	e.Buffer = e.Buffer[len(%[1]s):]
	`, name)
}

func BuildEncodeArray(name, elemVarName, elemSection string, options *Options) string {
	return fmt.Sprintf(`
	// %[1]s
	for _, %[2]s := range %[1]s {
		%[3]s
	}
	`, name, elemVarName, elemSection)
}

func BuildEncodeByteSlice(name string, options *Options) string {
	body := fmt.Sprintf(`
	%[2]s

	// %[1]s length check
	if len(%[1]s) > math.MaxUint32 {
		return errors.New("%[1]s length exceeds math.MaxUint32")
	}

	// %[1]s length
	e.Uint32(uint32(len(%[1]s)))

	// %[1]s copy
	copy(e.Buffer[:], %[1]s[:])
	e.Buffer = e.Buffer[len(%[1]s):]
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

func BuildEncodeSlice(name, elemVarName, elemSection string, options *Options) string {
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

func BuildEncodeMap(name, keyVarName, elemVarName, keySection, elemSection string, options *Options) string {
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

func WrapDecodeFunc(structName, funcBody string) []byte {
	return []byte(fmt.Sprintf(`
// Decode%[1]s decodes an object of type %[1]s from the buffer in encoder.Decoder
func Decode%[1]s(d *encoder.Decoder, obj *%[1]s) error {
	%[2]s

	if len(d.Buffer) != 0 {
		return encoder.ErrRemainingBytes
	}

	return nil
}
`, structName, funcBody))
}

func BuildDecodeBool(name string, castType bool, typeName string, options *Options) string {
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

func BuildDecodeUint8(name string, castType bool, typeName string, options *Options) string {
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

func BuildDecodeUint16(name string, castType bool, typeName string, options *Options) string {
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

func BuildDecodeUint32(name string, castType bool, typeName string, options *Options) string {
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

func BuildDecodeUint64(name string, castType bool, typeName string, options *Options) string {
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

func BuildDecodeInt8(name string, castType bool, typeName string, options *Options) string {
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

func BuildDecodeInt16(name string, castType bool, typeName string, options *Options) string {
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

func BuildDecodeInt32(name string, castType bool, typeName string, options *Options) string {
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

func BuildDecodeInt64(name string, castType bool, typeName string, options *Options) string {
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

func BuildDecodeByteArray(name string, options *Options) string {
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

func BuildDecodeArray(name, elemVarName, elemSection string, options *Options) string {
	return fmt.Sprintf(`{
	// %[1]s
	for _, %[2]s := range %[1]s {
		%[3]s
	}
	}
	`, name, elemVarName, elemSection)
}

func BuildDecodeByteSlice(name string, options *Options) string {
	return fmt.Sprintf(`{
	// %[1]s

	%[3]s

	if len(d.Buffer) < 4 {
		return encoder.ErrBufferUnderflow
	}

	ul, err := d.Uint32()
	if err != nil {
		return err
	}

	length := int(ul)
	if length < 0 || length > len(d.Buffer) {
		return encoder.ErrBufferUnderflow
	}

	%[2]s

	copy(%[1]s[:], d.Buffer[:length])
	d.Buffer = d.Buffer[length:]
	}`, name, decodeMaxLengthCheck(options), decodeOmitEmptyCheck(options))
}

func BuildDecodeSlice(name, elemVarName, elemSection, typeName string, options *Options) string {
	return fmt.Sprintf(`{
	// %[1]s

	%[6]s

	if len(d.Buffer) < 4 {
		return encoder.ErrBufferUnderflow
	}

	ul, err := d.Uint32()
	if err != nil {
		return err
	}

	length := int(ul)
	if length < 0 || length > len(d.Buffer) {
		return encoder.ErrBufferUnderflow
	}

	%[5]s

	%[1]s = make(%[4]s, length)

	for _, %[2]s := range %[1]s {
		%[3]s
	}

	}`, name, elemVarName, elemSection, typeName, decodeMaxLengthCheck(options), decodeOmitEmptyCheck(options))
}

func BuildDecodeString(name string, options *Options) string {
	return fmt.Sprintf(`{
	// %[1]s

	%[3]s

	if len(d.Buffer) < 4 {
		return encoder.ErrBufferUnderflow
	}

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

func BuildDecodeMap(name, keyVarName, elemVarName, keySection, elemSection, typeName string, options *Options) string {
	return fmt.Sprintf(`{
	// %[1]s

	%[8]s

	if len(d.Buffer) < 4 {
		return encoder.ErrBufferUnderflow
	}

	ul, err := d.Uint32()
	if err != nil {
		return err
	}

	length := int(ul)
	if length < 0 || length > len(d.Buffer) {
		return encoder.ErrBufferUnderflow
	}

	%[7]s

	%[1]s = make(%[6]s)

	for i := 0; i < length; i++ {
		%[4]s

		%[5]s

		%[1]s[%[2]s] = %[3]s
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
