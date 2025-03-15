package main

import "fmt"

type M func(...V) V

type V struct {
	N string
	V any
	F map[string]M
}

func (v V) String() string {
	return v.N
}

func (v V) C(method string) any {
	return v.F[method](v).V
}

func _static[T any](x T) V {
	return V{fmt.Sprintf("%T", x), x, nil}
}

func __static[T any](x T) M {
	return func(...V) V { return _static[T](x) }
}

func rect__area(args ...V) V {
	rect := args[0]
	return _static[float64](rect.C("width").(float64) * rect.C("height").(float64))
}

// func rect__perim(rect rect) float64 {
// 	return ((2 * rect.width(rect)) + (2 * rect.height(rect)))
// }

func circle__area(args ...V) V {
	circle := args[0]
	return _static[float64](3.14 * circle.C("radius").(float64) * circle.C("radius").(float64))
}

// func circle__perim(circle circle) float64 {
// 	return ((2 * 3.14) * circle.radius(circle))
// }

func measure_geometry(args ...V) {
	g := args[0]
	fmt.Println(g)
	fmt.Println(g.C("area").(float64))
	// fmt.Println(g.perim(g))
}

func detectCircle_geometry(args ...V) {
	g := args[0]
	if g.N == "circle" {
		fmt.Println(fmt.Sprintf("circle with radius %v", g.C("radius").(float64)))
	}
}

func main() {
	var r = V{"rect", nil, map[string]M{
		"width":  __static[float64](3),
		"height": __static[float64](4),
		"area":   rect__area,
	}}
	fmt.Println(r.C("area").(float64))
	var c = V{"circle", nil, map[string]M{
		"radius": __static[float64](5),
		"area":   circle__area,
	}}
	fmt.Println(c.C("area").(float64))

	measure_geometry(r)
	measure_geometry(c)
	detectCircle_geometry(r)
	detectCircle_geometry(c)
}
