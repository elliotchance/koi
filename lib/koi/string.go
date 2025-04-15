package koi

var StringFields = map[string]M{
	"length": func(v ...V) V {
		return NewInt(len(v[0].V.(string)))
	},
}

func NewString(v string) V {
	return V{
		N: "String",
		V: v,
		F: StringFields,
	}
}
