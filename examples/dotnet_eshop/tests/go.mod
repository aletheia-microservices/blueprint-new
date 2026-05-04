module github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/tests

go 1.22.4

require github.com/blueprint-uservices/blueprint/runtime v0.0.0

replace github.com/blueprint-uservices/blueprint/runtime => ../../../runtime

require (
	github.com/blueprint-uservices/blueprint/examples/dotnet_eshop/workflow v0.0.0
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.mongodb.org/mongo-driver v1.17.4 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
