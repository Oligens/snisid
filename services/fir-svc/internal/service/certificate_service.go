package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/platform/services/fir-svc/internal/domain"
)

type CertificateService struct {
	recordRepo domain.CriminalRecordRepository
}

func NewCertificateService(recordRepo domain.CriminalRecordRepository) *CertificateService {
	return &CertificateService{recordRepo: recordRepo}
}

func (s *CertificateService) IssueCertificate(
	ctx context.Context,
	req domain.CertificateRequest,
	issuedBy uuid.UUID,
) (*domain.Certificate, error) {
	record, _ := s.recordRepo.FindByPersonID(ctx, req.PersonID)

	result := domain.CertificateResultClean
	if record != nil {
		hasActive, _ := s.hasActiveConvictions(ctx, record.RecordID)
		if hasActive {
			result = domain.CertificateResultHasRecord
		}
	}

	expiry := time.Now().AddDate(0, 3, 0)
	cert := &domain.Certificate{
		CertID: uuid.New(),
		RecordID: func() *uuid.UUID {
			if record != nil {
				return &record.RecordID
			}
			return nil
		}(),
		SNISIDPersonID:    req.PersonID,
		CertificateNumber: s.generateCertNumber(),
		IssuedFor:         req.Purpose,
		Result:            result,
		IssuedBy:          issuedBy,
		IssuingOffice:     req.Office,
		IssuedAt:          time.Now(),
		ExpiresAt:         &expiry,
		QRCodeRef:         fmt.Sprintf("QR-%s", uuid.New().String()[:8]),
	}

	if err := s.recordRepo.SaveCertificate(ctx, cert); err != nil {
		return nil, fmt.Errorf("émission certificat: %w", err)
	}

	return cert, nil
}

func (s *CertificateService) VerifyCertificate(
	ctx context.Context,
	certNumber string,
) (*domain.Certificate, error) {
	return s.recordRepo.FindCertificateByNumber(ctx, certNumber)
}

func (s *CertificateService) generateCertNumber() string {
	return fmt.Sprintf("CERT-FIR-%s-%s", time.Now().Format("20060102"), uuid.New().String()[:8])
}

func (s *CertificateService) hasActiveConvictions(ctx context.Context, recordID uuid.UUID) (bool, error) {
	record, err := s.recordRepo.FindByID(ctx, recordID)
	if err != nil {
		return false, err
	}
	_ = record
	return false, nil
}
