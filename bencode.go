package main

import (
	"bufio"
	"bytes"
	"io"
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

func getList(reader *bufio.Reader) []interface{} {
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
			listData = append(listData, getList(reader))
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// ByteType
			listData = append(listData, getBytes(reader))
		case 'd':
			// DictType
			listData = append(listData, getDict(reader))
		case 'e':
			_, _ = reader.ReadByte()
			return listData
		}
	}
}

func getDict(reader *bufio.Reader) map[string]interface{} {
	dictData := make(map[string]interface{})
	_, _ = reader.ReadByte()

	for {
		// get key
		nextBytes, _ := reader.Peek(1)
		nextByte := nextBytes[0]

		if nextByte == 'e' {
			_, _ = reader.ReadByte()
			if infoSection {
				// log.Println(reader.Buffered())
				infoSectionEnd = bufferLength - reader.Buffered()
				infoSection = false
			}
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

		if key == "info" {
			infoSectionStart = bufferLength - reader.Buffered()
			infoSection, infoSectionExist = true, true
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
			value = getList(reader)
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// ByteType
			value = getBytes(reader)
		case 'd':
			// DictType
			value = getDict(reader)
		}
		dictData[key] = value
	}
}

func bencode_decode(data []byte) []interface{} {
	readerSize := 1024 * 1024 * 4 // 4MB Max Files Size
	reader := bufio.NewReaderSize(bytes.NewReader(data), readerSize)

	_, _ = reader.Peek(1)
	bufferLength = reader.Buffered()

	var decodedData []interface{}
	for {
		nextBytes, _ := reader.Peek(1)
		nextByte := nextBytes[0]

		switch nextByte {
		case 'i':
			// IntType
			decodedData = append(decodedData, getInt(reader))
		case 'l':
			// ListType
			decodedData = append(decodedData, getList(reader))
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// ByteType
			decodedData = append(decodedData, getBytes(reader))
		case 'd':
			decodedData = append(decodedData, getDict(reader))
		}

		if _, err := reader.Peek(1); err == io.EOF {
			break
		}
	}

	return decodedData

}
