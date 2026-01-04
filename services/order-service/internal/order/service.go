package order

import (
	"context"
	"uop-ms/services/order-service/internal/core"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

type CreateOrderItemInput struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
}

type CreateOrderInput struct {
	Items []CreateOrderItemInput `json:"items"`
}

func (s *Service) Create(ctx context.Context, userSub string, input CreateOrderInput) (*Order, *core.AppError) {
	if userSub == "" {
		return nil, core.NewInternal("UNAUTHORIZED", "Missing user identity")
	}

	if len(input.Items) == 0 {
		return nil, core.NewInternal("EMPTY_ORDER", "Order must contain at least one item")
	}

	var total float64
	items := make([]OrderItem, 0, len(input.Items))

	for _, it := range input.Items {
		if it.ProductID == "" {
			return nil, core.NewInternal("INVALID_PRODUCT_ID", "Product id required")
		}
		if it.Quantity <= 0 {
			return nil, core.NewInternal("INVALID_QUANTITY", "Quantity must be greater than zero")
		}
		if it.UnitPrice < 0 {
			return nil, core.NewInternal("INVALID_PRICE", "Unit price must be non-negative")
		}

		total += it.UnitPrice * float64(it.Quantity)
		items = append(items, OrderItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			UnitPrice: it.UnitPrice,
		})
	}

	o := &Order{
		UserSub:     userSub,
		TotalAmount: total,
		Status:      "NEW",
		Items:       items,
	}

	if err := s.store.Create(ctx, o); err != nil {
		return nil, core.NewInternal("ORDER_CREATE_FAILED", "Failed to create order")
	}
	return o, nil

}

func (s *Service) ListMyOrders(ctx context.Context, userSub string, limit int) ([]Order, *core.AppError) {
	if userSub == "" {
		return nil, core.NewInternal("UNAUTHORIZED", "Missing user identity")
	}

	items, err := s.store.ListByUser(ctx, userSub, limit)
	if err != nil {
		return nil, core.NewInternal("ORDER_LIST_FAILED", "Failed to list orders")
	}
	return items, nil
}
