package internal

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func camelToTitle(s string) string {
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	title := re.ReplaceAllString(s, `$1 $2`)
	return strings.ToUpper(title)
}

func ParseColumns[T any](allColumns bool, customColumns []string, defaultColumns []string) []string {
	if allColumns {
		var fields []string
		t := reflect.TypeOf(*new(T))
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fields = append(fields, field.Name)
		}
		return fields
	}

	if len(customColumns) > 0 {
		return customColumns
	}

	return defaultColumns
}

func RenderTable[T any](items []T, columns []string) {
	if len(items) == 0 {
		fmt.Println("No data to display.")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)

	var tableHeaders []string
	for _, column := range columns {
		tableHeaders = append(tableHeaders, camelToTitle(column))
	}
	table.SetHeader(tableHeaders)

	for _, item := range items {
		value := reflect.ValueOf(item)
		var tableRow []string
		for _, column := range columns {
			field := value.FieldByNameFunc(func(s string) bool {
				return strings.EqualFold(s, column)
			})
			if !field.IsValid() || !field.CanInterface() {
				tableRow = append(tableRow, "")
			} else {
				tableRow = append(tableRow, fmt.Sprintf("%v", field.Interface()))
			}
		}
		table.Append(tableRow)
	}

	table.Render()
}
