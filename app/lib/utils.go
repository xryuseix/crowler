package lib

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"
)

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

func ToAbsoluteLinks(base *url.URL, links []string) []string {
	r := []string{}
	for _, link := range links {
		l := ToAbsoluteLink(base, link)
		if l != "" {
			r = append(r, l)
		}
	}
	return r
}

func ToAbsoluteLink(base *url.URL, link string) string {
	if strings.HasPrefix(link, "http") || strings.HasPrefix(link, "https") {
		return link
	} else if strings.HasPrefix(link, "//") {
		return "https:" + link
	} else if strings.HasPrefix(link, "/") {
		return base.ResolveReference(&url.URL{Path: link}).String()
	}
	return ""
}

func Hash(s string) string {
	b := sha256.Sum256([]byte(s))
	return hex.EncodeToString(b[:])
}
