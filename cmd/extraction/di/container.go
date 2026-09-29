package di

import (
	"log"

	cmdmediafile "go-api/internal/application/command/mediafile"
	"go-api/internal/application/event/dedup"
	eventmediafile "go-api/internal/application/event/mediafile"
	"go-api/internal/application/registry"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/infrastructure/config"
	inframedia "go-api/internal/infrastructure/media"
	"go-api/internal/infrastructure/messaging/rabbitmq"
	"go-api/internal/infrastructure/persistence/outbox"
	"go-api/internal/infrastructure/persistence/processed"
	"go-api/internal/infrastructure/persistence/write"
	"go-api/internal/infrastructure/storage"

	"gorm.io/gorm"
)

type Container struct {
	Consumer *rabbitmq.Consumer
	Conn     *rabbitmq.Connection
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	topology := rabbitmq.DefaultTopology(
		env.RabbitMQExchange,
		env.ExtractionQueue,
		env.ExtractionRoutingKey,
		env.RabbitMQRetryTTLMS,
	)

	conn, err := rabbitmq.Connect(env.RabbitMQURL, topology)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}

	minioStorage, err := storage.NewMinIOStorage(env)
	if err != nil {
		log.Fatalf("failed to create storage client: %v", err)
	}

	outboxRepo := outbox.NewRepository(db)
	dedupRepo := processed.NewRepository(db)
	mediaFileWriteRepo := write.NewMediaFileWriteRepository(db)
	mediaAudioWriteRepo := write.NewMediaAudioWriteRepository(db)

	extractAudioHandler := cmdmediafile.NewExtractAudioHandler(
		mediaFileWriteRepo,
		mediaAudioWriteRepo,
		minioStorage,
		inframedia.NewAudioExtractor(),
		outboxRepo,
	)

	reg := registry.NewHandlerRegistry()
	reg.Register(domainmediafile.EventTypeMediaFileSegmentsReady, dedup.With(
		dedupRepo,
		"extract_audio_on_segments_ready",
		eventmediafile.NewExtractAudioOnSegmentsReadyHandler(extractAudioHandler).Handle,
	))

	consumer := rabbitmq.NewConsumer(conn, reg, env.ExtractionConcurrency, env.WorkerMaxRetries)

	return &Container{
		Consumer: consumer,
		Conn:     conn,
	}
}
