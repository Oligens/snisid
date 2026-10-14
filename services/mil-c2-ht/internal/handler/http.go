package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/snisid/mil-c2-ht/internal/domain"
	"github.com/snisid/mil-c2-ht/internal/service"
)

type MilC2Handler struct {
	svc service.MilC2ServiceInterface
}

func NewMilC2Handler(svc service.MilC2ServiceInterface) *MilC2Handler {
	return &MilC2Handler{svc: svc}
}

type createUnitReq struct {
	UnitName          string  `json:"unit_name" binding:"required"`
	Branch            string  `json:"branch" binding:"required"`
	ParentUnitID      string  `json:"parent_unit_id"`
	CommanderName     string  `json:"commander_name"`
	PersonnelCount    int     `json:"personnel_count"`
	LocationLat       float64 `json:"location_lat"`
	LocationLng       float64 `json:"location_lng"`
	OperationalStatus string  `json:"operational_status"`
	EquipmentSummary  string  `json:"equipment_summary"`
}

func (h *MilC2Handler) CreateUnit(c *gin.Context) {
	var req createUnitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	unit := domain.MilitaryUnit{
		UnitName:          req.UnitName,
		Branch:            domain.Branch(req.Branch),
		CommanderName:     req.CommanderName,
		PersonnelCount:    req.PersonnelCount,
		LocationLat:       req.LocationLat,
		LocationLng:       req.LocationLng,
		OperationalStatus: domain.OperationalStatus(req.OperationalStatus),
		EquipmentSummary:  req.EquipmentSummary,
	}
	if req.ParentUnitID != "" {
		pid, err := uuid.Parse(req.ParentUnitID)
		if err == nil {
			unit.ParentUnitID = &pid
		}
	}
	if err := h.svc.CreateUnit(unit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"unit_id": unit.UnitID})
}

func (h *MilC2Handler) GetDeployedUnits(c *gin.Context) {
	units, err := h.svc.GetDeployedUnits()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, units)
}

type createOperationReq struct {
	OperationName     string `json:"operation_name" binding:"required"`
	OperationType     string `json:"operation_type" binding:"required"`
	CommanderID       string `json:"commander_id" binding:"required"`
	StartDate         string `json:"start_date" binding:"required"`
	ExpectedEndDate   string `json:"expected_end_date"`
	OperationalArea   string `json:"operational_area"`
	RulesOfEngagement string `json:"rules_of_engagement"`
	MissionObjective  string `json:"mission_objective"`
}

func (h *MilC2Handler) CreateOperation(c *gin.Context) {
	var req createOperationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	commanderID, err := uuid.Parse(req.CommanderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid commander_id"})
		return
	}
	startDate, err := parseTime(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date"})
		return
	}
	op := domain.Operation{
		OperationName:     req.OperationName,
		OperationType:     domain.OperationType(req.OperationType),
		CommanderID:       commanderID,
		StartDate:         startDate,
		OperationalArea:   req.OperationalArea,
		RulesOfEngagement: req.RulesOfEngagement,
		MissionObjective:  req.MissionObjective,
	}
	if req.ExpectedEndDate != "" {
		t, err := parseTime(req.ExpectedEndDate)
		if err == nil {
			op.ExpectedEndDate = &t
		}
	}
	if err := h.svc.CreateOperation(op); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"operation_id": op.OperationID})
}

func (h *MilC2Handler) GetActiveOperations(c *gin.Context) {
	ops, err := h.svc.GetActiveOperations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ops)
}

type submitReportReq struct {
	ReportingUnitID      string  `json:"reporting_unit_id" binding:"required"`
	ReportType           string  `json:"report_type" binding:"required"`
	PositionLat          float64 `json:"position_lat"`
	PositionLng          float64 `json:"position_lng"`
	EnemyActivity        string  `json:"enemy_activity"`
	CivilianInteractions string  `json:"civilian_interactions"`
	Casualties           int     `json:"casualties"`
	Detainees            int     `json:"detainees"`
	EquipmentStatus      string  `json:"equipment_status"`
}

func (h *MilC2Handler) SubmitReport(c *gin.Context) {
	opID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid operation id"})
		return
	}
	var req submitReportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	unitID, err := uuid.Parse(req.ReportingUnitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reporting_unit_id"})
		return
	}
	report := domain.TacticalReport{
		ReportingUnitID:      unitID,
		ReportType:           domain.ReportType(req.ReportType),
		PositionLat:          req.PositionLat,
		PositionLng:          req.PositionLng,
		EnemyActivity:        req.EnemyActivity,
		CivilianInteractions: req.CivilianInteractions,
		Casualties:           req.Casualties,
		Detainees:            req.Detainees,
		EquipmentStatus:      req.EquipmentStatus,
	}
	if err := h.svc.SubmitReport(opID, report); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"report_id": report.ReportID})
}

func (h *MilC2Handler) GetOperationTimeline(c *gin.Context) {
	opID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid operation id"})
		return
	}
	reports, err := h.svc.GetOperationTimeline(opID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reports)
}

func (h *MilC2Handler) GetCommonOperatingPicture(c *gin.Context) {
	cop, err := h.svc.GetCommonOperatingPicture()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cop)
}

func parseTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05Z", s)
	}
	return t, err
}
