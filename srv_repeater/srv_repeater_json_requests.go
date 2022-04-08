package srv_repeater

import (
	"encoding/json"
	"errors"
	"github.com/gazercloud/gazer_repeater/logger"
	"github.com/gazercloud/gazer_repeater/storage"
)

func (c *SrvRepeater) RequestJson(function string, requestText []byte, session *storage.Session, nodeId string, addr string) ([]byte, error) {
	c.mtx.Lock()
	c.counterRequests++
	c.mtx.Unlock()

	var err error
	var result []byte
	switch function {
	case "s-registered-nodes":
		result, err = c.Nodes(session)
	case "s-active-nodes":
		result, err = c.AllNodes(requestText)
	case "s-where-node":
		result, err = c.Where(requestText)
	case "s-repeater-for-node":
		result, err = c.RepeaterForNode(requestText)
	case "s-get-all-nodes":
		result, err = c.GetAllNodesOfCluster()
	case "s-get-reg-db":
		result, err = c.GetAllDB()
	case "s-account-info":
		result, err = c.AccountInfo(session, requestText)
	case "s-registration":
		result, err = c.Registration(requestText, addr)
	case "s-confirm-registration":
		result, err = c.ConfirmRegistration(requestText, addr)
	case "s-change-password":
		result, err = c.ChangePassword(session, requestText, addr)
	case "s-restore-password":
		result, err = c.RestorePassword(requestText, addr)
	case "s-reset-password":
		result, err = c.ResetPassword(session, requestText, addr)
	case "s-node-add":
		result, err = c.NodeAdd(session, requestText, addr)
	case "s-node-update":
		result, err = c.NodeUpdate(session, requestText, addr)
	case "s-node-remove":
		result, err = c.NodeRemove(session, requestText, addr)
	case "session_open":
		result, err = c.SessionOpen(requestText)
	case "session_activate":
		result, err = c.SessionActivate(requestText)
	case "session_close":
		err = c.SessionClose(requestText, session)
	case "s-state":
		result, err = c.StateAsString()
	case "s-buy":
		result, err = c.Buy(requestText, session, addr)
	case "s-stop-bin":
		result, err = c.StopBin(requestText, session, addr)
	case "s-stop-http":
		result, err = c.StopHttp(requestText, session, addr)
	default:
		{
			result, err = c.translateToNode(nodeId, function, requestText, session)
		}
	}

	if err == nil {
		return result, nil
	}
	return []byte(""), err
}

func (c *SrvRepeater) RequestJsonFromBinary(function string, requestText []byte, session *storage.Session, nodeId string, addr string) ([]byte, error) {
	c.mtx.Lock()
	c.counterRequests++
	c.mtx.Unlock()

	var err error
	var result []byte

	if nodeId == "" {
		switch function {
		case "s-registered-nodes":
			result, err = c.Nodes(session)
		case "s-account-info":
			result, err = c.AccountInfo(session, requestText)
		case "s-node-add":
			result, err = c.NodeAdd(session, requestText, addr)
		case "s-node-update":
			result, err = c.NodeUpdate(session, requestText, addr)
		case "s-node-remove":
			result, err = c.NodeRemove(session, requestText, addr)
		case "session_open":
			result, err = c.SessionOpen(requestText)
		case "session_activate":
			result, err = c.SessionActivate(requestText)
		case "session_close":
			err = c.SessionClose(requestText, session)
		}
	} else {
		result, err = c.translateToNode(nodeId, function, requestText, session)
	}

	if err == nil {
		return result, nil
	}
	return []byte(""), err
}

func (c *SrvRepeater) AllNodes(requestText []byte) (bs []byte, err error) {
	var req AllNodesRequest
	_ = json.Unmarshal(requestText, &req)
	resp := c.localNodesDb.AllNodes(req.Key)
	bs, err = json.MarshalIndent(resp, "", " ")
	return
}

func (c *SrvRepeater) Where(requestText []byte) (bs []byte, err error) {
	type WhereRequest struct {
		NodeId string `json:"node_id"`
	}

	var req WhereRequest
	err = json.Unmarshal(requestText, &req)
	if err != nil {
		return
	}

	var host string
	host, err = c.regDB.Where(req.NodeId)
	if err != nil {
		logger.Println("WHERE Leave error")
		return
	}

	type WhereResponse struct {
		NodeId string `json:"node_id"`
		Host   string `json:"host"`
	}

	var resp WhereResponse
	resp.NodeId = req.NodeId
	resp.Host = host
	bs, err = json.MarshalIndent(resp, "", " ")

	return
}

func (c *SrvRepeater) AccountInfo(session *storage.Session, requestText []byte) (bs []byte, err error) {
	if session == nil {
		return nil, errors.New("no session")
	}

	type AccountInfoRequest struct {
	}

	var req AccountInfoRequest
	err = json.Unmarshal(requestText, &req)
	if err != nil {
		return
	}

	var accountInfo storage.AccountInfo
	accountInfo, err = c.storage.GetAccountInfo(session.UserId)
	if err != nil {
		return
	}
	bs, err = json.MarshalIndent(accountInfo, "", " ")
	return
}

func (c *SrvRepeater) RepeaterForNode(requestText []byte) (bs []byte, err error) {
	type RepeaterForNodeRequest struct {
		NodeId string `json:"node_id"`
	}

	var req RepeaterForNodeRequest
	err = json.Unmarshal(requestText, &req)
	if err != nil {
		return
	}

	var response RepeaterForNodeResponse
	response, err = c.regDB.RepeaterForNode(req.NodeId)
	if err != nil {
		logger.Println("RepeaterForNode Leave error")
		return
	}

	bs, err = json.MarshalIndent(response, "", " ")

	return
}

func (c *SrvRepeater) GetAllNodesOfCluster() (bs []byte, err error) {
	resp := c.regDB.GetAll()
	bs, err = json.MarshalIndent(resp, "", " ")
	return
}

func (c *SrvRepeater) GetAllDB() (bs []byte, err error) {
	resp := c.regDB.GetAllDB()
	bs, err = json.MarshalIndent(resp, "", " ")
	return
}
