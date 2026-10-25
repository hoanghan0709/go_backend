package validation

import (
	"fmt"
	answer "github.com/han/go-ecommerce/internal/answer/model"
	"strings"
)

// Answers dùng chung cho mọi đường ghi dữ liệu. Chỉ kiểm tra lúc tạo câu hỏi
// là chưa đủ: sửa hoặc xóa riêng một đáp án cũng có thể phá vỡ quy tắc này.
func Answers(answers []answer.Answer) error {
	if len(answers) < 2 {
		return fmt.Errorf("question must have at least two answers")
	}
	labels, orders, contents := map[string]bool{}, map[int]bool{}, map[string]bool{}
	correct := 0
	for i := range answers {
		a := &answers[i]
		a.Content = strings.TrimSpace(a.Content)
		a.Label = strings.ToUpper(strings.TrimSpace(a.Label))
		content := strings.ToLower(a.Content)
		if content == "" || a.Label == "" || len(a.Label) > 2 || a.Order < 1 {
			return fmt.Errorf("invalid answer content, label or order")
		}
		if labels[a.Label] || orders[a.Order] || contents[content] {
			return fmt.Errorf("duplicate answer label, order or content")
		}
		labels[a.Label] = true
		orders[a.Order] = true
		contents[content] = true
		if a.IsCorrect {
			correct++
		}
	}
	if correct != 1 {
		return fmt.Errorf("question must have exactly one correct answer")
	}
	return nil
}
