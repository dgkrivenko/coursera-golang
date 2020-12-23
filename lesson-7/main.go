package main

import (
	"context"
	"google.golang.org/grpc/metadata"
	"log"
	"time"
)

var acl string = `{
	"logger":    ["/main.Admin/Logging"],
	"stat":      ["/main.Admin/Statistics"],
	"biz_user":  ["/main.Biz/Check", "/main.Biz/Add"],
	"biz_admin": ["/main.Biz/*"]
}`
func mainWait(amout int) {
	time.Sleep(time.Duration(amout) * 10 * time.Millisecond)
}


func getConsumer(consumerName string) context.Context {
	// ctx, _ := context.WithTimeout(context.Background(), time.Second)
	ctx := context.Background()
	md := metadata.Pairs(
		"consumer", consumerName,
	)
	return metadata.NewOutgoingContext(ctx, md)
}


// старт-стоп сервера
func TestStartStop() {
	ctx, finish := context.WithCancel(context.Background())
	err := StartMyMicroservice(ctx, "127.0.0.1:8082", acl)
	if err != nil {
		log.Fatalf("cant start server initial: %v", err)
	}
	mainWait(1)
	finish() // при вызове этой функции ваш сервер должен остановиться и освободить порт
	mainWait(1)
	// теперь проверим что вы освободили порт и мы можем стартовать сервер ещё раз
	ctx, finish = context.WithCancel(context.Background())
	err = StartMyMicroservice(ctx, "127.0.0.1:8082", acl)
	if err != nil {
		log.Fatalf("cant start server again: %v", err)
	}
	mainWait(1)
	finish()
	mainWait(1)
}

// TODO перепроверить в конце
//func TestLeak() {
//	goroutinesStart := runtime.NumGoroutine()
//	TestStartStop()
//	goroutinesPerTwoIterations := runtime.NumGoroutine() - goroutinesStart
//
//	goroutinesStart = runtime.NumGoroutine()
//	goroutinesStat := []int{}
//	for i := 0; i <= 25; i++ {
//		TestStartStop()
//		goroutinesStat = append(goroutinesStat, runtime.NumGoroutine())
//	}
//	goroutinesPerFiftyIterations := runtime.NumGoroutine() - goroutinesStart
//	if goroutinesPerFiftyIterations > goroutinesPerTwoIterations*5 {
//		log.Fatalf("looks like you have goroutines leak: %+v", goroutinesStat)
//	}
//}

func main() {
	//TestLeak()
	ctx, _ := context.WithCancel(context.Background())
	ctx = getConsumer("test")
	err := StartMyMicroservice(ctx, "127.0.0.1:8082", acl)
	if err != nil {
		log.Fatalf("cant start server initial: %v", err)
	}

	time.Sleep(time.Hour)

}
