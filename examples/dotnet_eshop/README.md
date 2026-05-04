# .NET eShop

This is a Blueprint re-implementation of the [.NET eShop application](https://github.com/dotnet/eShop).

* [workflow](workflow) contains service implementations
* [tests](tests) has tests of the workflow
* [wiring](wiring) configures the application's topology and deployment

## Architecture

The application consists of five services communicating over HTTP, backed by MongoDB databases and RabbitMQ queues:

| Service | Description |
|---------|-------------|
| `web_app` | Bundles the Razor page behaviors (catalog browsing, cart, checkout, orders) |
| `catalog_service` | Manages catalog items, types, brands, and stock |
| `basket_service` | Manages shopping cart state per user |
| `order_service` | Manages order lifecycle |
| `payment_service` | Processes payments (always succeeds in this implementation) |

Event flow:
```
checkout → order_service --[stock_confirmed_queue]--> payment_service
                                                           |
                         <--[payment_result_queue]---------+
                         --[order_paid_queue]--> catalog_service (stock deduction)
catalog_service --[catalog_price_queue]--> (price change subscribers)
```

## Prerequisites

* Docker is installed

## Running tests

```zsh
cd tests
go test
```

## Compiling the application

To view available wiring specs:

```
go run wiring/main.go -h
```

To compile the `docker` wiring spec to the `build` directory:

```
rm -rf build
go run wiring/main.go -w docker -o build
```

## Running the application

```zsh
docker-compose --env-file build/.env -f build/docker/docker-compose.yml up --build
```

If Docker complains about missing environment variables, edit `build/.env` and remove the `0.0.0.0:` prefix from all addresses. For example, `WEB_APP_HTTP_BIND_ADDR=0.0.0.0:12356` becomes `WEB_APP_HTTP_BIND_ADDR=12356`.

Check `build/.env` for the actual port assigned to each service after compilation.

## Sending HTTP requests (examples)

### Web App

The web app is the primary entry point, mirroring the Razor page behaviors of the original .NET eShop WebApp.

```zsh
# Browse all catalog items (Catalog.razor)
curl "http://localhost:12356/GetCatalogItems?typeId=0&brandId=0"

# Browse catalog items filtered by type
curl "http://localhost:12356/GetCatalogItems?typeId=1&brandId=0"

# Browse catalog items filtered by type and brand
curl "http://localhost:12356/GetCatalogItems?typeId=1&brandId=5"

# View a single item (ItemPage.razor)
curl "http://localhost:12356/GetCatalogItem?itemId=1001"

# Add item to cart (ItemPage.razor AddToCartAsync)
curl "http://localhost:12356/AddToCartAsync?productId=1001"

# View cart contents (CartPage.razor)
curl "http://localhost:12356/GetBasketItems"

# Update item quantity in cart (CartPage.razor UpdateQuantityAsync; quantity=0 removes)
curl "http://localhost:12356/SetQuantityAsync?productId=1001&quantity=3"

# Checkout (Checkout.razor HandleValidSubmitAsync)
curl "http://localhost:12356/CheckoutAsync"

# View orders (Orders.razor)
curl "http://localhost:12356/GetOrders"
```

### Catalog Service

```zsh
# Get all catalog items
curl "http://localhost:12349/GetAllItems"

# Get a catalog item by ID
curl "http://localhost:12349/GetItemById?ID=1001"

# Get catalog items by name
curl "http://localhost:12349/GetItemsByName?Name=Laptop"

# Get catalog items by brand and type
curl "http://localhost:12349/GetItemsByBrandAndTypeId?TypeId=1&BrandId=5"

# Get catalog items by brand
curl "http://localhost:12349/GetItemsByBrandId?BrandId=5"

# Get all catalog types
curl "http://localhost:12349/GetCatalogTypes"

# Get all catalog brands
curl "http://localhost:12349/GetCatalogBrands"

# Get item picture filename by ID
curl "http://localhost:12349/GetItemPictureById?ID=1001"

# Create a catalog item
curl -X POST "http://localhost:12349/CreateItem" \
  -H "Content-Type: application/json" \
  -d '{"ID":1001,"Name":"Laptop","Description":"A great laptop","Price":999.99,"CatalogTypeID":1,"CatalogBrandID":5,"AvailableStock":50,"RestockThreshold":10,"MaxStockThreshold":100}'

# Update a catalog item
curl -X POST "http://localhost:12349/UpdateItem" \
  -H "Content-Type: application/json" \
  -d '{"ID":1001,"Name":"Laptop Pro","Price":1199.99,"CatalogTypeID":1,"CatalogBrandID":5,"AvailableStock":45}'

# Delete a catalog item
curl "http://localhost:12349/DeleteItemById?ID=1001"

# Add stock
curl "http://localhost:12349/AddStock?ProductID=1001&Quantity=50"

# Remove stock
curl "http://localhost:12349/RemoveStock?ProductID=1001&Quantity=5"
```

### Basket Service

```zsh
# Get basket for a user
curl "http://localhost:12345/GetBasket?UserName=alice"

# Update basket (replace entire cart)
curl -X POST "http://localhost:12345/UpdateBasket" \
  -H "Content-Type: application/json" \
  -d '{"Cart":{"UserName":"alice","Items":[{"ProductID":1001,"ProductName":"Laptop","UnitPrice":999.99,"Quantity":1}]}}'

# Delete basket
curl "http://localhost:12345/DeleteBasket?UserName=alice"
```

### Order Service

```zsh
# Get an order by ID
curl "http://localhost:12352/GetOrder?OrderId=<order-id>"

# Get all orders for a user
curl "http://localhost:12352/GetOrdersByUser?CustomerId=<customer-id>"

# Cancel an order
curl "http://localhost:12352/CancelOrder?OrderId=<order-id>"

# Ship an order
curl "http://localhost:12352/ShipOrder?OrderId=<order-id>"

# Create an order draft (preview total without persisting)
curl -X POST "http://localhost:12352/CreateOrderDraft" \
  -H "Content-Type: application/json" \
  -d '{"BuyerId":"alice","Items":[{"ProductId":1001,"ProductName":"Laptop","UnitPrice":999.99,"Discount":0,"Units":1}]}'

# Create an order
curl -X POST "http://localhost:12352/CreateOrder" \
  -H "Content-Type: application/json" \
  -d '{"OrderDto":{"CustomerId":"<customer-id>","OrderName":"alice","Status":0,"OrderItems":[{"ProductID":1001,"Quantity":1,"Price":999.99}]}}'
```
