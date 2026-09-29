package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL             string
	ClerkWebhookSecret      string
	Port                    string
	Environment             string
	ClerkSecretKey          string
	ClerkFrontendAPI        string
	CORSAllowedOrigins      []string
	CORSAllowCredentials    bool
	CORSAllowMethods        []string
	CORSAllowHeaders        []string
	CORSMaxAge              int
	RateLimitMax            int
	RabbitMQURL             string
	RabbitMQExchange        string
	RabbitMQQueue           string
	RabbitMQRoutingKey      string
	RabbitMQRetryTTLMS      int
	WorkerMaxRetries        int
	OutboxPollInterval      time.Duration
	WorkerConcurrency       int
	CentrifugoURL           string
	CentrifugoAPIKey        string
	CentrifugoTokenSecret   string
	CentrifugoPublicWSURL   string
	StorageEndpoint         string
	StorageInternalEndpoint string
	StorageRegion           string
	StorageAccessKey        string
	StorageSecretKey        string
	StorageBucket           string
	StorageUsePathStyle     bool
	MinIOWebhookSecret      string
	UploadURLTTL            time.Duration
	ExpireUploadsInterval   time.Duration
	MediaMaxSizeBytes       int64
	ExtractionQueue         string
	ExtractionRoutingKey    string
	ExtractionConcurrency   int
	SilenceQueue            string
	SilenceRoutingKey       string
	SilenceConcurrency      int
	SilenceThresholdDB      float64
	SilenceMinDurationMs    int64
	TranscriptQueue         string
	TranscriptRoutingKey    string
	TranscriptConcurrency   int
	AssemblyAIAPIKey        string
	TranscriptUseFixtures   bool
	TranscriptFixtureJSON   string
	AnalysisQueue           string
	AnalysisRoutingKey      string
	AnalysisConcurrency     int
	ViralQueue              string
	ViralRoutingKey         string
	ViralConcurrency        int
	ViralLLMProvider        string
	ViralLLMModel           string
	ViralLLMAPIKey          string
	ViralLLMBaseURL         string
	ViralChunkDurationMs    int64
	ViralChunkOverlapMs     int64
	TimelineQueue           string
	TimelineRoutingKey      string
	TimelineConcurrency     int
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	return &Config{
		DatabaseURL:             getEnv("DATABASE_URL"),
		ClerkWebhookSecret:      getEnv("CLERK_WEBHOOK_SECRET"),
		Port:                    getEnv("PORT"),
		Environment:             getEnv("GO_ENV"),
		ClerkSecretKey:          getEnv("CLERK_SECRET_KEY"),
		ClerkFrontendAPI:        getEnv("CLERK_FRONTEND_API"),
		CORSAllowedOrigins:      strings.Split(getEnv("CORS_ALLOWED_ORIGINS"), ","),
		CORSAllowCredentials:    getEnvBool("CORS_ALLOW_CREDENTIALS"),
		CORSAllowMethods:        strings.Split(getEnv("CORS_ALLOW_METHODS"), ","),
		CORSAllowHeaders:        strings.Split(getEnv("CORS_ALLOW_HEADERS"), ","),
		CORSMaxAge:              getEnvInt("CORS_MAX_AGE"),
		RateLimitMax:            getEnvInt("RATE_LIMIT_MAX"),
		RabbitMQURL:             getEnv("RABBITMQ_URL"),
		RabbitMQExchange:        getEnvOrDefault("RABBITMQ_EXCHANGE", "domain.events"),
		RabbitMQQueue:           getEnvOrDefault("RABBITMQ_QUEUE", "domain.events"),
		RabbitMQRoutingKey:      getEnvOrDefault("RABBITMQ_ROUTING_KEY", "user.#,project.#,media_file.#,job.#"),
		RabbitMQRetryTTLMS:      getEnvIntOrDefault("RABBITMQ_RETRY_TTL_MS", 30000),
		WorkerMaxRetries:        getEnvIntOrDefault("WORKER_MAX_RETRIES", 3),
		OutboxPollInterval:      getEnvDuration("OUTBOX_POLL_INTERVAL", 2*time.Second),
		WorkerConcurrency:       getEnvIntOrDefault("WORKER_CONCURRENCY", 4),
		CentrifugoURL:           getEnv("CENTRIFUGO_URL"),
		CentrifugoAPIKey:        getEnv("CENTRIFUGO_API_KEY"),
		CentrifugoTokenSecret:   getEnv("CENTRIFUGO_TOKEN_SECRET"),
		CentrifugoPublicWSURL:   getEnvOrDefault("CENTRIFUGO_PUBLIC_WS_URL", ""),
		StorageEndpoint:         getEnv("STORAGE_ENDPOINT"),
		StorageInternalEndpoint: getEnvOrDefault("STORAGE_INTERNAL_ENDPOINT", ""),
		StorageRegion:           getEnvOrDefault("STORAGE_REGION", "us-east-1"),
		StorageAccessKey:        getEnv("STORAGE_ACCESS_KEY"),
		StorageSecretKey:        getEnv("STORAGE_SECRET_KEY"),
		StorageBucket:           getEnv("STORAGE_BUCKET"),
		StorageUsePathStyle:     getEnvBoolOrDefault("STORAGE_USE_PATH_STYLE", true),
		MinIOWebhookSecret:      getEnvOrDefault("MINIO_WEBHOOK_SECRET", "dev-minio-webhook-secret"),
		UploadURLTTL:            getEnvDuration("UPLOAD_URL_TTL", 15*time.Minute),
		ExpireUploadsInterval:   getEnvDuration("EXPIRE_UPLOADS_INTERVAL", time.Minute),
		MediaMaxSizeBytes:       getEnvInt64OrDefault("MEDIA_MAX_SIZE_BYTES", 2*1024*1024*1024),
		ExtractionQueue:         getEnvOrDefault("EXTRACTION_QUEUE", "extraction"),
		ExtractionRoutingKey:    getEnvOrDefault("EXTRACTION_ROUTING_KEY", "media_file.ready.v1"),
		ExtractionConcurrency:   getEnvIntOrDefault("EXTRACTION_CONCURRENCY", 5),
		SilenceQueue:            getEnvOrDefault("SILENCE_QUEUE", "silence"),
		SilenceRoutingKey:       getEnvOrDefault("SILENCE_ROUTING_KEY", "media_file.silence_requested.v1"),
		SilenceConcurrency:      getEnvIntOrDefault("SILENCE_CONCURRENCY", 5),
		SilenceThresholdDB:      getEnvFloat64OrDefault("SILENCE_THRESHOLD_DB", -35),
		SilenceMinDurationMs:    getEnvInt64OrDefault("SILENCE_MIN_DURATION_MS", 400),
		TranscriptQueue:         getEnvOrDefault("TRANSCRIPT_QUEUE", "transcript"),
		TranscriptRoutingKey:    getEnvOrDefault("TRANSCRIPT_ROUTING_KEY", "media_file.transcript_requested.v1"),
		TranscriptConcurrency:   getEnvIntOrDefault("TRANSCRIPT_CONCURRENCY", 5),
		AssemblyAIAPIKey:        getEnvOrDefault("ASSEMBLYAI_API_KEY", ""),
		TranscriptUseFixtures:   getEnvBoolOrDefault("TRANSCRIPT_USE_FIXTURES", false),
		TranscriptFixtureJSON:   getEnvOrDefault("TRANSCRIPT_FIXTURE_JSON", "fixtures/transcript.json"),
		AnalysisQueue:           getEnvOrDefault("ANALYSIS_QUEUE", "analysis"),
		AnalysisRoutingKey:      getEnvOrDefault("ANALYSIS_ROUTING_KEY", "media_file.analysis_requested.v1"),
		AnalysisConcurrency:     getEnvIntOrDefault("ANALYSIS_CONCURRENCY", 5),
		ViralQueue:              getEnvOrDefault("VIRAL_QUEUE", "viral"),
		ViralRoutingKey:         getEnvOrDefault("VIRAL_ROUTING_KEY", "media_file.viral_requested.v1"),
		ViralConcurrency:        getEnvIntOrDefault("VIRAL_CONCURRENCY", 2),
		ViralLLMProvider:        getEnvOrDefault("VIRAL_LLM_PROVIDER", "openai"),
		ViralLLMModel:           getEnvOrDefault("VIRAL_LLM_MODEL", "gpt-4o-mini"),
		ViralLLMAPIKey:          getEnvOrDefault("VIRAL_LLM_API_KEY", ""),
		ViralLLMBaseURL:         getEnvOrDefault("VIRAL_LLM_BASE_URL", ""),
		ViralChunkDurationMs:    getEnvInt64OrDefault("VIRAL_CHUNK_DURATION_MS", 600000),
		ViralChunkOverlapMs:     getEnvInt64OrDefault("VIRAL_CHUNK_OVERLAP_MS", 30000),
		TimelineQueue:           getEnvOrDefault("TIMELINE_QUEUE", "timeline"),
		TimelineRoutingKey:      getEnvOrDefault("TIMELINE_ROUTING_KEY", "media_file.timeline_rebuild_requested.v1"),
		TimelineConcurrency:     getEnvIntOrDefault("TIMELINE_CONCURRENCY", 5),
	}
}

func getEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	log.Panicf("required environment variable %s is not set", key)
	return ""
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string) bool {
	value := os.Getenv(key)
	if value == "" {
		return false
	}

	return value == "true"
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true"
}

func getEnvInt(key string) int {
	value := os.Getenv(key)
	if value == "" {
		log.Panicf("required environment variable %s is not set", key)
		return 0
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		log.Panicf("invalid integer for %s: %q", key, value)
		return 0
	}

	return parsedValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		log.Panicf("invalid integer for %s: %q", key, value)
		return 0
	}
	return parsedValue
}

func getEnvInt64OrDefault(key string, defaultValue int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsedValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Panicf("invalid int64 for %s: %q", key, value)
		return 0
	}
	return parsedValue
}

func getEnvFloat64OrDefault(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsedValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Panicf("invalid float64 for %s: %q", key, value)
		return 0
	}
	return parsedValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		log.Panicf("invalid duration for %s: %q", key, value)
		return 0
	}
	return parsed
}
