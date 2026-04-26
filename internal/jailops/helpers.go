//go:build freebsd

package jailops

import (
	"github.com/zombocoder/go-freebsd-jail/internal/c"
	"github.com/zombocoder/go-freebsd-jail/internal/param"
)

// paramsToCParams converts internal params to C layer params.
func paramsToCParams(params []param.Param) []c.Param {
	cparams := make([]c.Param, len(params))
	for i, p := range params {
		cparams[i] = c.Param{
			Name:   p.Name,
			Value:  p.Value,
			IsBool: p.IsBool,
		}
	}
	return cparams
}
