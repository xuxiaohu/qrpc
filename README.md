### 安装

go get https://github.com/xuxiaohu/qrpc


### 使用

```
improt "github.com/xuxiaohu/qrpc"
在main方法的最上面
 qrpc.InitGConn(服务的地址和端口, 服务名 string)
 defer qrpc.Close()

```

#### 方法

1 qrpc.Log(内容 , 服务类型, 服务标志, 基本 string)

只需要在需要的时候调用

2 qrpc.HttpLog(内容 , 服务类型, 服务标志, 基本 string，, http请求的上下文 context.Context, http请求 *http.Request)

需要以下2步   
在发送http请求前调用，产生traceid   
在对应的http处理中调用，记录traceid


3 qrpc.RpcLog(内容 , 服务类型, 服务标志, 基本 string，, rpc请求的上下文 context.Context)
  使用比较复杂，有问题可以问我
 例子如下:
 
 ```
opentracing.InitGlobalTracer(qrpc.Tracer)
ctx := context.Background()
span := opentracing.StartSpan("good")
defer span.Finish()
ctx2 := opentracing.ContextWithSpan(ctx, span)
qlog.RpcLog("rpc1", "rpc1", "rpc1", "rpc1",  ctx2)
conn, err := grpc.Dial("localhost:3000",grpc.WithInsecure(), grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(qrpc.Tracer)))
if err != nil {
    fmt.Println(44444)
    fmt.Println(err)
}
c := pb.NewLogClient(conn)
rq, err:= c.Record(ctx2, &pb.LogRequest{ServiceType: "2123123", ServiceFlag: "2123123", Level: "2123123", Content: "2123123"})
if err != nil {
    fmt.Println(err)
}
fmt.Printf("Greeting: %s", rq.Status)