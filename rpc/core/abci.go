package core

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/bytes"
	"github.com/cometbft/cometbft/proxy"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	rpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
)

var tracer = otel.Tracer("cometbft")

// ABCIQuery queries the application for some information.
// More: https://docs.cometbft.com/v0.37/rpc/#/ABCI/abci_query
func ABCIQuery(
	ctx *rpctypes.Context,
	path string,
	data bytes.HexBytes,
	height int64,
	prove bool,
) (*ctypes.ResultABCIQuery, error) {

	// Extract the existing span context from the incoming request
	parentCtx := otel.GetTextMapPropagator().Extract(ctx.Context(), propagation.HeaderCarrier(ctx.HTTPReq.Header))

	// Start a new span representing the request
	// The span ends when the request is complete
	_, span := tracer.Start(parentCtx, path, trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	span.SetAttributes(attribute.String("http.method", path))

	resQuery, err := env.ProxyAppQuery.QuerySync(abci.RequestQuery{
		Path:   path,
		Data:   data,
		Height: height,
		Prove:  prove,
	})

	if err != nil {
		span.SetStatus(otelcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otelcodes.Ok, "OK")

	return &ctypes.ResultABCIQuery{Response: *resQuery}, nil
}

// ABCIInfo gets some info about the application.
// More: https://docs.cometbft.com/v0.37/rpc/#/ABCI/abci_info
func ABCIInfo(ctx *rpctypes.Context) (*ctypes.ResultABCIInfo, error) {
	resInfo, err := env.ProxyAppQuery.InfoSync(proxy.RequestInfo)
	if err != nil {
		return nil, err
	}

	return &ctypes.ResultABCIInfo{Response: *resInfo}, nil
}
