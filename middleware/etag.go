package middleware

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

type bodyWriter struct {
	gin.ResponseWriter
	body []byte
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

func ETagMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Only GET
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		bw := &bodyWriter{ResponseWriter: c.Writer}
		c.Writer = bw

		c.Next()

		// Skip errors
		if c.Writer.Status() >= 400 {
			return
		}

		// IMPORTANT: compute AFTER handler finishes
		sum := md5.Sum(bw.body)
		etag := `"` + hex.EncodeToString(sum[:]) + `"`

		// Compare request header
		if match := c.GetHeader("If-None-Match"); match == etag {
			c.Header("ETag", etag)
			c.AbortWithStatus(http.StatusNotModified)
			return
		}

		// Set headers BEFORE final response is flushed
		c.Header("ETag", etag)
		c.Header("Cache-Control", "private, max-age=0, must-revalidate")
	}
}