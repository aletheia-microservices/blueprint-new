package specs

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/workflow/basket"
	"github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/workflow/catalog"
	"github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/workflow/order"
	"github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/workflow/payment"
	"github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/workflow/web"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/gotests"
	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/plugins/rabbitmq"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

var Docker = cmdbuilder.SpecOption{
	Name:  "docker",
	Build: makeDockerSpec,
}

func makeDockerSpec(spec wiring.WiringSpec) ([]string, error) {
	var containers []string
	var allServices []string

	catalog_db := mongodb.Container(spec, "catalog_db")
	allServices = append(allServices, catalog_db)
	basket_db := mongodb.Container(spec, "basket_db")
	allServices = append(allServices, basket_db)
	order_db := mongodb.Container(spec, "order_db")
	allServices = append(allServices, order_db)

	// order -> catalog: stock validation requests (OrderStatusChangedToAwaitingValidationIntegrationEvent)
	awaiting_validation_queue := rabbitmq.Container(spec, "awaiting_validation_queue", "awaiting_validation_queue")
	allServices = append(allServices, awaiting_validation_queue)
	// catalog -> order: stock validation results (OrderStockConfirmed / OrderStockRejected)
	stock_validation_result_queue := rabbitmq.Container(spec, "stock_validation_result_queue", "stock_validation_result_queue")
	allServices = append(allServices, stock_validation_result_queue)
	// order -> payment: stock confirmed events (triggers payment processing)
	stock_confirmed_queue := rabbitmq.Container(spec, "stock_confirmed_queue", "stock_confirmed_queue")
	allServices = append(allServices, stock_confirmed_queue)
	// catalog -> (subscribers): product price changed events
	catalog_price_queue := rabbitmq.Container(spec, "catalog_price_queue", "catalog_price_queue")
	allServices = append(allServices, catalog_price_queue)
	// payment -> order: payment succeeded/failed events
	payment_result_queue := rabbitmq.Container(spec, "payment_result_queue", "payment_result_queue")
	allServices = append(allServices, payment_result_queue)
	// order -> catalog: order paid events (triggers stock deduction)
	order_paid_queue := rabbitmq.Container(spec, "order_paid_queue", "order_paid_queue")
	allServices = append(allServices, order_paid_queue)

	catalog_service := workflow.Service[catalog.CatalogService](spec, "catalog_service", catalog_db, catalog_price_queue, awaiting_validation_queue, stock_validation_result_queue, order_paid_queue)
	catalog_service_ctr := applyHTTPDefaults(spec, catalog_service, "catalog_service_proc", "catalog_service_container")
	containers = append(containers, catalog_service_ctr)
	allServices = append(allServices, "catalog_service")

	basket_service := workflow.Service[basket.BasketService](spec, "basket_service", basket_db)
	basket_service_ctr := applyHTTPDefaults(spec, basket_service, "basket_service_proc", "basket_service_container")
	containers = append(containers, basket_service_ctr)
	allServices = append(allServices, "basket_service")

	order_service := workflow.Service[order.OrderService](spec, "order_service", order_db, awaiting_validation_queue, stock_validation_result_queue, stock_confirmed_queue, payment_result_queue, order_paid_queue)
	order_service_ctr := applyHTTPDefaults(spec, order_service, "order_service_proc", "order_service_container")
	containers = append(containers, order_service_ctr)
	allServices = append(allServices, "order_service")

	payment_service := workflow.Service[payment.PaymentService](spec, "payment_service", stock_confirmed_queue, payment_result_queue)
	payment_service_ctr := applyDockerDefaults(spec, payment_service, "payment_service_proc", "payment_service_container")
	containers = append(containers, payment_service_ctr)
	allServices = append(allServices, "payment_service")

	web_app := workflow.Service[web.WebApp](spec, "web_app", basket_service, catalog_service, order_service)
	web_app_ctr := applyHTTPDefaults(spec, web_app, "web_app_proc", "web_app_container")
	containers = append(containers, web_app_ctr)
	allServices = append(allServices, "web_app")

	tests := gotests.Test(spec, allServices...)
	containers = append(containers, tests)

	return containers, nil
}
