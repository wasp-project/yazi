// Copyright 2024 mlycore. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v3.19.4
// source: rpcserver.proto

package pb

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

type KVRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key   string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *KVRequest) Reset() {
	*x = KVRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpcserver_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KVRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KVRequest) ProtoMessage() {}

func (x *KVRequest) ProtoReflect() protoreflect.Message {
	mi := &file_rpcserver_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KVRequest.ProtoReflect.Descriptor instead.
func (*KVRequest) Descriptor() ([]byte, []int) {
	return file_rpcserver_proto_rawDescGZIP(), []int{0}
}

func (x *KVRequest) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *KVRequest) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type KVResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value string `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *KVResponse) Reset() {
	*x = KVResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpcserver_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KVResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KVResponse) ProtoMessage() {}

func (x *KVResponse) ProtoReflect() protoreflect.Message {
	mi := &file_rpcserver_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KVResponse.ProtoReflect.Descriptor instead.
func (*KVResponse) Descriptor() ([]byte, []int) {
	return file_rpcserver_proto_rawDescGZIP(), []int{1}
}

func (x *KVResponse) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type MKVRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Keys   []string `protobuf:"bytes,1,rep,name=keys,proto3" json:"keys,omitempty"`
	Values []string `protobuf:"bytes,2,rep,name=values,proto3" json:"values,omitempty"`
}

func (x *MKVRequest) Reset() {
	*x = MKVRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpcserver_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MKVRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MKVRequest) ProtoMessage() {}

func (x *MKVRequest) ProtoReflect() protoreflect.Message {
	mi := &file_rpcserver_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MKVRequest.ProtoReflect.Descriptor instead.
func (*MKVRequest) Descriptor() ([]byte, []int) {
	return file_rpcserver_proto_rawDescGZIP(), []int{2}
}

func (x *MKVRequest) GetKeys() []string {
	if x != nil {
		return x.Keys
	}
	return nil
}

func (x *MKVRequest) GetValues() []string {
	if x != nil {
		return x.Values
	}
	return nil
}

type MKVResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Values []string `protobuf:"bytes,1,rep,name=values,proto3" json:"values,omitempty"`
}

func (x *MKVResponse) Reset() {
	*x = MKVResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpcserver_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MKVResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MKVResponse) ProtoMessage() {}

func (x *MKVResponse) ProtoReflect() protoreflect.Message {
	mi := &file_rpcserver_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MKVResponse.ProtoReflect.Descriptor instead.
func (*MKVResponse) Descriptor() ([]byte, []int) {
	return file_rpcserver_proto_rawDescGZIP(), []int{3}
}

func (x *MKVResponse) GetValues() []string {
	if x != nil {
		return x.Values
	}
	return nil
}

var File_rpcserver_proto protoreflect.FileDescriptor

var file_rpcserver_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x72, 0x70, 0x63, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x02, 0x70, 0x62, 0x22, 0x33, 0x0a, 0x09, 0x4b, 0x56, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x22, 0x0a, 0x0a, 0x4b, 0x56,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x38,
	0x0a, 0x0a, 0x4d, 0x4b, 0x56, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04,
	0x6b, 0x65, 0x79, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x65, 0x79, 0x73,
	0x12, 0x16, 0x0a, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x22, 0x25, 0x0a, 0x0b, 0x4d, 0x4b, 0x56, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x32,
	0x84, 0x02, 0x0a, 0x09, 0x52, 0x50, 0x43, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x12, 0x26, 0x0a,
	0x03, 0x53, 0x65, 0x74, 0x12, 0x0d, 0x2e, 0x70, 0x62, 0x2e, 0x4b, 0x56, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x0e, 0x2e, 0x70, 0x62, 0x2e, 0x4b, 0x56, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x26, 0x0a, 0x03, 0x47, 0x65, 0x74, 0x12, 0x0d, 0x2e, 0x70,
	0x62, 0x2e, 0x4b, 0x56, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0e, 0x2e, 0x70, 0x62,
	0x2e, 0x4b, 0x56, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x26, 0x0a,
	0x03, 0x44, 0x65, 0x6c, 0x12, 0x0d, 0x2e, 0x70, 0x62, 0x2e, 0x4b, 0x56, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x0e, 0x2e, 0x70, 0x62, 0x2e, 0x4b, 0x56, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x29, 0x0a, 0x04, 0x4b, 0x65, 0x79, 0x73, 0x12, 0x0e, 0x2e,
	0x70, 0x62, 0x2e, 0x4d, 0x4b, 0x56, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e,
	0x70, 0x62, 0x2e, 0x4d, 0x4b, 0x56, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x12, 0x29, 0x0a, 0x04, 0x4d, 0x47, 0x65, 0x74, 0x12, 0x0e, 0x2e, 0x70, 0x62, 0x2e, 0x4d, 0x4b,
	0x56, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x70, 0x62, 0x2e, 0x4d, 0x4b,
	0x56, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x29, 0x0a, 0x04, 0x4d,
	0x53, 0x65, 0x74, 0x12, 0x0e, 0x2e, 0x70, 0x62, 0x2e, 0x4d, 0x4b, 0x56, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x70, 0x62, 0x2e, 0x4d, 0x4b, 0x56, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x2c, 0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x77, 0x61, 0x73, 0x70, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x2f, 0x79, 0x61, 0x7a, 0x69, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_rpcserver_proto_rawDescOnce sync.Once
	file_rpcserver_proto_rawDescData = file_rpcserver_proto_rawDesc
)

func file_rpcserver_proto_rawDescGZIP() []byte {
	file_rpcserver_proto_rawDescOnce.Do(func() {
		file_rpcserver_proto_rawDescData = protoimpl.X.CompressGZIP(file_rpcserver_proto_rawDescData)
	})
	return file_rpcserver_proto_rawDescData
}

var file_rpcserver_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_rpcserver_proto_goTypes = []interface{}{
	(*KVRequest)(nil),   // 0: pb.KVRequest
	(*KVResponse)(nil),  // 1: pb.KVResponse
	(*MKVRequest)(nil),  // 2: pb.MKVRequest
	(*MKVResponse)(nil), // 3: pb.MKVResponse
}
var file_rpcserver_proto_depIdxs = []int32{
	0, // 0: pb.RPCServer.Set:input_type -> pb.KVRequest
	0, // 1: pb.RPCServer.Get:input_type -> pb.KVRequest
	0, // 2: pb.RPCServer.Del:input_type -> pb.KVRequest
	2, // 3: pb.RPCServer.Keys:input_type -> pb.MKVRequest
	2, // 4: pb.RPCServer.MGet:input_type -> pb.MKVRequest
	2, // 5: pb.RPCServer.MSet:input_type -> pb.MKVRequest
	1, // 6: pb.RPCServer.Set:output_type -> pb.KVResponse
	1, // 7: pb.RPCServer.Get:output_type -> pb.KVResponse
	1, // 8: pb.RPCServer.Del:output_type -> pb.KVResponse
	3, // 9: pb.RPCServer.Keys:output_type -> pb.MKVResponse
	3, // 10: pb.RPCServer.MGet:output_type -> pb.MKVResponse
	3, // 11: pb.RPCServer.MSet:output_type -> pb.MKVResponse
	6, // [6:12] is the sub-list for method output_type
	0, // [0:6] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_rpcserver_proto_init() }
func file_rpcserver_proto_init() {
	if File_rpcserver_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_rpcserver_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KVRequest); i {
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
		file_rpcserver_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KVResponse); i {
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
		file_rpcserver_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MKVRequest); i {
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
		file_rpcserver_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MKVResponse); i {
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
			RawDescriptor: file_rpcserver_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_rpcserver_proto_goTypes,
		DependencyIndexes: file_rpcserver_proto_depIdxs,
		MessageInfos:      file_rpcserver_proto_msgTypes,
	}.Build()
	File_rpcserver_proto = out.File
	file_rpcserver_proto_rawDesc = nil
	file_rpcserver_proto_goTypes = nil
	file_rpcserver_proto_depIdxs = nil
}
