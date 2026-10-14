package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snisid/platform/services/fir-svc/internal/service"
	"net/http"
)

type RecordHandler struct {
	recordSvc *service.RecordService
}

func NewRecordHandler(recordSvc *service.RecordService) *RecordHandler {
	return &RecordHandler{recordSvc: recordSvc}
}

type CreateRecordRequest struct {
	SNISIDPersonID string `json:"snisid_person_id" binding:"required"`
}

func (h *RecordHandler) CreateRecord(c *gin.Context) {
	var req CreateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	personID, err := uuid.Parse(req.SNISIDPersonID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID invalide"})
		return
	}

	record, err := h.recordSvc.GetOrCreateRecord(c.Request.Context(), personID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, record)
}

func (h *RecordHandler) GetRecord(c *gin.Context) {
	personIDStr := c.Param("person_id")
	personID, err := uuid.Parse(personIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID invalide"})
		return
	}

	record, err := h.recordSvc.GetRecord(c.Request.Context(), personID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Casier non trouvé"})
		return
	}

	c.JSON(http.StatusOK, record)
}

func (h *RecordHandler) AddArrest(c *gin.Context) {
	recordIDStr := c.Param("id")
	recordID, err := uuid.Parse(recordIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID invalide"})
		return
	}

	var arrest struct {
		ArrestDate     string `json:"arrest_date" binding:"required"`
		ArrestingUnit  string `json:"arresting_unit" binding:"required"`
		ArrestLocation string `json:"arrest_location"`
		DeptCode       string `json:"dept_code"`
		ChargesText    string `json:"charges_text" binding:"required"`
		OffenseClass   string `json:"offense_class" binding:"required"`
		CaseReference  string `json:"case_reference"`
	}
	if err := c.ShouldBindJSON(&arrest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Arrestation ajoutée", "record_id": recordID})
}

func (h *RecordHandler) AddConviction(c *gin.Context) {
	recordIDStr := c.Param("id")
	recordID, err := uuid.Parse(recordIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID invalide"})
		return
	}

	var conviction struct {
		CaseReference      string `json:"case_reference" binding:"required"`
		CourtName          string `json:"court_name" binding:"required"`
		OffenseDescription string `json:"offense_description" binding:"required"`
		VerdictDate        string `json:"verdict_date" binding:"required"`
		CaseStatus         string `json:"case_status" binding:"required"`
		SentenceType       string `json:"sentence_type"`
	}
	if err := c.ShouldBindJSON(&conviction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Condamnation enregistrée", "record_id": recordID})
}

func (h *RecordHandler) GetArrests(c *gin.Context) {
	recordIDStr := c.Param("id")
	recordID, err := uuid.Parse(recordIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID invalide"})
		return
	}

	arrests, err := h.recordSvc.GetArrests(c.Request.Context(), recordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, arrests)
}

func (h *RecordHandler) GetConvictions(c *gin.Context) {
	recordIDStr := c.Param("id")
	recordID, err := uuid.Parse(recordIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID invalide"})
		return
	}

	convictions, err := h.recordSvc.GetConvictions(c.Request.Context(), recordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, convictions)
}
