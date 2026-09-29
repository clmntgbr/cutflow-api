package di

import (
	"log"
	"os"
	"strings"

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
	"go-api/internal/infrastructure/viralllm"

	"gorm.io/gorm"
)

type Container struct {
	Consumer *rabbitmq.Consumer
	Conn     *rabbitmq.Connection
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	topology := rabbitmq.DefaultTopology(
		env.RabbitMQExchange,
		env.ViralQueue,
		env.ViralRoutingKey,
		env.RabbitMQRetryTTLMS,
	)

	conn, err := rabbitmq.Connect(env.RabbitMQURL, topology)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}

	apiKey := env.ViralLLMAPIKey
	if apiKey == "" {
		switch strings.ToLower(env.ViralLLMProvider) {
		case "deepseek":
			apiKey = os.Getenv("DEEPSEEK_API_KEY")
		default:
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
	}

	analyzer := viralllm.New(
		env.ViralLLMProvider,
		env.ViralLLMModel,
		apiKey,
		env.ViralLLMBaseURL,
		env.ViralLLMDumpJSON,
	)
	log.Printf("viral llm provider=%s model=%s dump=%s", analyzer.Provider(), analyzer.Model(), env.ViralLLMDumpJSON)

	outboxRepo := outbox.NewRepository(db)
	dedupRepo := processed.NewRepository(db)
	analyzeHandler := cmdmediafile.NewAnalyzeViralHandler(
		write.NewTranscriptWriteRepository(db),
		write.NewMediaFileWriteRepository(db),
		write.NewMediaConfigurationWriteRepository(db),
		write.NewViralCandidateWriteRepository(db),
		write.NewJobWriteRepository(db),
		analyzer,
		outboxRepo,
		env.ViralChunkDurationMs,
		env.ViralChunkOverlapMs,
	)

	reg := registry.NewHandlerRegistry()
	reg.Register(domainmediafile.EventTypeMediaFileViralRequested, dedup.With(
		dedupRepo,
		"analyze_viral_on_requested",
		eventmediafile.NewAnalyzeViralOnRequestedHandler(analyzeHandler).Handle,
	))

	consumer := rabbitmq.NewConsumer(conn, reg, env.ViralConcurrency, env.WorkerMaxRetries)
	return &Container{Consumer: consumer, Conn: conn}
}
