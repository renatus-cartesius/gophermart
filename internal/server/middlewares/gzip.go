package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/renatus-cartesius/gophermart/pkg/logger"
	"go.uber.org/zap"
)

type gzipWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (gw *gzipWriter) Write(b []byte) (int, error) {

	return gw.zw.Write(b)
}

func (gw *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		gw.w.Header().Set("Content-Encoding", "gzip")
	}
	gw.w.WriteHeader(statusCode)
}

func (gw *gzipWriter) Header() http.Header {
	return gw.w.Header()
}

func (gw *gzipWriter) Close() error {
	return gw.zw.Close()
}

type gzipReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		r:  r,
		zr: zr,
	}, nil
}

func (gr gzipReader) Read(p []byte) (n int, err error) {
	return gr.zr.Read(p)
}

func (gr *gzipReader) Close() error {
	if err := gr.r.Close(); err != nil {
		return err
	}
	return gr.zr.Close()
}

func Gzipper(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ow := w

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gw := newGzipWriter(w)
			ow = gw

			w.Header().Set("Content-Encoding", "gzip")

			defer gw.Close()
		}

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gr, err := newGzipReader(r.Body)
			if err != nil {
				logger.Log.Error(
					"error on creating new gzip reader",
					zap.Error(err),
				)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = gr
			defer gr.Close()
		}

		h.ServeHTTP(ow, r)

	})

}
