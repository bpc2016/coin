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
	port            = ":50051"
	enoughGetWorkrs = 3
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

// Login implements cpb.CoinServer
func (s *server) Login(ctx context.Context, in *cpb.LoginRequest) (*cpb.LoginReply, error) {
	users.mu.Lock()
	if _, ok := users.loggedIn[in.Name]; ok {
		return nil, errors.New("You are already logged in!")
	}
	users.nextID++
	users.loggedIn[in.Name] = users.nextID
	users.mu.Unlock()
	return &cpb.LoginReply{Id: uint32(users.nextID)}, nil
}

type starter struct {
	linedUp int
	waiting chan struct{}
	mu      sync.Mutex
}

var getSet starter

// GetWork implements cpb.CoinServer, synchronise start of miners
func (s *server) GetWork(ctx context.Context, in *cpb.GetWorkRequest) (*cpb.GetWorkReply, error) {
	getSet.mu.Lock()
	getSet.linedUp++
	if getSet.linedUp == enoughGetWorkrs {
		close(getSet.waiting)
		getSet.linedUp = 0
	}
	getSet.mu.Unlock()

	work := fetchWork(in) // get this in advance
	for {                 // wait
		if pistolFired() {
			break
		}
	}
	if getSet.linedUp == 0 { // we have just closed
		getSet.waiting = make(chan struct{})
	}
	return &cpb.GetWorkReply{Work: work}, nil
}

// pistolFired  returns false until channel waiting is closed.
func pistolFired() bool {
	select {
	case <-getSet.waiting:
		return true
	default:
		return false
	}
}

//TODO - move these out of here

// prepares teh candidatee block and also provides user specific data
func fetchWork(in *cpb.GetWorkRequest) *cpb.Work {
	return &cpb.Work{Specific: in.Name, Block: []byte("hello world")}
}

// initalise
func init() {
	users.loggedIn = make(map[string]int)
	users.nextID = -1

	getSet.waiting = make(chan struct{})
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
