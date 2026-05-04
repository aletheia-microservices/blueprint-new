package basket

type GetBasketRequest struct {
	UserName string
}

type UpdateBasketRequest struct {
	Cart CustomerBasket
}

type DeleteBasketRequest struct {
	UserName string
}

type CustomerBasketResponse struct {
	Cart CustomerBasket
}

type DeleteBasketResponse struct {
	IsSuccess bool
}
