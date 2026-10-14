package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/snisid/platform/services/sanc-svc/internal/domain"
)

type SanctionsService struct {
	repo domain.SanctionsRepository
	log  *zap.Logger
}

func NewSanctionsService(repo domain.SanctionsRepository, log *zap.Logger) *SanctionsService {
	return &SanctionsService{repo: repo, log: log}
}

type sdnXmlRoot struct {
	XMLName xml.Name   `xml:"sdnList"`
	Entries []sdnEntry `xml:"sdnEntry"`
}

type sdnEntry struct {
	UID       string    `xml:"uid"`
	LastName  string    `xml:"lastName"`
	FirstName string    `xml:"firstName"`
	SDNType   string    `xml:"sdnType"`
	Remarks   string    `xml:"remarks"`
	IDList    []sdnID   `xml:"idList>id"`
	AkaList   []sdnAka  `xml:"akaList>aka"`
	AddrList  []sdnAddr `xml:"addressList>address"`
	DateList  []sdnDate `xml:"dateList>date"`
	NatList   []sdnNat  `xml:"nationalityList>nationality"`
	PassList  []sdnPass `xml:"passportList>passport"`
}

type sdnID struct {
	ID   string `xml:"id"`
	Type string `xml:"idType"`
}

type sdnAka struct {
	Type      string `xml:"type"`
	LastName  string `xml:"lastName"`
	FirstName string `xml:"firstName"`
}

type sdnAddr struct {
	City    string `xml:"city"`
	Country string `xml:"country"`
}

type sdnDate struct {
	DatePart string `xml:"datePart"`
}

type sdnNat struct {
	Country string `xml:"country"`
}

type sdnPass struct {
	PassportNumber string `xml:"passportNumber"`
}

func (s *SanctionsService) SyncOFAC(ctx context.Context) (*domain.SyncResult, error) {
	syncLog := &domain.SyncLog{
		Source:    domain.OFAC_SDN,
		StartedAt: time.Now().UTC(),
		Status:    "RUNNING",
	}
	if err := s.repo.SaveSyncLog(syncLog); err != nil {
		return nil, fmt.Errorf("save sync log: %w", err)
	}

	entries, err := s.fetchOFACXml()
	if err != nil {
		now := time.Now().UTC()
		syncLog.CompletedAt = &now
		syncLog.Status = "FAILED"
		errMsg := err.Error()
		syncLog.ErrorDetails = &errMsg
		_ = s.repo.UpdateSyncLog(syncLog)
		return &domain.SyncResult{Log: syncLog}, fmt.Errorf("fetch OFAC XML: %w", err)
	}

	var added, errors int
	for i := range entries {
		if err := s.repo.UpsertEntry(&entries[i]); err != nil {
			s.log.Error("upsert entry failed", zap.String("source_ref", entries[i].SourceRefID), zap.Error(err))
			errors++
			continue
		}
		added++
	}

	now := time.Now().UTC()
	syncLog.CompletedAt = &now
	syncLog.EntriesProcessed = len(entries)
	syncLog.EntriesAdded = added
	syncLog.Errors = errors
	syncLog.Status = "COMPLETED"
	_ = s.repo.UpdateSyncLog(syncLog)

	return &domain.SyncResult{
		Log:          syncLog,
		EntriesAdded: added,
		Errors:       errors,
	}, nil
}

func (s *SanctionsService) SyncUN2653(ctx context.Context) (*domain.SyncResult, error) {
	syncLog := &domain.SyncLog{
		Source:    domain.UN_2653,
		StartedAt: time.Now().UTC(),
		Status:    "RUNNING",
	}
	if err := s.repo.SaveSyncLog(syncLog); err != nil {
		return nil, fmt.Errorf("save sync log: %w", err)
	}

	now := time.Now().UTC()
	syncLog.CompletedAt = &now
	syncLog.Status = "COMPLETED"
	_ = s.repo.SaveSyncLog(syncLog)

	return &domain.SyncResult{Log: syncLog}, nil
}

func (s *SanctionsService) SyncEU(ctx context.Context) (*domain.SyncResult, error) {
	syncLog := &domain.SyncLog{
		Source:    domain.EU_CONSOLIDATED,
		StartedAt: time.Now().UTC(),
		Status:    "RUNNING",
	}
	if err := s.repo.SaveSyncLog(syncLog); err != nil {
		return nil, fmt.Errorf("save sync log: %w", err)
	}

	now := time.Now().UTC()
	syncLog.CompletedAt = &now
	syncLog.Status = "COMPLETED"
	_ = s.repo.SaveSyncLog(syncLog)

	return &domain.SyncResult{Log: syncLog}, nil
}

func (s *SanctionsService) CheckPersonRealTime(ctx context.Context, personID uuid.UUID) (*domain.PersonSanctionsResult, error) {
	result := &domain.PersonSanctionsResult{
		PersonID: personID,
	}

	entries, _, err := s.repo.GetActiveEntries(1000, 0)
	if err != nil {
		return nil, fmt.Errorf("get active entries: %w", err)
	}

	for _, entry := range entries {
		if entry.SNISIDPersonID != nil && *entry.SNISIDPersonID == personID {
			result.IsSanctioned = true
			result.Entries = append(result.Entries, entry)
		}
	}

	return result, nil
}

func (s *SanctionsService) SearchByName(ctx context.Context, name string, dob *time.Time) ([]domain.SanctionEntry, error) {
	return s.repo.SearchByNameAndDOB(name, dob)
}

func (s *SanctionsService) GetActiveEntries(ctx context.Context, limit, offset int) ([]domain.SanctionEntry, int, error) {
	return s.repo.GetActiveEntries(limit, offset)
}

func (s *SanctionsService) GetEntriesBySource(ctx context.Context, source domain.Source, limit, offset int) ([]domain.SanctionEntry, int, error) {
	return s.repo.GetEntriesBySource(source, limit, offset)
}

func (s *SanctionsService) GetUnconfirmedMatches(ctx context.Context) ([]domain.IdentityMatch, error) {
	return s.repo.GetUnconfirmedMatches()
}

func (s *SanctionsService) ConfirmMatch(ctx context.Context, matchID uuid.UUID, confirmedBy uuid.UUID) error {
	return s.repo.ConfirmMatch(matchID, confirmedBy)
}

func (s *SanctionsService) GetSyncStatus(ctx context.Context) ([]domain.SyncLog, error) {
	return s.repo.GetSyncStatus(10)
}

func (s *SanctionsService) fetchOFACXml() ([]domain.SanctionEntry, error) {
	resp, err := http.Get("https://www.treasury.gov/ofac/downloads/sdn.xml")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var root sdnXmlRoot
	if err := xml.Unmarshal(body, &root); err != nil {
		return nil, fmt.Errorf("parse XML: %w", err)
	}

	var entries []domain.SanctionEntry
	for _, e := range root.Entries {
		entry := domain.SanctionEntry{
			Source:      domain.OFAC_SDN,
			SourceRefID: fmt.Sprintf("OFAC-%s", e.UID),
			EntityType:  mapSDNType(e.SDNType),
			EntityName:  buildName(e.FirstName, e.LastName),
			IsActive:    true,
			ListingDate: time.Now().UTC(),
		}

		if entry.EntityName == "" {
			entry.EntityName = e.Remarks
		}

		for _, a := range e.AkaList {
			alias := buildName(a.FirstName, a.LastName)
			if alias != "" {
				entry.Aliases = append(entry.Aliases, alias)
			}
		}

		for _, n := range e.NatList {
			if n.Country != "" {
				entry.Nationality = append(entry.Nationality, n.Country)
			}
		}

		for _, p := range e.PassList {
			if p.PassportNumber != "" {
				entry.PassportNumbers = append(entry.PassportNumbers, p.PassportNumber)
			}
		}

		for _, id := range e.IDList {
			if id.ID != "" {
				entry.NationalIDNumbers = append(entry.NationalIDNumbers, id.ID)
			}
		}

		for _, d := range e.DateList {
			if d.DatePart != "" {
				if t, err := time.Parse("01/02/2006", d.DatePart); err == nil {
					entry.DateOfBirth = &t
					break
				}
			}
		}

		if len(e.AddrList) > 0 {
			addr := e.AddrList[0]
			if addr.City != "" || addr.Country != "" {
				pob := fmt.Sprintf("%s, %s", addr.City, addr.Country)
				entry.PlaceOfBirth = &pob
			}
		}

		if e.Remarks != "" {
			entry.ListingReason = &e.Remarks
		}

		entry.MeasureTypes = []domain.Measure{domain.ALL_MEASURES}

		entries = append(entries, entry)
	}

	return entries, nil
}

func mapSDNType(sdnType string) domain.EntityType {
	switch strings.ToUpper(sdnType) {
	case "INDIVIDUAL", "PERSON":
		return domain.INDIVIDUAL
	case "ENTITY", "ORGANIZATION":
		return domain.ORGANIZATION
	case "VESSEL":
		return domain.VESSEL
	case "AIRCRAFT":
		return domain.AIRCRAFT
	default:
		return domain.INDIVIDUAL
	}
}

func buildName(first, last string) string {
	first = strings.TrimSpace(first)
	last = strings.TrimSpace(last)
	if first != "" && last != "" {
		return first + " " + last
	}
	if first != "" {
		return first
	}
	return last
}
