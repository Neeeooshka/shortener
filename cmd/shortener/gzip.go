package main

import (
	"compress/gzip"
	"io"
	"net/http"
)

type gzipCompressor struct {
	encoding string
}

func (c *gzipCompressor) NewWriter(w http.ResponseWriter) io.WriteCloser {
	return gzip.NewWriter(w)
}

func (c *gzipCompressor) NewReader(r io.ReadCloser) (io.ReadCloser, error) {
	return gzip.NewReader(r)
}

func (c *gzipCompressor) GetEncoding() string {
	return c.encoding
}

func newGzipCompressor() *gzipCompressor {
	return &gzipCompressor{
		encoding: "gzip",
	}
}
