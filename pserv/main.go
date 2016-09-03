package main

import (
	"errors"
	"log"
	"net"
	"sync"

	"coin/cpb"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	port         = ":50051"
	enoughMiners = 3
)

// server is used to implement cpb.CoinServer.
type server struct{}

// logger type is for the users var  accounts for logins
type logger struct {
	nextID   int
	mu       sync.Mutex
	loggedIn map[string]int
}

var users logger

// initalise
func init() {
	users.loggedIn = make(map[string]int)
	users.nextID = -1
}

// Login implements cpb.CoinServer
func (s *server) Login(ctx context.Context, in *cpb.LoginRequest) (*cpb.LoginReply, error) {
	users.mu.Lock()
	if _, ok := users.loggedIn[in.Name]; ok {
		return nil, errors.New("You are already logged in!")
	}
	users.nextID++
	users.loggedIn[in.Name] = users.nextID
	users.mu.Unlock()
	return &cpb.LoginReply{Id: int32(users.nextID), Work: "work for " + in.Name}, nil
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
