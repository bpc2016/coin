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

var (
	index     = flag.Int("index", 0, "RPC port is 50052+index") //; debug port is 36661+index")
	numMiners = flag.Int("m", 3, "number of miners + plus controller")
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

type blockdata struct {
	sync.Mutex
	data string
}

var block blockdata
var start sync.WaitGroup

// GetWork implements cpb.CoinServer, synchronise start of miners
func (s *server) GetWork(ctx context.Context, in *cpb.GetWorkRequest) (*cpb.GetWorkReply, error) {
	s.signIn <- in.Name // register
	start.Wait()        // all must wait, start when this is Done()
	block.Lock()
	work := &cpb.Work{Coinbase: in.Name, Block: []byte(block.data)}
	block.Unlock()
	return &cpb.GetWorkReply{Work: work}, nil
}

// Announce responds to a proposed solution : implements cpb.CoinServer
func (s *server) Announce(ctx context.Context, soln *cpb.AnnounceRequest) (*cpb.AnnounceReply, error) {
	won := s.vetWin(*soln.Win)
	if won { // declare the winner
		fmt.Printf("Winner: %+v\n", *soln.Win)
		for i := 0; i < *numMiners; i++ {
			<-s.signOut
		}
		stop.Done()  // issue cancellation to our clients
		start.Add(1) // reset start waitgroup
	}
	return &cpb.AnnounceReply{Ok: won}, nil
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
	won := false
	select {
	case <-s.search.ch: // closed if already have a winner
		//return false
	default:
		// should vet the solution, return false otherwise!!
		//
		close(s.search.ch) // until call for new run resets this one
		won = true         //return true
	}
	return won
}

//=============================================

// this comes from this server's role as a client to frontend
func askForNew() {
	time.Sleep(8 * time.Second)
	block.Lock()
	block.data = fmt.Sprintf("BLOCK: %v", time.Now())
	block.Unlock()
}

// // extAnnounce is the analogue of 'Announce'
// func (s *server) extAnnounce(ch chan struct{}) {
// 	select {
// 	case <-ch: //s.search.ch:
// 		return
// 	case <-time.After(timeOut * time.Second):
// 		s.vetWin(cpb.Win{Coinbase: "external", Nonce: 0}) // bogus
// 		return
// 	}
// }

//----------------------------------------------------

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

	port := fmt.Sprintf(":%d", 50052+*index)
	lis, err := net.Listen("tcp", port) // RPC port // HL
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := new(server)

	s.signIn = make(chan string)
	s.signOut = make(chan string)
	start.Add(1)

	fmt.Printf("Server up on port: %+v\n", port)

	go func() {
		for {
			askForNew()
			for i := 0; i < *numMiners; i++ { // loop blocks here until miners are ready
				si := <-s.signIn
				if *debug {
					fmt.Printf("work request: %+v\n", si)
				}
			}
			fmt.Printf("\n--------------------\nNew race!\n")
			stop.Add(1)                       // prep channel for getcancels
			s.search.ch = make(chan struct{}) // reset this channel
			start.Done()                      // start our miners
		}
	}()

	g := grpc.NewServer()
	cpb.RegisterCoinServer(g, s)
	g.Serve(lis)
}
