package di

import (
	"log"

	cmdmediafile "go-api/internal/application/command/mediafile"
	"go-api/internal/application/event/dedup"
	eventmediafile "go-api/internal/application/event/mediafile"
	"go-api/internal/application/registry"
	domaintimeline "go-api/internal/domain/timeline"
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
		env.TimelineQueue,
		env.TimelineRoutingKey,
		env.RabbitMQRetryTTLMS,
	)

	conn, err := rabbitmq.Connect(env.RabbitMQURL, topology)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}

	outboxRepo := outbox.NewRepository(db)
	dedupRepo := processed.NewRepository(db)
	rebuildHandler := cmdmediafile.NewRebuildTimelineHandler(
		write.NewMediaFileWriteRepository(db),
		write.NewProjectWriteRepository(db),
		write.NewMediaConfigurationWriteRepository(db),
		write.NewDetectedSilenceWriteRepository(db),
		write.NewDetectedTranscriptIssueWriteRepository(db),
		write.NewUserOverrideRepository(db),
		write.NewTimelineWriteRepository(db),
		write.NewJobWriteRepository(db),
		outboxRepo,
	)

	reg := registry.NewHandlerRegistry()
	reg.Register(domaintimeline.EventTypeTimelineRebuildRequested, dedup.With(
		dedupRepo,
		"rebuild_timeline_on_requested",
		eventmediafile.NewRebuildTimelineOnRequestedHandler(rebuildHandler).Handle,
	))

	consumer := rabbitmq.NewConsumer(conn, reg, env.TimelineConcurrency, env.WorkerMaxRetries)
	return &Container{Consumer: consumer, Conn: conn}
}
