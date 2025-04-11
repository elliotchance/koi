package koi

import "fmt"

type M func(...V) V

type V struct {
	N string
	V any
	F map[string]M
}

func (v V) String() string {
	return fmt.Sprintf("%v", v.V)
}

func (v V) C(method string) V {
	return v.F[method](v)
}

// func Static__[T any](x T) M {
// 	return func(...V) V { return Static_(x) }
// }
