package main

import (
	"bytes"
	"fmt"

	"github.com/pojoin/golis"
)

type TextLineCodecFilter struct {
	golis.IoFilterAdapter
}

func (*TextLineCodecFilter) Decode(message interface{}) (interface{}, bool) {
	enter := []byte("\n")
	if buffer, ok := message.(*golis.Buffer); ok {
		readpos := buffer.GetReadPos()
		fmt.Printf("readpos = %d,writepos = %d\n", readpos, buffer.GetWritePos())
		data, err := buffer.ReadBytes(buffer.GetWritePos() - readpos)
		if err != nil {
			return message, false
		}
		fmt.Println("data.len = ", len(data))
		if pos := bytes.Index(data, enter); pos != -1 {
			if readpos+pos+len(enter) == buffer.GetWritePos() {
				buffer.ResetRead()
				buffer.ResetWrite()
			} else {
				buffer.SetReadPos(readpos + pos + len(enter))
			}
			return data[0:pos], true
		}
	}
	return message, false
}

func (*TextLineCodecFilter) Encode(message interface{}) (interface{}, bool) {
	enter := []byte("\n")
	if msg, ok := message.([]byte); ok {
		msg = append(msg, enter...)
		return msg, true
	}
	return message, false
}
