package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"regexp"
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

// checkUnaryPermission - check permissions for unary methods
func (acl *Permissions) checkUnaryPermission(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	consumer := md.Get("consumer")
	if len(consumer) != 0 {
		allowedMethods := acl.Data[consumer[0]]
		for _, m := range allowedMethods {
			matched, err := regexp.Match(m, []byte(info.FullMethod))
			if err != nil {
				return nil, status.Error(codes.Internal, "internal server error")
			}
			if matched {
				return handler(ctx, req)
			}
		}
	}
	return nil, status.Error(codes.Unauthenticated, "unknown consumer")
}

// checkStreamPermission - check permissions for streaming method
func (acl *Permissions) checkStreamPermission(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler) error {
	md, _ := metadata.FromIncomingContext(ss.Context())
	consumer := md.Get("consumer")
	if len(consumer) != 0 {
		allowedMethods := acl.Data[consumer[0]]
		for _, m := range allowedMethods {
			matched, err := regexp.Match(m, []byte(info.FullMethod))
			if err != nil {
				return status.Error(codes.Internal, "internal server error")
			}
			if matched {
				return handler(srv, ss)
			}
		}
	}
	return status.Error(codes.Unauthenticated, "unknown consumer")
}

// ================================ Admin methods ================================

func NewAdminService() *AdminService {
	return &AdminService{}
}

func (as *AdminService) Logging(n *Nothing, server Admin_LoggingServer) error {
	// Добавить сервер в поле структуры
	// Если там единственный сервер запустить горутину
	go func() {
		for {
			// Ждем сообщение из канала
			// Обходим массив с серверами и в каждом делаем send
			break
		}
	}()
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
	return &Nothing{Dummy: false}, nil
}

func (bs *BizService) Add(ctx context.Context, n *Nothing) (*Nothing, error) {
	return &Nothing{Dummy: false}, nil
}

func (bs *BizService) Test(ctx context.Context, n *Nothing) (*Nothing, error) {
	return &Nothing{Dummy: false}, nil
}

// StartMyMicroservice - start server
func StartMyMicroservice(ctx context.Context, addr string, ACLData string) error {
	// Parse ACL data
	data := map[string][]string{}
	err := json.Unmarshal([]byte(ACLData), &data)
	if err != nil {
		return err
	}
	acl := Permissions{Data: data}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("cant listet port", err)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(acl.checkUnaryPermission),
		grpc.StreamInterceptor(acl.checkStreamPermission),
	)

	RegisterAdminServer(server, NewAdminService())
	RegisterBizServer(server, NewBizService())

	fmt.Println("starting server at :8082")
	go server.Serve(lis)
	go func() {
		<- ctx.Done()
		server.Stop()
	}()
	return nil
}
