package module

type Question struct {
	ID      uint   `gorm:"primaryKey"`
	Title   string `gorm:"size:255"`
	Content string `gorm:"type:text"`
	Image   string

	IsCritical bool

	CategoryID uint
	Category   Category

	Answers []Answer
}
