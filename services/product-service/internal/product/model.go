package product

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	ID        string  `json:"id" gorm:"primaryKey; size:36"`
	Name      string  `json:"name" gorm:"size:200;not null"`
	Price     float64 `json:"price" gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Product) TableName() string {
	return "products"
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	return nil
}
