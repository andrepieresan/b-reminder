package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/andrepieresan/birth-reminder/internal/birthday"
	"github.com/andrepieresan/birth-reminder/internal/config"
)

type Client struct {
	config config.WhatsAppConfig
	http   *http.Client
}

func New(cfg config.WhatsAppConfig) *Client {
	return &Client{
		config: cfg,
		http: &http.Client{
			Timeout: 15 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func TodayMessage(result birthday.Result) string {
	names := make([]string, 0, len(result.People))
	for _, person := range result.People {
		names = append(names, "🎂 *"+person.Name+"*")
	}

	date := result.Date
	if parsed, err := time.Parse("2006-01-02", result.Date); err == nil {
		date = parsed.Format("02/01")
	}

	return fmt.Sprintf(
		"✨ Hoje o nosso time tem um motivo especial para celebrar!\n\nHoje comemoramos o aniversário de:\n%s\n\nEntre projetos, ideias e desafios, são as pessoas que fazem tudo acontecer. Que este novo ciclo venha cheio de boas conquistas, aprendizados e momentos felizes! 🚀\n\nPessoal, vamos deixar uma mensagem de carinho e tornar este dia ainda mais especial? 💙\n\nE já fica o aviso: queremos churrasco, hein?! 🔥🥩\n\n📅 %s",
		strings.Join(names, "\n"),
		date,
	)
}

func TomorrowMessage(result birthday.Result) string {
	names := make([]string, 0, len(result.People))
	for _, person := range result.People {
		names = append(names, "🎂 *"+person.Name+"*")
	}

	date := result.Date
	if parsed, err := time.Parse("2006-01-02", result.Date); err == nil {
		date = parsed.Format("02/01")
	}

	return fmt.Sprintf(
		"🔔 Uma celebração especial está chegando!\nAmanhã é aniversário de:\n%s\n\n📅 Amanhã, %s",
		strings.Join(names, "\n"),
		date,
	)
}

func (client *Client) Send(ctx context.Context, message string) error {
	payload, err := json.Marshal(struct {
		Phone       string `json:"phone"`
		Message     string `json:"message"`
		SentBy      string `json:"sent_by"`
		LinkPreview bool   `json:"linkPreview"`
	}{
		Phone:       client.config.Recipient,
		Message:     message,
		SentBy:      client.config.SentBy,
		LinkPreview: true,
	})
	if err != nil {
		return fmt.Errorf("montar mensagem: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.config.WebhookURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("criar requisicao: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "birth-reminder/1.0")

	response, err := client.http.Do(request)
	if err != nil {
		return fmt.Errorf("enviar requisicao: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 8*1024))
	if err != nil {
		return fmt.Errorf("ler resposta do webhook: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("webhook respondeu com status %d", response.StatusCode)
	}

	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return fmt.Errorf("webhook retornou uma resposta invalida")
	}
	if !result.Success {
		if result.Message == "" {
			result.Message = "envio recusado sem motivo informado"
		}
		return fmt.Errorf("webhook recusou o envio: %s", result.Message)
	}

	return nil
}
