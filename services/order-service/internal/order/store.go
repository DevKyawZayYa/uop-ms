package order

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(ctx context.Context, o *Order) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		o.ID = uuid.NewString()
		for i := range o.Items {
			o.Items[i].OrderID = o.ID
		}

		if err := tx.Create(o).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *Store) ListByUser(ctx context.Context, userSub string, limit int) ([]Order, error) {
	var items []Order
	q := s.db.WithContext(ctx).Where("user_sub = ?", userSub).Order("created_at desc").Preload("Items")

	if limit > 0 {
		q = q.Limit(limit)
	}

	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) GetByID(ctx context.Context, orderID string) (*Order, error) {
	var o Order
	if err := s.db.
		WithContext(ctx).
		Where("id = ?", orderID).
		Preload("Items").
		First(&o).Error; err != nil {
		return nil, err
	}
	return &o, nil
}
