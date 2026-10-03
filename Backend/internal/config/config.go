package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName      string
	AppEnv       string
	AppPort      string
	AppPublicURL string
	APIPublicURL string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret    string
	JWTExpiresIn time.Duration

	MidtransServerKey string
	MidtransClientKey string
	MidtransBaseURL   string
	MidtransMock      bool

	PaymentCredentialEncryptionKey string

	XenditSecretKey     string
	XenditCallbackToken string
	XenditBaseURL       string
	XenditMock          bool
	XenditReturnBaseURL string
	CheckoutTTL         time.Duration
	AutopayEnabled      bool
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		AppName:      getEnv("APP_NAME", "PAY-LANGGAN.COM"),
		AppEnv:       getEnv("APP_ENV", "development"),
		AppPort:      getEnv("APP_PORT", "8080"),
		AppPublicURL: getEnv("APP_PUBLIC_URL", "http://localhost:5173"),
		APIPublicURL: getEnv("API_PUBLIC_URL", "http://localhost:8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "pay_langgan"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret:    getEnv("JWT_SECRET", "change_this_secret"),
		JWTExpiresIn: getDuration("JWT_EXPIRES_IN", 24*time.Hour),

		MidtransServerKey:              getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransClientKey:              getEnv("MIDTRANS_CLIENT_KEY", ""),
		MidtransBaseURL:                getEnv("MIDTRANS_BASE_URL", "https://api.sandbox.midtrans.com"),
		MidtransMock:                   getBool("MIDTRANS_MOCK", true),
		PaymentCredentialEncryptionKey: getEnv("PAYMENT_CREDENTIAL_ENCRYPTION_KEY", ""),

		XenditSecretKey:     getEnv("XENDIT_SECRET_KEY", ""),
		XenditCallbackToken: getEnv("XENDIT_CALLBACK_TOKEN", ""),
		XenditBaseURL:       getEnv("XENDIT_BASE_URL", "https://api.xendit.co"),
		XenditMock:          getBool("XENDIT_MOCK", true),
		XenditReturnBaseURL: getEnv("XENDIT_RETURN_BASE_URL", "https://paylanggan.example"),
		CheckoutTTL:         getDuration("CHECKOUT_TTL", 24*time.Hour),
		AutopayEnabled:      getBool("AUTOPAY_ENABLED", false),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return b
}

func getInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return i
}

func getDuration(key string, fallback time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return fallback
	}
	return d
}
