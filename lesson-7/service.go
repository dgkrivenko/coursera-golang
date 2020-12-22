package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

type AdminService struct {
	UnimplementedAdminServer
}

type BizService struct {
	UnimplementedBizServer
}

type Permissions struct {
	Data map[string][]string
}

func (acl *Permissions) checkPermission(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	// TODO тут провести валидацию по данным полученным из контекста и того что есть в структуре
	fmt.Println("Check permissions", info.FullMethod)
	reply, err := handler(ctx, req)
	return reply, err
}

// ================================ Admin methods ================================


func NewAdminService() *AdminService {
	return &AdminService{}
}

func (as *AdminService) Logging(n *Nothing, server Admin_LoggingServer) error {
	return nil
}

func (as *AdminService) Statistics(interval *StatInterval, srv Admin_StatisticsServer) error {
	return nil
}

// ============================== Biz methods (mock) ==============================

func NewBizService() *BizService {
	return &BizService{}
}

func (bs *BizService) Check(ctx context.Context, n *Nothing) (*Nothing, error) {
	fmt.Println("Biz check working")
	return &Nothing{Dummy: false}, nil
}

func (bs *BizService) Add(ctx context.Context, n *Nothing) (*Nothing, error) {
	return nil, nil
}

func (bs *BizService) Test(ctx context.Context, n *Nothing) (*Nothing, error) {
	return nil, nil
}

// StartMyMicroservice - start server
func StartMyMicroservice(ctx context.Context, addr string, ACLData string) error {
	// TODO тут распарсить ACL и сделать валидацию json. Положить в структуру
	acl := Permissions{}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("cant listet port", err)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(acl.checkPermission),
	)

	RegisterAdminServer(server, NewAdminService())
	RegisterBizServer(server, NewBizService())

	fmt.Println("starting server at :8082")
	server.Serve(lis)
	return nil
}
