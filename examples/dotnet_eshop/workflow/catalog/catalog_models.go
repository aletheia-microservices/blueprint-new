package catalog

type CatalogBrand struct {
	ID    int    `bson:"ID"`
	Brand string `bson:"Brand"`
}

type CatalogItem struct {
	ID                int          `bson:"ID"`
	Name              string       `bson:"Name"`
	Description       string       `bson:"Description"`
	Price             float64      `bson:"Price"`
	PictureFileName   string       `bson:"PictureFileName"`
	CatalogTypeID     int          `bson:"CatalogTypeID"`
	CatalogType       CatalogType  `bson:"CatalogType"`
	CatalogBrandID    int          `bson:"CatalogBrandID"`
	CatalogBrand      CatalogBrand `bson:"CatalogBrand"`
	AvailableStock    int          `bson:"AvailableStock"`
	RestockThreshold  int          `bson:"RestockThreshold"`
	MaxStockThreshold int          `bson:"MaxStockThreshold"`
}

type CatalogType struct {
	ID   int    `bson:"ID"`
	Type string `bson:"Type"`
}

type PaginatedItems struct {
	PageIndex int
	PageSize  int
	Count     int64
}
