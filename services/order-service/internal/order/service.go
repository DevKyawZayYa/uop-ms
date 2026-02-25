package order

import (
	"context"
	"log"
	"strings"
	"time"
	"uop-ms/services/order-service/internal/core"

	ordergrpc "uop-ms/services/order-service/internal/grpc"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	store         *Store
	publisher     *Publisher
	redis         *redis.Client
	productClient *ordergrpc.ProductClient
}

func NewService(
	store *Store,
	publisher *Publisher,
	redis *redis.Client,
	productClient *ordergrpc.ProductClient,
) *Service {
	return &Service{
		store:         store,
		publisher:     publisher,
		redis:         redis,
		productClient: productClient,
	}
}

type CreateOrderItemInput struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

type CreateOrderInput struct {
	Items []CreateOrderItemInput `json:"items"`
}

func (s *Service) Create(
	ctx context.Context,
	userSub string,
	idempotencyKey string,
	input CreateOrderInput,
) (*Order, *core.AppError) {

	// 1. Validate user identity
	if userSub == "" {
		return nil, core.NewBadRequest("UNAUTHORIZED", "Missing user identity")
	}

	// 2. redis IdempotencyKey
	if idempotencyKey == "" {
		return nil, core.NewBadRequest(
			"IDEMPOTENCY_KEY_REQUIRED",
			"Missing Idempotency-Key header",
		)
	}

	// 3. Validate request payload
	if len(input.Items) == 0 {
		return nil, core.NewBadRequest("EMPTY_ORDER", "Order must contain at least one item")
	}

	//4. idemp key check before DB
	redisKey := "idemp:order:" + userSub + ":" + idempotencyKey

	ok, err := s.redis.SetNX(
		ctx,
		redisKey,
		"processing",
		24*time.Hour,
	).Result()
	if err != nil {
		return nil, core.NewServiceUnavailable("REDIS_ERROR", "Failed to check idempotency")
	}

	// 5. Short-circuit duplicate requests
	if !ok {
		val, err := s.redis.Get(ctx, redisKey).Result()
		if err == nil && strings.HasPrefix(val, "order:") {
			orderID := strings.TrimPrefix(val, "order:")
			existing, err := s.store.GetByID(ctx, orderID)
			if err == nil {
				return existing, nil
			}
		}
		return nil, core.NewConflict("DUPLICATE_REQUEST", "Order already processed")
	}

	// 6. Collect product IDs and build grpc order items
	productIDs := make([]string, 0, len(input.Items))
	grpcItems := make([]ordergrpc.OrderItem, 0, len(input.Items))

	for _, it := range input.Items {
		if it.Quantity <= 0 {
			return nil, core.NewBadRequest("INVALID_QUANTITY", "Quantity must be greater than zero")
		}

		productIDs = append(productIDs, it.ProductID)
		grpcItems = append(grpcItems, ordergrpc.OrderItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
		})
	}

	// 7. Validate product existence via product-service
	_, missing, err := s.productClient.ValidateProducts(ctx, productIDs)
	if err != nil {
		return nil, core.NewServiceUnavailable(
			"PRODUCT_SERVICE_UNAVAILABLE",
			"Product service unavailable",
		)
	}
	if len(missing) > 0 {
		return nil, core.NewBadRequest(
			"PRODUCT_NOT_FOUND",
			"Missing products: "+strings.Join(missing, ","),
		)
	}

	// 8. Check inventory availability via product-service
	available, _, err := s.productClient.CheckAvailability(ctx, grpcItems)
	if err != nil {
		return nil, core.NewServiceUnavailable(
			"PRODUCT_SERVICE_UNAVAILABLE",
			"Product service unavailable",
		)
	}
	if !available {
		return nil, core.NewConflict(
			"INSUFFICIENT_STOCK",
			"Some items are out of stock",
		)
	}

	// 9. Resolve authoritative product snapshot for order
	resolvedProducts, err := s.productClient.ResolveProductsForOrder(ctx, grpcItems)
	if err != nil {
		return nil, core.NewServiceUnavailable(
			"PRODUCT_SERVICE_UNAVAILABLE",
			"Failed to resolve products for order",
		)
	}

	// 10. Compute order totals from resolved snapshot
	var total float64
	items := make([]OrderItem, 0, len(resolvedProducts))

	for _, rp := range resolvedProducts {
		total += rp.UnitPrice * float64(rp.Quantity)
		items = append(items, OrderItem{
			ProductID: rp.ProductID,
			Quantity:  rp.Quantity,
			UnitPrice: rp.UnitPrice,
		})
	}

	o := &Order{
		UserSub:     userSub,
		TotalAmount: total,
		Status:      "NEW",
		Items:       items,
	}

	// 11. Persist order in database
	if err := s.store.Create(ctx, o); err != nil {
		return nil, core.NewInternal("ORDER_CREATE_FAILED", "Failed to create order")
	}

	// 12. Publish OrderCreated event
	traceID := "no-trace"
	if v := ctx.Value("traceId"); v != nil {
		if s, ok := v.(string); ok {
			traceID = s
		}
	}

	if err := s.publisher.PublishOrderCreated(ctx, traceID, OrderCreatedPayload{
		OrderID:  o.ID,
		UserSub:  userSub,
		Total:    o.TotalAmount,
		Currency: "MYR",
	}); err != nil {
		log.Println("[order-service] kafka publish failed:", err)
	} else {
		log.Println("[order-service] kafka published OrderCreated:", o.ID)
	}

	// 13. store in redis
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
