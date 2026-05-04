package order

// OrderStockItem carries per-item stock validation data,
// matching the original OrderStockItem record in Catalog.API
type OrderStockItem struct {
	ProductId int
	Units     int
}

// OrderStatusChangedToAwaitingValidationEvent is published by order service on CreateOrder
// Catalog.API handles it to validate stock availability per item
type OrderStatusChangedToAwaitingValidationEvent struct {
	OrderId         string
	OrderStockItems []OrderStockItem
}

// OrderStockValidationResultEvent is published by catalog service after stock check
// Replaces OrderStockConfirmedIntegrationEvent / OrderStockRejectedIntegrationEvent
type OrderStockValidationResultEvent struct {
	OrderId   string
	Confirmed bool
}

// OrderStatusChangedToStockConfirmedEvent is published by order service
// after stock is confirmed, triggering payment processing
type OrderStatusChangedToStockConfirmedEvent struct {
	OrderId string
}

type OrderPaymentResultEvent struct {
	OrderId   string
	Succeeded bool
}

type OrderPaidEvent struct {
	OrderId    string
	OrderItems []OrderItemDto
}
