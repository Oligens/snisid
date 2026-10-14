package domain

type ExtorsType string

const (
	KIDNAPPING_RANSOM          ExtorsType = "KIDNAPPING_RANSOM"
	ROAD_TOLL_ILLEGAL          ExtorsType = "ROAD_TOLL_ILLEGAL"
	BUSINESS_PROTECTION_RACKET ExtorsType = "BUSINESS_PROTECTION_RACKET"
	REAL_ESTATE_EXTORTION      ExtorsType = "REAL_ESTATE_EXTORTION"
	PUBLIC_SERVANT_EXTORTION   ExtorsType = "PUBLIC_SERVANT_EXTORTION"
	NGO_EXTORTION              ExtorsType = "NGO_EXTORTION"
	FUEL_TRUCK_HIJACK          ExtorsType = "FUEL_TRUCK_HIJACK"
	OTHER                      ExtorsType = "OTHER"
)

type PaymentChannel string

const (
	MONCASH        PaymentChannel = "MONCASH"
	NATCASH        PaymentChannel = "NATCASH"
	DIGICEL_MONEY  PaymentChannel = "DIGICEL_MONEY"
	WIRE_TRANSFER  PaymentChannel = "WIRE_TRANSFER"
	CASH_DROP      PaymentChannel = "CASH_DROP"
	CRYPTOCURRENCY PaymentChannel = "CRYPTOCURRENCY"
	INTERMEDIARY   PaymentChannel = "INTERMEDIARY"
	UNKNOWN        PaymentChannel = "UNKNOWN"
)

type ExtorsStatus string

const (
	ACTIVE                   ExtorsStatus = "ACTIVE"
	PAID                     ExtorsStatus = "PAID"
	REFUSED                  ExtorsStatus = "REFUSED"
	NEGOTIATING              ExtorsStatus = "NEGOTIATING"
	LAW_ENFORCEMENT_INVOLVED ExtorsStatus = "LAW_ENFORCEMENT_INVOLVED"
	RESOLVED                 ExtorsStatus = "RESOLVED"
	VICTIM_HARMED            ExtorsStatus = "VICTIM_HARMED"
)
