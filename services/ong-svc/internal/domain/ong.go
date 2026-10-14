package domain

import (
	"time"

	"github.com/google/uuid"
)

type RegistrationStatus string

const (
	Registered          RegistrationStatus = "REGISTERED"
	Pending             RegistrationStatus = "PENDING"
	Suspended           RegistrationStatus = "SUSPENDED"
	Revoked             RegistrationStatus = "REVOKED"
	OperatingWithoutReg RegistrationStatus = "OPERATING_WITHOUT_REGISTRATION"
)

type ONGType string

const (
	Humanitarian ONGType = "HUMANITARIAN"
	Development  ONGType = "DEVELOPMENT"
	Advocacy     ONGType = "ADVOCACY"
	FaithBased   ONGType = "FAITH_BASED"
	Diaspora     ONGType = "DIASPORA"
	Research     ONGType = "RESEARCH"
	Mixed        ONGType = "MIXED"
	Unknown      ONGType = "UNKNOWN"
)

type RiskFlag string

const (
	None                       RiskFlag = "NONE"
	FinancialIrregularity      RiskFlag = "FINANCIAL_IRREGULARITY"
	StaffSecurityConcern       RiskFlag = "STAFF_SECURITY_CONCERN"
	OperatingInRestrictedZone  RiskFlag = "OPERATING_IN_RESTRICTED_ZONE"
	SanctionMatch              RiskFlag = "SANCTION_MATCH"
	SuspectedFrontOrganization RiskFlag = "SUSPECTED_FRONT_ORGANIZATION"
	UnregisteredIllegal        RiskFlag = "UNREGISTERED_ILLEGAL"
)

type Organization struct {
	ID                     uuid.UUID          `json:"org_id" db:"org_id"`
	NationalONGID          string             `json:"national_ong_id" db:"national_ong_id"`
	OrgName                string             `json:"org_name" db:"org_name"`
	OrgNameLocal           *string            `json:"org_name_local,omitempty" db:"org_name_local"`
	Acronym                *string            `json:"acronym,omitempty" db:"acronym"`
	OrgType                ONGType            `json:"org_type" db:"org_type"`
	RegistrationStatus     RegistrationStatus `json:"registration_status" db:"registration_status"`
	MJSPRegistrationNumber *string            `json:"mjsp_registration_number,omitempty" db:"mjsp_registration_number"`
	RegistrationDate       *time.Time         `json:"registration_date,omitempty" db:"registration_date"`
	HeadquarterCountry     string             `json:"headquarter_country" db:"headquarter_country"`
	HeadquarterCity        *string            `json:"headquarter_city,omitempty" db:"headquarter_city"`
	HaitiOfficeDept        *string            `json:"haiti_office_dept,omitempty" db:"haiti_office_dept"`
	HaitiOfficeAddress     *string            `json:"haiti_office_address,omitempty" db:"haiti_office_address"`
	HaitiOfficeLat         *float64           `json:"haiti_office_lat,omitempty" db:"haiti_office_lat"`
	HaitiOfficeLng         *float64           `json:"haiti_office_lng,omitempty" db:"haiti_office_lng"`
	OperatingDepts         []string           `json:"operating_depts" db:"operating_depts"`
	Sectors                []string           `json:"sectors" db:"sectors"`
	AnnualBudgetUSD        *float64           `json:"annual_budget_usd,omitempty" db:"annual_budget_usd"`
	HaitiStaffCount        *int               `json:"haiti_staff_count,omitempty" db:"haiti_staff_count"`
	ExpatStaffCount        *int               `json:"expat_staff_count,omitempty" db:"expat_staff_count"`
	DirectorName           *string            `json:"director_name,omitempty" db:"director_name"`
	DirectorNationality    *string            `json:"director_nationality,omitempty" db:"director_nationality"`
	ContactEmail           *string            `json:"contact_email,omitempty" db:"contact_email"`
	ContactPhone           *string            `json:"contact_phone,omitempty" db:"contact_phone"`
	RiskFlag               RiskFlag           `json:"risk_flag" db:"risk_flag"`
	RiskNotes              *string            `json:"risk_notes,omitempty" db:"risk_notes"`
	IsAccessRestricted     *bool              `json:"is_access_restricted,omitempty" db:"is_access_restricted"`
	CreatedBy              uuid.UUID          `json:"created_by" db:"created_by"`
	CreatedAt              time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time          `json:"updated_at" db:"updated_at"`
}

type Staff struct {
	ID             uuid.UUID  `json:"staff_id" db:"staff_id"`
	OrgID          uuid.UUID  `json:"org_id" db:"org_id"`
	SNISIDPersonID *uuid.UUID `json:"snisid_person_id,omitempty" db:"snisid_person_id"`
	FullName       string     `json:"full_name" db:"full_name"`
	Nationality    string     `json:"nationality" db:"nationality"`
	Role           *string    `json:"role,omitempty" db:"role"`
	IsExpatriate   *bool      `json:"is_expatriate,omitempty" db:"is_expatriate"`
	PassportNumber *string    `json:"passport_number,omitempty" db:"passport_number"`
	IsActive       *bool      `json:"is_active,omitempty" db:"is_active"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

type AccessRequest struct {
	ID                uuid.UUID `json:"request_id" db:"request_id"`
	OrgID             uuid.UUID `json:"org_id" db:"org_id"`
	AccessType        *string   `json:"access_type,omitempty" db:"access_type"`
	RequestedZones    []string  `json:"requested_zones" db:"requested_zones"`
	AccessDate        time.Time `json:"access_date" db:"access_date"`
	Purpose           string    `json:"purpose" db:"purpose"`
	Status            *string   `json:"status,omitempty" db:"status"`
	PNHEscortRequired *bool     `json:"pnh_escort_required,omitempty" db:"pnh_escort_required"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

type RegisterOrgRequest struct {
	OrgName            string   `json:"org_name" binding:"required"`
	OrgType            string   `json:"org_type" binding:"required"`
	HeadquarterCountry string   `json:"headquarter_country" binding:"required"`
	HaitiOfficeDept    string   `json:"haiti_office_dept"`
	OperatingDepts     []string `json:"operating_depts"`
	Sectors            []string `json:"sectors"`
	DirectorName       string   `json:"director_name"`
	ContactEmail       string   `json:"contact_email"`
	ContactPhone       string   `json:"contact_phone"`
}

type RegisterStaffRequest struct {
	OrgID          string `json:"org_id" binding:"required"`
	FullName       string `json:"full_name" binding:"required"`
	Nationality    string `json:"nationality" binding:"required"`
	Role           string `json:"role"`
	IsExpatriate   *bool  `json:"is_expatriate"`
	PassportNumber string `json:"passport_number"`
}

type RequestAccessRequest struct {
	OrgID          string   `json:"org_id" binding:"required"`
	AccessType     string   `json:"access_type"`
	RequestedZones []string `json:"requested_zones"`
	AccessDate     string   `json:"access_date" binding:"required"`
	Purpose        string   `json:"purpose" binding:"required"`
}

type ApproveAccessRequest struct {
	Status        string `json:"status" binding:"required"`
	ApprovalNotes string `json:"approval_notes"`
}

type ONGScreeningResult struct {
	OrgID     string   `json:"org_id"`
	OrgName   string   `json:"org_name"`
	Flags     []string `json:"flags"`
	RiskLevel string   `json:"risk_level"`
}

type ONGRepository interface {
	Create(org *Organization) (*Organization, error)
	FindByID(id uuid.UUID) (*Organization, error)
	FindAll() ([]Organization, error)
	FindFlagged() ([]Organization, error)
	FindUnregistered() ([]Organization, error)
	UpdateRiskFlag(id uuid.UUID, flag RiskFlag) error
	CreateStaff(staff *Staff) (*Staff, error)
	CreateAccessRequest(ar *AccessRequest) (*AccessRequest, error)
	UpdateAccessStatus(id uuid.UUID, status string, notes string) error
}
