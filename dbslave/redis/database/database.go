package database

import (
	"fmt"
	"redis/aof"
	"redis/interface/database"
	"redis/interface/resp"
	"redis/interface/utils"
	"redis/lib/config"
	"redis/lib/logger"
	"redis/resp/reply"
	"strconv"
	"strings"
)

type DataBase struct {
	dbSet      []*DB
	aofHandler *aof.AofHandler
}

func NewDataBase() *DataBase {
	dbNum := config.Conf.Redis["databases"].(int64)
	if dbNum == 0 {
		dbNum = 16
	}
	logger.Default().Info(fmt.Sprintf("init redis for %d db", dbNum))
	dbs := make([]*DB, dbNum)
	for i := 0; i < int(dbNum); i++ {
		dbs[i] = NewDB()
		dbs[i].index = i
	}
	d := &DataBase{dbSet: dbs}
	if config.Conf.Redis["appendonly"].(bool) {
		handler, err := aof.NewAofHandler(d)
		if err != nil {
			logger.Default().Error(err)
			panic(err)
		}
		d.aofHandler = handler
		for _, db := range dbs {
			var i = db.index
			db.AddAof = func(cmdline database.Cmdline) {
				handler.AddAof(i, cmdline)
			}
		}
	}
	return d
}

func (d *DataBase) Exec(client resp.Connection, args database.Cmdline) resp.Reply {
	defer func() {
		if err := recover(); err != nil {
			logger.Default().Error(err)
		}
	}()
	cmdName := strings.ToUpper(utils.BytesToString(args[0]))
	if cmdName == "SELECT" {
		if len(args) != 2 {
			return reply.NewArgNumErrReply(cmdName)
		}
		return execSelect(client, d, args[1:])
	}
	dbIndex := client.GetDBIndex()
	logger.Default().Info(fmt.Sprintf("client %s select db %d", client.RemoteAddr().String(), dbIndex))
	db := d.dbSet[dbIndex]
	return db.Exec(client, args)
}

func (d *DataBase) Close() {
}

func (d *DataBase) AfterClose(client resp.Connection) {
}

// select dbnum db相关命令交由executor 处理，select 与DB无关
func execSelect(connection resp.Connection, database *DataBase, args database.Cmdline) resp.Reply {
	dbIndex, err := strconv.Atoi(utils.BytesToString(args[0]))
	logger.Default().Info(fmt.Sprintf("client %s select db %d", connection.RemoteAddr().String(), dbIndex))
	if err != nil {
		return reply.NewStandardErrReply("Err invalid DB index ")
	}
	if dbIndex >= len(database.dbSet) {
		return reply.NewStandardErrReply("Err DB index is out of range")
	}
	connection.SelectDB(dbIndex)
	return reply.NewOKReply()
}
