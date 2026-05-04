package basket

type CustomerBasket struct {
	UserName string       `bson:"UserName"`
	BuyerID  string       `bson:"BuyerID"`
	Items    []BasketItem `bson:"Items"`
}

type BasketItem struct {
	ID           string  `bson:"ID"`
	ProductID    int     `bson:"ProductID"`
	ProductName  string  `bson:"ProductName"`
	UnitPrice    float64 `bson:"UnitPrice"`
	OldUnitPrice float64 `bson:"OldUnitPrice"`
	Quantity     int     `bson:"Quantity"`
	PictureUrl   string  `bson:"PictureUrl"`
}
