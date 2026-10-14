package entity

import "time"

type AuditEvent struct {
	EventID       string    `json:"eventId" gorm:"primaryKey"`
	CorrelationID string    `json:"correlationId" gorm:"index"`
	EventType     string    `json:"eventType" gorm:"index"`
	Actor         string    `json:"actor" gorm:"index"`
	Action        string    `json:"action"`
	Resource      string    `json:"resource" gorm:"index"`
	Status        string    `json:"status"`  // success, failed, denied
	Payload       string    `json:"payload"` // JSON stringified for hashing stability
	PreviousHash  string    `json:"previousHash"`
	Hash          string    `json:"hash"`
	SequenceID    int64     `json:"sequenceId" gorm:"autoIncrement;uniqueIndex"`
	Timestamp     time.Time `json:"timestamp"`
}
