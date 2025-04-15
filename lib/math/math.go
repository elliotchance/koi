package math

import (
	"math"

	"github.com/elliotchance/koi/lib/koi"
)

func Koi_Sin(args ...koi.V) koi.V {
	return koi.NewFloat(math.Sin(float64(args[0].V.(int))))
}
