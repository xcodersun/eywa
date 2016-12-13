package utils

import (
	"github.com/unrolled/render"
)

var Render *render.Render

func init() {
	Render = render.New(render.Options{})
}
