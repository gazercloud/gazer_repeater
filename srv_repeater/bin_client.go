package srv_repeater

import (
	"crypto/tls"
	"encoding/binary"
	"github.com/gazercloud/gazer_repeater/logger"
	"github.com/gazercloud/gazer_repeater/tools"
	"net"
	"sync"
	"time"
)

type RepeaterBinClient struct {
	mtx              sync.Mutex
	addr             string
	conn             net.Conn
	chProcessingData chan BinFrameTask
	started          bool
	stopping         bool
	lastError        error
	sessionId        string
	userName         string
	password         string
	//auth             *users.Users
	stat     *tools.Statistics
	lastAddr string

	disconnectedSent bool
}

func NewByConn(conn net.Conn, chProcessingData chan BinFrameTask, stat *tools.Statistics) *RepeaterBinClient {
	var c RepeaterBinClient
	c.chProcessingData = chProcessingData
	c.conn = conn

	if conn != nil {
		c.lastAddr = conn.RemoteAddr().String()
	}

	//c.auth = auth
	c.stat = stat
	c.applyConnected()
	return &c
}

func New(addr string, userName string, password string, chProcessingData chan BinFrameTask) *RepeaterBinClient {
	var c RepeaterBinClient
	c.chProcessingData = chProcessingData
	c.addr = addr
	c.userName = userName
	c.password = password
	c.stat = tools.NewStatistics()
	return &c
}

func (c *RepeaterBinClient) SetStat(statistics *tools.Statistics) {
	c.stat = statistics
}

func (c *RepeaterBinClient) Connected() bool {
	return c.conn != nil
}

func (c *RepeaterBinClient) Start() {
	if c.started {
		return
	}
	c.started = true
	c.stopping = false
	go c.thConn()
	go c.thBackground()
}

func (c *RepeaterBinClient) Started() bool {
	return c.started
}

func (c *RepeaterBinClient) Stop() {
	if !c.started {
		return
	}
	c.stopping = true

	if c.conn != nil {
		_ = c.conn.Close()
	}
	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		if !c.started {
			break
		}
	}

	c.started = false
}

func (c *RepeaterBinClient) ShortString() string {
	result := ""
	if c.conn != nil {
		result += "A:[" + c.conn.RemoteAddr().String() + "]"
	} else {
		result += "A:[" + c.lastAddr + "]"
	}
	return result
}

func (c *RepeaterBinClient) SessionId() string {
	c.mtx.Lock()
	result := c.sessionId
	c.mtx.Unlock()
	return result
}

func (c *RepeaterBinClient) SetSession(sessionId string, userName string) {
	c.mtx.Lock()
	c.sessionId = sessionId
	c.userName = userName
	c.mtx.Unlock()
}

func (c *RepeaterBinClient) GetRemoteAddr() string {
	c.mtx.Lock()
	conn := c.conn
	c.mtx.Unlock()
	if conn != nil {
		return conn.RemoteAddr().String()
	}
	return "[no addr]"
}

func (c *RepeaterBinClient) UserName() string {
	return c.userName
}

func (c *RepeaterBinClient) LastError() error {
	return c.lastError
}

func (c *RepeaterBinClient) thBackground() {
	for !c.stopping {
		time.Sleep(200 * time.Millisecond)
	}
}

func (c *RepeaterBinClient) thConn() {
	logger.Println("binClient th started", c.ShortString())
	const inputBufferSize = 1024 * 1024
	inputBuffer := make([]byte, inputBufferSize)
	inputBufferOffset := 0
	for !c.stopping {
		if c.conn == nil {
			c.sessionId = ""
			inputBufferOffset = 0
			if len(c.addr) > 0 {
				var err error
				var conn net.Conn
				conn, err = tls.DialWithDialer(&net.Dialer{Timeout: time.Second * 1}, "tcp", c.addr, &tls.Config{})
				if err != nil {
					c.lastError = err
					c.conn = nil
					logger.Println("binClient th dial error", err, c.ShortString())
					time.Sleep(100 * time.Millisecond)
					continue
				}
				c.lastError = nil
				c.mtx.Lock()
				c.conn = conn
				c.mtx.Unlock()

				c.applyConnected()
			} else {
				logger.Println("binClient exiting", c.ShortString())
				break
			}
		}

		if inputBufferOffset >= inputBufferSize {
			logger.Println("max buffer size", c.ShortString())
			c.applyDisconnected()
			continue
		}

		n, err := c.conn.Read(inputBuffer[inputBufferOffset:])
		if err != nil {
			// connection closed: n = 0; err = EOF
			logger.Println("binClient read error:", err, c.ShortString())
			c.applyDisconnected()
			continue
		}
		if n == 0 {
			logger.Println("read 0", c.ShortString())
			c.applyDisconnected()
			continue
		}

		c.stat.Add("rcv", n)

		inputBufferOffset += n

		needExit := false
		processed := 0
		for inputBufferOffset-processed >= 4 {
			frameLen := int(binary.LittleEndian.Uint32(inputBuffer[processed:]))
			if frameLen < 8 || frameLen > inputBufferSize {
				logger.Println("wrong frame len", frameLen, c.ShortString())
				needExit = true
				break // critical error
			}
			unprocessedBufferLen := inputBufferOffset - processed
			if unprocessedBufferLen < frameLen {
				break // no enough data
			}

			var frameData BinFrameTask
			frameData.Client = c
			frameData.Frame, err = UnmarshalBinFrame(inputBuffer[processed : processed+frameLen])
			if err != nil {
				logger.Println("Error parse frame", err, c.ShortString())
			} else {

				if c.chProcessingData != nil {
					c.chProcessingData <- frameData
				} else {
					logger.Println("no processor", c.ShortString())
				}
			}

			processed += frameLen
		}

		if needExit {
			c.applyDisconnected()
			continue
		}

		if processed > 0 {
			copy(inputBuffer, inputBuffer[processed:inputBufferOffset])
			inputBufferOffset -= processed
		}
	}

	c.applyDisconnected()

	logger.Println("binClient exit", c.ShortString())
	c.started = false
}

func (c *RepeaterBinClient) applyConnected() {
	c.mtx.Lock()
	var frameData BinFrameTask
	frameData.Client = c
	frameData.IsConnectedSignal = true
	if c.chProcessingData != nil {
		c.chProcessingData <- frameData
	} else {
	}
	c.mtx.Unlock()
}

func (c *RepeaterBinClient) applyDisconnected() {
	c.mtx.Lock()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil

		var frameData BinFrameTask
		frameData.IsDisconnectedSignal = true
		frameData.Client = c
		if c.chProcessingData != nil {
			c.chProcessingData <- frameData
		}
	}
	/*if c.auth != nil {
		c.auth.CloseSession(c.sessionId)
	}*/
	c.mtx.Unlock()
}

func (c *RepeaterBinClient) SendData(data *BinFrame) {
	c.mtx.Lock()
	conn := c.conn
	if conn != nil {
		frameBytes, _ := data.Marshal()

		sent := 0
		for sent < len(frameBytes) {
			n, err := conn.Write(frameBytes)
			if err != nil {
				break
			}
			sent += n
			c.stat.Add("snd", n)
		}
	}
	c.mtx.Unlock()
}

func (c *RepeaterBinClient) Stat() *tools.Statistics {
	return c.stat
}
