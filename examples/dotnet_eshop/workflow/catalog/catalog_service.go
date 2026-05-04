package catalog

import (
	"context"
	"fmt"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/workflow/order"
)

type CatalogService interface {
	// Routes for querying catalog items
	GetAllItems(ctx context.Context) (GetAllItemsResponse, error)
	GetItemsByIds(ctx context.Context, query GetItemsByIDsRequest) (GetItemsByIDsResponse, error)
	GetItemById(ctx context.Context, query GetItemByIDRequest) (GetItemByIDResponse, error)
	GetItemsByName(ctx context.Context, query GetItemsByNameRequest) (GetItemsByNameResponse, error)
	GetItemPictureById(ctx context.Context, query GetItemPictureByIdRequest) (GetItemPictureByIdResponse, error)

	// Routes for resolving catalog items by type and brand
	GetItemsByBrandAndTypeId(ctx context.Context, query GetItemsByBrandAndTypeRequest) (GetItemsByBrandAndTypeResponse, error)
	GetItemsByBrandId(ctx context.Context, query GetItemsByBrandRequest) (GetItemsByBrandResponse, error)
	GetCatalogTypes(ctx context.Context) (GetCatalogTypesResponse, error)
	GetCatalogBrands(ctx context.Context) (GetCatalogBrandsResponse, error)

	// Routes for modifying catalog items
	UpdateItem(ctx context.Context, productToUpdate CatalogItem) (UpdateItemResponse, error)
	CreateItem(ctx context.Context, command CreateItemRequest) (CreateItemResponse, error)
	DeleteItemById(ctx context.Context, command DeleteItemRequest) error

	// Stock management API
	AddStock(ctx context.Context, command AddStockRequest) (AddStockResponse, error)
	RemoveStock(ctx context.Context, command RemoveStockRequest) (RemoveStockResponse, error)

	// Background integration event handlers:
	// - OrderStatusChangedToAwaitingValidationIntegrationEventHandler (validates stock per item, replies confirmed/rejected)
	// - OrderStatusChangedToPaidIntegrationEventHandler (removes stock for paid order items)
	Init(ctx context.Context) error
}

type CatalogServiceImpl struct {
	database                   backend.NoSQLDatabase
	queue                      backend.Queue // catalog_price_queue: product price changed events
	awaitingValidationQueue    backend.Queue // order -> catalog: stock validation requests
	stockValidationResultQueue backend.Queue // catalog -> order: stock validation results
	paidQueue                  backend.Queue // order -> catalog: order paid events
}

func NewCatalogServiceImpl(ctx context.Context, database backend.NoSQLDatabase, queue backend.Queue, awaitingValidationQueue backend.Queue, stockValidationResultQueue backend.Queue, paidQueue backend.Queue) (CatalogService, error) {
	s := &CatalogServiceImpl{
		database:                   database,
		queue:                      queue,
		awaitingValidationQueue:    awaitingValidationQueue,
		stockValidationResultQueue: stockValidationResultQueue,
		paidQueue:                  paidQueue,
	}
	return s, nil
}

// Routes for querying catalog items

func (s *CatalogServiceImpl) GetAllItems(ctx context.Context) (GetAllItemsResponse, error) {
	products, err := s.getAll(ctx)
	if err != nil {
		return GetAllItemsResponse{}, err
	}
	return GetAllItemsResponse{products}, nil
}

func (s *CatalogServiceImpl) GetItemsByIds(ctx context.Context, request GetItemsByIDsRequest) (GetItemsByIDsResponse, error) {
	items, err := s.getMany(ctx, request.IDs)
	if err != nil {
		return GetItemsByIDsResponse{}, err
	}
	return GetItemsByIDsResponse{items}, nil
}

func (s *CatalogServiceImpl) GetItemById(ctx context.Context, request GetItemByIDRequest) (GetItemByIDResponse, error) {
	item, err := s.get(ctx, request.ID)
	if err != nil {
		return GetItemByIDResponse{}, err
	}
	return GetItemByIDResponse{item}, nil
}

func (s *CatalogServiceImpl) GetItemPictureById(ctx context.Context, query GetItemPictureByIdRequest) (GetItemPictureByIdResponse, error) {
	item, err := s.get(ctx, query.ID)
	if err != nil {
		return GetItemPictureByIdResponse{}, err
	}
	return GetItemPictureByIdResponse{PictureFileName: item.PictureFileName}, nil
}

func (s *CatalogServiceImpl) GetItemsByName(ctx context.Context, query GetItemsByNameRequest) (GetItemsByNameResponse, error) {
	collection, err := s.database.GetCollection(ctx, "catalog_db", "item")
	if err != nil {
		return GetItemsByNameResponse{}, err
	}
	filter := bson.D{{Key: "Name", Value: query.Name}}
	cursor, err := collection.FindMany(ctx, filter)
	if err != nil {
		return GetItemsByNameResponse{}, err
	}
	var items []CatalogItem
	err = cursor.All(ctx, &items)
	if err != nil {
		return GetItemsByNameResponse{}, err
	}
	return GetItemsByNameResponse{Items: items}, nil
}

// Routes for resolving catalog items by type and brand

func (s *CatalogServiceImpl) GetItemsByBrandAndTypeId(ctx context.Context, query GetItemsByBrandAndTypeRequest) (GetItemsByBrandAndTypeResponse, error) {
	collection, err := s.database.GetCollection(ctx, "catalog_db", "item")
	if err != nil {
		return GetItemsByBrandAndTypeResponse{}, err
	}
	filter := bson.D{
		{Key: "CatalogTypeID", Value: query.TypeId},
		{Key: "CatalogBrandID", Value: query.BrandId},
	}
	cursor, err := collection.FindMany(ctx, filter)
	if err != nil {
		return GetItemsByBrandAndTypeResponse{}, err
	}
	var items []CatalogItem
	err = cursor.All(ctx, &items)
	if err != nil {
		return GetItemsByBrandAndTypeResponse{}, err
	}
	return GetItemsByBrandAndTypeResponse{Items: items}, nil
}

func (s *CatalogServiceImpl) GetItemsByBrandId(ctx context.Context, query GetItemsByBrandRequest) (GetItemsByBrandResponse, error) {
	collection, err := s.database.GetCollection(ctx, "catalog_db", "item")
	if err != nil {
		return GetItemsByBrandResponse{}, err
	}
	filter := bson.D{{Key: "CatalogBrandID", Value: query.BrandId}}
	cursor, err := collection.FindMany(ctx, filter)
	if err != nil {
		return GetItemsByBrandResponse{}, err
	}
	var items []CatalogItem
	err = cursor.All(ctx, &items)
	if err != nil {
		return GetItemsByBrandResponse{}, err
	}
	return GetItemsByBrandResponse{Items: items}, nil
}

func (s *CatalogServiceImpl) GetCatalogTypes(ctx context.Context) (GetCatalogTypesResponse, error) {
	products, err := s.getAll(ctx)
	if err != nil {
		return GetCatalogTypesResponse{}, err
	}
	seen := make(map[int]bool)
	var types []CatalogType
	for _, p := range products {
		if p.CatalogTypeID != 0 && !seen[p.CatalogTypeID] {
			seen[p.CatalogTypeID] = true
			types = append(types, p.CatalogType)
		}
	}
	return GetCatalogTypesResponse{Types: types}, nil
}

func (s *CatalogServiceImpl) GetCatalogBrands(ctx context.Context) (GetCatalogBrandsResponse, error) {
	products, err := s.getAll(ctx)
	if err != nil {
		return GetCatalogBrandsResponse{}, err
	}
	seen := make(map[int]bool)
	var brands []CatalogBrand
	for _, p := range products {
		if p.CatalogBrandID != 0 && !seen[p.CatalogBrandID] {
			seen[p.CatalogBrandID] = true
			brands = append(brands, p.CatalogBrand)
		}
	}
	return GetCatalogBrandsResponse{Brands: brands}, nil
}

// Routes for modifying catalog items

func (s *CatalogServiceImpl) UpdateItem(ctx context.Context, productToUpdate CatalogItem) (UpdateItemResponse, error) {
	catalogItem, err := s.get(ctx, productToUpdate.ID)
	if err != nil {
		return UpdateItemResponse{}, err
	}

	oldPrice := catalogItem.Price

	catalogItem.Name = productToUpdate.Name
	catalogItem.Description = productToUpdate.Description
	catalogItem.Price = productToUpdate.Price
	catalogItem.PictureFileName = productToUpdate.PictureFileName
	catalogItem.AvailableStock = productToUpdate.AvailableStock
	catalogItem.CatalogBrandID = productToUpdate.CatalogBrandID
	catalogItem.CatalogTypeID = productToUpdate.CatalogTypeID

	if oldPrice != catalogItem.Price {
		priceChangedEvent := ProductPriceChangedEvent{
			CatalogItemID: catalogItem.ID,
			NewPrice:      catalogItem.Price,
			OldPrice:      oldPrice,
		}

		err := s.update(ctx, catalogItem)
		if err != nil {
			return UpdateItemResponse{}, fmt.Errorf("update catalog item: %w", err)
		}

		s.queue.Push(ctx, priceChangedEvent)
	} else {
		err := s.update(ctx, catalogItem)
		if err != nil {
			return UpdateItemResponse{}, fmt.Errorf("update catalog item: %w", err)
		}
	}

	return UpdateItemResponse{ID: catalogItem.ID}, nil
}

func (s *CatalogServiceImpl) CreateItem(ctx context.Context, request CreateItemRequest) (CreateItemResponse, error) {
	item := CatalogItem{
		ID:                request.ID,
		Name:              request.Name,
		Description:       request.Description,
		Price:             request.Price,
		PictureFileName:   request.PriceFileName,
		CatalogTypeID:     request.CatalogTypeID,
		CatalogType:       request.CatalogType,
		CatalogBrandID:    request.CatalogBrandID,
		CatalogBrand:      request.CatalogBrand,
		AvailableStock:    request.AvailableStock,
		RestockThreshold:  request.RestockThreshold,
		MaxStockThreshold: request.MaxStockThreshold,
	}
	err := s.save(ctx, item)
	if err != nil {
		return CreateItemResponse{}, err
	}
	return CreateItemResponse{item}, nil
}

func (s *CatalogServiceImpl) DeleteItemById(ctx context.Context, request DeleteItemRequest) error {
	return s.remove(ctx, request.ID)
}

// Stock management (integration event handlers)

func (s *CatalogServiceImpl) RemoveStock(ctx context.Context, command RemoveStockRequest) (RemoveStockResponse, error) {
	item, err := s.get(ctx, command.ProductID)
	if err != nil {
		return RemoveStockResponse{}, err
	}
	item.AvailableStock -= command.Quantity
	err = s.update(ctx, item)
	if err != nil {
		return RemoveStockResponse{}, err
	}
	return RemoveStockResponse{Depleted: item.AvailableStock <= item.RestockThreshold}, nil
}

func (s *CatalogServiceImpl) AddStock(ctx context.Context, command AddStockRequest) (AddStockResponse, error) {
	item, err := s.get(ctx, command.ProductID)
	if err != nil {
		return AddStockResponse{}, err
	}
	item.AvailableStock += command.Quantity
	if item.MaxStockThreshold > 0 && item.AvailableStock > item.MaxStockThreshold {
		item.AvailableStock = item.MaxStockThreshold
	}
	err = s.update(ctx, item)
	if err != nil {
		return AddStockResponse{}, err
	}
	return AddStockResponse{AvailableStock: item.AvailableStock}, nil
}

// Init runs two concurrent background loops matching the original Catalog.API integration event handlers.
func (s *CatalogServiceImpl) Init(ctx context.Context) error {
	go s.processAwaitingValidationEvents(ctx)
	return s.processOrderPaidEvents(ctx)
}

// processAwaitingValidationEvents handles OrderStatusChangedToAwaitingValidationIntegrationEvent.
// For each item checks AvailableStock >= requested units, then publishes confirmed/rejected result.
func (s *CatalogServiceImpl) processAwaitingValidationEvents(ctx context.Context) error {
	for {
		var event order.OrderStatusChangedToAwaitingValidationEvent
		ok, err := s.awaitingValidationQueue.Pop(ctx, &event)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		confirmed := true
		for _, stockItem := range event.OrderStockItems {
			item, err := s.get(ctx, stockItem.ProductId)
			if err != nil || item.AvailableStock < stockItem.Units {
				confirmed = false
				break
			}
		}
		s.stockValidationResultQueue.Push(ctx, order.OrderStockValidationResultEvent{
			OrderId:   event.OrderId,
			Confirmed: confirmed,
		})
	}
}

// processOrderPaidEvents handles OrderStatusChangedToPaidIntegrationEvent.
// Removes stock for each item in the paid order.
func (s *CatalogServiceImpl) processOrderPaidEvents(ctx context.Context) error {
	for {
		var event order.OrderPaidEvent
		ok, err := s.paidQueue.Pop(ctx, &event)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		for _, item := range event.OrderItems {
			s.RemoveStock(ctx, RemoveStockRequest{ProductID: item.ProductID, Quantity: item.Quantity})
		}
	}
}

func (s *CatalogServiceImpl) save(ctx context.Context, item CatalogItem) error {
	collection, err := s.database.GetCollection(ctx, "catalog_db", "item")
	if err != nil {
		return err
	}
	return collection.InsertOne(ctx, item)
}

func (s *CatalogServiceImpl) update(ctx context.Context, item CatalogItem) error {
	collection, err := s.database.GetCollection(ctx, "catalog_db", "item")
	if err != nil {
		return err
	}
	filter := bson.D{{Key: "ID", Value: item.ID}}
	_, err = collection.ReplaceOne(ctx, filter, item)
	return err
}

func (s *CatalogServiceImpl) remove(ctx context.Context, id int) error {
	collection, err := s.database.GetCollection(ctx, "catalog_db", "item")
	if err != nil {
		return err
	}
	filter := bson.D{{Key: "ID", Value: id}}
	return collection.DeleteOne(ctx, filter)
}

func (s *CatalogServiceImpl) get(ctx context.Context, id int) (CatalogItem, error) {
	collection, err := s.database.GetCollection(ctx, "catalog_db", "item")
	if err != nil {
		return CatalogItem{}, err
	}
	filter := bson.D{{Key: "ID", Value: id}}
	cursor, err := collection.FindOne(ctx, filter)
	if err != nil {
		return CatalogItem{}, err
	}
	var item CatalogItem
	ok, err := cursor.One(ctx, &item)
	if err != nil {
		return CatalogItem{}, err
	}
	if !ok {
		return CatalogItem{}, fmt.Errorf("item not found for id (%d)", id)
	}
	return item, nil
}

func (s *CatalogServiceImpl) getMany(ctx context.Context, ids []int) ([]CatalogItem, error) {
	collection, err := s.database.GetCollection(ctx, "catalog_db", "item")
	if err != nil {
		return nil, err
	}
	filter := bson.D{
		{Key: "ID", Value: bson.D{
			{Key: "$in", Value: ids},
		}},
	}
	cursor, err := collection.FindMany(ctx, filter)
	if err != nil {
		return nil, err
	}
	var products []CatalogItem
	err = cursor.All(ctx, &products)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (s *CatalogServiceImpl) getAll(ctx context.Context) ([]CatalogItem, error) {
	collection, err := s.database.GetCollection(ctx, "catalog_db", "item")
	if err != nil {
		return nil, err
	}
	cursor, err := collection.FindMany(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	var products []CatalogItem
	err = cursor.All(ctx, &products)
	if err != nil {
		return nil, err
	}
	return products, nil
}
