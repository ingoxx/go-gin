// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.2
// source: command.proto

package command

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
	FileTransferService_SendFile_FullMethodName = "/command.FileTransferService/SendFile"
)

// FileTransferServiceClient is the client API for FileTransferService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FileTransferServiceClient interface {
	SendFile(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[FileMessage, FileMessage], error)
}

type fileTransferServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewFileTransferServiceClient(cc grpc.ClientConnInterface) FileTransferServiceClient {
	return &fileTransferServiceClient{cc}
}

func (c *fileTransferServiceClient) SendFile(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[FileMessage, FileMessage], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &FileTransferService_ServiceDesc.Streams[0], FileTransferService_SendFile_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[FileMessage, FileMessage]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type FileTransferService_SendFileClient = grpc.BidiStreamingClient[FileMessage, FileMessage]

// FileTransferServiceServer is the server API for FileTransferService service.
// All implementations must embed UnimplementedFileTransferServiceServer
// for forward compatibility.
type FileTransferServiceServer interface {
	SendFile(grpc.BidiStreamingServer[FileMessage, FileMessage]) error
	mustEmbedUnimplementedFileTransferServiceServer()
}

// UnimplementedFileTransferServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedFileTransferServiceServer struct{}

func (UnimplementedFileTransferServiceServer) SendFile(grpc.BidiStreamingServer[FileMessage, FileMessage]) error {
	return status.Errorf(codes.Unimplemented, "method SendFile not implemented")
}
func (UnimplementedFileTransferServiceServer) mustEmbedUnimplementedFileTransferServiceServer() {}
func (UnimplementedFileTransferServiceServer) testEmbeddedByValue()                             {}

// UnsafeFileTransferServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FileTransferServiceServer will
// result in compilation errors.
type UnsafeFileTransferServiceServer interface {
	mustEmbedUnimplementedFileTransferServiceServer()
}

func RegisterFileTransferServiceServer(s grpc.ServiceRegistrar, srv FileTransferServiceServer) {
	// If the following call pancis, it indicates UnimplementedFileTransferServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&FileTransferService_ServiceDesc, srv)
}

func _FileTransferService_SendFile_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(FileTransferServiceServer).SendFile(&grpc.GenericServerStream[FileMessage, FileMessage]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type FileTransferService_SendFileServer = grpc.BidiStreamingServer[FileMessage, FileMessage]

// FileTransferService_ServiceDesc is the grpc.ServiceDesc for FileTransferService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FileTransferService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "command.FileTransferService",
	HandlerType: (*FileTransferServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SendFile",
			Handler:       _FileTransferService_SendFile_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "command.proto",
}

const (
	StreamUpdateProgramService_DockerUpdate_FullMethodName    = "/command.StreamUpdateProgramService/DockerUpdate"
	StreamUpdateProgramService_JavaUpdate_FullMethodName      = "/command.StreamUpdateProgramService/JavaUpdate"
	StreamUpdateProgramService_DockerReload_FullMethodName    = "/command.StreamUpdateProgramService/DockerReload"
	StreamUpdateProgramService_JavaReload_FullMethodName      = "/command.StreamUpdateProgramService/JavaReload"
	StreamUpdateProgramService_JavaUpdateLog_FullMethodName   = "/command.StreamUpdateProgramService/JavaUpdateLog"
	StreamUpdateProgramService_DockerUpdateLog_FullMethodName = "/command.StreamUpdateProgramService/DockerUpdateLog"
)

// StreamUpdateProgramServiceClient is the client API for StreamUpdateProgramService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type StreamUpdateProgramServiceClient interface {
	DockerUpdate(ctx context.Context, in *StreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamReply], error)
	JavaUpdate(ctx context.Context, in *StreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamReply], error)
	DockerReload(ctx context.Context, in *StreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamReply], error)
	JavaReload(ctx context.Context, in *StreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamReply], error)
	JavaUpdateLog(ctx context.Context, in *StreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamReply], error)
	DockerUpdateLog(ctx context.Context, in *StreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamReply], error)
}

type streamUpdateProgramServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewStreamUpdateProgramServiceClient(cc grpc.ClientConnInterface) StreamUpdateProgramServiceClient {
	return &streamUpdateProgramServiceClient{cc}
}

func (c *streamUpdateProgramServiceClient) DockerUpdate(ctx context.Context, in *StreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamReply], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &StreamUpdateProgramService_ServiceDesc.Streams[0], StreamUpdateProgramService_DockerUpdate_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[StreamRequest, StreamReply]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type StreamUpdateProgramService_DockerUpdateClient = grpc.ServerStreamingClient[StreamReply]

func (c *streamUpdateProgramServiceClient) JavaUpdate(ctx context.Context, in *StreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamReply], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &StreamUpdateProgramService_ServiceDesc.Streams[1], StreamUpdateProgramService_JavaUpdate_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[StreamRequest, StreamReply]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type StreamUpdateProgramService_JavaUpdateClient = grpc.ServerStreamingClient[StreamReply]

func (c *streamUpdateProgramServiceClient) DockerReload(ctx context.Context, in *StreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamReply], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &StreamUpdateProgramService_ServiceDesc.Streams[2], StreamUpdateProgramService_DockerReload_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[StreamRequest, StreamReply]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type StreamUpdateProgramService_DockerReloadClient = grpc.ServerStreamingClient[StreamReply]

func (c *streamUpdateProgramServiceClient) JavaReload(ctx context.Context, in *StreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamReply], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &StreamUpdateProgramService_ServiceDesc.Streams[3], StreamUpdateProgramService_JavaReload_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[StreamRequest, StreamReply]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type StreamUpdateProgramService_JavaReloadClient = grpc.ServerStreamingClient[StreamReply]

func (c *streamUpdateProgramServiceClient) JavaUpdateLog(ctx context.Context, in *StreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamReply], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &StreamUpdateProgramService_ServiceDesc.Streams[4], StreamUpdateProgramService_JavaUpdateLog_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[StreamRequest, StreamReply]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type StreamUpdateProgramService_JavaUpdateLogClient = grpc.ServerStreamingClient[StreamReply]

func (c *streamUpdateProgramServiceClient) DockerUpdateLog(ctx context.Context, in *StreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamReply], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &StreamUpdateProgramService_ServiceDesc.Streams[5], StreamUpdateProgramService_DockerUpdateLog_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[StreamRequest, StreamReply]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type StreamUpdateProgramService_DockerUpdateLogClient = grpc.ServerStreamingClient[StreamReply]

// StreamUpdateProgramServiceServer is the server API for StreamUpdateProgramService service.
// All implementations must embed UnimplementedStreamUpdateProgramServiceServer
// for forward compatibility.
type StreamUpdateProgramServiceServer interface {
	DockerUpdate(*StreamRequest, grpc.ServerStreamingServer[StreamReply]) error
	JavaUpdate(*StreamRequest, grpc.ServerStreamingServer[StreamReply]) error
	DockerReload(*StreamRequest, grpc.ServerStreamingServer[StreamReply]) error
	JavaReload(*StreamRequest, grpc.ServerStreamingServer[StreamReply]) error
	JavaUpdateLog(*StreamRequest, grpc.ServerStreamingServer[StreamReply]) error
	DockerUpdateLog(*StreamRequest, grpc.ServerStreamingServer[StreamReply]) error
	mustEmbedUnimplementedStreamUpdateProgramServiceServer()
}

// UnimplementedStreamUpdateProgramServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedStreamUpdateProgramServiceServer struct{}

func (UnimplementedStreamUpdateProgramServiceServer) DockerUpdate(*StreamRequest, grpc.ServerStreamingServer[StreamReply]) error {
	return status.Errorf(codes.Unimplemented, "method DockerUpdate not implemented")
}
func (UnimplementedStreamUpdateProgramServiceServer) JavaUpdate(*StreamRequest, grpc.ServerStreamingServer[StreamReply]) error {
	return status.Errorf(codes.Unimplemented, "method JavaUpdate not implemented")
}
func (UnimplementedStreamUpdateProgramServiceServer) DockerReload(*StreamRequest, grpc.ServerStreamingServer[StreamReply]) error {
	return status.Errorf(codes.Unimplemented, "method DockerReload not implemented")
}
func (UnimplementedStreamUpdateProgramServiceServer) JavaReload(*StreamRequest, grpc.ServerStreamingServer[StreamReply]) error {
	return status.Errorf(codes.Unimplemented, "method JavaReload not implemented")
}
func (UnimplementedStreamUpdateProgramServiceServer) JavaUpdateLog(*StreamRequest, grpc.ServerStreamingServer[StreamReply]) error {
	return status.Errorf(codes.Unimplemented, "method JavaUpdateLog not implemented")
}
func (UnimplementedStreamUpdateProgramServiceServer) DockerUpdateLog(*StreamRequest, grpc.ServerStreamingServer[StreamReply]) error {
	return status.Errorf(codes.Unimplemented, "method DockerUpdateLog not implemented")
}
func (UnimplementedStreamUpdateProgramServiceServer) mustEmbedUnimplementedStreamUpdateProgramServiceServer() {
}
func (UnimplementedStreamUpdateProgramServiceServer) testEmbeddedByValue() {}

// UnsafeStreamUpdateProgramServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to StreamUpdateProgramServiceServer will
// result in compilation errors.
type UnsafeStreamUpdateProgramServiceServer interface {
	mustEmbedUnimplementedStreamUpdateProgramServiceServer()
}

func RegisterStreamUpdateProgramServiceServer(s grpc.ServiceRegistrar, srv StreamUpdateProgramServiceServer) {
	// If the following call pancis, it indicates UnimplementedStreamUpdateProgramServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&StreamUpdateProgramService_ServiceDesc, srv)
}

func _StreamUpdateProgramService_DockerUpdate_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(StreamUpdateProgramServiceServer).DockerUpdate(m, &grpc.GenericServerStream[StreamRequest, StreamReply]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type StreamUpdateProgramService_DockerUpdateServer = grpc.ServerStreamingServer[StreamReply]

func _StreamUpdateProgramService_JavaUpdate_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(StreamUpdateProgramServiceServer).JavaUpdate(m, &grpc.GenericServerStream[StreamRequest, StreamReply]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type StreamUpdateProgramService_JavaUpdateServer = grpc.ServerStreamingServer[StreamReply]

func _StreamUpdateProgramService_DockerReload_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(StreamUpdateProgramServiceServer).DockerReload(m, &grpc.GenericServerStream[StreamRequest, StreamReply]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type StreamUpdateProgramService_DockerReloadServer = grpc.ServerStreamingServer[StreamReply]

func _StreamUpdateProgramService_JavaReload_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(StreamUpdateProgramServiceServer).JavaReload(m, &grpc.GenericServerStream[StreamRequest, StreamReply]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type StreamUpdateProgramService_JavaReloadServer = grpc.ServerStreamingServer[StreamReply]

func _StreamUpdateProgramService_JavaUpdateLog_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(StreamUpdateProgramServiceServer).JavaUpdateLog(m, &grpc.GenericServerStream[StreamRequest, StreamReply]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type StreamUpdateProgramService_JavaUpdateLogServer = grpc.ServerStreamingServer[StreamReply]

func _StreamUpdateProgramService_DockerUpdateLog_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(StreamUpdateProgramServiceServer).DockerUpdateLog(m, &grpc.GenericServerStream[StreamRequest, StreamReply]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type StreamUpdateProgramService_DockerUpdateLogServer = grpc.ServerStreamingServer[StreamReply]

// StreamUpdateProgramService_ServiceDesc is the grpc.ServiceDesc for StreamUpdateProgramService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var StreamUpdateProgramService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "command.StreamUpdateProgramService",
	HandlerType: (*StreamUpdateProgramServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "DockerUpdate",
			Handler:       _StreamUpdateProgramService_DockerUpdate_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "JavaUpdate",
			Handler:       _StreamUpdateProgramService_JavaUpdate_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "DockerReload",
			Handler:       _StreamUpdateProgramService_DockerReload_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "JavaReload",
			Handler:       _StreamUpdateProgramService_JavaReload_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "JavaUpdateLog",
			Handler:       _StreamUpdateProgramService_JavaUpdateLog_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "DockerUpdateLog",
			Handler:       _StreamUpdateProgramService_DockerUpdateLog_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "command.proto",
}
