package platform

import (
	"runtime"
)

type Info struct {
	OS   string
	Arch string
}

func Detect() Info {
	return Info{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}
