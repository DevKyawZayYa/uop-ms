package grpc

import (
	"context"

	productv1 "uop-ms/gen/product/v1"
	"uop-ms/services/product-service/internal/product"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements productv1.ProductServiceServer
type Server struct {
	productv1.UnimplementedProductServiceServer
	store *product.Store
}

func NewServer(store *product.Store) *Server {
	return &Server{store: store}
}

func (s *Server) GetProductsByIds(
	ctx context.Context,
	req *productv1.GetProductsByIdsRequest,
) (*productv1.GetProductsByIdsResponse, error) {

	if len(req.ProductIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "product_ids is required")
	}

	products, err := s.store.GetByIDs(ctx, req.ProductIds)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch products")
	}

	resp := &productv1.GetProductsByIdsResponse{
		Products: make([]*productv1.Product, 0, len(products)),
	}

	for _, p := range products {
		resp.Products = append(resp.Products, &productv1.Product{
			Id:    p.ID,
			Name:  p.Name,
			Price: p.Price,
		})
	}

	return resp, nil
}

func (s *Server) ValidateProducts(
	ctx context.Context,
	req *productv1.ValidateProductsRequest,
) (*productv1.ValidateProductsResponse, error) {
	if len(req.ProductIds) == 0 {
		return &productv1.ValidateProductsResponse{}, nil
	}

	products, err := s.store.GetByIDs(ctx, req.ProductIds)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to validate products")
	}

	found := make(map[string]struct{})
	for _, p := range products {
		found[p.ID] = struct{}{}
	}

	valid := make([]string, 0)
	missing := make([]string, 0)

	for _, id := range req.ProductIds {
		if _, ok := found[id]; ok {
			valid = append(valid, id)
		} else {
			missing = append(missing, id)
		}
	}

	return &productv1.ValidateProductsResponse{
		ValidProductIds:   valid,
		MissingProductIds: missing,
	}, nil
}

func (s *Server) CheckAvailability(
	ctx context.Context,
	req *productv1.CheckAvailabilityRequest,
) (*productv1.CheckAvailabilityResponse, error) {

	insufficient := make([]*productv1.InsufficientProduct, 0)

	for _, item := range req.Items {
		p, err := s.store.GetByID(ctx, item.ProductId)
		if err != nil {
			insufficient = append(insufficient, &productv1.InsufficientProduct{
				ProductId: item.ProductId,
				Requested: item.Quantity,
				Available: 0,
			})
			continue
		}

		if p.Stock < int(item.Quantity) {
			insufficient = append(insufficient, &productv1.InsufficientProduct{
				ProductId: item.ProductId,
				Requested: item.Quantity,
				Available: int32(p.Stock),
			})
		}
	}

	return &productv1.CheckAvailabilityResponse{
		Available:    len(insufficient) == 0,
		Insufficient: insufficient,
	}, nil
}

func (s *Server) ResolveProductsForOrder(
	ctx context.Context,
	req *productv1.ResolveProductsForOrderRequest,
) (*productv1.ResolveProductsForOrderResponse, error) {

	resolved := make([]*productv1.ResolvedProduct, 0, len(req.Items))

	for _, item := range req.Items {
		p, err := s.store.GetByID(ctx, item.ProductId)
		if err != nil {
			return nil, status.Errorf(
				codes.NotFound,
				"product %s not found",
				item.ProductId,
			)
		}

		resolved = append(resolved, &productv1.ResolvedProduct{
			ProductId: p.ID,
			Name:      p.Name,
			UnitPrice: p.Price,
			Quantity:  item.Quantity,
		})
	}

	return &productv1.ResolveProductsForOrderResponse{
		Products: resolved,
	}, nil
}
