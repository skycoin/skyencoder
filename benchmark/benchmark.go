package benchmark

type StaticStruct struct {
	A byte
	B uint64
}

type DynamicStruct struct {
	C string
}

type BenchmarkStruct struct {
	Int64              int64
	String             string
	StringSlice        []string
	StaticStructArray  [3]StaticStruct
	DynamicStructSlice []DynamicStruct
	ByteArray          [3]uint8
	ByteSlice          []uint8
	StringMaxLen       string `enc:",maxlen=4"`
}
