// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.21.9
// source: transformer/spec/standard_transformer.proto

package spec

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type StandardTransformerConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TransformerConfig   *TransformerConfig   `protobuf:"bytes,1,opt,name=transformerConfig,proto3" json:"transformerConfig,omitempty"`
	PredictionLogConfig *PredictionLogConfig `protobuf:"bytes,2,opt,name=predictionLogConfig,proto3" json:"predictionLogConfig,omitempty"`
}

func (x *StandardTransformerConfig) Reset() {
	*x = StandardTransformerConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_spec_standard_transformer_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StandardTransformerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StandardTransformerConfig) ProtoMessage() {}

func (x *StandardTransformerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_spec_standard_transformer_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StandardTransformerConfig.ProtoReflect.Descriptor instead.
func (*StandardTransformerConfig) Descriptor() ([]byte, []int) {
	return file_transformer_spec_standard_transformer_proto_rawDescGZIP(), []int{0}
}

func (x *StandardTransformerConfig) GetTransformerConfig() *TransformerConfig {
	if x != nil {
		return x.TransformerConfig
	}
	return nil
}

func (x *StandardTransformerConfig) GetPredictionLogConfig() *PredictionLogConfig {
	if x != nil {
		return x.PredictionLogConfig
	}
	return nil
}

type TransformerConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Feast       []*FeatureTable `protobuf:"bytes,1,rep,name=feast,proto3" json:"feast,omitempty"` // for backward compatibility
	Preprocess  *Pipeline       `protobuf:"bytes,2,opt,name=preprocess,proto3" json:"preprocess,omitempty"`
	Postprocess *Pipeline       `protobuf:"bytes,3,opt,name=postprocess,proto3" json:"postprocess,omitempty"`
}

func (x *TransformerConfig) Reset() {
	*x = TransformerConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_spec_standard_transformer_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TransformerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransformerConfig) ProtoMessage() {}

func (x *TransformerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_spec_standard_transformer_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TransformerConfig.ProtoReflect.Descriptor instead.
func (*TransformerConfig) Descriptor() ([]byte, []int) {
	return file_transformer_spec_standard_transformer_proto_rawDescGZIP(), []int{1}
}

func (x *TransformerConfig) GetFeast() []*FeatureTable {
	if x != nil {
		return x.Feast
	}
	return nil
}

func (x *TransformerConfig) GetPreprocess() *Pipeline {
	if x != nil {
		return x.Preprocess
	}
	return nil
}

func (x *TransformerConfig) GetPostprocess() *Pipeline {
	if x != nil {
		return x.Postprocess
	}
	return nil
}

type Pipeline struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Inputs          []*Input          `protobuf:"bytes,1,rep,name=inputs,proto3" json:"inputs,omitempty"`
	Transformations []*Transformation `protobuf:"bytes,2,rep,name=transformations,proto3" json:"transformations,omitempty"`
	Outputs         []*Output         `protobuf:"bytes,3,rep,name=outputs,proto3" json:"outputs,omitempty"`
}

func (x *Pipeline) Reset() {
	*x = Pipeline{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_spec_standard_transformer_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Pipeline) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Pipeline) ProtoMessage() {}

func (x *Pipeline) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_spec_standard_transformer_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Pipeline.ProtoReflect.Descriptor instead.
func (*Pipeline) Descriptor() ([]byte, []int) {
	return file_transformer_spec_standard_transformer_proto_rawDescGZIP(), []int{2}
}

func (x *Pipeline) GetInputs() []*Input {
	if x != nil {
		return x.Inputs
	}
	return nil
}

func (x *Pipeline) GetTransformations() []*Transformation {
	if x != nil {
		return x.Transformations
	}
	return nil
}

func (x *Pipeline) GetOutputs() []*Output {
	if x != nil {
		return x.Outputs
	}
	return nil
}

type Input struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Ideally, Input definition should be
	// ```
	//
	//	 oneof input {
	//	   repeated Variable variables = 1;
	//	   repeated FeatureTable feast = 2;
	//	   repeated Table tables = 3;
	//	}
	//
	// ```
	// however it's not possible to have repeated field in oneof
	// https://github.com/protocolbuffers/protobuf/issues/2592
	// Thus we will handle the oneof behavior in the code side
	Variables []*Variable     `protobuf:"bytes,1,rep,name=variables,proto3" json:"variables,omitempty"`
	Feast     []*FeatureTable `protobuf:"bytes,2,rep,name=feast,proto3" json:"feast,omitempty"`
	Tables    []*Table        `protobuf:"bytes,3,rep,name=tables,proto3" json:"tables,omitempty"`
	Encoders  []*Encoder      `protobuf:"bytes,4,rep,name=encoders,proto3" json:"encoders,omitempty"`
	Autoload  *UPIAutoload    `protobuf:"bytes,5,opt,name=autoload,proto3" json:"autoload,omitempty"`
}

func (x *Input) Reset() {
	*x = Input{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_spec_standard_transformer_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Input) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Input) ProtoMessage() {}

func (x *Input) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_spec_standard_transformer_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Input.ProtoReflect.Descriptor instead.
func (*Input) Descriptor() ([]byte, []int) {
	return file_transformer_spec_standard_transformer_proto_rawDescGZIP(), []int{3}
}

func (x *Input) GetVariables() []*Variable {
	if x != nil {
		return x.Variables
	}
	return nil
}

func (x *Input) GetFeast() []*FeatureTable {
	if x != nil {
		return x.Feast
	}
	return nil
}

func (x *Input) GetTables() []*Table {
	if x != nil {
		return x.Tables
	}
	return nil
}

func (x *Input) GetEncoders() []*Encoder {
	if x != nil {
		return x.Encoders
	}
	return nil
}

func (x *Input) GetAutoload() *UPIAutoload {
	if x != nil {
		return x.Autoload
	}
	return nil
}

type Transformation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Ideally, Transformation definition should be
	// ```
	//
	//	 oneof input {
	//	   repeated TableJoin tableJoin = 1;
	//	   repeated TableTransformation tableTransformation = 2;
	//	}
	//
	// ```
	// however it's not possible to have repeated field in oneof
	// https://github.com/protocolbuffers/protobuf/issues/2592
	// Thus we will handle the oneof behavior in the code side
	TableJoin           *TableJoin           `protobuf:"bytes,1,opt,name=tableJoin,proto3" json:"tableJoin,omitempty"`
	TableTransformation *TableTransformation `protobuf:"bytes,2,opt,name=tableTransformation,proto3" json:"tableTransformation,omitempty"`
	Variables           []*Variable          `protobuf:"bytes,3,rep,name=variables,proto3" json:"variables,omitempty"`
}

func (x *Transformation) Reset() {
	*x = Transformation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_spec_standard_transformer_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Transformation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Transformation) ProtoMessage() {}

func (x *Transformation) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_spec_standard_transformer_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Transformation.ProtoReflect.Descriptor instead.
func (*Transformation) Descriptor() ([]byte, []int) {
	return file_transformer_spec_standard_transformer_proto_rawDescGZIP(), []int{4}
}

func (x *Transformation) GetTableJoin() *TableJoin {
	if x != nil {
		return x.TableJoin
	}
	return nil
}

func (x *Transformation) GetTableTransformation() *TableTransformation {
	if x != nil {
		return x.TableTransformation
	}
	return nil
}

func (x *Transformation) GetVariables() []*Variable {
	if x != nil {
		return x.Variables
	}
	return nil
}

type Output struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	JsonOutput           *JsonOutput           `protobuf:"bytes,1,opt,name=jsonOutput,proto3" json:"jsonOutput,omitempty"`
	UpiPreprocessOutput  *UPIPreprocessOutput  `protobuf:"bytes,2,opt,name=upiPreprocessOutput,proto3" json:"upiPreprocessOutput,omitempty"`
	UpiPostprocessOutput *UPIPostprocessOutput `protobuf:"bytes,3,opt,name=upiPostprocessOutput,proto3" json:"upiPostprocessOutput,omitempty"`
}

func (x *Output) Reset() {
	*x = Output{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_spec_standard_transformer_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Output) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Output) ProtoMessage() {}

func (x *Output) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_spec_standard_transformer_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Output.ProtoReflect.Descriptor instead.
func (*Output) Descriptor() ([]byte, []int) {
	return file_transformer_spec_standard_transformer_proto_rawDescGZIP(), []int{5}
}

func (x *Output) GetJsonOutput() *JsonOutput {
	if x != nil {
		return x.JsonOutput
	}
	return nil
}

func (x *Output) GetUpiPreprocessOutput() *UPIPreprocessOutput {
	if x != nil {
		return x.UpiPreprocessOutput
	}
	return nil
}

func (x *Output) GetUpiPostprocessOutput() *UPIPostprocessOutput {
	if x != nil {
		return x.UpiPostprocessOutput
	}
	return nil
}

var File_transformer_spec_standard_transformer_proto protoreflect.FileDescriptor

var file_transformer_spec_standard_transformer_proto_rawDesc = []byte{
	0x0a, 0x2b, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2f, 0x73, 0x70,
	0x65, 0x63, 0x2f, 0x73, 0x74, 0x61, 0x6e, 0x64, 0x61, 0x72, 0x64, 0x5f, 0x74, 0x72, 0x61, 0x6e,
	0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x6d,
	0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65,
	0x72, 0x1a, 0x1c, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2f, 0x73,
	0x70, 0x65, 0x63, 0x2f, 0x66, 0x65, 0x61, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x1c, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2f, 0x73, 0x70, 0x65,
	0x63, 0x2f, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f,
	0x76, 0x61, 0x72, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b,
	0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2f, 0x73, 0x70, 0x65, 0x63,
	0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x65, 0x6e,
	0x63, 0x6f, 0x64, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x21, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x75, 0x70,
	0x69, 0x5f, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x23,
	0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2f, 0x73, 0x70, 0x65, 0x63,
	0x2f, 0x75, 0x70, 0x69, 0x5f, 0x61, 0x75, 0x74, 0x6f, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x25, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72,
	0x2f, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x70, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x5f, 0x6c, 0x6f, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xcb, 0x01, 0x0a, 0x19, 0x53,
	0x74, 0x61, 0x6e, 0x64, 0x61, 0x72, 0x64, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d,
	0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x53, 0x0a, 0x11, 0x74, 0x72, 0x61, 0x6e,
	0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f,
	0x72, 0x6d, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x11, 0x74, 0x72, 0x61, 0x6e,
	0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x59, 0x0a,
	0x13, 0x70, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4c, 0x6f, 0x67, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x6d, 0x65, 0x72,
	0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e,
	0x50, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4c, 0x6f, 0x67, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x52, 0x13, 0x70, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4c,
	0x6f, 0x67, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0xc9, 0x01, 0x0a, 0x11, 0x54, 0x72, 0x61,
	0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x36,
	0x0a, 0x05, 0x66, 0x65, 0x61, 0x73, 0x74, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e,
	0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d,
	0x65, 0x72, 0x2e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x52,
	0x05, 0x66, 0x65, 0x61, 0x73, 0x74, 0x12, 0x3c, 0x0a, 0x0a, 0x70, 0x72, 0x65, 0x70, 0x72, 0x6f,
	0x63, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x6d, 0x65, 0x72,
	0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e,
	0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x0a, 0x70, 0x72, 0x65, 0x70, 0x72, 0x6f,
	0x63, 0x65, 0x73, 0x73, 0x12, 0x3e, 0x0a, 0x0b, 0x70, 0x6f, 0x73, 0x74, 0x70, 0x72, 0x6f, 0x63,
	0x65, 0x73, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x6d, 0x65, 0x72, 0x6c,
	0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x50,
	0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x0b, 0x70, 0x6f, 0x73, 0x74, 0x70, 0x72, 0x6f,
	0x63, 0x65, 0x73, 0x73, 0x22, 0xc1, 0x01, 0x0a, 0x08, 0x50, 0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e,
	0x65, 0x12, 0x31, 0x0a, 0x06, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x19, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73,
	0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x52, 0x06, 0x69, 0x6e,
	0x70, 0x75, 0x74, 0x73, 0x12, 0x4c, 0x0a, 0x0f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72,
	0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e,
	0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d,
	0x65, 0x72, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x0f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x12, 0x34, 0x0a, 0x07, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x73, 0x18, 0x03, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x52,
	0x07, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x73, 0x22, 0xa4, 0x02, 0x0a, 0x05, 0x49, 0x6e, 0x70,
	0x75, 0x74, 0x12, 0x3a, 0x0a, 0x09, 0x76, 0x61, 0x72, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x56, 0x61, 0x72, 0x69, 0x61,
	0x62, 0x6c, 0x65, 0x52, 0x09, 0x76, 0x61, 0x72, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x73, 0x12, 0x36,
	0x0a, 0x05, 0x66, 0x65, 0x61, 0x73, 0x74, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e,
	0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d,
	0x65, 0x72, 0x2e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x52,
	0x05, 0x66, 0x65, 0x61, 0x73, 0x74, 0x12, 0x31, 0x0a, 0x06, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x73,
	0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e,
	0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x54, 0x61, 0x62, 0x6c,
	0x65, 0x52, 0x06, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x73, 0x12, 0x37, 0x0a, 0x08, 0x65, 0x6e, 0x63,
	0x6f, 0x64, 0x65, 0x72, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x6d, 0x65,
	0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72,
	0x2e, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x65, 0x72, 0x52, 0x08, 0x65, 0x6e, 0x63, 0x6f, 0x64, 0x65,
	0x72, 0x73, 0x12, 0x3b, 0x0a, 0x08, 0x61, 0x75, 0x74, 0x6f, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72,
	0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x55, 0x50, 0x49, 0x41, 0x75, 0x74,
	0x6f, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x08, 0x61, 0x75, 0x74, 0x6f, 0x6c, 0x6f, 0x61, 0x64, 0x22,
	0xe4, 0x01, 0x0a, 0x0e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x3b, 0x0a, 0x09, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x4a, 0x6f, 0x69, 0x6e, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x54, 0x61, 0x62, 0x6c, 0x65,
	0x4a, 0x6f, 0x69, 0x6e, 0x52, 0x09, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x4a, 0x6f, 0x69, 0x6e, 0x12,
	0x59, 0x0a, 0x13, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72,
	0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x6d,
	0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65,
	0x72, 0x2e, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x13, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x54, 0x72, 0x61, 0x6e,
	0x73, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x3a, 0x0a, 0x09, 0x76, 0x61,
	0x72, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e,
	0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d,
	0x65, 0x72, 0x2e, 0x56, 0x61, 0x72, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x52, 0x09, 0x76, 0x61, 0x72,
	0x69, 0x61, 0x62, 0x6c, 0x65, 0x73, 0x22, 0x81, 0x02, 0x0a, 0x06, 0x4f, 0x75, 0x74, 0x70, 0x75,
	0x74, 0x12, 0x3e, 0x0a, 0x0a, 0x6a, 0x73, 0x6f, 0x6e, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x4a, 0x73, 0x6f, 0x6e, 0x4f,
	0x75, 0x74, 0x70, 0x75, 0x74, 0x52, 0x0a, 0x6a, 0x73, 0x6f, 0x6e, 0x4f, 0x75, 0x74, 0x70, 0x75,
	0x74, 0x12, 0x59, 0x0a, 0x13, 0x75, 0x70, 0x69, 0x50, 0x72, 0x65, 0x70, 0x72, 0x6f, 0x63, 0x65,
	0x73, 0x73, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27,
	0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72,
	0x6d, 0x65, 0x72, 0x2e, 0x55, 0x50, 0x49, 0x50, 0x72, 0x65, 0x70, 0x72, 0x6f, 0x63, 0x65, 0x73,
	0x73, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x52, 0x13, 0x75, 0x70, 0x69, 0x50, 0x72, 0x65, 0x70,
	0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x5c, 0x0a, 0x14,
	0x75, 0x70, 0x69, 0x50, 0x6f, 0x73, 0x74, 0x70, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x4f, 0x75,
	0x74, 0x70, 0x75, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x28, 0x2e, 0x6d, 0x65, 0x72,
	0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e,
	0x55, 0x50, 0x49, 0x50, 0x6f, 0x73, 0x74, 0x70, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x4f, 0x75,
	0x74, 0x70, 0x75, 0x74, 0x52, 0x14, 0x75, 0x70, 0x69, 0x50, 0x6f, 0x73, 0x74, 0x70, 0x72, 0x6f,
	0x63, 0x65, 0x73, 0x73, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x42, 0x33, 0x5a, 0x31, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x61, 0x72, 0x61, 0x6d, 0x6c, 0x2d,
	0x64, 0x65, 0x76, 0x2f, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_transformer_spec_standard_transformer_proto_rawDescOnce sync.Once
	file_transformer_spec_standard_transformer_proto_rawDescData = file_transformer_spec_standard_transformer_proto_rawDesc
)

func file_transformer_spec_standard_transformer_proto_rawDescGZIP() []byte {
	file_transformer_spec_standard_transformer_proto_rawDescOnce.Do(func() {
		file_transformer_spec_standard_transformer_proto_rawDescData = protoimpl.X.CompressGZIP(file_transformer_spec_standard_transformer_proto_rawDescData)
	})
	return file_transformer_spec_standard_transformer_proto_rawDescData
}

var file_transformer_spec_standard_transformer_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_transformer_spec_standard_transformer_proto_goTypes = []interface{}{
	(*StandardTransformerConfig)(nil), // 0: merlin.transformer.StandardTransformerConfig
	(*TransformerConfig)(nil),         // 1: merlin.transformer.TransformerConfig
	(*Pipeline)(nil),                  // 2: merlin.transformer.Pipeline
	(*Input)(nil),                     // 3: merlin.transformer.Input
	(*Transformation)(nil),            // 4: merlin.transformer.Transformation
	(*Output)(nil),                    // 5: merlin.transformer.Output
	(*PredictionLogConfig)(nil),       // 6: merlin.transformer.PredictionLogConfig
	(*FeatureTable)(nil),              // 7: merlin.transformer.FeatureTable
	(*Variable)(nil),                  // 8: merlin.transformer.Variable
	(*Table)(nil),                     // 9: merlin.transformer.Table
	(*Encoder)(nil),                   // 10: merlin.transformer.Encoder
	(*UPIAutoload)(nil),               // 11: merlin.transformer.UPIAutoload
	(*TableJoin)(nil),                 // 12: merlin.transformer.TableJoin
	(*TableTransformation)(nil),       // 13: merlin.transformer.TableTransformation
	(*JsonOutput)(nil),                // 14: merlin.transformer.JsonOutput
	(*UPIPreprocessOutput)(nil),       // 15: merlin.transformer.UPIPreprocessOutput
	(*UPIPostprocessOutput)(nil),      // 16: merlin.transformer.UPIPostprocessOutput
}
var file_transformer_spec_standard_transformer_proto_depIdxs = []int32{
	1,  // 0: merlin.transformer.StandardTransformerConfig.transformerConfig:type_name -> merlin.transformer.TransformerConfig
	6,  // 1: merlin.transformer.StandardTransformerConfig.predictionLogConfig:type_name -> merlin.transformer.PredictionLogConfig
	7,  // 2: merlin.transformer.TransformerConfig.feast:type_name -> merlin.transformer.FeatureTable
	2,  // 3: merlin.transformer.TransformerConfig.preprocess:type_name -> merlin.transformer.Pipeline
	2,  // 4: merlin.transformer.TransformerConfig.postprocess:type_name -> merlin.transformer.Pipeline
	3,  // 5: merlin.transformer.Pipeline.inputs:type_name -> merlin.transformer.Input
	4,  // 6: merlin.transformer.Pipeline.transformations:type_name -> merlin.transformer.Transformation
	5,  // 7: merlin.transformer.Pipeline.outputs:type_name -> merlin.transformer.Output
	8,  // 8: merlin.transformer.Input.variables:type_name -> merlin.transformer.Variable
	7,  // 9: merlin.transformer.Input.feast:type_name -> merlin.transformer.FeatureTable
	9,  // 10: merlin.transformer.Input.tables:type_name -> merlin.transformer.Table
	10, // 11: merlin.transformer.Input.encoders:type_name -> merlin.transformer.Encoder
	11, // 12: merlin.transformer.Input.autoload:type_name -> merlin.transformer.UPIAutoload
	12, // 13: merlin.transformer.Transformation.tableJoin:type_name -> merlin.transformer.TableJoin
	13, // 14: merlin.transformer.Transformation.tableTransformation:type_name -> merlin.transformer.TableTransformation
	8,  // 15: merlin.transformer.Transformation.variables:type_name -> merlin.transformer.Variable
	14, // 16: merlin.transformer.Output.jsonOutput:type_name -> merlin.transformer.JsonOutput
	15, // 17: merlin.transformer.Output.upiPreprocessOutput:type_name -> merlin.transformer.UPIPreprocessOutput
	16, // 18: merlin.transformer.Output.upiPostprocessOutput:type_name -> merlin.transformer.UPIPostprocessOutput
	19, // [19:19] is the sub-list for method output_type
	19, // [19:19] is the sub-list for method input_type
	19, // [19:19] is the sub-list for extension type_name
	19, // [19:19] is the sub-list for extension extendee
	0,  // [0:19] is the sub-list for field type_name
}

func init() { file_transformer_spec_standard_transformer_proto_init() }
func file_transformer_spec_standard_transformer_proto_init() {
	if File_transformer_spec_standard_transformer_proto != nil {
		return
	}
	file_transformer_spec_feast_proto_init()
	file_transformer_spec_table_proto_init()
	file_transformer_spec_variable_proto_init()
	file_transformer_spec_json_proto_init()
	file_transformer_spec_encoder_proto_init()
	file_transformer_spec_upi_output_proto_init()
	file_transformer_spec_upi_autoload_proto_init()
	file_transformer_spec_prediction_log_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_transformer_spec_standard_transformer_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StandardTransformerConfig); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_transformer_spec_standard_transformer_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TransformerConfig); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_transformer_spec_standard_transformer_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Pipeline); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_transformer_spec_standard_transformer_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Input); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_transformer_spec_standard_transformer_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Transformation); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_transformer_spec_standard_transformer_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Output); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_transformer_spec_standard_transformer_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transformer_spec_standard_transformer_proto_goTypes,
		DependencyIndexes: file_transformer_spec_standard_transformer_proto_depIdxs,
		MessageInfos:      file_transformer_spec_standard_transformer_proto_msgTypes,
	}.Build()
	File_transformer_spec_standard_transformer_proto = out.File
	file_transformer_spec_standard_transformer_proto_rawDesc = nil
	file_transformer_spec_standard_transformer_proto_goTypes = nil
	file_transformer_spec_standard_transformer_proto_depIdxs = nil
}
