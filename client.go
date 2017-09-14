package qrpc

import (
	"google.golang.org/grpc"
	pb "github.com/xuxiaohu/qrpc/protos"
	"log"
	"golang.org/x/net/context"
	"fmt"
)
var (
	Gconn *grpc.ClientConn
)

func InitGConn(){
	conn, err := grpc.Dial("localhost:50051",grpc.WithInsecure())
	Gconn = conn
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
}

func Log(content string, serviceType, serviceFlag, level string){
	c := pb.NewLogClient(Gconn)

	r, err := c.Record(context.Background(), &pb.LogRequest{ServiceType: serviceType, ServiceFlag: serviceFlag, Level: level, Content: content})
	if err != nil {
		fmt.Println("could not greet: %v", err)
		//Close()
		//InitGConn()
		//Log(content)

	}
	log.Printf("Greeting: %s", r.Status)
}

func Close(){
	defer Gconn.Close()
}