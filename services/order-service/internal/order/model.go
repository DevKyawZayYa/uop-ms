package order

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Order struct {
	ID          string      `json:"id" gorm:"primaryKey;size:36"`
	UserSub     string      `json:"userSub" gorm:"size:64;index;not null"`
	TotalAmount float64     `json:"totalAmount" gorm:"not null"`
	Status      string      `json:"status" gorm:"size:32;not null"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
	Items       []OrderItem `json:"items,omitempty" gorm:"foreignKey:OrderID"`
}

func (Order) TableName() string { return "orders" }

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.NewString()
	}
	return nil
}

type OrderItem struct {
	OrderID   string  `json:"orderId" gorm:"primaryKey;size:36;not null"`
	ProductID string  `json:"productId" gorm:"primaryKey;size:36;not null"`
	Quantity  int     `json:"quantity" gorm:"not null"`
	UnitPrice float64 `json:"unitPrice" gorm:"not null"`
}

func (OrderItem) TableName() string { return "order_items" }
