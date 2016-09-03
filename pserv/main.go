package main

import (
	"errors"
	"log"
	"net"

	"coin/cpb"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	port         = ":50051"
	enoughMiners = 2
)

// server is used to implement cpb.CoinServer.
type server struct{}

// loggedIn keeps track of logins, prevents duplicates
var loggedIn = make(map[string]int)

var nextID = -1 // we want id assigned from 0, see below

// Login implements cpb.CoinServer
func (s *server) Login(ctx context.Context, in *cpb.LoginRequest) (*cpb.LoginReply, error) {
	if _, ok := loggedIn[in.Name]; ok {
		return nil, errors.New("You are already logged in!")
	}
	nextID++
	loggedIn[in.Name] = nextID
	return &cpb.LoginReply{Id: int32(nextID), Work: "work for " + in.Name}, nil
}

var linedUp int
var waiting = make(chan struct{})

// Mine implements cpb.CoinServer, synchronise start of miners
func (s *server) Mine(ctx context.Context, in *cpb.MineRequest) (*cpb.MineReply, error) {
	linedUp++
	if linedUp == enoughMiners {
		close(waiting)
	}
	for {
		if pistolFired() {
			break
		}
	}
	return &cpb.MineReply{Ok: true}, nil
}

// pistolFired  returns false until channel waiting is closed.
func pistolFired() bool {
	select {
	case <-waiting:
		return true
	default:
		return false
	}
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	cpb.RegisterCoinServer(s, &server{})
	s.Serve(lis)
}
