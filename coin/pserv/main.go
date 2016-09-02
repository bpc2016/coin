package main

import (
	"errors"
	"log"
	"net"

	pb "coin/cpb"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

// server is used to implement cpb.CoinServer.
type server struct{}

// loggedIn keeps track of logins, prevents duplicates
var loggedIn = make(map[string]int)

var nextID = -1 // we want id assigned from 0, see below

// Login implements cpb.CoinServer
func (s *server) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginReply, error) {
	if _, ok := loggedIn[in.Name]; ok {
		return nil, errors.New("You are already logged in!")
	}
	nextID++
	loggedIn[in.Name] = nextID
	return &pb.LoginReply{Id: int32(nextID), Work: "work for " + in.Name}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCoinServer(s, &server{})
	s.Serve(lis)
}
