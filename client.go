package qrpc

import (
	"google.golang.org/grpc"
	pb "github.com/xuxiaohu/qrpc/protos"
	"log"
	"golang.org/x/net/context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"runtime"
	"net/http"
)

var (
	Gconn *grpc.ClientConn
	ZCollector zipkin.Collector
	Tracer opentracing.Tracer
	ZipkinIpPort  string
	QLogIpPort    string
)

func InitGConn(addPort, serverName string){
	collector, err := zipkin.NewHTTPCollector("http://localhost:9411/api/v1/spans")
	if err != nil {
		log.Fatal(err)
		return
	}
	ZCollector = collector

	tracer, err := zipkin.NewTracer(
		zipkin.NewRecorder(collector, false, addPort, serverName),
		zipkin.ClientServerSameSpan(true),
		zipkin.TraceID128Bit(true),
	)
	if err != nil {
		log.Fatal(err)
		return
	}
	Tracer = tracer
	opentracing.InitGlobalTracer(tracer)

	conn, err := grpc.Dial("localhost:50051",grpc.WithInsecure(), grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(tracer)))
	Gconn = conn
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
}

func InitGLogConn(addPort, serverName , zkIpPort, qlogIpPort string){
	zkUrl := ""
	if zkIpPort == "" {
		zkUrl = "http://localhost:9411/api/v1/spans"
	} else {
		zkUrl = zkIpPort + "/api/v1/spans"
	}
	collector, err := zipkin.NewHTTPCollector(zkUrl)
	if err != nil {
		log.Fatal(err)
		return
	}
	ZCollector = collector

	tracer, err := zipkin.NewTracer(
		zipkin.NewRecorder(collector, false, addPort, serverName),
		zipkin.ClientServerSameSpan(true),
		zipkin.TraceID128Bit(true),
	)
	if err != nil {
		log.Fatal(err)
		return
	}
	Tracer = tracer
	opentracing.InitGlobalTracer(tracer)

	qlogUrl := ""
	if zkIpPort == "" {
		qlogUrl = "localhost:50051"
	} else {
		qlogUrl = qlogIpPort
	}
	conn, err := grpc.Dial(qlogUrl,grpc.WithInsecure(), grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(tracer)))
	Gconn = conn
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
}

func Log(content string, serviceType, serviceFlag, level string){
	content = funcInfo() + content
	c := pb.NewLogClient(Gconn)

	r, err := c.Record(context.Background(), &pb.LogRequest{ServiceType: serviceType, ServiceFlag: serviceFlag, Level: level, Content: content})
	if err != nil {
		fmt.Println("could not greet2: ", err)
		//Close()
		//InitGConn()
		//Log(content)

	}
	log.Printf("Greeting: %s", r.Status)
}

func RpcLog(content string, serviceType, serviceFlag, level string, logCtx context.Context){
	content = funcInfo() + content
	c := pb.NewLogClient(Gconn)
	oldSpan := opentracing.SpanFromContext(logCtx)
	var span opentracing.Span
	if oldSpan == nil {
		span = opentracing.StartSpan(funcName())

	} else {
		span = opentracing.StartSpan(funcName(), opentracing.ChildOf(oldSpan.Context()))
	}
	defer span.Finish()
	ctx := opentracing.ContextWithSpan(logCtx, span)
	r, err := c.Record(ctx, &pb.LogRequest{ServiceType: serviceType, ServiceFlag: serviceFlag, Level: level, Content: content})
	if err != nil {
		fmt.Println("could not greet: %v", err)

	}
	log.Printf("Greeting: %s", r.Status)
}

func HttpLog(content string, serviceType, serviceFlag, level string, logCtx context.Context, req *http.Request){
	content = funcInfo() + content
	c := pb.NewLogClient(Gconn)
	var span opentracing.Span
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header))
	if err != nil {
		span = opentracing.StartSpan(funcName())
	} else {
		span = opentracing.StartSpan(funcName(), opentracing.ChildOf(wireContext))
	}
	defer span.Finish()
	ctx := opentracing.ContextWithSpan(logCtx, span)
	opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header))
	r, err := c.Record(ctx, &pb.LogRequest{ServiceType: serviceType, ServiceFlag: serviceFlag, Level: level, Content: content})
	if err != nil {
		fmt.Println("could not greet: %v", err)
                return
	}
	log.Printf("Greeting: %s", r.Status)
}

func Close(){
	defer Gconn.Close()
	defer ZCollector.Close()
}

func funcName() string {
	pc, _, _, _ := runtime.Caller(2)
	return runtime.FuncForPC(pc).Name()
}

func funcInfo() string {
	pc, _, line,_  := runtime.Caller(2)
	return fmt.Sprintf("%s:%d:",runtime.FuncForPC(pc).Name(), line)
}
