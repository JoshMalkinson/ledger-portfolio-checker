// Package portfolio validates holdings and evaluates an example concentration rule.
// It deliberately has no dependency on HTTP, Protobuf or a database.
package portfolio

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Issue struct {
	Row                  int
	Field, Code, Message string
}
type Holding struct {
	InstrumentID, InstrumentName, Currency string
	MarketValueMinor                       int64
	AllocationBasisPoints                  uint32
	Breached                               bool
}
type Analysis struct {
	TotalMinor int64
	Holdings   []Holding
}
type Result struct {
	Issues   []Issue
	Analysis *Analysis
}

var amountPattern = regexp.MustCompile(`^[0-9]+(?:\.[0-9]{1,2})?$`)

// Check is atomic: if any input is invalid, it returns issues without analysis.
func Check(data []byte, limit uint32) Result {
	var result Result
	add := func(row int, field, code, message string) {
		result.Issues = append(result.Issues, Issue{row, field, code, message})
	}
	if len(data) > 1<<20 {
		add(0, "file", "file_too_large", "Choose a CSV no larger than 1 MiB.")
		return result
	}
	if !utf8.Valid(data) {
		add(0, "file", "invalid_utf8", "Save the CSV with UTF-8 encoding.")
		return result
	}
	if limit == 0 || limit > 10000 {
		add(0, "limit", "invalid_limit", "The limit must be between 0.01% and 100%.")
		return result
	}
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1
	fields, err := reader.Read()
	if errors.Is(err, io.EOF) {
		add(0, "file", "empty_portfolio", "The file contains no holdings.")
		return result
	}
	if err != nil {
		add(1, "file", "invalid_csv", "The CSV header cannot be read.")
		return result
	}
	expected := []string{"instrument_id", "instrument_name", "currency", "market_value"}
	if len(fields) != len(expected) {
		add(1, "header", "invalid_header", "Use the four columns in the downloadable sample, in the same order.")
		return result
	}
	for i, name := range expected {
		if fields[i] != name {
			add(1, "header", "invalid_header", "Expected column: "+name+".")
			return result
		}
	}
	reader.FieldsPerRecord = 4
	analysis := &Analysis{}
	seen := map[string]bool{}
	count := 0
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			line := 0
			var pe *csv.ParseError
			if errors.As(err, &pe) {
				line = pe.Line
			}
			add(line, "file", "invalid_csv", "Expected four CSV fields with correctly escaped quotes.")
			return result
		}
		count++
		if count > 1000 {
			add(0, "file", "too_many_holdings", "Use at most 1,000 holdings.")
			return result
		}
		row, _ := reader.FieldPos(0)
		id := strings.ToUpper(strings.TrimSpace(record[0]))
		name := strings.TrimSpace(record[1])
		if id == "" {
			add(row, "instrument_id", "required", "Instrument ID is required.")
		} else if seen[id] {
			add(row, "instrument_id", "duplicate_id", "Instrument ID is duplicated: "+id+".")
		}
		seen[id] = true
		if name == "" {
			add(row, "instrument_name", "required", "Instrument name is required.")
		}
		if record[2] != "ZAR" {
			add(row, "currency", "unsupported_currency", "Only ZAR is supported; convert the data before importing.")
		}
		value, ok := parseAmount(record[3])
		if !ok {
			add(row, "market_value", "invalid_amount", "Use a positive decimal value up to 1000000000.00, with at most two decimal places.")
		}
		analysis.Holdings = append(analysis.Holdings, Holding{InstrumentID: id, InstrumentName: name, Currency: record[2], MarketValueMinor: value})
		analysis.TotalMinor += value
	}
	if count == 0 {
		add(0, "file", "empty_portfolio", "The file contains no holdings.")
	}
	if len(result.Issues) > 0 {
		return result
	}
	for i := range analysis.Holdings {
		h := &analysis.Holdings[i]
		// With 1,000 holdings capped at 1e11 cents, both products fit in int64.
		weighted := h.MarketValueMinor * 10000
		h.Breached = weighted > analysis.TotalMinor*int64(limit)
		h.AllocationBasisPoints = uint32((weighted + analysis.TotalMinor/2) / analysis.TotalMinor)
	}
	result.Analysis = analysis
	return result
}

func parseAmount(value string) (int64, bool) {
	if !amountPattern.MatchString(value) {
		return 0, false
	}
	parts := strings.SplitN(value, ".", 2)
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || whole > 1000000000 {
		return 0, false
	}
	var cents int64
	if len(parts) == 2 {
		fraction := parts[1]
		if len(fraction) == 1 {
			fraction += "0"
		}
		cents, _ = strconv.ParseInt(fraction, 10, 64)
	}
	amount := whole*100 + cents
	return amount, amount > 0 && amount <= 100000000000
}
