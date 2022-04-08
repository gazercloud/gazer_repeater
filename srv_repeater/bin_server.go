package srv_repeater

import (
	"crypto/tls"
	"encoding/hex"
	"http-server.org/gazer/credentials"
	"http-server.org/gazer/logger"
	"http-server.org/gazer/traffic_control"
	"log"
	"net"
	"sync"
	"time"
)

type BinServer struct {
	mtx              sync.Mutex
	serverListener   net.Listener
	clients          []*RepeaterBinClient
	chProcessingData chan BinFrameTask
	//auth             *users.Users
	started        bool
	startedWorker  bool
	startedCleaner bool
	stopping       bool
}

func NewServer(chProcessingData chan BinFrameTask) *BinServer {
	var c BinServer
	//c.auth = auth
	c.clients = make([]*RepeaterBinClient, 0)
	c.chProcessingData = chProcessingData
	return &c
}

func (c *BinServer) Start() {
	if c.started {
		return
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.started = true
	c.stopping = false
	c.startedWorker = true
	c.startedCleaner = true
	go c.thServerWorker()
	go c.thClientClearUp()
}

func (c *BinServer) Started() bool {
	return c.started
}

func (c *BinServer) Stop() {
	if !c.started {
		logger.Println("[BinServer]", "[error]", "already stopped")
		return
	}
	_ = c.serverListener.Close()
	c.stopping = true

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		if !c.startedWorker && !c.startedCleaner {
			break
		}
	}

	c.started = false
}

func (c *BinServer) thServerWorker() {
	var err error

	c.serverListener, err = tls.Listen("tcp", ":1077", c.tlsConfig())
	if err != nil {
		logger.Println("[BinServer]", "[error]", "tls.Listen error:", err)
		c.startedWorker = false
		return
	}

	logger.Println("[BinServer]", "thServerWorker started")

	for !c.stopping {
		logger.Println("[BinServer]", "accepting ...")
		conn, err := c.serverListener.Accept()
		if err != nil {
			logger.Println("[BinServer]", "[error]", "Accept error:", err)
			time.Sleep(1 * time.Second)
			continue
		}
		logger.Println("[BinServer]", "accepted client", conn.RemoteAddr().String())

		client := NewByConn(conn, c.chProcessingData, traffic_control.Stat())

		c.mtx.Lock()
		c.clients = append(c.clients, client)
		client.Start()
		c.mtx.Unlock()
	}

	c.startedWorker = false
}

func (c *BinServer) thClientClearUp() {
	for !c.stopping {
		time.Sleep(100 * time.Second)
		c.mtx.Lock()
		found := true
		for found {
			found = false
			for clientIndex, client := range c.clients {
				if !client.Started() {
					c.clients = append(c.clients[:clientIndex], c.clients[clientIndex+1:]...)
					found = true
					break
				}
			}
		}
		c.mtx.Unlock()
	}
	c.startedCleaner = false
}

func (c *BinServer) tlsConfig() *tls.Config {
	var err error
	var crt []byte
	var key []byte
	var cert tls.Certificate

	crt, err = hex.DecodeString(credentials.HttpsBundle)
	if err != nil {
		logger.Println("[BinServer]", "Start error(HttpsBundle):", err)
		return &tls.Config{}
	}
	key, err = hex.DecodeString(credentials.HttpsPrivate)
	if err != nil {
		logger.Println("[BinServer]", "Start error(HttpsPrivate):", err)
		return &tls.Config{}
	}

	cert, err = tls.X509KeyPair(crt, key)
	if err != nil {
		logger.Println("[BinServer]", "[error]", "X509KeyPair error:", err)
		log.Fatal(err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ServerName:   "gazer.cloud",
	}
}
