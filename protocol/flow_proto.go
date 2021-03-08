package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

const (
	HEAD_SIZE = 4
	HEADER    = "socket-cluster"
	// 每个消息(包括头部)的最大长度， 这里最大可以设置4G
	MAX_LENGTH = 1024 * 70
)

var ExceededErr = errors.New("The maximum length of the packet is exceeded")
var HeaderErr = errors.New("The message header is incorrect")

var DefaultFlowProto = NewFlowProto(HEADER, MAX_LENGTH)

type FlowProto struct {
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

func NewFlowProto(header string, bufLength int) *FlowProto {
	buf := make([]byte, bufLength)
	return &FlowProto{
		header:     header,
		headLength: len(header) + HEAD_SIZE,
		buf:        buf,
		bufLength:  bufLength,
		start:      0,
		end:        0,
	}
}

// Read Read and parse the received message
func (this *FlowProto) Read(conn Connection, onMsg func(conn Connection, msg []byte)) error {
	for {
		// 移动数据
		if this.start > 0 {
			copy(this.buf, this.buf[this.start:this.end])
			this.end -= this.start
			this.start = 0
		}
		// 读数据拼接到定额缓存后面
		if this.end == this.bufLength {
			return ExceededErr
		}
		n, err := conn.ReadMsg(this.buf[this.end:])
		if err != nil {
			return err
		}
		this.end += n
		// 检查定额缓存里面的数据有几个消息(可能不到1个，可能连一个消息头都不够，可能有几个完整消息+一个消息的部分)
		for {
			if this.end-this.start < this.headLength {
				// 一个消息头都不够， 跳出去继续读吧, 但是这不是一种错误
				break
			}
			headBuf := this.buf[this.start : this.start+this.headLength]
			if string(headBuf[:len(this.header)]) != this.header { // 判断消息头正确性
				return HeaderErr
			}
			contentSize := int(binary.BigEndian.Uint32(headBuf[len(this.header):]))
			if this.end-this.start < contentSize-this.headLength { // 一个消息体都不够， 跳出去继续读吧, 但是这不是一种错误
				break
			}
			// 把消息读出来，把start往后移
			this.start += this.headLength
			contentBuf := this.buf[this.start : this.start+contentSize]
			this.start += contentSize
			onMsg(conn, contentBuf)
		}
	}

}

// Pack Package raw data
func (this *FlowProto) Pack(data []byte) ([]byte, error) {
	headSize := len(data)
	if headSize+this.headLength > this.bufLength {
		return nil, ExceededErr
	}
	var headBytes = make([]byte, HEAD_SIZE)
	binary.BigEndian.PutUint32(headBytes, uint32(headSize))
	var buf bytes.Buffer
	buf.Write([]byte(this.header))
	buf.Write(headBytes)
	buf.Write(data)
	return buf.Bytes(), nil
}
