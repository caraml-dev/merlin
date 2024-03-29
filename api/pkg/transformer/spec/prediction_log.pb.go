// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.21.9
// source: transformer/spec/prediction_log.proto

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

// PredictionLogConfig contains information about prediction log
type PredictionLogConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// flag to enable the prediction log
	Enable bool `protobuf:"varint,1,opt,name=enable,proto3" json:"enable,omitempty"`
	// name of table that will be used to populate `rawFeaturesTable` field in prediction log
	RawFeaturesTable string `protobuf:"bytes,2,opt,name=rawFeaturesTable,proto3" json:"rawFeaturesTable,omitempty"`
	// name of table that will be used to populate `entitiesTable` field in prediction log
	EntitiesTable string `protobuf:"bytes,3,opt,name=entitiesTable,proto3" json:"entitiesTable,omitempty"`
}

func (x *PredictionLogConfig) Reset() {
	*x = PredictionLogConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_spec_prediction_log_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PredictionLogConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PredictionLogConfig) ProtoMessage() {}

func (x *PredictionLogConfig) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_spec_prediction_log_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PredictionLogConfig.ProtoReflect.Descriptor instead.
func (*PredictionLogConfig) Descriptor() ([]byte, []int) {
	return file_transformer_spec_prediction_log_proto_rawDescGZIP(), []int{0}
}

func (x *PredictionLogConfig) GetEnable() bool {
	if x != nil {
		return x.Enable
	}
	return false
}

func (x *PredictionLogConfig) GetRawFeaturesTable() string {
	if x != nil {
		return x.RawFeaturesTable
	}
	return ""
}

func (x *PredictionLogConfig) GetEntitiesTable() string {
	if x != nil {
		return x.EntitiesTable
	}
	return ""
}

var File_transformer_spec_prediction_log_proto protoreflect.FileDescriptor

var file_transformer_spec_prediction_log_proto_rawDesc = []byte{
	0x0a, 0x25, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2f, 0x73, 0x70,
	0x65, 0x63, 0x2f, 0x70, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6c, 0x6f,
	0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e,
	0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x22, 0x7f, 0x0a, 0x13, 0x50,
	0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4c, 0x6f, 0x67, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x06, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x2a, 0x0a, 0x10, 0x72, 0x61,
	0x77, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x10, 0x72, 0x61, 0x77, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x73, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x24, 0x0a, 0x0d, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x69,
	0x65, 0x73, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x65,
	0x6e, 0x74, 0x69, 0x74, 0x69, 0x65, 0x73, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x42, 0x33, 0x5a, 0x31,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x61, 0x72, 0x61, 0x6d,
	0x6c, 0x2d, 0x64, 0x65, 0x76, 0x2f, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2f, 0x70, 0x6b, 0x67,
	0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2f, 0x73, 0x70, 0x65,
	0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_transformer_spec_prediction_log_proto_rawDescOnce sync.Once
	file_transformer_spec_prediction_log_proto_rawDescData = file_transformer_spec_prediction_log_proto_rawDesc
)

func file_transformer_spec_prediction_log_proto_rawDescGZIP() []byte {
	file_transformer_spec_prediction_log_proto_rawDescOnce.Do(func() {
		file_transformer_spec_prediction_log_proto_rawDescData = protoimpl.X.CompressGZIP(file_transformer_spec_prediction_log_proto_rawDescData)
	})
	return file_transformer_spec_prediction_log_proto_rawDescData
}

var file_transformer_spec_prediction_log_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_transformer_spec_prediction_log_proto_goTypes = []interface{}{
	(*PredictionLogConfig)(nil), // 0: merlin.transformer.PredictionLogConfig
}
var file_transformer_spec_prediction_log_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_transformer_spec_prediction_log_proto_init() }
func file_transformer_spec_prediction_log_proto_init() {
	if File_transformer_spec_prediction_log_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_transformer_spec_prediction_log_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PredictionLogConfig); i {
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
			RawDescriptor: file_transformer_spec_prediction_log_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transformer_spec_prediction_log_proto_goTypes,
		DependencyIndexes: file_transformer_spec_prediction_log_proto_depIdxs,
		MessageInfos:      file_transformer_spec_prediction_log_proto_msgTypes,
	}.Build()
	File_transformer_spec_prediction_log_proto = out.File
	file_transformer_spec_prediction_log_proto_rawDesc = nil
	file_transformer_spec_prediction_log_proto_goTypes = nil
	file_transformer_spec_prediction_log_proto_depIdxs = nil
}
