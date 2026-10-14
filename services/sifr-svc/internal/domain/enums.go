package domain

type CrossingDirection string

const (
	ENTRY CrossingDirection = "ENTRY"
	EXIT  CrossingDirection = "EXIT"
)

type DocType string

const (
	PASSPORT          DocType = "PASSPORT"
	NATIONAL_ID       DocType = "NATIONAL_ID"
	LAISSEZ_PASSER    DocType = "LAISSEZ_PASSER"
	BIRTH_CERTIFICATE DocType = "BIRTH_CERTIFICATE"
	TRAVEL_DOCUMENT   DocType = "TRAVEL_DOCUMENT"
	NONE              DocType = "NONE"
)

type AlertType string

const (
	WANTED_PERSON   AlertType = "WANTED_PERSON"
	STOLEN_DOCUMENT AlertType = "STOLEN_DOCUMENT"
	BLACKLIST       AlertType = "BLACKLIST"
	ACTIVE_WARRANT  AlertType = "ACTIVE_WARRANT"
	SANCTIONS       AlertType = "SANCTIONS"
	CUSTOMS_ALERT   AlertType = "CUSTOMS_ALERT"
)

const (
	ClearanceDenied  = false
	ClearanceGranted = true
)
