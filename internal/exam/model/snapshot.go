package exam

// Snapshot giữ nguyên đề và đáp án lúc bắt đầu. Không phụ thuộc vào nội dung
// ngân hàng câu hỏi khi người dùng xem lại bài thi sau này.
type Snapshot struct {
	Name        string             `json:"name"`
	LicenseType string             `json:"license_type"`
	Version     string             `json:"version"`
	Questions   []SnapshotQuestion `json:"questions"`
}
type SnapshotQuestion struct {
	ID               uint             `json:"id"`
	Title            string           `json:"title"`
	Content          string           `json:"content"`
	Image            string           `json:"image"`
	CategoryID       uint             `json:"category_id"`
	IsCritical       bool             `json:"is_critical"`
	Explanation      string           `json:"explanation,omitempty"`
	Answers          []SnapshotAnswer `json:"answers"`
	SelectedAnswerID uint             `json:"selected_answer_id,omitempty"`
	IsCorrect        *bool            `json:"is_correct,omitempty"`
}
type SnapshotAnswer struct {
	ID        uint   `json:"id"`
	Content   string `json:"content"`
	Label     string `json:"label"`
	Order     int    `json:"order"`
	IsCorrect *bool  `json:"is_correct,omitempty"`
}
