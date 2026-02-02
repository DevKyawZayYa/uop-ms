package grpc

import (
	"context"
	"time"
	productv1 "uop-ms/gen/product/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProductClient struct {
	client productv1.ProductServiceClient
}

func NewProductClient(addr string) (*ProductClient, error) {
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, err
	}

	return &ProductClient{
		client: productv1.NewProductServiceClient(conn),
	}, nil
}

func (c *ProductClient) GetProductsByIDs(
	ctx context.Context,
	ids []string,
) (map[string]*productv1.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := c.client.GetProductsByIds(ctx, &productv1.GetProductsByIdsRequest{
		ProductIds: ids,
	})

	if err != nil {
		return nil, err
	}

	result := make(map[string]*productv1.Product)
	for _, p := range resp.Products {
		result[p.Id] = p
	}

	return result, nil

}
