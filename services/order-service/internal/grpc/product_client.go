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

type OrderItem struct {
	ProductID string
	Quantity  int
}

type CheckItem struct {
	ProductID string
	Quantity  int
}

type ResolvedProduct struct {
	ProductID string
	Quantity  int
	UnitPrice float64
}

// ValidateProducts calls product-service ValidateProducts RPC
func (c *ProductClient) ValidateProducts(
	ctx context.Context,
	productIDs []string,
) (valid []string, missing []string, err error) {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := c.client.ValidateProducts(
		ctx,
		&productv1.ValidateProductsRequest{
			ProductIds: productIDs,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return resp.ValidProductIds, resp.MissingProductIds, nil
}

// CheckAvailability calls product-service CheckAvailability RPC
func (c *ProductClient) CheckAvailability(
	ctx context.Context,
	items []OrderItem,
) (available bool, insufficient []CheckItem, err error) {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	reqItems := make([]*productv1.CheckAvailabilityItem, 0, len(items))
	for _, it := range items {
		reqItems = append(reqItems, &productv1.CheckAvailabilityItem{
			ProductId: it.ProductID,
			Quantity:  int32(it.Quantity),
		})
	}

	resp, err := c.client.CheckAvailability(
		ctx,
		&productv1.CheckAvailabilityRequest{
			Items: reqItems,
		},
	)
	if err != nil {
		return false, nil, err
	}

	ins := make([]CheckItem, 0, len(resp.Insufficient))
	for _, i := range resp.Insufficient {
		ins = append(ins, CheckItem{
			ProductID: i.ProductId,
			Quantity:  int(i.Requested),
		})
	}

	return resp.Available, ins, nil
}

// ResolveProductsForOrder calls product-service ResolveProductsForOrder RPC
func (c *ProductClient) ResolveProductsForOrder(
	ctx context.Context,
	items []OrderItem,
) ([]ResolvedProduct, error) {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	reqItems := make([]*productv1.ResolveProductItem, 0, len(items))
	for _, it := range items {
		reqItems = append(reqItems, &productv1.ResolveProductItem{
			ProductId: it.ProductID,
			Quantity:  int32(it.Quantity),
		})
	}

	resp, err := c.client.ResolveProductsForOrder(
		ctx,
		&productv1.ResolveProductsForOrderRequest{
			Items: reqItems,
		},
	)
	if err != nil {
		return nil, err
	}

	resolved := make([]ResolvedProduct, 0, len(resp.Products))
	for _, p := range resp.Products {
		resolved = append(resolved, ResolvedProduct{
			ProductID: p.ProductId,
			Quantity:  int(p.Quantity),
			UnitPrice: p.UnitPrice,
		})
	}

	return resolved, nil
}
