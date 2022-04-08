package srv_repeater

import (
	"fmt"
	"http-server.org/gazer/logger"
	"math/rand"
	"sync"
	"time"
)

type LocalNodesDb struct {
	mtx         sync.Mutex
	nodes       map[string]*Node
	nodesIDs    []string
	nodesIDsKey string
}

func NewLocalNodesDb() *LocalNodesDb {
	var c LocalNodesDb
	c.nodes = make(map[string]*Node)
	c.nodesIDs = make([]string, 0)
	c.nodesIDsKey = c.NewNodesIDsKey()
	return &c
}

func (c *LocalNodesDb) NewNodesIDsKey() string {
	return fmt.Sprint(rand.Int(), "_", time.Now().UnixNano())
}

func (c *LocalNodesDb) updateNodesIDs() {
	c.nodesIDs = make([]string, 0)
	for _, n := range c.nodes {
		c.nodesIDs = append(c.nodesIDs, n.NodeId)
	}
	c.nodesIDsKey = c.NewNodesIDsKey()
}

func (c *LocalNodesDb) nodeById(nodeId string) *Node {
	var node *Node
	var ok bool
	c.mtx.Lock()
	node, ok = c.nodes[nodeId]
	if !ok {
		node = nil
	}
	c.mtx.Unlock()
	return node
}

func (c *LocalNodesDb) deactivateNodeByClient(client *RepeaterBinClient) {
	nodesIDsForRemove := make([]string, 0)

	for _, n := range c.nodes {
		if n.sourceClient == client {
			nodesIDsForRemove = append(nodesIDsForRemove, n.NodeId)
		}
	}
	for _, nodeId := range nodesIDsForRemove {
		delete(c.nodes, nodeId)
	}

	c.updateNodesIDs()
}

type AllNodesRequest struct {
	Key string `json:"key"`
}

type AllNodesResponse struct {
	Key       string   `json:"key"`
	NoChanges bool     `json:"no_changes"`
	Nodes     []string `json:"nodes"`
}

func (c *LocalNodesDb) AllNodes(key string) *AllNodesResponse {
	var resp AllNodesResponse
	if key == c.nodesIDsKey {
		c.mtx.Lock()
		resp.Nodes = make([]string, 0)
		resp.Key = c.nodesIDsKey
		resp.NoChanges = true
		c.mtx.Unlock()
	} else {
		c.mtx.Lock()
		resp.Nodes = make([]string, len(c.nodesIDs))
		resp.Key = c.nodesIDsKey
		resp.NoChanges = false
		copy(resp.Nodes, c.nodesIDs)
		c.mtx.Unlock()
	}
	return &resp
}

func (c *LocalNodesDb) ActiveNodesCount() int {
	c.mtx.Lock()
	res := len(c.nodes)
	c.mtx.Unlock()
	return res
}

func (c *LocalNodesDb) SetNode(nodeId string, name string, userId int64, client *RepeaterBinClient) {
	c.mtx.Lock()
	if existingNode, ok := c.nodes[nodeId]; ok {
		if existingNode.sourceClient != client {
			logger.Println("[LocalNodesDb]", "[error]", "SrvRepeater #iam one node - many connections ERROR")
			//existingNode.sourceClient.Stop()
			//delete(c.nodes, req.NodeId)
		}
	}
	var node Node
	node.NodeId = nodeId
	node.NodeName = name
	node.UserId = userId
	node.sourceClient = client
	c.nodes[nodeId] = &node
	c.updateNodesIDs()
	c.mtx.Unlock()

}
