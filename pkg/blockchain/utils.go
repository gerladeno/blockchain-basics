package blockchain

import (
	"bytes"
	"encoding/binary"
	"log"
	"strconv"
)

func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func intToHex(i int64) []byte {
	return []byte(strconv.FormatInt(i, 16))
}