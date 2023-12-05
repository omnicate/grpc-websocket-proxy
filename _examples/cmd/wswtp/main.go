package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "github.com/sirupsen/logrus"
	"github.com/tmc/grpc-websocket-proxy/examples/cmd/wswtp/wtpserver/proto/external/wgtwo/webterminal/v0"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	"golang.org/x/net/context"
	_ "golang.org/x/net/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcMetadata "google.golang.org/grpc/metadata"
)

var (
	grpcAddr    = flag.String("grpcaddr", ":8001", "listen grpc addr")
	httpAddr    = flag.String("addr", ":8000", "listen http addr")
	debugAddr   = flag.String("debugaddr", ":8002", "listen debug addr")
	accessToken = flag.String("wg2_token", "token", "WG2 api access tocken")
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := listenGRPC(*grpcAddr); err != nil {
		return err
	}

	mux := runtime.NewServeMux()
	// opts := []grpc.DialOption{grpc.WithInsecure()}

	conf := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	opts := []grpc.DialOption{grpc.WithTransportCredentials(conf), grpc.WithBlock()}
	// Add token to gRPC Request.
	ctx = grpcMetadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+*accessToken)
	ctx = grpcMetadata.AppendToOutgoingContext(ctx, "wg2-msisdn", "4799999120")

	err := v0.RegisterWebTerminalServiceHandlerFromEndpoint(ctx, mux, "api-gateway.dub.dev.wgtwo.com:443", opts)
	//opts = []grpc.DialOption{grpc.WithInsecure()}
	//err := v0.RegisterWebTerminalServiceHandlerFromEndpoint(ctx, mux, *grpcAddr, opts)
	if err != nil {
		return err
	}
	go http.ListenAndServe(*debugAddr, nil)
	fmt.Println("listening")

	l := log.New()
	l.SetLevel(log.DebugLevel)
	l.Debugln("hei from debug logger")

	http.ListenAndServe(*httpAddr, wsproxy.WebsocketProxy(mux, wsproxy.WithLogger(l)))
	return nil
}

func listenGRPC(listenAddr string) error {
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	v0.RegisterWebTerminalServiceServer(grpcServer, &Server{})
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Println("serveGRPC err:", err)
		}
	}()
	return nil

}
func main() {
	flag.Parse()
	defer glog.Flush()

	if err := run(); err != nil {
		glog.Fatal(err)
	}
}
