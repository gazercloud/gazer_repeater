package srv_repeater

import (
	"encoding/json"
	"errors"
	"fmt"
	"http-server.org/gazer/credentials"
	"http-server.org/gazer/geolite"
	"http-server.org/gazer/logger"
	"http-server.org/gazer/pay/fastspring"
	"http-server.org/gazer/state"
	"http-server.org/gazer/storage"
	"math"
	"sync"
	"time"
)

type SrvRepeater struct {
	mtx                          sync.Mutex
	localNodesDb                 *LocalNodesDb
	lastBackgroundOperationsTime time.Time

	httpSrv          *HttpServer
	binSrv           *BinServer
	chProcessingData chan BinFrameTask

	proxyTasks map[string]*ProxyTask

	storage *storage.Storage
	regDB   *RegDB

	started  bool
	stopping bool

	// Statistics
	counterRequests                  int64
	counterRetranslations            int64
	counterRetranslationsNullSession int64
	counterRetranslationsNoNodeFound int64
	counterRetranslationsWrongUser   int64
	counterRetranslationsAdded       int64
	counterRetranslationsSuccess     int64
	counterRetranslationsNodeTimeout int64

	counterRequestsLast                  int64
	counterRetranslationsLast            int64
	counterRetranslationsNullSessionLast int64
	counterRetranslationsNoNodeFoundLast int64
	counterRetranslationsWrongUserLast   int64
	counterRetranslationsAddedLast       int64
	counterRetranslationsSuccessLast     int64
	counterRetranslationsNodeTimeoutLast int64

	counterRequestsSpeed                  float64
	counterRetranslationsSpeed            float64
	counterRetranslationsNullSessionSpeed float64
	counterRetranslationsNoNodeFoundSpeed float64
	counterRetranslationsWrongUserSpeed   float64
	counterRetranslationsAddedSpeed       float64
	counterRetranslationsSuccessSpeed     float64
	counterRetranslationsNodeTimeoutSpeed float64

	lastStatTime time.Time
}

type ProxyTask struct {
	TransactionId string
	Function      string
	RequestText   []byte

	ResponseReceived bool
	ErrorReceived    bool
	ResponseText     []byte
}

func NewSrvRepeater() *SrvRepeater {
	var c SrvRepeater
	c.localNodesDb = NewLocalNodesDb()
	c.proxyTasks = make(map[string]*ProxyTask)

	c.storage = storage.NewStorage()
	c.regDB = NewRegDB(c.storage)

	c.chProcessingData = make(chan BinFrameTask)
	c.httpSrv = NewHttpServer(&c, c.storage)
	c.binSrv = NewServer(c.chProcessingData)

	return &c
}

func (c *SrvRepeater) Start() {
	logger.Println("[SrvRepeater]", "Start. ID=", credentials.ServerId)
	if c.started {
		logger.Println("[SrvRepeater]", "[error]", "already started")
		return
	}
	c.stopping = false
	c.started = true

	c.httpSrv.Start()
	c.binSrv.Start()

	c.storage.Log("i", "started")

	go c.thWorker()
	logger.Println("[SrvRepeater]", "Started")
}

func (c *SrvRepeater) Stop() {
	logger.Println("[SrvRepeater]", "Stop. ID=", credentials.ServerId)
	if !c.started {
		return
	}
	c.stopping = true
	c.binSrv.Stop()
	c.httpSrv.Stop()

	c.storage.Log("i", "stopped")

	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		if !c.started {
			break
		}
	}
	if c.started {
		logger.Println("[SrvRepeater]", "[error]", "Stop - timeout")
	}
	logger.Println("[SrvRepeater]", "Stopped")
}

func (c *SrvRepeater) processStat() {
	//logger.Println("[SrvRepeater]", "processStat begin")
	c.mtx.Lock()
	now := time.Now()

	durationSec := now.Sub(c.lastStatTime).Seconds()
	c.counterRequestsSpeed = float64(c.counterRequests-c.counterRequestsLast) / durationSec
	c.counterRetranslationsSpeed = float64(c.counterRetranslations-c.counterRetranslationsLast) / durationSec
	c.counterRetranslationsNullSessionSpeed = float64(c.counterRetranslationsNullSession-c.counterRetranslationsNullSessionLast) / durationSec
	c.counterRetranslationsNoNodeFoundSpeed = float64(c.counterRetranslationsNoNodeFound-c.counterRetranslationsNoNodeFoundLast) / durationSec
	c.counterRetranslationsWrongUserSpeed = float64(c.counterRetranslationsWrongUser-c.counterRetranslationsWrongUserLast) / durationSec
	c.counterRetranslationsAddedSpeed = float64(c.counterRetranslationsAdded-c.counterRetranslationsAddedLast) / durationSec
	c.counterRetranslationsSuccessSpeed = float64(c.counterRetranslationsSuccess-c.counterRetranslationsSuccessLast) / durationSec
	c.counterRetranslationsNodeTimeoutSpeed = float64(c.counterRetranslationsNodeTimeout-c.counterRetranslationsNodeTimeoutLast) / durationSec

	c.counterRequestsLast = c.counterRequests
	c.counterRetranslationsLast = c.counterRetranslations
	c.counterRetranslationsNullSessionLast = c.counterRetranslationsNullSession
	c.counterRetranslationsNoNodeFoundLast = c.counterRetranslationsNoNodeFound
	c.counterRetranslationsWrongUserLast = c.counterRetranslationsWrongUser
	c.counterRetranslationsAddedLast = c.counterRetranslationsAdded
	c.counterRetranslationsSuccessLast = c.counterRetranslationsSuccess
	c.counterRetranslationsNodeTimeoutLast = c.counterRetranslationsNodeTimeout
	c.lastStatTime = time.Now()
	c.mtx.Unlock()
	//logger.Println("[SrvRepeater]", "processStat end")
}

func (c *SrvRepeater) backgroundOperations() {
	//logger.Println("[SrvRepeater]", "backgroundOperations")

	c.regDB.requestFromAllHosts()

	if time.Now().Sub(c.lastStatTime) > 1*time.Second {
		c.processStat()
	}
}

func (c *SrvRepeater) thWorker() {
	for !c.stopping {
		var frame BinFrameTask
		select {
		case frame = <-c.chProcessingData:
			go c.processData(frame)
		case <-time.After(50 * time.Millisecond):
		}

		needBackgroundOperations := false
		c.mtx.Lock()
		if time.Now().UTC().Sub(c.lastBackgroundOperationsTime) > 1000*time.Millisecond {
			c.lastBackgroundOperationsTime = time.Now().UTC()
			needBackgroundOperations = true
		}
		c.mtx.Unlock()

		if needBackgroundOperations {
			c.backgroundOperations()
		}
	}

	c.started = false
}

func (c *SrvRepeater) Nodes(session *storage.Session) (bs []byte, err error) {
	var resp *storage.NodesResponse
	resp, err = c.storage.Nodes(session.UserId)
	if err != nil {
		return
	}

	for index, item := range resp.Items {
		h, e := c.regDB.Where(item.Id)
		if e != nil {
		} else {
			resp.Items[index].CurrentRepeater = h
		}
	}

	bs, err = json.MarshalIndent(resp, "", " ")
	return
}

func (c *SrvRepeater) State() *state.System {
	var res state.System
	return &res
}

func (c *SrvRepeater) Buy(requestText []byte, session *storage.Session, addr string) (bs []byte, err error) {
	logger.Println("[SrvRepeater]", "buy")

	type BuyRequest struct {
		Quantity int64 `json:"quantity"`
	}
	var request BuyRequest
	err = json.Unmarshal(requestText, &request)
	if err != nil {
		return
	}

	var location geolite.Location
	location, err = geolite.Get(addr)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "get country:", err, "addr:", addr)
		return
	}

	var accountInfo storage.AccountInfo
	accountInfo, err = c.storage.GetAccountInfo(session.UserId)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "get account:", err, "userId:", session.UserId)
		return
	}

	email := accountInfo.Email

	type BuyInfo struct {
		StoreFrontURL string `json:"store_front_url"`
	}

	if len(credentials.AllowBuyWithEmail) > 0 {
		if email != credentials.AllowBuyWithEmail {
			logger.Println("[SrvRepeater]", "buy rejected. AllowBuyWithEmail:"+credentials.AllowBuyWithEmail+" email of the user: "+email)
			c.storage.Log("e", "buy rejected. AllowBuyWithEmail:"+credentials.AllowBuyWithEmail+" email of the user: "+email)
			err = errors.New("the payment system is currently unavailable. email us (admin@gazer.cloud) and we will give you free nodes")
			return
		}
	}

	conn := fastspring.NewConnection()

	var accountId string

	// Find account
	var acc fastspring.AccountGet
	acc, err = conn.GetAccountByEmail(email)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "get account from FastSpring:", err, "email:", email)
		return
	}

	if acc.Result != "success" || len(acc.Accounts) == 0 {
		logger.Println("[SrvRepeater]", "account"+email+"not found on FastSpring. Creating ...")
		var createAccRes fastspring.AccountCreateResponse
		createAccRes, err = conn.CreateAccount(email, location.CountryISO)
		if err != nil {
			return
		}
		if createAccRes.Result != "success" {
			err = errors.New("can not create account in FastSpring")
			return
		}

		accountId = createAccRes.Account

		logger.Println("[SrvRepeater]", "account "+email+" created in FastSpring.", createAccRes.Account)
	} else {
		logger.Println("[SrvRepeater]", "account "+email+" found in FastSpring. ID: ", acc.Accounts[0].Id)
		accountId = acc.Accounts[0].Id
	}

	logger.Println("[SrvRepeater]", "account "+email+" creating session ...")

	var fsSession fastspring.SessionCreateResponse
	fsSession, err = conn.CreateSession(accountId, request.Quantity)
	if err != nil {
		logger.Println("[SrvRepeater]", "[error]", "account "+email+" creating session error:", err)
		return
	}

	logger.Println("[SrvRepeater]", "account "+email+" creating session. sessionId: ", fsSession.Id)

	resURL := credentials.FastSpringUrlBuy + fsSession.Id

	logger.Println("[SrvRepeater]", "result URL: ", resURL)

	c.storage.Log("i", "buy begin sessionId: "+fsSession.Id)

	var info BuyInfo
	info.StoreFrontURL = resURL
	//info.StoreFrontURL = "https://gazer.cloud"
	bs, err = json.Marshal(info)
	return
}

func (c *SrvRepeater) StateAsString() ([]byte, error) {
	type State struct {
		Result         string `json:"result"`
		Retranslations int64  `json:"retranslations"`
		NodesCount     int    `json:"nodes_count"`

		RequestsSpeed                  string `json:"requests_speed"`
		RetranslationsSpeed            string `json:"retranslations_speed"`
		RetranslationsNullSessionSpeed string `json:"retranslations_null_session_speed"`
		RetranslationsNoNodeFoundSpeed string `json:"retranslations_no_node_found_speed"`
		RetranslationsWrongUserSpeed   string `json:"retranslations_wrong_user_speed"`
		RetranslationsAddedSpeed       string `json:"retranslations_added_speed"`
		RetranslationsSuccessSpeed     string `json:"retranslations_success_speed"`
		RetranslationsNodeTimeoutSpeed string `json:"retranslations_node_timeout_speed"`
	}

	c.mtx.Lock()
	var res State
	res.Result = fmt.Sprint(time.Now().Unix())
	res.NodesCount = c.localNodesDb.ActiveNodesCount()
	res.Retranslations = c.counterRetranslations
	res.RequestsSpeed = fmt.Sprint(math.Round(c.counterRequestsSpeed*100) / 100)
	res.RetranslationsSpeed = fmt.Sprint(math.Round(c.counterRetranslationsSpeed*100) / 100)
	res.RetranslationsNullSessionSpeed = fmt.Sprint(math.Round(c.counterRetranslationsNullSessionSpeed*100) / 100)
	res.RetranslationsNoNodeFoundSpeed = fmt.Sprint(math.Round(c.counterRetranslationsNoNodeFoundSpeed*100) / 100)
	res.RetranslationsWrongUserSpeed = fmt.Sprint(math.Round(c.counterRetranslationsWrongUserSpeed*100) / 100)
	res.RetranslationsAddedSpeed = fmt.Sprint(math.Round(c.counterRetranslationsAddedSpeed*100) / 100)
	res.RetranslationsSuccessSpeed = fmt.Sprint(math.Round(c.counterRetranslationsSuccessSpeed*100) / 100)
	res.RetranslationsNodeTimeoutSpeed = fmt.Sprint(math.Round(c.counterRetranslationsNodeTimeoutSpeed*100) / 100)
	c.mtx.Unlock()

	bs, err := json.MarshalIndent(res, "", " ")
	return bs, err
}

func (c *SrvRepeater) StopBin(requestText []byte, session *storage.Session, addr string) (bs []byte, err error) {
	c.binSrv.Stop()
	return make([]byte, 0), nil
}

func (c *SrvRepeater) StopHttp(requestText []byte, session *storage.Session, addr string) (bs []byte, err error) {
	c.httpSrv.Stop()
	return make([]byte, 0), nil
}
