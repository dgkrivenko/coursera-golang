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
	StatCh          chan *Stat
	Middleware      *Middleware
	sync.Mutex
}

type BizService struct {
	UnimplementedBizServer
}

type Permissions struct {
	Data map[string][]string
}

type Middleware struct {
	Ch    chan *Event
	ChStats []chan *StatMessage
	sync.Mutex
}

type StatMessage struct {
	Method string
	Consumer string
}

func (m *Middleware) SendEvent(ctx context.Context, info *tap.Info) (context.Context, error) {
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
		m.Ch <- event
	}(consumerVal, info.FullMethodName, p.Addr.String())

	go func(consumer, method string) {
		msg := &StatMessage{
			Method: method,
			Consumer: consumer,
		}
		m.Lock()
		for _, c := range m.ChStats {
			c <- msg
		}
		m.Unlock()
	}(consumerVal, info.FullMethodName)

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
	as.Lock()
	as.OpenConnections = append(as.OpenConnections, server)
	as.Unlock()
	as.Wg.Wait()
	return nil
}

func (as *AdminService) Statistics(interval *StatInterval, server Admin_StatisticsServer) error {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	ch := make(chan *StatMessage, 10)
	as.Middleware.Lock()
	as.Middleware.ChStats = append(as.Middleware.ChStats, ch)
	as.Middleware.Unlock()
	st := &Stat{ByMethod: map[string]uint64{}, ByConsumer: map[string]uint64{}}

	mu := &sync.Mutex{}

	go func() {
		for {
			time.Sleep(time.Second * time.Duration(interval.IntervalSeconds))
			mu.Lock()
			err := server.Send(st)
			if err != nil {
				fmt.Println("Server err:", err)
				break
			}
			st.ByMethod = map[string]uint64{}
			st.ByConsumer = map[string]uint64{}
			mu.Unlock()
		}
		wg.Done()
	}()

	go func(interval *StatInterval, group *sync.WaitGroup, ch chan *StatMessage) {
		for {
			 msg := <- ch
			 mu.Lock()
			 st.ByMethod[msg.Method] += 1
			 st.ByConsumer[msg.Consumer] += 1
			 mu.Unlock()
		}
	}(interval, wg, ch)
	wg.Wait()
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
	mw := &Middleware{Ch: ch, ChStats: []chan *StatMessage{}}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(acl.checkUnaryPermission),
		grpc.StreamInterceptor(acl.checkStreamPermission),
		grpc.InTapHandle(mw.SendEvent),
	)
	as := NewAdminService(ch)
	as.Middleware = mw
	RegisterAdminServer(server, as)
	RegisterBizServer(server, NewBizService())

	// Middleware worker for logging
	quit := make(chan struct{})
	as.Wg.Add(1)
	go func(quit chan struct{}) {
		for {
			select {
			case <-quit:
				return
			case event := <-as.Ch:
				as.Lock()
				for _, srv := range as.OpenConnections {
					err := srv.Send(event)
					if err != nil {
						fmt.Println("Server err:", err)
					}
				}
				as.Unlock()
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
