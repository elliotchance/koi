package koi

func Koi_int_float64(args ...V) V {
	return Static_(float64(args[0].V.(int)))
}
