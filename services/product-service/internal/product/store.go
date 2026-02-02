package product

import (
	"context"

	"gorm.io/gorm"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) List(ctx context.Context, limit int) ([]Product, error) {
	var items []Product
	q := s.db.WithContext(ctx).Order("created_at desc")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) GetByID(ctx context.Context, id string) (*Product, error) {
	var p Product
	if err := s.db.WithContext(ctx).First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) GetByIDs(ctx context.Context, ids []string) ([]Product, error) {
	var products []Product

	if err := s.db.WithContext(ctx).Where("id IN ?", ids).Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}
