package srv_repeater

import (
	"encoding/json"
	"errors"
	"github.com/gazercloud/gazer_repeater/logger"
	"github.com/gazercloud/gazer_repeater/storage"
)

func (c *SrvRepeater) processData(task BinFrameTask) {
	var err error
	var session *storage.Session

	// Local signals
	if task.Frame == nil {
		if task.IsConnectedSignal {
			clientInfo := "-"
			if task.Client != nil {
				clientInfo = task.Client.ShortString()
			}
			logger.Println("SrvRepeater connected", clientInfo)
		}
		if task.IsDisconnectedSignal {
			clientInfo := ""
			if task.Client != nil {
				clientInfo = task.Client.ShortString()
				c.localNodesDb.deactivateNodeByClient(task.Client)
				logger.Println("SrvRepeater disconnected:", clientInfo)
			} else {
				logger.Println("SrvRepeater disconnected (no clientInfo)")
			}
		}
		return
	}

	// Check session
	if task.Frame.Header.Function != "session_open" {
		session, err = c.storage.SessionCheck(task.Frame.Header.CloudSessionId)
		if err != nil {
			logger.Println("SrvRepeater wrong session:", task.Frame.Header.CloudSessionId)
			return
		}
	}

	if task.Frame.Header.Function == "session_close" {
		clientInfo := task.Client.ShortString()
		c.localNodesDb.deactivateNodeByClient(task.Client)
		logger.Println("SrvRepeater logout(bin):", clientInfo)
	}

	// Process response
	if !task.Frame.Header.IsRequest {
		c.mtx.Lock()
		if tr, ok := c.proxyTasks[task.Frame.Header.TransactionId]; ok {
			tr.ResponseText = task.Frame.Data
			tr.ResponseReceived = true
			if len(task.Frame.Header.Error) > 0 {
				tr.ErrorReceived = true
				tr.ResponseText = []byte(task.Frame.Header.Error)
			}
		}
		c.mtx.Unlock()
	} else {
		originalTransactionId := task.Frame.Header.TransactionId

		// Process internal functions
		var bs []byte
		if task.Frame.Header.Function == "#iam" {
			err = c.IAm(task, session)
			bs = []byte("{}")
		} else {
			bs, err = c.RequestJsonFromBinary(task.Frame.Header.Function, task.Frame.Data, session, task.Frame.Header.TargetNodeId, task.Client.ShortString())
		}

		if err != nil {
			bs = []byte(err.Error())
		}

		var frame BinFrame
		frame.Header.Function = task.Frame.Header.Function
		frame.Header.TransactionId = originalTransactionId
		frame.Data = bs
		if err != nil {
			frame.Header.Error = err.Error()
		}
		task.Client.SendData(&frame)
	}
}

func (c *SrvRepeater) OpenSession(task BinFrameTask) {
	var err error
	clientInfo := "-"
	if task.Client != nil {
		clientInfo = task.Client.ShortString()
	}

	logger.Println("SrvRepeater #session_open ", clientInfo)

	type OpenSessionRequest struct {
		UserName string `json:"user_name"`
		Password string `json:"password"`
	}

	var req OpenSessionRequest

	err = json.Unmarshal(task.Frame.Data, &req)
	if err == nil {
		type SessionInfo struct {
			Key string `json:"key"`
		}
		var sessionInfo SessionInfo

		sessionInfo.Key, err = c.storage.SessionOpen(req.UserName, req.Password)
		logger.Println("SrvRepeater #session_open username:", req.UserName, "password:", req.Password)
		if err == nil {
			logger.Println("SrvRepeater #session_open sessionKey:", sessionInfo.Key)
			var frame BinFrame
			frame.Header.Function = "#session"
			frame.Header.TransactionId = ""
			frame.Data, _ = json.Marshal(sessionInfo)
			task.Client.SendData(&frame)
		} else {
			type ErrorStruct struct {
				Error string `json:"error"`
			}
			var e ErrorStruct
			e.Error = err.Error()
			var frame BinFrame
			frame.Header.Function = "#session"
			frame.Header.TransactionId = ""
			frame.Data, _ = json.Marshal(e)
			task.Client.SendData(&frame)
		}
	} else {
		logger.Println("SrvRepeater #session_open error", err)
	}

}

func (c *SrvRepeater) IAm(task BinFrameTask, session *storage.Session) (err error) {
	clientInfo := "-"
	if task.Client != nil {
		clientInfo = task.Client.ShortString()
	}

	logger.Println("SrvRepeater #iam", clientInfo)
	type IAmRequest struct {
		NodeId string `json:"node_id"`
	}

	if session == nil {
		err = errors.New("SrvRepeater #iam no session " + clientInfo)
		logger.Println("SrvRepeater #iam error:", err)
		return
	}

	var req IAmRequest
	err = json.Unmarshal(task.Frame.Data, &req)

	if err == nil && session != nil {
		var nodeName string
		nodeName, err = c.storage.CheckNode(session.UserId, req.NodeId)

		if err == nil {
			c.localNodesDb.SetNode(req.NodeId, nodeName, session.UserId, task.Client)
		} else {
			logger.Println("SrvRepeater #iam error", err)
			return
		}

		logger.Println("SrvRepeater #iam ok", clientInfo, "nodeId:", req.NodeId, "userId:", session.UserId)
	} else {
		sessionInfo := ""
		if session != nil {
			sessionInfo = session.Key
		}
		logger.Println("SrvRepeater #iam error", clientInfo, "nodeId:", req.NodeId, "session:", sessionInfo, "err:", err)
	}

	return
}

func (c *SrvRepeater) RegNode() {
}
