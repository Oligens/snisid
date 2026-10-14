package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/snisid/card-ht/internal/domain"
	"github.com/snisid/card-ht/internal/kafka"
	"github.com/snisid/card-ht/internal/repository"
)

type CardService struct {
	repo     repository.Repository
	producer *kafka.Producer
}

func NewCardService(repo repository.Repository, producer *kafka.Producer) *CardService {
	return &CardService{repo: repo, producer: producer}
}

func generateDocNumber(docType domain.DocType, year int, seq int) string {
	typePrefix := map[domain.DocType]string{
		domain.DocNationalID: "ID",
		domain.DocPassport:   "PP",
		domain.DocResidence:  "RP",
		domain.DocRefugee:    "RD",
	}
	return fmt.Sprintf("HTI-%s-%04d-%06d", typePrefix[docType], year, seq)
}

func (s *CardService) Issue(ctx context.Context, req domain.IssueRequest) (*domain.CardDocument, error) {
	citizenID, err := uuid.Parse(req.CitizenID)
	if err != nil {
		return nil, fmt.Errorf("invalid citizen_id: %w", err)
	}
	createdBy, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("invalid created_by: %w", err)
	}

	docType := domain.DocType(req.DocType)
	now := time.Now()
	expiryDate := time.Date(now.Year()+10, now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	doc := &domain.CardDocument{
		DocumentID:              uuid.New(),
		DocumentNumber:          generateDocNumber(docType, now.Year(), 1),
		DocType:                 docType,
		CitizenID:               citizenID,
		Status:                  domain.CardIssued,
		IssueDate:               now,
		ExpiryDate:              expiryDate,
		IssuingOffice:           strPtr(req.IssuingOffice),
		PersonalizationFacility: "Imprimerie Nationale PAP",
		PhotoRef:                strPtr(req.PhotoRef),
		SignatureRef:            strPtr(req.SignatureRef),
		SLTDReported:            false,
		CreatedBy:               createdBy,
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("create document: %w", err)
	}

	s.publishEvent(ctx, "card.document.issued", doc)
	return doc, nil
}

func (s *CardService) Verify(ctx context.Context, docNumber string) (*domain.CardDocument, error) {
	return s.repo.FindByDocumentNumber(ctx, docNumber)
}

func (s *CardService) ReportLost(ctx context.Context, documentID string) error {
	did, err := uuid.Parse(documentID)
	if err != nil {
		return fmt.Errorf("invalid document_id: %w", err)
	}
	if err := s.repo.UpdateStatus(ctx, did, domain.CardLost); err != nil {
		return fmt.Errorf("report lost: %w", err)
	}
	return nil
}

func (s *CardService) Revoke(ctx context.Context, documentID string) error {
	did, err := uuid.Parse(documentID)
	if err != nil {
		return fmt.Errorf("invalid document_id: %w", err)
	}
	if err := s.repo.UpdateStatus(ctx, did, domain.CardRevoked); err != nil {
		return fmt.Errorf("revoke: %w", err)
	}
	return nil
}

func (s *CardService) publishEvent(ctx context.Context, eventType string, doc *domain.CardDocument) {
	if s.producer == nil {
		return
	}
	evt := kafka.Event{
		EventType:      eventType,
		DocumentID:     doc.DocumentID.String(),
		DocumentNumber: doc.DocumentNumber,
		Timestamp:      time.Now().UTC(),
		Data:           doc,
	}
	if err := s.producer.Publish(ctx, evt); err != nil {
		log.Printf("failed to publish event: %v", err)
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
