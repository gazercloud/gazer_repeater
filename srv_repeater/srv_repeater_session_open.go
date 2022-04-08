package srv_repeater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/gazercloud/gazer_repeater/storage"
)

func (c *SrvRepeater) SessionOpen(requestText []byte) ([]byte, error) {
	var err error
	type SessionOpenRequest struct {
		UserName string `json:"user_name"`
		Password string `json:"password"`
	}
	var req SessionOpenRequest
	err = json.Unmarshal(requestText, &req)
	if err != nil {
		return nil, err
	}

	hash := sha256.New()
	hash.Write([]byte(req.Password))
	shaPassword := hex.EncodeToString(hash.Sum(nil))

	var sessionKey string
	sessionKey, err = c.storage.SessionOpen(req.UserName, shaPassword)
	if err != nil {
		return nil, err
	}

	type SessionOpenResponse struct {
		SessionToken string `json:"session_token"`
	}
	var bs []byte
	var resp SessionOpenResponse
	resp.SessionToken = sessionKey
	bs, err = json.MarshalIndent(resp, "", " ")
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (c *SrvRepeater) SessionActivate(requestText []byte) ([]byte, error) {
	var err error
	type SessionActivateRequest struct {
		SessionToken string `json:"session_token"`
	}
	var req SessionActivateRequest
	err = json.Unmarshal(requestText, &req)
	if err != nil {
		return nil, err
	}

	type SessionOpenResponse struct {
		SessionToken string `json:"session_token"`
	}
	var bs []byte
	var resp SessionOpenResponse
	resp.SessionToken = req.SessionToken
	bs, err = json.MarshalIndent(resp, "", " ")
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (c *SrvRepeater) SessionClose(requestText []byte, session *storage.Session) (err error) {
	if session == nil {
		return errors.New("no session")
	}

	type SessionCloseRequest struct {
		Key string `json:"key"`
	}
	var req SessionCloseRequest
	err = json.Unmarshal(requestText, &req)
	if err != nil {
		return
	}

	if req.Key == "" {
		req.Key = session.Key
	}

	err = c.storage.SessionClose(req.Key, session.UserId)
	return
}
