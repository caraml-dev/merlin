package models

type Protocol string

const (
	// ProtocolHttpJson protocol to be used when deploying model as HTTP/JSON server
	ProtocolHttpJson Protocol = "HTTP_JSON"
	// ProtocolUpiV1 protocol to be used when deploying UPI-compatible model
	ProtocolUpiV1    Protocol = "UPI_V1"
)
