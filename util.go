package main

import (
	"encoding/hex"
	"os"
	"strconv"
)

func readTorrent(filename string) []byte {
	data, _ := os.ReadFile(filename)
	return data
}

func bytesToInt(b []byte) int {
	s := string(b)
	i, _ := strconv.Atoi(s)
	return i
}

// func checkSingleMapInside(i []interface{}) bool {
// 	if len(i) == 1 {
// 		if _, ok := i[0].(map[string]interface{}); ok {
// 			return true
// 		}
// 	}
// 	return false
// }

func hashURLEncode(h [20]byte) string {
	// Note that all binary data in the URL (particularly info_hash and peer_id) must be properly escaped.
	// This means any byte not in the set 0-9, a-z, A-Z, '.', '-', '_' and '~', must be encoded using the "%nn" format,
	// where nn is the hexadecimal value of the byte. (See RFC1738 for details.)

	// For a 20-byte hash of \x12\x34\x56\x78\x9a\xbc\xde\xf1\x23\x45\x67\x89\xab\xcd\xef\x12\x34\x56\x78\x9a,
	// The right encoded form is %124Vx%9A%BC%DE%F1%23Eg%89%AB%CD%EF%124Vx%9A

	var encoded string
	for _, b := range h {

		if (b >= '0' && b <= '9') || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || b == '.' || b == '-' || b == '_' || b == '~' {
			encoded += string(b)
			continue
		}

		hexOfByte := hex.EncodeToString([]byte{b})
		encoded += "%" + hexOfByte

	}
	return encoded
}
