// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.12.3
// source: transformer/spec/variable.proto

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

type Variable struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Types that are assignable to Value:
	//	*Variable_Literal
	//	*Variable_Expression
	//	*Variable_JsonPath
	Value isVariable_Value `protobuf_oneof:"value"`
}

func (x *Variable) Reset() {
	*x = Variable{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_spec_variable_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Variable) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Variable) ProtoMessage() {}

func (x *Variable) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_spec_variable_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Variable.ProtoReflect.Descriptor instead.
func (*Variable) Descriptor() ([]byte, []int) {
	return file_transformer_spec_variable_proto_rawDescGZIP(), []int{0}
}

func (x *Variable) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (m *Variable) GetValue() isVariable_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (x *Variable) GetLiteral() *Literal {
	if x, ok := x.GetValue().(*Variable_Literal); ok {
		return x.Literal
	}
	return nil
}

func (x *Variable) GetExpression() string {
	if x, ok := x.GetValue().(*Variable_Expression); ok {
		return x.Expression
	}
	return ""
}

func (x *Variable) GetJsonPath() string {
	if x, ok := x.GetValue().(*Variable_JsonPath); ok {
		return x.JsonPath
	}
	return ""
}

type isVariable_Value interface {
	isVariable_Value()
}

type Variable_Literal struct {
	Literal *Literal `protobuf:"bytes,2,opt,name=literal,proto3,oneof"`
}

type Variable_Expression struct {
	Expression string `protobuf:"bytes,3,opt,name=expression,proto3,oneof"`
}

type Variable_JsonPath struct {
	JsonPath string `protobuf:"bytes,4,opt,name=jsonPath,proto3,oneof"`
}

func (*Variable_Literal) isVariable_Value() {}

func (*Variable_Expression) isVariable_Value() {}

func (*Variable_JsonPath) isVariable_Value() {}

type Literal struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to LiteralValue:
	//	*Literal_StringValue
	//	*Literal_IntValue
	//	*Literal_FloatValue
	LiteralValue isLiteral_LiteralValue `protobuf_oneof:"literal_value"`
}

func (x *Literal) Reset() {
	*x = Literal{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_spec_variable_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Literal) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Literal) ProtoMessage() {}

func (x *Literal) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_spec_variable_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Literal.ProtoReflect.Descriptor instead.
func (*Literal) Descriptor() ([]byte, []int) {
	return file_transformer_spec_variable_proto_rawDescGZIP(), []int{1}
}

func (m *Literal) GetLiteralValue() isLiteral_LiteralValue {
	if m != nil {
		return m.LiteralValue
	}
	return nil
}

func (x *Literal) GetStringValue() string {
	if x, ok := x.GetLiteralValue().(*Literal_StringValue); ok {
		return x.StringValue
	}
	return ""
}

func (x *Literal) GetIntValue() int64 {
	if x, ok := x.GetLiteralValue().(*Literal_IntValue); ok {
		return x.IntValue
	}
	return 0
}

func (x *Literal) GetFloatValue() float64 {
	if x, ok := x.GetLiteralValue().(*Literal_FloatValue); ok {
		return x.FloatValue
	}
	return 0
}

type isLiteral_LiteralValue interface {
	isLiteral_LiteralValue()
}

type Literal_StringValue struct {
	StringValue string `protobuf:"bytes,1,opt,name=stringValue,proto3,oneof"`
}

type Literal_IntValue struct {
	IntValue int64 `protobuf:"varint,2,opt,name=intValue,proto3,oneof"`
}

type Literal_FloatValue struct {
	FloatValue float64 `protobuf:"fixed64,3,opt,name=floatValue,proto3,oneof"`
}

func (*Literal_StringValue) isLiteral_LiteralValue() {}

func (*Literal_IntValue) isLiteral_LiteralValue() {}

func (*Literal_FloatValue) isLiteral_LiteralValue() {}

var File_transformer_spec_variable_proto protoreflect.FileDescriptor

var file_transformer_spec_variable_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2f, 0x73, 0x70,
	0x65, 0x63, 0x2f, 0x76, 0x61, 0x72, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x12, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66,
	0x6f, 0x72, 0x6d, 0x65, 0x72, 0x22, 0xa0, 0x01, 0x0a, 0x08, 0x56, 0x61, 0x72, 0x69, 0x61, 0x62,
	0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x37, 0x0a, 0x07, 0x6c, 0x69, 0x74, 0x65, 0x72, 0x61,
	0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e,
	0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x4c, 0x69, 0x74,
	0x65, 0x72, 0x61, 0x6c, 0x48, 0x00, 0x52, 0x07, 0x6c, 0x69, 0x74, 0x65, 0x72, 0x61, 0x6c, 0x12,
	0x20, 0x0a, 0x0a, 0x65, 0x78, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x0a, 0x65, 0x78, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f,
	0x6e, 0x12, 0x1c, 0x0a, 0x08, 0x6a, 0x73, 0x6f, 0x6e, 0x50, 0x61, 0x74, 0x68, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x08, 0x6a, 0x73, 0x6f, 0x6e, 0x50, 0x61, 0x74, 0x68, 0x42,
	0x07, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x7e, 0x0a, 0x07, 0x4c, 0x69, 0x74, 0x65,
	0x72, 0x61, 0x6c, 0x12, 0x22, 0x0a, 0x0b, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x0b, 0x73, 0x74, 0x72, 0x69,
	0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x1c, 0x0a, 0x08, 0x69, 0x6e, 0x74, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x48, 0x00, 0x52, 0x08, 0x69, 0x6e, 0x74,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x20, 0x0a, 0x0a, 0x66, 0x6c, 0x6f, 0x61, 0x74, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x01, 0x48, 0x00, 0x52, 0x0a, 0x66, 0x6c, 0x6f,
	0x61, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x42, 0x0f, 0x0a, 0x0d, 0x6c, 0x69, 0x74, 0x65, 0x72,
	0x61, 0x6c, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x42, 0x2e, 0x5a, 0x2c, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x6f, 0x6a, 0x65, 0x6b, 0x2f, 0x6d, 0x65, 0x72,
	0x6c, 0x69, 0x6e, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72,
	0x6d, 0x65, 0x72, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_transformer_spec_variable_proto_rawDescOnce sync.Once
	file_transformer_spec_variable_proto_rawDescData = file_transformer_spec_variable_proto_rawDesc
)

func file_transformer_spec_variable_proto_rawDescGZIP() []byte {
	file_transformer_spec_variable_proto_rawDescOnce.Do(func() {
		file_transformer_spec_variable_proto_rawDescData = protoimpl.X.CompressGZIP(file_transformer_spec_variable_proto_rawDescData)
	})
	return file_transformer_spec_variable_proto_rawDescData
}

var file_transformer_spec_variable_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_transformer_spec_variable_proto_goTypes = []interface{}{
	(*Variable)(nil), // 0: merlin.transformer.Variable
	(*Literal)(nil),  // 1: merlin.transformer.Literal
}
var file_transformer_spec_variable_proto_depIdxs = []int32{
	1, // 0: merlin.transformer.Variable.literal:type_name -> merlin.transformer.Literal
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_transformer_spec_variable_proto_init() }
func file_transformer_spec_variable_proto_init() {
	if File_transformer_spec_variable_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_transformer_spec_variable_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Variable); i {
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
		file_transformer_spec_variable_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Literal); i {
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
	file_transformer_spec_variable_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Variable_Literal)(nil),
		(*Variable_Expression)(nil),
		(*Variable_JsonPath)(nil),
	}
	file_transformer_spec_variable_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*Literal_StringValue)(nil),
		(*Literal_IntValue)(nil),
		(*Literal_FloatValue)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_transformer_spec_variable_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transformer_spec_variable_proto_goTypes,
		DependencyIndexes: file_transformer_spec_variable_proto_depIdxs,
		MessageInfos:      file_transformer_spec_variable_proto_msgTypes,
	}.Build()
	File_transformer_spec_variable_proto = out.File
	file_transformer_spec_variable_proto_rawDesc = nil
	file_transformer_spec_variable_proto_goTypes = nil
	file_transformer_spec_variable_proto_depIdxs = nil
}
