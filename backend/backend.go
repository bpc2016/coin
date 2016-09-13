package main

import (
	cpb "coin/service"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	port    = ":50051"
	timeOut = 14
)

var (
	numMiners = flag.Int("m", 3, "number of miners")
	debug     = flag.Bool("d", false, "debug mode")
)

// logger type is for the users login details
type logger struct {
	sync.Mutex //anonymous
	nextID     int
	loggedIn   map[string]int
}

var users logger

// Login implements cpb.CoinServer
func (s *server) Login(ctx context.Context, in *cpb.LoginRequest) (*cpb.LoginReply, error) {
	users.Lock()
	defer users.Unlock()
	if _, ok := users.loggedIn[in.Name]; ok {
		return nil, errors.New("You are already logged in!")
	}
	users.nextID++
	users.loggedIn[in.Name] = users.nextID
	return &cpb.LoginReply{Id: uint32(users.nextID)}, nil
}

var start sync.WaitGroup

// GetWork implements cpb.CoinServer, synchronise start of miners
func (s *server) GetWork(ctx context.Context, in *cpb.GetWorkRequest) (*cpb.GetWorkReply, error) {
	if *debug {
		fmt.Printf("Work request: %+v\n", in)
	}
	s.signIn <- in.Name // register
	start.Wait()        // all must wait, start when this is Done()
	return &cpb.GetWorkReply{Work: fetchWork(in.Name)}, nil
}

var block string

// prepares the candidate block and also provides user specific coibase data
func fetchWork(name string) *cpb.Work { // TODO -this should return err as well
	return &cpb.Work{Coinbase: name, Block: []byte(block)}
}

// Announce responds to a proposed solution : implements cpb.CoinServer
func (s *server) Announce(ctx context.Context, soln *cpb.AnnounceRequest) (*cpb.AnnounceReply, error) {
	won := s.vetWin(*soln.Win)
	return &cpb.AnnounceReply{Ok: won}, nil
}

// extAnnounce is the analogue of 'Announce'
func (s *server) extAnnounce(ch chan struct{}) {
	select {
	case <-ch: //s.search.ch:
		return
	case <-time.After(timeOut * time.Second):
		s.vetWin(cpb.Win{Coinbase: "external", Nonce: 0}) // bogus
		return
	}
}

var stop sync.WaitGroup

// GetCancel blocks until a valid stop condition then broadcasts a cancel instruction : implements cpb.CoinServer
func (s *server) GetCancel(ctx context.Context, in *cpb.GetCancelRequest) (*cpb.GetCancelReply, error) {
	s.signOut <- in.Name // register
	stop.Wait()          // wait for valid solution  OR timeout
	return &cpb.GetCancelReply{Ok: true}, nil
}

// vetWins handles wins - all are directed here
func (s *server) vetWin(thewin cpb.Win) bool {
	s.search.Lock()
	defer s.search.Unlock()
	select {
	case <-s.search.ch: // closed if already have a winner
		return false
	default:
		fmt.Printf("Winner: %+v\n", thewin)
		close(s.search.ch) // until call for new run resets this one
		for i := 0; i < *numMiners; i++ {
			<-s.signOut
		}
		stop.Done()  // issue cancellation to our clients
		start.Add(1) // reset start waitgroup
		block = fmt.Sprintf("BLOCK: %v", time.Now())

		fmt.Printf("\n--------------------\nNew race!\n")
		return true
	}
}

type winchannel struct {
	sync.Mutex
	ch chan struct{}
}

// server is used to implement cpb.CoinServer.
type server struct {
	signIn  chan string // for registering users in getwork
	signOut chan string // for registering leaving users in getcancel
	search  winchannel  // search.ch is closed when we have dclared a winner
}

// initalise
func init() {
	users.loggedIn = make(map[string]int)
	users.nextID = -1
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := new(server)

	s.signIn = make(chan string)
	s.signOut = make(chan string)
	block = fmt.Sprintf("BLOCK: %v", time.Now())
	start.Add(1)

	go func() {
		for {
			for i := 0; i < *numMiners; i++ { // loop blocks here until miners are ready
				<-s.signIn
			}
			stop.Add(1)                       // prep channel for getcancels
			s.search.ch = make(chan struct{}) // reset this channel
			start.Done()                      // start our miners
			go s.extAnnounce(s.search.ch)     // start external miners
		}
	}()

	g := grpc.NewServer()
	cpb.RegisterCoinServer(g, s)
	g.Serve(lis)
}
