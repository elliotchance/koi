package io

import (
	"fmt"

	"github.com/elliotchance/koi/lib/koi"
)

func Koi_PrintLine_(args ...koi.V) koi.V {
	fmt.Println(args[0])
	return koi.V{}
}
