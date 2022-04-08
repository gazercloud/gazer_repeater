package storage

import (
	"github.com/gazercloud/gazer_repeater/credentials"
	"github.com/gazercloud/gazer_repeater/logger"
	"github.com/jackc/pgx"
	"sync"
)

type Storage struct {
	db       *pgx.ConnPool
	address  string
	database string
	user     string
	password string

	mtx           sync.Mutex
	sessionsCache map[string]*Session
}

func NewStorage() *Storage {
	var c Storage
	/*c.address = "gazer-cluster-1-instance-1.ceovd2dqka5z.eu-central-1.rds.amazonaws.com"
	c.user = "gazer"
	c.password = "PtaMn6csVdmQw1zp#"*/

	c.address = credentials.DbHost
	c.user = credentials.DbUser
	c.password = credentials.DbPassword

	c.sessionsCache = make(map[string]*Session)
	return &c
}

func (c *Storage) checkConnection() error {
	var err error

	if c.db == nil {
		logger.Println("[Storage]", "checkConnection begin (c.db == nil)", c.address)
		c.db, err = pgx.NewConnPool(pgx.ConnPoolConfig{
			ConnConfig: pgx.ConnConfig{
				Host:                 c.address,
				Port:                 5432,
				Database:             "gazercloud",
				User:                 c.user,
				Password:             c.password,
				TLSConfig:            nil,
				UseFallbackTLS:       false,
				FallbackTLSConfig:    nil,
				Logger:               nil,
				LogLevel:             0,
				Dial:                 nil,
				RuntimeParams:        nil,
				OnNotice:             nil,
				CustomConnInfo:       nil,
				CustomCancel:         nil,
				PreferSimpleProtocol: false,
				TargetSessionAttrs:   "",
			},
			MaxConnections: 0,
			AfterConnect:   nil,
			AcquireTimeout: 0,
		})

		if err != nil {
			logger.Println("[Storage]", "checkConnection error:", err)
			c.db = nil
			return err
		}
	}

	return nil
}

func (c *Storage) Directory() (res map[string]string, err error) {
	res = make(map[string]string)
	logger.Println("[Storage]", "Directory begin")
	err = c.checkConnection()
	if err != nil {
		logger.Println("[Storage]", "Directory error (checkConnection):", err)
		return
	}

	var rows *pgx.Rows
	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		logger.Println("[Storage]", "Directory error (BEGIN):", err)
		return nil, err
	}

	rows, err = c.db.Query("SELECT code, value FROM directory")
	if err != nil {
		_ = tr.Rollback()
		logger.Println("[Storage]", "Directory error (SELECT):", err)
		return nil, err
	}
	for rows.Next() {
		var code string
		var value string
		err = rows.Scan(&code, &value)
		if err != nil {
			break
		}
		res[code] = value
	}
	rows.Close()
	_ = tr.Rollback()
	logger.Println("[Storage]", "Directory end")
	return
}

func (c *Storage) ServerRole(code string) (role string, err error) {
	logger.Println("[Storage]", "ServerRole begin")
	err = c.checkConnection()
	if err != nil {
		logger.Println("[Storage]", "ServerRole error (checkConnection):", err)
		return
	}

	var rows *pgx.Rows
	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		logger.Println("[Storage]", "ServerRole error (BEGIN):", err)
		return "", err
	}

	rows, err = c.db.Query("SELECT role FROM servers WHERE code=$1", code)
	if err != nil {
		_ = tr.Rollback()
		logger.Println("[Storage]", "ServerRole error (SELECT):", err)
		return "", err
	}
	if rows.Next() {
		err = rows.Scan(&role)
		if err != nil {
			logger.Println("[Storage]", "ServerRole error (Scan):", err)
			return "", err
		}
		logger.Println("[Storage]", "Role for", code, "is", role)
	}
	rows.Close()
	_ = tr.Rollback()
	logger.Println("[Storage]", "ServerRole end")
	return
}

type ServerInfo struct {
	Id      int64
	Name    string
	Code    string
	Role    string
	Config  string
	Enabled bool
}

func (c *Storage) Servers() (res []ServerInfo, err error) {
	res = make([]ServerInfo, 0)
	logger.Println("[Storage]", "Servers begin")
	err = c.checkConnection()
	if err != nil {
		logger.Println("[Storage]", "Servers error (checkConnection):", err)
		return
	}

	var rows *pgx.Rows
	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		logger.Println("[Storage]", "Servers error (BEGIN):", err)
		return nil, err
	}

	rows, err = c.db.Query("SELECT id, name, code, role, config, enabled FROM servers ORDER BY id")
	if err != nil {
		_ = tr.Rollback()
		logger.Println("[Storage]", "Servers error (SELECT):", err)
		return nil, err
	}
	for rows.Next() {
		var serverInfo ServerInfo
		err = rows.Scan(&serverInfo.Id, &serverInfo.Name, &serverInfo.Code, &serverInfo.Role, &serverInfo.Config, &serverInfo.Enabled)
		if err != nil {
			break
		}
		res = append(res, serverInfo)
	}
	rows.Close()
	_ = tr.Rollback()
	logger.Println("[Storage]", "Servers end")
	return
}
