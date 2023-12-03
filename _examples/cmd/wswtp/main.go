package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"

	_ "golang.org/x/net/trace"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/tmc/grpc-websocket-proxy/examples/cmd/wswtp/wtpserver/proto/external/wgtwo/webterminal/v0"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	grpcAddr  = flag.String("grpcaddr", ":8001", "listen grpc addr")
	httpAddr  = flag.String("addr", ":8000", "listen http addr")
	debugAddr = flag.String("debugaddr", ":8002", "listen debug addr")
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// if err := listenGRPC(*grpcAddr); err != nil {
	//	return err
	//}

	mux := runtime.NewServeMux()
	// opts := []grpc.DialOption{grpc.WithInsecure()}

	conf := credentials.NewTLS(&tls.Config{InsecureSkipVerify: false})
	opts := []grpc.DialOption{grpc.WithTransportCredentials(conf)}

	err := v0.RegisterWebTerminalServiceHandlerFromEndpoint(ctx, mux, net.JoinHostPort("api.wgtwo.com", "443"), opts)
	if err != nil {
		return err
	}
	go http.ListenAndServe(*debugAddr, nil)
	fmt.Println("listening")
	http.ListenAndServe(*httpAddr, wsproxy.WebsocketProxy(mux))
	return nil
}

func main() {
	flag.Parse()
	defer glog.Flush()

	if err := run(); err != nil {
		glog.Fatal(err)
	}
}
