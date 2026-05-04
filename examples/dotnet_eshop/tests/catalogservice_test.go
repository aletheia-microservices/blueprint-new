package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/workflow/catalog"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplequeue"
	"github.com/stretchr/testify/assert"
)

var catalogServiceRegistry = registry.NewServiceRegistry[catalog.CatalogService]("catalog_service")

func init() {
	catalogServiceRegistry.Register("local", func(ctx context.Context) (catalog.CatalogService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		priceQueue, err := simplequeue.NewSimpleQueue(ctx)
		if err != nil {
			return nil, err
		}
		awaitingValidationQueue, err := simplequeue.NewSimpleQueue(ctx)
		if err != nil {
			return nil, err
		}
		stockValidationResultQueue, err := simplequeue.NewSimpleQueue(ctx)
		if err != nil {
			return nil, err
		}
		paidQueue, err := simplequeue.NewSimpleQueue(ctx)
		if err != nil {
			return nil, err
		}
		return catalog.NewCatalogServiceImpl(ctx, db, priceQueue, awaitingValidationQueue, stockValidationResultQueue, paidQueue)
	})
}

func makeElectronicsType() catalog.CatalogType {
	return catalog.CatalogType{ID: 1, Type: "Electronics"}
}

func makeClothingType() catalog.CatalogType {
	return catalog.CatalogType{ID: 2, Type: "Clothing"}
}

func makeDefaultBrand() catalog.CatalogBrand {
	return catalog.CatalogBrand{ID: 1, Brand: "Generic"}
}

func TestCatalogServiceCreateItem(t *testing.T) {
	ctx := context.Background()
	catalogService, err := catalogServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	resp, err := catalogService.CreateItem(ctx, catalog.CreateItemRequest{
		ID:             101,
		Name:           "Test Laptop",
		Description:    "A test laptop",
		Price:          999.99,
		CatalogType:    makeElectronicsType(),
		CatalogTypeID:  1,
		CatalogBrand:   makeDefaultBrand(),
		CatalogBrandID: 1,
		AvailableStock: 10,
	})
	assert.NoError(t, err)
	assert.Equal(t, "Test Laptop", resp.Item.Name)
	assert.Equal(t, 999.99, resp.Item.Price)
}

func TestCatalogServiceGetItemById(t *testing.T) {
	ctx := context.Background()
	catalogService, err := catalogServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	_, err = catalogService.CreateItem(ctx, catalog.CreateItemRequest{
		ID:            201,
		Name:          "Get Laptop",
		Price:         899.99,
		CatalogType:   makeElectronicsType(),
		CatalogTypeID: 1,
	})
	assert.NoError(t, err)

	resp, err := catalogService.GetItemById(ctx, catalog.GetItemByIDRequest{ID: 201})
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.Item.ID)
	assert.Equal(t, "Get Laptop", resp.Item.Name)
}

func TestCatalogServiceGetItemsByIds(t *testing.T) {
	ctx := context.Background()
	catalogService, err := catalogServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	_, err = catalogService.CreateItem(ctx, catalog.CreateItemRequest{
		ID: 301, Name: "Item A", Price: 10.0, CatalogType: makeElectronicsType(), CatalogTypeID: 1,
	})
	assert.NoError(t, err)
	_, err = catalogService.CreateItem(ctx, catalog.CreateItemRequest{
		ID: 302, Name: "Item B", Price: 20.0, CatalogType: makeClothingType(), CatalogTypeID: 2,
	})
	assert.NoError(t, err)

	resp, err := catalogService.GetItemsByIds(ctx, catalog.GetItemsByIDsRequest{IDs: []int{301, 302}})
	assert.NoError(t, err)
	assert.Len(t, resp.Item, 2)
}

func TestCatalogServiceGetItemsByName(t *testing.T) {
	ctx := context.Background()
	catalogService, err := catalogServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	_, err = catalogService.CreateItem(ctx, catalog.CreateItemRequest{
		ID: 401, Name: "Unique Shirt", Price: 29.99, CatalogType: makeClothingType(), CatalogTypeID: 2,
	})
	assert.NoError(t, err)
	_, err = catalogService.CreateItem(ctx, catalog.CreateItemRequest{
		ID: 402, Name: "Unique Shirt", Price: 39.99, CatalogType: makeClothingType(), CatalogTypeID: 2,
	})
	assert.NoError(t, err)

	resp, err := catalogService.GetItemsByName(ctx, catalog.GetItemsByNameRequest{Name: "Unique Shirt"})
	assert.NoError(t, err)
	assert.True(t, len(resp.Items) >= 2)
	for _, item := range resp.Items {
		assert.Equal(t, "Unique Shirt", item.Name)
	}
}

func TestCatalogServiceGetAllItems(t *testing.T) {
	ctx := context.Background()
	catalogService, err := catalogServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	_, err = catalogService.CreateItem(ctx, catalog.CreateItemRequest{
		ID: 501, Name: "All Item A", Price: 5.0, CatalogType: makeElectronicsType(), CatalogTypeID: 1,
	})
	assert.NoError(t, err)
	_, err = catalogService.CreateItem(ctx, catalog.CreateItemRequest{
		ID: 502, Name: "All Item B", Price: 6.0, CatalogType: makeClothingType(), CatalogTypeID: 2,
	})
	assert.NoError(t, err)

	resp, err := catalogService.GetAllItems(ctx)
	assert.NoError(t, err)
	assert.True(t, len(resp.Items) >= 2)
}

func TestCatalogServiceGetItemsByBrandAndType(t *testing.T) {
	ctx := context.Background()
	catalogService, err := catalogServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	brand := catalog.CatalogBrand{ID: 10, Brand: "Acme"}
	_, err = catalogService.CreateItem(ctx, catalog.CreateItemRequest{
		ID: 701, Name: "Acme Laptop", Price: 1200.0,
		CatalogType: makeElectronicsType(), CatalogTypeID: 1, CatalogBrand: brand, CatalogBrandID: 10,
	})
	assert.NoError(t, err)
	_, err = catalogService.CreateItem(ctx, catalog.CreateItemRequest{
		ID: 702, Name: "Acme Shirt", Price: 35.0,
		CatalogType: makeClothingType(), CatalogTypeID: 2, CatalogBrand: brand, CatalogBrandID: 10,
	})
	assert.NoError(t, err)

	resp, err := catalogService.GetItemsByBrandAndTypeId(ctx, catalog.GetItemsByBrandAndTypeRequest{TypeId: 1, BrandId: 10})
	assert.NoError(t, err)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, "Acme Laptop", resp.Items[0].Name)
}

func TestCatalogServiceGetItemsByBrandId(t *testing.T) {
	ctx := context.Background()
	catalogService, err := catalogServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	brand := catalog.CatalogBrand{ID: 20, Brand: "BrandX"}
	_, err = catalogService.CreateItem(ctx, catalog.CreateItemRequest{
		ID: 801, Name: "BrandX Item A", Price: 10.0,
		CatalogType: makeElectronicsType(), CatalogTypeID: 1, CatalogBrand: brand, CatalogBrandID: 20,
	})
	assert.NoError(t, err)
	_, err = catalogService.CreateItem(ctx, catalog.CreateItemRequest{
		ID: 802, Name: "BrandX Item B", Price: 20.0,
		CatalogType: makeClothingType(), CatalogTypeID: 2, CatalogBrand: brand, CatalogBrandID: 20,
	})
	assert.NoError(t, err)

	resp, err := catalogService.GetItemsByBrandId(ctx, catalog.GetItemsByBrandRequest{BrandId: 20})
	assert.NoError(t, err)
	assert.True(t, len(resp.Items) >= 2)
	for _, item := range resp.Items {
		assert.Equal(t, 20, item.CatalogBrandID)
	}
}

func TestCatalogServiceGetCatalogTypes(t *testing.T) {
	ctx := context.Background()
	catalogService, err := catalogServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	_, err = catalogService.CreateItem(ctx, catalog.CreateItemRequest{
		ID: 901, Name: "Type Test Item A", Price: 5.0, CatalogType: makeElectronicsType(), CatalogTypeID: 1,
	})
	assert.NoError(t, err)
	_, err = catalogService.CreateItem(ctx, catalog.CreateItemRequest{
		ID: 902, Name: "Type Test Item B", Price: 6.0, CatalogType: makeClothingType(), CatalogTypeID: 2,
	})
	assert.NoError(t, err)

	resp, err := catalogService.GetCatalogTypes(ctx)
	assert.NoError(t, err)
	assert.True(t, len(resp.Types) >= 2)
	typeNames := make(map[string]bool)
	for _, ct := range resp.Types {
		typeNames[ct.Type] = true
	}
	assert.True(t, typeNames["Electronics"])
	assert.True(t, typeNames["Clothing"])
}

func TestCatalogServiceGetCatalogBrands(t *testing.T) {
	ctx := context.Background()
	catalogService, err := catalogServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	brand := catalog.CatalogBrand{ID: 30, Brand: "UniqueTestBrand"}
	_, err = catalogService.CreateItem(ctx, catalog.CreateItemRequest{
		ID: 1001, Name: "Brand Test Item", Price: 5.0,
		CatalogType: makeElectronicsType(), CatalogTypeID: 1, CatalogBrand: brand, CatalogBrandID: 30,
	})
	assert.NoError(t, err)

	resp, err := catalogService.GetCatalogBrands(ctx)
	assert.NoError(t, err)
	found := false
	for _, b := range resp.Brands {
		if b.Brand == "UniqueTestBrand" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestCatalogServiceRemoveStock(t *testing.T) {
	ctx := context.Background()
	catalogService, err := catalogServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	_, err = catalogService.CreateItem(ctx, catalog.CreateItemRequest{
		ID: 1101, Name: "Stockable Item", Price: 50.0,
		CatalogType: makeElectronicsType(), CatalogTypeID: 1,
		AvailableStock: 100, RestockThreshold: 10, MaxStockThreshold: 200,
	})
	assert.NoError(t, err)

	resp, err := catalogService.RemoveStock(ctx, catalog.RemoveStockRequest{ProductID: 1101, Quantity: 5})
	assert.NoError(t, err)
	assert.False(t, resp.Depleted)

	item, err := catalogService.GetItemById(ctx, catalog.GetItemByIDRequest{ID: 1101})
	assert.NoError(t, err)
	assert.Equal(t, 95, item.Item.AvailableStock)

	resp, err = catalogService.RemoveStock(ctx, catalog.RemoveStockRequest{ProductID: 1101, Quantity: 90})
	assert.NoError(t, err)
	assert.True(t, resp.Depleted)
}

func TestCatalogServiceAddStock(t *testing.T) {
	ctx := context.Background()
	catalogService, err := catalogServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	_, err = catalogService.CreateItem(ctx, catalog.CreateItemRequest{
		ID: 1201, Name: "Restockable Item", Price: 50.0,
		CatalogType: makeElectronicsType(), CatalogTypeID: 1,
		AvailableStock: 10, MaxStockThreshold: 50,
	})
	assert.NoError(t, err)

	resp, err := catalogService.AddStock(ctx, catalog.AddStockRequest{ProductID: 1201, Quantity: 20})
	assert.NoError(t, err)
	assert.Equal(t, 30, resp.AvailableStock)

	respCapped, err := catalogService.AddStock(ctx, catalog.AddStockRequest{ProductID: 1201, Quantity: 100})
	assert.NoError(t, err)
	assert.Equal(t, 50, respCapped.AvailableStock)
}

func TestCatalogServiceUpdateItemPublishesPriceChange(t *testing.T) {
	ctx := context.Background()
	db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
	assert.NoError(t, err)
	priceQueue, err := simplequeue.NewSimpleQueue(ctx)
	assert.NoError(t, err)
	awaitingValidationQueue, err := simplequeue.NewSimpleQueue(ctx)
	assert.NoError(t, err)
	stockValidationResultQueue, err := simplequeue.NewSimpleQueue(ctx)
	assert.NoError(t, err)
	paidQueue, err := simplequeue.NewSimpleQueue(ctx)
	assert.NoError(t, err)
	catalogService, err := catalog.NewCatalogServiceImpl(ctx, db, priceQueue, awaitingValidationQueue, stockValidationResultQueue, paidQueue)
	assert.NoError(t, err)

	_, err = catalogService.CreateItem(ctx, catalog.CreateItemRequest{
		ID: 1301, Name: "Priceable Item", Price: 100.0,
		CatalogType: makeElectronicsType(), CatalogTypeID: 1,
	})
	assert.NoError(t, err)

	_, err = catalogService.UpdateItem(ctx, catalog.CatalogItem{
		ID: 1301, Name: "Priceable Item", Price: 200.0,
		CatalogType: makeElectronicsType(), CatalogTypeID: 1,
	})
	assert.NoError(t, err)

	var event catalog.ProductPriceChangedEvent
	ok, err := priceQueue.Pop(ctx, &event)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 1301, event.CatalogItemID)
	assert.Equal(t, 100.0, event.OldPrice)
	assert.Equal(t, 200.0, event.NewPrice)
}

func TestCatalogServiceDeleteItemById(t *testing.T) {
	ctx := context.Background()
	catalogService, err := catalogServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	_, err = catalogService.CreateItem(ctx, catalog.CreateItemRequest{
		ID: 601, Name: "Delete Me", Price: 1.99, CatalogType: makeElectronicsType(), CatalogTypeID: 1,
	})
	assert.NoError(t, err)

	err = catalogService.DeleteItemById(ctx, catalog.DeleteItemRequest{ID: 601})
	assert.NoError(t, err)

	_, err = catalogService.GetItemById(ctx, catalog.GetItemByIDRequest{ID: 601})
	assert.Error(t, err)
}
