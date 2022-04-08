package starter

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/gazercloud/gazer_repeater/credentials"
	"github.com/gazercloud/gazer_repeater/logger"
	"github.com/gazercloud/gazer_repeater/storage"
	"github.com/gorilla/mux"
	"net/http"
)

type HttpServer struct {
	srv        *http.Server
	r          *mux.Router
	storage    *storage.Storage
	Started    bool
	privateKey *rsa.PrivateKey
}

func NewHttpServer() *HttpServer {
	var c HttpServer
	c.privateKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	return &c
}

func (c *HttpServer) Start() {
	c.Started = true
	logger.Println("[STARTER]", "HttpServer start")
	c.r = mux.NewRouter()
	c.r.NotFoundHandler = http.HandlerFunc(c.process)
	c.srv = &http.Server{Addr: ":80"}
	c.srv.Handler = c.r
	go c.thListen()
}

func (c *HttpServer) thListen() {
	logger.Println("[STARTER]", "HttpServer thListen begin")
	err := c.srv.ListenAndServe()
	if err != nil {
		logger.Println("[STARTER]", "[error]", "HttpServer thListen error: ", err)
	}
	logger.Println("[STARTER]", "HttpServer thListen end")
}

func (c *HttpServer) Stop() error {
	c.srv.Close()
	c.Started = false
	return nil
}

func (c *HttpServer) process(w http.ResponseWriter, r *http.Request) {
	logger.Println("[Starter HttpServer]", "process begin")

	if r.URL.Path == "/key" {
		keyBS, _ := json.MarshalIndent(c.privateKey.PublicKey, "", " ")
		w.Write(keyBS)
		logger.Println("[Starter HttpServer]", "process send key")
		return
	}

	if r.URL.Path == "/start" {
		logger.Println("[Starter HttpServer]", "process start begin")
		label := []byte("")
		hash := sha256.New()
		dataHex := r.FormValue("data")
		data, _ := hex.DecodeString(dataHex)
		plainText, err := rsa.DecryptOAEP(
			hash,
			rand.Reader,
			c.privateKey,
			data,
			label,
		)

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		type StartBlock struct {
			ServerId   string `json:"server_id"`
			DbHost     string `json:"db_host"`
			DbUser     string `json:"db_user"`
			DbPassword string `json:"db_password"`
		}

		var startBlock StartBlock

		err = json.Unmarshal(plainText, &startBlock)

		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		logger.Println("[Starter HttpServer]", "process start server_id:", startBlock.ServerId)

		w.Write([]byte("loading credentials"))

		credentials.ServerId = startBlock.ServerId
		credentials.DbHost = startBlock.DbHost
		credentials.DbUser = startBlock.DbUser
		credentials.DbPassword = startBlock.DbPassword

		logger.Println("[Starter HttpServer]", "process loading directory")

		err = LoadDirectory(startBlock.ServerId)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		logger.Println("[Starter HttpServer]", "process loading directory ok")

		w.Write([]byte("ok: " + string(plainText)))
		c.Stop()
		logger.Println("[Starter HttpServer]", "process STOP")
		return
	}

	logger.Println("[Starter HttpServer]", "wrong path:", r.URL.Path)
	w.Write([]byte("wrong path:" + r.URL.Path))
}

func LoadDirectory(addr string) error {
	var err error
	st := storage.NewStorage()
	credentials.ServerRole, err = st.ServerRole(addr)
	if err != nil {
		logger.Println("[Starter LoadDirectory]", "error role:", err)
		return err
	}

	dir := make(map[string]string)
	dir, err = st.Directory()
	if err != nil {
		return err
	}
	logger.Println("[LoadDirectory]", "count of keys:", len(dir))

	//

	for key, value := range dir {
		logger.Println("[LoadDirectory]", key)
		switch key {
		case "email_smtp_server":
			credentials.EmailSmtpServer = value
		case "email_smtp_from":
			credentials.EmailSmtpFrom = value
		case "email_smtp_user":
			credentials.EmailSmtpUser = value
		case "email_smtp_password":
			credentials.EmailSmtpPassword = value

		case "https_bundle":
			credentials.HttpsBundle = value
		case "https_private":
			credentials.HttpsPrivate = value

		case "fastspring_product":
			credentials.FastSpringProduct = value
		case "fastspring_username":
			credentials.FastSpringUserName = value
		case "fastspring_password":
			credentials.FastSpringPassword = value
		case "fastspring_url_buy":
			credentials.FastSpringUrlBuy = value

		case "public_channels_user":
			credentials.PublicChannelUser = value
		case "public_channels_password":
			credentials.PublicChannelPassword = value
		case "public_channels_secret":
			credentials.PublicChannelsSecret = value

		case "allow_buy_with_email":
			credentials.AllowBuyWithEmail = value
		}
	}

	return nil
}
