package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/andrepieresan/birth-reminder/internal/birthday"
	"github.com/andrepieresan/birth-reminder/internal/config"
	"github.com/andrepieresan/birth-reminder/internal/googlesheets"
	"github.com/andrepieresan/birth-reminder/internal/logging"
	"github.com/andrepieresan/birth-reminder/internal/whatsapp"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := run(ctx); err != nil {
		logging.New(os.Getenv("LOG_FORMAT")).Error("[SYNC] Falha ao verificar aniversarios", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := logging.New(cfg.LogFormat)
	logger.Info("[SYNC] Iniciando verificacao de aniversarios",
		"range", cfg.GoogleSheetRange,
		"timezone", cfg.Timezone,
	)

	rows, err := googlesheets.ReadRows(ctx, cfg)
	if err != nil {
		return fmt.Errorf("ler planilha do Google: %w", err)
	}

	now := time.Now().In(cfg.Location)
	today, err := birthday.FindOnDate(rows, now, func(row int, reason string) {
		logger.Warn("[SYNC] Linha ignorada", "row", row, "reason", reason)
	})
	if err != nil {
		return err
	}
	tomorrow, err := birthday.FindOnDate(rows, now.AddDate(0, 0, 1), nil)
	if err != nil {
		return err
	}

	if len(today.People) == 0 && len(tomorrow.People) == 0 {
		logger.Info("[SYNC] Nenhum aniversario hoje ou amanha", "date", today.Date)
		return nil
	}

	alerts := make([]struct {
		when    string
		result  birthday.Result
		message string
	}, 0, 2)
	if len(today.People) > 0 {
		alerts = append(alerts, struct {
			when    string
			result  birthday.Result
			message string
		}{"hoje", today, whatsapp.TodayMessage(today)})
	}
	if len(tomorrow.People) > 0 {
		alerts = append(alerts, struct {
			when    string
			result  birthday.Result
			message string
		}{"amanha", tomorrow, whatsapp.TomorrowMessage(tomorrow)})
	}

	client := whatsapp.New(cfg.WhatsApp)
	for _, alert := range alerts {
		names := make([]string, 0, len(alert.result.People))
		for _, person := range alert.result.People {
			names = append(names, person.Name)
		}

		logger.LogAttrs(ctx, slog.LevelInfo, "[SYNC] Aniversario encontrado "+alert.when,
			slog.String("date", alert.result.Date),
			slog.Int("count", len(alert.result.People)),
			slog.Any("names", names),
		)

		if err := client.Send(ctx, alert.message); err != nil {
			return fmt.Errorf("enviar alerta de aniversario %s pelo webhook do WhatsApp: %w", alert.when, err)
		}
		logger.Info("[SYNC] Alerta enviado pelo WhatsApp", "when", alert.when, "count", len(alert.result.People))
	}

	return nil
}
