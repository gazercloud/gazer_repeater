package srv_repeater

import (
	"errors"
	"fmt"
	"http-server.org/gazer/storage"
	"math/rand"
	"strconv"
	"time"
)

func (c *SrvRepeater) translateToNode(nodeId string, function string, requestText []byte, session *storage.Session) ([]byte, error) {
	var err error

	c.mtx.Lock()
	c.counterRetranslations++
	c.mtx.Unlock()

	if session == nil {
		c.mtx.Lock()
		c.counterRetranslationsNullSession++
		c.mtx.Unlock()
		return []byte(`{"error":"access denied: wrong session"}`), nil
	}

	// Search node
	var node *Node

	node = c.localNodesDb.nodeById(nodeId)

	if node == nil || node.sourceClient == nil {
		c.mtx.Lock()
		c.counterRetranslationsNoNodeFound++
		c.mtx.Unlock()
		return nil, errors.New("no node found [" + fmt.Sprint(nodeId) + "]")
	}

	if node.UserId != session.UserId {
		c.mtx.Lock()
		c.counterRetranslationsWrongUser++
		c.mtx.Unlock()
		return nil, errors.New("access denied [" + fmt.Sprint(nodeId) + "]")
	}

	// Unique Transaction Id
	transactionId := strconv.FormatInt(rand.Int63(), 16) + strconv.FormatInt(time.Now().UnixNano(), 16)

	// ProxyTask
	var task ProxyTask
	task.Function = function
	task.RequestText = requestText
	task.TransactionId = transactionId
	task.ResponseReceived = false
	c.mtx.Lock()
	c.proxyTasks[transactionId] = &task
	c.counterRetranslationsAdded++
	c.mtx.Unlock()

	// Send frame to node
	var frame BinFrame
	frame.Header.Function = function
	frame.Header.TransactionId = transactionId
	frame.Header.IsRequest = true
	frame.Data = requestText
	node.sourceClient.SendData(&frame)

	// Waiting for response
	tBegin := time.Now()
	for time.Now().Sub(tBegin) < 5*time.Second && !task.ResponseReceived {
		time.Sleep(10 * time.Millisecond)
	}

	// Remove task
	c.mtx.Lock()
	delete(c.proxyTasks, transactionId)
	c.mtx.Unlock()

	var resultBytes []byte

	if task.ResponseReceived {
		c.mtx.Lock()
		c.counterRetranslationsSuccess++
		c.mtx.Unlock()

		resultBytes = task.ResponseText
		if task.ErrorReceived {
			err = errors.New(string(resultBytes))
		}
	} else {
		err = errors.New("node timeout")
		c.mtx.Lock()
		c.counterRetranslationsNodeTimeout++
		c.mtx.Unlock()
	}

	return resultBytes, err
}
