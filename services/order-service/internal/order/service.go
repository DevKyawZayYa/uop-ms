package order

import (
	"context"
	"log"
	"strings"
	"time"
	"uop-ms/services/order-service/internal/core"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	store     *Store
	publisher *Publisher
	redis     *redis.Client
}

func NewService(store *Store, publisher *Publisher, redis *redis.Client) *Service {
	return &Service{store: store, publisher: publisher, redis: redis}
}

type CreateOrderItemInput struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
}

type CreateOrderInput struct {
	Items []CreateOrderItemInput `json:"items"`
}

func (s *Service) Create(ctx context.Context, userSub string, idempotencyKey string, input CreateOrderInput) (*Order, *core.AppError) {
	if userSub == "" {
		return nil, core.NewInternal("UNAUTHORIZED", "Missing user identity")
	}

	//redis IdempotencyKey
	if idempotencyKey == "" {
		return nil, core.NewInternal(
			"IDEMPOTENCY_KEY_REQUIRED",
			"Missing Idempotency-Key header",
		)
	}

	if len(input.Items) == 0 {
		return nil, core.NewInternal("EMPTY_ORDER", "Order must contain at least one item")
	}

	//idemp key check before DB
	redisKey := "idemp:order:" + userSub + ":" + idempotencyKey

	ok, err := s.redis.SetNX(
		ctx,
		redisKey,
		"processing",
		24*time.Hour,
	).Result()

	if err != nil {
		return nil, core.NewInternal("REDIS_ERROR", "Failed to check idempotency")
	}

	if !ok {
		// Request already processed
		val, err := s.redis.Get(ctx, redisKey).Result()
		if err == nil && strings.HasPrefix(val, "order:") {
			orderID := strings.TrimPrefix(val, "order:")
			existing, err := s.store.GetByID(ctx, orderID)
			if err == nil {
				return existing, nil
			}
		}
		return nil, core.NewInternal("DUPLICATE_REQUEST", "Order already processed")
	}

	// Prepare order
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

	// DB hit
	if err := s.store.Create(ctx, o); err != nil {
		return nil, core.NewInternal("ORDER_CREATE_FAILED", "Failed to create order")
	}

	// Kafka publish happens AFTER DB commit
	traceID := "no-trace"
	if v := ctx.Value("traceId"); v != nil {
		if s, ok := v.(string); ok {
			traceID = s
		}
	}

	err = s.publisher.PublishOrderCreated(ctx, traceID, OrderCreatedPayload{
		OrderID:  o.ID,
		UserSub:  userSub,
		Total:    o.TotalAmount,
		Currency: "MYR",
	})

	if err != nil {
		log.Println("[order-service] kafka publish failed:", err)
	} else {
		log.Println("[order-service] kafka published OrderCreated:", o.ID)
	}

	// store in redis
	_ = s.redis.Set(
		ctx,
		redisKey,
		"order:"+o.ID,
		24*time.Hour,
	).Err()

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
