package category

import (
	// "github.com/han/go-ecommerce/internal/common/enums"
	common "github.com/han/go-ecommerce/internal/common/model"
	"github.com/han/go-ecommerce/internal/question/model"
)

type Category struct {
	common.BaseModel
	Name        string           `gorm:"size:255;not null"`
	Description string           `gorm:"size:255"`
	Questions   []model.Question `gorm:"foreignKey:CategoryID"`
	// Type        enums.CategoryType `gorm:"type:varchar(30);not null;unique"`
	//Sau khi xem source project của bạn, mình không khuyến khích thêm CategoryType.
	//Lý do là bạn đã có Name.

}
