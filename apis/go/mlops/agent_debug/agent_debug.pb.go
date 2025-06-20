/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

// Debug service for agent replica

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.27.2
// source: mlops/agent_debug/agent_debug.proto

package agent_debug

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ModelReplicaState_ModelState int32

const (
	ModelReplicaState_Evicted  ModelReplicaState_ModelState = 0
	ModelReplicaState_InMemory ModelReplicaState_ModelState = 1
)

// Enum value maps for ModelReplicaState_ModelState.
var (
	ModelReplicaState_ModelState_name = map[int32]string{
		0: "Evicted",
		1: "InMemory",
	}
	ModelReplicaState_ModelState_value = map[string]int32{
		"Evicted":  0,
		"InMemory": 1,
	}
)

func (x ModelReplicaState_ModelState) Enum() *ModelReplicaState_ModelState {
	p := new(ModelReplicaState_ModelState)
	*p = x
	return p
}

func (x ModelReplicaState_ModelState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ModelReplicaState_ModelState) Descriptor() protoreflect.EnumDescriptor {
	return file_mlops_agent_debug_agent_debug_proto_enumTypes[0].Descriptor()
}

func (ModelReplicaState_ModelState) Type() protoreflect.EnumType {
	return &file_mlops_agent_debug_agent_debug_proto_enumTypes[0]
}

func (x ModelReplicaState_ModelState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ModelReplicaState_ModelState.Descriptor instead.
func (ModelReplicaState_ModelState) EnumDescriptor() ([]byte, []int) {
	return file_mlops_agent_debug_agent_debug_proto_rawDescGZIP(), []int{0, 0}
}

type ModelReplicaState struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	State        ModelReplicaState_ModelState `protobuf:"varint,1,opt,name=state,proto3,enum=seldon.mlops.agent_debug.ModelReplicaState_ModelState" json:"state,omitempty"`
	LastAccessed *timestamppb.Timestamp       `protobuf:"bytes,2,opt,name=lastAccessed,proto3" json:"lastAccessed,omitempty"`
	Name         string                       `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *ModelReplicaState) Reset() {
	*x = ModelReplicaState{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mlops_agent_debug_agent_debug_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ModelReplicaState) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ModelReplicaState) ProtoMessage() {}

func (x *ModelReplicaState) ProtoReflect() protoreflect.Message {
	mi := &file_mlops_agent_debug_agent_debug_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ModelReplicaState.ProtoReflect.Descriptor instead.
func (*ModelReplicaState) Descriptor() ([]byte, []int) {
	return file_mlops_agent_debug_agent_debug_proto_rawDescGZIP(), []int{0}
}

func (x *ModelReplicaState) GetState() ModelReplicaState_ModelState {
	if x != nil {
		return x.State
	}
	return ModelReplicaState_Evicted
}

func (x *ModelReplicaState) GetLastAccessed() *timestamppb.Timestamp {
	if x != nil {
		return x.LastAccessed
	}
	return nil
}

func (x *ModelReplicaState) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type ReplicaStatusResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AvailableMemoryBytes uint64               `protobuf:"varint,1,opt,name=availableMemoryBytes,proto3" json:"availableMemoryBytes,omitempty"`
	Models               []*ModelReplicaState `protobuf:"bytes,2,rep,name=models,proto3" json:"models,omitempty"`
}

func (x *ReplicaStatusResponse) Reset() {
	*x = ReplicaStatusResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mlops_agent_debug_agent_debug_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReplicaStatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReplicaStatusResponse) ProtoMessage() {}

func (x *ReplicaStatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mlops_agent_debug_agent_debug_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReplicaStatusResponse.ProtoReflect.Descriptor instead.
func (*ReplicaStatusResponse) Descriptor() ([]byte, []int) {
	return file_mlops_agent_debug_agent_debug_proto_rawDescGZIP(), []int{1}
}

func (x *ReplicaStatusResponse) GetAvailableMemoryBytes() uint64 {
	if x != nil {
		return x.AvailableMemoryBytes
	}
	return 0
}

func (x *ReplicaStatusResponse) GetModels() []*ModelReplicaState {
	if x != nil {
		return x.Models
	}
	return nil
}

type ReplicaStatusRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ReplicaStatusRequest) Reset() {
	*x = ReplicaStatusRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mlops_agent_debug_agent_debug_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReplicaStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReplicaStatusRequest) ProtoMessage() {}

func (x *ReplicaStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mlops_agent_debug_agent_debug_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReplicaStatusRequest.ProtoReflect.Descriptor instead.
func (*ReplicaStatusRequest) Descriptor() ([]byte, []int) {
	return file_mlops_agent_debug_agent_debug_proto_rawDescGZIP(), []int{2}
}

var File_mlops_agent_debug_agent_debug_proto protoreflect.FileDescriptor

var file_mlops_agent_debug_agent_debug_proto_rawDesc = []byte{
	0x0a, 0x23, 0x6d, 0x6c, 0x6f, 0x70, 0x73, 0x2f, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x5f, 0x64, 0x65,
	0x62, 0x75, 0x67, 0x2f, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x5f, 0x64, 0x65, 0x62, 0x75, 0x67, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x18, 0x73, 0x65, 0x6c, 0x64, 0x6f, 0x6e, 0x2e, 0x6d, 0x6c,
	0x6f, 0x70, 0x73, 0x2e, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x5f, 0x64, 0x65, 0x62, 0x75, 0x67, 0x1a,
	0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xde, 0x01, 0x0a, 0x11, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63,
	0x61, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x4c, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x36, 0x2e, 0x73, 0x65, 0x6c, 0x64, 0x6f, 0x6e, 0x2e, 0x6d,
	0x6c, 0x6f, 0x70, 0x73, 0x2e, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x5f, 0x64, 0x65, 0x62, 0x75, 0x67,
	0x2e, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x2e, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73,
	0x74, 0x61, 0x74, 0x65, 0x12, 0x3e, 0x0a, 0x0c, 0x6c, 0x61, 0x73, 0x74, 0x41, 0x63, 0x63, 0x65,
	0x73, 0x73, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0c, 0x6c, 0x61, 0x73, 0x74, 0x41, 0x63, 0x63, 0x65,
	0x73, 0x73, 0x65, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x27, 0x0a, 0x0a, 0x4d, 0x6f, 0x64, 0x65,
	0x6c, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x45, 0x76, 0x69, 0x63, 0x74, 0x65,
	0x64, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x49, 0x6e, 0x4d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x10,
	0x01, 0x22, 0x90, 0x01, 0x0a, 0x15, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x32, 0x0a, 0x14, 0x61,
	0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x4d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x42, 0x79,
	0x74, 0x65, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x14, 0x61, 0x76, 0x61, 0x69, 0x6c,
	0x61, 0x62, 0x6c, 0x65, 0x4d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12,
	0x43, 0x0a, 0x06, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x2b, 0x2e, 0x73, 0x65, 0x6c, 0x64, 0x6f, 0x6e, 0x2e, 0x6d, 0x6c, 0x6f, 0x70, 0x73, 0x2e, 0x61,
	0x67, 0x65, 0x6e, 0x74, 0x5f, 0x64, 0x65, 0x62, 0x75, 0x67, 0x2e, 0x4d, 0x6f, 0x64, 0x65, 0x6c,
	0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x06, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x73, 0x22, 0x16, 0x0a, 0x14, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x32, 0x87, 0x01, 0x0a,
	0x11, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x44, 0x65, 0x62, 0x75, 0x67, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x72, 0x0a, 0x0d, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x12, 0x2e, 0x2e, 0x73, 0x65, 0x6c, 0x64, 0x6f, 0x6e, 0x2e, 0x6d, 0x6c, 0x6f,
	0x70, 0x73, 0x2e, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x5f, 0x64, 0x65, 0x62, 0x75, 0x67, 0x2e, 0x52,
	0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x2f, 0x2e, 0x73, 0x65, 0x6c, 0x64, 0x6f, 0x6e, 0x2e, 0x6d, 0x6c, 0x6f,
	0x70, 0x73, 0x2e, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x5f, 0x64, 0x65, 0x62, 0x75, 0x67, 0x2e, 0x52,
	0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x3e, 0x5a, 0x3c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x65, 0x6c, 0x64, 0x6f, 0x6e, 0x69, 0x6f, 0x2f, 0x73, 0x65,
	0x6c, 0x64, 0x6f, 0x6e, 0x2d, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x67,
	0x6f, 0x2f, 0x76, 0x32, 0x2f, 0x6d, 0x6c, 0x6f, 0x70, 0x73, 0x2f, 0x61, 0x67, 0x65, 0x6e, 0x74,
	0x5f, 0x64, 0x65, 0x62, 0x75, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mlops_agent_debug_agent_debug_proto_rawDescOnce sync.Once
	file_mlops_agent_debug_agent_debug_proto_rawDescData = file_mlops_agent_debug_agent_debug_proto_rawDesc
)

func file_mlops_agent_debug_agent_debug_proto_rawDescGZIP() []byte {
	file_mlops_agent_debug_agent_debug_proto_rawDescOnce.Do(func() {
		file_mlops_agent_debug_agent_debug_proto_rawDescData = protoimpl.X.CompressGZIP(file_mlops_agent_debug_agent_debug_proto_rawDescData)
	})
	return file_mlops_agent_debug_agent_debug_proto_rawDescData
}

var file_mlops_agent_debug_agent_debug_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_mlops_agent_debug_agent_debug_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_mlops_agent_debug_agent_debug_proto_goTypes = []any{
	(ModelReplicaState_ModelState)(0), // 0: seldon.mlops.agent_debug.ModelReplicaState.ModelState
	(*ModelReplicaState)(nil),         // 1: seldon.mlops.agent_debug.ModelReplicaState
	(*ReplicaStatusResponse)(nil),     // 2: seldon.mlops.agent_debug.ReplicaStatusResponse
	(*ReplicaStatusRequest)(nil),      // 3: seldon.mlops.agent_debug.ReplicaStatusRequest
	(*timestamppb.Timestamp)(nil),     // 4: google.protobuf.Timestamp
}
var file_mlops_agent_debug_agent_debug_proto_depIdxs = []int32{
	0, // 0: seldon.mlops.agent_debug.ModelReplicaState.state:type_name -> seldon.mlops.agent_debug.ModelReplicaState.ModelState
	4, // 1: seldon.mlops.agent_debug.ModelReplicaState.lastAccessed:type_name -> google.protobuf.Timestamp
	1, // 2: seldon.mlops.agent_debug.ReplicaStatusResponse.models:type_name -> seldon.mlops.agent_debug.ModelReplicaState
	3, // 3: seldon.mlops.agent_debug.AgentDebugService.ReplicaStatus:input_type -> seldon.mlops.agent_debug.ReplicaStatusRequest
	2, // 4: seldon.mlops.agent_debug.AgentDebugService.ReplicaStatus:output_type -> seldon.mlops.agent_debug.ReplicaStatusResponse
	4, // [4:5] is the sub-list for method output_type
	3, // [3:4] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_mlops_agent_debug_agent_debug_proto_init() }
func file_mlops_agent_debug_agent_debug_proto_init() {
	if File_mlops_agent_debug_agent_debug_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mlops_agent_debug_agent_debug_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*ModelReplicaState); i {
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
		file_mlops_agent_debug_agent_debug_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*ReplicaStatusResponse); i {
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
		file_mlops_agent_debug_agent_debug_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*ReplicaStatusRequest); i {
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
			RawDescriptor: file_mlops_agent_debug_agent_debug_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_mlops_agent_debug_agent_debug_proto_goTypes,
		DependencyIndexes: file_mlops_agent_debug_agent_debug_proto_depIdxs,
		EnumInfos:         file_mlops_agent_debug_agent_debug_proto_enumTypes,
		MessageInfos:      file_mlops_agent_debug_agent_debug_proto_msgTypes,
	}.Build()
	File_mlops_agent_debug_agent_debug_proto = out.File
	file_mlops_agent_debug_agent_debug_proto_rawDesc = nil
	file_mlops_agent_debug_agent_debug_proto_goTypes = nil
	file_mlops_agent_debug_agent_debug_proto_depIdxs = nil
}
