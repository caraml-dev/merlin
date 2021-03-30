// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.12.3
// source: transformer/table.proto

package transformer

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

type SortOrder int32

const (
	SortOrder_INVALID_SORT_ORDER SortOrder = 0
	SortOrder_ASC                SortOrder = 1
	SortOrder_DESC               SortOrder = 2
)

// Enum value maps for SortOrder.
var (
	SortOrder_name = map[int32]string{
		0: "INVALID_SORT_ORDER",
		1: "ASC",
		2: "DESC",
	}
	SortOrder_value = map[string]int32{
		"INVALID_SORT_ORDER": 0,
		"ASC":                1,
		"DESC":               2,
	}
)

func (x SortOrder) Enum() *SortOrder {
	p := new(SortOrder)
	*p = x
	return p
}

func (x SortOrder) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SortOrder) Descriptor() protoreflect.EnumDescriptor {
	return file_transformer_table_proto_enumTypes[0].Descriptor()
}

func (SortOrder) Type() protoreflect.EnumType {
	return &file_transformer_table_proto_enumTypes[0]
}

func (x SortOrder) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SortOrder.Descriptor instead.
func (SortOrder) EnumDescriptor() ([]byte, []int) {
	return file_transformer_table_proto_rawDescGZIP(), []int{0}
}

type JoinMethod int32

const (
	JoinMethod_INVALID_JOIN JoinMethod = 0
	JoinMethod_LEFT         JoinMethod = 1
	JoinMethod_RIGHT        JoinMethod = 2
	JoinMethod_INNER        JoinMethod = 3
	JoinMethod_OUTER        JoinMethod = 4
	JoinMethod_CROSS        JoinMethod = 5
	JoinMethod_CONCAT       JoinMethod = 6
)

// Enum value maps for JoinMethod.
var (
	JoinMethod_name = map[int32]string{
		0: "INVALID_JOIN",
		1: "LEFT",
		2: "RIGHT",
		3: "INNER",
		4: "OUTER",
		5: "CROSS",
		6: "CONCAT",
	}
	JoinMethod_value = map[string]int32{
		"INVALID_JOIN": 0,
		"LEFT":         1,
		"RIGHT":        2,
		"INNER":        3,
		"OUTER":        4,
		"CROSS":        5,
		"CONCAT":       6,
	}
)

func (x JoinMethod) Enum() *JoinMethod {
	p := new(JoinMethod)
	*p = x
	return p
}

func (x JoinMethod) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (JoinMethod) Descriptor() protoreflect.EnumDescriptor {
	return file_transformer_table_proto_enumTypes[1].Descriptor()
}

func (JoinMethod) Type() protoreflect.EnumType {
	return &file_transformer_table_proto_enumTypes[1]
}

func (x JoinMethod) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use JoinMethod.Descriptor instead.
func (JoinMethod) EnumDescriptor() ([]byte, []int) {
	return file_transformer_table_proto_rawDescGZIP(), []int{1}
}

type Table struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name      string     `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	BaseTable *BaseTable `protobuf:"bytes,2,opt,name=baseTable,proto3" json:"baseTable,omitempty"`
	Columns   []*Column  `protobuf:"bytes,3,rep,name=columns,proto3" json:"columns,omitempty"`
}

func (x *Table) Reset() {
	*x = Table{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_table_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Table) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Table) ProtoMessage() {}

func (x *Table) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_table_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Table.ProtoReflect.Descriptor instead.
func (*Table) Descriptor() ([]byte, []int) {
	return file_transformer_table_proto_rawDescGZIP(), []int{0}
}

func (x *Table) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Table) GetBaseTable() *BaseTable {
	if x != nil {
		return x.BaseTable
	}
	return nil
}

func (x *Table) GetColumns() []*Column {
	if x != nil {
		return x.Columns
	}
	return nil
}

type BaseTable struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to BaseTable:
	//	*BaseTable_FromJson
	//	*BaseTable_FromTable
	BaseTable isBaseTable_BaseTable `protobuf_oneof:"baseTable"`
}

func (x *BaseTable) Reset() {
	*x = BaseTable{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_table_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BaseTable) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BaseTable) ProtoMessage() {}

func (x *BaseTable) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_table_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BaseTable.ProtoReflect.Descriptor instead.
func (*BaseTable) Descriptor() ([]byte, []int) {
	return file_transformer_table_proto_rawDescGZIP(), []int{1}
}

func (m *BaseTable) GetBaseTable() isBaseTable_BaseTable {
	if m != nil {
		return m.BaseTable
	}
	return nil
}

func (x *BaseTable) GetFromJson() *FromJson {
	if x, ok := x.GetBaseTable().(*BaseTable_FromJson); ok {
		return x.FromJson
	}
	return nil
}

func (x *BaseTable) GetFromTable() *FromTable {
	if x, ok := x.GetBaseTable().(*BaseTable_FromTable); ok {
		return x.FromTable
	}
	return nil
}

type isBaseTable_BaseTable interface {
	isBaseTable_BaseTable()
}

type BaseTable_FromJson struct {
	FromJson *FromJson `protobuf:"bytes,1,opt,name=fromJson,proto3,oneof"`
}

type BaseTable_FromTable struct {
	FromTable *FromTable `protobuf:"bytes,2,opt,name=fromTable,proto3,oneof"`
}

func (*BaseTable_FromJson) isBaseTable_BaseTable() {}

func (*BaseTable_FromTable) isBaseTable_BaseTable() {}

type Column struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Types that are assignable to Value:
	//	*Column_FromJson
	//	*Column_Expression
	Value isColumn_Value `protobuf_oneof:"value"`
}

func (x *Column) Reset() {
	*x = Column{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_table_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Column) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Column) ProtoMessage() {}

func (x *Column) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_table_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Column.ProtoReflect.Descriptor instead.
func (*Column) Descriptor() ([]byte, []int) {
	return file_transformer_table_proto_rawDescGZIP(), []int{2}
}

func (x *Column) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (m *Column) GetValue() isColumn_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (x *Column) GetFromJson() *FromJson {
	if x, ok := x.GetValue().(*Column_FromJson); ok {
		return x.FromJson
	}
	return nil
}

func (x *Column) GetExpression() string {
	if x, ok := x.GetValue().(*Column_Expression); ok {
		return x.Expression
	}
	return ""
}

type isColumn_Value interface {
	isColumn_Value()
}

type Column_FromJson struct {
	FromJson *FromJson `protobuf:"bytes,2,opt,name=fromJson,proto3,oneof"`
}

type Column_Expression struct {
	Expression string `protobuf:"bytes,3,opt,name=expression,proto3,oneof"`
}

func (*Column_FromJson) isColumn_Value() {}

func (*Column_Expression) isColumn_Value() {}

type TableTransformation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InputTable  string                `protobuf:"bytes,1,opt,name=inputTable,proto3" json:"inputTable,omitempty"`
	OutputTable string                `protobuf:"bytes,2,opt,name=outputTable,proto3" json:"outputTable,omitempty"`
	Steps       []*TransformationStep `protobuf:"bytes,3,rep,name=steps,proto3" json:"steps,omitempty"`
}

func (x *TableTransformation) Reset() {
	*x = TableTransformation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_table_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TableTransformation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TableTransformation) ProtoMessage() {}

func (x *TableTransformation) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_table_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TableTransformation.ProtoReflect.Descriptor instead.
func (*TableTransformation) Descriptor() ([]byte, []int) {
	return file_transformer_table_proto_rawDescGZIP(), []int{3}
}

func (x *TableTransformation) GetInputTable() string {
	if x != nil {
		return x.InputTable
	}
	return ""
}

func (x *TableTransformation) GetOutputTable() string {
	if x != nil {
		return x.OutputTable
	}
	return ""
}

func (x *TableTransformation) GetSteps() []*TransformationStep {
	if x != nil {
		return x.Steps
	}
	return nil
}

type TransformationStep struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DropColumns       []string           `protobuf:"bytes,1,rep,name=dropColumns,proto3" json:"dropColumns,omitempty"`
	Sort              []*SortColumnRule  `protobuf:"bytes,2,rep,name=sort,proto3" json:"sort,omitempty"`
	RenameColumn      *RenameColumn      `protobuf:"bytes,3,opt,name=renameColumn,proto3" json:"renameColumn,omitempty"`
	AddOrUpdateColumn *AddOrUpdateColumn `protobuf:"bytes,4,opt,name=addOrUpdateColumn,proto3" json:"addOrUpdateColumn,omitempty"`
}

func (x *TransformationStep) Reset() {
	*x = TransformationStep{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_table_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TransformationStep) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransformationStep) ProtoMessage() {}

func (x *TransformationStep) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_table_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TransformationStep.ProtoReflect.Descriptor instead.
func (*TransformationStep) Descriptor() ([]byte, []int) {
	return file_transformer_table_proto_rawDescGZIP(), []int{4}
}

func (x *TransformationStep) GetDropColumns() []string {
	if x != nil {
		return x.DropColumns
	}
	return nil
}

func (x *TransformationStep) GetSort() []*SortColumnRule {
	if x != nil {
		return x.Sort
	}
	return nil
}

func (x *TransformationStep) GetRenameColumn() *RenameColumn {
	if x != nil {
		return x.RenameColumn
	}
	return nil
}

func (x *TransformationStep) GetAddOrUpdateColumn() *AddOrUpdateColumn {
	if x != nil {
		return x.AddOrUpdateColumn
	}
	return nil
}

type SortColumnRule struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Column string    `protobuf:"bytes,1,opt,name=column,proto3" json:"column,omitempty"`
	Order  SortOrder `protobuf:"varint,2,opt,name=order,proto3,enum=merlin.transformer.SortOrder" json:"order,omitempty"`
}

func (x *SortColumnRule) Reset() {
	*x = SortColumnRule{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_table_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SortColumnRule) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SortColumnRule) ProtoMessage() {}

func (x *SortColumnRule) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_table_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SortColumnRule.ProtoReflect.Descriptor instead.
func (*SortColumnRule) Descriptor() ([]byte, []int) {
	return file_transformer_table_proto_rawDescGZIP(), []int{5}
}

func (x *SortColumnRule) GetColumn() string {
	if x != nil {
		return x.Column
	}
	return ""
}

func (x *SortColumnRule) GetOrder() SortOrder {
	if x != nil {
		return x.Order
	}
	return SortOrder_INVALID_SORT_ORDER
}

type RenameColumn struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	From string `protobuf:"bytes,1,opt,name=from,proto3" json:"from,omitempty"`
	To   string `protobuf:"bytes,2,opt,name=to,proto3" json:"to,omitempty"`
}

func (x *RenameColumn) Reset() {
	*x = RenameColumn{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_table_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RenameColumn) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RenameColumn) ProtoMessage() {}

func (x *RenameColumn) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_table_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RenameColumn.ProtoReflect.Descriptor instead.
func (*RenameColumn) Descriptor() ([]byte, []int) {
	return file_transformer_table_proto_rawDescGZIP(), []int{6}
}

func (x *RenameColumn) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *RenameColumn) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

type AddOrUpdateColumn struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Column     string `protobuf:"bytes,1,opt,name=column,proto3" json:"column,omitempty"`
	Expression string `protobuf:"bytes,2,opt,name=expression,proto3" json:"expression,omitempty"`
}

func (x *AddOrUpdateColumn) Reset() {
	*x = AddOrUpdateColumn{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_table_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddOrUpdateColumn) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddOrUpdateColumn) ProtoMessage() {}

func (x *AddOrUpdateColumn) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_table_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddOrUpdateColumn.ProtoReflect.Descriptor instead.
func (*AddOrUpdateColumn) Descriptor() ([]byte, []int) {
	return file_transformer_table_proto_rawDescGZIP(), []int{7}
}

func (x *AddOrUpdateColumn) GetColumn() string {
	if x != nil {
		return x.Column
	}
	return ""
}

func (x *AddOrUpdateColumn) GetExpression() string {
	if x != nil {
		return x.Expression
	}
	return ""
}

type TableJoin struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Left    string     `protobuf:"bytes,1,opt,name=left,proto3" json:"left,omitempty"`
	Right   string     `protobuf:"bytes,2,opt,name=right,proto3" json:"right,omitempty"`
	Output  string     `protobuf:"bytes,3,opt,name=output,proto3" json:"output,omitempty"`
	How     JoinMethod `protobuf:"varint,4,opt,name=how,proto3,enum=merlin.transformer.JoinMethod" json:"how,omitempty"`
	OnLeft  string     `protobuf:"bytes,5,opt,name=onLeft,proto3" json:"onLeft,omitempty"`
	OnRight string     `protobuf:"bytes,6,opt,name=onRight,proto3" json:"onRight,omitempty"`
}

func (x *TableJoin) Reset() {
	*x = TableJoin{}
	if protoimpl.UnsafeEnabled {
		mi := &file_transformer_table_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TableJoin) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TableJoin) ProtoMessage() {}

func (x *TableJoin) ProtoReflect() protoreflect.Message {
	mi := &file_transformer_table_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TableJoin.ProtoReflect.Descriptor instead.
func (*TableJoin) Descriptor() ([]byte, []int) {
	return file_transformer_table_proto_rawDescGZIP(), []int{8}
}

func (x *TableJoin) GetLeft() string {
	if x != nil {
		return x.Left
	}
	return ""
}

func (x *TableJoin) GetRight() string {
	if x != nil {
		return x.Right
	}
	return ""
}

func (x *TableJoin) GetOutput() string {
	if x != nil {
		return x.Output
	}
	return ""
}

func (x *TableJoin) GetHow() JoinMethod {
	if x != nil {
		return x.How
	}
	return JoinMethod_INVALID_JOIN
}

func (x *TableJoin) GetOnLeft() string {
	if x != nil {
		return x.OnLeft
	}
	return ""
}

func (x *TableJoin) GetOnRight() string {
	if x != nil {
		return x.OnRight
	}
	return ""
}

var File_transformer_table_proto protoreflect.FileDescriptor

var file_transformer_table_proto_rawDesc = []byte{
	0x0a, 0x17, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2f, 0x74, 0x61,
	0x62, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x6d, 0x65, 0x72, 0x6c, 0x69,
	0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x1a, 0x18, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8e, 0x01, 0x0a, 0x05, 0x54, 0x61, 0x62, 0x6c,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x3b, 0x0a, 0x09, 0x62, 0x61, 0x73, 0x65, 0x54, 0x61, 0x62,
	0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69,
	0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x42, 0x61,
	0x73, 0x65, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x52, 0x09, 0x62, 0x61, 0x73, 0x65, 0x54, 0x61, 0x62,
	0x6c, 0x65, 0x12, 0x34, 0x0a, 0x07, 0x63, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x73, 0x18, 0x03, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x52,
	0x07, 0x63, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x73, 0x22, 0x93, 0x01, 0x0a, 0x09, 0x42, 0x61, 0x73,
	0x65, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x3a, 0x0a, 0x08, 0x66, 0x72, 0x6f, 0x6d, 0x4a, 0x73,
	0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69,
	0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x46, 0x72,
	0x6f, 0x6d, 0x4a, 0x73, 0x6f, 0x6e, 0x48, 0x00, 0x52, 0x08, 0x66, 0x72, 0x6f, 0x6d, 0x4a, 0x73,
	0x6f, 0x6e, 0x12, 0x3d, 0x0a, 0x09, 0x66, 0x72, 0x6f, 0x6d, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x46, 0x72, 0x6f, 0x6d, 0x54,
	0x61, 0x62, 0x6c, 0x65, 0x48, 0x00, 0x52, 0x09, 0x66, 0x72, 0x6f, 0x6d, 0x54, 0x61, 0x62, 0x6c,
	0x65, 0x42, 0x0b, 0x0a, 0x09, 0x62, 0x61, 0x73, 0x65, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x22, 0x83,
	0x01, 0x0a, 0x06, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x3a, 0x0a,
	0x08, 0x66, 0x72, 0x6f, 0x6d, 0x4a, 0x73, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1c, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f,
	0x72, 0x6d, 0x65, 0x72, 0x2e, 0x46, 0x72, 0x6f, 0x6d, 0x4a, 0x73, 0x6f, 0x6e, 0x48, 0x00, 0x52,
	0x08, 0x66, 0x72, 0x6f, 0x6d, 0x4a, 0x73, 0x6f, 0x6e, 0x12, 0x20, 0x0a, 0x0a, 0x65, 0x78, 0x70,
	0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52,
	0x0a, 0x65, 0x78, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x42, 0x07, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x22, 0x95, 0x01, 0x0a, 0x13, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x54, 0x72,
	0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1e, 0x0a, 0x0a,
	0x69, 0x6e, 0x70, 0x75, 0x74, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0a, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x20, 0x0a, 0x0b,
	0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0b, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x3c,
	0x0a, 0x05, 0x73, 0x74, 0x65, 0x70, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x26, 0x2e,
	0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d,
	0x65, 0x72, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x53, 0x74, 0x65, 0x70, 0x52, 0x05, 0x73, 0x74, 0x65, 0x70, 0x73, 0x22, 0x89, 0x02, 0x0a,
	0x12, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53,
	0x74, 0x65, 0x70, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x72, 0x6f, 0x70, 0x43, 0x6f, 0x6c, 0x75, 0x6d,
	0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x72, 0x6f, 0x70, 0x43, 0x6f,
	0x6c, 0x75, 0x6d, 0x6e, 0x73, 0x12, 0x36, 0x0a, 0x04, 0x73, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x53, 0x6f, 0x72, 0x74, 0x43, 0x6f, 0x6c,
	0x75, 0x6d, 0x6e, 0x52, 0x75, 0x6c, 0x65, 0x52, 0x04, 0x73, 0x6f, 0x72, 0x74, 0x12, 0x44, 0x0a,
	0x0c, 0x72, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x43,
	0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x52, 0x0c, 0x72, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x43, 0x6f, 0x6c,
	0x75, 0x6d, 0x6e, 0x12, 0x53, 0x0a, 0x11, 0x61, 0x64, 0x64, 0x4f, 0x72, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25,
	0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72,
	0x6d, 0x65, 0x72, 0x2e, 0x41, 0x64, 0x64, 0x4f, 0x72, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x43,
	0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x52, 0x11, 0x61, 0x64, 0x64, 0x4f, 0x72, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x22, 0x5d, 0x0a, 0x0e, 0x53, 0x6f, 0x72, 0x74,
	0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x52, 0x75, 0x6c, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x63, 0x6f,
	0x6c, 0x75, 0x6d, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x63, 0x6f, 0x6c, 0x75,
	0x6d, 0x6e, 0x12, 0x33, 0x0a, 0x05, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x1d, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73,
	0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x53, 0x6f, 0x72, 0x74, 0x4f, 0x72, 0x64, 0x65, 0x72,
	0x52, 0x05, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x22, 0x32, 0x0a, 0x0c, 0x52, 0x65, 0x6e, 0x61, 0x6d,
	0x65, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x74,
	0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x74, 0x6f, 0x22, 0x4b, 0x0a, 0x11, 0x41,
	0x64, 0x64, 0x4f, 0x72, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e,
	0x12, 0x16, 0x0a, 0x06, 0x63, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x63, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x78, 0x70, 0x72,
	0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x65, 0x78,
	0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0xb1, 0x01, 0x0a, 0x09, 0x54, 0x61, 0x62,
	0x6c, 0x65, 0x4a, 0x6f, 0x69, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6c, 0x65, 0x66, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6c, 0x65, 0x66, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x69,
	0x67, 0x68, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x72, 0x69, 0x67, 0x68, 0x74,
	0x12, 0x16, 0x0a, 0x06, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x30, 0x0a, 0x03, 0x68, 0x6f, 0x77, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1e, 0x2e, 0x6d, 0x65, 0x72, 0x6c, 0x69, 0x6e, 0x2e, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65, 0x72, 0x2e, 0x4a, 0x6f, 0x69, 0x6e, 0x4d,
	0x65, 0x74, 0x68, 0x6f, 0x64, 0x52, 0x03, 0x68, 0x6f, 0x77, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x6e,
	0x4c, 0x65, 0x66, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6f, 0x6e, 0x4c, 0x65,
	0x66, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x6f, 0x6e, 0x52, 0x69, 0x67, 0x68, 0x74, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x6f, 0x6e, 0x52, 0x69, 0x67, 0x68, 0x74, 0x2a, 0x36, 0x0a, 0x09,
	0x53, 0x6f, 0x72, 0x74, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x12, 0x49, 0x4e, 0x56,
	0x41, 0x4c, 0x49, 0x44, 0x5f, 0x53, 0x4f, 0x52, 0x54, 0x5f, 0x4f, 0x52, 0x44, 0x45, 0x52, 0x10,
	0x00, 0x12, 0x07, 0x0a, 0x03, 0x41, 0x53, 0x43, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x44, 0x45,
	0x53, 0x43, 0x10, 0x02, 0x2a, 0x60, 0x0a, 0x0a, 0x4a, 0x6f, 0x69, 0x6e, 0x4d, 0x65, 0x74, 0x68,
	0x6f, 0x64, 0x12, 0x10, 0x0a, 0x0c, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x5f, 0x4a, 0x4f,
	0x49, 0x4e, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x4c, 0x45, 0x46, 0x54, 0x10, 0x01, 0x12, 0x09,
	0x0a, 0x05, 0x52, 0x49, 0x47, 0x48, 0x54, 0x10, 0x02, 0x12, 0x09, 0x0a, 0x05, 0x49, 0x4e, 0x4e,
	0x45, 0x52, 0x10, 0x03, 0x12, 0x09, 0x0a, 0x05, 0x4f, 0x55, 0x54, 0x45, 0x52, 0x10, 0x04, 0x12,
	0x09, 0x0a, 0x05, 0x43, 0x52, 0x4f, 0x53, 0x53, 0x10, 0x05, 0x12, 0x0a, 0x0a, 0x06, 0x43, 0x4f,
	0x4e, 0x43, 0x41, 0x54, 0x10, 0x06, 0x42, 0x29, 0x5a, 0x27, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x6f, 0x6a, 0x65, 0x6b, 0x2f, 0x6d, 0x65, 0x72, 0x6c, 0x69,
	0x6e, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x6f, 0x72, 0x6d, 0x65,
	0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_transformer_table_proto_rawDescOnce sync.Once
	file_transformer_table_proto_rawDescData = file_transformer_table_proto_rawDesc
)

func file_transformer_table_proto_rawDescGZIP() []byte {
	file_transformer_table_proto_rawDescOnce.Do(func() {
		file_transformer_table_proto_rawDescData = protoimpl.X.CompressGZIP(file_transformer_table_proto_rawDescData)
	})
	return file_transformer_table_proto_rawDescData
}

var file_transformer_table_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_transformer_table_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_transformer_table_proto_goTypes = []interface{}{
	(SortOrder)(0),              // 0: merlin.transformer.SortOrder
	(JoinMethod)(0),             // 1: merlin.transformer.JoinMethod
	(*Table)(nil),               // 2: merlin.transformer.Table
	(*BaseTable)(nil),           // 3: merlin.transformer.BaseTable
	(*Column)(nil),              // 4: merlin.transformer.Column
	(*TableTransformation)(nil), // 5: merlin.transformer.TableTransformation
	(*TransformationStep)(nil),  // 6: merlin.transformer.TransformationStep
	(*SortColumnRule)(nil),      // 7: merlin.transformer.SortColumnRule
	(*RenameColumn)(nil),        // 8: merlin.transformer.RenameColumn
	(*AddOrUpdateColumn)(nil),   // 9: merlin.transformer.AddOrUpdateColumn
	(*TableJoin)(nil),           // 10: merlin.transformer.TableJoin
	(*FromJson)(nil),            // 11: merlin.transformer.FromJson
	(*FromTable)(nil),           // 12: merlin.transformer.FromTable
}
var file_transformer_table_proto_depIdxs = []int32{
	3,  // 0: merlin.transformer.Table.baseTable:type_name -> merlin.transformer.BaseTable
	4,  // 1: merlin.transformer.Table.columns:type_name -> merlin.transformer.Column
	11, // 2: merlin.transformer.BaseTable.fromJson:type_name -> merlin.transformer.FromJson
	12, // 3: merlin.transformer.BaseTable.fromTable:type_name -> merlin.transformer.FromTable
	11, // 4: merlin.transformer.Column.fromJson:type_name -> merlin.transformer.FromJson
	6,  // 5: merlin.transformer.TableTransformation.steps:type_name -> merlin.transformer.TransformationStep
	7,  // 6: merlin.transformer.TransformationStep.sort:type_name -> merlin.transformer.SortColumnRule
	8,  // 7: merlin.transformer.TransformationStep.renameColumn:type_name -> merlin.transformer.RenameColumn
	9,  // 8: merlin.transformer.TransformationStep.addOrUpdateColumn:type_name -> merlin.transformer.AddOrUpdateColumn
	0,  // 9: merlin.transformer.SortColumnRule.order:type_name -> merlin.transformer.SortOrder
	1,  // 10: merlin.transformer.TableJoin.how:type_name -> merlin.transformer.JoinMethod
	11, // [11:11] is the sub-list for method output_type
	11, // [11:11] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_transformer_table_proto_init() }
func file_transformer_table_proto_init() {
	if File_transformer_table_proto != nil {
		return
	}
	file_transformer_common_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_transformer_table_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Table); i {
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
		file_transformer_table_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BaseTable); i {
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
		file_transformer_table_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Column); i {
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
		file_transformer_table_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TableTransformation); i {
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
		file_transformer_table_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TransformationStep); i {
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
		file_transformer_table_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SortColumnRule); i {
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
		file_transformer_table_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RenameColumn); i {
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
		file_transformer_table_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddOrUpdateColumn); i {
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
		file_transformer_table_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TableJoin); i {
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
	file_transformer_table_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*BaseTable_FromJson)(nil),
		(*BaseTable_FromTable)(nil),
	}
	file_transformer_table_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*Column_FromJson)(nil),
		(*Column_Expression)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_transformer_table_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_transformer_table_proto_goTypes,
		DependencyIndexes: file_transformer_table_proto_depIdxs,
		EnumInfos:         file_transformer_table_proto_enumTypes,
		MessageInfos:      file_transformer_table_proto_msgTypes,
	}.Build()
	File_transformer_table_proto = out.File
	file_transformer_table_proto_rawDesc = nil
	file_transformer_table_proto_goTypes = nil
	file_transformer_table_proto_depIdxs = nil
}
