package main

import (
	"coin"
	cpb "coin/service"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	allowedTime          = 2  // "number of seconds before miner declared NOT alive"
	allowedConductorTime = 20 // number of seconds for the conductor
)

var (
	index     = flag.Int("index", -1, "RPC port is 50051+index") // must be at least 0
	numMiners = flag.Int("miners", 3, "number of miners")        // DOESNT include the external one
	debug     = flag.Bool("d", false, "debug mode")
)

type lockMap struct {
	sync.Mutex
	countIN  int
	loggedIn map[string]string
}

type blockdata struct {
	u      []byte // upper coinbase
	l      []byte // lower coinbase
	height uint32 // blockheight
	blk    []byte // 80 byte block header partially filled
	merk   []byte // merkle root skeleton - multiple of 32 bytes
	bits   uint32 // for target computation
}

type lockBlock struct {
	sync.Mutex
	data blockdata
}

type lockChan struct {
	sync.Mutex
	winnerFound bool
	ch          chan struct{}
}

var (
	users      lockMap
	block      lockBlock      // models the block information - basis of 'work'
	run        lockChan       // channel that controls start of run
	signIn     chan string    // for registering users in getwork
	signOut    chan string    // for registering leaving users in getcancel
	stop       sync.WaitGroup // control cancellation issue
	blockchan  chan blockdata // for incoming block
	resultchan chan cpb.Win   // for the winner decision
)

var mysql map[string]string

func auth(credentials string) (string, string, bool) {
	arr := strings.Split(credentials, ",")
	login := arr[0]
	time := arr[1]
	user := arr[2]

	key := mysql[user]
	if key == "" {
		return "", "", true
	}

	expected, err := coin.GenLogin(user, key, time)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if expected != login {
		return "", "", true
	}
	// otherwise
	return login, user, false
}

// Login implements cpb.CoinServer
func (s *server) Login(ctx context.Context, in *cpb.LoginRequest) (*cpb.LoginReply, error) { // HL
	users.Lock()
	defer users.Unlock()
	nxtplus := users.countIN + 2
	// fmt.Println("countIN:", users.countIN, " MINERS:", *numMiners)  // OMIT
	if nxtplus == *numMiners {
		return nil, errors.New("Capacity reached!")
	}
	// authenticate user
	login, userid, nogood := auth(in.Name)
	if nogood {
		return nil, errors.New("Authentication failure")
	}
	users.countIN++
	users.loggedIn[login] = userid     // HL
	return &cpb.LoginReply{Id: 0}, nil // FIXME - we do not need to return any value, not used
}

// GetWork implements cpb.CoinServer, synchronises start of miners, hands out work
func (s *server) GetWork(ctx context.Context, in *cpb.GetWorkRequest) (*cpb.GetWorkReply, error) {
	debugF("Work request: %+v\n", in) // OMIT
	signIn <- in.Name                 // HL
	<-run.ch                          // HL
	// customise work for this miner
	work := setWork(in.Name)
	return &cpb.GetWorkReply{Work: work}, nil
}

// this is needed because we no longer issue miner IDs - needed by steWork below
func minerID(name string) int {
	return 1 // FIXME - this should be a unqiue ID assigned to miner for purposes of mining
}

func setWork(name string) *cpb.Work {
	if name == "EXTERNAL" {
		return &cpb.Work{Coinbase: []byte{}, Block: []byte{}, Skel: []byte{}}
	}
	block.Lock()
	minername := fmt.Sprintf("%d:%s", *index, name)
	miner := minerID(name) // we return an ID attahed to this miner by name
	upper := block.data.u
	lower := block.data.l
	blockHeight := block.data.height
	partblock := block.data.blk
	merkSkel := block.data.merk
	bits := block.data.bits
	// generate actual coinbase txn
	coinbaseBytes, err := coin.GenCoinbase(upper, lower, blockHeight, miner, minername)
	fatalF("failed to set block data", err)
	block.Unlock()
	// fmt.Printf("miner: %s\ncoinbase:\n%x\n", minername, coinbaseBytes)
	return &cpb.Work{Coinbase: coinbaseBytes, Block: partblock, Skel: merkSkel, Bits: bits}
}

// Announce responds to a proposed solution : implements cpb.CoinServer
func (s *server) Announce(ctx context.Context, soln *cpb.AnnounceRequest) (*cpb.AnnounceReply, error) {
	run.Lock()
	defer run.Unlock()
	if run.winnerFound { // reject all but the first
		return &cpb.AnnounceReply{Ok: false}, nil
	}
	// we have a  winner
	run.winnerFound = true                                   // HL
	resultchan <- *soln.Win                                  // HL
	fmt.Println("starting signout numminers = ", *numMiners) // OMIT
	WaitFor(signOut, "out")
	run.ch = make(chan struct{}) // HL
	stop.Done()                  // HL
	return &cpb.AnnounceReply{Ok: true}, nil
}

// GetCancel broadcasts a cancel instruction : implements cpb.CoinServer
func (s *server) GetCancel(ctx context.Context, in *cpb.GetCancelRequest) (*cpb.GetCancelReply, error) {
	signOut <- in.Name // HL
	stop.Wait()        // HL
	serv := *index
	return &cpb.GetCancelReply{Index: uint32(serv)}, nil // HL
}

// server is used to implement cpb.CoinServer.
type server struct{}

// IssueBlock receives the new block from Conductor : implements cpb.CoinServer
func (s *server) IssueBlock(ctx context.Context, in *cpb.IssueBlockRequest) (*cpb.IssueBlockReply, error) {
	blockchan <- blockdata{in.Lower, in.Upper, in.Blockheight, in.Block, in.Merkle, in.Bits}
	users.loggedIn["EXTERNAL"] = "" //1 // we login conductor here
	return &cpb.IssueBlockReply{Ok: true}, nil
}

// GetResult sends back win to Conductor : implements cpb.CoinServer
func (s *server) GetResult(ctx context.Context, in *cpb.GetResultRequest) (*cpb.GetResultReply, error) {
	result := <-resultchan // wait for a result
	//fmt.Printf("sendresult: %d, %v\n", *index, result) // OMIT
	fmt.Printf("sendresult: %d\n", *index) // OMIT
	return &cpb.GetResultReply{Winner: &result, Index: uint32(*index)}, nil
}

// utilities
func fatalF(message string, err error) {
	if err != nil {
		log.Fatalf(message+": %v", err)
	}
}

func debugF(format string, args ...interface{}) {
	if *debug {
		log.Printf(format, args...)
	}
}

// WaitFor allows for the loss of a miners
func WaitFor(sign chan string, direction string) {
	alive := make(map[string]bool) // HL
	count := 1
	alive[<-sign] = true // we need at least one! ... then the rest ...
	stopWaiting := false
	for i := 1; i < *numMiners; i++ {
		select {
		case <-time.After(allowedTime * time.Second):
			stopWaiting = true //  exit, time is up
		case c := <-sign:
			alive[c] = true
			count++
		}
		if stopWaiting { // the remaining miners are taking too long, abandon them
			break
		}
	}
	// done OMIT
	if direction == "in" && count < *numMiners {
		for name := range users.loggedIn {
			if !alive[name] && name != "EXTERNAL" {
				fmt.Printf("DEAD: %s\n", name)
				delete(users.loggedIn, name)
				users.countIN--
			}
		}
	}
	fmt.Printf("miners %s = %d\n", direction, count)
}

// coinbase accepts data from work, result is tailored to miner
// func coinbase(upper []byte, lower []byte, blockHeight int,
// 	miner int, minername string) coin.Transaction {
// 	txn, err := coin.GenCoinbase(upper, lower, blockHeight, miner, minername)
// 	fatalF("failed to generate coinbase transaction", err)
// 	// fmt.Printf("%x", txn)
// 	return coin.Transaction(txn) // convert to a transaction type
// }

func main() {
	flag.Parse() // HL

	users.loggedIn = make(map[string]string)
	users.countIN = -1
	*numMiners++      // to include the Conductor (EXTERNAL)
	if *index == -1 { // mandatory
		log.Fatalf("%s", "Server port missing! use -index i, i=0,1, ...")
	}
	port := fmt.Sprintf(":%d", 50051+*index) // HL
	lis, err := net.Listen("tcp", port)
	fatalF("failed to listen", err)

	signIn = make(chan string, *numMiners)  // register incoming miners
	signOut = make(chan string, *numMiners) // register miners receipt of cancel instructions
	blockchan = make(chan blockdata, 1)     // transfer block data
	run.ch = make(chan struct{})            // signal to start mining
	resultchan = make(chan cpb.Win)         // transfer solution data

	mysql = make(map[string]string)
	mysql["one"] = "thekey"
	mysql["two"] = "anotherthekey"

	go func() {
		for {
			haveBlock := false
			for {
				select {
				case block.data = <-blockchan: // HL
					haveBlock = true //break out of this loop
				case <-time.After(allowedConductorTime * time.Second): // HL
					fmt.Println("Need a live conductor!")
				}
				if haveBlock {
					break
				}
			}
			WaitFor(signIn, "in") // HL
			fmt.Printf("\n--------------------\nNew race!\n")
			run.winnerFound = false // HL
			stop.Add(1)             // HL
			close(run.ch)           // HL
		}
	}()
	s := new(server)
	g := grpc.NewServer()
	cpb.RegisterCoinServer(g, s)
	g.Serve(lis)
}
