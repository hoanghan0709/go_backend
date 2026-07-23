package enums

type ExamStatus string

const (
	ExamDoing     ExamStatus = "DOING"
	ExamCompleted ExamStatus = "COMPLETED"
	ExamTimeout   ExamStatus = "TIMEOUT"
)
