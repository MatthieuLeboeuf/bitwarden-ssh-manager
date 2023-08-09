package backend

import (
	"net/url"
	"runtime"
)

const (
	LoginOk              = 0
	LoginWaitOtp         = 1
	LoginInvalidOtp      = 2
	LoginError           = 3
	LoginToManyRequests  = 4
	LoginInvalidPassword = 5
)

func urlValues(pairs ...string) url.Values {
	if len(pairs)%2 != 0 {
		panic("pairs must be of even length")
	}
	vals := make(url.Values)
	for i := 0; i < len(pairs); i += 2 {
		vals.Set(pairs[i], pairs[i+1])
	}
	return vals
}

func deviceType() string {
	switch runtime.GOOS {
	case "linux":
		return "8" // Linux Desktop
	case "darwin":
		return "7" // MacOS Desktop
	case "windows":
		return "6" // Windows Desktop
	default:
		return "14" // Unknown Browser, since we don't have a better fallback
	}
}
