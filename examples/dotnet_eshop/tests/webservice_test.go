package tests

import (
	"context"
	"testing"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/workflow/basket"
	"github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/workflow/catalog"
	"github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/workflow/order"
	"github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/workflow/payment"
	"github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/workflow/web"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplequeue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testWebApp struct {
	webapp     web.WebApp
	basketSvc  basket.BasketService
	catalogSvc catalog.CatalogService
	orderSvc   order.OrderService
}

func newTestWebApp(t *testing.T, ctx context.Context) testWebApp {
	basketDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	require.NoError(t, err)
	orderDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	require.NoError(t, err)
	catalogDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	require.NoError(t, err)

	awaitingValidationQueue, err := simplequeue.NewSimpleQueue(ctx)
	require.NoError(t, err)
	stockValidationResultQueue, err := simplequeue.NewSimpleQueue(ctx)
	require.NoError(t, err)
	stockConfirmedQueue, err := simplequeue.NewSimpleQueue(ctx)
	require.NoError(t, err)
	paymentResultQueue, err := simplequeue.NewSimpleQueue(ctx)
	require.NoError(t, err)
	catalogPriceQueue, err := simplequeue.NewSimpleQueue(ctx)
	require.NoError(t, err)
	catalogPaidQueue, err := simplequeue.NewSimpleQueue(ctx)
	require.NoError(t, err)

	basketSvc, err := basket.NewBasketServiceImpl(ctx, basketDB)
	require.NoError(t, err)
	catalogSvc, err := catalog.NewCatalogServiceImpl(ctx, catalogDB, catalogPriceQueue, awaitingValidationQueue, stockValidationResultQueue, catalogPaidQueue)
	require.NoError(t, err)
	orderSvc, err := order.NewOrderServiceImpl(ctx, orderDB, awaitingValidationQueue, stockValidationResultQueue, stockConfirmedQueue, paymentResultQueue, catalogPaidQueue)
	require.NoError(t, err)

	webapp, err := web.NewWebAppImpl(ctx, basketSvc, catalogSvc, orderSvc)
	require.NoError(t, err)

	_, err = basketSvc.UpdateBasket(ctx, basket.UpdateBasketRequest{Cart: basket.CustomerBasket{UserName: "swn"}})
	require.NoError(t, err)

	return testWebApp{webapp: webapp, basketSvc: basketSvc, catalogSvc: catalogSvc, orderSvc: orderSvc}
}

func makeItem(id int, name string, itemType catalog.CatalogType) catalog.CreateItemRequest {
	return catalog.CreateItemRequest{
		ID:            id,
		Name:          name,
		Price:         99.99,
		CatalogType:   itemType,
		CatalogTypeID: itemType.ID,
	}
}

func TestWebAppGetCatalogItems(t *testing.T) {
	ctx := context.Background()
	tw := newTestWebApp(t, ctx)

	electronics := catalog.CatalogType{ID: 1, Type: "Electronics"}
	clothing := catalog.CatalogType{ID: 2, Type: "Clothing"}

	_, err := tw.catalogSvc.CreateItem(ctx, makeItem(1001, "Laptop", electronics))
	require.NoError(t, err)
	_, err = tw.catalogSvc.CreateItem(ctx, makeItem(1002, "Shirt", clothing))
	require.NoError(t, err)

	products, err := tw.webapp.GetCatalogItems(ctx, 0, 0)
	assert.NoError(t, err)
	assert.True(t, len(products) >= 2)
}

func TestWebAppGetCatalogItemsByType(t *testing.T) {
	ctx := context.Background()
	tw := newTestWebApp(t, ctx)

	electronics := catalog.CatalogType{ID: 1, Type: "Electronics"}
	clothing := catalog.CatalogType{ID: 2, Type: "Clothing"}

	_, err := tw.catalogSvc.CreateItem(ctx, makeItem(2001, "Phone", electronics))
	require.NoError(t, err)
	_, err = tw.catalogSvc.CreateItem(ctx, makeItem(2002, "Jacket", clothing))
	require.NoError(t, err)

	products, err := tw.webapp.GetCatalogItems(ctx, 1, 0)
	assert.NoError(t, err)
	for _, p := range products {
		assert.Equal(t, 1, p.CatalogTypeID)
	}
}

func TestWebAppGetCatalogItemsByBrandAndType(t *testing.T) {
	ctx := context.Background()
	tw := newTestWebApp(t, ctx)

	electronics := catalog.CatalogType{ID: 1, Type: "Electronics"}
	brand := catalog.CatalogBrand{ID: 5, Brand: "Acme"}

	_, err := tw.catalogSvc.CreateItem(ctx, catalog.CreateItemRequest{
		ID: 2101, Name: "Acme Phone", Price: 499.99,
		CatalogType: electronics, CatalogTypeID: 1, CatalogBrand: brand, CatalogBrandID: 5,
	})
	require.NoError(t, err)

	products, err := tw.webapp.GetCatalogItems(ctx, 1, 5)
	assert.NoError(t, err)
	assert.True(t, len(products) >= 1)
	for _, p := range products {
		assert.Equal(t, 1, p.CatalogTypeID)
		assert.Equal(t, 5, p.CatalogBrandID)
	}
}

func TestWebAppAddToCart(t *testing.T) {
	ctx := context.Background()
	tw := newTestWebApp(t, ctx)

	electronics := catalog.CatalogType{ID: 1, Type: "Electronics"}
	_, err := tw.catalogSvc.CreateItem(ctx, makeItem(3001, "Headphones", electronics))
	require.NoError(t, err)

	err = tw.webapp.AddToCartAsync(ctx, 3001)
	assert.NoError(t, err)
}

func TestWebAppRemoveFromCart(t *testing.T) {
	ctx := context.Background()
	tw := newTestWebApp(t, ctx)

	electronics := catalog.CatalogType{ID: 1, Type: "Electronics"}
	_, err := tw.catalogSvc.CreateItem(ctx, makeItem(4001, "Keyboard", electronics))
	require.NoError(t, err)

	err = tw.webapp.AddToCartAsync(ctx, 4001)
	require.NoError(t, err)

	err = tw.webapp.SetQuantityAsync(ctx, 4001, 0)
	assert.NoError(t, err)
}

func TestWebAppCheckoutAndGetOrders(t *testing.T) {
	ctx := context.Background()
	tw := newTestWebApp(t, ctx)

	electronics := catalog.CatalogType{ID: 1, Type: "Electronics"}
	_, err := tw.catalogSvc.CreateItem(ctx, makeItem(5001, "Monitor", electronics))
	require.NoError(t, err)

	err = tw.webapp.AddToCartAsync(ctx, 5001)
	require.NoError(t, err)

	err = tw.webapp.CheckoutAsync(ctx)
	require.NoError(t, err)

	orders, err := tw.webapp.GetOrders(ctx)
	assert.NoError(t, err)
	assert.True(t, len(orders) >= 1)
}

func TestWebAppAddSameProductTwiceIncrementsQuantity(t *testing.T) {
	ctx := context.Background()
	tw := newTestWebApp(t, ctx)

	electronics := catalog.CatalogType{ID: 1, Type: "Electronics"}
	_, err := tw.catalogSvc.CreateItem(ctx, makeItem(6001, "Headset", electronics))
	require.NoError(t, err)

	err = tw.webapp.AddToCartAsync(ctx, 6001)
	require.NoError(t, err)
	err = tw.webapp.AddToCartAsync(ctx, 6001)
	require.NoError(t, err)

	resp, err := tw.basketSvc.GetBasket(ctx, basket.GetBasketRequest{UserName: "swn"})
	require.NoError(t, err)
	var headsetItems []basket.BasketItem
	for _, item := range resp.Cart.Items {
		if item.ProductID == 6001 {
			headsetItems = append(headsetItems, item)
		}
	}
	assert.Len(t, headsetItems, 1, "should be a single entry, not duplicates")
	assert.Equal(t, 2, headsetItems[0].Quantity)
}

func TestWebAppFullPaymentFlow(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	basketDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	require.NoError(t, err)
	orderDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	require.NoError(t, err)
	catalogDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	require.NoError(t, err)

	awaitingValidationQueue, err := simplequeue.NewSimpleQueue(ctx)
	require.NoError(t, err)
	stockValidationResultQueue, err := simplequeue.NewSimpleQueue(ctx)
	require.NoError(t, err)
	stockConfirmedQueue, err := simplequeue.NewSimpleQueue(ctx)
	require.NoError(t, err)
	paymentResultQueue, err := simplequeue.NewSimpleQueue(ctx)
	require.NoError(t, err)
	catalogPriceQueue, err := simplequeue.NewSimpleQueue(ctx)
	require.NoError(t, err)
	catalogPaidQueue, err := simplequeue.NewSimpleQueue(ctx)
	require.NoError(t, err)

	basketSvc, err := basket.NewBasketServiceImpl(ctx, basketDB)
	require.NoError(t, err)
	catalogSvc, err := catalog.NewCatalogServiceImpl(ctx, catalogDB, catalogPriceQueue, awaitingValidationQueue, stockValidationResultQueue, catalogPaidQueue)
	require.NoError(t, err)
	orderSvc, err := order.NewOrderServiceImpl(ctx, orderDB, awaitingValidationQueue, stockValidationResultQueue, stockConfirmedQueue, paymentResultQueue, catalogPaidQueue)
	require.NoError(t, err)
	paymentSvc, err := payment.NewPaymentServiceImpl(ctx, stockConfirmedQueue, paymentResultQueue)
	require.NoError(t, err)

	webapp, err := web.NewWebAppImpl(ctx, basketSvc, catalogSvc, orderSvc)
	require.NoError(t, err)

	_, err = basketSvc.UpdateBasket(ctx, basket.UpdateBasketRequest{Cart: basket.CustomerBasket{UserName: "swn"}})
	require.NoError(t, err)

	go orderSvc.Init(ctx)
	go catalogSvc.Init(ctx)
	go paymentSvc.Run(ctx)

	electronics := catalog.CatalogType{ID: 1, Type: "Electronics"}
	_, err = catalogSvc.CreateItem(ctx, catalog.CreateItemRequest{
		ID: 7001, Name: "Payment Test Item", Price: 199.99,
		CatalogType: electronics, CatalogTypeID: 1,
	})
	require.NoError(t, err)

	err = webapp.AddToCartAsync(ctx, 7001)
	require.NoError(t, err)
	err = webapp.CheckoutAsync(ctx)
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	orders, err := webapp.GetOrders(ctx)
	require.NoError(t, err)
	require.True(t, len(orders) >= 1)

	assert.Equal(t, order.Paid, orders[0].Status, "order should reach Paid after payment processing")
	assert.Len(t, orders[0].OrderItems, 1, "order should contain the basket item")
	assert.Equal(t, 199.99, orders[0].OrderItems[0].Price)
}
