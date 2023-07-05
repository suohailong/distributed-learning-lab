// Code generated by protoc-gen-go. DO NOT EDIT.
// source: api/v1/harmoniakv.proto

package v1

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Object struct {
	Key                  []byte   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value                []byte   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Object) Reset()         { *m = Object{} }
func (m *Object) String() string { return proto.CompactTextString(m) }
func (*Object) ProtoMessage()    {}
func (*Object) Descriptor() ([]byte, []int) {
	return fileDescriptor_d5ef6af278ddee73, []int{0}
}

func (m *Object) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Object.Unmarshal(m, b)
}
func (m *Object) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Object.Marshal(b, m, deterministic)
}
func (m *Object) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Object.Merge(m, src)
}
func (m *Object) XXX_Size() int {
	return xxx_messageInfo_Object.Size(m)
}
func (m *Object) XXX_DiscardUnknown() {
	xxx_messageInfo_Object.DiscardUnknown(m)
}

var xxx_messageInfo_Object proto.InternalMessageInfo

func (m *Object) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *Object) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

type GetRequest struct {
	Key                  []byte   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetRequest) Reset()         { *m = GetRequest{} }
func (m *GetRequest) String() string { return proto.CompactTextString(m) }
func (*GetRequest) ProtoMessage()    {}
func (*GetRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d5ef6af278ddee73, []int{1}
}

func (m *GetRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetRequest.Unmarshal(m, b)
}
func (m *GetRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetRequest.Marshal(b, m, deterministic)
}
func (m *GetRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetRequest.Merge(m, src)
}
func (m *GetRequest) XXX_Size() int {
	return xxx_messageInfo_GetRequest.Size(m)
}
func (m *GetRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetRequest proto.InternalMessageInfo

func (m *GetRequest) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

type GetResponse struct {
	Objects              []*Object `protobuf:"bytes,1,rep,name=objects,proto3" json:"objects,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *GetResponse) Reset()         { *m = GetResponse{} }
func (m *GetResponse) String() string { return proto.CompactTextString(m) }
func (*GetResponse) ProtoMessage()    {}
func (*GetResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_d5ef6af278ddee73, []int{2}
}

func (m *GetResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetResponse.Unmarshal(m, b)
}
func (m *GetResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetResponse.Marshal(b, m, deterministic)
}
func (m *GetResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetResponse.Merge(m, src)
}
func (m *GetResponse) XXX_Size() int {
	return xxx_messageInfo_GetResponse.Size(m)
}
func (m *GetResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetResponse proto.InternalMessageInfo

func (m *GetResponse) GetObjects() []*Object {
	if m != nil {
		return m.Objects
	}
	return nil
}

type PutRequest struct {
	Key                  []byte   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Context              []byte   `protobuf:"bytes,2,opt,name=context,proto3" json:"context,omitempty"`
	Value                []byte   `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PutRequest) Reset()         { *m = PutRequest{} }
func (m *PutRequest) String() string { return proto.CompactTextString(m) }
func (*PutRequest) ProtoMessage()    {}
func (*PutRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d5ef6af278ddee73, []int{3}
}

func (m *PutRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PutRequest.Unmarshal(m, b)
}
func (m *PutRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PutRequest.Marshal(b, m, deterministic)
}
func (m *PutRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PutRequest.Merge(m, src)
}
func (m *PutRequest) XXX_Size() int {
	return xxx_messageInfo_PutRequest.Size(m)
}
func (m *PutRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PutRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PutRequest proto.InternalMessageInfo

func (m *PutRequest) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *PutRequest) GetContext() []byte {
	if m != nil {
		return m.Context
	}
	return nil
}

func (m *PutRequest) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

type PutResponse struct {
	Success              bool     `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PutResponse) Reset()         { *m = PutResponse{} }
func (m *PutResponse) String() string { return proto.CompactTextString(m) }
func (*PutResponse) ProtoMessage()    {}
func (*PutResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_d5ef6af278ddee73, []int{4}
}

func (m *PutResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PutResponse.Unmarshal(m, b)
}
func (m *PutResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PutResponse.Marshal(b, m, deterministic)
}
func (m *PutResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PutResponse.Merge(m, src)
}
func (m *PutResponse) XXX_Size() int {
	return xxx_messageInfo_PutResponse.Size(m)
}
func (m *PutResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PutResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PutResponse proto.InternalMessageInfo

func (m *PutResponse) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func init() {
	proto.RegisterType((*Object)(nil), "Object")
	proto.RegisterType((*GetRequest)(nil), "GetRequest")
	proto.RegisterType((*GetResponse)(nil), "GetResponse")
	proto.RegisterType((*PutRequest)(nil), "PutRequest")
	proto.RegisterType((*PutResponse)(nil), "PutResponse")
}

func init() { proto.RegisterFile("api/v1/harmoniakv.proto", fileDescriptor_d5ef6af278ddee73) }

var fileDescriptor_d5ef6af278ddee73 = []byte{
	// 261 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x91, 0x31, 0x4f, 0xc3, 0x30,
	0x10, 0x85, 0x15, 0x2c, 0x1a, 0x74, 0xc9, 0x80, 0x2c, 0x24, 0xac, 0x0e, 0x28, 0x64, 0xa1, 0x42,
	0xd4, 0x29, 0xe5, 0x1f, 0xb0, 0x94, 0x09, 0xa2, 0x8c, 0x6c, 0x4e, 0x7a, 0x02, 0xd3, 0x60, 0x87,
	0xd8, 0x8e, 0xe0, 0xdf, 0xa3, 0x3a, 0x0d, 0xc9, 0x40, 0x37, 0xbf, 0xb3, 0xfd, 0xde, 0x77, 0x77,
	0x70, 0x29, 0x1a, 0x99, 0x75, 0xf7, 0xd9, 0xbb, 0x68, 0x3f, 0xb5, 0x92, 0x62, 0xd7, 0xf1, 0xa6,
	0xd5, 0x56, 0xa7, 0x2b, 0x98, 0xbd, 0x94, 0x1f, 0x58, 0x59, 0x7a, 0x0e, 0x64, 0x87, 0x3f, 0x2c,
	0x48, 0x82, 0x45, 0x5c, 0xec, 0x8f, 0xf4, 0x02, 0x4e, 0x3b, 0x51, 0x3b, 0x64, 0x27, 0xbe, 0xd6,
	0x8b, 0xf4, 0x0a, 0x60, 0x83, 0xb6, 0xc0, 0x2f, 0x87, 0xe6, 0x9f, 0x5f, 0xe9, 0x0a, 0x22, 0x7f,
	0x6f, 0x1a, 0xad, 0x0c, 0xd2, 0x6b, 0x08, 0xb5, 0x0f, 0x30, 0x2c, 0x48, 0xc8, 0x22, 0x5a, 0x87,
	0xbc, 0x0f, 0x2c, 0x86, 0x7a, 0xfa, 0x0c, 0x90, 0xbb, 0xe3, 0x8e, 0x94, 0x41, 0x58, 0x69, 0x65,
	0xf1, 0xdb, 0x1e, 0x48, 0x06, 0x39, 0x12, 0x92, 0x29, 0xe1, 0x0d, 0x44, 0xde, 0xef, 0x40, 0xc0,
	0x20, 0x34, 0xae, 0xaa, 0xd0, 0x18, 0x6f, 0x7a, 0x56, 0x0c, 0x72, 0x9d, 0x03, 0x3c, 0xfd, 0x0d,
	0x84, 0x26, 0x40, 0x36, 0x68, 0x69, 0xc4, 0xc7, 0xf6, 0xe6, 0x31, 0x9f, 0xf6, 0x92, 0x00, 0xc9,
	0xdd, 0xfe, 0xc5, 0x88, 0x3b, 0x8f, 0xf9, 0x24, 0xeb, 0xf1, 0xee, 0xf5, 0x76, 0x2b, 0x8d, 0x6d,
	0x65, 0xe9, 0x2c, 0x6e, 0x97, 0x35, 0x8a, 0x56, 0x49, 0xf5, 0xb6, 0xac, 0x45, 0x39, 0x99, 0x7d,
	0xd6, 0x6f, 0xa3, 0x9c, 0xf9, 0x1d, 0x3c, 0xfc, 0x06, 0x00, 0x00, 0xff, 0xff, 0x98, 0x99, 0xb6,
	0x2c, 0x9e, 0x01, 0x00, 0x00,
}