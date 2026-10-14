package grpc

import (
	"context"

	"github.com/snisid/vehicle-criminal-svc/internal/service"
)

type AlertServiceServer struct {
	UnimplementedAlertServiceServer
	alertSvc *service.CriminalAlertService
}

func NewAlertServiceServer(alertSvc *service.CriminalAlertService) *AlertServiceServer {
	return &AlertServiceServer{alertSvc: alertSvc}
}

func (s *AlertServiceServer) CheckPlate(ctx context.Context, req *CheckPlateRequest) (*CheckPlateResponse, error) {
	result, err := s.alertSvc.CheckPlate(ctx, req.PlateNumber)
	if err != nil {
		return nil, err
	}

	resp := &CheckPlateResponse{
		PlateNumber:      result.PlateNumber,
		HasCriminalAlert: result.HasCriminalAlert,
		HasStolenPlate:   result.HasStolenPlate,
		Source:           result.Source,
	}

	if result.Alert != nil {
		resp.AlertLevel = string(result.Alert.AlertLevel)
		resp.CrimeCategory = string(result.Alert.CrimeCategory)
		resp.ArmedAndDangerous = result.Alert.ArmedAndDangerous
		resp.DoNotStopAlone = result.Alert.DoNotStopAlone
		resp.Make = result.Alert.Make
		resp.Model = result.Alert.Model
		resp.ColorPrimary = result.Alert.ColorPrimary
		if result.Alert.OfficerSafetyNotes != nil {
			resp.OfficerSafetyNotes = *result.Alert.OfficerSafetyNotes
		}
		if result.Alert.LastSeenLocation != nil {
			resp.LastSeenLocation = *result.Alert.LastSeenLocation
		}
	}

	return resp, nil
}

func (s *AlertServiceServer) GetAlert(ctx context.Context, req *GetAlertRequest) (*GetAlertResponse, error) {
	alert, err := s.alertSvc.GetAlert(ctx, parseUUID(req.AlertId))
	if err != nil {
		return nil, err
	}
	if alert == nil {
		return nil, nil
	}

	resp := &GetAlertResponse{
		AlertId:           alert.AlertID.String(),
		PlateNumber:       alert.PlateNumber,
		CrimeCategory:     string(alert.CrimeCategory),
		AlertLevel:        string(alert.AlertLevel),
		Status:            string(alert.Status),
		ArmedAndDangerous: alert.ArmedAndDangerous,
		ReportingUnit:     alert.ReportingUnit,
		Make:              alert.Make,
		Model:             alert.Model,
		ColorPrimary:      alert.ColorPrimary,
	}

	return resp, nil
}

func parseUUID(s string) [16]byte {
	var id [16]byte
	copy(id[:], s)
	return id
}
