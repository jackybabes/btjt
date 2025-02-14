package main

import (
	"log"
	"os"
	"strconv"
)

func readTorrent(filename string) []byte {
	data, _ := os.ReadFile(filename)
	return data
}

type BencodeType int

const (
	ByteType BencodeType = iota
	ListType
	DictType
	IntType
	NoneType
)

func checkNextDataStructure(b byte) BencodeType {
	switch b {
	case 'd':
		return DictType
	case 'i':
		return IntType
	case 'l':
		return ListType
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return ByteType
	default:
		return NoneType
	}
}

func bytesToInt(b []byte) int {
	s := string(b)
	i, _ := strconv.Atoi(s)
	return i
}

func getBytes(data []byte) ([]byte, []byte) {

	var lengthBytes []byte
	for _, c := range data {
		if c == ':' {
			break
		}
		lengthBytes = append(lengthBytes, c)
	}
	length := bytesToInt(lengthBytes)
	bytes := data[len(lengthBytes)+1 : len(lengthBytes)+1+length]

	// log.Printf("%v, %v", length, string(bytes))

	data = data[len(lengthBytes)+1+length:]

	return bytes, data

}

func getInt(data []byte) (int, []byte) {
	var intBytes []byte
	for _, b := range data[1:] {
		if b == 'e' {
			break
		}
		intBytes = append(intBytes, b)
	}
	i := bytesToInt(intBytes)
	data = data[2+len(intBytes):]
	return i, data
}

func getList(data []byte) ([]interface{}, []byte) {
	var listData []interface{}
	data = data[1:]
	for {
		nextByte := data[0]
		nextDataStructure := checkNextDataStructure(nextByte)
		if nextDataStructure == ByteType {
			var bytes []byte
			bytes, data = getBytes(data)
			listData = append(listData, bytes)
		} else if nextDataStructure == IntType {
			var interger int
			interger, data = getInt(data)
			listData = append(listData, interger)
		} else if nextDataStructure == ListType {
			var list []interface{}
			list, data = getList(data)
			listData = append(listData, list)
		} else if nextDataStructure == DictType {
			var dict map[string]interface{}
			dict, data = getDict(data)
			listData = append(listData, dict)
		} else {
			panic("next data undefined in list")
		}
		if data[0] == 'e' {
			break
		}
	}
	return listData, data[1:]
}

func getDict(data []byte) (map[string]interface{}, []byte) {

	dictData := make(map[string]interface{})

	data = data[1:]
	for {
		// get key
		nextByte := data[0]
		nextDataStructure := checkNextDataStructure(nextByte)
		var key string

		if nextDataStructure == ByteType {
			var bytes []byte
			bytes, data = getBytes(data)
			key = string(bytes)
		} else {
			panic("Not key")
		}
		// get value
		nextByte = data[0]
		nextDataStructure = checkNextDataStructure(nextByte)

		var value interface{}

		if nextDataStructure == ByteType {
			var bytes []byte
			bytes, data = getBytes(data)
			value = bytes
		} else if nextDataStructure == IntType {
			var interger int
			interger, data = getInt(data)
			value = interger
		} else if nextDataStructure == ListType {
			value, data = getList(data)
		} else if nextDataStructure == DictType {
			value, data = getDict(data)
		} else {
			panic("Could not find value")
		}
		// store in dict
		dictData[key] = value

		if data[0] == 'e' {
			break
		}

	}
	return dictData, data[1:]
}

func main() {
	torrentFile := readTorrent("./sample.torrent")
	log.Printf("%v character file", len(torrentFile))
	// log.Println(string(torrentFile))

	data := torrentFile

	var torrentData []interface{}

	for len(data) > 0 {
		nextByte := data[0]
		nextDataStructure := checkNextDataStructure(nextByte)
		if nextDataStructure == ByteType {
			var bytes []byte
			bytes, data = getBytes(data)
			torrentData = append(torrentData, bytes)
		} else if nextDataStructure == IntType {
			var interger int
			interger, data = getInt(data)
			torrentData = append(torrentData, interger)
		} else if nextDataStructure == ListType {
			var list []interface{}
			list, data = getList(data)
			torrentData = append(torrentData, list)
		} else if nextDataStructure == DictType {
			var dict map[string]interface{}
			dict, data = getDict(data)
			torrentData = append(torrentData, dict)
		}
	}

	log.Println(torrentData)
}
