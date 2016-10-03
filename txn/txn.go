// Package txn has the code for manipulating bitcoin transactions
package txn

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
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

// coinbaseData is the alternative to scriptSig ("unlocking" script) in a coinbase
func coinbaseData(bh int, extra int, user string) ([]byte, error) {
	totlen := 0 // bytes consumed
	bhlen := 4  // max lenngth of blockheight data

	bhMaxBytes := make([]byte, 4) // will accomodate largest possible blockheighth of 500 million
	binary.LittleEndian.PutUint32(bhMaxBytes, uint32(bh))
	// decide the actual length required
	for bhMaxBytes[bhlen-1] == 0 {
		bhlen--
	}
	lenBlockHeight := make([]byte, 1)
	lenBlockHeight[0] = uint8(bhlen)
	totlen += bhlen + 1
	// the desired slice
	blockHeight := bhMaxBytes[0:bhlen]
	//length of extranonce always 3 in our case)
	lenextra, err := hex.DecodeString("03")
	if err != nil {
		return nil, err
	}
	totlen += 3 + 1
	// extranonce
	extranonce := make([]byte, 3)
	temp := make([]byte, 4)
	binary.BigEndian.PutUint32(temp, uint32(extra))
	copy(extranonce[0:3], temp[1:4])
	// username
	userBytes := []byte(user)
	//length of username
	if len(userBytes) > 100-totlen-1 {
		return nil, errors.New("username too long")
	}
	lenuser := make([]byte, 1)
	lenuser[0] = uint8(len(userBytes))

	var buffer bytes.Buffer
	buffer.Write(lenBlockHeight)
	buffer.Write(blockHeight)
	buffer.Write(lenextra)
	buffer.Write(extranonce)
	buffer.Write(lenuser)
	buffer.Write(userBytes)

	return buffer.Bytes(), nil
}

// NewCoinBase returns a coinbase transaction given blockHeight, blockFees (satoshi), extraNonce and extraData
func NewCoinBase(blockHeight int, blockFees int, extraNonce int, extraData string) ([]byte, error) {
	inputTx := "0000000000000000000000000000000000000000000000000000000000000000" // coinbase
	satoshis := getValue(blockHeight) + blockFees
	scriptpubkey := make([]byte, 1)
	outputIndex := -1
	coinbasedata, err := coinbaseData(blockHeight, extraNonce, extraData)
	if err != nil {
		return nil, err
	}
	return NewRawTransaction(inputTx, satoshis, outputIndex, coinbasedata, scriptpubkey)
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
