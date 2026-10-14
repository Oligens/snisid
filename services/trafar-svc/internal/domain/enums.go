package domain

type RouteType string

const (
	RouteTypeMaritimeDirect     RouteType = "MARITIME_DIRECT"
	RouteTypeMaritimeViaBahamas RouteType = "MARITIME_VIA_BAHAMAS"
	RouteTypeAirCargo           RouteType = "AIR_CARGO"
	RouteTypeAirPassenger       RouteType = "AIR_PASSENGER"
	RouteTypeLandBorderDom      RouteType = "LAND_BORDER_DOM"
	RouteTypeLandBorderOther    RouteType = "LAND_BORDER_OTHER"
	RouteTypePostal             RouteType = "POSTAL"
	RouteTypeMixed              RouteType = "MIXED"
)

type TraffickingMethod string

const (
	TraffickingMethodStrawPurchase    TraffickingMethod = "STRAW_PURCHASE"
	TraffickingMethodStolenDiverted   TraffickingMethod = "STOLEN_DIVERTED"
	TraffickingMethodCorruptOfficial  TraffickingMethod = "CORRUPT_OFFICIAL"
	TraffickingMethodFalseEndUser     TraffickingMethod = "FALSE_END_USER"
	TraffickingMethodDarkWeb          TraffickingMethod = "DARK_WEB"
	TraffickingMethodDiplomaticPouch  TraffickingMethod = "DIPLOMATIC_POUCH"
	TraffickingMethodConcealedCargo   TraffickingMethod = "CONCEALED_CARGO"
	TraffickingMethodDrugsForGunsSwap TraffickingMethod = "DRUGS_FOR_GUNS_SWAP"
	TraffickingMethodUnknown          TraffickingMethod = "UNKNOWN"
)
