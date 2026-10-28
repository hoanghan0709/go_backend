package crawler

import (
	"context"
	"errors"
	"net"

	"github.com/gin-gonic/gin"
	"github.com/han/go-ecommerce/internal/common/httpapi"
)

func RegisterRoutes(r *gin.Engine, s *Service, auth, admin gin.HandlerFunc) {
	r.POST("/api/v1/admin/questions/crawl", auth, admin, func(c *gin.Context) {
		// Một thao tác cố định: không có body/query, không có URL do client điều khiển.
		if c.Request.URL.RawQuery != "" || c.Request.ContentLength > 0 || len(c.Request.TransferEncoding) > 0 {
			c.JSON(400, gin.H{"message": "this endpoint accepts no query parameters or request body"})
			return
		}
		result, err := s.Crawl(c.Request.Context())
		if err == nil {
			c.JSON(200, result)
			return
		}
		var networkError net.Error
		switch {
		case errors.Is(err, ErrBusy):
			c.JSON(409, gin.H{"message": err.Error()})
		case errors.Is(err, ErrCooldown):
			c.Header("Retry-After", "60")
			c.JSON(429, gin.H{"message": err.Error()})
		case errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &networkError) && networkError.Timeout()):
			c.JSON(504, gin.H{"message": "crawl timed out; no partial import was committed"})
		case errors.Is(err, ErrUpstream):
			c.JSON(502, gin.H{"message": err.Error()})
		default:
			httpapi.Error(c, err)
		}
	})
}
