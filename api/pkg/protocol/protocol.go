package protocol

type Protocol string

const (
	// HttpJson protocol to be used when deploying model as HTTP/JSON server
	HttpJson Protocol = "HTTP_JSON"
	// UpiV1 protocol to be used when deploying UPI-compatible model
	UpiV1 Protocol = "UPI_V1"
)
