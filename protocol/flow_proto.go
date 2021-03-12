package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	HEAD_SIZE = 4
	HEADER    = "socket-cluster"
	// The maximum length of each message (including headers), which can be set to 4G
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
