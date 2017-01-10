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
	GetCancelRequest
	IssueBlockRequest
	GetResultRequest
	LoginReply
	GetWorkReply
	AnnounceReply
	GetCancelReply
	IssueBlockReply
	GetResultReply
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
	Time string `protobuf:"bytes,2,opt,name=time" json:"time,omitempty"`
	User uint32 `protobuf:"varint,3,opt,name=user" json:"user,omitempty"`
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

// Announce request is a Win struct - coinbase + nonce
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

// GetCancel requests carries the same name as login
type GetCancelRequest struct {
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
}

func (m *GetCancelRequest) Reset()                    { *m = GetCancelRequest{} }
func (m *GetCancelRequest) String() string            { return proto.CompactTextString(m) }
func (*GetCancelRequest) ProtoMessage()               {}
func (*GetCancelRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

// IssueBlock requests carries the string block
type IssueBlockRequest struct {
	Upper       []byte `protobuf:"bytes,1,opt,name=upper,proto3" json:"upper,omitempty"`
	Lower       []byte `protobuf:"bytes,2,opt,name=lower,proto3" json:"lower,omitempty"`
	Blockheight uint32 `protobuf:"varint,3,opt,name=blockheight" json:"blockheight,omitempty"`
	Block       []byte `protobuf:"bytes,4,opt,name=block,proto3" json:"block,omitempty"`
	Merkle      []byte `protobuf:"bytes,5,opt,name=merkle,proto3" json:"merkle,omitempty"`
	Bits        uint32 `protobuf:"varint,6,opt,name=bits" json:"bits,omitempty"`
	Server      string `protobuf:"bytes,7,opt,name=server" json:"server,omitempty"`
}

func (m *IssueBlockRequest) Reset()                    { *m = IssueBlockRequest{} }
func (m *IssueBlockRequest) String() string            { return proto.CompactTextString(m) }
func (*IssueBlockRequest) ProtoMessage()               {}
func (*IssueBlockRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

// GetResult requests carries the same name as login
type GetResultRequest struct {
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
}

func (m *GetResultRequest) Reset()                    { *m = GetResultRequest{} }
func (m *GetResultRequest) String() string            { return proto.CompactTextString(m) }
func (*GetResultRequest) ProtoMessage()               {}
func (*GetResultRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

// Login response message containing the assigned id and work
type LoginReply struct {
	Id uint32 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
}

func (m *LoginReply) Reset()                    { *m = LoginReply{} }
func (m *LoginReply) String() string            { return proto.CompactTextString(m) }
func (*LoginReply) ProtoMessage()               {}
func (*LoginReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

// GetWork response is a work struct
type GetWorkReply struct {
	Work *Work `protobuf:"bytes,1,opt,name=work" json:"work,omitempty"`
}

func (m *GetWorkReply) Reset()                    { *m = GetWorkReply{} }
func (m *GetWorkReply) String() string            { return proto.CompactTextString(m) }
func (*GetWorkReply) ProtoMessage()               {}
func (*GetWorkReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

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
func (*AnnounceReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

// GetCancel response is the canonical name of server // index of server
type GetCancelReply struct {
	Server string `protobuf:"bytes,1,opt,name=server" json:"server,omitempty"`
}

func (m *GetCancelReply) Reset()                    { *m = GetCancelReply{} }
func (m *GetCancelReply) String() string            { return proto.CompactTextString(m) }
func (*GetCancelReply) ProtoMessage()               {}
func (*GetCancelReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

// IssueBlock response is boolean
type IssueBlockReply struct {
	Ok bool `protobuf:"varint,1,opt,name=ok" json:"ok,omitempty"`
}

func (m *IssueBlockReply) Reset()                    { *m = IssueBlockReply{} }
func (m *IssueBlockReply) String() string            { return proto.CompactTextString(m) }
func (*IssueBlockReply) ProtoMessage()               {}
func (*IssueBlockReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

// GetResult response is the winner details + server name // index
type GetResultReply struct {
	Winner *Win   `protobuf:"bytes,1,opt,name=winner" json:"winner,omitempty"`
	Server string `protobuf:"bytes,2,opt,name=server" json:"server,omitempty"`
}

func (m *GetResultReply) Reset()                    { *m = GetResultReply{} }
func (m *GetResultReply) String() string            { return proto.CompactTextString(m) }
func (*GetResultReply) ProtoMessage()               {}
func (*GetResultReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *GetResultReply) GetWinner() *Win {
	if m != nil {
		return m.Winner
	}
	return nil
}

type Work struct {
	Coinbase []byte `protobuf:"bytes,1,opt,name=coinbase,proto3" json:"coinbase,omitempty"`
	Block    []byte `protobuf:"bytes,2,opt,name=block,proto3" json:"block,omitempty"`
	Skel     []byte `protobuf:"bytes,3,opt,name=skel,proto3" json:"skel,omitempty"`
	Bits     uint32 `protobuf:"varint,4,opt,name=bits" json:"bits,omitempty"`
	Share    uint32 `protobuf:"varint,5,opt,name=share" json:"share,omitempty"`
}

func (m *Work) Reset()                    { *m = Work{} }
func (m *Work) String() string            { return proto.CompactTextString(m) }
func (*Work) ProtoMessage()               {}
func (*Work) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

type Win struct {
	Block    []byte `protobuf:"bytes,1,opt,name=block,proto3" json:"block,omitempty"`
	Nonce    uint32 `protobuf:"varint,2,opt,name=nonce" json:"nonce,omitempty"`
	Identity string `protobuf:"bytes,3,opt,name=identity" json:"identity,omitempty"`
}

func (m *Win) Reset()                    { *m = Win{} }
func (m *Win) String() string            { return proto.CompactTextString(m) }
func (*Win) ProtoMessage()               {}
func (*Win) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

func init() {
	proto.RegisterType((*LoginRequest)(nil), "cpb.LoginRequest")
	proto.RegisterType((*GetWorkRequest)(nil), "cpb.GetWorkRequest")
	proto.RegisterType((*AnnounceRequest)(nil), "cpb.AnnounceRequest")
	proto.RegisterType((*GetCancelRequest)(nil), "cpb.GetCancelRequest")
	proto.RegisterType((*IssueBlockRequest)(nil), "cpb.IssueBlockRequest")
	proto.RegisterType((*GetResultRequest)(nil), "cpb.GetResultRequest")
	proto.RegisterType((*LoginReply)(nil), "cpb.LoginReply")
	proto.RegisterType((*GetWorkReply)(nil), "cpb.GetWorkReply")
	proto.RegisterType((*AnnounceReply)(nil), "cpb.AnnounceReply")
	proto.RegisterType((*GetCancelReply)(nil), "cpb.GetCancelReply")
	proto.RegisterType((*IssueBlockReply)(nil), "cpb.IssueBlockReply")
	proto.RegisterType((*GetResultReply)(nil), "cpb.GetResultReply")
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
	// GetWork is a request to start mining
	GetWork(ctx context.Context, in *GetWorkRequest, opts ...grpc.CallOption) (*GetWorkReply, error)
	// Announce is a request to accept a win discovery
	Announce(ctx context.Context, in *AnnounceRequest, opts ...grpc.CallOption) (*AnnounceReply, error)
	// GetCancel is a request for a cancellation of search
	GetCancel(ctx context.Context, in *GetCancelRequest, opts ...grpc.CallOption) (*GetCancelReply, error)
	// IssueBlock is an offer of search block data
	IssueBlock(ctx context.Context, in *IssueBlockRequest, opts ...grpc.CallOption) (*IssueBlockReply, error)
	// GetResult is a  request for a solution
	GetResult(ctx context.Context, in *GetResultRequest, opts ...grpc.CallOption) (*GetResultReply, error)
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

func (c *coinClient) GetCancel(ctx context.Context, in *GetCancelRequest, opts ...grpc.CallOption) (*GetCancelReply, error) {
	out := new(GetCancelReply)
	err := grpc.Invoke(ctx, "/cpb.Coin/GetCancel", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coinClient) IssueBlock(ctx context.Context, in *IssueBlockRequest, opts ...grpc.CallOption) (*IssueBlockReply, error) {
	out := new(IssueBlockReply)
	err := grpc.Invoke(ctx, "/cpb.Coin/IssueBlock", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coinClient) GetResult(ctx context.Context, in *GetResultRequest, opts ...grpc.CallOption) (*GetResultReply, error) {
	out := new(GetResultReply)
	err := grpc.Invoke(ctx, "/cpb.Coin/GetResult", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Coin service

type CoinServer interface {
	// very first message client -> server requests login details
	Login(context.Context, *LoginRequest) (*LoginReply, error)
	// GetWork is a request to start mining
	GetWork(context.Context, *GetWorkRequest) (*GetWorkReply, error)
	// Announce is a request to accept a win discovery
	Announce(context.Context, *AnnounceRequest) (*AnnounceReply, error)
	// GetCancel is a request for a cancellation of search
	GetCancel(context.Context, *GetCancelRequest) (*GetCancelReply, error)
	// IssueBlock is an offer of search block data
	IssueBlock(context.Context, *IssueBlockRequest) (*IssueBlockReply, error)
	// GetResult is a  request for a solution
	GetResult(context.Context, *GetResultRequest) (*GetResultReply, error)
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

func _Coin_GetCancel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCancelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoinServer).GetCancel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cpb.Coin/GetCancel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoinServer).GetCancel(ctx, req.(*GetCancelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Coin_IssueBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IssueBlockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoinServer).IssueBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cpb.Coin/IssueBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoinServer).IssueBlock(ctx, req.(*IssueBlockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Coin_GetResult_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetResultRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoinServer).GetResult(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cpb.Coin/GetResult",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoinServer).GetResult(ctx, req.(*GetResultRequest))
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
		{
			MethodName: "GetCancel",
			Handler:    _Coin_GetCancel_Handler,
		},
		{
			MethodName: "IssueBlock",
			Handler:    _Coin_IssueBlock_Handler,
		},
		{
			MethodName: "GetResult",
			Handler:    _Coin_GetResult_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

func init() { proto.RegisterFile("coin.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 559 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x7c, 0x54, 0xcd, 0x6e, 0xda, 0x4c,
	0x14, 0x8d, 0x8d, 0x21, 0x70, 0x03, 0xe1, 0x63, 0xc2, 0x17, 0x59, 0x56, 0xab, 0x52, 0xab, 0xaa,
	0xd8, 0xc0, 0x22, 0x91, 0x2a, 0x55, 0xea, 0xa6, 0xcd, 0x22, 0x6a, 0xd4, 0x6e, 0x66, 0xc3, 0x1a,
	0x9b, 0x51, 0x18, 0x61, 0x66, 0x5c, 0xff, 0x04, 0xf1, 0x6c, 0x7d, 0x8f, 0x3e, 0x4f, 0x35, 0x77,
	0xc6, 0xf6, 0x98, 0x54, 0xec, 0xe6, 0x1c, 0xdf, 0xff, 0x73, 0xaf, 0x01, 0x62, 0xc9, 0xc5, 0x32,
	0xcd, 0x64, 0x21, 0x49, 0x27, 0x4e, 0xa3, 0xf0, 0x09, 0x86, 0x3f, 0xe4, 0x33, 0x17, 0x94, 0xfd,
	0x2a, 0x59, 0x5e, 0x10, 0x02, 0x9e, 0x58, 0xef, 0x99, 0xef, 0xcc, 0x9c, 0xf9, 0x80, 0xe2, 0x5b,
	0x71, 0x05, 0xdf, 0x33, 0xdf, 0xd5, 0x9c, 0x7a, 0x2b, 0xae, 0xcc, 0x59, 0xe6, 0x77, 0x66, 0xce,
	0x7c, 0x44, 0xf1, 0x1d, 0x7e, 0x80, 0xeb, 0x47, 0x56, 0xac, 0x64, 0xb6, 0x3b, 0x13, 0x2d, 0x5c,
	0xc0, 0xf8, 0xab, 0x10, 0xb2, 0x14, 0x31, 0xab, 0xcc, 0x02, 0xe8, 0x1c, 0xb8, 0x40, 0xab, 0xab,
	0xbb, 0xfe, 0x32, 0x4e, 0xa3, 0xe5, 0x8a, 0x0b, 0xaa, 0xc8, 0xf0, 0x23, 0xfc, 0xf7, 0xc8, 0x8a,
	0x87, 0xb5, 0x88, 0x59, 0x72, 0x2e, 0xec, 0x6f, 0x07, 0x26, 0xdf, 0xf3, 0xbc, 0x64, 0xdf, 0x12,
	0x19, 0xd7, 0x05, 0x4c, 0xa1, 0x5b, 0xa6, 0x29, 0xcb, 0xd0, 0x74, 0x48, 0x35, 0x50, 0x6c, 0x22,
	0x0f, 0x2c, 0xc3, 0x8e, 0x86, 0x54, 0x03, 0x32, 0x83, 0xab, 0x48, 0xf9, 0x6e, 0x19, 0x7f, 0xde,
	0x16, 0xa6, 0x33, 0x9b, 0x52, 0x7e, 0x08, 0x7d, 0x4f, 0xfb, 0x21, 0x20, 0xb7, 0xd0, 0xdb, 0xb3,
	0x6c, 0x97, 0x30, 0xbf, 0x8b, 0xb4, 0x41, 0xaa, 0xca, 0x88, 0x17, 0xb9, 0xdf, 0xd3, 0x23, 0x52,
	0x6f, 0x65, 0x9b, 0xb3, 0xec, 0x85, 0x65, 0xfe, 0x25, 0xd6, 0x6e, 0x90, 0xe9, 0x92, 0xb2, 0xbc,
	0x4c, 0x8a, 0x73, 0x5d, 0xbe, 0x01, 0x30, 0x72, 0xa5, 0xc9, 0x91, 0x5c, 0x83, 0xcb, 0x37, 0xf8,
	0x7d, 0x44, 0x5d, 0xbe, 0x09, 0x17, 0x30, 0xac, 0x05, 0x50, 0xdf, 0xdf, 0x82, 0x77, 0x90, 0xd9,
	0xce, 0x0c, 0x76, 0xa0, 0x07, 0xab, 0xbe, 0x22, 0x1d, 0xbe, 0x83, 0x51, 0xa3, 0x84, 0x89, 0x27,
	0xb5, 0x75, 0x9f, 0xba, 0x72, 0x17, 0xce, 0x51, 0xd0, 0x6a, 0xf6, 0xca, 0xa2, 0xa9, 0xdf, 0x69,
	0xd5, 0xff, 0x1e, 0xc6, 0xf6, 0xf0, 0xff, 0x15, 0xec, 0x09, 0x83, 0x55, 0x2d, 0x2a, 0x8b, 0x19,
	0xf4, 0x0e, 0x5c, 0x08, 0x13, 0xcc, 0x56, 0xde, 0xf0, 0x56, 0x3a, 0xb7, 0x95, 0xee, 0x05, 0x3c,
	0xd5, 0x07, 0x09, 0xa0, 0xaf, 0x16, 0x3a, 0x5a, 0xe7, 0xcc, 0x28, 0x5c, 0xe3, 0x46, 0x2c, 0xd7,
	0x16, 0x8b, 0x80, 0x97, 0xef, 0x58, 0x82, 0xea, 0x0e, 0x29, 0xbe, 0x6b, 0xa1, 0x3c, 0x4b, 0xa8,
	0x29, 0x74, 0xf3, 0xed, 0x3a, 0xd3, 0x9a, 0x8e, 0xa8, 0x06, 0xe1, 0x4f, 0xe8, 0xac, 0xb8, 0x68,
	0x42, 0x3b, 0x76, 0xe8, 0x29, 0x74, 0x85, 0x14, 0xb1, 0xbe, 0x93, 0x11, 0xd5, 0x40, 0x95, 0xc8,
	0x37, 0x4c, 0x14, 0xbc, 0x38, 0x62, 0xd2, 0x01, 0xad, 0xf1, 0xdd, 0x1f, 0x17, 0xbc, 0x07, 0xc9,
	0x05, 0x59, 0x40, 0x17, 0x65, 0x25, 0x13, 0x1c, 0x81, 0x7d, 0x91, 0xc1, 0xd8, 0xa6, 0xd2, 0xe4,
	0x18, 0x5e, 0x90, 0x7b, 0xb8, 0x34, 0x3a, 0x93, 0x1b, 0xfc, 0xda, 0x3e, 0xbb, 0x60, 0xd2, 0x26,
	0xb5, 0xd3, 0x27, 0xe8, 0x57, 0x6a, 0x93, 0x29, 0x1a, 0x9c, 0x9c, 0x61, 0x40, 0x4e, 0x58, 0xed,
	0xf7, 0x19, 0x06, 0xf5, 0x12, 0x90, 0xff, 0xab, 0xc8, 0xad, 0x83, 0x0c, 0x6e, 0x4e, 0x69, 0xed,
	0xfa, 0x05, 0xa0, 0xd9, 0x0a, 0x72, 0x8b, 0x46, 0xaf, 0x6e, 0x34, 0x98, 0xbe, 0xe2, 0xed, 0xc4,
	0x7a, 0x61, 0x9a, 0xc4, 0xad, 0x1b, 0x69, 0x12, 0x5b, 0x7b, 0x15, 0x5e, 0x44, 0x3d, 0xfc, 0xc3,
	0xdd, 0xff, 0x0d, 0x00, 0x00, 0xff, 0xff, 0x6b, 0x34, 0x31, 0x1f, 0xef, 0x04, 0x00, 0x00,
}
