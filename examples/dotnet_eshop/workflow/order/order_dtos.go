package order

type OrderStatus int

const (
	Draft              OrderStatus = iota + 1
	Pending
	Completed
	Cancelled
	Submitted
	AwaitingValidation
	StockConfirmed
	Paid
	Shipped
)

type OrderDto struct {
	Id              string         `bson:"Id"`
	CustomerId      string         `bson:"CustomerId"`
	OrderName       string         `bson:"OrderName"`
	OrderDate       string         `bson:"OrderDate"`
	Description     string         `bson:"Description"`
	ShippingAddress AddressDto     `bson:"ShippingAddress"`
	BillingAddress  AddressDto     `bson:"BillingAddress"`
	Payment         PaymentDto     `bson:"Payment"`
	Status          OrderStatus    `bson:"Status"`
	OrderItems      []OrderItemDto `bson:"OrderItems"`
}

type OrderItemDto struct {
	OrderId   string  `bson:"OrderId"`
	ProductID int     `bson:"ProductID"`
	Quantity  int     `bson:"Quantity"`
	Price     float64 `bson:"Price"`
}

type AddressDto struct {
	FirstName    string `bson:"FirstName"`
	LastName     string `bson:"LastName"`
	EmailAddress string `bson:"EmailAddress"`
	AddressLine  string `bson:"AddressLine"`
	Country      string `bson:"Country"`
	State        string `bson:"State"`
	ZipCode      string `bson:"ZipCode"`
}

type PaymentDto struct {
	CardName      string `bson:"CardName"`
	CardNumber    string `bson:"CardNumber"`
	Expiration    string `bson:"Expiration"`
	CCV           string `bson:"CCV"`
	PaymentMethod int    `bson:"PaymentMethod"`
}
