package lib

import "strings"

func Filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func SplitBySpace(s []string) []string {
	r := []string{}
	for _, v := range s {
		r = append(r, strings.Split(v, " ")...)
	}
	return r
}

func Unique[T comparable](vs []T) []T {
	vsm := make(map[T]struct{})
	for _, v := range vs {
		vsm[v] = struct{}{}
	}
	r := make([]T, 0, len(vsm))
	for v := range vsm {
		r = append(r, v)
	}
	return r
}
