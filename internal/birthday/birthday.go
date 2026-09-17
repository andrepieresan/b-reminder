package birthday

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	brazilianDate = regexp.MustCompile(`^(\d{1,2})[/-](\d{1,2})(?:[/-]\d{2,4})?$`)
	isoDate       = regexp.MustCompile(`^\d{4}-(\d{1,2})-(\d{1,2})$`)
	activeValues  = map[string]struct{}{
		"1": {}, "ativo": {}, "sim": {}, "s": {}, "true": {}, "yes": {},
	}
)

type Person struct {
	Name     string `json:"name"`
	Birthday string `json:"birthday"`
}

type Result struct {
	Date   string
	People []Person
}

type WarnFunc func(row int, reason string)

func FindOnDate(rows [][]string, date time.Time, warn WarnFunc) (Result, error) {
	result := Result{Date: date.Format("2006-01-02"), People: []Person{}}

	if len(rows) == 0 {
		return result, nil
	}

	nameColumn := findColumn(rows[0], "nome", "funcionario")
	birthdayColumn := findColumn(rows[0], "aniversario", "data de aniversario")
	activeColumn := findColumn(rows[0], "ativo", "status")
	if nameColumn < 0 || birthdayColumn < 0 {
		return Result{}, fmt.Errorf(`a planilha precisa das colunas "nome" e "aniversario"`)
	}

	for index, row := range rows[1:] {
		rowNumber := index + 2
		name := cell(row, nameColumn)
		birthdayText := cell(row, birthdayColumn)

		if name == "" && birthdayText == "" {
			continue
		}
		if name == "" {
			warnRow(warn, rowNumber, "nome vazio")
			continue
		}
		if activeColumn >= 0 && !isActive(cell(row, activeColumn)) {
			continue
		}

		day, month, ok := parseDayMonth(birthdayText)
		if !ok {
			warnRow(warn, rowNumber, "data de aniversario invalida")
			continue
		}
		if day != date.Day() || month != date.Month() {
			continue
		}

		result.People = append(result.People, Person{Name: name, Birthday: birthdayText})
	}

	return result, nil
}

func findColumn(headers []string, aliases ...string) int {
	for column, header := range headers {
		normalizedHeader := normalize(header)
		for _, alias := range aliases {
			if normalizedHeader == alias {
				return column
			}
		}
	}
	return -1
}

func parseDayMonth(value string) (int, time.Month, bool) {
	value = strings.TrimSpace(value)
	if match := brazilianDate.FindStringSubmatch(value); match != nil {
		return validDayMonth(match[1], match[2])
	}
	if match := isoDate.FindStringSubmatch(value); match != nil {
		return validDayMonth(match[2], match[1])
	}
	return 0, 0, false
}

func validDayMonth(dayText, monthText string) (int, time.Month, bool) {
	day, dayErr := strconv.Atoi(dayText)
	monthNumber, monthErr := strconv.Atoi(monthText)
	if dayErr != nil || monthErr != nil || monthNumber < 1 || monthNumber > 12 || day < 1 {
		return 0, 0, false
	}

	month := time.Month(monthNumber)
	candidate := time.Date(2024, month, day, 0, 0, 0, 0, time.UTC)
	if candidate.Day() != day || candidate.Month() != month {
		return 0, 0, false
	}
	return day, month, true
}

func isActive(value string) bool {
	_, ok := activeValues[normalize(value)]
	return ok
}

func normalize(value string) string {
	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u", "ç", "c",
	)
	return replacer.Replace(strings.ToLower(strings.TrimSpace(value)))
}

func cell(row []string, column int) string {
	if column < 0 || column >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[column])
}

func warnRow(warn WarnFunc, row int, reason string) {
	if warn != nil {
		warn(row, reason)
	}
}
