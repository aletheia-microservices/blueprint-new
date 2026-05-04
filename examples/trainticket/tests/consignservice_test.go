package tests

import (
	"context"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/trainticket/workflow/trainticket"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/assert"
)

var consignServiceRegistry = registry.NewServiceRegistry[trainticket.ConsignService]("consign_service")

func init() {
	consignServiceRegistry.Register("local", func(ctx context.Context) (trainticket.ConsignService, error) {
		priceDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		consignPriceService, err := trainticket.NewConsignPriceServiceImpl(ctx, priceDB)
		if err != nil {
			return nil, err
		}
		// seed the price config so GetPriceByWeightAndRegion works
		priceConfig := trainticket.ConsignPrice{
			ID:            "cp001",
			Index:         0,
			InitialWeight: 2.0,
			InitialPrice:  5.0,
			WithinPrice:   1.0,
			BeyondPrice:   1.5,
		}
		_, err = consignPriceService.CreateAndModifyPriceConfig(ctx, priceConfig)
		if err != nil {
			return nil, err
		}

		consignDB, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		return trainticket.NewConsignServiceImpl(ctx, consignPriceService, consignDB)
	})
}

func TestConsignServiceInsert(t *testing.T) {
	ctx := context.Background()
	service, err := consignServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	req := trainticket.ConsignRequest{
		OrderID:   "consign_ord001",
		AccountID: "consign_acc001",
		From:      "shanghai",
		To:        "beijing",
		Consignee: "Alice",
		Phone:     "13800000001",
		Weight:    3.0,
		IsWithin:  true,
	}
	record, err := service.InsertConsign(ctx, req)
	assert.NoError(t, err)
	assert.NotEmpty(t, record.ID)
	assert.Equal(t, "consign_ord001", record.OrderID)
	assert.Equal(t, "Alice", record.Consignee)
	assert.Greater(t, record.Price, 0.0)
}

func TestConsignServiceFindByOrderId(t *testing.T) {
	ctx := context.Background()
	service, err := consignServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	req := trainticket.ConsignRequest{
		OrderID:   "consign_ord002",
		AccountID: "consign_acc002",
		Consignee: "Bob",
		Weight:    1.0,
		IsWithin:  true,
	}
	_, err = service.InsertConsign(ctx, req)
	assert.NoError(t, err)

	found, err := service.FindByOrderId(ctx, "consign_ord002")
	assert.NoError(t, err)
	assert.Equal(t, "consign_ord002", found.OrderID)
}

func TestConsignServiceFindByAccountId(t *testing.T) {
	ctx := context.Background()
	service, err := consignServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	req := trainticket.ConsignRequest{
		OrderID:   "consign_ord003",
		AccountID: "consign_acc003",
		Consignee: "Carol",
		Weight:    2.5,
		IsWithin:  false,
	}
	_, err = service.InsertConsign(ctx, req)
	assert.NoError(t, err)

	found, err := service.FindByAccountId(ctx, "consign_acc003")
	assert.NoError(t, err)
	assert.Equal(t, "consign_acc003", found.AccountID)
}

func TestConsignServiceFindByConsignee(t *testing.T) {
	ctx := context.Background()
	service, err := consignServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	req := trainticket.ConsignRequest{
		OrderID:   "consign_ord004",
		AccountID: "consign_acc004",
		Consignee: "David",
		Weight:    1.5,
		IsWithin:  true,
	}
	_, err = service.InsertConsign(ctx, req)
	assert.NoError(t, err)

	found, err := service.FindByConsignee(ctx, "David")
	assert.NoError(t, err)
	assert.Equal(t, "David", found.Consignee)
}
