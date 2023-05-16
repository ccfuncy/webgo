package orm

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"orm/log"
	"orm/session"
)

type Engine struct {
	db *sql.DB
}

func NewEngine(driver, source string) (*Engine, error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		log.Error(err)
		return nil, err
	}

	e := &Engine{db: db}
	log.Info("Connect database success")
	return e, nil
}

func (e *Engine) Close() {
	if err := e.db.Close(); err != nil {
		log.Error("Failed to close database")
		return
	}
	log.Info("close database success")
}
func (e *Engine) NewSession() *session.Session {
	return session.New(e.db)
}
