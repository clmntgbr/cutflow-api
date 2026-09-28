package di

import (
	"log"

	authcmd "go-api/internal/application/command/auth"
	identitycmd "go-api/internal/application/command/identity"
	cmdmediafile "go-api/internal/application/command/mediafile"
	cmdproject "go-api/internal/application/command/project"
	usercmd "go-api/internal/application/command/user"
	queryuser "go-api/internal/application/query/user"
	"go-api/internal/infrastructure/centrifugo"
	infraClerk "go-api/internal/infrastructure/clerk"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/persistence/outbox"
	"go-api/internal/infrastructure/persistence/read"
	"go-api/internal/infrastructure/persistence/write"
	"go-api/internal/infrastructure/storage"
	httphandler "go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/middleware"

	"gorm.io/gorm"
)

type Container struct {
	AuthenticateMiddleware   *middleware.AuthenticateMiddleware
	UserWebhookMiddleware    *middleware.UserWebhookMiddleware
	StorageWebhookMiddleware *middleware.StorageWebhookMiddleware
	UserWebhookHandler       *httphandler.UserWebhookHandler
	StorageWebhookHandler    *httphandler.StorageWebhookHandler
	UserHandler              *httphandler.UserHandler
	ProjectHandler           *httphandler.ProjectHandler
	RealtimeHandler          *httphandler.RealtimeHandler
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	jwksProvider, err := infraClerk.NewJWKSProvider(env)
	if err != nil {
		log.Fatalf("failed to create JWKS provider: %v", err)
	}

	minioStorage, err := storage.NewMinIOStorage(env)
	if err != nil {
		log.Fatalf("failed to create storage client: %v", err)
	}

	userWriteRepo := write.NewUserWriteRepository(db)
	userReadRepo := read.NewUserReadRepository(db)
	projectWriteRepo := write.NewProjectWriteRepository(db)
	mediaFileWriteRepo := write.NewMediaFileWriteRepository(db)
	outboxRepo := outbox.NewRepository(db)

	createUserHandler := usercmd.NewCreateUserHandler(userWriteRepo, outboxRepo)
	updateUserHandler := usercmd.NewUpdateUserHandler(userWriteRepo, outboxRepo)
	getUserByExternalIDHandler := usercmd.NewGetUserByExternalIDHandler(userWriteRepo)
	deleteUserByExternalIDHandler := usercmd.NewDeleteUserByExternalIDHandler(userWriteRepo, outboxRepo)
	validateTokenHandler := authcmd.NewValidateTokenHandler(jwksProvider, userWriteRepo)
	fetchUserHandler := identitycmd.NewFetchUserHandler(infraClerk.NewUserGateway(env.ClerkSecretKey))
	getUserByIDHandler := queryuser.NewGetUserByIDHandler(userReadRepo)

	requestUploadURLHandler := cmdproject.NewRequestUploadURLHandler(
		projectWriteRepo,
		mediaFileWriteRepo,
		outboxRepo,
		minioStorage,
		env.UploadURLTTL,
		env.MediaMaxSizeBytes,
	)
	confirmUploadHandler := cmdmediafile.NewConfirmUploadHandler(
		mediaFileWriteRepo,
		outboxRepo,
		env.MediaMaxSizeBytes,
	)

	return &Container{
		AuthenticateMiddleware: middleware.NewAuthenticateMiddleware(
			validateTokenHandler,
			fetchUserHandler,
			createUserHandler,
		),
		UserWebhookMiddleware:    middleware.NewUserWebhookMiddleware(env.ClerkWebhookSecret),
		StorageWebhookMiddleware: middleware.NewStorageWebhookMiddleware(env.MinIOWebhookSecret),
		UserWebhookHandler: httphandler.NewUserWebhookHandler(
			getUserByExternalIDHandler,
			createUserHandler,
			updateUserHandler,
			deleteUserByExternalIDHandler,
		),
		StorageWebhookHandler: httphandler.NewStorageWebhookHandler(env.StorageBucket, confirmUploadHandler),
		UserHandler:           httphandler.NewUserHandler(getUserByIDHandler),
		ProjectHandler:        httphandler.NewProjectHandler(requestUploadURLHandler),
		RealtimeHandler:       httphandler.NewRealtimeHandler(centrifugo.NewConnectionInfoCreator(env)),
	}
}
