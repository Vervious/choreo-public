// Code generated by protoc-gen-go. DO NOT EDIT.
// source: choreo.proto

/*
Package p2p is a generated protocol buffer package.

It is generated from these files:
	choreo.proto

It has these top-level messages:
	ReadStateRequest
	ReadStateReply
	WriteStateRequest
	WriteStateReply
*/
package p2p

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type ReadStateRequest struct {
	Address string `protobuf:"bytes,1,opt,name=address" json:"address,omitempty"`
}

func (m *ReadStateRequest) Reset()                    { *m = ReadStateRequest{} }
func (m *ReadStateRequest) String() string            { return proto.CompactTextString(m) }
func (*ReadStateRequest) ProtoMessage()               {}
func (*ReadStateRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *ReadStateRequest) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

type ReadStateReply struct {
	Address string `protobuf:"bytes,1,opt,name=address" json:"address,omitempty"`
	Value   []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *ReadStateReply) Reset()                    { *m = ReadStateReply{} }
func (m *ReadStateReply) String() string            { return proto.CompactTextString(m) }
func (*ReadStateReply) ProtoMessage()               {}
func (*ReadStateReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ReadStateReply) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *ReadStateReply) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

type WriteStateRequest struct {
	Address string `protobuf:"bytes,1,opt,name=address" json:"address,omitempty"`
	Value   []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *WriteStateRequest) Reset()                    { *m = WriteStateRequest{} }
func (m *WriteStateRequest) String() string            { return proto.CompactTextString(m) }
func (*WriteStateRequest) ProtoMessage()               {}
func (*WriteStateRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *WriteStateRequest) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *WriteStateRequest) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

type WriteStateReply struct {
}

func (m *WriteStateReply) Reset()                    { *m = WriteStateReply{} }
func (m *WriteStateReply) String() string            { return proto.CompactTextString(m) }
func (*WriteStateReply) ProtoMessage()               {}
func (*WriteStateReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func init() {
	proto.RegisterType((*ReadStateRequest)(nil), "p2p.ReadStateRequest")
	proto.RegisterType((*ReadStateReply)(nil), "p2p.ReadStateReply")
	proto.RegisterType((*WriteStateRequest)(nil), "p2p.WriteStateRequest")
	proto.RegisterType((*WriteStateReply)(nil), "p2p.WriteStateReply")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for State service

type StateClient interface {
	Get(ctx context.Context, in *ReadStateRequest, opts ...grpc.CallOption) (*ReadStateReply, error)
	Set(ctx context.Context, in *WriteStateRequest, opts ...grpc.CallOption) (*WriteStateReply, error)
}

type stateClient struct {
	cc *grpc.ClientConn
}

func NewStateClient(cc *grpc.ClientConn) StateClient {
	return &stateClient{cc}
}

func (c *stateClient) Get(ctx context.Context, in *ReadStateRequest, opts ...grpc.CallOption) (*ReadStateReply, error) {
	out := new(ReadStateReply)
	err := grpc.Invoke(ctx, "/p2p.State/Get", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *stateClient) Set(ctx context.Context, in *WriteStateRequest, opts ...grpc.CallOption) (*WriteStateReply, error) {
	out := new(WriteStateReply)
	err := grpc.Invoke(ctx, "/p2p.State/Set", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for State service

type StateServer interface {
	Get(context.Context, *ReadStateRequest) (*ReadStateReply, error)
	Set(context.Context, *WriteStateRequest) (*WriteStateReply, error)
}

func RegisterStateServer(s *grpc.Server, srv StateServer) {
	s.RegisterService(&_State_serviceDesc, srv)
}

func _State_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadStateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StateServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/p2p.State/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StateServer).Get(ctx, req.(*ReadStateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _State_Set_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WriteStateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StateServer).Set(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/p2p.State/Set",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StateServer).Set(ctx, req.(*WriteStateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _State_serviceDesc = grpc.ServiceDesc{
	ServiceName: "p2p.State",
	HandlerType: (*StateServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _State_Get_Handler,
		},
		{
			MethodName: "Set",
			Handler:    _State_Set_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "choreo.proto",
}

func init() { proto.RegisterFile("choreo.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 190 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x49, 0xce, 0xc8, 0x2f,
	0x4a, 0xcd, 0xd7, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2e, 0x30, 0x2a, 0x50, 0xd2, 0xe1,
	0x12, 0x08, 0x4a, 0x4d, 0x4c, 0x09, 0x2e, 0x49, 0x2c, 0x49, 0x0d, 0x4a, 0x2d, 0x2c, 0x4d, 0x2d,
	0x2e, 0x11, 0x92, 0xe0, 0x62, 0x4f, 0x4c, 0x49, 0x29, 0x4a, 0x2d, 0x2e, 0x96, 0x60, 0x54, 0x60,
	0xd4, 0xe0, 0x0c, 0x82, 0x71, 0x95, 0x1c, 0xb8, 0xf8, 0x90, 0x54, 0x17, 0xe4, 0x54, 0xe2, 0x56,
	0x2b, 0x24, 0xc2, 0xc5, 0x5a, 0x96, 0x98, 0x53, 0x9a, 0x2a, 0xc1, 0xa4, 0xc0, 0xa8, 0xc1, 0x13,
	0x04, 0xe1, 0x28, 0x39, 0x73, 0x09, 0x86, 0x17, 0x65, 0x96, 0xa4, 0x12, 0x67, 0x21, 0x0e, 0x43,
	0x04, 0xb9, 0xf8, 0x91, 0x0d, 0x29, 0xc8, 0xa9, 0x34, 0x2a, 0xe6, 0x62, 0x05, 0xf3, 0x84, 0x8c,
	0xb9, 0x98, 0xdd, 0x53, 0x4b, 0x84, 0x44, 0xf5, 0x0a, 0x8c, 0x0a, 0xf4, 0xd0, 0xbd, 0x26, 0x25,
	0x8c, 0x2e, 0x5c, 0x90, 0x53, 0xa9, 0xc4, 0x20, 0x64, 0xca, 0xc5, 0x1c, 0x9c, 0x5a, 0x22, 0x24,
	0x06, 0x96, 0xc5, 0x70, 0x9f, 0x94, 0x08, 0x86, 0x38, 0x58, 0x5b, 0x12, 0x1b, 0x38, 0x20, 0x8d,
	0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0xec, 0xb4, 0x18, 0x96, 0x58, 0x01, 0x00, 0x00,
}