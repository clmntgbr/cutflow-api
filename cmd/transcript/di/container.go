package di

import (
	"log"

	cmdmediafile "go-api/internal/application/command/mediafile"
	"go-api/internal/application/event/dedup"
	eventmediafile "go-api/internal/application/event/mediafile"
	"go-api/internal/application/registry"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/infrastructure/assemblyai"
	"go-api/internal/infrastructure/config"
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
		env.TranscriptQueue,
		env.TranscriptRoutingKey,
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
	transcribeHandler := cmdmediafile.NewTranscribeAudioHandler(
		write.NewTranscriptWriteRepository(db),
		minioStorage,
		assemblyai.NewClient(env.AssemblyAIAPIKey),
		outboxRepo,
	)

	reg := registry.NewHandlerRegistry()
	reg.Register(domainmediafile.EventTypeMediaFileTranscriptRequested, dedup.With(
		dedupRepo,
		"transcribe_audio_on_requested",
		eventmediafile.NewTranscribeAudioOnRequestedHandler(transcribeHandler).Handle,
	))

	consumer := rabbitmq.NewConsumer(conn, reg, env.TranscriptConcurrency, env.WorkerMaxRetries)
	return &Container{Consumer: consumer, Conn: conn}
}
