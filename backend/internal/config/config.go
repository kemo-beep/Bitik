package config

import (
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App           AppConfig           `mapstructure:"app"`
	HTTP          HTTPConfig          `mapstructure:"http"`
	CORS          CORSConfig          `mapstructure:"cors"`
	RateLimit     RateLimitConfig     `mapstructure:"rate_limit"`
	Security      SecurityConfig      `mapstructure:"security"`
	Auth          AuthConfig          `mapstructure:"auth"`
	Payments      PaymentsConfig      `mapstructure:"payments"`
	InternalAPI   InternalAPIConfig   `mapstructure:"internal_api"`
	Idempotency   IdempotencyConfig   `mapstructure:"idempotency"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Redis         RedisConfig         `mapstructure:"redis"`
	RabbitMQ      RabbitMQConfig      `mapstructure:"rabbitmq"`
	Storage       StorageConfig       `mapstructure:"storage"`
	Search        SearchConfig        `mapstructure:"search"`
	Observability ObservabilityConfig `mapstructure:"observability"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"`
	Environment string `mapstructure:"environment"`
	Version     string `mapstructure:"version"`
	Commit      string `mapstructure:"commit"`
}

type HTTPConfig struct {
	Addr            string        `mapstructure:"addr"`
	MetricsAddr     string        `mapstructure:"metrics_addr"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

type RateLimitConfig struct {
	RequestsPerSecond float64 `mapstructure:"requests_per_second"`
	Burst             int     `mapstructure:"burst"`
}

type SecurityConfig struct {
	MaxRequestBodyBytes int64 `mapstructure:"max_request_body_bytes"`
	MaxUploadBytes      int64 `mapstructure:"max_upload_bytes"`
}

type AuthConfig struct {
	JWTIssuer            string        `mapstructure:"jwt_issuer"`
	JWTSecret            string        `mapstructure:"jwt_secret"`
	AccessTokenTTL       time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL      time.Duration `mapstructure:"refresh_token_ttl"`
	PublicBaseURL        string        `mapstructure:"public_base_url"`
	RefreshCookieName    string        `mapstructure:"refresh_cookie_name"`
	RefreshCookieSecure  bool          `mapstructure:"refresh_cookie_secure"`
	OAuthRedirectBaseURL string        `mapstructure:"oauth_redirect_base_url"`

	OAuthGoogleClientID     string `mapstructure:"oauth_google_client_id"`
	OAuthGoogleClientSecret string `mapstructure:"oauth_google_client_secret"`

	OAuthFacebookAppID     string `mapstructure:"oauth_facebook_app_id"`
	OAuthFacebookAppSecret string `mapstructure:"oauth_facebook_app_secret"`

	OAuthAppleClientID      string `mapstructure:"oauth_apple_client_id"`
	OAuthAppleTeamID        string `mapstructure:"oauth_apple_team_id"`
	OAuthAppleKeyID         string `mapstructure:"oauth_apple_key_id"`
	OAuthApplePrivateKeyPEM string `mapstructure:"oauth_apple_private_key_pem"`

	OTPCodeTTL            time.Duration `mapstructure:"otp_code_ttl"`
	OTPMaxPerHour         int           `mapstructure:"otp_max_per_hour"`
	EmailVerifyTokenTTL   time.Duration `mapstructure:"email_verify_token_ttl"`
	PasswordResetTokenTTL time.Duration `mapstructure:"password_reset_token_ttl"`
	LogSecrets            bool          `mapstructure:"log_secrets"`
	MaxLoginFailures      int           `mapstructure:"max_login_failures"`
	LoginLockoutDuration  time.Duration `mapstructure:"login_lockout_duration"`
	OTPMaxVerifyFailures  int           `mapstructure:"otp_max_verify_failures"`
	RefreshCookieSameSite string        `mapstructure:"refresh_cookie_same_site"`
}

type PaymentsConfig struct {
	WebhookSecret string `mapstructure:"webhook_secret"`
}

type InternalAPIConfig struct {
	// Token is required to call internal worker/job endpoints.
	Token string `mapstructure:"token"`
	// AllowedCIDRs optionally restricts internal endpoints by client IP (e.g. ["10.0.0.0/8","127.0.0.1/32"]).
	AllowedCIDRs []string `mapstructure:"allowed_cidrs"`
}

type IdempotencyConfig struct {
	TTL          time.Duration `mapstructure:"ttl"`
	MaxBodyBytes int64         `mapstructure:"max_body_bytes"`
	LockTTL      time.Duration `mapstructure:"lock_ttl"`
	WaitTimeout  time.Duration `mapstructure:"wait_timeout"`
	PollInterval time.Duration `mapstructure:"poll_interval"`
}

type DatabaseConfig struct {
	URL             string        `mapstructure:"url"`
	MaxOpenConns    int32         `mapstructure:"max_open_conns"`
	MaxIdleConns    int32         `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnectTimeout  time.Duration `mapstructure:"connect_timeout"`
	QueryTimeout    time.Duration `mapstructure:"query_timeout"`
	// AutoMigrate runs Goose "up" on API startup before serving (disable in prod if you use a release job).
	AutoMigrate bool `mapstructure:"auto_migrate"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type RabbitMQConfig struct {
	URL string `mapstructure:"url"`
}

type StorageConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Bucket          string `mapstructure:"bucket"`
	UseSSL          bool   `mapstructure:"use_ssl"`
}

type SearchConfig struct {
	URL      string `mapstructure:"url"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	// ProductsIndex is the OpenSearch index name for products.
	ProductsIndex string `mapstructure:"products_index"`
}

type ObservabilityConfig struct {
	LogLevel       string  `mapstructure:"log_level"`
	TracingEnabled bool    `mapstructure:"tracing_enabled"`
	OTLPEndpoint   string  `mapstructure:"otlp_endpoint"`
	SampleRatio    float64 `mapstructure:"sample_ratio"`
}

func Load() (Config, error) {
	loadDotEnv()

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.SetEnvPrefix("BITIK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return Config{}, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func loadDotEnv() {
	original := map[string]struct{}{}
	for _, entry := range os.Environ() {
		key, _, ok := strings.Cut(entry, "=")
		if ok {
			original[key] = struct{}{}
		}
	}

	// Support both `go run` from `backend/` and from the monorepo root (`Bitik/`).
	if fileExists("backend/go.mod") {
		loadDotEnvFiles(original, "backend/.env", "backend/internal/.env")
		return
	}
	loadDotEnvFiles(original, ".env", "internal/.env")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func loadDotEnvFiles(original map[string]struct{}, paths ...string) {
	values := map[string]string{}
	for _, p := range paths {
		if !fileExists(p) {
			continue
		}
		fileValues, err := godotenv.Read(p)
		if err != nil {
			continue
		}
		for key, value := range fileValues {
			values[key] = value
		}
	}

	for key, value := range values {
		if _, exists := original[key]; exists {
			continue
		}
		_ = os.Setenv(key, value)
	}
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "bitik-backend")
	v.SetDefault("app.environment", "local")
	v.SetDefault("app.version", "dev")
	v.SetDefault("app.commit", "local")

	v.SetDefault("http.addr", ":8080")
	v.SetDefault("http.metrics_addr", "")
	v.SetDefault("http.read_timeout", "10s")
	v.SetDefault("http.write_timeout", "15s")
	v.SetDefault("http.shutdown_timeout", "10s")

	v.SetDefault("cors.allowed_origins", []string{"http://localhost:3000", "http://localhost:5173"})
	v.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allowed_headers", []string{"Authorization", "Content-Type", "Idempotency-Key", "X-Request-ID", "X-Device-Id", "X-Platform", "X-App-Version"})

	v.SetDefault("internal_api.token", "")
	v.SetDefault("internal_api.allowed_cidrs", []string{})

	v.SetDefault("rate_limit.requests_per_second", 20)
	v.SetDefault("rate_limit.burst", 40)
	v.SetDefault("security.max_request_body_bytes", 1<<20)
	v.SetDefault("security.max_upload_bytes", 10<<20)

	v.SetDefault("auth.jwt_issuer", "bitik")
	v.SetDefault("auth.jwt_secret", "change-me-in-local-env")
	v.SetDefault("auth.access_token_ttl", "15m")
	v.SetDefault("auth.refresh_token_ttl", "720h")
	v.SetDefault("auth.public_base_url", "http://127.0.0.1:8080")
	v.SetDefault("auth.refresh_cookie_name", "bitik_refresh")
	v.SetDefault("auth.refresh_cookie_secure", false)
	v.SetDefault("auth.oauth_redirect_base_url", "http://localhost:3000")
	v.SetDefault("auth.otp_code_ttl", "10m")
	v.SetDefault("auth.otp_max_per_hour", 5)
	v.SetDefault("auth.email_verify_token_ttl", "48h")
	v.SetDefault("auth.password_reset_token_ttl", "1h")
	v.SetDefault("auth.log_secrets", false)
	v.SetDefault("auth.max_login_failures", 8)
	v.SetDefault("auth.login_lockout_duration", "20m")
	v.SetDefault("auth.otp_max_verify_failures", 5)
	v.SetDefault("auth.refresh_cookie_same_site", "lax")
	v.SetDefault("payments.webhook_secret", "")

	v.SetDefault("idempotency.ttl", "24h")
	v.SetDefault("idempotency.max_body_bytes", 262144)
	v.SetDefault("idempotency.lock_ttl", "2m")
	v.SetDefault("idempotency.wait_timeout", "30s")
	v.SetDefault("idempotency.poll_interval", "50ms")

	v.SetDefault("database.url", "postgres://bitik:bitik@127.0.0.1:55432/bitik?sslmode=disable")
	v.SetDefault("database.auto_migrate", true)
	v.SetDefault("database.max_open_conns", 20)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", "30m")
	v.SetDefault("database.connect_timeout", "5s")
	v.SetDefault("database.query_timeout", "10s")

	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("rabbitmq.url", "amqp://bitik:bitik@localhost:5672/")

	v.SetDefault("storage.endpoint", "localhost:9000")
	v.SetDefault("storage.access_key_id", "minioadmin")
	v.SetDefault("storage.secret_access_key", "minioadmin")
	v.SetDefault("storage.bucket", "bitik-local")
	v.SetDefault("storage.use_ssl", false)

	v.SetDefault("search.url", "http://localhost:9200")
	v.SetDefault("search.username", "")
	v.SetDefault("search.password", "")
	v.SetDefault("search.products_index", "products_v1")

	v.SetDefault("observability.log_level", "info")
	v.SetDefault("observability.tracing_enabled", false)
	v.SetDefault("observability.otlp_endpoint", "localhost:4317")
	v.SetDefault("observability.sample_ratio", 1.0)
}
