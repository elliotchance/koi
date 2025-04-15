package koi

var IntFields = map[string]M{}

func NewInt(v int) V {
	return V{
		N: "Int",
		V: v,
		F: IntFields,
	}
}
