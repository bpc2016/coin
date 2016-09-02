package main

import (
	"log"
	"net"

	pb "coin/proto"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

// server is used to implement cpb.CoinServer.
type server struct{}

// Login implements cpb.CoinServer
func (s *server) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginReply, error) {
	return &pb.LoginReply{Id: 1234, Work: "Belongs to " + in.Name}, nil
}

// could have used in.Name

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCoinServer(s, &server{})
	s.Serve(lis)
}
