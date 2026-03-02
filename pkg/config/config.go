package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	DB       DBConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Services ServicesConfig
	Stripe   StripeConfig
	Maps     MapsConfig
	Matching MatchingConfig
	Pricing  PricingConfig
}

type AppConfig struct {
	Env         string `mapstructure:"APP_ENV"`
	Debug       bool   `mapstructure:"APP_DEBUG"`
	ServiceName string `mapstructure:"SERVICE_NAME"`
	ServicePort string `mapstructure:"SERVICE_PORT"`
}

type DBConfig struct {
	Host           string `mapstructure:"DB_HOST"`
	Port           string `mapstructure:"DB_PORT"`
	User           string `mapstructure:"DB_USER"`
	Password       string `mapstructure:"DB_PASSWORD"`
	Name           string `mapstructure:"DB_NAME"`
	SSLMode        string `mapstructure:"DB_SSL_MODE"`
	MaxConnections int    `mapstructure:"DB_MAX_CONNECTIONS"`
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}

type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     string `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

type JWTConfig struct {
	Secret        string        `mapstructure:"JWT_SECRET"`
	AccessExpiry  time.Duration `mapstructure:"JWT_ACCESS_EXPIRY"`
	RefreshExpiry time.Duration `mapstructure:"JWT_REFRESH_EXPIRY"`
}

type ServicesConfig struct {
	UserServiceURL        string `mapstructure:"USER_SERVICE_URL"`
	RideServiceURL        string `mapstructure:"RIDE_SERVICE_URL"`
	MatchingServiceURL    string `mapstructure:"MATCHING_SERVICE_URL"`
	GeolocationServiceURL string `mapstructure:"GEOLOCATION_SERVICE_URL"`
	PricingServiceURL     string `mapstructure:"PRICING_SERVICE_URL"`
	PaymentServiceURL     string `mapstructure:"PAYMENT_SERVICE_URL"`
}

type StripeConfig struct {
	SecretKey     string `mapstructure:"STRIPE_SECRET_KEY"`
	WebhookSecret string `mapstructure:"STRIPE_WEBHOOK_SECRET"`
}

type MapsConfig struct {
	GoogleAPIKey string `mapstructure:"GOOGLE_MAPS_API_KEY"`
}

type MatchingConfig struct {
	CorridorMeters        int `mapstructure:"MATCHING_CORRIDOR_METERS"`
	MaxDetourPercent      int `mapstructure:"MATCHING_MAX_DETOUR_PERCENT"`
	MinBatchSize          int `mapstructure:"MATCHING_MIN_BATCH_SIZE"`
	MaxBatchSize          int `mapstructure:"MATCHING_MAX_BATCH_SIZE"`
	RouteExtensionPercent int `mapstructure:"MATCHING_ROUTE_EXTENSION_PERCENT"`
}

type PricingConfig struct {
	BaseFareRoundTrip        float64 `mapstructure:"BASE_FARE_ROUND_TRIP"`
	BaseFareOneWay           float64 `mapstructure:"BASE_FARE_ONE_WAY"`
	PlatformCommissionPct    int     `mapstructure:"PLATFORM_COMMISSION_PERCENT"`
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("SERVICE_PORT", "8080")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_SSL_MODE", "disable")
	viper.SetDefault("DB_MAX_CONNECTIONS", 25)
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("JWT_ACCESS_EXPIRY", "15m")
	viper.SetDefault("JWT_REFRESH_EXPIRY", "720h")
	viper.SetDefault("MATCHING_CORRIDOR_METERS", 1000)
	viper.SetDefault("MATCHING_MAX_DETOUR_PERCENT", 20)
	viper.SetDefault("MATCHING_MIN_BATCH_SIZE", 1)
	viper.SetDefault("MATCHING_MAX_BATCH_SIZE", 4)
	viper.SetDefault("MATCHING_ROUTE_EXTENSION_PERCENT", 15)
	viper.SetDefault("BASE_FARE_ROUND_TRIP", 20.00)
	viper.SetDefault("BASE_FARE_ONE_WAY", 10.00)
	viper.SetDefault("PLATFORM_COMMISSION_PERCENT", 8)

	if err := viper.ReadInConfig(); err != nil {
		// .env file is optional; env vars are fine
	}

	cfg := &Config{}
	cfg.App.Env = viper.GetString("APP_ENV")
	cfg.App.Debug = viper.GetBool("APP_DEBUG")
	cfg.App.ServiceName = viper.GetString("SERVICE_NAME")
	cfg.App.ServicePort = viper.GetString("SERVICE_PORT")

	cfg.DB.Host = viper.GetString("DB_HOST")
	cfg.DB.Port = viper.GetString("DB_PORT")
	cfg.DB.User = viper.GetString("DB_USER")
	cfg.DB.Password = viper.GetString("DB_PASSWORD")
	cfg.DB.Name = viper.GetString("DB_NAME")
	cfg.DB.SSLMode = viper.GetString("DB_SSL_MODE")
	cfg.DB.MaxConnections = viper.GetInt("DB_MAX_CONNECTIONS")

	cfg.Redis.Host = viper.GetString("REDIS_HOST")
	cfg.Redis.Port = viper.GetString("REDIS_PORT")
	cfg.Redis.Password = viper.GetString("REDIS_PASSWORD")
	cfg.Redis.DB = viper.GetInt("REDIS_DB")

	cfg.JWT.Secret = viper.GetString("JWT_SECRET")
	cfg.JWT.AccessExpiry = viper.GetDuration("JWT_ACCESS_EXPIRY")
	cfg.JWT.RefreshExpiry = viper.GetDuration("JWT_REFRESH_EXPIRY")

	cfg.Services.UserServiceURL = viper.GetString("USER_SERVICE_URL")
	cfg.Services.MatchingServiceURL = viper.GetString("MATCHING_SERVICE_URL")
	cfg.Services.GeolocationServiceURL = viper.GetString("GEOLOCATION_SERVICE_URL")
	cfg.Services.PricingServiceURL = viper.GetString("PRICING_SERVICE_URL")

	cfg.Stripe.SecretKey = viper.GetString("STRIPE_SECRET_KEY")
	cfg.Stripe.WebhookSecret = viper.GetString("STRIPE_WEBHOOK_SECRET")

	cfg.Maps.GoogleAPIKey = viper.GetString("GOOGLE_MAPS_API_KEY")

	cfg.Matching.CorridorMeters = viper.GetInt("MATCHING_CORRIDOR_METERS")
	cfg.Matching.MaxDetourPercent = viper.GetInt("MATCHING_MAX_DETOUR_PERCENT")
	cfg.Matching.MinBatchSize = viper.GetInt("MATCHING_MIN_BATCH_SIZE")
	cfg.Matching.MaxBatchSize = viper.GetInt("MATCHING_MAX_BATCH_SIZE")
	cfg.Matching.RouteExtensionPercent = viper.GetInt("MATCHING_ROUTE_EXTENSION_PERCENT")

	cfg.Pricing.BaseFareRoundTrip = viper.GetFloat64("BASE_FARE_ROUND_TRIP")
	cfg.Pricing.BaseFareOneWay = viper.GetFloat64("BASE_FARE_ONE_WAY")
	cfg.Pricing.PlatformCommissionPct = viper.GetInt("PLATFORM_COMMISSION_PERCENT")

	return cfg, nil
}
