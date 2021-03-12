package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
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
	header     string
	headLength int
	bufLength  int
}

func NewFlowProto(header string, bufLength int) *FlowProto {
	return &FlowProto{
		header:     header,
		headLength: len(header) + HEAD_SIZE,
		bufLength:  bufLength,
	}
}

// Read Read and parse the received message
func (this *FlowProto) ReadMsg(conn FlowConnection, onMsg func(conn Connection, msg []byte)) error {
	var start, end int
	needBytesCount := this.headLength

	buf := make([]byte, this.bufLength)

	for {
		// 移动数据
		if start > 0 {
			copy(buf, buf[start:end])
			end -= start
			start = 0
		}
		// 读数据拼接到定额缓存后面
		if end == this.bufLength {
			return ExceededErr
		}
		n, err := conn.ReadMsg(buf[end:])
		if err != nil {
			return err
		}
		end += n
		// 检查定额缓存里面的数据有几个消息(可能不到1个，可能连一个消息头都不够，可能有几个完整消息+一个消息的部分)
		for end-start >= needBytesCount {
			headBuf := buf[start : start+this.headLength]
			if string(headBuf[:len(this.header)]) != this.header { // 判断消息头正确性
				return HeaderErr
			}
			contentSize := int(binary.BigEndian.Uint32(headBuf[len(this.header):]))
			totalLenth := this.headLength + contentSize
			if totalLenth > this.bufLength {
				return ExceededErr
			}
			if end-start < totalLenth { // 一个消息体都不够， 跳出去继续读吧, 但是这不是一种错误
				needBytesCount = totalLenth
				break
			}
			// 把消息读出来，把start往后移
			start += this.headLength
			contentBuf := buf[start : start+contentSize]
			start += contentSize
			onMsg(conn, contentBuf)
			needBytesCount = this.headLength
		}
	}

}

// Read Read and parse the received message
func (this *FlowProto) Read(conn FlowConnection, onMsg func(conn Connection, msg []byte)) error {
	var start, end int
	buf := make([]byte, this.bufLength)
	for {
		// read header
		for end-start < this.headLength {
			// reset start
			if start > 0 {
				copy(buf, buf[start:end])
				end -= start
				start = 0
			}
			n, err := conn.ReadMsg(buf[end:])
			if err != nil {
				return err
			}
			end += n
			// The maximum length of the packet is exceeded
			if end == this.bufLength {
				return ExceededErr
			}
		}

		headBuf := buf[start : start+this.headLength]
		// Verify that the message header is correct
		if string(headBuf[:len(this.header)]) != this.header {
			return HeaderErr
		}

		// read content
		contentSize := int(binary.BigEndian.Uint32(headBuf[len(this.header):]))
		totalLenth := this.headLength + contentSize
		if totalLenth > this.bufLength {
			return ExceededErr
		}
		for end-start < totalLenth {
			n, err := conn.ReadMsg(buf[end:])
			if err != nil {
				return err
			}
			end += n
			// The maximum length of the packet is exceeded
			if end == this.bufLength {
				return ExceededErr
			}
		}
		start += this.headLength
		contentBuf := buf[start : start+contentSize]
		start += contentSize
		onMsg(conn, contentBuf)
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
