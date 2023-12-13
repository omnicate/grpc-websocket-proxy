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
	//	grpcMetadata "google.golang.org/grpc/metadata"
)

var (
	grpcAddr       = flag.String("grpcaddr", "", "listen grpc addr") // ""
	httpAddr       = flag.String("addr", ":8000", "listen http addr")
	debugAddr      = flag.String("debugaddr", ":8002", "listen debug addr")
	wg2AccessToken = flag.String("wg2_token", "token", "WG2 api access tocken")
	wg2Msisdn      = flag.String("wg2_msisdn", "", "WG2 Sub msisdn")
	wg2ApiAddress  = flag.String("wg2_api_address", "", "WG2 Api address")
)
var defaultHeadersToForward = map[string]bool{
	"Origin":     true,
	"origin":     true,
	"Referer":    true,
	"referer":    true,
	"wg2-msisdn": true,
}

type bearerTokenCreds struct {
	token  string
	msisdn string
}

func (c *bearerTokenCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"Authorization": "Bearer " + c.token,
		"wg2-msisdn":    c.msisdn,
	}, nil
}

func (c *bearerTokenCreds) RequireTransportSecurity() bool {
	return false
}
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
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(conf),
		grpc.WithPerRPCCredentials(&bearerTokenCreds{token: *wg2AccessToken, msisdn: *wg2Msisdn}),
	}
	// Add token to gRPC Request.

	err := v0.RegisterWebTerminalServiceHandlerFromEndpoint(ctx, mux, *wg2ApiAddress, opts)
	opts = []grpc.DialOption{grpc.WithInsecure(),
		grpc.WithPerRPCCredentials(&bearerTokenCreds{token: *wg2AccessToken, msisdn: *wg2Msisdn}),
	}
	//err := v0.RegisterWebTerminalServiceHandlerFromEndpoint(ctx, mux, *grpcAddr, opts)
	if err != nil {
		return err
	}
	go http.ListenAndServe(*debugAddr, nil)
	fmt.Println("listening")

	l := log.New()
	l.SetLevel(log.DebugLevel)
	l.Debugln("hei from debug logger", *wg2Msisdn)
	http.ListenAndServe(*httpAddr, wsproxy.WebsocketProxy(mux,
		wsproxy.WithLogger(l),
		wsproxy.WithRequestMutator(
			func(incoming *http.Request, outgoing *http.Request) *http.Request {
				return outgoing
			})))
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
