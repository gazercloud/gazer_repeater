package public

import (
	"encoding/json"
	"errors"
	"github.com/gazercloud/gazer_repeater/logger"
	"github.com/gazercloud/gazer_repeater/state"
	"sync"
)

type SrvPublic struct {
	mtx      sync.Mutex
	stopping bool
	started  bool

	httpSrv *HttpServer
	//auth              *users.Users
}

func NewSrvPublic() *SrvPublic {
	var c SrvPublic
	//c.auth = users.NewUsers()
	c.httpSrv = NewHttpServer(&c)
	return &c
}

func (c *SrvPublic) Start() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.started {
		logger.Println("SrvPublic already started")
		return
	}

	c.httpSrv.Start()
	c.started = true
}

func (c *SrvPublic) Stop() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if !c.started {
		logger.Println("SrvPublic already stopped")
		return
	}

	c.httpSrv.Stop()

	c.stopping = true

	// nothing to stop
	c.started = false
}

func (c *SrvPublic) StateBinary() ([]byte, error) {
	result := c.State()
	bs, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (c *SrvPublic) State() *state.System {
	var result state.System
	var st state.Public
	result.Public = &st
	return &result
}

func (c *SrvPublic) RequestJson(function string, requestText []byte) ([]byte, error) {
	var err error
	var result []byte
	switch function {
	case "state":
		result, err = c.StateBinary()
	/*case "usage_statistics":
	c.SaveUsageStatistics(requestText)*/
	default:
		err = errors.New("function not supported")
	}

	if err == nil {
		return result, nil
	}

	return []byte(""), err
}

/*func (c *SrvPublic) SaveUsageStatistics(data []byte) {
	path := config.CurrentExePath() + "/usage/" + time.Now().Format("2006-01-02")
	_ = os.MkdirAll(path, 0777)
	filename := time.Now().Format("2006-01-02 15-04-05") + " - " + strconv.FormatInt(rand.Int63()%1000000, 10) + ".json"
	err := ioutil.WriteFile(path+"/"+filename, data, 0666)
	if err != nil {
		logger.Println("SaveUsageStatistics error: ", err)
	}
}
*/
