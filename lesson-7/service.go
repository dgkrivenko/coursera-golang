package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/tap"
	"log"
	"net"
	"regexp"
	"sync"
	"time"
)

type AdminService struct {
	UnimplementedAdminServer
	OpenConnections []Admin_LoggingServer
	Ch              chan *Event
	Wg              *sync.WaitGroup
	Stats           *Stat
}

type BizService struct {
	UnimplementedBizServer
}

type Permissions struct {
	Data map[string][]string
}

type Logger struct {
	Ch chan *Event
}

func (lg *Logger) SendEvent(ctx context.Context, info *tap.Info) (context.Context, error) {
	var consumerVal string
	md, _ := metadata.FromIncomingContext(ctx)
	p, _ := peer.FromContext(ctx)
	consumer := md.Get("consumer")
	if len(consumer) > 0 {
		consumerVal = consumer[0]
	}
	go func(consumer, method, host string) {
		event := &Event{
			Timestamp: time.Now().Unix(),
			Consumer:  consumer,
			Method:    method,
			Host:      host,
		}
		lg.Ch <- event
	}(consumerVal, info.FullMethodName, p.Addr.String())
	return ctx, nil
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

func NewAdminService(ch chan *Event) *AdminService {
	return &AdminService{Ch: ch, Wg: &sync.WaitGroup{}}
}

func (as *AdminService) Logging(n *Nothing, server Admin_LoggingServer) error {
	as.OpenConnections = append(as.OpenConnections, server)
	as.Wg.Wait()
	return nil
}

func (as *AdminService) Statistics(interval *StatInterval, srv Admin_StatisticsServer) error {
	go func(interval *StatInterval) {
		// по таймеру отправляем результат
	}(interval)
	return nil
}

// ============================== Biz methods ==============================

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
	ch := make(chan *Event, 10)
	logger := Logger{Ch: ch}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(acl.checkUnaryPermission),
		grpc.StreamInterceptor(acl.checkStreamPermission),
		grpc.InTapHandle(logger.SendEvent),
	)
	as := NewAdminService(ch)
	RegisterAdminServer(server, as)
	RegisterBizServer(server, NewBizService())

	quit := make(chan struct{})
	as.Wg.Add(1)
	go func(quit chan struct{}) {
		for {
			select {
			case <-quit:
				return
			case event := <-as.Ch:
				for _, srv := range as.OpenConnections {
					err := srv.Send(event)
					if err != nil {
						fmt.Println("Server err:", err)
					}
				}
			}
		}
	}(quit)

	fmt.Println("starting server at :8082")
	go server.Serve(lis)

	go func() {
		<-ctx.Done()
		quit <- struct{}{}
		server.Stop()
	}()
	return nil
}
