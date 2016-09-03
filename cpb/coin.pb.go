// Code generated by protoc-gen-go.
// source: coin.proto
// DO NOT EDIT!

/*
Package cpb is a generated protocol buffer package.

It is generated from these files:
	coin.proto

It has these top-level messages:
	LoginRequest
	GetWorkRequest
	AnnounceRequest
	LoginReply
	GetWorkReply
	AnnounceReply
	Work
	Win
*/
package cpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
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

// The Login request message containing the user's name.
type LoginRequest struct {
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
}

func (m *LoginRequest) Reset()                    { *m = LoginRequest{} }
func (m *LoginRequest) String() string            { return proto.CompactTextString(m) }
func (*LoginRequest) ProtoMessage()               {}
func (*LoginRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// GetWork request carries the same name as login
type GetWorkRequest struct {
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
}

func (m *GetWorkRequest) Reset()                    { *m = GetWorkRequest{} }
func (m *GetWorkRequest) String() string            { return proto.CompactTextString(m) }
func (*GetWorkRequest) ProtoMessage()               {}
func (*GetWorkRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

// Announce request carries the same name as login
type AnnounceRequest struct {
	Win *Win `protobuf:"bytes,1,opt,name=win" json:"win,omitempty"`
}

func (m *AnnounceRequest) Reset()                    { *m = AnnounceRequest{} }
func (m *AnnounceRequest) String() string            { return proto.CompactTextString(m) }
func (*AnnounceRequest) ProtoMessage()               {}
func (*AnnounceRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *AnnounceRequest) GetWin() *Win {
	if m != nil {
		return m.Win
	}
	return nil
}

// Login response message containing the assigned id and work
type LoginReply struct {
	Id uint32 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
}

func (m *LoginReply) Reset()                    { *m = LoginReply{} }
func (m *LoginReply) String() string            { return proto.CompactTextString(m) }
func (*LoginReply) ProtoMessage()               {}
func (*LoginReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

// GetWork response is a work struct
type GetWorkReply struct {
	Work *Work `protobuf:"bytes,1,opt,name=work" json:"work,omitempty"`
}

func (m *GetWorkReply) Reset()                    { *m = GetWorkReply{} }
func (m *GetWorkReply) String() string            { return proto.CompactTextString(m) }
func (*GetWorkReply) ProtoMessage()               {}
func (*GetWorkReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *GetWorkReply) GetWork() *Work {
	if m != nil {
		return m.Work
	}
	return nil
}

// Announce response is boolean
type AnnounceReply struct {
	Ok bool `protobuf:"varint,1,opt,name=ok" json:"ok,omitempty"`
}

func (m *AnnounceReply) Reset()                    { *m = AnnounceReply{} }
func (m *AnnounceReply) String() string            { return proto.CompactTextString(m) }
func (*AnnounceReply) ProtoMessage()               {}
func (*AnnounceReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

type Work struct {
	Coinbase string `protobuf:"bytes,1,opt,name=coinbase" json:"coinbase,omitempty"`
	Block    []byte `protobuf:"bytes,2,opt,name=block,proto3" json:"block,omitempty"`
}

func (m *Work) Reset()                    { *m = Work{} }
func (m *Work) String() string            { return proto.CompactTextString(m) }
func (*Work) ProtoMessage()               {}
func (*Work) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type Win struct {
	Coinbase string `protobuf:"bytes,1,opt,name=coinbase" json:"coinbase,omitempty"`
	Nonce    uint32 `protobuf:"varint,2,opt,name=nonce" json:"nonce,omitempty"`
}

func (m *Win) Reset()                    { *m = Win{} }
func (m *Win) String() string            { return proto.CompactTextString(m) }
func (*Win) ProtoMessage()               {}
func (*Win) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func init() {
	proto.RegisterType((*LoginRequest)(nil), "cpb.LoginRequest")
	proto.RegisterType((*GetWorkRequest)(nil), "cpb.GetWorkRequest")
	proto.RegisterType((*AnnounceRequest)(nil), "cpb.AnnounceRequest")
	proto.RegisterType((*LoginReply)(nil), "cpb.LoginReply")
	proto.RegisterType((*GetWorkReply)(nil), "cpb.GetWorkReply")
	proto.RegisterType((*AnnounceReply)(nil), "cpb.AnnounceReply")
	proto.RegisterType((*Work)(nil), "cpb.Work")
	proto.RegisterType((*Win)(nil), "cpb.Win")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion3

// Client API for Coin service

type CoinClient interface {
	// very first message client -> server requests login details
	Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginReply, error)
	// this is a request to start mining
	GetWork(ctx context.Context, in *GetWorkRequest, opts ...grpc.CallOption) (*GetWorkReply, error)
	//
	Announce(ctx context.Context, in *AnnounceRequest, opts ...grpc.CallOption) (*AnnounceReply, error)
}

type coinClient struct {
	cc *grpc.ClientConn
}

func NewCoinClient(cc *grpc.ClientConn) CoinClient {
	return &coinClient{cc}
}

func (c *coinClient) Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginReply, error) {
	out := new(LoginReply)
	err := grpc.Invoke(ctx, "/cpb.Coin/Login", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coinClient) GetWork(ctx context.Context, in *GetWorkRequest, opts ...grpc.CallOption) (*GetWorkReply, error) {
	out := new(GetWorkReply)
	err := grpc.Invoke(ctx, "/cpb.Coin/GetWork", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coinClient) Announce(ctx context.Context, in *AnnounceRequest, opts ...grpc.CallOption) (*AnnounceReply, error) {
	out := new(AnnounceReply)
	err := grpc.Invoke(ctx, "/cpb.Coin/Announce", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Coin service

type CoinServer interface {
	// very first message client -> server requests login details
	Login(context.Context, *LoginRequest) (*LoginReply, error)
	// this is a request to start mining
	GetWork(context.Context, *GetWorkRequest) (*GetWorkReply, error)
	//
	Announce(context.Context, *AnnounceRequest) (*AnnounceReply, error)
}

func RegisterCoinServer(s *grpc.Server, srv CoinServer) {
	s.RegisterService(&_Coin_serviceDesc, srv)
}

func _Coin_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoinServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cpb.Coin/Login",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoinServer).Login(ctx, req.(*LoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Coin_GetWork_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetWorkRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoinServer).GetWork(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cpb.Coin/GetWork",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoinServer).GetWork(ctx, req.(*GetWorkRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Coin_Announce_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AnnounceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoinServer).Announce(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cpb.Coin/Announce",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoinServer).Announce(ctx, req.(*AnnounceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Coin_serviceDesc = grpc.ServiceDesc{
	ServiceName: "cpb.Coin",
	HandlerType: (*CoinServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Login",
			Handler:    _Coin_Login_Handler,
		},
		{
			MethodName: "GetWork",
			Handler:    _Coin_GetWork_Handler,
		},
		{
			MethodName: "Announce",
			Handler:    _Coin_Announce_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

func init() { proto.RegisterFile("coin.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 301 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x84, 0x92, 0x4f, 0x4f, 0xc2, 0x40,
	0x10, 0xc5, 0x6d, 0x29, 0x5a, 0x46, 0xfe, 0x84, 0x91, 0x03, 0x69, 0x34, 0x92, 0x8d, 0x07, 0x2e,
	0xed, 0x01, 0x12, 0xf5, 0x6a, 0x3c, 0x78, 0xf1, 0xb4, 0x97, 0x9e, 0xe9, 0xb2, 0x31, 0x9b, 0xd6,
	0x99, 0x5a, 0x4a, 0x08, 0x5f, 0xc7, 0x4f, 0x6a, 0xba, 0x6d, 0x11, 0x38, 0xe8, 0xad, 0xf3, 0xe6,
	0xf7, 0x66, 0x76, 0x5f, 0x17, 0x40, 0xb1, 0xa1, 0x28, 0x2f, 0xb8, 0x64, 0xec, 0xa8, 0x3c, 0x11,
	0x02, 0xfa, 0xef, 0xfc, 0x61, 0x48, 0xea, 0xaf, 0xad, 0xde, 0x94, 0x88, 0xe0, 0xd1, 0xea, 0x53,
	0x4f, 0x9d, 0x99, 0x33, 0xef, 0x49, 0xfb, 0x2d, 0x1e, 0x60, 0xf8, 0xa6, 0xcb, 0x98, 0x8b, 0xf4,
	0x2f, 0x2a, 0x84, 0xd1, 0x0b, 0x11, 0x6f, 0x49, 0xe9, 0x16, 0x0b, 0xa0, 0xb3, 0x33, 0x64, 0xa9,
	0xeb, 0x85, 0x1f, 0xa9, 0x3c, 0x89, 0x62, 0x43, 0xb2, 0x12, 0xc5, 0x2d, 0x40, 0xb3, 0x38, 0xcf,
	0xf6, 0x38, 0x04, 0xd7, 0xac, 0x2d, 0x38, 0x90, 0xae, 0x59, 0x8b, 0x10, 0xfa, 0x87, 0x95, 0x55,
	0xff, 0x0e, 0xbc, 0x1d, 0x17, 0x69, 0x33, 0xaa, 0x57, 0x8f, 0xaa, 0xba, 0x56, 0x16, 0xf7, 0x30,
	0xf8, 0xdd, 0xdd, 0xcc, 0xe3, 0x9a, 0xf6, 0xa5, 0xcb, 0xa9, 0x78, 0x06, 0xaf, 0xc2, 0x31, 0x00,
	0xbf, 0x4a, 0x20, 0x59, 0x6d, 0xda, 0xc3, 0x1f, 0x6a, 0x9c, 0x40, 0x37, 0xc9, 0x58, 0xa5, 0x53,
	0x77, 0xe6, 0xcc, 0xfb, 0xb2, 0x2e, 0xc4, 0x13, 0x74, 0x62, 0x43, 0xff, 0x19, 0x89, 0x49, 0x69,
	0x6b, 0x1c, 0xc8, 0xba, 0x58, 0x7c, 0x3b, 0xe0, 0xbd, 0xb2, 0x21, 0x0c, 0xa1, 0x6b, 0x6f, 0x8a,
	0x63, 0x7b, 0xec, 0xe3, 0xb8, 0x83, 0xd1, 0xb1, 0x94, 0x67, 0x7b, 0x71, 0x81, 0x4b, 0xb8, 0x6a,
	0xae, 0x8e, 0x37, 0xb6, 0x7b, 0x9a, 0x7d, 0x30, 0x3e, 0x15, 0x6b, 0xd3, 0x23, 0xf8, 0x6d, 0x00,
	0x38, 0xb1, 0xc0, 0xd9, 0xbf, 0x08, 0xf0, 0x4c, 0xb5, 0xbe, 0xe4, 0xd2, 0x3e, 0x85, 0xe5, 0x4f,
	0x00, 0x00, 0x00, 0xff, 0xff, 0x15, 0x22, 0xa8, 0xd9, 0x18, 0x02, 0x00, 0x00,
}