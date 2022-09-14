package types

type TracingDetail struct {
	Input  map[string]interface{} `json:"input"`
	Output map[string]interface{} `json:"output"`
	Spec   interface{}            `json:"spec"`
	OpType OperationType          `json:"operation_type"`
}

type OperationType string

const (
	VariableOpType    OperationType = "variable_op"
	CreateTableOpType OperationType = "create_table_op"
	FeastOpType       OperationType = "feast_op"
	JsonOutputOpType  OperationType = "json_output_op"
	TableJoinOpType   OperationType = "table_join_op"
	TableTransformOp  OperationType = "table_transform_op"
)

type PredictResponse struct {
	Response Payload           `json:"response"`
	Tracing  *OperationTracing `json:"operation_tracing"`
}

type OperationTracing struct {
	PreprocessTracing  []TracingDetail `json:"preprocess"`
	PostprocessTracing []TracingDetail `json:"postprocess"`
}
