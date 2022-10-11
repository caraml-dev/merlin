package types

type TracingDetail struct {
	Input  map[string]interface{} `json:"input"`
	Output map[string]interface{} `json:"output"`
	Spec   interface{}            `json:"spec"`
	OpType OperationType          `json:"operation_type"`
}

type OperationType string

const (
	VariableOpType         OperationType = "variable_op"
	CreateTableOpType      OperationType = "create_table_op"
	FeastOpType            OperationType = "feast_op"
	JsonOutputOpType       OperationType = "json_output_op"
	EncoderOpType          OperationType = "encoder_op"
	TableJoinOpType        OperationType = "table_join_op"
	TableTransformOp       OperationType = "table_transform_op"
	UPIAutoloadingOp       OperationType = "upi_autoloading_op"
	UPIPreprocessOutputOp  OperationType = "upi_preprocess_output_op"
	UPIPostprocessOutputOp OperationType = "upi_postprocess_output_op"
)

type PredictResponse struct {
	Response Payload           `json:"response"`
	Tracing  *OperationTracing `json:"operation_tracing"`
}

type OperationTracing struct {
	PreprocessTracing  []TracingDetail `json:"preprocess"`
	PostprocessTracing []TracingDetail `json:"postprocess"`
}
