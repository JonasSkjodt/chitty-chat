// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.24.3
// source: proto/template.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Chat_SendMessage_FullMethodName          = "/proto.Chat/SendMessage"
	Chat_ReceiveMessageStream_FullMethodName = "/proto.Chat/ReceiveMessageStream"
)

// ChatClient is the client API for Chat service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ChatClient interface {
	SendMessage(ctx context.Context, in *ChatMessage, opts ...grpc.CallOption) (*Ack, error)
	ReceiveMessageStream(ctx context.Context, in *ClientName, opts ...grpc.CallOption) (Chat_ReceiveMessageStreamClient, error)
}

type chatClient struct {
	cc grpc.ClientConnInterface
}

func NewChatClient(cc grpc.ClientConnInterface) ChatClient {
	return &chatClient{cc}
}

func (c *chatClient) SendMessage(ctx context.Context, in *ChatMessage, opts ...grpc.CallOption) (*Ack, error) {
	out := new(Ack)
	err := c.cc.Invoke(ctx, Chat_SendMessage_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chatClient) ReceiveMessageStream(ctx context.Context, in *ClientName, opts ...grpc.CallOption) (Chat_ReceiveMessageStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Chat_ServiceDesc.Streams[0], Chat_ReceiveMessageStream_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &chatReceiveMessageStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Chat_ReceiveMessageStreamClient interface {
	Recv() (*ChatMessage, error)
	grpc.ClientStream
}

type chatReceiveMessageStreamClient struct {
	grpc.ClientStream
}

func (x *chatReceiveMessageStreamClient) Recv() (*ChatMessage, error) {
	m := new(ChatMessage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ChatServer is the server API for Chat service.
// All implementations must embed UnimplementedChatServer
// for forward compatibility
type ChatServer interface {
	SendMessage(context.Context, *ChatMessage) (*Ack, error)
	ReceiveMessageStream(*ClientName, Chat_ReceiveMessageStreamServer) error
	mustEmbedUnimplementedChatServer()
}

// UnimplementedChatServer must be embedded to have forward compatible implementations.
type UnimplementedChatServer struct {
}

func (UnimplementedChatServer) SendMessage(context.Context, *ChatMessage) (*Ack, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMessage not implemented")
}
func (UnimplementedChatServer) ReceiveMessageStream(*ClientName, Chat_ReceiveMessageStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method ReceiveMessageStream not implemented")
}
func (UnimplementedChatServer) mustEmbedUnimplementedChatServer() {}

// UnsafeChatServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ChatServer will
// result in compilation errors.
type UnsafeChatServer interface {
	mustEmbedUnimplementedChatServer()
}

func RegisterChatServer(s grpc.ServiceRegistrar, srv ChatServer) {
	s.RegisterService(&Chat_ServiceDesc, srv)
}

func _Chat_SendMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChatMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChatServer).SendMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Chat_SendMessage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChatServer).SendMessage(ctx, req.(*ChatMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _Chat_ReceiveMessageStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ClientName)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ChatServer).ReceiveMessageStream(m, &chatReceiveMessageStreamServer{stream})
}

type Chat_ReceiveMessageStreamServer interface {
	Send(*ChatMessage) error
	grpc.ServerStream
}

type chatReceiveMessageStreamServer struct {
	grpc.ServerStream
}

func (x *chatReceiveMessageStreamServer) Send(m *ChatMessage) error {
	return x.ServerStream.SendMsg(m)
}

// Chat_ServiceDesc is the grpc.ServiceDesc for Chat service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Chat_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Chat",
	HandlerType: (*ChatServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendMessage",
			Handler:    _Chat_SendMessage_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ReceiveMessageStream",
			Handler:       _Chat_ReceiveMessageStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "proto/template.proto",
}
