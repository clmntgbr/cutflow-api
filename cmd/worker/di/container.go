package di

import (
	"log"
	"time"

	"go-api/internal/application/event/dedup"
	eventmediafile "go-api/internal/application/event/mediafile"
	eventuser "go-api/internal/application/event/user"
	cmdmediafile "go-api/internal/application/command/mediafile"
	"go-api/internal/application/registry"
	domainmediafile "go-api/internal/domain/mediafile"
	domainuser "go-api/internal/domain/user"
	"go-api/internal/infrastructure/centrifugo"
	"go-api/internal/infrastructure/config"
	inframedia "go-api/internal/infrastructure/media"
	"go-api/internal/infrastructure/messaging/rabbitmq"
	"go-api/internal/infrastructure/notification"
	"go-api/internal/infrastructure/persistence/outbox"
	"go-api/internal/infrastructure/persistence/processed"
	"go-api/internal/infrastructure/persistence/write"
	"go-api/internal/infrastructure/storage"

	"gorm.io/gorm"
)

type Container struct {
	Relay                 *outbox.Relay
	Consumer              *rabbitmq.Consumer
	Conn                  *rabbitmq.Connection
	ExpireStaleUploads    *cmdmediafile.ExpireStaleUploadsHandler
	ExpireUploadsInterval time.Duration
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	topology := rabbitmq.DefaultTopology(
		env.RabbitMQExchange,
		env.RabbitMQQueue,
		env.RabbitMQRoutingKey,
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

	publisher := rabbitmq.NewPublisher(conn, env.RabbitMQExchange)
	outboxRepo := outbox.NewRepository(db)
	relay := outbox.NewRelay(outboxRepo, publisher, env.OutboxPollInterval, 50)

	mediaFileWriteRepo := write.NewMediaFileWriteRepository(db)
	dedupRepo := processed.NewRepository(db)
	notifier := notification.NewLogNotifier()
	realtimePublisher := centrifugo.NewPublisher(env)
	publishUserRealtime := eventuser.NewPublishRealtimeHandler(realtimePublisher)
	publishMediaRealtime := eventmediafile.NewPublishRealtimeHandler(realtimePublisher)

	generateThumbnailHandler := cmdmediafile.NewGenerateThumbnailHandler(
		mediaFileWriteRepo,
		minioStorage,
		inframedia.NewFrameExtractor(),
	)
	onUploadedThumbnail := eventmediafile.NewGenerateThumbnailOnUploadedHandler(generateThumbnailHandler)

	reg := registry.NewHandlerRegistry()

	reg.Register(domainuser.EventTypeUserCreated, dedup.With(
		dedupRepo,
		"user_created",
		eventuser.NewUserCreatedHandler().Handle,
	))
	reg.Register(domainuser.EventTypeUserCreated, dedup.With(
		dedupRepo,
		"notify_user_on_created",
		eventuser.NewNotifyUserOnCreatedHandler(notifier).Handle,
	))
	reg.Register(domainuser.EventTypeUserCreated, dedup.With(
		dedupRepo,
		"publish_user_created_realtime",
		publishUserRealtime.OnCreated,
	))
	reg.Register(domainuser.EventTypeUserUpdated, dedup.With(
		dedupRepo,
		"user_updated",
		eventuser.NewUserUpdatedHandler().Handle,
	))
	reg.Register(domainuser.EventTypeUserUpdated, dedup.With(
		dedupRepo,
		"publish_user_updated_realtime",
		publishUserRealtime.OnUpdated,
	))
	reg.Register(domainuser.EventTypeUserDeleted, dedup.With(
		dedupRepo,
		"user_deleted",
		eventuser.NewUserDeletedHandler().Handle,
	))
	reg.Register(domainuser.EventTypeUserDeleted, dedup.With(
		dedupRepo,
		"publish_user_deleted_realtime",
		publishUserRealtime.OnDeleted,
	))

	reg.Register(domainmediafile.EventTypeMediaFileUploaded, dedup.With(
		dedupRepo,
		"generate_media_file_thumbnail",
		onUploadedThumbnail.Handle,
	))
	reg.Register(domainmediafile.EventTypeMediaFileUploaded, dedup.With(
		dedupRepo,
		"publish_media_file_uploaded_realtime",
		publishMediaRealtime.OnUploaded,
	))

	consumer := rabbitmq.NewConsumer(conn, reg, env.WorkerConcurrency, env.WorkerMaxRetries)

	return &Container{
		Relay:    relay,
		Consumer: consumer,
		Conn:     conn,
		ExpireStaleUploads: cmdmediafile.NewExpireStaleUploadsHandler(
			mediaFileWriteRepo,
			outboxRepo,
			env.UploadURLTTL,
		),
		ExpireUploadsInterval: env.ExpireUploadsInterval,
	}
}
