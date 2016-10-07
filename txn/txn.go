// Package txn has the code for manipulating bitcoin transactions
package txn

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"golang.org/x/crypto/ripemd160"
)

// NewRawTransaction creates a Bitcoin transaction given inputs, output satoshi amount, outputindex, scriptSig and scriptPubKey
func NewRawTransaction(inputTxHash string, satoshis int, outputindex int, scriptSig []byte, scriptPubKey []byte) ([]byte, error) {
	//Version field
	version, err := hex.DecodeString("01000000")
	if err != nil {
		return nil, err
	}
	//# of inputs (always 1 in our case)
	inputs, err := hex.DecodeString("01")
	if err != nil {
		return nil, err
	}
	//Input transaction hash
	inputTxBytes, err := hex.DecodeString(inputTxHash)
	if err != nil {
		return nil, err
	}
	//Convert input transaction hash to little-endian form
	inputTxBytesReversed := make([]byte, len(inputTxBytes))
	for i := 0; i < len(inputTxBytes); i++ {
		inputTxBytesReversed[i] = inputTxBytes[len(inputTxBytes)-i-1]
	}
	//Ouput index of input transaction. Normally starts from 0 but is -1 for coinbase
	outputIndexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(outputIndexBytes, uint32(outputindex))

	//scriptSig length. To allow scriptSig > 255 bytes, we use variable length integer syntax from protocol spec
	var scriptSigLengthBytes []byte
	if len(scriptSig) < 253 {
		scriptSigLengthBytes = []byte{byte(len(scriptSig))}
	} else {
		scriptSigLengthBytes = make([]byte, 3)
		binary.LittleEndian.PutUint16(scriptSigLengthBytes, uint16(len(scriptSig)))
		copy(scriptSigLengthBytes[1:3], scriptSigLengthBytes[0:2])
		scriptSigLengthBytes[0] = 253 //Signifies that next two bytes are 2-byte representation of scriptSig length

	}
	//sequence_no. Normally 0xFFFFFFFF. Always in this case.
	sequence, err := hex.DecodeString("ffffffff")
	if err != nil {
		return nil, err
	}
	//Numbers of outputs for the transaction being created. Always one in this example.
	numOutputs, err := hex.DecodeString("01")
	if err != nil {
		return nil, err
	}
	//Satoshis to send.
	satoshiBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(satoshiBytes, uint64(satoshis))
	//Lock time field
	lockTimeField, err := hex.DecodeString("00000000")
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	buffer.Write(version)
	buffer.Write(inputs)
	buffer.Write(inputTxBytesReversed)
	buffer.Write(outputIndexBytes)
	buffer.Write(scriptSigLengthBytes)
	buffer.Write(scriptSig)
	buffer.Write(sequence)
	buffer.Write(numOutputs)
	buffer.Write(satoshiBytes)
	buffer.WriteByte(byte(len(scriptPubKey)))
	buffer.Write(scriptPubKey)
	buffer.Write(lockTimeField)

	return buffer.Bytes(), nil
}

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

const extralen = 4      // number of bytes for the extranonce
const mineridlen = 3    // number of bytes for the miner id
const minerhashlen = 20 // number of bytes for miner user hash

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
	poolIdentity, err := hex.DecodeString("2f5a6f6368657a612f")
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

// NewCoinBase returns a coinbase transaction given blockHeight, blockFees (satoshi), extraNonce and miner
func NewCoinBase(blockHeight int, blockFees int, pubkey string, extraNonce int, miner int, minerlist []string) ([]byte, error) {
	inputTx := "0000000000000000000000000000000000000000000000000000000000000000" // string - coinbase input txn
	satoshis := getValue(blockHeight) + blockFees                                 // int - satoshis
	scriptpubkey, err := P2PKH(pubkey)                                            // []byte - script
	if err != nil {
		return nil, err
	}
	outputIndex := -1
	coinbasedata, err := coinbaseData(blockHeight, extraNonce, miner, minerlist[miner])
	if err != nil {
		return nil, err
	}
	return NewRawTransaction(inputTx, satoshis, outputIndex, coinbasedata, scriptpubkey)
}

// FromTemplate creates a  coinbase transaction given slices upperTemplate, lowerTemplate
// which represent the parts of the txn aside from the coinbasedata
func FromTemplate(upperTemplate []byte, lowerTemplate []byte,
	blockHeight int, miner int, minerlist []string) ([]byte, error) {
	// contruct the coinbasedata, extranince=0
	coinbasedata, err := coinbaseData(blockHeight, 0, miner, minerlist[miner])
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

// Templates is what the server uses to deploy the upper & lower templates
func Templates(blockHeight int, blockFees int, pubkey string) (upper, lower []byte, err error) {
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

// BTC is the number of satoshi in a single bitcoin : 10^8
const BTC = 100000000

// HalvingInterval is how often the 'subsidy' or reward is halved, in blocks
const HalvingInterval = 210000

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

// littleEndian returns teh integer value of littleendian encoded byte sequence v
func littleEndian(v []byte) int {
	val := 0
	for i := 0; i < len(v); i++ {
		pt := int(v[i]) << uint(8*i)
		val += pt
	}
	return val
}

// assuming we have base58 encoding, this is how we do it, note
// 76a914 164f1d1d6fce7e2e491352b95b4ea47b880c1546 88ac
func pkhash2wif() {
	// (1) 164F1D1D6FCE7E2E491352B95B4EA47B880C1546 - the entry here
	// (2) 00164F1D1D6FCE7E2E491352B95B4EA47B880C1546 - add network '00' to front
	// (3) 1FA2C09F7505B401B0536F916A0ACD70941E3DBEAD54ACF843A7240C5A1B6101 - doubleSha256
	// (4) 1FA2C09F - grab first 4 bytes
	// (5) 00164F1D1D6FCE7E2E491352B95B4EA47B880C15461FA2C09F - append to (2) above
	// (6) 132xe93LdrdGa39vN7su1shRpcBwMdAX4J - base58 encode
}

// Transaction type allows to dissemble byte sequence outputs
type Transaction []byte

// CoinBase returns the transaction slice - typed as 'coinbase' rather than slice
func CoinBase(blockHeight int, blockFees int, pubkey string, extraNonce int, miner int, minerslist []string) (Transaction, error) {
	res, err := NewCoinBase(blockHeight, blockFees, pubkey, extraNonce, miner, minerslist)
	if err != nil {
		return nil, err
	}
	return Transaction(res), nil
}

// .entry returns the contents at posiion n (>=0) as an integer in [0,255]
func (t Transaction) entry(n int) int {
	return int(t[n])
}

// .get returns the  transaction's slice from s to s+length-1 (an array of size length)
func (t Transaction) get(s, length int) []byte {
	return t[s : s+length]
}

// .put modifies the slice between positions s and s+len(vals)-1. safe at the end of slice - drops vals
func (t Transaction) put(s int, vals []byte) {
	copy(t[s:s+len(vals)], vals)
}

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

// IncrementNonce raises the nonce value by 1
func (t Transaction) IncrementNonce() error {
	val, err := t.getNonce()
	if err != nil {
		return err
	}
	val++
	return t.putNonce(val)
}

// setNonce resets the nnce of a coinbase transaction
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

const posOutIndex = 37     // position of the output index data
const posLenScriptSig = 41 // position of the length srciptsig data

func (t Transaction) scriptsig() []byte {
	length := t.entry(posLenScriptSig)
	return t.get(posLenScriptSig+1, length)
}

func (t Transaction) outindex() string {
	return fmt.Sprintf("%x", t.get(posOutIndex, 4))
}
