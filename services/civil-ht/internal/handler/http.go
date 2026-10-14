package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/snisid/civil-ht/internal/domain"
	"github.com/snisid/civil-ht/internal/service"
)

type Handler struct {
	svc *service.CivilService
}

func NewHandler(svc *service.CivilService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/births", h.DeclareBirth)
	r.POST("/deaths", h.DeclareDeath)
	r.POST("/marriages", h.RegisterMarriage)
	r.GET("/acts/:number", h.GetAct)
	r.GET("/acts/citizen/:nin", h.GetCitizenActs)
}

func (h *Handler) DeclareBirth(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var req struct {
		ChildFullName         string `json:"child_full_name"`
		ChildGender           string `json:"child_gender,omitempty"`
		MotherCitizenID       string `json:"mother_citizen_id,omitempty"`
		FatherCitizenID       string `json:"father_citizen_id,omitempty"`
		BirthWeightG          int    `json:"birth_weight_g,omitempty"`
		BirthFacility         string `json:"birth_facility,omitempty"`
		AttendingProfessional string `json:"attending_professional,omitempty"`
		EventDate             string `json:"event_date"`
		RegisteringOffice     string `json:"registering_office"`
		DeptCode              string `json:"dept_code"`
		Commune               string `json:"commune"`
		OfficerName           string `json:"officer_name,omitempty"`
		OfficerID             string `json:"officer_id,omitempty"`
		IsLateDeclaration     bool   `json:"is_late_declaration"`
		IsReconstructed       bool   `json:"is_reconstructed"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	eventDate, err := time.Parse("2006-01-02", req.EventDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_date, use YYYY-MM-DD"})
		return
	}

	birth := domain.BirthDeclaration{
		ChildFullName:         req.ChildFullName,
		ChildGender:           strPtr(req.ChildGender),
		BirthWeightG:          intPtr(req.BirthWeightG),
		BirthFacility:         strPtr(req.BirthFacility),
		AttendingProfessional: strPtr(req.AttendingProfessional),
	}
	if req.MotherCitizenID != "" {
		if id, err := uuid.Parse(req.MotherCitizenID); err == nil {
			birth.MotherCitizenID = &id
		}
	}
	if req.FatherCitizenID != "" {
		if id, err := uuid.Parse(req.FatherCitizenID); err == nil {
			birth.FatherCitizenID = &id
		}
	}

	registerInfo := domain.CivilAct{
		RegisteringOffice: req.RegisteringOffice,
		DeptCode:          req.DeptCode,
		Commune:           req.Commune,
		EventDate:         eventDate,
		OfficerName:       strPtr(req.OfficerName),
		IsLateDeclaration: req.IsLateDeclaration,
		IsReconstructed:   req.IsReconstructed,
	}
	if req.OfficerID != "" {
		if id, err := uuid.Parse(req.OfficerID); err == nil {
			registerInfo.OfficerID = &id
		}
	}

	result, err := h.svc.DeclareBirth(c.Request.Context(), birth, registerInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) DeclareDeath(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var req struct {
		DeceasedCitizenID string `json:"deceased_citizen_id"`
		CauseOfDeath      string `json:"cause_of_death,omitempty"`
		DeathLocation     string `json:"death_location,omitempty"`
		MedicalCertifier  string `json:"medical_certifier,omitempty"`
		IsViolentDeath    bool   `json:"is_violent_death"`
		FIRCaseReference  string `json:"fir_case_reference,omitempty"`
		EventDate         string `json:"event_date"`
		RegisteringOffice string `json:"registering_office"`
		DeptCode          string `json:"dept_code"`
		Commune           string `json:"commune"`
		OfficerName       string `json:"officer_name,omitempty"`
		OfficerID         string `json:"officer_id,omitempty"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	eventDate, err := time.Parse("2006-01-02", req.EventDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_date"})
		return
	}

	death := domain.DeathDeclaration{
		CauseOfDeath:     strPtr(req.CauseOfDeath),
		DeathLocation:    strPtr(req.DeathLocation),
		MedicalCertifier: strPtr(req.MedicalCertifier),
		IsViolentDeath:   req.IsViolentDeath,
		FIRCaseReference: strPtr(req.FIRCaseReference),
	}
	if id, err := uuid.Parse(req.DeceasedCitizenID); err == nil {
		death.DeceasedCitizenID = id
	}

	registerInfo := domain.CivilAct{
		RegisteringOffice: req.RegisteringOffice,
		DeptCode:          req.DeptCode,
		Commune:           req.Commune,
		EventDate:         eventDate,
		OfficerName:       strPtr(req.OfficerName),
	}
	if req.OfficerID != "" {
		if id, err := uuid.Parse(req.OfficerID); err == nil {
			registerInfo.OfficerID = &id
		}
	}

	result, err := h.svc.DeclareDeath(c.Request.Context(), death, registerInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) RegisterMarriage(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var req struct {
		SpouseACitizenID    string `json:"spouse_a_citizen_id"`
		SpouseBCitizenID    string `json:"spouse_b_citizen_id"`
		MarriageRegime      string `json:"marriage_regime,omitempty"`
		PrenuptialAgreement bool   `json:"prenuptial_agreement"`
		EventDate           string `json:"event_date"`
		RegisteringOffice   string `json:"registering_office"`
		DeptCode            string `json:"dept_code"`
		Commune             string `json:"commune"`
		OfficerName         string `json:"officer_name,omitempty"`
		OfficerID           string `json:"officer_id,omitempty"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	eventDate, err := time.Parse("2006-01-02", req.EventDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_date"})
		return
	}

	marriage := domain.MarriageDeclaration{
		MarriageRegime:      strPtr(req.MarriageRegime),
		PrenuptialAgreement: req.PrenuptialAgreement,
	}
	if id, err := uuid.Parse(req.SpouseACitizenID); err == nil {
		marriage.SpouseACitizenID = id
	}
	if id, err := uuid.Parse(req.SpouseBCitizenID); err == nil {
		marriage.SpouseBCitizenID = id
	}

	registerInfo := domain.CivilAct{
		RegisteringOffice: req.RegisteringOffice,
		DeptCode:          req.DeptCode,
		Commune:           req.Commune,
		EventDate:         eventDate,
		OfficerName:       strPtr(req.OfficerName),
	}
	if req.OfficerID != "" {
		if id, err := uuid.Parse(req.OfficerID); err == nil {
			registerInfo.OfficerID = &id
		}
	}

	result, err := h.svc.RegisterMarriage(c.Request.Context(), marriage, registerInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) GetAct(c *gin.Context) {
	actNumber := c.Param("number")
	act, err := h.svc.GetAct(c.Request.Context(), actNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "act not found"})
		return
	}

	var details any
	switch act.ActType {
	case domain.ActBirth:
		details, _ = h.svc.GetBirthDetails(c.Request.Context(), act.ActID)
	case domain.ActDeath:
		details, _ = h.svc.GetDeathDetails(c.Request.Context(), act.ActID)
	case domain.ActMarriage:
		details, _ = h.svc.GetMarriageDetails(c.Request.Context(), act.ActID)
	}

	c.JSON(http.StatusOK, gin.H{"act": act, "details": details})
}

func (h *Handler) GetCitizenActs(c *gin.Context) {
	citizenID := c.Param("nin")
	acts, err := h.svc.GetCitizenActs(c.Request.Context(), citizenID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "citizen not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": acts})
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func intPtr(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}
