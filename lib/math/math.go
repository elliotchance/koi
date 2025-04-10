package math

import (
	"math"

	"github.com/elliotchance/koi/lib/koi"
)

func Sin_int(args ...koi.V) koi.V {
	return koi.Static_(math.Sin(float64(args[0].V.(int))))
}
