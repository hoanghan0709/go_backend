package module

type Test struct {
	ID    uint   `gorm:"primaryKey"`
	Title string `gorm:"size:255"`

	DurationMinutes int
	PassScore       int

	Questions []Question `gorm:"many2many:test_questions;"`
}
