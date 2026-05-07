# eShopMicroservices

This is a Blueprint re-implementation of the [eShopMicroservices application](https://github.com/mehmetozkaya/EShopMicroservices).

## Getting started

Prerequisites for this tutorial:
* docker is installed

## Running tests

```zsh
cd tests
go test
```

## Compiling the application

To compile the application, we execute `wiring/main.go` and specify which wiring spec to compile. To view options and list wiring specs, run:

```zsh
go run wiring/main.go -h
```

If you encounter errors because of missing modules that are supposed to be replaced by local ones, do:

```zsh
cd wiring
go clean -cache -modcache
export GOFLAGS=-mod=mod
export GOWORK=off
go mod tidy
cd ..
```

The following will compile the `docker` wiring spec to the directory `build`.

```zsh
rm -rf build
go run wiring/main.go -w docker -o build
```

## Running the application

To run the application, navigate to `build/docker` and run `docker compose up`. Use flag `--build` to build images if code is changed.

```zsh
docker-compose --env-file build/.env -f build/docker/docker-compose.yml up --build
```

If you see Docker complain about missing environment variables, edit the `.env` file in `build/docker` and remove `0.0.0.0:` in all addresses. For example, `BASKET_DB_BIND_ADDR=0.0.0.0:12345` becomes `BASKET_DB_BIND_ADDR=12345`.

## Sending HTTP requests (examples)

### Catalog Service (port 12348)

```zsh
# Create a product
curl -G 'http://localhost:12348/CreateProduct' \
--data-urlencode 'command={"Name":"Widget","Category":["electronics"],"Description":"A great widget","ImageFile":"widget.jpg","Price":19.99}'

# Get a product by ID
curl -G 'http://localhost:12348/GetProductById' \
--data-urlencode 'query={"Id":"f38000b7-4be2-417c-ba4e-66bdb29379fa"}'

# Get products by category
curl -G 'http://localhost:12348/GetProductByCategory' \
--data-urlencode 'query={"Category":"electronics"}'

# Get all products
curl 'http://localhost:12348/GetProducts'

# Delete a product
curl -G 'http://localhost:12348/DeleteProduct' \
--data-urlencode 'command={"Id":"f38000b7-4be2-417c-ba4e-66bdb29379fa"}'
```

### Discount Service (port 12350)

```zsh
# Create a discount coupon
curl -G 'http://localhost:12350/CreateDiscount' \
--data-urlencode 'request={"Coupon":{"Id":1,"ProductName":"Widget","Description":"10% off","Amount":5.0}}'

# Get a discount by product name
curl -G 'http://localhost:12350/GetDiscount' \
--data-urlencode 'request={"ProductName":"Widget"}'

# Update a discount
curl -G 'http://localhost:12350/UpdateDiscount' \
--data-urlencode 'request={"Coupon":{"Id":1,"ProductName":"Widget","Description":"20% off","Amount":10.0}}'

# Delete a discount
curl -G 'http://localhost:12350/DeleteDiscount' \
--data-urlencode 'request={"ProductName":"Widget"}'
```

### Basket Service (port 12346)

```zsh
# Store a basket
curl -G 'http://localhost:12346/StoreBasket' \
--data-urlencode 'request={"Cart":{"UserName":"swn","Items":[{"Quantity":1,"Color":"red","Price":19.99,"ProductName":"Widget"}],"TotalPrice":19.99}}'

# Get a basket
curl -G 'http://localhost:12346/GetBasket' \
--data-urlencode 'query={"UserName":"swn"}'

# Checkout a basket
curl -G 'http://localhost:12346/CheckoutBasket' \
--data-urlencode 'request={"BasketCheckoutDto":{"UserName":"swn","FirstName":"John","LastName":"Doe","EmailAddress":"john@example.com","AddressLine":"123 Main St","Country":"US","State":"NY","ZipCode":"10001","CardName":"John Doe","CardNumber":"4242424242424242","Expiration":"12/25","CVV":"123","PaymentMethod":1}}'

# Delete a basket
curl -G 'http://localhost:12346/DeleteBasket' \
--data-urlencode 'query={"UserName":"swn"}'
```

### Order Service (port 12353)

```zsh
# Create a new order
curl -G 'http://localhost:12353/CreateNewOrder' \
--data-urlencode 'command={"OrderDto":{"CustomerId":"swn", "OrderName":"john@example.com","Status":2,"ShippingAddress":{"FirstName":"John","LastName":"Doe","AddressLine":"123 Main St","Country":"US","State":"NY","ZipCode":"10001"},"Payment":{"CardName":"John Doe","CardNumber":"4242424242424242","Expiration":"12/25","CCV":"123","PaymentMethod":1}}}'

# Get orders by customer ID
curl -G 'http://localhost:12353/GetOrdersByCustomer' \
--data-urlencode 'query={"CustomerId":"swn"}'

# Update an order
curl -G 'http://localhost:12353/UpdateOrder' \
--data-urlencode 'command={"OrderDto":{"Id":"<order-id>","Status":3}}'

# Delete an order
curl -G 'http://localhost:12353/DeleteOrder' \
--data-urlencode 'command={"Id":"<order-id>"}'
```

### Web App (port 12354)

```zsh
# Get all products (optionally filtered by category)
curl 'http://localhost:12354/OnGetProductsAsync?categoryName=electronics'

# Add a product to cart
curl -X POST 'http://localhost:12354/OnPostAddToCartAsync?productId=f38000b7-4be2-417c-ba4e-66bdb29379fa'

# Remove a product from cart
curl -X POST 'http://localhost:12354/OnPostRemoveToCartAsync?productId=f38000b7-4be2-417c-ba4e-66bdb29379fa'

# Checkout
curl -X POST 'http://localhost:12354/OnPostCheckoutAsync'

# Get orders
curl 'http://localhost:12354/OnGetOrdersAsync'
```
