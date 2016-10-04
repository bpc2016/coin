package txn

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNewRawTransaction(t *testing.T) {
	testInputTx := "3ad337270ac0ba14fbce812291b7d95338c878709ea8123a4d88c3c29efbc6ac"
	testAmount := 65600
	testOutputindex := 0
	testScriptSig := []byte{118, 169, 20, 146, 3, 228, 122, 22, 247, 153, 222, 208, 53, 50, 227, 228, 82, 96, 111, 220, 82, 0, 126, 136, 172}
	testScriptPubKey := []byte{169, 20, 26, 139, 0, 38, 52, 49, 102, 98, 92, 116, 117, 240, 30, 72, 181, 237, 232, 192, 37, 46, 135}
	testRawTx := []byte{1, 0, 0, 0, 1, 172, 198, 251, 158, 194, 195, 136, 77, 58, 18, 168, 158, 112, 120, 200, 56, 83, 217, 183, 145, 34, 129, 206, 251, 20, 186, 192, 10, 39, 55, 211, 58, 0, 0, 0, 0, 25, 118, 169, 20, 146, 3, 228, 122, 22, 247, 153, 222, 208, 53, 50, 227, 228, 82, 96, 111, 220, 82, 0, 126, 136, 172, 255, 255, 255, 255, 1, 64, 0, 1, 0, 0, 0, 0, 0, 23, 169, 20, 26, 139, 0, 38, 52, 49, 102, 98, 92, 116, 117, 240, 30, 72, 181, 237, 232, 192, 37, 46, 135, 0, 0, 0, 0}

	rawTx, err := NewRawTransaction(testInputTx, testAmount, testOutputindex, testScriptSig, testScriptPubKey)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(rawTx, testRawTx) {
		fmt.Printf("Raw transaction different from expected transaction.\n%v\n%v\n", testRawTx, rawTx)
	}

	L := len(rawTx)

	piece(0, 3, rawTx)        // version - 4 bytes
	piece(4, 4, rawTx)        // number - 1 byte
	piece(5, 32+5-1, rawTx)   // input tx hash - 32 bytes
	piece(37, 40, rawTx)      // output index (starts at 0) - 4 bytes
	piece(41, 41, rawTx)      // length of following script = 25 or 0x19 here
	piece(42, 42+25-1, rawTx) // script sig "unlocking script"
	piece(67, 67+4-1, rawTx)  // sequence set at 0xffffff -4 bytes
	piece(71, 71, rawTx)      // number of outputs - 1 byte
	piece(72, 72+8-1, rawTx)  // satoshis : little endian - 8 bytes
	piece(80, 80, rawTx)      // length of following script = 23 - 1 (variable)
	piece(81, 81+23-1, rawTx) // script pub key "locking script"
	piece(104, 107, rawTx)    // locktime - 4 byte

	fmt.Printf("Length: %d\nAmount: %x\n", L, testAmount)

}

func piece(a int, b int, v []byte) {
	fmt.Printf(">> %d - %d\n%v\n", a, b, v[a:b+1])
}

func TestCoinbase(t *testing.T) {
	// testInputTx := "0000000000000000000000000000000000000000000000000000000000000000" // coinbase
	// testAmount := 2509094928                                                          // satoshi  25.09094928 BTC
	// testOutputindex := -1
	// testBlockHeight := 277316
	// testExtraNonce := 0x858402
	// testUserName := "busiso"
	// testCoinBaseData, err := coinbaseData(testBlockHeight, testExtraNonce, testUserName)
	// if err != nil {
	// 	t.Error(err)
	// }
	// testScriptPubKey, err := P2PKH("0225c141d69b74adac8ab984a8eb9fee42c4ce79cf6cb2be166b1ddc0356b37086") //hex.DecodeString("2102aa970c592640d19de03ff6f329d6fd2eecb023263b9ba5d1b81c29b523da8b21ac")
	// if err != nil {
	// 	t.Error(err)
	// }
	// testRawTx, err := hex.DecodeString("01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0f03443b0403858402062f503253482fffffffff0110c08d9500000000232102aa970c592640d19de03ff6f329d6fd2eecb023263b9ba5d1b81c29b523da8b21ac00000000")
	// if err != nil {
	// 	t.Error(err)
	// }

	rawTx, err := NewCoinBase(277316, 9094928, "0225c141d69b74adac8ab984a8eb9fee42c4ce79cf6cb2be166b1ddc0356b37086", 0x858402, "user:busiso")
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("rawtx:\n%x\n", rawTx)

	// if !reflect.DeepEqual(rawTx, testRawTx) {
	// 	fmt.Printf("Raw transaction different from expected transaction.\n%v\n%v\n", testRawTx, rawTx)
	// }

	// res, _ := P2PKH("0225c141d69b74adac8ab984a8eb9fee42c4ce79cf6cb2be166b1ddc0356b37086")

	// fmt.Printf("p2kh :\n%x\n", res)
}

// 15 - length of script sig
// [3 68 59 4 3 133 132 2 6 47 80 50 83 72 47] - scriptsig  itself
// 35 - len scriptpubkey
// [33 2 170 151 12 89 38 64 209 157 224 63 246 243 41 214 253 46 236 176 35 38 59 155 165 209 184 28 41 181 35 218 139 33 172]

// raw: 01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0f03443b0403858402062f503253482fffffffff0110c08d9500000000232102aa970c592640d19de03ff6f329d6fd2eecb023263b9ba5d1b81c29b523da8b21ac00000000
// act: 01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0f03443b0403858402062f503253482fffffffff0110c08d9500000000232102aa970c592640d19de03ff6f329d6fd2eecb023263b9ba5d1b81c29b523da8b21ac00000000

// from ... http://bitcoin.stackexchange.com/questions/21557/how-to-fully-decode-a-coinbase-transaction

// coinbase: 038abd07062f503253482f048725ee5208083865a409000000092f7374726174756d2f
// 03 - length opcode
// 8abd07 - data with length 03
// 06 - length opcode
// 2f503253482f - data with length 06
// 04 - length opcode
// 8725ee52 - data with length 04
// 08 - length opcode
// 083865a409000000 - data with length 08
// 09 - length opcode
// 2f7374726174756d2f - data with length 09

// scriptPubKey: 76a914975efcba1e058667594dc57146022ec46560a63c88ac
// 76 - OP_DUP opcode
// a9 - HASH160 opcode
// 14 - length opcode
// 975efcba1e058667594dc57146022ec46560a63c - data with length 14 (20 in dec)
// 88 - OP_EQUALVERIFY opcode
// ac - OP_CHECKSIG opcode

// 132xe93LdrdGa39vN7su1shRpcBwMdAX4J - wif
// 0225c141d69b74adac8ab984a8eb9fee42c4ce79cf6cb2be166b1ddc0356b37086 - pubkey
// 164f1d1d6fce7e2e491352b95b4ea47b880c1546 - after Hash160
// KyufBz2L22mZgxgeftJuDK7Fot4rMarX4sQ7v5SNE9eZhq1wSqVf - privkey
