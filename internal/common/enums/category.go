package enums

type CategoryType string

const (
	CategoryConcept CategoryType = "CONCEPT" // Khái niệm
	CategoryRule    CategoryType = "RULE"    // Quy tắc giao thông
	CategorySign    CategoryType = "SIGN"    // Biển báo
	CategorySaHinh  CategoryType = "SA_HINH" // Sa hình
	CategoryCulture CategoryType = "CULTURE" // Văn hóa giao thông
)
