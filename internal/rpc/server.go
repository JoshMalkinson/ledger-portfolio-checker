package rpc

import (
	"connectrpc.com/connect"
	"context"
	"net/http"
	v1 "portfolio-checker/gen/portfolio/v1"
	"portfolio-checker/gen/portfolio/v1/portfoliov1connect"
	"portfolio-checker/internal/portfolio"
)

type Server struct{}

func (s *Server) CheckPortfolio(ctx context.Context, req *connect.Request[v1.CheckPortfolioRequest]) (*connect.Response[v1.CheckPortfolioResponse], error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	result := portfolio.Check(req.Msg.CsvData, req.Msg.LimitBasisPoints)
	response := &v1.CheckPortfolioResponse{}
	if result.Analysis == nil {
		failure := &v1.ValidationFailure{}
		for _, i := range result.Issues {
			failure.Issues = append(failure.Issues, &v1.ValidationIssue{Row: uint32(i.Row), Field: i.Field, Code: i.Code, Message: i.Message})
		}
		response.Result = &v1.CheckPortfolioResponse_ValidationFailure{ValidationFailure: failure}
	} else {
		a := &v1.PortfolioAnalysis{TotalMinor: result.Analysis.TotalMinor}
		for _, h := range result.Analysis.Holdings {
			a.Holdings = append(a.Holdings, &v1.Holding{InstrumentId: h.InstrumentID, InstrumentName: h.InstrumentName, Currency: h.Currency, MarketValueMinor: h.MarketValueMinor, AllocationBasisPoints: h.AllocationBasisPoints, Breached: h.Breached})
		}
		response.Result = &v1.CheckPortfolioResponse_Analysis{Analysis: a}
	}
	return connect.NewResponse(response), nil
}

func NewHandler() http.Handler {
	path, handler := portfoliov1connect.NewPortfolioServiceHandler(&Server{}, connect.WithReadMaxBytes(2<<20))
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	return mux
}
