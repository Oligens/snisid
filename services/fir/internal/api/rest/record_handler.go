package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snisid/platform/services/fir/internal/service"
)

type RecordHandler struct {
	recordSvc *service.RecordService
	chargeSvc *service.ChargeService
}

func NewRecordHandler(rs *service.RecordService, cs *service.ChargeService) *RecordHandler {
	return &RecordHandler{recordSvc: rs, chargeSvc: cs}
}

func (h *RecordHandler) CreateRecord(c *gin.Context) {
	var req struct {
		SNISIDPersonID uuid.UUID  `json:"snisid_person_id" binding:"required"`
		IsHaitian      bool       `json:"is_haitian_national"`
		AFISSubjectID  *uuid.UUID `json:"afis_subject_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := h.recordSvc.Create(c.Request.Context(), req.SNISIDPersonID, req.IsHaitian, req.AFISSubjectID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, record)
}

func (h *RecordHandler) GetRecord(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	record, err := h.recordSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	charges, _ := h.chargeSvc.ListByRecord(c.Request.Context(), id)

	c.JSON(http.StatusOK, gin.H{
		"record":  record,
		"charges": charges,
	})
}

func (h *RecordHandler) GetRecordByPerson(c *gin.Context) {
	personID, err := uuid.Parse(c.Param("person_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID personne invalide"})
		return
	}

	record, err := h.recordSvc.GetByPersonID(c.Request.Context(), personID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	charges, _ := h.chargeSvc.ListByRecord(c.Request.Context(), record.RecordID)

	c.JSON(http.StatusOK, gin.H{
		"record":  record,
		"charges": charges,
	})
}

func (h *RecordHandler) ListRecords(c *gin.Context) {
	records, err := h.recordSvc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"records": records})
}

func (h *RecordHandler) ExpungeRecord(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	record, err := h.recordSvc.Expunge(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, record)
}

func (h *RecordHandler) AddArrest(c *gin.Context) {
	recordID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID casier invalide"})
		return
	}

	var req struct {
		ArrestDate     string `json:"arrest_date"`
		ArrestingUnit  string `json:"arresting_unit"`
		ArrestLocation string `json:"arrest_location"`
		DeptCode       string `json:"dept_code"`
		ChargesText    string `json:"charges_text"`
		OffenseClass   string `json:"offense_class"`
		CaseReference  string `json:"case_reference"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{
		"error":     "non implémenté - utiliser le endpoint charges",
		"record_id": recordID,
	})
}

func (h *RecordHandler) AddConviction(c *gin.Context) {
	recordID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID casier invalide"})
		return
	}

	var req struct {
		CaseReference string `json:"case_reference"`
		CourtName     string `json:"court_name"`
		OffenseClass  string `json:"offense_class"`
		VerdictDate   string `json:"verdict_date"`
		SentenceType  string `json:"sentence_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{
		"error":     "non implémenté - utiliser le endpoint charges",
		"record_id": recordID,
	})
}
