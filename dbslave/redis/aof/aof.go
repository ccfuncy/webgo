package aof

import (
	"io"
	"os"
	"redis/interface/database"
	"redis/interface/utils"
	"redis/lib/config"
	"redis/lib/logger"
	"redis/resp/connection"
	"redis/resp/parser"
	"redis/resp/reply"
	"strconv"
)

const aofBufferSize = 1 << 16

type payload struct {
	cmd     database.Cmdline
	dbIndex int
}

type AofHandler struct {
	database    database.Database
	aofFile     *os.File
	aofFileName string
	currentDB   int
	aofChan     chan *payload
}

func NewAofHandler(database database.Database) (*AofHandler, error) {
	handler := &AofHandler{database: database}
	handler.aofFileName = config.Conf.Redis["appendonlyfilename"].(string)
	//loadAof
	handler.loadAof()
	file, err := os.OpenFile(handler.aofFileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	handler.aofFile = file
	handler.aofChan = make(chan *payload, aofBufferSize)
	go func() {
		handler.handleAof()
	}()
	return handler, nil
}

func (h *AofHandler) AddAof(dbIndex int, line database.Cmdline) {
	if !(config.Conf.Redis["appendonly"].(bool)) && h.aofChan == nil {
		return
	}
	h.aofChan <- &payload{
		cmd:     line,
		dbIndex: dbIndex,
	}
}

func (h *AofHandler) handleAof() {
	//落盘
	h.currentDB = 0
	for p := range h.aofChan {
		if p.dbIndex != h.currentDB {
			bytes := reply.NewMultiBulkReply(utils.ToCmdLine("select", strconv.Itoa(p.dbIndex))).ToBytes()
			_, err := h.aofFile.Write(bytes)
			if err != nil {
				logger.Default().Error(err)
				continue
			}
			h.currentDB = p.dbIndex
		}
		bytes := reply.NewMultiBulkReply(p.cmd).ToBytes()
		_, err := h.aofFile.Write(bytes)
		if err != nil {
			logger.Default().Error(err)
		}
	}
}

func (h *AofHandler) loadAof() {
	open, err := os.Open(h.aofFileName)
	if err != nil {
		return
	}
	defer open.Close()
	ch := parser.ParseStream(open)
	for p := range ch {
		if p.Err != nil {
			if p.Err == io.EOF {
				break
			}
			logger.Default().Error(err)
			continue
		}
		if p.Data == nil {
			logger.Default().Info("empty data")
		}
		bulkReply, ok := p.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Default().Error("exec err")
			continue
		}
		c := connection.NewFakeConnection()
		resp := h.database.Exec(c, bulkReply.Args)
		if reply.IsErrReply(resp) {
			logger.Default().Error(resp)
		}
	}
}
