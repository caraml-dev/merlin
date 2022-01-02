// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.1
// source: transformer/spec/encoder.proto

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

type PeriodType int32

const (
	PeriodType_UNDEFINED PeriodType = 0 //in case not field not defined in config
	PeriodType_HOUR      PeriodType = 1
	PeriodType_DAY       PeriodType = 2
	PeriodType_WEEK      PeriodType = 3
	PeriodType_MONTH     PeriodType = 4
	PeriodType_QUARTER   PeriodType = 5
	PeriodType_HALF      PeriodType = 6
	PeriodType_YEAR      PeriodType = 7
)

// Enum value maps for PeriodType.
var (
	PeriodType_name = map[int32]string{
		0: "UNDEFINED",
		1: "HOUR",
		2: "DAY",
		3: "WEEK",
		4: "MONTH",
		5: "QUARTER",
		6: "HALF",
		7: "YEAR",
	}
	PeriodType_value = map[string]int32{
		"UNDEFINED": 0,
		"HOUR":      1,
		"DAY":       2,
		"WEEK":      3,
		"MONTH":     4,
		"QUARTER":   5,
		"HALF":      6,
		"YEAR":      7,
	}
)

func (x PeriodType) Enum() *PeriodType {
	p := new(PeriodType)
	*p = x
	return p
}

func (x PeriodType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PeriodType) Descriptor() protoreflect.EnumDescriptor {
	return file_transformer_spec_encoder_proto_enumTypes[0].Descriptor()
}

func (PeriodType) Type() protoreflect.EnumType {
	return &file_transformer_spec_encoder_proto_enumTypes[0]
}

func (x PeriodType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PeriodType.Descriptor instead.
func (PeriodType) EnumDescriptor() ([]byte, []int) {
	return file_transformer_spec_encoder_proto_rawDescGZIP(), []int{0}
}

type ValueType int32

const (
	ValueType_STRING ValueType = 0
	ValueType_INT    ValueType = 1
	ValueType_FLOAT  ValueType = 2
	ValueType_BOOL   ValueType = 3
)

// Enum value maps for ValueType.
var (
	ValueType_name = map[int32]string{
		0: "STRING",
		1: "INT",
		2: "FLOAT",
		3: "BOOL",
	}
	ValueType_value = map[string]int32{
		"STRING": 0,
		"INT":    1,
		"FLOAT":  2,
		"BOOL":   3,
	}
)

func (x ValueType) Enum() *ValueType {
	p := new(ValueType)
	*p = x
	return p
}

func (x ValueType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ValueType) Descriptor() protoreflect.EnumDescriptor {
	return file_transformer_spec_encoder_proto_enumTypes[1].Descriptor()
}

func (ValueType) Type() protoreflect.EnumType {
	return &file_transformer_spec_encoder_proto_enumTypes[1]
}

func (x ValueType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ValueType.Descriptor instead.
func (ValueType) EnumDescriptor() ([]byte, []int) {
	return file_transformer_spec_encoder_proto_rawDescGZIP(), []int{1}
}

type Encoder struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Types that are assignable to EncoderConfig:
	//	*Encoder_OrdinalEncoderConfig
	//	*Encoder_CyclicalEncoderConfig
	EncoderConfig isEncoder_EncoderConfig `protobuf_oneof:"encoderConfig"`
}

func (x *Encoder) Reset() {
	*x = Encoder{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_spec_encoder_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Encoder) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Encoder) ProtoMessage() {}

func (x *Encoder) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_spec_encoder_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Encoder.ProtoReflect.Descriptor instead.
func (*Encoder) Descriptor() ([]byte, []int) {
	return file_transformer_spec_encoder_proto_rawDescGZIP(), []int{0}
}

func (x *Encoder) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (m *Encoder) GetEncoderConfig() isEncoder_EncoderConfig {
	if m != nil {
		return m.EncoderConfig
	}
	return nil
}

func (x *Encoder) GetOrdinalEncoderConfig() *OrdinalEncoderConfig {
	if x, ok := x.GetEncoderConfig().(*Encoder_OrdinalEncoderConfig); ok {
		return x.OrdinalEncoderConfig
	}
	return nil
}

func (x *Encoder) GetCyclicalEncoderConfig() *CyclicalEncoderConfig {
	if x, ok := x.GetEncoderConfig().(*Encoder_CyclicalEncoderConfig); ok {
		return x.CyclicalEncoderConfig
	}
	return nil
}

type isEncoder_EncoderConfig interface {
	isEncoder_EncoderConfig()
}

type Encoder_OrdinalEncoderConfig struct {
	OrdinalEncoderConfig *OrdinalEncoderConfig `protobuf:"bytes,2,opt,name=ordinalEncoderConfig,proto3,oneof"`
}

type Encoder_CyclicalEncoderConfig struct {
	CyclicalEncoderConfig *CyclicalEncoderConfig `protobuf:"bytes,3,opt,name=cyclicalEncoderConfig,proto3,oneof"`
}

func (*Encoder_OrdinalEncoderConfig) isEncoder_EncoderConfig() {}

func (*Encoder_CyclicalEncoderConfig) isEncoder_EncoderConfig() {}

type OrdinalEncoderConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DefaultValue    string            `protobuf:"bytes,1,opt,name=defaultValue,proto3" json:"defaultValue,omitempty"`
	TargetValueType ValueType         `protobuf:"varint,2,opt,name=targetValueType,proto3,enum=merlin.transformer.ValueType" json:"targetValueType,omitempty"`
	Mapping         map[string]string `protobuf:"bytes,3,rep,name=mapping,proto3" json:"mapping,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *OrdinalEncoderConfig) Reset() {
	*x = OrdinalEncoderConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_spec_encoder_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OrdinalEncoderConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OrdinalEncoderConfig) ProtoMessage() {}

func (x *OrdinalEncoderConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_spec_encoder_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OrdinalEncoderConfig.ProtoReflect.Descriptor instead.
func (*OrdinalEncoderConfig) Descriptor() ([]byte, []int) {
	return file_transformer_spec_encoder_proto_rawDescGZIP(), []int{1}
}

func (x *OrdinalEncoderConfig) GetDefaultValue() string {
	if x != nil {
		return x.DefaultValue
	}
	return ""
}

func (x *OrdinalEncoderConfig) GetTargetValueType() ValueType {
	if x != nil {
		return x.TargetValueType
	}
	return ValueType_STRING
}

func (x *OrdinalEncoderConfig) GetMapping() map[string]string {
	if x != nil {
		return x.Mapping
	}
	return nil
}

type CyclicalEncoderConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to EncodeBy:
	//	*CyclicalEncoderConfig_ByEpochTime
	//	*CyclicalEncoderConfig_ByRange
	EncodeBy isCyclicalEncoderConfig_EncodeBy `protobuf_oneof:"encodeBy"`
}

func (x *CyclicalEncoderConfig) Reset() {
	*x = CyclicalEncoderConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_spec_encoder_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CyclicalEncoderConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CyclicalEncoderConfig) ProtoMessage() {}

func (x *CyclicalEncoderConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_spec_encoder_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CyclicalEncoderConfig.ProtoReflect.Descriptor instead.
func (*CyclicalEncoderConfig) Descriptor() ([]byte, []int) {
	return file_transformer_spec_encoder_proto_rawDescGZIP(), []int{2}
}

func (m *CyclicalEncoderConfig) GetEncodeBy() isCyclicalEncoderConfig_EncodeBy {
	if m != nil {
		return m.EncodeBy
	}
	return nil
}

func (x *CyclicalEncoderConfig) GetByEpochTime() *ByEpochTime {
	if x, ok := x.GetEncodeBy().(*CyclicalEncoderConfig_ByEpochTime); ok {
		return x.ByEpochTime
	}
	return nil
}

func (x *CyclicalEncoderConfig) GetByRange() *ByRange {
	if x, ok := x.GetEncodeBy().(*CyclicalEncoderConfig_ByRange); ok {
		return x.ByRange
	}
	return nil
}

type isCyclicalEncoderConfig_EncodeBy interface {
	isCyclicalEncoderConfig_EncodeBy()
}

type CyclicalEncoderConfig_ByEpochTime struct {
	ByEpochTime *ByEpochTime `protobuf:"bytes,1,opt,name=byEpochTime,proto3,oneof"`
}

type CyclicalEncoderConfig_ByRange struct {
	ByRange *ByRange `protobuf:"bytes,2,opt,name=byRange,proto3,oneof"`
}

func (*CyclicalEncoderConfig_ByEpochTime) isCyclicalEncoderConfig_EncodeBy() {}

func (*CyclicalEncoderConfig_ByRange) isCyclicalEncoderConfig_EncodeBy() {}

type ByEpochTime struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Period PeriodType `protobuf:"varint,1,opt,name=period,proto3,enum=merlin.transformer.PeriodType" json:"period,omitempty"`
}

func (x *ByEpochTime) Reset() {
	*x = ByEpochTime{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_spec_encoder_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ByEpochTime) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ByEpochTime) ProtoMessage() {}

func (x *ByEpochTime) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_spec_encoder_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ByEpochTime.ProtoReflect.Descriptor instead.
func (*ByEpochTime) Descriptor() ([]byte, []int) {
	return file_transformer_spec_encoder_proto_rawDescGZIP(), []int{3}
}

func (x *ByEpochTime) GetPeriod() PeriodType {
	if x != nil {
		return x.Period
	}
	return PeriodType_UNDEFINED
}

type ByRange struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Min float64 `protobuf:"fixed64,1,opt,name=min,proto3" json:"min,omitempty"`
	Max float64 `protobuf:"fixed64,2,opt,name=max,proto3" json:"max,omitempty"`
}

func (x *ByRange) Reset() {
	*x = ByRange{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_spec_encoder_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ByRange) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ByRange) ProtoMessage() {}

func (x *ByRange) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_spec_encoder_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ByRange.ProtoReflect.Descriptor instead.
func (*ByRange) Descriptor() ([]byte, []int) {
	return file_transformer_spec_encoder_proto_rawDescGZIP(), []int{4}
}

func (x *ByRange) GetMin() float64 {
	if x != nil {
		return x.Min
	}
	return 0
}

func (x *ByRange) GetMax() float64 {
	if x != nil {
		return x.Max
	}
	return 0
}

var File_transformer_spec_encoder_proto protoreflect.FileDescriptor

var file_transformer_spec_encoder_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2f, 0x73, 0x70,
	0x65, 0x63, 0x2f, 0x65, 0x6e, 0x63, 0x6f, 0x64, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x12, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f,
	0x72, 0x6d, 0x65, 0x72, 0x22, 0xf1, 0x01, 0x0a, 0x07, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x65, 0x72,
	0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x5e, 0x0a, 0x14, 0x6f, 0x72, 0x64, 0x69, 0x6e, 0x61, 0x6c, 0x45,
	0x6e, 0x63, 0x6f, 0x64, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x28, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e,
	0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x4f, 0x72, 0x64, 0x69, 0x6e, 0x61, 0x6c, 0x45,
	0x6e, 0x63, 0x6f, 0x64, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x48, 0x00, 0x52, 0x14,
	0x6f, 0x72, 0x64, 0x69, 0x6e, 0x61, 0x6c, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x65, 0x72, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x12, 0x61, 0x0a, 0x15, 0x63, 0x79, 0x63, 0x6c, 0x69, 0x63, 0x61, 0x6c,
	0x45, 0x6e, 0x63, 0x6f, 0x64, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x43, 0x79, 0x63, 0x6c, 0x69, 0x63, 0x61,
	0x6c, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x48, 0x00,
	0x52, 0x15, 0x63, 0x79, 0x63, 0x6c, 0x69, 0x63, 0x61, 0x6c, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x65,
	0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x42, 0x0f, 0x0a, 0x0d, 0x65, 0x6e, 0x63, 0x6f, 0x64,
	0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0x90, 0x02, 0x0a, 0x14, 0x4f, 0x72, 0x64,
	0x69, 0x6e, 0x61, 0x6c, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x12, 0x22, 0x0a, 0x0c, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x47, 0x0a, 0x0f, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x54, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1d,
	0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72,
	0x6d, 0x65, 0x72, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0f, 0x74,
	0x61, 0x72, 0x67, 0x65, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x4f,
	0x0a, 0x07, 0x6d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x35, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f,
	0x72, 0x6d, 0x65, 0x72, 0x2e, 0x4f, 0x72, 0x64, 0x69, 0x6e, 0x61, 0x6c, 0x45, 0x6e, 0x63, 0x6f,
	0x64, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x4d, 0x61, 0x70, 0x70, 0x69, 0x6e,
	0x67, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x6d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x1a,
	0x3a, 0x0a, 0x0c, 0x4d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xa1, 0x01, 0x0a, 0x15,
	0x43, 0x79, 0x63, 0x6c, 0x69, 0x63, 0x61, 0x6c, 0x45, 0x6e, 0x63, 0x6f, 0x64, 0x65, 0x72, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x43, 0x0a, 0x0b, 0x62, 0x79, 0x45, 0x70, 0x6f, 0x63, 0x68,
	0x54, 0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x6d, 0x65, 0x72,
	0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e,
	0x42, 0x79, 0x45, 0x70, 0x6f, 0x63, 0x68, 0x54, 0x69, 0x6d, 0x65, 0x48, 0x00, 0x52, 0x0b, 0x62,
	0x79, 0x45, 0x70, 0x6f, 0x63, 0x68, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x37, 0x0a, 0x07, 0x62, 0x79,
	0x52, 0x61, 0x6e, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x6d, 0x65,
	0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72,
	0x2e, 0x42, 0x79, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x48, 0x00, 0x52, 0x07, 0x62, 0x79, 0x52, 0x61,
	0x6e, 0x67, 0x65, 0x42, 0x0a, 0x0a, 0x08, 0x65, 0x6e, 0x63, 0x6f, 0x64, 0x65, 0x42, 0x79, 0x22,
	0x45, 0x0a, 0x0b, 0x42, 0x79, 0x45, 0x70, 0x6f, 0x63, 0x68, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x36,
	0x0a, 0x06, 0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1e,
	0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72,
	0x6d, 0x65, 0x72, 0x2e, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x54, 0x79, 0x70, 0x65, 0x52, 0x06,
	0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x22, 0x2d, 0x0a, 0x07, 0x42, 0x79, 0x52, 0x61, 0x6e, 0x67,
	0x65, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x01, 0x52, 0x03,
	0x6d, 0x69, 0x6e, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x61, 0x78, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01,
	0x52, 0x03, 0x6d, 0x61, 0x78, 0x2a, 0x64, 0x0a, 0x0a, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x54,
	0x79, 0x70, 0x65, 0x12, 0x0d, 0x0a, 0x09, 0x55, 0x4e, 0x44, 0x45, 0x46, 0x49, 0x4e, 0x45, 0x44,
	0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x48, 0x4f, 0x55, 0x52, 0x10, 0x01, 0x12, 0x07, 0x0a, 0x03,
	0x44, 0x41, 0x59, 0x10, 0x02, 0x12, 0x08, 0x0a, 0x04, 0x57, 0x45, 0x45, 0x4b, 0x10, 0x03, 0x12,
	0x09, 0x0a, 0x05, 0x4d, 0x4f, 0x4e, 0x54, 0x48, 0x10, 0x04, 0x12, 0x0b, 0x0a, 0x07, 0x51, 0x55,
	0x41, 0x52, 0x54, 0x45, 0x52, 0x10, 0x05, 0x12, 0x08, 0x0a, 0x04, 0x48, 0x41, 0x4c, 0x46, 0x10,
	0x06, 0x12, 0x08, 0x0a, 0x04, 0x59, 0x45, 0x41, 0x52, 0x10, 0x07, 0x2a, 0x35, 0x0a, 0x09, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0a, 0x0a, 0x06, 0x53, 0x54, 0x52, 0x49,
	0x4e, 0x47, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03, 0x49, 0x4e, 0x54, 0x10, 0x01, 0x12, 0x09, 0x0a,
	0x05, 0x46, 0x4c, 0x4f, 0x41, 0x54, 0x10, 0x02, 0x12, 0x08, 0x0a, 0x04, 0x42, 0x4f, 0x4f, 0x4c,
	0x10, 0x03, 0x42, 0x2e, 0x5a, 0x2c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x67, 0x6f, 0x6a, 0x65, 0x6b, 0x2f, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2f, 0x70, 0x6b,
	0x67, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2f, 0x73, 0x70,
	0x65, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_transformer_spec_encoder_proto_rawDescOnce sync.Once
	file_transformer_spec_encoder_proto_rawDescData = file_transformer_spec_encoder_proto_rawDesc
)

func file_transformer_spec_encoder_proto_rawDescGZIP() []byte {
	file_transformer_spec_encoder_proto_rawDescOnce.Do(func() {
		file_transformer_spec_encoder_proto_rawDescData = protoimpl.X.CompressGZIP(file_transformer_spec_encoder_proto_rawDescData)
	})
	return file_transformer_spec_encoder_proto_rawDescData
}

var file_transformer_spec_encoder_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_transformer_spec_encoder_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_transformer_spec_encoder_proto_goTypes = []interface{}{
	(PeriodType)(0),               // 0: merlin.transformer.PeriodType
	(ValueType)(0),                // 1: merlin.transformer.ValueType
	(*Encoder)(nil),               // 2: merlin.transformer.Encoder
	(*OrdinalEncoderConfig)(nil),  // 3: merlin.transformer.OrdinalEncoderConfig
	(*CyclicalEncoderConfig)(nil), // 4: merlin.transformer.CyclicalEncoderConfig
	(*ByEpochTime)(nil),           // 5: merlin.transformer.ByEpochTime
	(*ByRange)(nil),               // 6: merlin.transformer.ByRange
	nil,                           // 7: merlin.transformer.OrdinalEncoderConfig.MappingEntry
}
var file_transformer_spec_encoder_proto_depIdxs = []int32{
	3, // 0: merlin.transformer.Encoder.ordinalEncoderConfig:type_name -> merlin.transformer.OrdinalEncoderConfig
	4, // 1: merlin.transformer.Encoder.cyclicalEncoderConfig:type_name -> merlin.transformer.CyclicalEncoderConfig
	1, // 2: merlin.transformer.OrdinalEncoderConfig.targetValueType:type_name -> merlin.transformer.ValueType
	7, // 3: merlin.transformer.OrdinalEncoderConfig.mapping:type_name -> merlin.transformer.OrdinalEncoderConfig.MappingEntry
	5, // 4: merlin.transformer.CyclicalEncoderConfig.byEpochTime:type_name -> merlin.transformer.ByEpochTime
	6, // 5: merlin.transformer.CyclicalEncoderConfig.byRange:type_name -> merlin.transformer.ByRange
	0, // 6: merlin.transformer.ByEpochTime.period:type_name -> merlin.transformer.PeriodType
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_transformer_spec_encoder_proto_init() }
func file_transformer_spec_encoder_proto_init() {
	if File_transformer_spec_encoder_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_transformer_spec_encoder_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Encoder); i {
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
		file_transformer_spec_encoder_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OrdinalEncoderConfig); i {
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
		file_transformer_spec_encoder_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CyclicalEncoderConfig); i {
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
		file_transformer_spec_encoder_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ByEpochTime); i {
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
		file_transformer_spec_encoder_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ByRange); i {
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
	file_transformer_spec_encoder_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Encoder_OrdinalEncoderConfig)(nil),
		(*Encoder_CyclicalEncoderConfig)(nil),
	}
	file_transformer_spec_encoder_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*CyclicalEncoderConfig_ByEpochTime)(nil),
		(*CyclicalEncoderConfig_ByRange)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_transformer_spec_encoder_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transformer_spec_encoder_proto_goTypes,
		DependencyIndexes: file_transformer_spec_encoder_proto_depIdxs,
		EnumInfos:         file_transformer_spec_encoder_proto_enumTypes,
		MessageInfos:      file_transformer_spec_encoder_proto_msgTypes,
	}.Build()
	File_transformer_spec_encoder_proto = out.File
	file_transformer_spec_encoder_proto_rawDesc = nil
	file_transformer_spec_encoder_proto_goTypes = nil
	file_transformer_spec_encoder_proto_depIdxs = nil
}
