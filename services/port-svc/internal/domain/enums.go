package domain

type RiskLevel string

const (
	LOW      RiskLevel = "LOW"
	MEDIUM   RiskLevel = "MEDIUM"
	HIGH     RiskLevel = "HIGH"
	CRITICAL RiskLevel = "CRITICAL"
)

type ContainerStatus string

const (
	PENDING_INSPECTION        ContainerStatus = "PENDING_INSPECTION"
	CLEARED                   ContainerStatus = "CLEARED"
	HELD_FOR_INSPECTION       ContainerStatus = "HELD_FOR_INSPECTION"
	SEIZED                    ContainerStatus = "SEIZED"
	RELEASED_AFTER_INSPECTION ContainerStatus = "RELEASED_AFTER_INSPECTION"
)
