package bio_adn

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type KafkaConsumerConfig struct {
	Brokers []string
	GroupID string
	Topic   string
	Handler func(ctx context.Context, event BioEvent) error
}

type BioEvent struct {
	Topic          string                 `json:"topic"`
	EventID        string                 `json:"event_id"`
	EventType      string                 `json:"event_type"`
	Source         string                 `json:"source"`
	Timestamp      time.Time              `json:"timestamp"`
	Payload        map[string]interface{} `json:"payload"`
	Classification string                 `json:"classification"`
}

type BioADNEventProcessor struct {
	dnaRepo      *DNARepository
	personRepo   *PersonRepository
	propertyRepo *PropertyRepository
}

func NewBioADNEventProcessor(dna *DNARepository, person *PersonRepository, property *PropertyRepository) *BioADNEventProcessor {
	return &BioADNEventProcessor{
		dnaRepo:      dna,
		personRepo:   person,
		propertyRepo: property,
	}
}

func (p *BioADNEventProcessor) HandleDNAProfileCreated(ctx context.Context, event BioEvent) error {
	profileID, _ := event.Payload["profile_id"].(string)
	niu, _ := event.Payload["niu"].(string)
	log.Printf("[BIO-ADN] DNA profile created: %s (NIU: %s)", profileID, niu)
	return nil
}

func (p *BioADNEventProcessor) HandleDNAMatchFound(ctx context.Context, event BioEvent) error {
	matchID, _ := event.Payload["match_id"].(string)
	score, _ := event.Payload["match_score"].(float64)
	log.Printf("[BIO-ADN] DNA match found: %s (score: %.4f)", matchID, score)
	return nil
}

func (p *BioADNEventProcessor) HandleDNAMatchReviewed(ctx context.Context, event BioEvent) error {
	matchID, _ := event.Payload["match_id"].(string)
	status, _ := event.Payload["review_status"].(string)
	log.Printf("[BIO-ADN] DNA match reviewed: %s → %s", matchID, status)
	return nil
}

func (p *BioADNEventProcessor) HandlePersonRecordCreated(ctx context.Context, event BioEvent) error {
	recordID, _ := event.Payload["record_id"].(string)
	recordType, _ := event.Payload["record_type"].(string)
	log.Printf("[BIO-ADN] Person record created: %s (type: %s)", recordID, recordType)
	return nil
}

func (p *BioADNEventProcessor) HandlePersonRecordUpdated(ctx context.Context, event BioEvent) error {
	recordID, _ := event.Payload["record_id"].(string)
	log.Printf("[BIO-ADN] Person record updated: %s", recordID)
	return nil
}

func (p *BioADNEventProcessor) HandlePersonLocated(ctx context.Context, event BioEvent) error {
	niu, _ := event.Payload["niu"].(string)
	location, _ := event.Payload["location"].(string)
	log.Printf("[BIO-ADN] Person located: NIU %s at %s", niu, location)
	return nil
}

func (p *BioADNEventProcessor) HandlePersonApprehended(ctx context.Context, event BioEvent) error {
	niu, _ := event.Payload["niu"].(string)
	log.Printf("[BIO-ADN] Person apprehended: NIU %s", niu)
	return nil
}

func (p *BioADNEventProcessor) HandlePropertyStolen(ctx context.Context, event BioEvent) error {
	recordID, _ := event.Payload["record_id"].(string)
	log.Printf("[BIO-ADN] Property reported stolen: %s", recordID)
	return nil
}

func (p *BioADNEventProcessor) HandlePropertyRecovered(ctx context.Context, event BioEvent) error {
	recordID, _ := event.Payload["record_id"].(string)
	log.Printf("[BIO-ADN] Property recovered: %s", recordID)
	return nil
}

func (p *BioADNEventProcessor) HandleFPRWantedEvent(ctx context.Context, event BioEvent) error {
	niu, _ := event.Payload["niu"].(string)
	log.Printf("[BIO-ADN] FPR wanted event for NIU: %s", niu)
	return nil
}

func (p *BioADNEventProcessor) HandleLAPIPlateQuery(ctx context.Context, event BioEvent) error {
	plate, _ := event.Payload["plate_number"].(string)
	log.Printf("[BIO-ADN] LAPI plate query: %s", plate)
	return nil
}

func (p *BioADNEventProcessor) HandleFOVeSVehicleEvent(ctx context.Context, event BioEvent) error {
	vin, _ := event.Payload["vin"].(string)
	log.Printf("[BIO-ADN] FOVeS vehicle event: VIN %s", vin)
	return nil
}

func (p *BioADNEventProcessor) HandleIdentityVerified(ctx context.Context, event BioEvent) error {
	niu, _ := event.Payload["niu"].(string)
	log.Printf("[BIO-ADN] Identity verified: NIU %s", niu)
	return nil
}

func (p *BioADNEventProcessor) HandleDCPJIntelEvent(ctx context.Context, event BioEvent) error {
	alertID, _ := event.Payload["alert_id"].(string)
	log.Printf("[BIO-ADN] DCPJ intel event: %s", alertID)
	return nil
}

func (p *BioADNEventProcessor) HandleMLFraudScored(ctx context.Context, event BioEvent) error {
	niu, _ := event.Payload["niu"].(string)
	score, _ := event.Payload["fraud_score"].(float64)
	log.Printf("[BIO-ADN] ML fraud scored for NIU %s: %.4f", niu, score)
	return nil
}

func (p *BioADNEventProcessor) HandleMLRouting(ctx context.Context, event BioEvent) error {
	modelName, _ := event.Payload["model_name"].(string)
	log.Printf("[BIO-ADN] ML routing decision: model %s", modelName)
	return nil
}

func (p *BioADNEventProcessor) HandleMLFeatureUpdated(ctx context.Context, event BioEvent) error {
	feature, _ := event.Payload["feature_name"].(string)
	niu, _ := event.Payload["niu"].(string)
	log.Printf("[BIO-ADN] ML feature updated: %s for NIU %s", feature, niu)
	return nil
}

var topicHandlers = map[string]func(ctx context.Context, event BioEvent) error{
	"snisid.bio.dna.profile.created":   nil,
	"snisid.bio.dna.match.found":       nil,
	"snisid.bio.dna.match.reviewed":    nil,
	"snisid.bio.person.record.created": nil,
	"snisid.bio.person.record.updated": nil,
	"snisid.bio.person.located":        nil,
	"snisid.bio.person.apprehended":    nil,
	"snisid.bio.property.stolen":       nil,
	"snisid.bio.property.recovered":    nil,
	"snisid.fpr.wanted.events":         nil,
	"snisid.lapi.plate.query":          nil,
	"snisid.foves.vehicle.events":      nil,
	"snisid.identity.verified":         nil,
	"snisid.dcpj.intel.events":         nil,
	"snisid.ml.fraud.scored":           nil,
	"snisid.ml.routing":                nil,
	"snisid.ml.feature.updated":        nil,
}

func (p *BioADNEventProcessor) RegisterHandlers() {
	topicHandlers["snisid.bio.dna.profile.created"] = p.HandleDNAProfileCreated
	topicHandlers["snisid.bio.dna.match.found"] = p.HandleDNAMatchFound
	topicHandlers["snisid.bio.dna.match.reviewed"] = p.HandleDNAMatchReviewed
	topicHandlers["snisid.bio.person.record.created"] = p.HandlePersonRecordCreated
	topicHandlers["snisid.bio.person.record.updated"] = p.HandlePersonRecordUpdated
	topicHandlers["snisid.bio.person.located"] = p.HandlePersonLocated
	topicHandlers["snisid.bio.person.apprehended"] = p.HandlePersonApprehended
	topicHandlers["snisid.bio.property.stolen"] = p.HandlePropertyStolen
	topicHandlers["snisid.bio.property.recovered"] = p.HandlePropertyRecovered
	topicHandlers["snisid.fpr.wanted.events"] = p.HandleFPRWantedEvent
	topicHandlers["snisid.lapi.plate.query"] = p.HandleLAPIPlateQuery
	topicHandlers["snisid.foves.vehicle.events"] = p.HandleFOVeSVehicleEvent
	topicHandlers["snisid.identity.verified"] = p.HandleIdentityVerified
	topicHandlers["snisid.dcpj.intel.events"] = p.HandleDCPJIntelEvent
	topicHandlers["snisid.ml.fraud.scored"] = p.HandleMLFraudScored
	topicHandlers["snisid.ml.routing"] = p.HandleMLRouting
	topicHandlers["snisid.ml.feature.updated"] = p.HandleMLFeatureUpdated
}

func StartConsumer(ctx context.Context, brokers []string, groupID, topic string, handler func(ctx context.Context, event BioEvent) error) error {
	r := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		Topic:       topic,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafkago.LastOffset,
	})

	defer r.Close()

	log.Printf("[BIO-ADN] Consumer started on topic: %s (group: %s)", topic, groupID)

	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("read message from %s: %w", topic, err)
		}

		var event BioEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("[BIO-ADN] Failed to unmarshal event on %s: %v", topic, err)
			continue
		}

		event.Topic = topic

		if err := handler(ctx, event); err != nil {
			log.Printf("[BIO-ADN] Handler error on %s: %v", topic, err)
			continue
		}
	}
}

func StartAllConsumers(ctx context.Context, brokers []string, groupID string) error {
	processor := &BioADNEventProcessor{}
	processor.RegisterHandlers()

	for topic, handler := range topicHandlers {
		if handler == nil {
			continue
		}
		t := topic
		h := handler
		go func() {
			if err := StartConsumer(ctx, brokers, groupID, t, h); err != nil {
				log.Printf("[BIO-ADN] Consumer for %s exited: %v", t, err)
			}
		}()
	}

	<-ctx.Done()
	return ctx.Err()
}
