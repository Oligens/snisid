package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snisid/platform/services/fir-svc/internal/service"
	"net/http"
)

type SearchHandler struct {
	recordSvc *service.RecordService
}

func NewSearchHandler(recordSvc *service.RecordService) *SearchHandler {
	return &SearchHandler{recordSvc: recordSvc}
}

func (h *SearchHandler) Search(c *gin.Context) {
	firID := c.Query("fir_id")
	personIDStr := c.Query("person_id")

	if firID == "" && personIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fir_id ou person_id requis"})
		return
	}

	if personIDStr != "" {
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

		c.JSON(http.StatusOK, gin.H{"record": record})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Recherche par FIR ID", "fir_id": firID})
}

func (h *SearchHandler) GetRecordByFIRID(c *gin.Context) {
	firID := c.Param("fir_id")
	if firID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "FIR ID requis"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"fir_id": firID, "message": "Recherche par FIR ID"})
}
