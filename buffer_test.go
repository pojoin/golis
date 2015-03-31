package golis

import (
	"log"
	"testing"
)

func TestBuffer(T *testing.T) {
	buf := NewBuffer()
	buf.PutUint32(123)
	buf.PutUint32(456)
	buf.PutUint32(789)
	s := "你好，我在测试你3453453453jjjj"
	buf.PutString(s)
	log.Println(len(s))
	i, _ := buf.ReadUint32()
	log.Println("[buf123] = ", i)
	i, _ = buf.ReadUint32()
	log.Println("[buf456] = ", i)
	i, _ = buf.ReadUint32()
	log.Println("[buf789] = ", i)
	ss, err := buf.ReadString(len(s))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[buf string] = ", ss)
}
