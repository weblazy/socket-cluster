package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

const HEAD_SIZE = 4

type Buffer struct {
	reader     io.Reader
	header     string
	buf        []byte
	buf_length int
	start      int
	end        int
}

func NewBuffer(reader io.Reader, header string, len int) *Buffer {
	buf := make([]byte, len)
	return &Buffer{reader, header, buf, len, 0, 0}
}

// grow 将有用的字节前移
func (b *Buffer) grow() {
	if b.start == 0 {
		return
	}
	copy(b.buf, b.buf[b.start:b.end])
	b.end -= b.start
	b.start = 0
}

func (b *Buffer) len() int {
	return b.end - b.start
}

//返回n个字节，而不产生移位
func (b *Buffer) seek(n int) ([]byte, error) {
	if b.end-b.start >= n {
		buf := b.buf[b.start : b.start+n]
		return buf, nil
	}
	return nil, errors.New("not enough")
}

//舍弃offset个字段，读取n个字段
func (b *Buffer) read(offset, n int) []byte {
	b.start += offset
	buf := b.buf[b.start : b.start+n]
	b.start += n
	return buf
}

//从reader里面读取数据，如果reader阻塞，会发生阻塞
func (b *Buffer) readFromReader() error {
	if b.end == b.buf_length {
		return errors.New(fmt.Sprintf("一个完整的数据包太长已经超过你定义的example.BUFFER_LENGTH(%d)\n", b.buf_length))
	}
	n, err := b.reader.Read(b.buf[b.end:])
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second) // 便于观察这里sleep了一下
	b.end += n
	return nil
}

func (buffer *Buffer) Read(msg chan string) error {
	for {
		buffer.grow()                   // 移动数据
		err1 := buffer.readFromReader() // 读数据拼接到定额缓存后面
		if err1 != nil {
			return err1
		}
		// 检查定额缓存里面的数据有几个消息(可能不到1个，可能连一个消息头都不够，可能有几个完整消息+一个消息的部分)
		err2 := buffer.checkMsg(msg)
		if err2 != nil {
			return err2
		}
	}
}

func (buffer *Buffer) checkMsg(msg chan string) error {
	HEADER_LENG := HEAD_SIZE + len(buffer.header)
	headBuf, err1 := buffer.seek(HEADER_LENG)
	if err1 != nil { // 一个消息头都不够， 跳出去继续读吧, 但是这不是一种错误
		return nil
	}
	if string(headBuf[:len(buffer.header)]) == buffer.header { // 判断消息头正确性

	} else {
		return errors.New("消息头部不正确")
	}
	contentSize := int(binary.BigEndian.Uint32(headBuf[len(buffer.header):]))
	if buffer.len() >= contentSize-HEADER_LENG { // 一个消息体也是够的
		contentBuf := buffer.read(HEADER_LENG, contentSize) // 把消息读出来，把start往后移
		msg <- string(contentBuf)
		// 递归，看剩下的还够一个消息不
		err3 := buffer.checkMsg(msg)
		if err3 != nil {
			return err3
		}
	}
	return nil
}
