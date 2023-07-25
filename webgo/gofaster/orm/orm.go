package orm

import (
	"database/sql"
	"gofaster/log"
	"reflect"
	"time"
)

type FsDB struct {
	db              *sql.DB
	MaxOpenConn     int           //最大连接数
	MaxIdleConn     int           //最大空闲链接数
	ConnMaxLifetime time.Duration //链接存活时间
	ConnMaxIdletime time.Duration //空闲存活时间
	logger          *log.Logger
	Prefix          string
}

func Open(driverName, source string) *FsDB {
	open, err := sql.Open(driverName, source)
	if err != nil {
		panic(err)
	}
	fsDb := &FsDB{
		db:     open,
		logger: log.Default()}
	err = open.Ping()
	if err != nil {
		panic(err)
	}
	open.SetMaxIdleConns(5)
	open.SetMaxOpenConns(100)
	open.SetConnMaxLifetime(time.Minute * 3)
	open.SetConnMaxIdleTime(time.Minute * 1)
	return fsDb
}
func (d *FsDB) SetMaxIdleConns(n int) {
	d.db.SetMaxIdleConns(n)
}

func (d *FsDB) SetMaxOpenConns(n int) {
	d.db.SetMaxOpenConns(n)
}

func (d *FsDB) SetConnMaxLifetime(duration time.Duration) {
	d.db.SetConnMaxLifetime(duration)
}

func (d *FsDB) SetConnMaxIdleTime(duration time.Duration) {
	d.db.SetConnMaxIdleTime(duration)
}

func (d *FsDB) New(data any) *FsSession {
	typeOf := reflect.TypeOf(data)
	name := typeOf.Elem().Name()
	return &FsSession{
		db:        d,
		tableName: name,
	}
}
func (d *FsDB) Close() error {
	return d.db.Close()
}
