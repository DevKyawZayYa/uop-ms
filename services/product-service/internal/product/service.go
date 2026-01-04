package product

import (
	"context"

	"uop-ms/services/product-service/internal/core"

	"gorm.io/gorm"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) List(ctx context.Context, limit int) ([]Product, *core.AppError) {
	items, err := s.store.List(ctx, limit)
	if err != nil {
		return nil, core.NewInternal("PRODUCT_LIST_FAILED", "Failed to list products")
	}
	return items, nil
}

func (s *Service) Get(ctx context.Context, id string) (*Product, *core.AppError) {
	p, err := s.store.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, core.NewInternal("PRODUCT_NOT_FOUND", "Product not found")
		}
		return nil, core.NewInternal("PRODUCT_GET_FAILED", "Failed to get product")
	}
	return p, nil
}
