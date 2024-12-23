// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.21.12
// source: chat/service/v1/ws.proto

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	Ws_BindMember_FullMethodName      = "/chat.service.v1.Ws/BindMember"
	Ws_BindGroup_FullMethodName       = "/chat.service.v1.Ws/BindGroup"
	Ws_CancelGroup_FullMethodName     = "/chat.service.v1.Ws/CancelGroup"
	Ws_SendMsg_FullMethodName         = "/chat.service.v1.Ws/SendMsg"
	Ws_DelGroupColl_FullMethodName    = "/chat.service.v1.Ws/DelGroupColl"
	Ws_GetGroupHistory_FullMethodName = "/chat.service.v1.Ws/GetGroupHistory"
)

// WsClient is the client API for Ws service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WsClient interface {
	BindMember(ctx context.Context, in *BindMemberRequest, opts ...grpc.CallOption) (*BindMemberReply, error)
	BindGroup(ctx context.Context, in *BindGroupRequest, opts ...grpc.CallOption) (*BindGroupReply, error)
	CancelGroup(ctx context.Context, in *CancelGroupRequest, opts ...grpc.CallOption) (*CancelGroupReply, error)
	SendMsg(ctx context.Context, in *SendMsgRequest, opts ...grpc.CallOption) (*SendMsgReply, error)
	DelGroupColl(ctx context.Context, in *DelGroupCollRequest, opts ...grpc.CallOption) (*DelGroupCollReply, error)
	GetGroupHistory(ctx context.Context, in *GetGroupHistoryRequest, opts ...grpc.CallOption) (*GetGroupHistoryReply, error)
}

type wsClient struct {
	cc grpc.ClientConnInterface
}

func NewWsClient(cc grpc.ClientConnInterface) WsClient {
	return &wsClient{cc}
}

func (c *wsClient) BindMember(ctx context.Context, in *BindMemberRequest, opts ...grpc.CallOption) (*BindMemberReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BindMemberReply)
	err := c.cc.Invoke(ctx, Ws_BindMember_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *wsClient) BindGroup(ctx context.Context, in *BindGroupRequest, opts ...grpc.CallOption) (*BindGroupReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BindGroupReply)
	err := c.cc.Invoke(ctx, Ws_BindGroup_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *wsClient) CancelGroup(ctx context.Context, in *CancelGroupRequest, opts ...grpc.CallOption) (*CancelGroupReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CancelGroupReply)
	err := c.cc.Invoke(ctx, Ws_CancelGroup_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *wsClient) SendMsg(ctx context.Context, in *SendMsgRequest, opts ...grpc.CallOption) (*SendMsgReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SendMsgReply)
	err := c.cc.Invoke(ctx, Ws_SendMsg_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *wsClient) DelGroupColl(ctx context.Context, in *DelGroupCollRequest, opts ...grpc.CallOption) (*DelGroupCollReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DelGroupCollReply)
	err := c.cc.Invoke(ctx, Ws_DelGroupColl_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *wsClient) GetGroupHistory(ctx context.Context, in *GetGroupHistoryRequest, opts ...grpc.CallOption) (*GetGroupHistoryReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetGroupHistoryReply)
	err := c.cc.Invoke(ctx, Ws_GetGroupHistory_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WsServer is the server API for Ws service.
// All implementations must embed UnimplementedWsServer
// for forward compatibility.
type WsServer interface {
	BindMember(context.Context, *BindMemberRequest) (*BindMemberReply, error)
	BindGroup(context.Context, *BindGroupRequest) (*BindGroupReply, error)
	CancelGroup(context.Context, *CancelGroupRequest) (*CancelGroupReply, error)
	SendMsg(context.Context, *SendMsgRequest) (*SendMsgReply, error)
	DelGroupColl(context.Context, *DelGroupCollRequest) (*DelGroupCollReply, error)
	GetGroupHistory(context.Context, *GetGroupHistoryRequest) (*GetGroupHistoryReply, error)
	mustEmbedUnimplementedWsServer()
}

// UnimplementedWsServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedWsServer struct{}

func (UnimplementedWsServer) BindMember(context.Context, *BindMemberRequest) (*BindMemberReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BindMember not implemented")
}
func (UnimplementedWsServer) BindGroup(context.Context, *BindGroupRequest) (*BindGroupReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BindGroup not implemented")
}
func (UnimplementedWsServer) CancelGroup(context.Context, *CancelGroupRequest) (*CancelGroupReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelGroup not implemented")
}
func (UnimplementedWsServer) SendMsg(context.Context, *SendMsgRequest) (*SendMsgReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMsg not implemented")
}
func (UnimplementedWsServer) DelGroupColl(context.Context, *DelGroupCollRequest) (*DelGroupCollReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DelGroupColl not implemented")
}
func (UnimplementedWsServer) GetGroupHistory(context.Context, *GetGroupHistoryRequest) (*GetGroupHistoryReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGroupHistory not implemented")
}
func (UnimplementedWsServer) mustEmbedUnimplementedWsServer() {}
func (UnimplementedWsServer) testEmbeddedByValue()            {}

// UnsafeWsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WsServer will
// result in compilation errors.
type UnsafeWsServer interface {
	mustEmbedUnimplementedWsServer()
}

func RegisterWsServer(s grpc.ServiceRegistrar, srv WsServer) {
	// If the following call pancis, it indicates UnimplementedWsServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Ws_ServiceDesc, srv)
}

func _Ws_BindMember_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BindMemberRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WsServer).BindMember(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Ws_BindMember_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WsServer).BindMember(ctx, req.(*BindMemberRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Ws_BindGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BindGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WsServer).BindGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Ws_BindGroup_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WsServer).BindGroup(ctx, req.(*BindGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Ws_CancelGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CancelGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WsServer).CancelGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Ws_CancelGroup_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WsServer).CancelGroup(ctx, req.(*CancelGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Ws_SendMsg_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendMsgRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WsServer).SendMsg(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Ws_SendMsg_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WsServer).SendMsg(ctx, req.(*SendMsgRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Ws_DelGroupColl_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DelGroupCollRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WsServer).DelGroupColl(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Ws_DelGroupColl_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WsServer).DelGroupColl(ctx, req.(*DelGroupCollRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Ws_GetGroupHistory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetGroupHistoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WsServer).GetGroupHistory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Ws_GetGroupHistory_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WsServer).GetGroupHistory(ctx, req.(*GetGroupHistoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Ws_ServiceDesc is the grpc.ServiceDesc for Ws service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Ws_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "chat.service.v1.Ws",
	HandlerType: (*WsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "BindMember",
			Handler:    _Ws_BindMember_Handler,
		},
		{
			MethodName: "BindGroup",
			Handler:    _Ws_BindGroup_Handler,
		},
		{
			MethodName: "CancelGroup",
			Handler:    _Ws_CancelGroup_Handler,
		},
		{
			MethodName: "SendMsg",
			Handler:    _Ws_SendMsg_Handler,
		},
		{
			MethodName: "DelGroupColl",
			Handler:    _Ws_DelGroupColl_Handler,
		},
		{
			MethodName: "GetGroupHistory",
			Handler:    _Ws_GetGroupHistory_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "chat/service/v1/ws.proto",
}
