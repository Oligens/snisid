package domain

type VesselType string

const (
	VesselTypeCARGO   VesselType = "CARGO_SHIP"
	VesselTypeTANKER  VesselType = "TANKER"
	VesselTypeFISHING VesselType = "FISHING_BOAT"
	VesselTypeGOFAST  VesselType = "GO_FAST"
	VesselTypeSAIL    VesselType = "SAILBOAT"
	VesselTypeYACHT   VesselType = "YACHT"
	VesselTypeFERRY   VesselType = "FERRY"
	VesselTypePATROL  VesselType = "PATROL_BOAT"
	VesselTypeWOODEN  VesselType = "WOODEN_BOAT"
	VesselTypeCANOE   VesselType = "CANOE"
	VesselTypeUNKNOWN VesselType = "UNKNOWN"
)

type VesselStatus string

const (
	StatusREGISTERED VesselStatus = "REGISTERED"
	StatusSTOLEN     VesselStatus = "STOLEN"
	StatusSUSPECTED  VesselStatus = "SUSPECTED"
	StatusDETAINED   VesselStatus = "DETAINED"
	StatusSUNK       VesselStatus = "SUNK"
	StatusDESTROYED  VesselStatus = "DESTROYED"
	StatusMISSING    VesselStatus = "MISSING"
	StatusINTERPOL   VesselStatus = "INTERPOL_ALERT"
)

type IncidentType string

const (
	IncidentDRUG     IncidentType = "DRUG_SEIZURE"
	IncidentARMS     IncidentType = "ARMS_SEIZURE"
	IncidentMIGRANT  IncidentType = "MIGRANT_INTERDICTION"
	IncidentSMUGGLE  IncidentType = "SMUGGLING"
	IncidentSUSPECT  IncidentType = "SUSPICIOUS_ACTIVITY"
	IncidentDISTRESS IncidentType = "DISTRESS"
	IncidentPIRACY   IncidentType = "PIRACY"
	IncidentILLEGAL  IncidentType = "ILLEGAL_FISHING"
	IncidentTRAFFICK IncidentType = "HUMAN_TRAFFICKING"
)
