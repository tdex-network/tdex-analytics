// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: tdexa/v1/analytics.proto

package tdexav1

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

// AnalyticsClient is the client API for Analytics service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AnalyticsClient interface {
	// returns all markets and its balances in time series
	MarketsBalances(ctx context.Context, in *MarketsBalancesRequest, opts ...grpc.CallOption) (*MarketsBalancesReply, error)
	// returns all markets and its prices in time series
	MarketsPrices(ctx context.Context, in *MarketsPricesRequest, opts ...grpc.CallOption) (*MarketsPricesReply, error)
	// return market id's to be used, if needed, as filter for MarketsBalances/MarketsPrices rpcs
	ListMarkets(ctx context.Context, in *ListMarketsRequest, opts ...grpc.CallOption) (*ListMarketsReply, error)
}

type analyticsClient struct {
	cc grpc.ClientConnInterface
}

func NewAnalyticsClient(cc grpc.ClientConnInterface) AnalyticsClient {
	return &analyticsClient{cc}
}

func (c *analyticsClient) MarketsBalances(ctx context.Context, in *MarketsBalancesRequest, opts ...grpc.CallOption) (*MarketsBalancesReply, error) {
	out := new(MarketsBalancesReply)
	err := c.cc.Invoke(ctx, "/tdexa.v1.Analytics/MarketsBalances", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *analyticsClient) MarketsPrices(ctx context.Context, in *MarketsPricesRequest, opts ...grpc.CallOption) (*MarketsPricesReply, error) {
	out := new(MarketsPricesReply)
	err := c.cc.Invoke(ctx, "/tdexa.v1.Analytics/MarketsPrices", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *analyticsClient) ListMarkets(ctx context.Context, in *ListMarketsRequest, opts ...grpc.CallOption) (*ListMarketsReply, error) {
	out := new(ListMarketsReply)
	err := c.cc.Invoke(ctx, "/tdexa.v1.Analytics/ListMarkets", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AnalyticsServer is the server API for Analytics service.
// All implementations should embed UnimplementedAnalyticsServer
// for forward compatibility
type AnalyticsServer interface {
	// returns all markets and its balances in time series
	MarketsBalances(context.Context, *MarketsBalancesRequest) (*MarketsBalancesReply, error)
	// returns all markets and its prices in time series
	MarketsPrices(context.Context, *MarketsPricesRequest) (*MarketsPricesReply, error)
	// return market id's to be used, if needed, as filter for MarketsBalances/MarketsPrices rpcs
	ListMarkets(context.Context, *ListMarketsRequest) (*ListMarketsReply, error)
}

// UnimplementedAnalyticsServer should be embedded to have forward compatible implementations.
type UnimplementedAnalyticsServer struct {
}

func (UnimplementedAnalyticsServer) MarketsBalances(context.Context, *MarketsBalancesRequest) (*MarketsBalancesReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MarketsBalances not implemented")
}
func (UnimplementedAnalyticsServer) MarketsPrices(context.Context, *MarketsPricesRequest) (*MarketsPricesReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MarketsPrices not implemented")
}
func (UnimplementedAnalyticsServer) ListMarkets(context.Context, *ListMarketsRequest) (*ListMarketsReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListMarkets not implemented")
}

// UnsafeAnalyticsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AnalyticsServer will
// result in compilation errors.
type UnsafeAnalyticsServer interface {
	mustEmbedUnimplementedAnalyticsServer()
}

func RegisterAnalyticsServer(s grpc.ServiceRegistrar, srv AnalyticsServer) {
	s.RegisterService(&Analytics_ServiceDesc, srv)
}

func _Analytics_MarketsBalances_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MarketsBalancesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnalyticsServer).MarketsBalances(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tdexa.v1.Analytics/MarketsBalances",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnalyticsServer).MarketsBalances(ctx, req.(*MarketsBalancesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Analytics_MarketsPrices_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MarketsPricesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnalyticsServer).MarketsPrices(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tdexa.v1.Analytics/MarketsPrices",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnalyticsServer).MarketsPrices(ctx, req.(*MarketsPricesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Analytics_ListMarkets_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListMarketsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnalyticsServer).ListMarkets(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tdexa.v1.Analytics/ListMarkets",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnalyticsServer).ListMarkets(ctx, req.(*ListMarketsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Analytics_ServiceDesc is the grpc.ServiceDesc for Analytics service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Analytics_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tdexa.v1.Analytics",
	HandlerType: (*AnalyticsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "MarketsBalances",
			Handler:    _Analytics_MarketsBalances_Handler,
		},
		{
			MethodName: "MarketsPrices",
			Handler:    _Analytics_MarketsPrices_Handler,
		},
		{
			MethodName: "ListMarkets",
			Handler:    _Analytics_ListMarkets_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tdexa/v1/analytics.proto",
}
