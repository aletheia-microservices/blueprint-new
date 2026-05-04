package order

type CreateOrderCommand struct {
	OrderDto OrderDto
}

type CancelOrderRequest struct {
	OrderId string
}

type CreateOrderResult struct {
	Id string
}

type CancelOrderResponse struct {
	IsSuccess bool
}

type GetOrdersByUserRequest struct {
	CustomerId string
}

type GetOrdersByUserResponse struct {
	Orders []OrderDto
}

type GetOrderRequest struct {
	OrderId string
}

type GetOrderResponse struct {
	Order OrderDto
}

type ShipOrderRequest struct {
	OrderId string
}

type ShipOrderResponse struct {
	IsSuccess bool
}

type DraftOrderItem struct {
	ProductId   int
	ProductName string
	UnitPrice   float64
	Discount    float64
	PictureUrl  string
	Units       int
}

type CreateOrderDraftCommand struct {
	BuyerId string
	Items   []DraftOrderItem
}

type CreateOrderDraftResponse struct {
	OrderItems []OrderItemDto
	Total      float64
}
