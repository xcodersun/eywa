package utils

import (
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/unrolled/render"
)

var Render *render.Render

func init() {
	Render = render.New(render.Options{})
}
