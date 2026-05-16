package middleware

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

type bodyWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *bodyWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

func ETagMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		originalWriter := c.Writer

		bw := &bodyWriter{
			ResponseWriter: originalWriter,
			body:           bytes.NewBuffer(nil),
			status:         http.StatusOK,
		}

		c.Writer = bw
		c.Next()

		body := bw.body.Bytes()

		hash := md5.Sum(body)
		etag := `"` + hex.EncodeToString(hash[:]) + `"`

		if c.GetHeader("If-None-Match") == etag {
			originalWriter.Header().Set("ETag", etag)
			originalWriter.WriteHeader(http.StatusNotModified)
			return
		}

		headers := originalWriter.Header()
		headers.Set("ETag", etag)
		headers.Set("Cache-Control", "private, max-age=0, must-revalidate")
		headers.Set("Content-Type", "application/json")

		originalWriter.WriteHeader(bw.status)
		_, _ = originalWriter.Write(body)
	}
}