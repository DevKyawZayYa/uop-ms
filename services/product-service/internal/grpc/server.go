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
