package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/workflow/basket"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/assert"
)

var basketServiceRegistry = registry.NewServiceRegistry[basket.BasketService]("basket_service")

func init() {
	basketServiceRegistry.Register("local", func(ctx context.Context) (basket.BasketService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		return basket.NewBasketServiceImpl(ctx, db)
	})
}

func TestBasketServiceUpdateBasket(t *testing.T) {
	ctx := context.Background()
	basketService, err := basketServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	cart := basket.CustomerBasket{
		UserName: "updateuser",
		BuyerID:  "buyer1",
		Items: []basket.BasketItem{
			{ProductID: 1, ProductName: "Laptop", UnitPrice: 999.99, Quantity: 1},
		},
	}
	resp, err := basketService.UpdateBasket(ctx, basket.UpdateBasketRequest{Cart: cart})
	assert.NoError(t, err)
	assert.Equal(t, "updateuser", resp.Cart.UserName)
	assert.Len(t, resp.Cart.Items, 1)
}

func TestBasketServiceGetBasket(t *testing.T) {
	ctx := context.Background()
	basketService, err := basketServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	cart := basket.CustomerBasket{
		UserName: "getuser",
		BuyerID:  "buyer2",
		Items: []basket.BasketItem{
			{ProductID: 2, ProductName: "Phone", UnitPrice: 499.99, Quantity: 2},
		},
	}
	_, err = basketService.UpdateBasket(ctx, basket.UpdateBasketRequest{Cart: cart})
	assert.NoError(t, err)

	result, err := basketService.GetBasket(ctx, basket.GetBasketRequest{UserName: "getuser"})
	assert.NoError(t, err)
	assert.Equal(t, "getuser", result.Cart.UserName)
	assert.Len(t, result.Cart.Items, 1)
}

func TestBasketServiceDeleteBasket(t *testing.T) {
	ctx := context.Background()
	basketService, err := basketServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	cart := basket.CustomerBasket{
		UserName: "deleteuser",
		Items:    []basket.BasketItem{},
	}
	_, err = basketService.UpdateBasket(ctx, basket.UpdateBasketRequest{Cart: cart})
	assert.NoError(t, err)

	result, err := basketService.DeleteBasket(ctx, basket.DeleteBasketRequest{UserName: "deleteuser"})
	assert.NoError(t, err)
	assert.True(t, result.IsSuccess)

	_, err = basketService.GetBasket(ctx, basket.GetBasketRequest{UserName: "deleteuser"})
	assert.Error(t, err)
}
