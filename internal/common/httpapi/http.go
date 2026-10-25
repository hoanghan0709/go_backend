package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/han/go-ecommerce/internal/learning"
	"gorm.io/gorm"
)

func Error(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "internal server error"
	if errors.Is(err, gorm.ErrRecordNotFound) {
		status = 404
		message = "not found"
	} else if errors.Is(err, learning.ErrInvalid) {
		status = 400
		message = err.Error()
	}
	if status == 500 {
		_ = c.Error(err)
	} // Ghi lỗi nội bộ vào Gin, không lộ SQL ra client.
	c.JSON(status, gin.H{"message": message})
}
func ID(c *gin.Context, name string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 32)
	if err != nil || id == 0 {
		c.JSON(400, gin.H{"message": "invalid " + name})
		return 0, false
	}
	return uint(id), true
}
func User(c *gin.Context) uint { return c.MustGet("userID").(uint) }
func Bind(c *gin.Context, out interface{}) bool {
	// Giới hạn import/request để tránh một body quá lớn giữ bộ nhớ của server.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 5<<20)
	if err := c.ShouldBindJSON(out); err != nil {
		c.JSON(400, gin.H{"message": "invalid request body", "detail": err.Error()})
		return false
	}
	return true
}
func Filter(c *gin.Context) (learning.Filter, bool) {
	f := learning.Filter{License: c.Query("license_type"), Search: c.Query("search")}
	if f.License != "" && !learning.ValidLicense(f.License) {
		c.JSON(400, gin.H{"message": "invalid license_type"})
		return f, false
	}
	for _, name := range []string{"page", "limit", "category_id"} {
		raw := c.Query(name)
		if raw == "" {
			continue
		}
		n, err := strconv.ParseUint(raw, 10, 31)
		if err != nil || n == 0 || (name == "page" && n > 1000000) {
			c.JSON(400, gin.H{"message": "invalid " + name})
			return f, false
		}
		switch name {
		case "page":
			f.Page = int(n)
		case "limit":
			f.Limit = int(n)
		case "category_id":
			f.CategoryID = uint(n)
		}
	}
	if raw := c.Query("is_critical"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			c.JSON(400, gin.H{"message": "invalid is_critical"})
			return f, false
		}
		f.Critical = &v
	}
	f.Page, f.Limit = f.Bounds()
	return f, true
}
func Page(c *gin.Context, data interface{}, total int64, f learning.Filter) {
	c.JSON(200, gin.H{"data": data, "total": total, "page": f.Page, "limit": f.Limit})
}
