package koi

var FloatFields = map[string]M{}

func NewFloat(v float64) V {
	return V{
		N: "Float",
		V: v,
		F: FloatFields,
	}
}
