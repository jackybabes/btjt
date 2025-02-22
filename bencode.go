package main

import (
	"bufio"
	"bytes"
	"log"
)

func getBytes(reader *bufio.Reader) []byte {
	var lengthBytes []byte
	for {
		b, _ := reader.ReadByte()
		if b == ':' {
			break
		}
		lengthBytes = append(lengthBytes, b)
	}

	length := bytesToInt(lengthBytes)
	bytesBuffer := make([]byte, length)
	_, _ = reader.Read(bytesBuffer)

	// log.Printf("%v, %v", length, string(bytesBuffer))
	return bytesBuffer
}

func getInt(reader *bufio.Reader) int {
	var intBytes []byte
	for {
		b, _ := reader.ReadByte()
		if b == 'i' {
			continue
		}
		if b == 'e' {
			break
		}
		intBytes = append(intBytes, b)
	}
	i := bytesToInt(intBytes)
	return i
}

func getList(reader *bufio.Reader, infoSection *KeyIndex) []interface{} {
	var listData []interface{}

	_, _ = reader.ReadByte()
	for {
		nextBytes, _ := reader.Peek(1)
		nextByte := nextBytes[0]

		switch nextByte {
		case 'i':
			// IntType
			listData = append(listData, getInt(reader))
		case 'l':
			// ListType
			listData = append(listData, getList(reader, infoSection))
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// ByteType
			listData = append(listData, getBytes(reader))
		case 'd':
			// DictType
			listData = append(listData, getDict(reader, infoSection))
		case 'e':
			_, _ = reader.ReadByte()
			return listData
		}
	}
}

func getDict(reader *bufio.Reader, keyIndex *KeyIndex) map[string]interface{} {
	dictData := make(map[string]interface{})
	_, _ = reader.ReadByte()
	// KEY CHECEK VAR
	var thisLevelKeyIndex bool
	for {
		// get key
		nextBytes, _ := reader.Peek(1)
		nextByte := nextBytes[0]

		if nextByte == 'e' {
			// if thisLevelInfoSection {
			// 	infoSection.BufferedWhenEnd = reader.Buffered()
			// 	thisLevelInfoSection = false
			// }
			_, _ = reader.ReadByte()
			return dictData
		}

		var key string
		switch nextByte {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// ByteType
			key = string(getBytes(reader))
		default:
			log.Fatalln("key must be string")
		}

		// KEY CHECK START
		if keyIndex.Enabled {
			if key == keyIndex.Key {
				keyIndex.BufferedWhenStart = reader.Buffered()
				thisLevelKeyIndex = true
			}
		}

		// get value
		nextBytes, _ = reader.Peek(1)
		nextByte = nextBytes[0]

		var value interface{}

		switch nextByte {
		case 'i':
			// IntType
			value = getInt(reader)
		case 'l':
			// ListType
			value = getList(reader, keyIndex)
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// ByteType
			value = getBytes(reader)
		case 'd':
			// DictType
			value = getDict(reader, keyIndex)
		}

		dictData[key] = value
		// KEY CHECK END
		if thisLevelKeyIndex {
			keyIndex.BufferedWhenEnd = reader.Buffered()
			thisLevelKeyIndex = false
		}
	}
}

type KeyIndex struct {
	Enabled           bool
	Key               string
	BufferedWhenStart int
	BufferedWhenEnd   int
}

func bencode_decode(data []byte) map[string]interface{} {
	reader := bufio.NewReaderSize(bytes.NewReader(data), len(data))
	var decodedData map[string]interface{}

	keyIndex := KeyIndex{Enabled: false}

	decodedData = getDict(reader, &keyIndex)

	return decodedData

}

func bencode_decode_byte_data_of_key(data []byte, key string) (map[string]interface{}, []byte) {
	reader := bufio.NewReaderSize(bytes.NewReader(data), len(data))
	var decodedData map[string]interface{}

	keyIndex := KeyIndex{Key: key, Enabled: true}

	decodedData = getDict(reader, &keyIndex)

	keyValue := data[len(data)-keyIndex.BufferedWhenStart : len(data)-keyIndex.BufferedWhenEnd]

	return decodedData, keyValue
}
