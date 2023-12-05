package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/tmc/grpc-websocket-proxy/examples/cmd/wswtp/wtpserver/proto/external/wgtwo/webterminal/v0"
)

type Server struct{
	v0.UnimplementedWebTerminalServiceServer
}
func (s *Server) Pipe(srv v0.WebTerminalService_PipeServer) error {
	for {
		req, err := srv.Recv()
		log.Println("Recv() req, err:", req, err)
		if err != nil {
			return err
		}
		m := v0.WebTerminalMessage{
			Message: &v0.WebTerminalMessage_Answer{Answer: &v0.Answer{Sdp:"hoho"}},
		}
		if err := srv.Send(&m); err != nil {
			return err
		}
	}
}
