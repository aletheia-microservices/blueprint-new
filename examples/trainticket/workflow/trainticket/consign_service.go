package trainticket

import (
	"context"
	"fmt"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type ConsignService interface {
	InsertConsign(ctx context.Context, consignRequest ConsignRequest) (ConsignRecord, error)
	UpdateConsign(ctx context.Context, consignRequest ConsignRequest) (ConsignRecord, error)
	FindByOrderId(ctx context.Context, orderID string) (ConsignRecord, error)
	FindByAccountId(ctx context.Context, accountID string) (ConsignRecord, error)
	FindByConsignee(ctx context.Context, consignee string) (ConsignRecord, error)
}

type ConsignServiceImpl struct {
	consignPriceService ConsignPriceService
	consignDB           backend.NoSQLDatabase
}

func NewConsignServiceImpl(ctx context.Context, consignPriceService ConsignPriceService, consignDB backend.NoSQLDatabase) (ConsignService, error) {
	return &ConsignServiceImpl{consignPriceService: consignPriceService, consignDB: consignDB}, nil
}

func (c *ConsignServiceImpl) InsertConsign(ctx context.Context, consignRequest ConsignRequest) (ConsignRecord, error) {
	price, err := c.consignPriceService.GetPriceByWeightAndRegion(ctx, consignRequest.Weight, consignRequest.IsWithin)
	if err != nil {
		return ConsignRecord{}, err
	}

	consignRecord := ConsignRecord{
		ID:         uuid.New().String(),
		OrderID:    consignRequest.OrderID,
		AccountID:  consignRequest.AccountID,
		HandleDate: consignRequest.HandleDate,
		TargetDate: consignRequest.TargetDate,
		FromPlace:  consignRequest.From,
		ToPlace:    consignRequest.To,
		Consignee:  consignRequest.Consignee,
		Phone:      consignRequest.Phone,
		Weight:     consignRequest.Weight,
		Price:      price,
	}

	collection, err := c.consignDB.GetCollection(ctx, "consign_db", "consign_record")
	if err != nil {
		return ConsignRecord{}, err
	}
	err = collection.InsertOne(ctx, consignRecord)
	if err != nil {
		return ConsignRecord{}, err
	}

	return consignRecord, nil
}

func (c *ConsignServiceImpl) UpdateConsign(ctx context.Context, consignRequest ConsignRequest) (ConsignRecord, error) {
	collection, err := c.consignDB.GetCollection(ctx, "consign_db", "consign_record")
	if err != nil {
		return ConsignRecord{}, err
	}

	filter := bson.D{{Key: "ID", Value: consignRequest.ID}}
	cursor, err := collection.FindOne(ctx, filter)
	if err != nil {
		return ConsignRecord{}, err
	}

	var originalRecord ConsignRecord
	ok, err := cursor.One(ctx, &originalRecord)
	if err != nil {
		return ConsignRecord{}, err
	}
	if !ok {
		return ConsignRecord{}, fmt.Errorf("consign not found for ID (%s)", consignRequest.ID)
	}

	// recalculate price
	if originalRecord.Weight != consignRequest.Weight {
		price, err := c.consignPriceService.GetPriceByWeightAndRegion(ctx, consignRequest.Weight, consignRequest.IsWithin)
		if err != nil {
			return ConsignRecord{}, err
		}
		originalRecord.Price = price
	}

	originalRecord.Consignee = consignRequest.Consignee
	originalRecord.Phone = consignRequest.Phone
	originalRecord.Weight = consignRequest.Weight

	err = collection.InsertOne(ctx, originalRecord)
	if err != nil {
		return ConsignRecord{}, err
	}

	return originalRecord, nil
}

func (c *ConsignServiceImpl) FindByOrderId(ctx context.Context, orderID string) (ConsignRecord, error) {
	collection, err := c.consignDB.GetCollection(ctx, "consign_db", "consign_record")
	if err != nil {
		return ConsignRecord{}, err
	}

	filter := bson.D{{Key: "OrderID", Value: orderID}}
	cursor, err := collection.FindOne(ctx, filter)
	if err != nil {
		return ConsignRecord{}, err
	}

	var consignRecord ConsignRecord
	ok, err := cursor.One(ctx, &consignRecord)
	if err != nil {
		return ConsignRecord{}, err
	}
	if !ok {
		return ConsignRecord{}, fmt.Errorf("consign not found for orderID (%s)", orderID)
	}
	return consignRecord, nil
}

func (c *ConsignServiceImpl) FindByAccountId(ctx context.Context, accountID string) (ConsignRecord, error) {
	collection, err := c.consignDB.GetCollection(ctx, "consign_db", "consign_record")
	if err != nil {
		return ConsignRecord{}, err
	}

	filter := bson.D{{Key: "AccountID", Value: accountID}}
	cursor, err := collection.FindOne(ctx, filter)
	if err != nil {
		return ConsignRecord{}, err
	}

	var consignRecord ConsignRecord
	ok, err := cursor.One(ctx, &consignRecord)
	if err != nil {
		return ConsignRecord{}, err
	}
	if !ok {
		return ConsignRecord{}, fmt.Errorf("consign not found for accountID (%s)", accountID)
	}
	return consignRecord, nil
}

func (c *ConsignServiceImpl) FindByConsignee(ctx context.Context, consignee string) (ConsignRecord, error) {
	collection, err := c.consignDB.GetCollection(ctx, "consign_db", "consign_record")
	if err != nil {
		return ConsignRecord{}, err
	}

	filter := bson.D{{Key: "Consignee", Value: consignee}}
	cursor, err := collection.FindOne(ctx, filter)
	if err != nil {
		return ConsignRecord{}, err
	}

	var consignRecord ConsignRecord
	ok, err := cursor.One(ctx, &consignRecord)
	if err != nil {
		return ConsignRecord{}, err
	}
	if !ok {
		return ConsignRecord{}, fmt.Errorf("consign not found for consignee (%s)", consignee)
	}
	return consignRecord, nil
}
