package rpc

import (
	"connectrpc.com/connect"
	"context"
	"net/http/httptest"
	v1 "portfolio-checker/gen/portfolio/v1"
	"portfolio-checker/gen/portfolio/v1/portfoliov1connect"
	"testing"
)

func TestCheckPortfolioProtocols(t *testing.T) {
	for name, opt := range map[string]connect.ClientOption{"grpc-web": connect.WithGRPCWeb(), "grpc": connect.WithGRPC()} {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewUnstartedServer(NewHandler())
			server.EnableHTTP2 = true
			server.StartTLS()
			defer server.Close()
			client := portfoliov1connect.NewPortfolioServiceClient(server.Client(), server.URL, opt)
			data := []byte("instrument_id,instrument_name,currency,market_value\nA,Alpha,ZAR,60.00\nB,Beta,ZAR,40.00\n")
			res, err := client.CheckPortfolio(context.Background(), connect.NewRequest(&v1.CheckPortfolioRequest{CsvData: data, LimitBasisPoints: 5000}))
			if err != nil {
				t.Fatal(err)
			}
			a := res.Msg.GetAnalysis()
			if a == nil || a.TotalMinor != 10000 || len(a.Holdings) != 2 || !a.Holdings[0].Breached || a.Holdings[1].Breached {
				t.Fatalf("wrong result: %v", a)
			}
			bad, err := client.CheckPortfolio(context.Background(), connect.NewRequest(&v1.CheckPortfolioRequest{LimitBasisPoints: 5000}))
			if err != nil || bad.Msg.GetAnalysis() != nil || len(bad.Msg.GetValidationFailure().GetIssues()) == 0 {
				t.Fatalf("invalid empty input: %v %v", bad, err)
			}
		})
	}
}

func TestRPCOversizedMessage(t *testing.T) {
	s := httptest.NewServer(NewHandler())
	defer s.Close()
	client := portfoliov1connect.NewPortfolioServiceClient(s.Client(), s.URL, connect.WithGRPCWeb())
	_, err := client.CheckPortfolio(context.Background(), connect.NewRequest(&v1.CheckPortfolioRequest{CsvData: make([]byte, (2<<20)+1), LimitBasisPoints: 2000}))
	if connect.CodeOf(err) != connect.CodeResourceExhausted {
		t.Fatalf("wire cap: %v", err)
	}
	res, err := client.CheckPortfolio(context.Background(), connect.NewRequest(&v1.CheckPortfolioRequest{CsvData: make([]byte, (1<<20)+1), LimitBasisPoints: 2000}))
	if err != nil {
		t.Fatal(err)
	}
	issues := res.Msg.GetValidationFailure().GetIssues()
	if res.Msg.GetAnalysis() != nil || len(issues) != 1 || issues[0].Code != "file_too_large" {
		t.Fatal(res.Msg)
	}
}
