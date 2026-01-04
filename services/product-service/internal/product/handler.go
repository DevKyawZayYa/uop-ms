package product

import (
	"net/http"
	"strconv"

	"uop-ms/services/product-service/internal/core"

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
		v1.GET("/products", h.List)
		v1.GET("/products/:id", h.Get)
	}
}

func (h *Handler) List(c *gin.Context) {
	limitStr := c.Query("limit")
	limit := 0
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}

	items, appErr := h.svc.List(c.Request.Context(), limit)
	if appErr != nil {
		c.Error(appErr)
		return
	}

	c.JSON(http.StatusOK, core.APIResponse{
		Success: true,
		Data:    items,
	})
}

func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.Error(core.NewInternal("INVALID_PRODUCT_ID", "Invalid product id"))
		return
	}

	p, appErr := h.svc.Get(c.Request.Context(), id)
	if appErr != nil {
		c.Error(appErr)
		return
	}

	c.JSON(http.StatusOK, core.APIResponse{
		Success: true,
		Data:    p,
	})
}
