package main

import (
	"coin"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	cpb "coin/service"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	debug         = flag.Bool("d", false, "debug mode")
	servers       = flag.String("s", "", "Servers - list url_1:i_1,url_2:i_2, i_j=0,.. port")
	timeOut       = flag.Int("o", 14, "timeout for EXTERNAL")
	numServers    int // count of expected servers
	dialedServers []cpb.CoinClient
)

type server struct {
	host string
	port int
}

// Bitcoin stuff =========================================

// newBlock packages the block information that becomes 'work' for each run
func newBlock() (upper, lower, bheader, merkle []byte, blockheight, bits uint32) { // TODO - this data NOT fixed
	blockHeight := uint32(433789) // should come from unix time
	blockFees := 8756123          // satoshis
	bits = 0x19015f53             // difficulty
	pubkey := "0225c141d69b74adac8ab984a8eb9fee42c4ce79cf6cb2be166b1ddc0356b37086"
	// conductor generates this ...
	upper, lower, err := coin.CoinbaseTemplates(blockHeight, blockFees, pubkey)
	if err != nil {
		log.Fatalf("failed to generate coinbase: %v", err)
	}
	// call for a blockheader template
	bheader = blockHeader(int(bits))
	// fetch the  skeleton mr
	merkle = merkleRoot()
	// sends upper, lower , blockHeight --> server
	return upper, lower, bheader, merkle, blockHeight, bits
}

// blockHeader supplies the 80 byte bh template
func blockHeader(bits int) []byte {
	Version := 2
	PrevBlock := "000000000000000117c80378b8da0e33559b5997f2ad55e2f7d18ec1975b9717"
	TimeStamp := 0x53058b35 // tt := fmt.Sprintf("%x",uint32(time.Now().Unix()))
	bh, err := coin.BlockHeader(Version, PrevBlock, TimeStamp, bits)
	if err != nil {
		log.Fatalf("failed to generate blockheader: %v", err)
	}
	return bh
}

// merkelRoot hands out the skeleton in lieu of actual txn hashes, a 32*n byte sequence
func merkleRoot() []byte {
	// 		"00baf6626abc2df808da36a518c69f09b0d2ed0a79421ccfde4f559d2e42128b",
	var txHashes = []string{
		"91c5e9f288437262f218c60f986e8bc10fb35ab3b9f6de477ff0eb554da89dea",
		"46685c94b82b84fa05b6a0f36de6ff46475520113d5cb8c6fb060e043a0dbc5c",
		"ba7ed2544c78ad793ef5bb0ebe0b1c62e8eb9404691165ffcb08662d1733d7a8",
		"b8dc1b7b7ed847c3595e7b02dbd7372aa221756b718c5f2943c75654faf48589",
		"25074ef168a061fcc8663b4554a31b617683abc33b72d2e2834f9329c93f8214",
		"0fb8e311bffffadc6dc4928d7da9e142951d3ba726c8bde2cf1489b62fb9ebc5",
		"c67c79204e681c8bb453195db8ca7d61d4692f0098514ca198ccfd1b59dbcee3",
		"bd27570a6cbd8ad026bfdb8909fdae9321788f0643dea195f39cd84a60a1901b",
		"41a06e53ffc5108358ddcec05b029763d714ae9f33c5403735e8dee78027fe74",
		"cc2696b44cb07612c316f24c07092956f7d8b6e0d48f758572e0d611d1da6fb9",
		"8fc508772c60ace7bfeb3f5f3a507659285ea6f351ac0474a0a9710c7673d4fd",
		"62fed508c095446d971580099f976428fc069f32e966a40a991953b798b28684",
		"928eadbc39196b95147416eedf6f635dcff818916da65419904df8fde977d5db",
		"b137e685df7c1dffe031fb966a0923bb5d0e56f381e730bc01c6d5244cfe47c1",
		"b92207cee1f9e0bfbd797b05a738fab9de9c799b74f54f6b922f20bd5ec23dd6",
		"29d6f37ada0481375b6903c6480a81f8deaf2dcdba03411ed9e8d3e5684d02dd",
		"48158deb116e4fd0429fbbbae61e8e68cb6d0e0c4465ff9a6a990037f88c489c",
		"be64ea86960864cc0a0236bbb11f232faf5b19ae6e2c85518628f5fae37ec1ca",
		"081363552e9fff7461f1fc6663e1abd0fb2dd1c54931e177479a18c4c26260e8",
		"eb87c25dd2b2537b1ff3dbabc420e422e2a801f1bededa6fa49ef7980feaef70",
		"339e16fcc11deb61ccb548239270af43f5ad34c321416bada4b8d66467b1c697",
		"4ad6417a3a04179482ed2e4b7251c396e38841c6fba8d2ce9543337ab7c93c02",
		"c28a45cded020bf424b400ffc9cb6f2f85601934f18c34a4f78283247192056a",
		"882037cc9e3ee6ddc2d3eba86b7ca163533b5d3cbb16eaa38696bb0a2ea1137e",
		"179bb936305b46bb0a9df330f8701984c725a60e063ad5892fa97461570b5c04",
		"9517c585d1578cb327b7988f38e1a15c663955ea288a2292b40d27f232fbb980",
		"2c7e07d0cf42e5520bcbfe2f5ef63761a9ab9d7ccb00ea346195eae030f3b86f",
		"534f631fc42ae2d309670e01c7a0890e4bfb65bae798522ca14df09c81b09734",
		"104643385619adb848593eb668a8066d1f32650edf35e74b0fc3306cb6719448",
		"87ac990808239c768182a752f4f71cd98558397072883c7e137efb49d22b9231",
		"9b3e2f1c47d59a444e9b6dc725f0ac6baf160d22f3a9d399434e5e65b14eccb0",
		"fbe123066ae5add633a542f151663db4eb5a7053e388faadb40240671ae1b09b",
		"1dd07e92e20b3cb9208af040031f7cfc4efd46cc31ec27be20a1047965a42849",
		"2709bb9ed27353c1fd76b9240cab7576a44de68945e256ad44b2cb8d849a8060",
		"d0174db2c712573432a7869c1508f371f3a1058aeedddc1b53a7e04d7c56c725",
		"b4a16f724cddb8f77ddf3d2146a12c4be13d503885eaba3518a03da005009f62",
		"2aa706d75decbe57745e01d46f9f5d30a08dedaf3288cee14cc4948e3684e1d4",
		"ee49c5f6a5129ccaf2abebbc1d6d07a402a600af6221476b89aafaa683ca95b7",
		"bea1011c77874845e9b4c876ed2ceebd530d428dd4a564ad003d9211d40bb091",
		"f1e88ffc2b1de2aa4827002f06943ce5468735f7433f960bf01e75885b9f832b",
		"19247d017e002fb9143d1a89eb921222a94f8a3d0faaf2e05b0f594989edc4c4",
		"13f714ff62ee7d26b6d69ca980c141ebc54e9f71d2697083fe6c5efc1b02bd0f",
		"0c78cbb8246572f015fbdc53dc9798fa54d1119ec77c1f07ac310bcbcc40dbf8",
		"4bcde0ef92a6d24a2be7be50ac5e5299d776df2e6229ba5d475c2491da94f255",
		"0cfd7d1058502730cf0b2ffa880c78ef534651e06832b5d87c0d7eb84eac5b0c",
		"3a168f794d6e0c614429ad874317cc4cd67a8177214880ff6ea1704d29228c2f",
		"f9a555d817334397b402518d6fd959dc73d981ee7f5fe67969b63974ebbef127",
		"24b52691f66eaed4ce391a473902e309018257c98b9f02aaa33b399c9e6f3168",
		"a37b5e623dc26a180d9e2c9510d06885b014e86e533adb63ec40511e10b55046",
		"9dbaeb485e51d9e25a5621dc46e0bc0aaf51fb26be5acc4e370b96f62c469b80",
		"a6431d3d39f6c38c5df48405090752cab03bfdf5c77cf881b18a946807fba74a",
		"faa77e309f125373acf19855dd496fffe2f74962e545420844557a3adc7ebc11",
		"3523f52543ecfea2f78486dc91550fad0e6467d46d9d9c82ca63b2e0230bfa71",
		"a0583e358e42d77d18d1fd0533ff0a65615fc3b3112061ef92f168a00bf640c1",
		"42ae900888d5e5dde59c8e3d06e13db9e84ef05d27726d4b67fd00c50cd9406a",
		"154940777d3ff78f592ef02790131a59263c36b4958bbc836f9a767ea1a9f178",
		"6a0337de6ac75eecf748306e8ebc5bfe5c811a1481ae50f6956a9e7f26a679f5",
		"c99530c2148e09688d0b88795625943371183bf1f5d56c7446c6ed51ea133589",
		"626421dbe8ad6a0fd0d622d5dd3308a1cdc00b98575a41a91fe01a439e6f40bd",
		"b2f3a559f605a158cc395126c3cf394a7e92a53b7514c75157e1dc43a6c7f93e",
		"dffe06d1bea81f2a01c76786404bb867258f9e68013bf25454097ce935090738",
		"0860159ec7a2a51ce107c182a988c40b4bc2057a734354a1219b6c65e72640ed",
		"a405ff1bb51846b1867acc0b0da17f6f9616e592a0a7ff5ef3297c1ecfd60911",
		"a7d451924263284765f6343bca8a21b79b89ebfe611c7355dd88e0ec1c29e232",
		"41c758d08a4d3fe4d90645711589b832a2cd54dd25bd5b66e463e5d389a53aff",
		"a05c1a93a521fa5dbc1790cfbb808893453a428a65f2c6b2d51249fbb12db309",
		"90997920aa9786e10f513cfdd14e294feee6739cee1ab61b3fb1e3f42e7a915d",
		"99fcb9cb62c20a3135484a70bd3f73983f8f3b7b26266dad34f3993958a7642c",
		"e05f9a668b37e5f78bd3b9d047f29f92b33a87f11dd48390410006f858188b7b",
		"56dbc65895f7992da4a6985e7edba4d1c00879f1b28442c644c8a07658ceab27",
		"5e9004fe262b829563d0804656ba68b1de1690401f08a1915273230d8c902fc0",
		"1ea9ed3717523c5e304b7a7ac8058a87fb4f3fed8c6004769f226c9bb67e79c5",
		"f0f1a4c009b3f1b2729e89898e2f5c0fcdc312edea5df884a9c897cb90e4c566",
		"b5bb4ddf04863e6a60f33cb96c20dac8175d3bae55f335781503143c97a50e43",
		"f14cc97a20c6f627b4b78301352ae35463bc359362589cd178a06c0fa90850b7",
		"628801c8f614015c0fa0ccb2768cccc3e7b9d41ceed06071ce2534d31f7236d6",
		"3be1013c8f8da150e2195408093153b55b08b037fd92db8bb5e803f4c2538aae",
		"c9e1f8777685f54ba65c4e02915fd649ee1edcbf9c77ddf584b943d27efb86c3",
		"4274e92ed3bd02eb101baa5fb8ff7b96236830762d08273749fbb5166db8ab0b",
		"aa84c955bea04c7cee8f5bbbec97d25930fcaca363eed1b8cad37b931556d3e3",
		"d6a29c948677fb1f71aaf16debc3d071a4dd349458eb9e056dce3a000ff853da",
		"ba84bdb3d78367ca365016ac4bff9269576eb010f874c2967af73e0de5638de0",
		"1546c79951e3b541bc64d1957b565b7a2850fc87192c7b374aee6cfc69b9805e",
		"f119227d492ebe27fe9aae321980802454dfa64b2691efbe796c5075d5b07f62",
		"b8cf13d64818b32f96bbb585998b1bc9505f6a94055488e5a71fee9479c6f2a9",
		"1aaf459705b6afef2d7b83e3f181f1af55be0813daf55edce104cc59abc28ed7",
		"61ac185c8f520b5e3134953dc52ff292a40e1e96b088dab259558a9d240ec02f",
		"2da96e3154d7ec2329f787b73cb8a436b92d64cf3cc28e920d073279ea73b5f8",
		"1c4d72ce733b971b9ec4e24f37d733355f6f2ea635cc67ffb3e22748484df446",
		"2a6f89769f3272ac8c7a36a42a57627eca6b260ab2c76d8046a27d44d4034893",
		"f8d11df51a2cc113698ebf39a958fe81179d7d973d2044322771c0fe63f4d7c9",
		"f2287f17a4fa232dca5715c24a92f7112402a8101b9a7b276fb8c8f617376b90",
		"bb5ee510a4fda29cae30c97e7eee80569d3ec3598465f2d7e0674c395e0256e9",
		"647ab8c84365620d60f2523505d14bd230b5e650c96dee48be47770063ee7461",
		"34b06018fcc33ba6ebb01198d785b0629fbdc5d1948f688059158f053093f08b",
		"ff58b258dab0d7f36a2908e6c75229ce308d34806289c912a1a5f39a5aa71f9f",
		"232fc124803668a9f23b1c3bcb1134274303f5c0e1b0e27c9b6c7db59f0e2a4d",
		"27a0797cc5b042ba4c11e72a9555d13a67f00161550b32ede0511718b22dbc2c",
	}
	// generate the skeleton - from all but the coinbase
	skel, err := coin.Skeleton(txHashes)
	if err != nil {
		log.Fatalf("failed to generate blockheader: %v", err)
	}
	return skel
}

// Networking ==============================================

// getCancel makes a blocking request to the server
func getCancel(c cpb.CoinClient, name string, stopLooking chan string, gotCancel chan struct{}) {
	_, err := c.GetCancel(context.Background(), &cpb.GetCancelRequest{Name: name})
	if skipF(c, "could not request cancellation", err) {
		return
	}
	stopLooking <- fmt.Sprintf("%v", c) // stop search indicate who from
	gotCancel <- struct{}{}             // quit loop
	fmt.Println("GETCANCEL ISSUED")
}

// getResult makes a blocking request to the server, comes back with theWinner
func getResult(c cpb.CoinClient, name string, theWinner chan string, ignore chan struct{}) {
	res, err := c.GetResult(context.Background(), &cpb.GetResultRequest{Name: name})
	if skipF(c, "could not request result", err) {
		return
	}
	select {
	case <-ignore:
		return // we are too late
	default: // carry on
	}
	fmt.Println("GETRESULT")
	// declare the winner
	if res.Winner.Identity != "EXTERNAL" { // avoid echoes
		declareWin(theWinner, int(res.Index), res.Winner.Identity, res.Winner.Nonce) // HL
	}
}

// annouceWin is what causes the server to issue a  cancellation
func annouceWin(c cpb.CoinClient, nonce uint32, block []byte, winner string) bool {
	win := &cpb.Win{Block: block, Nonce: nonce, Identity: winner}
	// win := &cpb.Win{Coinbase: coinbase, Nonce: nonce}
	r, err := c.Announce(context.Background(), &cpb.AnnounceRequest{Win: win})
	if skipF(c, "could not announce win", err) {
		return false
	}
	return r.Ok
}

func declareWin(theWinner chan string,
	index int, coinbase string, nonce uint32) {
	// check if already declared ..
	select {
	case thew := <-theWinner:
		fmt.Println("ALREADY HAVE WINNER!")
		theWinner <- thew // replace it
		return
	default: // go on
	}
	str := fmt.Sprintf("%s - ", time.Now().Format("15:04:05"))
	if index == -1 { //indicates winner is external
		str += "external" // HL
	} else {
		str += fmt.Sprintf("miner %d:%s, nonce %d", index, coinbase, nonce)
	}
	theWinner <- str // HL
	fmt.Println("DECLARE")

	for i, c := range dialedServers {
		if isDead(c) {
			continue
		}
		if i == index {
			// fmt.Println("DON'T TELL YOURSELF: ", i, c)
			continue
		}
		fmt.Printf("ANNOUNCE c=%v index=%d\n", c, index)
		annouceWin(c, 99, []byte{}, "EXTERNAL") // bogus  announcement
	}
	// }
}

// active[c] is now a struct{} *buffered* channel that is empty
// server c is no longer reachable. Tested by isDead(c)
var active, awake map[cpb.CoinClient]chan (struct{})

func isDead(c cpb.CoinClient) bool {
	select {
	case <-active[c]:
		active[c] <- struct{}{} // refill, so isDead again is false
		return false
	default:
		return true // so active is empty
	}
}

// higher status - now not asleep means has accepted a block
func isAsleep(c cpb.CoinClient) bool {
	select {
	case <-awake[c]:
		awake[c] <- struct{}{} // refill, so isAsleep again is false
		return false
	default:
		return true // so active is empty
	}
}

func mopup(stopLooking chan string) {
	// fmt.Println("DRAIN STOPL")
	// drain stopLooking
	done := false
	for !done {
		select {
		case <-stopLooking:
			// fmt.Println("STOPLOOKING")
		default:
			done = true
		}
	}
}

type result struct {
	nonce   uint32 // winning nonce
	ours    bool   // true if the winner is ours
	cstring string // cpb.CoinClient -> string || ""
}

func drain(ch chan struct{}) {
	done := false
	for !done {
		select {
		case <-ch: // again
		default:
			done = true
		}
	}
}

// based on proto

// getCancel makes a blocking request to the server
func getCancelNew(c cpb.CoinClient, name string, gotCancel chan struct{}) {
	_, err := c.GetCancel(context.Background(), &cpb.GetCancelRequest{Name: name})
	if skipF(c, "could not request cancellation", err) {
		return
	}
	gotCancel <- struct{}{} // acknowledge
	// fmt.Println("GETCANCELNEW ISSUED")
}

// issue blocks
func issueBlocks(cancelChan chan struct{}) {
	blockSendDone := make(chan struct{}, numServers)  //
	serverWrkAcks := make(chan *cpb.Work, numServers) // for gathering signin
	// reset being awake
	for _, c := range dialedServers {
		drain(awake[c])
	}
	// the block ....
	u, l, blk, m, h, bts := newBlock() // next block

	for _, c := range dialedServers { // RANGE DIALED
		go func(c cpb.CoinClient) {
			_, err := c.IssueBlock(context.Background(), // HL
				&cpb.IssueBlockRequest{ // HL
					Upper:       u,   // OMIT
					Lower:       l,   // OMIT
					Block:       blk, // OMIT
					Merkle:      m,   // OMIT
					Blockheight: h,   // OMIT
					Bits:        bts})
			if skipF(c, "could not issue block", err) {
				blockSendDone <- struct{}{}
				return
			}
			blockSendDone <- struct{}{}
			// get ready, get set ... this blocks!  OMIT
			r, err := c.GetWork(context.Background(), // HL
				&cpb.GetWorkRequest{Name: "EXTERNAL"}) // HL
			if skipF(c, "could not reconnect", err) { // HL
				return
			} else if isDead(c) { // were we previously declared dead? change that ..
				active[c] <- struct{}{} // 'revive' us
			}
			serverWrkAcks <- r.Work
			awake[c] <- struct{}{} // register that this server is awake
			go getCancelNew(c, "EXTERNAL", cancelChan)
		}(c)
	} // END  RANGE DIALED
	// wait a bit drain blockSendDone
	for i := 0; i < numServers; i++ {
		// fmt.Println("blocksenddone")
		<-blockSendDone
	}
	//  collect the work request acks from servers
	for _, c := range dialedServers {
		if isDead(c) {
			debugF("server DOWN: %v\n", c)
			continue
		}
		// fmt.Println("serverWrkAcks")
		<-serverWrkAcks
		debugF("server up: %v\n", c) // ?? how do we know
	}
}

// getResult makes a blocking request to the server, comes back with theWinner
func getResultNew(c cpb.CoinClient, name string, Winner chan *cpb.GetResultReply, ignore chan struct{}) {
	res, err := c.GetResult(context.Background(), &cpb.GetResultRequest{Name: name})
	if skipF(c, "could not request result", err) {
		return
	}
	select {
	case <-ignore:
		ignore <- struct{}{} // shut the gate once more - revive
		return               // we are too late
	default: // carry on
	}
	// fmt.Println("GETRESULT CALL")
	ignore <- struct{}{} // shut the gates
	Winner <- res
}

type winStruct struct {
	message string
	source  int
}

func main() {
	flag.Parse()
	myServers := checkMandatoryF()
	numServers = len(myServers)
	active = make(map[cpb.CoinClient]chan (struct{}))
	awake = make(map[cpb.CoinClient]chan (struct{}))
	// dial them
	for index := 0; index < numServers; index++ {
		addr := fmt.Sprintf("%s:%d", myServers[index].host, 50051+myServers[index].port)
		conn, err := grpc.Dial(addr, grpc.WithInsecure()) // HL
		if err != nil {
			log.Fatalf("fail to dial: %v", err)
		}
		defer conn.Close()
		c := cpb.NewCoinClient(conn) // note that we do not login!
		dialedServers = append(dialedServers, c)
		active[c] = make(chan struct{}, 1)
		active[c] <- struct{}{}           //  is alive!
		awake[c] = make(chan struct{}, 1) // do not assume awake
	}

	// initialise
	theEnd := make(chan struct{}) // required because we use go routines ... exit (never)

	startSearch := make(chan struct{})   // for firirng off external
	winner := make(chan winStruct)       // stores result of externals trials
	stopSearching := make(chan struct{}) // stop external
	ignore := make(chan struct{})        // where to declare no more anouncements allowed
	// theWinner := make(chan string)       //

	// external (search)
	// counts to timeOut unless disturbed
	// starts with channel startSearch
	go func() {
		fmt.Println("STARTING ...")
		tick := time.Tick(1 * time.Second)
		for {
			<-startSearch // wait here
			carryOn := true
			for cn := 0; carryOn; cn++ {
				if cn >= *timeOut {
					str := fmt.Sprintf("%s - ", time.Now().Format("15:04:05"))
					str += "External"
					winner <- winStruct{str, -1} // external wins
					carryOn = false
				} else { // check for win
					select {
					case <-stopSearching:
						fmt.Println("STOPPED!")
						// winner <- theWinner
						// str := <-theWinner
						// winner <- winStruct{str, 0} // we win
						// drain stopSearching
						// drain(stopSearching)
						carryOn = false
					default: // continue
					}
				}
				if carryOn {
					// wait for a second here ...
					<-tick
					debugF(" | EXT %d\n", cn)
				}
			}
		}
	}()

	// in parallel, listen for server announcement on getResult
	go func() {
		for {
			// fmt.Println("GETRESULTS ..")
			resCh := make(chan *cpb.GetResultReply)
			for _, c := range dialedServers {
				go getResultNew(c, "EXTERNAL", resCh, ignore)
			}
			<-ignore // wait here for a winner
			// fmt.Println("GOT A WINNER 1 ....")
			res := <-resCh
			// fmt.Println("GOT A WINNER  2....", res.Winner.Identity)

			// now declare the winner
			if res.Winner.Identity != "EXTERNAL" { // avoid echoes
				str := fmt.Sprintf("%s - ", time.Now().Format("15:04:05"))
				str += fmt.Sprintf("miner %d:%s, nonce %d",
					res.Index, res.Winner.Identity, res.Winner.Nonce)
				// fmt.Println("GOT A WINNER 3 ....")
				// theWinner <- str
				winner <- winStruct{str, 0} // we win
				// fmt.Println("GOT A WINNER 4 ....")
				stopSearching <- struct{}{} // data on to external search
				// fmt.Println("GOT A WINNER 5 ....")
			}
		}
	}()

	replies := make(chan struct{}, numServers)
	// resultCopy := make(chan string)

	// servers - receiving winner
	// sends replies back to conductor

	// conductor main cycle
	go func() {
		for {
			win := <-winner
			// fmt.Println("Sending backs acks - cancels? message:", win.message)
			// fmt.Println("Announce Winner ...", win.message)
			for _, c := range dialedServers {
				if isAsleep(c) {
					continue
				}
				if win.source != -1 {
					// there should be additional data carried ...
					// we want to avoid reflecting the win back there
					// is c== then skip
					if win.source == 0 { //FIX ME
						continue
					}
				}
				// fmt.Printf("ANNOUNCEWIN to %v\n", c)
				annouceWin(c, 99, []byte{}, "EXTERNAL") // bogus  announcement
			}
			// resultCopy <- win.message // pass this on
			// wait for cancellations
			drain(replies)
			// announce
			fmt.Println("---------------\nWinner: ", win.message, "\n---------------")
			// wait a bit
			// <-time.After(2 * time.Second)
			// awaken by issuing new blocks
			issueBlocks(replies)
			startSearch <- struct{}{} // restart external
			// empty ignore
			select {
			case <-ignore: // this clears it
			default: // is already empty
			}

		}
	}()

	// conductor
	// go func() {
	// 	for {
	// 		r := <-resultCopy // to sync
	// 		// wait for cancellations
	// 		drain(replies)
	// 		// announce
	// 		fmt.Println("Winner: ", r, "\n---------------")
	// 		// wait a bit
	// 		// <-time.After(2 * time.Second)
	// 		// awaken by issuing new blocks
	// 		issueBlocks(replies)
	// 		startSearch <- struct{}{} // restart external
	// 		// empty ignore
	// 		select {
	// 		case <-ignore: // this clears it
	// 		default: // is already empty
	// 		}
	// 		// resultAnnounce <- struct{}{} // also miners search
	// 	}
	// }()

	<-time.After(2 * time.Second)
	fmt.Println("kick off ...")
	startSearch <- struct{}{} // kick off

	/*
		// external
		// here models external net: timeout after timeOut seconds
		// now stopLooking is filled with the stringified server or ""
		go func() {
			tick := time.Tick(1 * time.Second)
			for {
				for {
					quit := false
					for cn := 0; ; cn++ {
						if quit {
							break
						}
						if cn >= *timeOut {
							fmt.Println("TO SET THEWINNER ..")
							theWinner <- "External"
							fmt.Println(" = ME!")

							// winner <- result{nonce: uint32(cn), ours: false, cstring: ""}
							// declareWin(theWinner, -1, "external", uint32(cn))
							for _, c := range dialedServers {
								if isDead(c) {
									continue
								}
								fmt.Printf("ANNOUNCEWIN\n")
								annouceWin(c, 99, []byte{}, "EXTERNAL") // bogus  announcement
							}
							quit = true //break
						}
						if !quit {
							// check for a stop order
							select {
							case <-stopLooking: // expect that getResult sets theWinner here
								// winner <- result{nonce: 0, ours: true, cstring: cs}
								theWinner <- "Yours"
								// fmt.Println("stoplooking ...")
								quit = true
							default: // continue
							}
						}
						if !quit {
							// wait for a second here ...
							<-tick
							debugF(" | EXT %d\n", cn)
						}
					}
				}
			}
		}()

		// issue blocks, get the 'runs' going
		go func() {
			for {
				// ISSUE BLOCK -- wait until we have a winner
				fmt.Println("WON: ", <-theWinner, "\n---------------------------") // a OMIT
				u, l, blk, m, h, bts := newBlock()                                 // next block

				for _, c := range dialedServers { // RANGE DIALED
					go func(c cpb.CoinClient, stopLooking chan string) {
						// start with block issue
						_, err := c.IssueBlock(context.Background(), // HL
							&cpb.IssueBlockRequest{ // HL
								Upper:       u,   // OMIT
								Lower:       l,   // OMIT
								Block:       blk, // OMIT
								Merkle:      m,   // OMIT
								Blockheight: h,   // OMIT
								Bits:        bts})
						if skipF(c, "could not issue block", err) {
							blockSendDone <- struct{}{}
							return
						}
						blockSendDone <- struct{}{}
						// get ready, get set ... this blocks!  OMIT
						r, err := c.GetWork(context.Background(), // HL
							&cpb.GetWorkRequest{Name: "EXTERNAL"}) // HL
						if skipF(c, "could not reconnect", err) { // HL
							return
						} else if isDead(c) { // were we previously decalred dead? change that ..
							active[c] <- struct{}{} // 'revive' us
						}
						serverWrkAcks <- r.Work // HL
						// wait for a cancel from c
						getCancel(c, "EXTERNAL", stopLooking, gotCancel)
					}(c, stopLooking)
				} // END  RANGE DIALED
				// wait a bit - drain blockSendDone
				for i := 0; i < numServers; i++ {
					// fmt.Println("blocksenddone")
					<-blockSendDone
				}
				//  collect the work request acks from servers
				for _, c := range dialedServers {
					if isDead(c) {
						debugF("server DOWN: %v\n", c)
						continue
					}
					// fmt.Println("serverWrkAcks")
					<-serverWrkAcks
					debugF("server up: %v\n", c)
				}
				// OMIT
				debugF("%s\n", "...") // OMIT
				//  wait for server cancellation responses
				for _, c := range dialedServers {
					if isDead(c) {
						continue
					}
					fmt.Println("wait gotCancel ..")
					<-gotCancel // wait for cancellation from each server
					fmt.Println("... gotCancel")
					// ignore[c] <- struct{}{} // issue an ignore .... ???
				}
				fmt.Println("end run")
			}
		}()

	*/

	<-theEnd // so that we can monitor ..
}

func main1() {
	flag.Parse()
	myServers := checkMandatoryF()
	numServers = len(myServers)
	active = make(map[cpb.CoinClient]chan (struct{}))
	for index := 0; index < numServers; index++ {
		addr := fmt.Sprintf("%s:%d", myServers[index].host, 50051+myServers[index].port)
		conn, err := grpc.Dial(addr, grpc.WithInsecure()) // HL
		if err != nil {
			log.Fatalf("fail to dial: %v", err)
		}
		defer conn.Close()
		c := cpb.NewCoinClient(conn) // note that we do not login!
		dialedServers = append(dialedServers, c)
		active[c] = make(chan struct{}, 1)
		active[c] <- struct{}{} //  is alive!
	}

	// winner := make(chan result)
	stopLooking := make(chan string, numServers) // for search

	tick := time.Tick(1 * time.Second)
	theEnd := make(chan struct{})

	serverWrkAcks := make(chan *cpb.Work, numServers) // for gathering signin
	blockSendDone := make(chan struct{}, numServers)  // for the next loop
	gotCancel := make(chan struct{}, numServers)      // for the next loop
	theWinner := make(chan string)                    //  OMIT

	// external
	// here models external net: timeout after timeOut seconds
	// now stopLooking is filled with the stringified server or ""
	go func() {
		for {
			quit := false
			for cn := 0; ; cn++ {
				if quit {
					break
				}
				if cn >= *timeOut {
					// winner <- result{nonce: uint32(cn), ours: false, cstring: ""}
					declareWin(theWinner, -1, "external", uint32(cn))
					quit = true //break
				}
				if !quit {
					// check for a stop order
					select {
					// case cs := <-stopLooking: // expect that getResult sets theWinner here
					case <-stopLooking: // expect that getResult sets theWinner here
						// winner <- result{nonce: 0, ours: true, cstring: cs}
						// fmt.Println("stoplooking ...")
						//go mopup(stopLooking)
						quit = true
					default: // continue
					}
				}
				if !quit {
					// wait for a second here ...
					<-tick
					debugF(" | EXT %d\n", cn)
				}
			}
		}
	}()

	// loop
	// issue blocks, get the 'runs' going
	go func() {
		for { // ISSUE BLOCK
			// wait until we have a winner
			fmt.Println(<-theWinner, "\n---------------------------") // a OMIT
			// <-winner // wait until we have one
			// this is where we declare the winner
			// mopup(stopLooking)
			// then issue new block
			u, l, blk, m, h, bts := newBlock() // next block
			// make(chan struct{}, numServers)
			// make map()

			ignore := make(map[cpb.CoinClient]chan (struct{}))
			for _, c := range dialedServers { // RANGE DIALED
				go func(c cpb.CoinClient) {
					// start with block issue
					_, err := c.IssueBlock(context.Background(), // HL
						&cpb.IssueBlockRequest{ // HL
							Upper:       u,   // OMIT
							Lower:       l,   // OMIT
							Block:       blk, // OMIT
							Merkle:      m,   // OMIT
							Blockheight: h,   // OMIT
							Bits:        bts})
					if skipF(c, "could not issue block", err) {
						blockSendDone <- struct{}{}
						return
					}
					blockSendDone <- struct{}{}
					// get ready, get set ... this blocks!  OMIT
					r, err := c.GetWork(context.Background(), // HL
						&cpb.GetWorkRequest{Name: "EXTERNAL"}) // HL
					if skipF(c, "could not reconnect", err) { // HL
						return
					} else if isDead(c) { // were we previously decalred dead? change that ..
						active[c] <- struct{}{} // 'revive' us
					}
					serverWrkAcks <- r.Work // HL
					// conductor handles results OMIT
					// ignore[c] = make(chan struct{})                   // initalise the ignore channel
					// go getResult(c, "EXTERNAL", theWinner, ignore[c]) // HL
					// in parallel - seek cancellation OMIT
					go getCancel(c, "EXTERNAL", stopLooking, gotCancel)
				}(c) // (c, stopLooking, endLoop, theWinner, lateEntry)

			} // END RANGE DIALED
			// wait a bit - drain blockSendDone
			for i := 0; i < numServers; i++ {
				// fmt.Println("blocksenddone")
				<-blockSendDone
			}
			//  collect the work request acks from servers
			for _, c := range dialedServers {
				if isDead(c) {
					debugF("server DOWN: %v\n", c)
					continue
				}
				// fmt.Println("serverWrkAcks")
				<-serverWrkAcks
				debugF("server up: %v\n", c)
			}
			// OMIT
			debugF("%s\n", "...") // OMIT
			//  wait for server cancellation responses
			for _, c := range dialedServers {
				if isDead(c) {
					continue
				}
				<-gotCancel // wait for cancellation from each server
				fmt.Println("gotCancel")
				ignore[c] <- struct{}{} // issue an ignore .... ???
			}
		} // END ISSUE BLOCK
	}()

	<-theEnd // so that we can monitor ..
}

// utilities

// skipF is a per connection function. logs message and returns true if err
// otherwise returns false. it also maintains the alive[] map
func skipF(c cpb.CoinClient, message string, err error) bool {
	if err != nil {
		log.Printf("SF: "+message+": %v", err)
		if !isDead(c) {
			<-active[c] // drain active[c] so that we are truly dead
		}
		return true // we have skipped
	}
	return false
}

func debugF(format string, args ...interface{}) {
	if *debug {
		log.Printf(format, args...)
	}
}

func checkMandatoryF() []server {
	if *servers == "" {
		log.Fatalf("%s\n", "Conductor must set servers. Use -s switch")
	}
	var servList []server
	cl := strings.Split(*servers, ",")
	for _, v := range cl {
		w := strings.Split(v, ":")
		s := w[0]
		p, err := strconv.Atoi(w[1])
		if err != nil {
			log.Fatalf("Servers type error : %s\n%v\n", w, err)
		}
		servList = append(servList, server{s, p})
	}
	return servList
}

//=================================== OLDER ==============================

// 'search' here models external net: timeout after timeOut seconds
func search(stopLooking chan struct{}) (uint32, bool) {
	var theNonce uint32
	var ok bool
	tick := time.Tick(1 * time.Second)
	for cn := 0; ; cn++ {
		if cn >= *timeOut {
			theNonce = uint32(cn)
			ok = true
			break
		}
		// check for a stop order
		select {
		case <-stopLooking:
			return theNonce, ok
		default: // continue
		}
		// wait for a second here ...
		<-tick
		debugF(" | EXT %d\n", cn)
	}
	return theNonce, ok
}

// getCancelOLD makes a blocking request to the server
func getCancelOLD(c cpb.CoinClient, name string, stopLooking chan struct{}, endLoop chan struct{}) {
	_, err := c.GetCancel(context.Background(), &cpb.GetCancelRequest{Name: name})
	if skipF(c, "could not request cancellation", err) {
		return
	}
	stopLooking <- struct{}{} // stop search  // <- fmt.Sprintf("%v",c)
	endLoop <- struct{}{}     // quit loop
}

// getResultOLD makes a blocking request to the server
func getResultOLD(c cpb.CoinClient, name string, theWinner chan string, lateEntry chan struct{}) {
	res, err := c.GetResult(context.Background(), &cpb.GetResultRequest{Name: name})
	if skipF(c, "could not request result", err) {
		return
	}

	if res.Winner.Identity != "EXTERNAL" { // avoid echoes
		declareWinOLD(theWinner, lateEntry, int(res.Index), res.Winner.Identity, res.Winner.Nonce) // HL
	}
}

func declareWinOLD(theWinner chan string, lateEntry chan struct{},
	index int, coinbase string, nonce uint32) {
	// index uint32, coinbase string, nonce uint32) {
	select {
	case <-lateEntry: // we already have declared a winner, do nothing
	default:
		close(lateEntry) // HL
		str := fmt.Sprintf("%s - ", time.Now().Format("15:04:05"))
		if index == -1 { //indicates winner is external
			str += "external" // HL
		} else {
			str += fmt.Sprintf("miner %d:%s, nonce %d", index, coinbase, nonce)
		}
		theWinner <- str // HL
		for i, c := range dialedServers {
			if isDead(c) {
				continue
			}
			if i == index {
				fmt.Println("DON'T TELL YOURSELF: ", i, c)
				continue
			}
			annouceWin(c, 99, []byte{}, "EXTERNAL") // bogus  announcement
		}
	}
}

func mainOLD() {
	flag.Parse()
	myServers := checkMandatoryF()
	numServers = len(myServers)
	active = make(map[cpb.CoinClient]chan (struct{}))
	for index := 0; index < numServers; index++ {
		addr := fmt.Sprintf("%s:%d", myServers[index].host, 50051+myServers[index].port)
		conn, err := grpc.Dial(addr, grpc.WithInsecure()) // HL
		if err != nil {
			log.Fatalf("fail to dial: %v", err)
		}
		defer conn.Close()
		c := cpb.NewCoinClient(conn) // note that we do not login!
		dialedServers = append(dialedServers, c)
		active[c] = make(chan struct{}, 1)
		active[c] <- struct{}{} //  is alive!
	}

	blockSendDone := make(chan struct{}, numServers) // for the next loop
	stopL := make(chan struct{}, numServers)         // for search OMIT

	// OMIT
	for {
		// stopLooking := make(chan struct{}, numServers)    // for search OMIT
		endLoop := make(chan struct{}, numServers)        // for this loop OMIT
		serverWrkAcks := make(chan *cpb.Work, numServers) // for gathering signins OMIT
		lateEntry := make(chan struct{})                  // no more results please OMIT
		theWinner := make(chan string, numServers)        //  OMIT
		u, l, blk, m, h, bts := newBlock()                // next block

		for _, c := range dialedServers {
			go func(c cpb.CoinClient, // HL
				stopLooking chan struct{}, endLoop chan struct{},
				theWinner chan string, lateEntry chan struct{}) {
				// start with block issue
				_, err := c.IssueBlock(context.Background(), // HL
					&cpb.IssueBlockRequest{ // HL
						Upper:       u,   // OMIT
						Lower:       l,   // OMIT
						Block:       blk, // OMIT
						Merkle:      m,   // OMIT
						Blockheight: h,   // OMIT
						Bits:        bts})
				if skipF(c, "could not issue block", err) {
					blockSendDone <- struct{}{}
					return
				}
				blockSendDone <- struct{}{}
				// get ready, get set ... this blocks!  OMIT
				r, err := c.GetWork(context.Background(), // HL
					&cpb.GetWorkRequest{Name: "EXTERNAL"}) // HL
				if skipF(c, "could not reconnect", err) { // HL
					return
				} else if isDead(c) { // were we previously decalred dead? change that ..
					active[c] <- struct{}{} // 'revive' us
				}
				serverWrkAcks <- r.Work // HL
				// conductor handles results OMIT
				go getResultOLD(c, "EXTERNAL", theWinner, lateEntry) // HL
				// in parallel - seek cancellation OMIT
				go getCancelOLD(c, "EXTERNAL", stopLooking, endLoop)
			}(c, stopL, endLoop, theWinner, lateEntry) // (c, stopLooking, endLoop, theWinner, lateEntry)
		}
		// wait a bit - drain blockSendDone
		for i := 0; i < numServers; i++ {
			fmt.Println("blocksenddone")
			<-blockSendDone
		}
		//  collect the work request acks from servers b OMIT
		for _, c := range dialedServers {
			if isDead(c) {
				debugF("server DOWN: %v\n", c)
				continue
			}
			fmt.Println("serverWrkAcks")
			<-serverWrkAcks
			debugF("server up: %v\n", c)
		}
		// OMIT
		debugF("%s\n", "...") // OMIT
		// 'search' - as the common 'External' miner
		theNonce, ok := search(stopL) // search(stopLooking)
		if ok {
			declareWinOLD(theWinner, lateEntry, -1, // HL
				"external", theNonce)
		}
		// drain stopL // stopLooking
		done := false
		for !done {
			select {
			case <-stopL: //stopLooking:
				fmt.Println("STOPLOOKING")
			default:
				done = true
			}
		}

		//  wait for server cancellation responses
		for _, c := range dialedServers {
			if isDead(c) {
				continue
			}
			fmt.Println("endLoop")
			<-endLoop // wait for cancellation from each server
		}
		//  OMIT
		fmt.Println(<-theWinner, "\n---------------------------") // a OMIT
	}
} // c OMIT
