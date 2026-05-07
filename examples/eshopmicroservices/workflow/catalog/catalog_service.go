package catalog

import (
	"context"
	"fmt"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type CatalogService interface {
	CreateProduct(ctx context.Context, command CreateProductCommand) (CreateProductResponse, error)
	DeleteProduct(ctx context.Context, command DeleteProductCommand) error
	GetProductById(ctx context.Context, query GetProductByIdQuery) (GetProductByIdResponse, error)
	GetProductByCategory(ctx context.Context, query GetProductByCategoryQuery) (GetProductByCategoryResponse, error)
	GetProducts(ctx context.Context) (GetProductsResponse, error)
}

type CatalogServiceImpl struct {
	database backend.NoSQLDatabase
}

func NewCatalogServiceImpl(ctx context.Context, database backend.NoSQLDatabase) (CatalogService, error) {
	s := &CatalogServiceImpl{
		database: database,
	}
	return s, nil
}

func (s *CatalogServiceImpl) CreateProduct(ctx context.Context, command CreateProductCommand) (CreateProductResponse, error) {
	product := Product{
		Id:          uuid.NewString(),
		Name:        command.Name,
		Category:    command.Category,
		Description: command.Description,
		ImageFile:   command.ImageFile,
		Price:       command.Price,
	}
	err := s.store(ctx, product)
	if err != nil {
		return CreateProductResponse{}, err
	}
	return CreateProductResponse{product}, nil
}

func (s *CatalogServiceImpl) DeleteProduct(ctx context.Context, command DeleteProductCommand) error {
	return s.delete(ctx, command.Id)
}

func (s *CatalogServiceImpl) GetProductById(ctx context.Context, query GetProductByIdQuery) (GetProductByIdResponse, error) {
	product, err := s.load(ctx, query.Id)
	if err != nil {
		return GetProductByIdResponse{}, err
	}
	return GetProductByIdResponse{product}, nil
}

func (s *CatalogServiceImpl) GetProductByCategory(ctx context.Context, query GetProductByCategoryQuery) (GetProductByCategoryResponse, error) {
	product, err := s.loadByCategory(ctx, query.Category)
	if err != nil {
		return GetProductByCategoryResponse{}, err
	}
	return GetProductByCategoryResponse{product}, nil
}

func (s *CatalogServiceImpl) GetProducts(ctx context.Context) (GetProductsResponse, error) {
	products, err := s.loadAll(ctx)
	if err != nil {
		return GetProductsResponse{}, err
	}
	return GetProductsResponse{products}, nil
}

func (s *CatalogServiceImpl) store(ctx context.Context, product Product) error {
	collection, err := s.database.GetCollection(ctx, "catalog_db", "product")
	if err != nil {
		return err
	}
	return collection.InsertOne(ctx, product)
}

func (s *CatalogServiceImpl) delete(ctx context.Context, id string) error {
	collection, err := s.database.GetCollection(ctx, "catalog_db", "product")
	if err != nil {
		return err
	}
	filter := bson.D{{Key: "Id", Value: id}}
	return collection.DeleteOne(ctx, filter)
}

func (s *CatalogServiceImpl) load(ctx context.Context, id string) (Product, error) {
	collection, err := s.database.GetCollection(ctx, "catalog_db", "product")
	if err != nil {
		return Product{}, err
	}
	filter := bson.D{{Key: "Id", Value: id}}
	cursor, err := collection.FindOne(ctx, filter)
	if err != nil {
		return Product{}, err
	}
	var product Product
	ok, err := cursor.One(ctx, &product)
	if err != nil {
		return Product{}, err
	}
	if !ok {
		return Product{}, fmt.Errorf("product not found for id (%s)", id)
	}
	return product, nil
}

func (s *CatalogServiceImpl) loadByCategory(ctx context.Context, category string) (Product, error) {
	collection, err := s.database.GetCollection(ctx, "catalog_db", "product")
	if err != nil {
		return Product{}, err
	}
	filter := bson.D{{Key: "Category", Value: category}}
	cursor, err := collection.FindOne(ctx, filter)
	if err != nil {
		return Product{}, err
	}
	var product Product
	ok, err := cursor.One(ctx, &product)
	if err != nil {
		return Product{}, err
	}
	if !ok {
		return Product{}, fmt.Errorf("product not found for category (%s)", category)
	}
	return product, nil
}

func (s *CatalogServiceImpl) loadAll(ctx context.Context) ([]Product, error) {
	collection, err := s.database.GetCollection(ctx, "catalog_db", "product")
	if err != nil {
		return nil, err
	}
	cursor, err := collection.FindMany(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	var products []Product
	err = cursor.All(ctx, &products)
	if err != nil {
		return nil, err
	}
	return products, nil
}
