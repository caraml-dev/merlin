package logger

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	PrometheusUAPrefix = "Prometheus/"

	// Since K8s 1.8, prober requests have
	// User-Agent = "kube-probe/{major-version}.{minor-version}".
	KubeProbeUAPrefix = "kube-probe/"

	// Istio with mTLS rewrites probes, but their probes pass a different
	// user-agent.  So we augment the probes with this header.
	KubeletProbeHeaderName = "K-Kubelet-Probe"

	MerlinLogIdHeader = "X-Merlin-Log-Id"
)

type LoggerHandler struct {
	dispatcher *Dispatcher
	logMode    LogMode
	next       http.Handler
	logger     *zap.SugaredLogger
}

func NewLoggerHandler(dispatcher *Dispatcher, logMode LogMode, next http.Handler, logger *zap.SugaredLogger) http.Handler {
	return &LoggerHandler{
		dispatcher: dispatcher,
		logMode:    logMode,
		next:       next,
		logger:     logger,
	}
}

func (eh *LoggerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if isKubeletProbe(r) || isPrometheusScrape(r) {
		if eh.next != nil {
			eh.next.ServeHTTP(w, r)
		}
		return
	}

	// Get or Create an ID
	id := getOrCreateID(r)
	logEntry := &LogEntry{
		RequestId:      id,
		EventTimestamp: timestamppb.Now(),
	}

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	// Read Request Payload
	if eh.logMode == LogModeAll || eh.logMode == LogModeRequestOnly {
		logEntry.RequestPayload = &RequestPayload{
			Headers: formatHeader(r.Header),
			Body:    requestBody,
		}
	}

	defer func() {
		if err := eh.dispatcher.Submit(logEntry); err != nil {
			eh.logger.Errorf("error submitting log entry: %v", err)
		}
	}()

	r.Header.Set(MerlinLogIdHeader, id)
	// Proxy Request
	r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	rr := httptest.NewRecorder()
	eh.next.ServeHTTP(rr, r)

	copyHeader(w.Header(), rr.Header())
	respBody, err := io.ReadAll(rr.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Read Response Payload
	if eh.logMode == LogModeAll || eh.logMode == LogModeResponseOnly {
		logEntry.ResponsePayload = &ResponsePayload{
			StatusCode: rr.Code,
			Body:       respBody,
		}
	}

	w.WriteHeader(rr.Code)
	_, err = w.Write(respBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getOrCreateID(r *http.Request) string {
	id := r.Header.Get(MerlinLogIdHeader)
	if id == "" {
		id = uuid.New().String()
	}
	return id
}

func formatHeader(header http.Header) map[string]string {
	formatted := map[string]string{}
	for k, v := range header {
		formatted[k] = strings.Join(v, ",")
	}

	return formatted
}

func isPrometheusScrape(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get("User-Agent"), PrometheusUAPrefix)
}

func isKubeletProbe(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get("User-Agent"), KubeProbeUAPrefix) ||
		r.Header.Get(KubeletProbeHeaderName) != ""
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Set(k, v)
		}
	}
}
