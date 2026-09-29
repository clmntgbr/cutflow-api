package di

import (
	"log"
	"time"

	"go-api/internal/application/event/dedup"
	eventmediafile "go-api/internal/application/event/mediafile"
	eventproject "go-api/internal/application/event/project"
	eventuser "go-api/internal/application/event/user"
	cmdmediafile "go-api/internal/application/command/mediafile"
	"go-api/internal/application/registry"
	domainmediaaudio "go-api/internal/domain/mediaaudio"
	domainmediafile "go-api/internal/domain/mediafile"
	domainproject "go-api/internal/domain/project"
	domainsegment "go-api/internal/domain/segment"
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
	segmentWriteRepo := write.NewSegmentWriteRepository(db)
	dedupRepo := processed.NewRepository(db)
	notifier := notification.NewLogNotifier()
	realtimePublisher := centrifugo.NewPublisher(env)
	publishUserRealtime := eventuser.NewPublishRealtimeHandler(realtimePublisher)
	publishMediaRealtime := eventmediafile.NewPublishRealtimeHandler(realtimePublisher)
	publishProjectRealtime := eventproject.NewPublishRealtimeHandler(realtimePublisher)

	generateThumbnailHandler := cmdmediafile.NewGenerateThumbnailHandler(
		mediaFileWriteRepo,
		minioStorage,
		inframedia.NewFrameExtractor(),
		outboxRepo,
	)
	onUploadedThumbnail := eventmediafile.NewGenerateThumbnailOnUploadedHandler(generateThumbnailHandler)

	probeMediaHandler := cmdmediafile.NewProbeMediaHandler(
		mediaFileWriteRepo,
		minioStorage,
		inframedia.NewMediaProber(),
		outboxRepo,
	)
	onUploadedProbe := eventmediafile.NewProbeMediaOnUploadedHandler(probeMediaHandler)

	createSegmentsHandler := cmdmediafile.NewCreateSegmentsHandler(
		mediaFileWriteRepo,
		segmentWriteRepo,
		outboxRepo,
		domainsegment.PlanConfig{
			DurationMs: env.SegmentDurationMs,
			OverlapMs:  env.SegmentOverlapMs,
			MaxCount:   env.SegmentMaxCount,
		},
	)
	onReadySegments := eventmediafile.NewCreateSegmentsOnReadyHandler(createSegmentsHandler)

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
		"probe_media_file_on_uploaded",
		onUploadedProbe.Handle,
	))
	reg.Register(domainmediafile.EventTypeMediaFileUploaded, dedup.With(
		dedupRepo,
		"publish_media_file_uploaded_realtime",
		publishMediaRealtime.OnUploaded,
	))
	reg.Register(domainmediafile.EventTypeMediaFileThumbnailReady, dedup.With(
		dedupRepo,
		"publish_media_file_thumbnail_ready_realtime",
		publishMediaRealtime.OnThumbnailReady,
	))
	reg.Register(domainmediafile.EventTypeMediaFileProbing, dedup.With(
		dedupRepo,
		"publish_media_file_probing_realtime",
		publishMediaRealtime.OnProbing,
	))
	reg.Register(domainmediafile.EventTypeMediaFileReady, dedup.With(
		dedupRepo,
		"publish_media_file_ready_realtime",
		publishMediaRealtime.OnReady,
	))
	reg.Register(domainmediafile.EventTypeMediaFileReady, dedup.With(
		dedupRepo,
		"create_segments_on_media_file_ready",
		onReadySegments.Handle,
	))
	reg.Register(domainmediafile.EventTypeMediaFileProbeFailed, dedup.With(
		dedupRepo,
		"publish_media_file_probe_failed_realtime",
		publishMediaRealtime.OnProbeFailed,
	))
	reg.Register(domainmediafile.EventTypeMediaFileSegmentsReady, dedup.With(
		dedupRepo,
		"publish_media_file_segments_ready_realtime",
		publishMediaRealtime.OnSegmentsReady,
	))
	reg.Register(domainmediaaudio.EventTypeMediaAudioReady, dedup.With(
		dedupRepo,
		"publish_media_file_audio_ready_realtime",
		publishMediaRealtime.OnAudioReady,
	))
	reg.Register(domainmediaaudio.EventTypeMediaAudioFailed, dedup.With(
		dedupRepo,
		"publish_media_file_audio_failed_realtime",
		publishMediaRealtime.OnAudioFailed,
	))
	reg.Register(domainproject.EventTypeProjectUpdated, dedup.With(
		dedupRepo,
		"publish_project_updated_realtime",
		publishProjectRealtime.OnUpdated,
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
