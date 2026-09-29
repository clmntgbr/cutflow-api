package di

import (
	"log"

	cmdmediafile "go-api/internal/application/command/mediafile"
	"go-api/internal/application/event/dedup"
	eventmediafile "go-api/internal/application/event/mediafile"
	"go-api/internal/application/registry"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/messaging/rabbitmq"
	"go-api/internal/infrastructure/persistence/outbox"
	"go-api/internal/infrastructure/persistence/processed"
	"go-api/internal/infrastructure/persistence/write"

	"gorm.io/gorm"
)

type Container struct {
	Consumer *rabbitmq.Consumer
	Conn     *rabbitmq.Connection
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	topology := rabbitmq.DefaultTopology(
		env.RabbitMQExchange,
		env.AnalysisQueue,
		env.AnalysisRoutingKey,
		env.RabbitMQRetryTTLMS,
	)

	conn, err := rabbitmq.Connect(env.RabbitMQURL, topology)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}

	outboxRepo := outbox.NewRepository(db)
	dedupRepo := processed.NewRepository(db)
	analyzeHandler := cmdmediafile.NewAnalyzeTranscriptHandler(
		write.NewTranscriptWriteRepository(db),
		write.NewDetectedTranscriptIssueWriteRepository(db),
		write.NewMediaConfigurationWriteRepository(db),
		write.NewJobWriteRepository(db),
		outboxRepo,
	)

	reg := registry.NewHandlerRegistry()
	reg.Register(domainmediafile.EventTypeMediaFileAnalysisRequested, dedup.With(
		dedupRepo,
		"analyze_transcript_on_requested",
		eventmediafile.NewAnalyzeTranscriptOnRequestedHandler(analyzeHandler).Handle,
	))

	consumer := rabbitmq.NewConsumer(conn, reg, env.AnalysisConcurrency, env.WorkerMaxRetries)
	return &Container{Consumer: consumer, Conn: conn}
}
