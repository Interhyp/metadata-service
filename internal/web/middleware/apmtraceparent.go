package middleware

// Elastic APM license: https://pkg.go.dev/go.elastic.co/apm/module/apmhttp/v2?tab=licenses

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"github.com/StephanHCB/go-backend-service-common/web/middleware/requestid"
	"go.elastic.co/apm/module/apmhttp/v2"
	"go.elastic.co/apm/v2"
	"math/rand"
	"net/http"
	"time"
)

var TraceContextFetcherForResponseHeaders = RestoreOrCreateTraceContextWithoutAPM

// RestoreOrCreateTraceContextWithoutAPM is designed as a fallback to enable trace propagation even if Elastic APM is not configured or disabled.
// It uses some Elastic apm go agent functions for compatibility but does not require the middleware to be active.
func RestoreOrCreateTraceContextWithoutAPM(r *http.Request) apm.TraceContext {
	// create own trace context
	traceContext, ok := getRequestTraceparent(r)
	if ok {
		traceContext.State, _ = apmhttp.ParseTracestateHeader(r.Header[apmhttp.TracestateHeader]...)
	} else {
		// no trace context restored from header -> we create our own
		// code mostly copied from https://pkg.go.dev/go.elastic.co/apm/v2@v2.1.0#Tracer.StartTransactionOptions
		var seed int64
		if err := binary.Read(cryptorand.Reader, binary.LittleEndian, &seed); err != nil {
			seed = time.Now().UnixNano()
		}
		random := rand.New(rand.NewSource(seed))
		binary.LittleEndian.PutUint64(traceContext.Trace[:8], random.Uint64())
		binary.LittleEndian.PutUint64(traceContext.Trace[8:], random.Uint64())
		copy(traceContext.Span[:], traceContext.Trace[:])
		//we do not set trace state by ourselves as it is vendor specific and contains the sample rate for Elastic apm
		traceContext.Options.WithRecorded(false) //do not record our custom IDs
	}
	return traceContext
}

// UseElasticApmTraceContext extracts the apm.TraceContext from the apm.Transaction stored in the Context. The transaction
// is contained in the context if Elastic APM is enabled and the apmhttp handler has started the transaction.
// See https://pkg.go.dev/go.elastic.co/apm/module/apmhttp/v2@v2.1.0#StartTransactionWithBody
func UseElasticApmTraceContext(r *http.Request) apm.TraceContext {
	ctx := r.Context()
	return apm.TransactionFromContext(ctx).TraceContext()
}

func ApmTraceResponseHeaders(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var traceContext = TraceContextFetcherForResponseHeaders(r)
		ctx := r.Context()

		//add traceparent as request id to context
		//(response header is added later by middleware requestidinresponse.AddRequestIdHeaderToResponse)
		requestId := apmhttp.FormatTraceparentHeader(traceContext)
		requestIdCtx := requestid.PutReqID(ctx, requestId)

		//propagate tracestate (containing sampling rate, vendor etc.) if present in the trace context
		//I appreciate a more consistent solution (one id is added to context, other is added as header directly)
		if tracestate := traceContext.State.String(); tracestate != "" {
			w.Header().Set(apmhttp.TracestateHeader, tracestate)
		}

		next.ServeHTTP(w, r.WithContext(requestIdCtx))
	}
	return http.HandlerFunc(fn)
}

// copied (and slightly modified) from pkg\mod\go.elastic.co\apm\module\apmhttp\v2@v2.1.0\handler.go
// why the hell is this private? :/
func getRequestTraceparent(req *http.Request) (apm.TraceContext, bool) {
	if values := req.Header[apmhttp.W3CTraceparentHeader]; len(values) == 1 && values[0] != "" {
		if c, err := apmhttp.ParseTraceparentHeader(values[0]); err == nil {
			return c, true
		}
	}
	return apm.TraceContext{}, false
}
