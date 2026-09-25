package metadata

import (
	"strconv"
	"strings"
)

// CompareVersions compares two dot-separated version strings numerically
// (e.g. "4.10" > "4.9", not lexicographically). A trailing "-mono" suffix
// is stripped before comparing. Non-numeric segments parse as 0 rather
// than causing an error or panic. Missing trailing segments on the
// shorter string are treated as 0, so "4.3" == "4.3.0".
//
// It returns a positive number if a is newer than b, a negative number
// if a is older than b, and zero if they are equal.
func CompareVersions(a, b string) int {
	as := versionSegments(a)
	bs := versionSegments(b)

	n := len(as)
	if len(bs) > n {
		n = len(bs)
	}
	for i := 0; i < n; i++ {
		var av, bv int
		if i < len(as) {
			av = as[i]
		}
		if i < len(bs) {
			bv = bs[i]
		}
		if av != bv {
			return av - bv
		}
	}
	return 0
}

// versionSegments splits a version string (with any trailing "-mono"
// suffix stripped) into its dot-separated numeric segments. Segments
// that fail to parse as integers become 0.
func versionSegments(v string) []int {
	v = strings.TrimSuffix(v, "-mono")
	parts := strings.Split(v, ".")
	segs := make([]int, len(parts))
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			n = 0
		}
		segs[i] = n
	}
	return segs
}
