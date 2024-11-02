package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

var (
	GetDoFunc func(req *http.Request) (*http.Response, error)
)

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

func TestGetWebPage(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`OK`))
	}))

	defer func() { testServer.Close() }()

	tests := []struct {
		name         string
		url          string
		expectedBody string
	}{
		{
			name:         "test1",
			url:          testServer.URL,
			expectedBody: "OK",
		},
	}

	for _, tc := range tests {
		parser := NewParser(tc.url)
		t.Run(tc.name, func(t *testing.T) {
			GetDoFunc = func(*http.Request) (*http.Response, error) {
				return &http.Response{
					Body: io.NopCloser(strings.NewReader(tc.expectedBody)),
				}, nil
			}

			err := parser.GetWebPage(tc.url)
			if err != nil {
				t.Fatal(err)
			}

			if parser.html != tc.expectedBody {
				t.Errorf("expected %s; got %s", tc.expectedBody, parser.html)
			}
		})
	}
}

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name            string
		html            string
		url             string
		expectedLinks   []string
		expectedDomains []string
	}{
		{
			name: "test1",
			html: `<html><head><link href="http://example.com/style.css"></head><body><img src="http://example.com/image.jpg"><script src="http://example.com/script.js"></script></body></html>`,
			url:  "http://example.com",
			expectedLinks: []string{
				"http://example.com/style.css",
				"http://example.com/image.jpg",
				"http://example.com/script.js",
			},
			expectedDomains: []string{
				"http://example.com",
			},
		},
		{
			name: "test2",
			html: `<html><head><link href="/style.css"></head><body><img src="/image.jpg"><script src="/script.js"></script></body></html>`,
			url:  "http://example.com",
			expectedLinks: []string{
				"http://example.com/style.css",
				"http://example.com/image.jpg",
				"http://example.com/script.js",
			},
			expectedDomains: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewParser(tc.url)
			parser.Parse()

			if len(parser.links) != len(tc.expectedLinks) {
				t.Errorf("expected %d links; got %d", len(tc.expectedLinks), len(parser.links))
			}

			for i, link := range parser.links {
				if link != tc.expectedLinks[i] {
					t.Errorf("expected link %s; got %s", tc.expectedLinks[i], link)
				}
			}

			if len(parser.externalDomains) != len(tc.expectedDomains) {
				t.Errorf("expected %d external domains; got %d", len(tc.expectedDomains), len(parser.externalDomains))
			}

			for i, domain := range parser.externalDomains {
				if domain != tc.expectedDomains[i] {
					t.Errorf("expected domain %s; got %s", tc.expectedDomains[i], domain)
				}
			}
		})
	}
}
