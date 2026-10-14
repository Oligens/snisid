package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	nexusv1 "github.com/snisid/platform/api/proto/nexus/v1"
	"google.golang.org/protobuf/proto"
)

// Runtime is the execution environment for a Kai agent.
type Runtime struct {
	ID           string
	Capabilities []string
	KafkaReader  *kafka.Reader
}

func NewRuntime(agentID string, brokers []string) *Runtime {
	return &Runtime{
		ID: agentID,
		KafkaReader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   "snisid.nexus.commands",
			GroupID: "kai-agent-" + agentID,
		}),
	}
}

func (r *Runtime) Start(ctx context.Context) {
	fmt.Printf("🤖 KAI-AGENT: Runtime %s operational. Awaiting tasks...\n", r.ID)

	for {
		msg, err := r.KafkaReader.ReadMessage(ctx)
		if err != nil {
			log.Printf("❌ KAI-AGENT: Message read error: %v", err)
			continue
		}

		var task nexusv1.TaskDefinition
		if err := proto.Unmarshal(msg.Value, &task); err != nil {
			log.Printf("❌ KAI-AGENT: Failed to unmarshal task: %v", err)
			continue
		}

		r.Execute(ctx, &task)
	}
}

func (r *Runtime) Execute(ctx context.Context, task *nexusv1.TaskDefinition) {
	log.Printf("⚡ KAI-AGENT: Executing task %s (Type: %s)", task.Id, task.Type)

	// Simulate execution delay
	time.Sleep(500 * time.Millisecond)

	log.Printf("✅ KAI-AGENT: Task %s completed successfully.", task.Id)

	// TODO: Send Signal back to Kernel
}
