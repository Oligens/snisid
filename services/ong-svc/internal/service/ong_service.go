package service

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/ong-svc/internal/domain"
)

type ONGService struct {
	repo domain.ONGRepository
	log  *zap.Logger
}

func NewONGService(repo domain.ONGRepository, log *zap.Logger) *ONGService {
	return &ONGService{repo: repo, log: log}
}

func (s *ONGService) RegisterOrg(req *domain.RegisterOrgRequest) (*domain.Organization, error) {
	org := &domain.Organization{
		OrgName:            req.OrgName,
		OrgType:            domain.ONGType(req.OrgType),
		RegistrationStatus: domain.Pending,
		HeadquarterCountry: req.HeadquarterCountry,
		OperatingDepts:     req.OperatingDepts,
		Sectors:            req.Sectors,
		RiskFlag:           domain.None,
		CreatedBy:          uuid.New(),
	}

	if req.HaitiOfficeDept != "" {
		org.HaitiOfficeDept = &req.HaitiOfficeDept
	}
	if req.DirectorName != "" {
		org.DirectorName = &req.DirectorName
	}
	if req.ContactEmail != "" {
		org.ContactEmail = &req.ContactEmail
	}
	if req.ContactPhone != "" {
		org.ContactPhone = &req.ContactPhone
	}

	return s.repo.Create(org)
}

func (s *ONGService) GetOrg(id uuid.UUID) (*domain.Organization, error) {
	return s.repo.FindByID(id)
}

func (s *ONGService) ListOrgs() ([]domain.Organization, error) {
	return s.repo.FindAll()
}

func (s *ONGService) ListFlagged() ([]domain.Organization, error) {
	return s.repo.FindFlagged()
}

func (s *ONGService) ListUnregistered() ([]domain.Organization, error) {
	return s.repo.FindUnregistered()
}

func (s *ONGService) ScreenOrg(id uuid.UUID) (*domain.ONGScreeningResult, error) {
	org, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	result := &domain.ONGScreeningResult{
		OrgID:     id.String(),
		OrgName:   org.OrgName,
		Flags:     []string{},
		RiskLevel: "NONE",
	}

	if org.RegistrationStatus == domain.OperatingWithoutReg {
		result.Flags = append(result.Flags, "UNREGISTERED_ILLEGAL")
	}

	if org.IsAccessRestricted != nil && *org.IsAccessRestricted {
		result.Flags = append(result.Flags, "ACCESS_RESTRICTED")
		result.RiskLevel = "HIGH"
	}

	if org.RiskFlag != domain.None {
		result.Flags = append(result.Flags, string(org.RiskFlag))
		result.RiskLevel = "HIGH"
	}

	return result, nil
}

func (s *ONGService) RegisterStaff(req *domain.RegisterStaffRequest) (*domain.Staff, error) {
	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		return nil, err
	}

	staff := &domain.Staff{
		OrgID:       orgID,
		FullName:    req.FullName,
		Nationality: req.Nationality,
	}

	if req.Role != "" {
		staff.Role = &req.Role
	}
	if req.IsExpatriate != nil {
		staff.IsExpatriate = req.IsExpatriate
	}
	if req.PassportNumber != "" {
		staff.PassportNumber = &req.PassportNumber
	}
	isActive := true
	staff.IsActive = &isActive

	return s.repo.CreateStaff(staff)
}

func (s *ONGService) RequestAccess(req *domain.RequestAccessRequest) (*domain.AccessRequest, error) {
	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		return nil, err
	}

	ad, _ := time.Parse("2006-01-02", req.AccessDate)

	ar := &domain.AccessRequest{
		OrgID:          orgID,
		RequestedZones: req.RequestedZones,
		AccessDate:     ad,
		Purpose:        req.Purpose,
	}

	if req.AccessType != "" {
		ar.AccessType = &req.AccessType
	}
	status := "PENDING"
	ar.Status = &status

	return s.repo.CreateAccessRequest(ar)
}

func (s *ONGService) ApproveAccess(id uuid.UUID, req *domain.ApproveAccessRequest) error {
	return s.repo.UpdateAccessStatus(id, req.Status, req.ApprovalNotes)
}
