package main

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const (
	httpPort = 8080
	grpcPort = 8002
)

func main() {
	l, _ := zap.NewProduction()
	log := l.Sugar()

	log.Infof("running mock model service at port: %d", httpPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), NewMockModelService(log)); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}

type MockModelService struct {
	logger *zap.SugaredLogger
}

func NewMockModelService(log *zap.SugaredLogger) http.Handler {
	return &MockModelService{logger: log}
}

func (m *MockModelService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	time.Sleep(20 * time.Millisecond)
	_, err := w.Write([]byte(`{"predictions" : [[1,2,3,4]]}`))
	if err != nil {
		fmt.Println(err)
	}
}
