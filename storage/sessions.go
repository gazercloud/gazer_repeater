package storage

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx"
	"http-server.org/gazer/credentials"
	"http-server.org/gazer/logger"
	"math/rand"
	"strconv"
	"time"
)

type Session struct {
	Id             int64
	UserId         int64
	Key            string
	LastUpdateTime time.Time
}

type RegRequest struct {
	Id       int64
	EMail    string
	DT       int64
	Password string
	Key      string
}

func (c *Storage) Log(tp string, content string) {
	var err error
	err = c.checkConnection()
	if err != nil {
		return
	}

	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		return
	}

	_, err = tr.Exec("INSERT INTO logs (time, src, tp, content) VALUES($1, $2, $3, $4)", time.Now().UTC(), credentials.ServerId, tp, content)
	if err != nil {
		_ = tr.Rollback()
		return
	}

	err = tr.Commit()
	if err != nil {
		_ = tr.Rollback()
		return
	}
}

func (c *Storage) SessionOpen(userName string, password string) (string, error) {
	var err error
	err = c.checkConnection()
	if err != nil {
		return "", err
	}

	var res *pgx.Rows
	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		return "", err
	}

	res, err = tr.Query("SELECT id FROM users WHERE name=$1 AND password=$2", userName, password)
	if err != nil {
		_ = tr.Rollback()
		return "", err
	}

	var userId int64
	var userFound bool
	var values []interface{}
	if res.Next() {
		values, err = res.Values()
		if err != nil {
			_ = tr.Rollback()
			return "", err
		}
		userId = values[0].(int64)
		userFound = true
	}
	res.Close()

	if !userFound {
		_ = tr.Rollback()
		return "", errors.New("we didn't recognize the email or password you entered")
	}

	sessionKey := "S-" + strconv.FormatInt(userId, 16) + "-" + strconv.FormatInt(rand.Int63(), 16) + "-" + strconv.FormatInt(time.Now().UnixNano(), 16)

	_, err = tr.Exec("INSERT INTO sessions (id, user_id, key) VALUES(nextval('seq_session_id'), $1, $2)", userId, sessionKey)
	if err != nil {
		_ = tr.Rollback()
		return "", err
	}

	_, err = tr.Exec("INSERT INTO changes (tp, time, content) VALUES($1, $2, $3)", "session_open", time.Now().UTC().UnixNano(), sessionKey)
	if err != nil {
		_ = tr.Rollback()
		return "", err
	}

	err = tr.Commit()
	if err != nil {
		_ = tr.Rollback()
		return "", err
	}

	return sessionKey, nil
}

func (c *Storage) SessionClose(key string, userId int64) (err error) {

	err = c.checkConnection()
	if err != nil {
		return
	}

	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		return
	}
	_, err = tr.Exec("DELETE FROM sessions WHERE user_id=$1 AND key=$2", userId, key)
	if err != nil {
		_ = tr.Rollback()
		return
	}

	_, err = tr.Exec("INSERT INTO changes (tp, time, content) VALUES($1, $2, $3)", "session_close", time.Now().UTC().UnixNano(), key)
	if err != nil {
		_ = tr.Rollback()
		return
	}

	err = tr.Commit()

	return
}

func (c *Storage) Registration(eMail string, password string, addr string, recaptcha string, score float64) (string, error) {
	var err error
	err = c.checkConnection()
	if err != nil {
		return "", err
	}

	var res *pgx.Rows
	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		return "", err
	}

	{
		res, err = tr.Query("SELECT id FROM users WHERE email=$1", eMail)
		if err != nil {
			_ = tr.Rollback()
			return "", err
		}

		userFound := false
		if res.Next() {
			userFound = true
		}
		res.Close()

		if userFound {
			_ = tr.Rollback()
			return "", errors.New("user already exists")
		}
	}

	regRequestKey := hex.EncodeToString([]byte(eMail)) + "_" + fmt.Sprint(time.Now().UnixNano()%10000000)

	_, err = tr.Exec("INSERT INTO reg_requests (id, email, dt, password, key, addr, comment, score) VALUES(nextval('seq_reg_request_id'), $1, $2, $3, $4, $5, $6, $7)", eMail, time.Now().Unix(), password, regRequestKey, addr, recaptcha, score)
	if err != nil {
		_ = tr.Rollback()
		return "", err
	}

	err = tr.Commit()
	if err != nil {
		_ = tr.Rollback()
		return "", err
	}

	return regRequestKey, nil
}

func (c *Storage) ChangePassword(session *Session, password string) (err error) {
	if session == nil {
		return errors.New("no session")
	}

	err = c.checkConnection()
	if err != nil {
		return
	}

	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		return
	}

	_, err = tr.Exec("UPDATE users SET password=$1 WHERE id=$2", password, session.UserId)
	if err != nil {
		_ = tr.Rollback()
		return
	}
	_ = tr.Commit()

	return
}

func (c *Storage) ChangePasswordByKey(key string, password string) (err error) {
	if key == "" {
		return errors.New("no key")
	}

	var keyBS []byte
	keyBS, err = hex.DecodeString(key)

	type ResetPasswordStruct struct {
		EMail    string `json:"e_mail"`
		Password string `json:"password"`
	}
	var keyObject ResetPasswordStruct
	err = json.Unmarshal(keyBS, &keyObject)
	if err != nil {
		return
	}

	err = c.checkConnection()
	if err != nil {
		return
	}

	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		return
	}

	//err = errors.New("ChangePasswordByKey/" +  password + "/" +  keyObject.EMail + "/" +  keyObject.Password + "/")
	//return

	_, err = tr.Exec("UPDATE users SET password=$1 WHERE email=$2 AND password=$3", password, keyObject.EMail, keyObject.Password)
	if err != nil {
		_ = tr.Rollback()
		return
	}
	_ = tr.Commit()

	return
}

func (c *Storage) ConfirmRegistration(key string) (email string, err error) {
	err = c.checkConnection()
	if err != nil {
		return
	}

	var res *pgx.Rows
	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		return
	}

	var regRequest RegRequest

	{
		res, err = tr.Query("SELECT id, email, password FROM reg_requests WHERE key=$1", key)
		if err != nil {
			_ = tr.Rollback()
			return
		}

		recordFound := false
		if res.Next() {
			err = res.Scan(&regRequest.Id, &regRequest.EMail, &regRequest.Password)
			recordFound = true
			email = regRequest.EMail
		}
		res.Close()

		if err != nil {
			_ = tr.Rollback()
			return
		}

		if !recordFound {
			_ = tr.Rollback()
			err = errors.New("no reg request found")
			return
		}
	}

	{
		res, err = tr.Query("SELECT id FROM users WHERE email=$1", regRequest.EMail)
		if err != nil {
			_ = tr.Rollback()
			return
		}

		recordFound := false
		if res.Next() {
			recordFound = true
		}
		res.Close()

		if recordFound {
			_ = tr.Rollback()
			err = errors.New("user exists already")
			return
		}
	}

	_, err = tr.Exec("INSERT INTO users (id, name, email, password) VALUES(nextval('seq_user_id'), $1, $2, $3)", regRequest.EMail, regRequest.EMail, regRequest.Password)
	if err != nil {
		_ = tr.Rollback()
		return
	}

	err = tr.Commit()
	if err != nil {
		_ = tr.Rollback()
		return
	}

	return
}

func (c *Storage) SessionCheck(sessionKey string) (*Session, error) {
	var err error
	var ok bool
	var session *Session

	if sessionKey == "" {
		return nil, errors.New("empty session key")
	}

	c.mtx.Lock()
	if session, ok = c.sessionsCache[sessionKey]; ok {
		if time.Now().Sub(session.LastUpdateTime) > 60*time.Second {
			delete(c.sessionsCache, sessionKey)
			session = nil
		}
	}
	c.mtx.Unlock()

	if session != nil {
		return session, nil
	}

	err = c.checkConnection()
	if err != nil {
		return nil, err
	}

	var res *pgx.Rows
	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		return nil, err
	}

	res, err = tr.Query("SELECT id, user_id FROM sessions WHERE key=$1", sessionKey)
	if err != nil {
		_ = tr.Rollback()
		return nil, err
	}

	if res.Next() {
		session = &Session{}
		err = res.Scan(&session.Id, &session.UserId)
		if err != nil {
			res.Close()
			_ = tr.Rollback()
			return nil, err
		}
		session.Key = sessionKey
	}
	res.Close()

	if session == nil {
		_ = tr.Rollback()
		return nil, errors.New("no session found")
	}

	c.mtx.Lock()
	session.LastUpdateTime = time.Now()
	c.sessionsCache[sessionKey] = session
	//logger.Println("session", sessionKey, "updated", time.Now())
	c.mtx.Unlock()

	_ = tr.Rollback()

	return session, nil
}

func (c *Storage) GetPassword(email string) (password string, err error) {
	err = c.checkConnection()
	if err != nil {
		return
	}

	var res *pgx.Rows
	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		return
	}

	res, err = tr.Query("SELECT password FROM users WHERE email=$1", email)
	if err != nil {
		_ = tr.Rollback()
		return
	}

	if res.Next() {
		err = res.Scan(&password)
		if err != nil {
			res.Close()
			_ = tr.Rollback()
			return
		}
	} else {
		err = errors.New("no user found")
	}
	res.Close()
	_ = tr.Rollback()

	return
}

type AccountInfo struct {
	Id            int64  `json:"id"`
	Email         string `json:"email"`
	MaxNodesCount int64  `json:"max_nodes_count"`
	NodesCount    int64  `json:"nodes_count"`
}

func (c *Storage) GetAccountInfo(userId int64) (accountInfo AccountInfo, err error) {
	err = c.checkConnection()
	if err != nil {
		return
	}

	var res *pgx.Rows
	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		return
	}

	{
		res, err = tr.Query("SELECT email, (max_nodes_count + free_nodes) as max_nodes_count  FROM users WHERE id=$1", userId)
		if err != nil {
			_ = tr.Rollback()
			return
		}

		if res.Next() {
			err = res.Scan(&accountInfo.Email, &accountInfo.MaxNodesCount)
			if err != nil {
				res.Close()
				_ = tr.Rollback()
				return
			}
		} else {
			err = errors.New("no user found")
		}
		res.Close()
	}

	{
		res, err = tr.Query("SELECT count(*) FROM nodes WHERE user_id=$1", userId)
		if err != nil {
			_ = tr.Rollback()
			return
		}

		if res.Next() {
			err = res.Scan(&accountInfo.NodesCount)
			if err != nil {
				res.Close()
				_ = tr.Rollback()
				return
			}
		}
		res.Close()
	}

	_ = tr.Rollback()

	return
}

func (c *Storage) GetAccountInfoByEMail(email string) (accountInfo AccountInfo, err error) {
	err = c.checkConnection()
	if err != nil {
		return
	}

	var res *pgx.Rows
	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		return
	}

	res, err = tr.Query("SELECT id, email, (max_nodes_count + free_nodes) as max_nodes_count FROM users WHERE email=$1", email)
	if err != nil {
		_ = tr.Rollback()
		return
	}

	if res.Next() {
		err = res.Scan(&accountInfo.Id, &accountInfo.Email, &accountInfo.MaxNodesCount)
		if err != nil {
			res.Close()
			_ = tr.Rollback()
			return
		}
	} else {
		err = errors.New("no user found")
	}

	res.Close()
	_ = tr.Rollback()

	return
}

func (c *Storage) ResetPasswordByIdAndKey(email string, oldPassword string) (string, error) {
	var err error

	err = c.checkConnection()
	if err != nil {
		return "", err
	}

	newPassword := time.Now().String()

	var tr *pgx.Tx
	tr, err = c.db.Begin()
	if err != nil {
		return "", err
	}

	_, err = tr.Exec("UPDATE users SET password=$1 WHERE user_email=$2 AND user_password=$3", newPassword, email, oldPassword)
	if err != nil {
		_ = tr.Rollback()
		return "", err
	}
	_ = tr.Rollback()

	return newPassword, nil
}

func (c *Storage) CheckNode(userId int64, nodeIdAsString string) (nodeName string, err error) {
	logger.Println("Storage CheckNode", nodeIdAsString, nodeName)

	var nodeId int64
	nodeId, err = strconv.ParseInt(nodeIdAsString, 10, 64)
	if err != nil {
		return
	}

	err = c.checkConnection()
	if err != nil {
		return
	}

	{
		var res *pgx.Rows
		res, err = c.db.Query("SELECT name FROM nodes WHERE id=$1 AND user_id=$2", nodeId, userId)
		if err != nil {
			return
		}
		if res.Next() {
			err = res.Scan(&nodeName)
			logger.Println("Storage CheckNode node found", nodeId, nodeName)
			if err != nil {
				res.Close()
				return
			}
		} else {
			err = errors.New("no node [" + nodeIdAsString + "] found for user")
		}
		res.Close()
	}

	return
}

/*func (c *Storage) WhereNode(nodeCode string, userId int64) (host string, err error) {
	err = c.checkConnection()
	if err != nil {
		return
	}

	logger.Println("Storage WhereNode", userId, nodeCode)

	// Search existing node
	{
		var res *pgx.Rows
		res, err = c.db.Query("SELECT active_host FROM nodes WHERE code=$1 AND user_id=$2", nodeCode, userId)
		if err != nil {
			return
		}
		if res.Next() {
			err = res.Scan(&host)
			logger.Println("Storage WhereNode node found", host)
			if err != nil {
				res.Close()
				return
			}
		} else {
			err = errors.New("no node found")
		}
		res.Close()
	}
	return
}
*/
