package googlesheets

import (
	"context"
	"fmt"
	"strings"

	"github.com/andrepieresan/birth-reminder/internal/config"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func ReadRows(ctx context.Context, cfg config.Config) ([][]string, error) {
	service, err := sheets.NewService(ctx,
		option.WithCredentialsFile(cfg.GoogleCredentialsFile),
		option.WithScopes(sheets.SpreadsheetsReadonlyScope),
	)
	if err != nil {
		return nil, fmt.Errorf("criar cliente: %w", err)
	}

	response, err := service.Spreadsheets.Values.Get(cfg.GoogleSheetID, cfg.GoogleSheetRange).
		ValueRenderOption("FORMATTED_VALUE").
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("consultar intervalo %q: %w", cfg.GoogleSheetRange, err)
	}

	rows := make([][]string, 0, len(response.Values))
	for _, sourceRow := range response.Values {
		row := make([]string, len(sourceRow))
		for column, value := range sourceRow {
			row[column] = strings.TrimSpace(fmt.Sprint(value))
		}
		rows = append(rows, row)
	}

	return rows, nil
}
