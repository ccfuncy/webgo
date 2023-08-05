package parser

import (
	"bufio"
	"errors"
	"io"
	"redis/interface/resp"
	"redis/interface/utils"
	"redis/lib/logger"
	"redis/resp/reply"
	"runtime/debug"
	"strconv"
	"strings"
)

// 请求解析
type Payload struct {
	Data resp.Reply
	Err  error
}

// 解析器状态
type readState struct {
	readingMultiLine bool     //解析单行还是多行数据
	exceptArgsCount  int      //正在读取的数据应该有几个参数
	msgType          byte     //消息类型
	args             [][]byte //消息本身
	bulkLen          int64    //字节组或包的大小
}

func (s *readState) finish() bool {
	//返回解析是否结束
	return s.exceptArgsCount > 0 && s.exceptArgsCount == len(s.args)
}

func ParseStream(reader io.Reader) <-chan *Payload {
	//对外的解析接口
	ch := make(chan *Payload)
	go parse0(reader, ch)
	return ch
}

// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n=>*3\r\n $3\r\n
// 抽取
func readLine(reader *bufio.Reader, state *readState) ([]byte, bool, error) {
	// bool 用于标识是否是IO错误
	//1.\r\n切分
	var msg []byte
	var err error
	if state.bulkLen == 0 {
		//\r\n区分
		msg, err = reader.ReadBytes('\n')
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' {
			return nil, false, errors.New("protocol errors: " + string(msg))
		}
	} else {
		//2. 如果之前读到$符，严格读取字符个数
		msg = make([]byte, state.bulkLen+2)
		_, err := io.ReadFull(reader, msg)
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' || msg[len(msg)-1] != '\n' {
			return nil, false, err
		}
		state.bulkLen = 0
	}
	return msg, false, nil
}

// 解析多个字符串头
// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n=>*3\r\n $3\r\n
func parseMultiBulkHeader(msg []byte, state *readState) error {
	//*3\r\n
	exceptedLine, err := strconv.ParseUint(utils.BytesToString(msg[1:len(msg)-2]), 10, 32)
	if err != nil {
		return errors.New("protocol error: " + utils.BytesToString(msg))
	}
	if exceptedLine == 0 {
		state.exceptArgsCount = int(exceptedLine)
		return nil
	} else if exceptedLine > 0 {
		state.readingMultiLine = true
		state.msgType = msg[0]
		state.exceptArgsCount = int(exceptedLine)
		state.args = make([][]byte, 0, exceptedLine)
		return nil
	} else {
		return errors.New("protocol error: " + utils.BytesToString(msg))
	}
}

// 解析单个字符串头
// $3\r\nSET\r\n
func parseBulkHeader(msg []byte, state *readState) error {
	var err error
	state.bulkLen, err = strconv.ParseInt(utils.BytesToString(msg[1:len(msg)-2]), 10, 64)
	if err != nil {
		return errors.New("protocol error: " + utils.BytesToString(msg))
	}
	if state.bulkLen == -1 {
		return nil
	} else if state.bulkLen > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.exceptArgsCount = 1
		state.args = make([][]byte, 0, 1)
		return nil
	} else {
		return errors.New("protocol error: " + utils.BytesToString(msg))
	}
}

// 解析单条信息头
// +ok\r\n -err\r\n :5\r\n
func parseSingleLineReply(msg []byte) (resp.Reply, error) {
	str := strings.TrimSuffix(utils.BytesToString(msg), "\r\n")
	var res resp.Reply
	switch str[0] {
	case '+':
		res = reply.NewStatusReply(str[1:])
	case '-':
		res = reply.NewStandardErrReply(str[1:])
	case ':':
		val, err := strconv.ParseInt(str[1:], 10, 64)
		if err != nil {
			return nil, errors.New("protocol error: " + utils.BytesToString(msg))
		}
		res = reply.NewIntReply(val)
	}
	return res, nil
}

// SET\r\n
// $3\r\n SET\r\n $3\r\n key\r\n $5\r\n value\r\n
func readBody(msg []byte, state *readState) error {
	line := msg[:len(msg)-2]
	var err error
	if line[0] == '$' {
		state.bulkLen, err = strconv.ParseInt(utils.BytesToString(line[1:]), 10, 64)
		if err != nil {
			return errors.New("protocol error: " + utils.BytesToString(msg))
		}
		if state.bulkLen <= 0 {
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		}
	} else {
		state.args = append(state.args, line)
	}
	return err
}
func parse0(reader io.Reader, payload chan<- *Payload) {
	defer func() {
		if err := recover(); err != nil {
			logger.Default().Error(err)
			logger.Default().Error(debug.Stack())
		}
	}()
	bufReader := bufio.NewReader(reader)
	var state readState
	for true {
		msg, ioErr, err := readLine(bufReader, &state)
		if err != nil {
			if ioErr {
				payload <- &Payload{
					Err: err,
				}
				close(payload)
				return
			}
			payload <- &Payload{Err: err}
			state = readState{}
			continue
		}
		//判断是否多行解析
		if !state.readingMultiLine {
			if msg[0] == '*' {
				err := parseMultiBulkHeader(msg, &state)
				if err != nil {
					payload <- &Payload{Err: err}
					state = readState{}
					continue
				}
				if state.exceptArgsCount == 0 {
					payload <- &Payload{Data: reply.NewEmptyBulkReply()}
					state = readState{}
					continue
				}
			} else if msg[0] == '$' {
				err := parseBulkHeader(msg, &state)
				if err != nil {
					payload <- &Payload{Err: err}
					state = readState{}
					continue
				}
				if state.bulkLen == -1 {
					payload <- &Payload{Data: reply.NewNULLBulkReply()}
					state = readState{}
					continue
				}
			} else {
				reply, err := parseSingleLineReply(msg)
				payload <- &Payload{Data: reply, Err: err}
				state = readState{}
				continue
			}
		} else {
			err := readBody(msg, &state)
			if err != nil {
				payload <- &Payload{Err: err}
				state = readState{}
				continue
			}
			if state.finish() {
				var res resp.Reply
				if state.msgType == '*' {
					res = reply.NewMultiBulkReply(state.args)
				} else if state.msgType == '$' {
					reply.NewBulkReply(state.args[0])
				}
				payload <- &Payload{Err: err, Data: res}
				state = readState{}
			}
		}
	}
}
