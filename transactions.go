// Package coin has the code for manipulating bitcoin transactions
package coin

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"golang.org/x/crypto/ripemd160"
)

const (
	poolname        = "2f5a6f6368657a612f" // /Zocheza/
	extralen        = 4                    // number of bytes for the extranonce
	mineridlen      = 3                    // number of bytes for the miner id
	minerhashlen    = 20                   // number of bytes for miner user hash
	posOutIndex     = 37                   // position of the output index data
	posLenScriptSig = 41                   // position of the length srciptsig data
)

// BTC is the number of satoshi in a single bitcoin : 10^8
const BTC = 100000000

// HalvingInterval is how often the 'subsidy' or reward is halved, in blocks
const HalvingInterval = 210000

/*
The only convention followed in contructing the coinbase 'scritpsig' is that it carry
the block height at teh beginning - in lower Endian - together with the number of bytes
this occupies, so bh = 277316 => coinbasedata starts with 003443b04, since 0x043b44
is hex for 277316. The only other limitation is that the script has a max length 100
and that the following data ususally encodes the 'extra nonce' and the identity of the
mining pool. We will thus fix teh following:
	bhlen - 1 byte
	bh - 3/4 bytes
	extranonce - 4 bytes
	miner id - 3 bytes
	miner hash - 20 bytes
	pool identity - 9 bytes
where
1. miner hash is encoded using hash160 - so the true extranonce is
miner = hash160(miner identity+extranonce)
this makes verifiaction of teh winner easy all around: users will have a directory of
miner identities, and we assume fewer than 2^24 = 16,777,216 miners and that the speed
of each will be such that we do not need to exhaust all 2^32 = 4,294,967,296 extra
nonce space.
2. the pool identity is fixed at /Zocheza/ --> 2f5a6f6368657a612f has 9 bytes
we will set this as a 'const'
3. what public methods?
	coinbaseData(bh, enonce, minerid) - will require map: minerid -> miner identity
	extract(key string) - to get bh, extranonce, miner id, minerhash
	IncrementNonce() - increment extra nonce, return new coinbasedata
*/

// coinbaseData is the alternative to scriptSig ("unlocking" script) in a coinbase
func coinbaseData(bh int, extra int, minerid int, miner string) ([]byte, error) {
	bhlen := 4                    // max length of blockheight data
	bhMaxBytes := make([]byte, 4) // will accomodate largest possible blockheighth of 500 million
	binary.LittleEndian.PutUint32(bhMaxBytes, uint32(bh))
	// decide the actual length required
	for bhMaxBytes[bhlen-1] == 0 {
		bhlen--
	}
	// the desired slice
	blockHeight := bhMaxBytes[0:bhlen]
	// extranonce
	extranonce := make([]byte, extralen) // 4 bytes
	binary.LittleEndian.PutUint32(extranonce, uint32(extra))
	// miner id
	minerIDBytes := make([]byte, mineridlen)
	temp := make([]byte, 4)
	binary.LittleEndian.PutUint32(temp, uint32(minerid))
	copy(minerIDBytes[:], temp[0:3]) // use the three bytes
	// minerhash
	var hashbuffer bytes.Buffer
	hashbuffer.Write(extranonce)
	hashbuffer.Write([]byte(miner))
	minerHashBytes, err := Hash160(hashbuffer.Bytes())
	if err != nil {
		return nil, err
	}
	// pool identity - /Zocheza/
	poolIdentity, err := hex.DecodeString(poolname)
	if err != nil {
		return nil, err
	}
	// assemble the bytes
	var buffer bytes.Buffer
	buffer.WriteByte(byte(bhlen))
	buffer.Write(blockHeight)
	buffer.Write(extranonce)
	buffer.Write(minerIDBytes)
	buffer.Write(minerHashBytes)
	buffer.Write(poolIdentity)

	return buffer.Bytes(), nil
}

// GenCoinbase creates a  coinbase transaction given slices upperTemplate, lowerTemplate
// which represent the parts of the txn aside from the coinbasedata
func GenCoinbase(upperTemplate []byte, lowerTemplate []byte,
	blockHeight int, miner int, minerstring string) ([]byte, error) {
	// contruct the coinbasedata, extranince=0
	coinbasedata, err := coinbaseData(blockHeight, 0, miner, minerstring)
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	buffer.Write(upperTemplate)
	buffer.WriteByte(byte(len(coinbasedata)))
	buffer.Write(coinbasedata)
	buffer.Write(lowerTemplate)

	return buffer.Bytes(), nil
}

// CoinbaseTemplates is what the server uses to deploy the upper & lower templates
func CoinbaseTemplates(blockHeight int, blockFees int, pubkey string) (upper, lower []byte, err error) {
	var upperBuffer, lowerBuffer bytes.Buffer

	//Version field
	version := make([]byte, 4)
	binary.LittleEndian.PutUint32(version, 1)
	// InputTxn - here all zeros
	coinbaseInput := make([]byte, 32) // all 0s
	//Ouput index of input transaction -1 for coinbase
	outputIndexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(outputIndexBytes, 0xffffffff)
	sequence := make([]byte, 4)
	binary.BigEndian.PutUint32(sequence, 0xffffffff)
	locktime := make([]byte, 4) // all 0s here
	//Satoshis to send.
	satoshis := getValue(blockHeight) + blockFees
	amount := make([]byte, 8)
	binary.LittleEndian.PutUint64(amount, uint64(satoshis))
	// outout script
	scriptpubkey, err := P2PKH(pubkey)
	if err != nil {
		return nil, nil, err
	}

	upperBuffer.Write(version)     // littleendian = 1
	upperBuffer.WriteByte(byte(1)) // # inputs
	upperBuffer.Write(coinbaseInput)
	upperBuffer.Write(outputIndexBytes)
	// lowerBuffer
	lowerBuffer.Write(sequence)
	lowerBuffer.WriteByte(byte(1)) // # outputs
	lowerBuffer.Write(amount)      // in satoshis
	lowerBuffer.WriteByte(byte(len(scriptpubkey)))
	lowerBuffer.Write(scriptpubkey)
	lowerBuffer.Write(locktime)

	return upperBuffer.Bytes(), lowerBuffer.Bytes(), nil
}

// calculate the mining reward at this height
func getValue(blockHeight int) int {
	subsidy := 50 * BTC
	halvings := uint(blockHeight / HalvingInterval)
	if halvings >= 64 {
		return 0
	}
	subsidy >>= halvings
	return subsidy
}

// Hash160 performs the same operations as OP_HASH160 in Bitcoin Script
// It hashes the given data first with SHA256, then RIPEMD160
func Hash160(data []byte) ([]byte, error) {
	//Does identical function to Script OP_HASH160. Hash once with SHA-256, then RIPEMD-160
	if data == nil {
		return nil, errors.New("Empty bytes cannot be hashed")
	}
	shaHash := sha256.New()
	shaHash.Write(data)
	hash := shaHash.Sum(nil) // SHA256 first
	ripemd160Hash := ripemd160.New()
	ripemd160Hash.Write(hash)
	hash = ripemd160Hash.Sum(nil) //RIPEMD160 second

	return hash, nil
}

// P2PKH returns the pay-to-public-key-hash script against hex address pubkey
func P2PKH(pubkey string) ([]byte, error) {
	// OP_DUP HASH160 0x14
	opduphash, err := hex.DecodeString("76a914")
	if err != nil {
		return nil, err
	}
	addr, err := hex.DecodeString(pubkey)
	if err != nil {
		return nil, err
	}
	pubkeyBytes, err := Hash160(addr)
	if err != nil {
		return nil, err
	}
	opverifychecksig, err := hex.DecodeString("88ac")
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	buffer.Write(opduphash)
	buffer.Write(pubkeyBytes)
	buffer.Write(opverifychecksig)

	return buffer.Bytes(), nil
}

// littleEndian returns the integer value of littleendian encoded byte sequence v
func littleEndian(v []byte) int {
	val := 0
	for i := 0; i < len(v); i++ {
		pt := int(v[i]) << uint(8*i)
		val += pt
	}
	return val
}

// Transaction type allows to dissemble byte sequence outputs
type Transaction []byte

// IncrementNonce raises the nonce value by 1
func (t Transaction) IncrementNonce() error {
	val, err := t.getNonce()
	if err != nil {
		return err
	}
	val++
	return t.putNonce(val)
}

// putNonce sets the nonce of a coinbase transaction
func (t Transaction) putNonce(nonce int) error {
	// make sure this is a coinbase txn
	if t.outindex() != "ffffffff" {
		return errors.New("not a coinbase transaction")
	}
	offset := t[posLenScriptSig+1]
	start := posLenScriptSig + 2 + offset
	v := t[start : start+4]
	fmt.Printf("current: %d\n", littleEndian(v))
	binary.LittleEndian.PutUint32(v, uint32(nonce))
	return nil
}

// getNonce fetches the nonce of a coinbase transaction
func (t Transaction) getNonce() (int, error) {
	// make sure this is a coinbase txn
	if t.outindex() != "ffffffff" {
		return -1, errors.New("not a coinbase transaction")
	}
	// the coinbase data really scriptsig
	ss := t.scriptsig()
	n := int(ss[0])        // either 3 or 4 - bytes length of hex(blockheight)
	nonce := ss[n+1 : n+5] // followed directly by the extranonce

	return littleEndian(nonce), nil
}

// .entry returns the contents at posiion n (>=0) as an integer in [0,255]
func (t Transaction) entry(n int) int {
	return int(t[n])
}

// .get returns the  transaction's slice from s to s+length-1 (an array of size length)
func (t Transaction) get(s, length int) []byte {
	return t[s : s+length]
}

// .detail displays information about the transaction
func (t Transaction) detail() {
	pos := 0
	fmt.Printf("length: %d\n---------\n", len(t))
	fmt.Printf("version: %v\n", littleEndian(t.get(0, 4)))
	pos += 4
	fmt.Printf("input count: %v\n", t.get(pos, 1))
	pos++
	fmt.Printf("input: %x\n", t.get(pos, 32))
	pos += 32
	fmt.Printf("output index: %x\n", t.get(pos, 4))
	pos += 4
	v := t.entry(pos)
	fmt.Printf("length scriptsig: %d\n", t.get(pos, 1))
	pos++
	fmt.Printf("scriptsig: %x\n", t.get(pos, v))
	pos += v
	fmt.Printf("sequence: %x\n", t.get(pos, 4))
	pos += 4
	fmt.Printf("output count: %d\n", t.get(pos, 1))
	pos++
	fmt.Printf("satoshi: %d\n", littleEndian(t.get(pos, 8)))
	pos += 8
	v = t.entry(pos)
	fmt.Printf("length scriptpubkey: %d\n", t.get(pos, 1))
	pos++
	fmt.Printf("scriptpubkey: %x\n", t.get(pos, v))
	pos += v
	fmt.Printf("locktime: %x\n", t.get(pos, 4))
}

func (t Transaction) scriptsig() []byte {
	length := t.entry(posLenScriptSig)
	return t.get(posLenScriptSig+1, length)
}

func (t Transaction) outindex() string {
	return fmt.Sprintf("%x", t.get(posOutIndex, 4))
}
