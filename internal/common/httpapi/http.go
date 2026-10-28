package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

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
		c.JSON(400, gin.H{"message": name + " must be a positive integer"})
		return 0, false
	}
	return uint(id), true
}
func User(c *gin.Context) uint { return c.MustGet("userID").(uint) }
func Bind(c *gin.Context, out interface{}) bool {
	// Giới hạn import/request để tránh một body quá lớn giữ bộ nhớ của server.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 5<<20)
	if err := c.ShouldBindJSON(out); err != nil {
		BindingError(c, out, err)
		return false
	}
	return true
}
func Filter(c *gin.Context) (learning.Filter, bool) {
	f := learning.Filter{License: c.Query("license_type"), Search: c.Query("search")}
	if f.License != "" && !learning.ValidLicense(f.License) {
		c.JSON(400, gin.H{"message": "license_type must be one of A1, A, B, C1, C, D"})
		return f, false
	}
	for _, name := range []string{"page", "limit", "category_id"} {
		raw := c.Query(name)
		if raw == "" {
			continue
		}
		n, err := strconv.ParseUint(raw, 10, 31)
		if name == "page" && err == nil && n > 1000000 {
			c.JSON(400, gin.H{"message": "page must be between 1 and 1000000"})
			return f, false
		}
		if err != nil || n == 0 {
			c.JSON(400, gin.H{"message": name + " must be a positive integer"})
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
			c.JSON(400, gin.H{"message": "is_critical must be a boolean (true or false)"})
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

// BindingError uses JSON field paths so clients can highlight nested inputs without
// depending on Go struct names. Never echo submitted values (passwords/tokens).
func BindingError(c *gin.Context, out interface{}, err error) {
	type issue struct {
		Field   string `json:"field"`
		Rule    string `json:"rule"`
		Message string `json:"message"`
	}
	issues := []issue{}
	var ve validator.ValidationErrors
	var syntax *json.SyntaxError
	var mismatch *json.UnmarshalTypeError
	var large *http.MaxBytesError
	switch {
	case errors.As(err, &ve):
		for _, e := range ve {
			field := jsonPath(reflect.TypeOf(out), e.StructNamespace())
			message := field + " is invalid"
			unit := ""
			switch e.Kind() {
			case reflect.String:
				unit = " characters"
			case reflect.Slice, reflect.Array, reflect.Map:
				unit = " items"
			}
			switch e.Tag() {
			case "required":
				message = field + " is required and must not be empty"
			case "email":
				message = field + " must be a valid email address"
			case "min":
				message = field + " must be at least " + e.Param() + unit
			case "max":
				message = field + " must be at most " + e.Param() + unit
			}
			issues = append(issues, issue{field, e.Tag(), message})
		}
	case errors.As(err, &large):
		issues = append(issues, issue{"body", "max_size", "request body must not exceed 5 MiB"})
	case errors.As(err, &mismatch):
		field := mismatch.Field
		if field == "" {
			field = "body"
		}
		issues = append(issues, issue{field, "type", field + " must have JSON type compatible with " + mismatch.Type.String()})
	case errors.As(err, &syntax):
		issues = append(issues, issue{"body", "json", fmt.Sprintf("malformed JSON near byte %d", syntax.Offset)})
	case errors.Is(err, io.EOF):
		issues = append(issues, issue{"body", "required", "request body is required; send a JSON object"})
	case errors.Is(err, io.ErrUnexpectedEOF):
		issues = append(issues, issue{"body", "json", "incomplete JSON; check closing quotes, braces and brackets"})
	default:
		issues = append(issues, issue{"body", "json", "request body must contain valid JSON matching the documented request"})
	}
	messages := make([]string, len(issues))
	for i, item := range issues {
		messages[i] = item.Message
	}
	c.JSON(400, gin.H{"status_code": 400, "message": strings.Join(messages, "; "), "errors": issues})
}

func jsonPath(t reflect.Type, namespace string) string {
	parts := strings.Split(namespace, ".")
	result := []string{}
	for _, part := range parts[1:] {
		for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
			t = t.Elem()
		}
		name, suffix := part, ""
		if i := strings.Index(part, "["); i >= 0 {
			name, suffix = part[:i], part[i:]
		}
		if t.Kind() != reflect.Struct {
			result = append(result, part)
			continue
		}
		f, ok := t.FieldByName(name)
		if !ok {
			result = append(result, part)
			continue
		}
		t = f.Type
		tag := strings.Split(f.Tag.Get("json"), ",")[0]
		if f.Anonymous && tag == "" {
			continue
		}
		if tag == "" {
			tag = name
		}
		result = append(result, tag+suffix)
	}
	return strings.Join(result, ".")
}
