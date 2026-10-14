package domain

type OrderType string

const (
	OrderTypeRestraintingOrder OrderType = "RESTRAINING_ORDER"
	OrderTypeNoContact         OrderType = "NO_CONTACT"
	OrderTypeStayAway          OrderType = "STAY_AWAY"
	OrderTypeProtective        OrderType = "PROTECTIVE"
	OrderTypeWitnessProtection OrderType = "WITNESS_PROTECTION"
	OrderTypeGangExclusionZone OrderType = "GANG_EXCLUSION_ZONE"
	OrderTypeTravelRestriction OrderType = "TRAVEL_RESTRICTION"
)

type OrderStatus string

const (
	StatusActive    OrderStatus = "ACTIVE"
	StatusExpired   OrderStatus = "EXPIRED"
	StatusViolated  OrderStatus = "VIOLATED"
	StatusDismissed OrderStatus = "DISMISSED"
	StatusAppealed  OrderStatus = "APPEALED"
)
