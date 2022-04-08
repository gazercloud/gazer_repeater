package srv_repeater

import (
	"encoding/json"
	"errors"
	"github.com/gazercloud/gazer_repeater/client"
	"github.com/gazercloud/gazer_repeater/logger"
	"github.com/gazercloud/gazer_repeater/storage"
	"sort"
	"sync"
	"time"
)

type RegDB struct {
	mtx                     sync.Mutex
	dbs                     map[string]*RegDBHostInfo
	hostByNodeId            map[string]*RegDBItem
	othersRepeaters         []string
	storage                 *storage.Storage
	lastUpdateRepeatersTime time.Time
}

type RegDBHostInfo struct {
	Host       string    `json:"host"`
	Nodes      []string  `json:"nodes"`
	Key        string    `json:"key"`
	DT         time.Time `json:"dt"`
	processing bool
}

type RegDBItem struct {
	NodeId string    `json:"node_id"`
	Host   string    `json:"host"`
	DT     time.Time `json:"dt"`
}

func NewRegDB(storage *storage.Storage) *RegDB {
	var c RegDB
	c.storage = storage
	c.dbs = make(map[string]*RegDBHostInfo)
	c.hostByNodeId = make(map[string]*RegDBItem)
	return &c
}

func (c *RegDB) CountDBS() int {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return len(c.dbs)
}

func (c *RegDB) CountOfNodes() int {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return len(c.hostByNodeId)
}

func (c *RegDB) Where(nodeId string) (host string, err error) {
	if item, ok := c.hostByNodeId[nodeId]; ok {
		return item.Host, nil
	}
	return "", errors.New("no node path found in system")
}

type RepeaterForNodeResponseItem struct {
	Host  string  `json:"host"`
	Score float64 `json:"score"`
}

type RepeaterForNodeResponse struct {
	NodeId string                        `json:"node_id"`
	Items  []RepeaterForNodeResponseItem `json:"items"`
	Host   string                        `json:"host"`
}

func (c *RegDB) RepeaterForNode(nodeId string) (result RepeaterForNodeResponse, err error) {
	result.NodeId = nodeId
	result.Items = make([]RepeaterForNodeResponseItem, 0)

	dt := time.Now().UTC()

	for _, hostInfo := range c.dbs {
		if dt.Sub(hostInfo.DT) < 10*time.Second {
			var item RepeaterForNodeResponseItem
			item.Score = float64(1000000 - len(hostInfo.Nodes))
			item.Host = hostInfo.Host
			result.Items = append(result.Items, item)
		}
	}

	sort.Slice(result.Items, func(i, j int) bool {
		return result.Items[i].Score < result.Items[j].Score
	})

	if len(result.Items) > 0 {
		result.Host = result.Items[0].Host
	}

	return
}

func (c *RegDB) ApplyDBFromHost(host string, db *AllNodesResponse) {
	c.mtx.Lock()
	var hostInfo *RegDBHostInfo
	dt := time.Now().UTC()

	var ok bool
	if hostInfo, ok = c.dbs[host]; !ok {
		hostInfo = &RegDBHostInfo{}
		c.dbs[host] = hostInfo
	}

	hostInfo.Host = host
	if !db.NoChanges {
		hostInfo.Nodes = db.Nodes
	}
	hostInfo.DT = dt
	hostInfo.Key = db.Key

	for _, node := range hostInfo.Nodes {
		var item *RegDBItem
		if item, ok = c.hostByNodeId[node]; ok {
			if dt.After(item.DT) {
				item.Host = host
				item.DT = dt
			}
		} else {
			var newItem RegDBItem
			newItem.Host = host
			newItem.DT = dt
			newItem.NodeId = node
			c.hostByNodeId[node] = &newItem
		}
	}

	// Remove old nodes
	nodesToRemove := make([]string, 0)
	for _, nodeInfo := range c.hostByNodeId {
		if dt.Sub(nodeInfo.DT) > 10*time.Second {
			nodesToRemove = append(nodesToRemove, nodeInfo.NodeId)
		}
	}
	for _, n := range nodesToRemove {
		delete(c.hostByNodeId, n)
	}

	c.mtx.Unlock()
}

func (c *RegDB) KeyForHost(host string) string {
	result := ""
	c.mtx.Lock()
	if hostInfo, ok := c.dbs[host]; ok {
		result = hostInfo.Key
	}
	c.mtx.Unlock()
	return result
}

type GetAllResponse struct {
	Items []RegDBItem `json:"items"`
}

type GetAllDBResponse struct {
	Hosts []RegDBHostInfo
	Items []RegDBItem `json:"items"`
}

func (c *RegDB) GetAll() *GetAllResponse {
	var resp GetAllResponse
	resp.Items = make([]RegDBItem, 0)
	c.mtx.Lock()
	for _, item := range c.hostByNodeId {
		resp.Items = append(resp.Items, *item)
	}
	sort.Slice(resp.Items, func(i, j int) bool {
		return resp.Items[i].NodeId < resp.Items[j].NodeId
	})
	c.mtx.Unlock()
	return &resp
}

func (c *RegDB) GetAllDB() *GetAllDBResponse {
	var resp GetAllDBResponse
	resp.Items = make([]RegDBItem, 0)
	c.mtx.Lock()
	for _, item := range c.hostByNodeId {
		resp.Items = append(resp.Items, *item)
	}
	resp.Hosts = make([]RegDBHostInfo, 0)
	for _, hostInfo := range c.dbs {
		resp.Hosts = append(resp.Hosts, *hostInfo)
	}

	sort.Slice(resp.Items, func(i, j int) bool {
		return resp.Items[i].NodeId < resp.Items[j].NodeId
	})

	sort.Slice(resp.Hosts, func(i, j int) bool {
		return resp.Hosts[i].Host < resp.Hosts[j].Host
	})

	c.mtx.Unlock()
	return &resp
}

func (c *RegDB) requestFromAllHosts() {
	if time.Now().Sub(c.lastUpdateRepeatersTime) > 5*time.Second {
		servers, err := c.storage.Servers()
		if err == nil {
			c.mtx.Lock()
			c.othersRepeaters = make([]string, 0)
			for _, s := range servers {
				if s.Role == "repeater" && s.Enabled {
					c.othersRepeaters = append(c.othersRepeaters, s.Code)
				}
			}
			c.mtx.Unlock()
		} else {
			logger.Println("[RegDB]", "requestFromAllHosts update other repeaters error:", err)
		}

		c.lastUpdateRepeatersTime = time.Now()
	}

	for _, rep := range c.othersRepeaters {
		go c.thRequestActiveNodes(rep)
	}
}

func (c *RegDB) thRequestActiveNodes(host string) {
	//logger.Println("request s-active-nodes", host)
	var req AllNodesRequest

	req.Key = c.KeyForHost(host)
	reqBS, _ := json.Marshal(req)

	var currentHostInfo *RegDBHostInfo
	alreadyProcessing := false
	currentHostInfo = nil
	c.mtx.Lock()
	if hostInfo, ok := c.dbs[host]; ok {
		currentHostInfo = hostInfo
	}
	if currentHostInfo != nil {
		alreadyProcessing = currentHostInfo.processing
		currentHostInfo.processing = true
	}
	c.mtx.Unlock()

	if alreadyProcessing {
		logger.Println("[RegDB]", "[error]", "request s-active-nodes alreadyProcessing", host)
		return
	}

	str, err := client.Call(host, "s-active-nodes", reqBS)

	if currentHostInfo != nil {
		currentHostInfo.processing = false
	}

	if err != nil {
		logger.Println("[RegDB]", "[error]", "request s-active-nodes error", host, err)
		return
	}

	var resp AllNodesResponse
	err = json.Unmarshal([]byte(str), &resp)
	if err != nil {
		logger.Println("[RegDB]", "[error]", "request s-active-nodes error parsing", host, err)
		return
	}

	c.ApplyDBFromHost(host, &resp)

	//logger.Println("request s-active-nodes ok", host)
}
