package delivery

import (
	"github.com/gin-gonic/gin"
	authdto "github.com/han/go-ecommerce/internal/auth/dto"
	authmodel "github.com/han/go-ecommerce/internal/auth/model"
	"github.com/han/go-ecommerce/internal/common/httpapi"
	"github.com/han/go-ecommerce/internal/learning"
	"gorm.io/gorm"
)

func mobileFilter(c *gin.Context, s *learning.Service, usePreference bool) (learning.Filter, bool) {
	f, ok := httpapi.Filter(c)
	if !ok {
		return f, false
	}
	if usePreference {
		license, err := s.EffectiveLicense(httpapi.User(c), f.License)
		if err != nil {
			httpapi.Error(c, err)
			return f, false
		}
		f.License = license
	}
	return f, true
}
func RegisterMobileRoutes(r *gin.Engine, s *learning.Service, auth, admin gin.HandlerFunc) {
	api := r.Group("/api/v1", auth)
	api.GET("/license-types", func(c *gin.Context) {
		out := []gin.H{}
		for _, code := range learning.LicenseTypes() {
			out = append(out, gin.H{"code": code, "name": "Hạng " + code})
		}
		c.JSON(200, gin.H{"data": out})
	})
	api.GET("/me", func(c *gin.Context) {
		var u authmodel.User
		if err := s.DB.First(&u, httpapi.User(c)).Error; err != nil {
			httpapi.Error(c, err)
			return
		}
		c.JSON(200, authdto.ToUserResponse(&u))
	})
	api.PUT("/me/license", func(c *gin.Context) {
		var in struct {
			LicenseType string `json:"license_type" binding:"required"`
		}
		if !httpapi.Bind(c, &in) {
			return
		}
		out, err := s.SetLicense(httpapi.User(c), in.LicenseType)
		if err != nil {
			httpapi.Error(c, err)
			return
		}
		c.JSON(200, authdto.ToUserResponse(out))
	})
	api.GET("/category-types", func(c *gin.Context) {
		names := []string{"Chưa phân loại", "Khái niệm", "Quy tắc giao thông", "Biển báo", "Sa hình", "Văn hóa giao thông", "Kỹ thuật lái xe", "Cấu tạo và sửa chữa"}
		out := []gin.H{}
		for i, code := range learning.CategoryTypes {
			out = append(out, gin.H{"code": code, "name": names[i]})
		}
		c.JSON(200, gin.H{"data": out})
	})
	api.GET("/road-sign-types", func(c *gin.Context) {
		names := []string{"Biển cấm", "Biển cảnh báo", "Biển hiệu lệnh", "Biển chỉ dẫn", "Biển phụ", "Khác"}
		out := []gin.H{}
		for i, code := range learning.SignTypes {
			out = append(out, gin.H{"code": code, "name": names[i]})
		}
		c.JSON(200, gin.H{"data": out})
	})
	listSigns := func(isAdmin bool) gin.HandlerFunc {
		return func(c *gin.Context) {
			f, ok := mobileFilter(c, s, !isAdmin)
			if !ok {
				return
			}
			out, n, err := s.RoadSigns(f, isAdmin)
			if err != nil {
				httpapi.Error(c, err)
				return
			}
			httpapi.Page(c, out, n, f)
		}
	}
	getSign := func(isAdmin bool) gin.HandlerFunc {
		return func(c *gin.Context) {
			id, ok := httpapi.ID(c, "id")
			if !ok {
				return
			}
			f, ok := mobileFilter(c, s, !isAdmin)
			if !ok {
				return
			}
			out, err := s.RoadSign(id, f.License, isAdmin)
			if err != nil {
				httpapi.Error(c, err)
				return
			}
			c.JSON(200, out)
		}
	}
	api.GET("/road-signs", listSigns(false))
	api.GET("/road-signs/:id", getSign(false))
	api.GET("/road-signs/:id/questions", func(c *gin.Context) {
		id, ok := httpapi.ID(c, "id")
		if !ok {
			return
		}
		f, ok := mobileFilter(c, s, true)
		if !ok {
			return
		}
		if _, err := s.RoadSign(id, f.License, false); err != nil {
			httpapi.Error(c, err)
			return
		}
		f.RoadSignID = id
		out, n, err := s.Questions(f, false)
		if err != nil {
			httpapi.Error(c, err)
			return
		}
		httpapi.Page(c, learnerQuestions(out), n, f)
	})
	a := api.Group("/admin", admin)
	a.GET("/road-signs", listSigns(true))
	a.GET("/road-signs/:id", getSign(true))
	for _, method := range []string{"POST", "PUT"} {
		path := "/road-signs"
		if method == "PUT" {
			path += "/:id"
		}
		a.Handle(method, path, func(c *gin.Context) {
			var id uint
			if c.Request.Method == "PUT" {
				var ok bool
				id, ok = httpapi.ID(c, "id")
				if !ok {
					return
				}
			}
			var in learning.RoadSignInput
			if !httpapi.Bind(c, &in) {
				return
			}
			out, err := s.SaveRoadSign(id, in)
			if err != nil {
				httpapi.Error(c, err)
				return
			}
			status := 200
			if id == 0 {
				status = 201
			}
			c.JSON(status, out)
		})
	}
	a.DELETE("/road-signs/:id", func(c *gin.Context) {
		id, ok := httpapi.ID(c, "id")
		if !ok {
			return
		}
		if err := s.DeleteRoadSign(id); err != nil {
			httpapi.Error(c, err)
			return
		}
		c.Status(204)
	})
	listTemplates := func(isAdmin bool) gin.HandlerFunc {
		return func(c *gin.Context) {
			f, ok := mobileFilter(c, s, !isAdmin)
			if !ok {
				return
			}
			out, n, err := s.Templates(f, isAdmin)
			if err != nil {
				httpapi.Error(c, err)
				return
			}
			httpapi.Page(c, out, n, f)
		}
	}
	api.GET("/exam-templates", listTemplates(false))
	a.GET("/exam-templates", listTemplates(true))
	a.GET("/exam-templates/:id", func(c *gin.Context) {
		id, ok := httpapi.ID(c, "id")
		if !ok {
			return
		}
		var out learning.ExamTemplate
		if err := s.DB.First(&out, id).Error; err != nil {
			httpapi.Error(c, err)
			return
		}
		c.JSON(200, out)
	})
	for _, method := range []string{"POST", "PUT"} {
		path := "/exam-templates"
		if method == "PUT" {
			path += "/:id"
		}
		a.Handle(method, path, func(c *gin.Context) {
			var id uint
			if c.Request.Method == "PUT" {
				var ok bool
				id, ok = httpapi.ID(c, "id")
				if !ok {
					return
				}
			}
			var in learning.GenerateInput
			if !httpapi.Bind(c, &in) {
				return
			}
			out, err := s.SaveTemplate(id, in)
			if err != nil {
				httpapi.Error(c, err)
				return
			}
			status := 200
			if id == 0 {
				status = 201
			}
			c.JSON(status, out)
		})
	}
	a.DELETE("/exam-templates/:id", func(c *gin.Context) {
		id, ok := httpapi.ID(c, "id")
		if !ok {
			return
		}
		result := s.DB.Delete(&learning.ExamTemplate{}, id)
		if result.Error != nil {
			httpapi.Error(c, result.Error)
			return
		}
		if result.RowsAffected == 0 {
			httpapi.Error(c, gorm.ErrRecordNotFound)
			return
		}
		c.Status(204)
	})
}
