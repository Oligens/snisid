package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/snisid/idcore-svc/internal/domain"
	"github.com/snisid/idcore-svc/internal/service"
)

type Handler struct {
	svc *service.IdentityService
}

func NewHandler(svc *service.IdentityService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/enroll", h.Enroll)
	r.GET("/citizens/:nin", h.GetByNIN)
	r.GET("/citizens/search", h.Search)
	r.POST("/dedup/resolve", h.ResolveDedup)
	r.PATCH("/citizens/:nin/status", h.UpdateStatus)
	r.GET("/citizens/:nin/history", h.GetHistory)
	r.GET("/stats/population", h.GetPopulationStats)
}

func (h *Handler) Enroll(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	var req struct {
		EnrollmentType  string `json:"enrollment_type"`
		FullNameLegal   string `json:"full_name_legal"`
		FirstName       string `json:"first_name"`
		MiddleNames     string `json:"middle_names,omitempty"`
		LastName        string `json:"last_name"`
		MaidenName      string `json:"maiden_name,omitempty"`
		DOB             string `json:"dob"`
		PobCommune      string `json:"pob_commune,omitempty"`
		PobDeptCode     string `json:"pob_dept_code,omitempty"`
		Gender          string `json:"gender,omitempty"`
		Nationality     string `json:"nationality"`
		DeptCode        string `json:"dept_code"`
		CurrentAddress  string `json:"current_address,omitempty"`
		CurrentCommune  string `json:"current_commune,omitempty"`
		PhotoRef        string `json:"photo_ref,omitempty"`
		MotherNIN       string `json:"mother_nin,omitempty"`
		FatherNIN       string `json:"father_nin,omitempty"`
		CreatedBy       string `json:"created_by"`
		BiometricSample []byte `json:"biometric_sample,omitempty"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
		return
	}

	dob, err := time.Parse("2006-01-02", req.DOB)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	age := int(time.Since(dob).Hours() / 24 / 365)

	enrollReq := domain.EnrollmentRequest{
		Age:             age,
		EnrollmentType:  domain.EnrollmentType(req.EnrollmentType),
		FullNameLegal:   req.FullNameLegal,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		DOB:             dob,
		Gender:          strPtr(req.Gender),
		Nationality:     req.Nationality,
		DeptCode:        req.DeptCode,
		CurrentAddress:  strPtr(req.CurrentAddress),
		PhotoRef:        strPtr(req.PhotoRef),
		BiometricSample: req.BiometricSample,
		CreatedBy:       req.CreatedBy,
	}

	if req.MiddleNames != "" {
		enrollReq.MiddleNames = &req.MiddleNames
	}
	if req.MaidenName != "" {
		enrollReq.MaidenName = &req.MaidenName
	}
	if req.PobCommune != "" {
		enrollReq.PobCommune = &req.PobCommune
	}
	if req.PobDeptCode != "" {
		enrollReq.PobDeptCode = &req.PobDeptCode
	}
	if req.CurrentCommune != "" {
		enrollReq.CurrentCommune = &req.CurrentCommune
	}
	if req.MotherNIN != "" {
		enrollReq.MotherNIN = &req.MotherNIN
	}
	if req.FatherNIN != "" {
		enrollReq.FatherNIN = &req.FatherNIN
	}
	if req.Nationality == "" {
		enrollReq.Nationality = "HTI"
	}

	result, err := h.svc.EnrollCitizen(c.Request.Context(), enrollReq)
	if err != nil {
		if err == domain.ErrDuplicateDetected {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *Handler) GetByNIN(c *gin.Context) {
	nin := c.Param("nin")
	citizen, err := h.svc.VerifyIdentity(c.Request.Context(), nin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "citizen not found"})
		return
	}
	c.JSON(http.StatusOK, citizen)
}

func (h *Handler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	citizens, err := h.svc.SearchCitizens(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": citizens})
}

func (h *Handler) ResolveDedup(c *gin.Context) {
	var req struct {
		CandidateID string `json:"candidate_id"`
		Resolution  string `json:"resolution"`
		ReviewedBy  string `json:"reviewed_by"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.ResolveDedup(c.Request.Context(), req.CandidateID, req.Resolution, req.ReviewedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "resolved"})
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	nin := c.Param("nin")

	var req struct {
		Status       string `json:"status"`
		Reason       string `json:"reason"`
		AuthorizedBy string `json:"authorized_by"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateStatus(c.Request.Context(), nin, domain.IDStatus(req.Status), req.Reason, req.AuthorizedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *Handler) GetHistory(c *gin.Context) {
	nin := c.Param("nin")
	history, err := h.svc.GetHistory(c.Request.Context(), nin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "citizen not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": history})
}

func (h *Handler) GetPopulationStats(c *gin.Context) {
	stats, err := h.svc.GetPopulationStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
