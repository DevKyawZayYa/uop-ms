package order

import (
	"net/http"
	"strconv"
	"uop-ms/services/order-service/internal/core"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		v1.POST("/orders", h.Create)
		v1.GET("/orders", h.ListMyOrders)
	}
}

func (h *Handler) Create(c *gin.Context) {
	userSub := c.GetHeader("X-User-Sub")

	var input CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(core.NewInternal("INVALID_BODY", "Invalid request body"))
		return
	}

	for _, it := range input.Items {
		if _, err := uuid.Parse(it.ProductID); err != nil {
			c.Error(core.NewInternal("INVALID_PRODUCT_ID", "Invalid product id format"))
			return
		}
	}

	o, appErr := h.svc.Create(c.Request.Context(), userSub, input)
	if appErr != nil {
		c.Error(appErr)
		return
	}

	c.JSON(http.StatusCreated, core.APIResponse{
		Success: true,
		Data:    o,
	})
}

func (h *Handler) ListMyOrders(c *gin.Context) {
	userSub := c.GetHeader("X-User-Sub")

	limitStr := c.Query("limit")
	limit := 0
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}

	items, appErr := h.svc.ListMyOrders(c.Request.Context(), userSub, limit)
	if appErr != nil {
		c.Error(appErr)
		return
	}

	c.JSON(http.StatusOK, core.APIResponse{
		Success: true,
		Data:    items,
	})
}
