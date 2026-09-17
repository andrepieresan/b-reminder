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
		"mode", cfg.ReminderMode,
	)

	rows, err := googlesheets.ReadRows(ctx, cfg)
	if err != nil {
		return fmt.Errorf("ler planilha do Google: %w", err)
	}

	warn := func(row int, reason string) {
		logger.Warn("[SYNC] Linha ignorada", "row", row, "reason", reason)
	}

	type alert struct {
		when    string
		result  birthday.Result
		message string
	}

	now := time.Now().In(cfg.Location)
	alerts := make([]alert, 0, 2)
	addAlert := func(date time.Time, when string, warning birthday.WarnFunc, message func(birthday.Result) string) error {
		result, err := birthday.FindOnDate(rows, date, warning)
		if err != nil {
			return err
		}
		if len(result.People) > 0 {
			alerts = append(alerts, alert{when: when, result: result, message: message(result)})
		}
		return nil
	}

	if cfg.ReminderMode == config.ReminderModeToday || cfg.ReminderMode == config.ReminderModeBoth {
		if err := addAlert(now, "hoje", warn, whatsapp.TodayMessage); err != nil {
			return err
		}
	}
	if cfg.ReminderMode == config.ReminderModeNext || cfg.ReminderMode == config.ReminderModeBoth {
		var warning birthday.WarnFunc
		if cfg.ReminderMode == config.ReminderModeNext {
			warning = warn
		}
		if err := addAlert(now.AddDate(0, 0, 1), "amanha", warning, whatsapp.TomorrowMessage); err != nil {
			return err
		}
	}

	if len(alerts) == 0 {
		logger.Info("[SYNC] Nenhum aniversario no periodo verificado", "date", now.Format("2006-01-02"), "mode", cfg.ReminderMode)
		return nil
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
