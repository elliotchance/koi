package koi

func Koi_Array_length(v ...V) V {
	return NewInt(len(v[0].V.([]V)))
}

func Koi_Array_append_(v ...V) V {
	newValues := append(v[0].V.([]V), NewInt(v[1].V.(int)))
	return NewArray(v[0].N, newValues...)
}

func NewArray(typ string, a ...V) V {
	return V{
		N: typ,
		V: a,
		// F: ArrayFields,
	}
}
