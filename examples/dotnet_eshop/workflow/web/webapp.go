package web

import (
	"context"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/workflow/basket"
	"github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/workflow/catalog"
	"github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/workflow/order"
)

// WebApp bundles the Razor page behaviors of the original WebApp into a single testable service.
// Each method corresponds to a page action in the original:
//   - GetCatalogItems  -> Catalog.razor OnInitializedAsync
//   - GetCatalogItem   -> ItemPage.razor OnInitializedAsync
//   - AddToCartAsync   -> ItemPage.razor AddToCartAsync
//   - GetBasketItems   -> CartPage.razor OnInitializedAsync
//   - SetQuantityAsync -> CartPage.razor UpdateQuantityAsync (qty=0 removes)
//   - CheckoutAsync    -> Checkout.razor HandleValidSubmitAsync
//   - GetOrders        -> Orders.razor OnInitializedAsync
type WebApp interface {
	GetCatalogItems(ctx context.Context, typeId int, brandId int) ([]catalog.CatalogItem, error)
	GetCatalogItem(ctx context.Context, itemId int) (catalog.CatalogItem, error)
	AddToCartAsync(ctx context.Context, productId int) error
	GetBasketItems(ctx context.Context) ([]basket.BasketItem, error)
	SetQuantityAsync(ctx context.Context, productId int, quantity int) error
	CheckoutAsync(ctx context.Context) error
	GetOrders(ctx context.Context) ([]order.OrderDto, error)
}

type WebAppImpl struct {
	basketService  basket.BasketService
	catalogService catalog.CatalogService
	orderService   order.OrderService
	customerId     string
	userName       string
}

func NewWebAppImpl(ctx context.Context, basketService basket.BasketService, catalogService catalog.CatalogService, orderService order.OrderService) (WebApp, error) {
	s := &WebAppImpl{
		basketService:  basketService,
		catalogService: catalogService,
		orderService:   orderService,
		customerId:     "5334c996-8457-4cf0-815c-ed2b77c4ff61",
		userName:       "swn",
	}
	return s, nil
}

func (webapp *WebAppImpl) GetCatalogItem(ctx context.Context, itemId int) (catalog.CatalogItem, error) {
	resp, err := webapp.catalogService.GetItemById(ctx, catalog.GetItemByIDRequest{ID: itemId})
	if err != nil {
		return catalog.CatalogItem{}, err
	}
	return resp.Item, nil
}

func (webapp *WebAppImpl) GetBasketItems(ctx context.Context) ([]basket.BasketItem, error) {
	resp, err := webapp.basketService.GetBasket(ctx, basket.GetBasketRequest{UserName: webapp.userName})
	if err != nil {
		return nil, err
	}
	return resp.Cart.Items, nil
}

func (webapp *WebAppImpl) GetCatalogItems(ctx context.Context, typeId int, brandId int) ([]catalog.CatalogItem, error) {
	if typeId != 0 && brandId != 0 {
		resp, err := webapp.catalogService.GetItemsByBrandAndTypeId(ctx, catalog.GetItemsByBrandAndTypeRequest{TypeId: typeId, BrandId: brandId})
		if err != nil {
			return nil, err
		}
		return resp.Items, nil
	} else if brandId != 0 {
		resp, err := webapp.catalogService.GetItemsByBrandId(ctx, catalog.GetItemsByBrandRequest{BrandId: brandId})
		if err != nil {
			return nil, err
		}
		return resp.Items, nil
	}
	resp, err := webapp.catalogService.GetAllItems(ctx)
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (webapp *WebAppImpl) AddToCartAsync(ctx context.Context, productId int) error {
	productResponse, err := webapp.catalogService.GetItemById(ctx, catalog.GetItemByIDRequest{ID: productId})
	if err != nil {
		return err
	}

	basketResponse, err := webapp.basketService.GetBasket(ctx, basket.GetBasketRequest{UserName: webapp.userName})
	if err != nil {
		return err
	}
	cart := basketResponse.Cart

	found := false
	for i, item := range cart.Items {
		if item.ProductID == productId {
			cart.Items[i].Quantity++
			found = true
			break
		}
	}
	if !found {
		cart.Items = append(cart.Items, basket.BasketItem{
			ProductID:   productId,
			ProductName: productResponse.Item.Name,
			UnitPrice:   productResponse.Item.Price,
			Quantity:    1,
		})
	}

	_, err = webapp.basketService.UpdateBasket(ctx, basket.UpdateBasketRequest{Cart: cart})
	return err
}

func (webapp *WebAppImpl) SetQuantityAsync(ctx context.Context, productId int, quantity int) error {
	basketResponse, err := webapp.basketService.GetBasket(ctx, basket.GetBasketRequest{UserName: webapp.userName})
	if err != nil {
		return err
	}
	cart := basketResponse.Cart

	var updated []basket.BasketItem
	for _, item := range cart.Items {
		if item.ProductID == productId {
			if quantity > 0 {
				item.Quantity = quantity
				updated = append(updated, item)
			}
		} else {
			updated = append(updated, item)
		}
	}
	cart.Items = updated

	_, err = webapp.basketService.UpdateBasket(ctx, basket.UpdateBasketRequest{Cart: cart})
	return err
}

func (webapp *WebAppImpl) CheckoutAsync(ctx context.Context) error {
	basketResponse, err := webapp.basketService.GetBasket(ctx, basket.GetBasketRequest{UserName: webapp.userName})
	if err != nil {
		return err
	}
	cart := basketResponse.Cart

	var orderItems []order.OrderItemDto
	for _, item := range cart.Items {
		orderItems = append(orderItems, order.OrderItemDto{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.UnitPrice,
		})
	}

	orderDto := order.OrderDto{
		CustomerId: webapp.customerId,
		OrderName:  webapp.userName,
		OrderDate:  time.Now().UTC().Format(time.RFC3339),
		Status:     order.Submitted,
		OrderItems: orderItems,
	}
	_, err = webapp.orderService.CreateOrder(ctx, order.CreateOrderCommand{OrderDto: orderDto})
	if err != nil {
		return err
	}

	_, err = webapp.basketService.DeleteBasket(ctx, basket.DeleteBasketRequest{UserName: webapp.userName})
	return err
}

func (webapp *WebAppImpl) GetOrders(ctx context.Context) ([]order.OrderDto, error) {
	response, err := webapp.orderService.GetOrdersByUser(ctx, order.GetOrdersByUserRequest{CustomerId: webapp.customerId})
	if err != nil {
		return nil, err
	}
	return response.Orders, nil
}
