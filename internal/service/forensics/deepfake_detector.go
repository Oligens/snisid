package forensics

import (
	"context"
	"fmt"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type DeepfakeDetector struct {
	engine ForensicEngine
}

func NewDeepfakeDetector(engine ForensicEngine) *DeepfakeDetector {
	return &DeepfakeDetector{engine: engine}
}

func (d *DeepfakeDetector) Detect(ctx context.Context, identityID string, mediaData []byte) (float32, []string, error) {
	logger.Info(ctx, "Starting deepfake analysis", zap.String("identity_id", identityID))

	// 1. Preprocessing (Mock: in prod use OpenCV/ffmpeg to extract faces)

	// 2. Inference
	result, err := d.engine.Analyze(ctx, mediaData)
	if err != nil {
		return 0, nil, fmt.Errorf("forensic inference failed: %w", err)
	}

	logger.Info(ctx, "Deepfake analysis complete",
		zap.String("identity_id", identityID),
		zap.Float64("probability", result.DeepfakeProbability),
	)

	return float32(result.DeepfakeProbability), result.Anomalies, nil
}
