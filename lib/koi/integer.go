package koi

var IntegerFields = map[string]M{}

func NewInteger(v int) V {
	return V{
		N: "Integer",
		V: v,
		F: IntegerFields,
	}
}
