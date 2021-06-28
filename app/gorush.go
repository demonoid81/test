package app

import (
	"context"
	"log"

	"github.com/appleboy/gorush/rpc/proto"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/grpc"
)

func (a *App) SendPush(address string, tokens []string, message string) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := proto.NewGorushClient(conn)

	r, err := c.Send(context.Background(), &proto.NotificationRequest{
		Platform: 2,
		Tokens:   tokens,
		Message:  message,
	})
	if err != nil {
		log.Println("could not greet: %v", err)
	}
	log.Printf("Success: %t\n", r.Success)
	log.Printf("Count: %d\n", r.Counts)
}

func (a *App) SendDataPush(address string, tokens []string, message string, data *structpb.Struct) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := proto.NewGorushClient(conn)

	r, err := c.Send(context.Background(), &proto.NotificationRequest{
		Platform: 2,
		Tokens:   tokens,
		Message:  message,
		Data:     data,
	})
	if err != nil {
		log.Println("could not greet: %v", err)
	}
	log.Printf("Success: %t\n", r.Success)
	log.Printf("Count: %d\n", r.Counts)
}
