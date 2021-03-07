package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const HEAD_SIZE = 4

type Buffer struct {
	Proto
	conn       Connection
	reader     io.Reader
	header     string
	headLength int
	buf        []byte
	bufLength  int
	start      int
	end        int
}

func NewBuffer(header string, bufLength int) *Buffer {
	buf := make([]byte, bufLength)
	return &Buffer{
		header:     header,
		headLength: len(header) + HEAD_SIZE,
		buf:        buf,
		bufLength:  bufLength,
		start:      0,
		end:        0,
	}
}

func (buffer *Buffer) Read(conn Connection, f func(conn Connection, msg []byte)) error {
	for {
		// 移动数据
		if buffer.start > 0 {
			copy(buffer.buf, buffer.buf[buffer.start:buffer.end])
			buffer.end -= buffer.start
			buffer.start = 0
		}
		// 读数据拼接到定额缓存后面
		if buffer.end == buffer.bufLength {
			return errors.New(fmt.Sprintf("一个完整的数据包太长已经超过你定义的MAX_LENGTH(%d)\n", buffer.bufLength))
		}
		n, err := conn.ReadMsg(buffer.buf[buffer.end:])
		if err != nil {
			return err
		}
		buffer.end += n

		// 检查定额缓存里面的数据有几个消息(可能不到1个，可能连一个消息头都不够，可能有几个完整消息+一个消息的部分)
		for {
			if buffer.end-buffer.start < buffer.headLength {
				// 一个消息头都不够， 跳出去继续读吧, 但是这不是一种错误
				break

			}
			headBuf := buffer.buf[buffer.start : buffer.start+buffer.headLength]
			if string(headBuf[:len(buffer.header)]) == buffer.header { // 判断消息头正确性

			} else {
				return errors.New("消息头部不正确")
			}
			contentSize := int(binary.BigEndian.Uint32(headBuf[len(buffer.header):]))
			if buffer.end-buffer.start < contentSize-buffer.headLength { // 一个消息体都不够， 跳出去继续读吧, 但是这不是一种错误
				break
			}
			// 把消息读出来，把start往后移
			buffer.start += buffer.headLength
			contentBuf := buffer.buf[buffer.start : buffer.start+contentSize]
			buffer.start += contentSize
			f(conn, contentBuf)
		}
	}

}

func (buffer *Buffer) Pack(data []byte) ([]byte, error) {
	headSize := len(data)
	var headBytes = make([]byte, HEAD_SIZE)
	binary.BigEndian.PutUint32(headBytes, uint32(headSize))
	var bufferClient bytes.Buffer
	bufferClient.Write([]byte(buffer.header))
	bufferClient.Write(headBytes)
	bufferClient.Write(data)
	return bufferClient.Bytes(), nil
}

func (buffer *Buffer) UnPack(data []byte) ([]byte, error) {
	return data, nil
}
