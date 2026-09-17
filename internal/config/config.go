package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	defaultRange    = "Funcionarios!A:C"
	defaultTimezone = "America/Sao_Paulo"
)

type Config struct {
	GoogleSheetID         string
	GoogleSheetRange      string
	GoogleCredentialsFile string
	Timezone              string
	Location              *time.Location
	LogFormat             string
	WhatsApp              WhatsAppConfig
}

type WhatsAppConfig struct {
	WebhookURL string
	Recipient  string
	SentBy     string
}

func Load() (Config, error) {
	sheetID, err := required("GOOGLE_SHEET_ID")
	if err != nil {
		return Config{}, err
	}

	credentialsFile, err := required("GOOGLE_APPLICATION_CREDENTIALS")
	if err != nil {
		return Config{}, err
	}
	if info, statErr := os.Stat(credentialsFile); statErr != nil {
		return Config{}, fmt.Errorf("acessar GOOGLE_APPLICATION_CREDENTIALS: %w", statErr)
	} else if info.IsDir() {
		return Config{}, fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS deve apontar para um arquivo")
	}

	timezone := valueOrDefault("TIMEZONE", defaultTimezone)
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return Config{}, fmt.Errorf("TIMEZONE invalida %q: %w", timezone, err)
	}

	logFormat := strings.ToLower(valueOrDefault("LOG_FORMAT", "human"))
	if logFormat != "human" && logFormat != "json" {
		return Config{}, fmt.Errorf("LOG_FORMAT deve ser human ou json")
	}

	whatsApp, err := loadWhatsApp()
	if err != nil {
		return Config{}, err
	}

	return Config{
		GoogleSheetID:         sheetID,
		GoogleSheetRange:      valueOrDefault("GOOGLE_SHEET_RANGE", defaultRange),
		GoogleCredentialsFile: credentialsFile,
		Timezone:              timezone,
		Location:              location,
		LogFormat:             logFormat,
		WhatsApp:              whatsApp,
	}, nil
}

func loadWhatsApp() (WhatsAppConfig, error) {
	webhookURL, err := required("WHATSAPP_WEBHOOK_URL")
	if err != nil {
		return WhatsAppConfig{}, err
	}
	if err := validateWebhookURL(webhookURL); err != nil {
		return WhatsAppConfig{}, err
	}
	recipient, err := required("WHATSAPP_RECIPIENT")
	if err != nil {
		return WhatsAppConfig{}, err
	}

	return WhatsAppConfig{
		WebhookURL: webhookURL,
		Recipient:  recipient,
		SentBy:     valueOrDefault("WHATSAPP_SENT_BY", "birth-reminder"),
	}, nil
}

func validateWebhookURL(value string) error {
	parsedURL, err := url.ParseRequestURI(value)
	if err != nil || parsedURL.Host == "" {
		return fmt.Errorf("WHATSAPP_WEBHOOK_URL deve ser uma URL HTTPS valida")
	}
	if parsedURL.Scheme == "https" {
		return nil
	}
	if parsedURL.Scheme == "http" && isLoopbackHost(parsedURL.Hostname()) {
		return nil
	}
	return fmt.Errorf("WHATSAPP_WEBHOOK_URL deve usar HTTPS; HTTP e permitido apenas para localhost")
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func required(name string) (string, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return "", fmt.Errorf("variavel obrigatoria ausente: %s", name)
	}
	return value, nil
}

func valueOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
