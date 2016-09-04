package main

import (
	"errors"
	"fmt"
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
	mu       sync.Mutex //TODO make what follows a struct too?
	nextID   int
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

type starts struct {
	linedUp int
	waiting chan struct{}
}

type starter struct {
	mu sync.Mutex
	in starts
}

var getSet starter

// GetWork implements cpb.CoinServer, synchronise start of miners
func (s *server) GetWork(ctx context.Context, in *cpb.GetWorkRequest) (*cpb.GetWorkReply, error) {
	fmt.Printf("GetWork req: %+v\n", in)
	getSet.mu.Lock()
	if getSet.in.linedUp == 0 {
		fmt.Printf("GetWork lined=0, resetting getEnd.waiting ...\n")
		getEnd.mu.Lock()
		getEnd.waiting = make(chan struct{})
		getEnd.mu.Unlock()
	}
	getSet.in.linedUp++
	fmt.Printf("GetWork linedUp = %d\n", getSet.in.linedUp)
	if getSet.in.linedUp == enoughGetWorkrs {
		close(getSet.in.waiting)
		getSet.in.linedUp = 0
	}
	getSet.mu.Unlock()

	work := fetchWork(in) // get this in advance
	<-getSet.in.waiting   // wait

	if getSet.in.linedUp == 0 { // we have just closed
		getSet.mu.Lock()
		getSet.in.waiting = make(chan struct{})
		getSet.mu.Unlock()
	}
	return &cpb.GetWorkReply{Work: work}, nil
}

// Announce implements cpb.CoinServer
func (s *server) Announce(ctx context.Context, soln *cpb.AnnounceRequest) (*cpb.AnnounceReply, error) {
	checked := verify(soln)
	return &cpb.AnnounceReply{Ok: checked}, nil
}

// GetCancel implements cpb.CoinServer
func (s *server) GetCancel(ctx context.Context, in *cpb.GetCancelRequest) (*cpb.GetCancelReply, error) {
	fmt.Printf("GetCancel request from %s\n", in.Name)

	<-getEnd.waiting // wait

	fmt.Printf("GetCancel OUT: %s\n", in.Name)
	return &cpb.GetCancelReply{Ok: true}, nil
}

// check whether the proposed nonce/coinbase works with current block
// TODO -this should return err as well
func verify(soln *cpb.AnnounceRequest) bool {
	// select {
	// case <-getEnd.waiting:
	// 	fmt.Printf("LATE proposed solution: %+v\n", soln)
	// 	return true
	// default:
	fmt.Printf("received proposed solution: %+v\n", soln)
	getEnd.mu.Lock()
	select {
	case <-getEnd.waiting: // already closed
	default:
		close(getEnd.waiting)
	}
	getEnd.mu.Unlock()
	return true
	// }
}

type stoper struct {
	mu      sync.Mutex
	waiting chan struct{}
}

var getEnd stoper

//TODO - move these out of here

// prepares the candidate block and also provides user specific coibase data
// TODO -this should return err as well
func fetchWork(in *cpb.GetWorkRequest) *cpb.Work {
	return &cpb.Work{Coinbase: in.Name, Block: []byte("hello world")}
}

// initalise
func init() {
	users.loggedIn = make(map[string]int)
	users.nextID = -1

	getSet.in.waiting = make(chan struct{})
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
