package delivery

import (
	"github.com/gin-gonic/gin"
	category "github.com/han/go-ecommerce/internal/category/model"
	"github.com/han/go-ecommerce/internal/common/httpapi"
	exam "github.com/han/go-ecommerce/internal/exam/model"
	"github.com/han/go-ecommerce/internal/learning"
	question "github.com/han/go-ecommerce/internal/question/model"
	test "github.com/han/go-ecommerce/internal/test/model"
	"gorm.io/gorm"
)

// LearnerQuestion không có khóa đáp án hay giải thích. API luyện tập chỉ cung cấp
// lời giải sau khi người học gửi lựa chọn; API quản trị sử dụng model đầy đủ.
func learnerQuestions(qs []question.Question) []exam.SnapshotQuestion {
	out := []exam.SnapshotQuestion{}
	for _, q := range qs {
		sq := exam.SnapshotQuestion{ID: q.ID, Title: q.Title, Content: q.Content, Image: q.Image, CategoryID: q.CategoryID, IsCritical: q.IsCritical, Answers: []exam.SnapshotAnswer{}}
		for _, a := range q.Answers {
			sq.Answers = append(sq.Answers, exam.SnapshotAnswer{ID: a.ID, Content: a.Content, Label: a.Label, Order: a.Order})
		}
		out = append(out, sq)
	}
	return out
}
func RegisterRoutes(r *gin.Engine, s *learning.Service, auth, admin gin.HandlerFunc, expire func(uint) error) {
	api := r.Group("/api/v1", auth)
	api.GET("/categories", func(c *gin.Context) {
		var out []category.Category
		err := s.DB.Order("id").Find(&out).Error
		if err != nil {
			httpapi.Error(c, err)
			return
		}
		c.JSON(200, gin.H{"data": out})
	})
	listQuestions := func(isAdmin, wrong bool) gin.HandlerFunc {
		return func(c *gin.Context) {
			f, ok := httpapi.Filter(c)
			if !ok {
				return
			}
			var qs []question.Question
			var total int64
			var err error
			if wrong {
				if err = expire(httpapi.User(c)); err != nil {
					httpapi.Error(c, err)
					return
				}
				qs, total, err = s.WrongQuestions(httpapi.User(c), f)
			} else {
				qs, total, err = s.Questions(f, isAdmin)
			}
			if err != nil {
				httpapi.Error(c, err)
				return
			}
			if isAdmin {
				httpapi.Page(c, qs, total, f)
			} else {
				httpapi.Page(c, learnerQuestions(qs), total, f)
			}
		}
	}
	api.GET("/practice/questions", listQuestions(false, false))
	api.GET("/practice/wrong-questions", listQuestions(false, true))
	api.POST("/practice/questions/:id/answer", func(c *gin.Context) {
		id, ok := httpapi.ID(c, "id")
		if !ok {
			return
		}
		var in struct {
			AnswerID uint `json:"answer_id" binding:"required"`
		}
		if !httpapi.Bind(c, &in) {
			return
		}
		a, q, err := s.Practice(httpapi.User(c), id, in.AnswerID)
		if err != nil {
			httpapi.Error(c, err)
			return
		}
		c.JSON(200, gin.H{"attempt": a, "question": q})
	})
	api.GET("/progress", func(c *gin.Context) {
		f, ok := httpapi.Filter(c)
		if !ok {
			return
		}
		if err := expire(httpapi.User(c)); err != nil {
			httpapi.Error(c, err)
			return
		}
		out, err := s.Progress(httpapi.User(c), f.License)
		if err != nil {
			httpapi.Error(c, err)
			return
		}
		c.JSON(200, out)
	})
	listTests := func(isAdmin bool) gin.HandlerFunc {
		return func(c *gin.Context) {
			f, ok := httpapi.Filter(c)
			if !ok {
				return
			}
			db := s.DB.Model(&test.Test{})
			if !isAdmin {
				db = db.Where("published = ?", true)
			}
			if f.License != "" {
				db = db.Where("license_type = ?", f.License)
			}
			var n int64
			if err := db.Count(&n).Error; err != nil {
				httpapi.Error(c, err)
				return
			}
			out := []test.Test{}
			if err := db.Order("id DESC").Offset((f.Page - 1) * f.Limit).Limit(f.Limit).Find(&out).Error; err != nil {
				httpapi.Error(c, err)
				return
			}
			httpapi.Page(c, out, n, f)
		}
	}
	api.GET("/tests", listTests(false))
	a := api.Group("/admin", admin)
	a.GET("/questions", listQuestions(true, false))
	a.GET("/questions/:id", func(c *gin.Context) {
		id, ok := httpapi.ID(c, "id")
		if !ok {
			return
		}
		var out question.Question
		if err := s.DB.Preload("Category").Preload("Answers", func(tx *gorm.DB) *gorm.DB { return tx.Order("\"order\" ASC") }).First(&out, id).Error; err != nil {
			httpapi.Error(c, err)
			return
		}
		c.JSON(200, out)
	})
	a.POST("/questions", func(c *gin.Context) {
		var in learning.QuestionInput
		if !httpapi.Bind(c, &in) {
			return
		}
		out, err := s.SaveQuestion(0, in)
		if err != nil {
			httpapi.Error(c, err)
			return
		}
		c.JSON(201, out)
	})
	a.PUT("/questions/:id", func(c *gin.Context) {
		id, ok := httpapi.ID(c, "id")
		if !ok {
			return
		}
		var in learning.QuestionInput
		if !httpapi.Bind(c, &in) {
			return
		}
		out, err := s.SaveQuestion(id, in)
		if err != nil {
			httpapi.Error(c, err)
			return
		}
		c.JSON(200, out)
	})
	a.DELETE("/questions/:id", func(c *gin.Context) {
		id, ok := httpapi.ID(c, "id")
		if !ok {
			return
		}
		if err := s.DeleteQuestion(id); err != nil {
			httpapi.Error(c, err)
			return
		}
		c.Status(204)
	})
	a.POST("/questions/import", func(c *gin.Context) {
		var in struct {
			Questions []learning.QuestionInput `json:"questions" binding:"required,min=1,max=500,dive"`
		}
		if !httpapi.Bind(c, &in) {
			return
		}
		out, err := s.ImportQuestions(in.Questions)
		if err != nil {
			httpapi.Error(c, err)
			return
		}
		c.JSON(201, gin.H{"data": out, "imported": len(out)})
	})
	for _, method := range []string{"POST", "PUT"} {
		path := "/categories"
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
			var in learning.CategoryInput
			if !httpapi.Bind(c, &in) {
				return
			}
			out, err := s.SaveCategory(id, in)
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
	a.DELETE("/categories/:id", func(c *gin.Context) {
		id, ok := httpapi.ID(c, "id")
		if !ok {
			return
		}
		if err := s.DeleteCategory(id); err != nil {
			httpapi.Error(c, err)
			return
		}
		c.Status(204)
	})
	a.GET("/tests", listTests(true))
	a.GET("/tests/:id", func(c *gin.Context) {
		id, ok := httpapi.ID(c, "id")
		if !ok {
			return
		}
		var out test.Test
		// Questions chỉ là bảng nối đề–câu. Tải cả câu, chủ đề và phương án
		// để admin xem trọn đề; route này đã được middleware giới hạn quyền admin.
		if err := s.DB.
			Preload("Questions", func(db *gorm.DB) *gorm.DB { return db.Order("\"order\" ASC") }).
			Preload("Questions.Question.Category").
			Preload("Questions.Question.Answers", func(db *gorm.DB) *gorm.DB { return db.Order("\"order\" ASC") }).
			First(&out, id).Error; err != nil {
			httpapi.Error(c, err)
			return
		}
		c.JSON(200, out)
	})
	for _, method := range []string{"POST", "PUT"} {
		path := "/tests"
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
			var in learning.TestInput
			if !httpapi.Bind(c, &in) {
				return
			}
			out, err := s.SaveTest(id, in)
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
	a.POST("/tests/generate", func(c *gin.Context) {
		var in learning.GenerateInput
		if !httpapi.Bind(c, &in) {
			return
		}
		out, err := s.GenerateTest(in)
		if err != nil {
			httpapi.Error(c, err)
			return
		}
		c.JSON(201, out)
	})
	// Xóa mềm đề không ảnh hưởng snapshot của các phiên thi đã bắt đầu.
	a.DELETE("/tests/:id", func(c *gin.Context) {
		id, ok := httpapi.ID(c, "id")
		if !ok {
			return
		}
		err := s.DB.Transaction(func(tx *gorm.DB) error {
			r := tx.Delete(&test.Test{}, id)
			if r.Error != nil {
				return r.Error
			}
			if r.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
			return tx.Unscoped().Where("test_id = ?", id).Delete(&test.TestQuestion{}).Error
		})
		if err != nil {
			httpapi.Error(c, err)
			return
		}
		c.Status(204)
	})
}
