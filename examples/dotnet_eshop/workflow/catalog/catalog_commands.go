package catalog

type CreateItemRequest struct {
	ID                int
	Name              string
	Description       string
	Price             float64
	PriceFileName     string
	CatalogTypeID     int
	CatalogType       CatalogType
	CatalogBrandID    int
	CatalogBrand      CatalogBrand
	AvailableStock    int
	RestockThreshold  int
	MaxStockThreshold int
}

type CreateItemResponse struct {
	Item CatalogItem
}

type UpdateItemResponse struct {
	ID int
}

type DeleteItemRequest struct {
	ID int
}

type GetItemByIDRequest struct {
	ID int
}

type GetItemsByIDsRequest struct {
	IDs []int
}

type GetItemByIDResponse struct {
	Item CatalogItem
}

type GetItemsByIDsResponse struct {
	Item []CatalogItem
}


type GetItemsByNameRequest struct {
	Name string
}

type GetItemsByNameResponse struct {
	Items []CatalogItem
}

type GetItemPictureByIdRequest struct {
	ID int
}

type GetItemPictureByIdResponse struct {
	PictureFileName string
}

type GetAllItemsResponse struct {
	Items []CatalogItem
}

type GetItemsByBrandAndTypeRequest struct {
	TypeId  int
	BrandId int
}

type GetItemsByBrandAndTypeResponse struct {
	Items []CatalogItem
}

type GetItemsByBrandRequest struct {
	BrandId int
}

type GetItemsByBrandResponse struct {
	Items []CatalogItem
}

type GetCatalogTypesResponse struct {
	Types []CatalogType
}

type GetCatalogBrandsResponse struct {
	Brands []CatalogBrand
}

type RemoveStockRequest struct {
	ProductID int
	Quantity  int
}

type RemoveStockResponse struct {
	Depleted bool
}

type AddStockRequest struct {
	ProductID int
	Quantity  int
}

type AddStockResponse struct {
	AvailableStock int
}
