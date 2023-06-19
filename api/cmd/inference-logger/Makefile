APP_NAME=inference-logger
LOG_URL?=kafka:localhost:9092

.PHONY: build
build:
	@go build -o bin/${APP_NAME} ./main.go

.PHONY: run
run:
	@rm /tmp/agent.sock || true
	@SERVING_READINESS_PROBE='{"tcpSocket":{"port":8080,"host":"127.0.0.1"},"successThreshold":1}' UNIX_SOCKET_PATH="/tmp/agent.sock" ./bin/${APP_NAME} -log-url="${LOG_URL}"

.PHONY: run-mock
run-mock:
	@go run ../../pkg/inference-logger/mock-server/mock_server.go
