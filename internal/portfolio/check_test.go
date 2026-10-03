package portfolio

import (
	"fmt"
	"strings"
	"testing"
)

const header = "instrument_id,instrument_name,currency,market_value\n"
const valid = header + "A,Alpha,ZAR,60.00\nB,Beta,ZAR,40.00\n"

func TestCheckAllocationAndBoundary(t *testing.T) {
	r := Check([]byte(valid), 5000)
	if r.Analysis == nil || len(r.Issues) != 0 {
		t.Fatalf("invalid result: %+v", r)
	}
	a := r.Analysis
	if a.TotalMinor != 10000 {
		t.Fatal(a.TotalMinor)
	}
	if len(a.Holdings) != 2 || a.Holdings[0].AllocationBasisPoints != 6000 || a.Holdings[1].AllocationBasisPoints != 4000 {
		t.Fatal(a.Holdings)
	}
	if !a.Holdings[0].Breached || a.Holdings[1].Breached {
		t.Fatal("60/40 at 50%")
	}
	if Check([]byte(valid), 6000).Analysis.Holdings[0].Breached {
		t.Fatal("equality must pass")
	}
	near := Check([]byte(header+"A,Alpha,ZAR,5000.01\nB,Beta,ZAR,4999.99\n"), 5000)
	if !near.Analysis.Holdings[0].Breached || near.Analysis.Holdings[0].AllocationBasisPoints != 5000 {
		t.Fatal("unrounded comparison required")
	}
}

func TestCheckRejectsInvalidInputs(t *testing.T) {
	cases := []struct {
		name, csv, code string
		limit           uint32
	}{
		{"empty", "", "empty_portfolio", 2000},
		{"header only", header, "empty_portfolio", 2000},
		{"missing header", "A,Alpha,ZAR,1\n", "invalid_header", 2000},
		{"extra header", strings.TrimSpace(header) + ",extra\n", "invalid_header", 2000},
		{"reordered header", "instrument_name,instrument_id,currency,market_value\n", "invalid_header", 2000},
		{"extra field", header + "A,Alpha,ZAR,1,extra\n", "invalid_csv", 2000},
		{"missing field", header + "A,Alpha,ZAR\n", "invalid_csv", 2000},
		{"duplicate", header + " a ,Alpha,ZAR,1\nA,Beta,ZAR,2\n", "duplicate_id", 2000},
		{"name", header + "A, ,ZAR,1\n", "required", 2000},
		{"id", header + ",Alpha,ZAR,1\n", "required", 2000},
		{"currency", header + "A,Alpha,USD,1\n", "unsupported_currency", 2000},
		{"zero", header + "A,Alpha,ZAR,0\n", "invalid_amount", 2000},
		{"negative", header + "A,Alpha,ZAR,-1\n", "invalid_amount", 2000},
		{"fraction", header + "A,Alpha,ZAR,1.001\n", "invalid_amount", 2000},
		{"exponent", header + "A,Alpha,ZAR,1e3\n", "invalid_amount", 2000},
		{"grouped", header + "A,Alpha,ZAR,\"1,000\"\n", "invalid_amount", 2000},
		{"symbol", header + "A,Alpha,ZAR,R1\n", "invalid_amount", 2000},
		{"cap", header + "A,Alpha,ZAR,1000000000.01\n", "invalid_amount", 2000},
		{"huge integer", header + "A,Alpha,ZAR,99999999999999999999999999\n", "invalid_amount", 2000},
		{"quote", header + "A,\"Alpha,ZAR,1\n", "invalid_csv", 2000},
		{"utf8", header + "A,\xff,ZAR,1\n", "invalid_utf8", 2000},
		{"file size", strings.Repeat("a", (1<<20)+1), "file_too_large", 2000},
		{"limit zero", valid, "invalid_limit", 0},
		{"limit high", valid, "invalid_limit", 10001},
	}
	var many strings.Builder
	many.WriteString(header)
	for i := 0; i < 1001; i++ {
		fmt.Fprintf(&many, "ID%d,Fund,ZAR,1\n", i)
	}
	cases = append(cases, struct {
		name, csv, code string
		limit           uint32
	}{"row cap", many.String(), "too_many_holdings", 2000})
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := Check([]byte(c.csv), c.limit)
			if r.Analysis != nil || len(r.Issues) == 0 {
				t.Fatalf("expected validation: %+v", r)
			}
			found := false
			for _, issue := range r.Issues {
				if issue.Code == c.code {
					found = true
				}
				if issue.Message == "" {
					t.Fatal("empty message")
				}
			}
			if !found {
				t.Fatalf("missing %s: %+v", c.code, r.Issues)
			}
		})
	}
}

func TestCheckCSVSourceLines(t *testing.T) {
	r := Check([]byte("\xef\xbb\xbf"+header+"\nA,\"Alpha\nFund\",ZAR,1\nB,Beta,ZAR,nope\n"), 2000)
	if r.Analysis != nil || len(r.Issues) != 1 || r.Issues[0].Row != 5 {
		t.Fatalf("physical row: %+v", r)
	}
	good := Check([]byte("\xef\xbb\xbf"+header+"\n a ,\"Alpha, Fund\",ZAR,0.1\n"), 10000)
	if good.Analysis == nil || good.Analysis.TotalMinor != 10 || good.Analysis.Holdings[0].InstrumentID != "A" || good.Analysis.Holdings[0].InstrumentName != "Alpha, Fund" {
		t.Fatalf("BOM and quotes: %+v", good)
	}
}

func TestCheckMaximumValues(t *testing.T) {
	var b strings.Builder
	b.WriteString(header)
	for i := 0; i < 1000; i++ {
		fmt.Fprintf(&b, "ID%d,Fund,ZAR,1000000000.00\n", i)
	}
	r := Check([]byte(b.String()), 1)
	if r.Analysis == nil || r.Analysis.TotalMinor != 100000000000000 {
		t.Fatalf("maximum: %+v", r)
	}
	if !r.Analysis.Holdings[0].Breached || r.Analysis.Holdings[0].AllocationBasisPoints != 10 {
		t.Fatal("overflow in comparison")
	}
}
