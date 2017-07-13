package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/hnlq715/doggy/utils"
	"go.uber.org/zap"
)

// Key to use when setting the request ID.
type ctxKeyRequestID int

// RequestIDKey is the key that holds th unique request ID in a request context.
const requestIDKey ctxKeyRequestID = 0

var prefix string
var reqid uint64

func init() {
	hostname, err := os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}
	var buf [12]byte
	var b64 string
	for len(b64) < 10 {
		rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}

	prefix = fmt.Sprintf("%s/%s", hostname, b64[0:10])
}

type TraceID struct {
}

// NewTraceID returns a new TraceID instance
func NewTraceID() *TraceID {
	return &TraceID{}
}

func (m *TraceID) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	myid := atomic.AddUint64(&reqid, 1)
	traceID := fmt.Sprintf("%s-%06d", prefix, myid)
	ctx := context.WithValue(r.Context(), requestIDKey, traceID)
	log := utils.LogFromContext(ctx).With(zap.String("trace_id", traceID))
	ctx = utils.ContextWithLog(ctx, log)
	next(rw, r.WithContext(ctx))
	rw.Header().Set("TraceID", traceID)
}

func GetReqID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return reqID
	}
	return ""
}
