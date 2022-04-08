package fastspring

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"http-server.org/gazer/credentials"
	"http-server.org/gazer/logger"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Connection struct {
	client *http.Client
}

func NewConnection() *Connection {
	var c Connection

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	c.client = &http.Client{Transport: tr}
	c.client.Timeout = 3 * time.Second

	return &c
}

func (c *Connection) GetProducts() (res ProductsGetAll, err error) {
	var resStr string
	resStr, err = c.Get("/products")
	if err != nil {
		logger.Println("FastSpring Error:", err)
		return
	}

	err = json.Unmarshal([]byte(resStr), &res)
	if err != nil {
		logger.Println("FastSpring Unmarshal Error:", err)
		return
	}

	return
}

func (c *Connection) CreateAccount(email string, countryISOCode string) (res AccountCreateResponse, err error) {
	var req AccountCreate
	req.Language = "en"
	req.Country = countryISOCode
	req.Contact.Last = "-"
	req.Contact.First = "email"
	req.Contact.Email = email
	bs, _ := json.Marshal(req)

	var resStr string
	resStr, err = c.Post("/accounts", bs)
	if err != nil {
		logger.Println("FastSpring Error:", err)
		return
	}

	err = json.Unmarshal([]byte(resStr), &res)
	if err != nil {
		logger.Println("FastSpring Unmarshal Error:", err)
		return
	}

	logger.Println("CreateAccount", res, "resp:", resStr)

	return
}

func (c *Connection) CreateSession(account string, quantity int64) (res SessionCreateResponse, err error) {
	var req SessionCreate
	req.Account = account
	req.Items = make([]SessionCreateItem, 1)
	req.Items[0].Product = credentials.FastSpringProduct
	req.Items[0].Quantity = quantity
	bs, _ := json.Marshal(req)

	var resStr string
	resStr, err = c.Post("/sessions", bs)
	if err != nil {
		logger.Println("FastSpring Error:", err)
		return
	}

	err = json.Unmarshal([]byte(resStr), &res)
	if err != nil {
		logger.Println("FastSpring Unmarshal Error:", err)
		return
	}

	logger.Println("CreateSession", res, "resp:", resStr)

	return
}

func (c *Connection) GetAllAccounts() (res AccountGetAll, err error) {
	var resStr string
	resStr, err = c.Get("/accounts/")
	if err != nil {
		logger.Println("FastSpring Error:", err)
		return
	}

	err = json.Unmarshal([]byte(resStr), &res)
	if err != nil {
		logger.Println("FastSpring Unmarshal Error:", err)
		return
	}

	return
}

func (c *Connection) GetAccountByEmail(email string) (res AccountGet, err error) {
	var resStr string
	resStr, err = c.Get("/accounts?email=" + email)
	if err != nil {
		logger.Println("FastSpring Error:", err)
		return
	}

	err = json.Unmarshal([]byte(resStr), &res)
	if err != nil {
		logger.Println("FastSpring Unmarshal Error:", err)
		return
	}

	logger.Println("FastSpring GetAccountByEmail:", resStr)

	return
}

func (c *Connection) GetAccounts(accountIDs []string) (res AccountGet, err error) {
	var resStr string

	accIDs := ""
	for i, accID := range accountIDs {
		if i > 0 {
			accIDs += ","
		}
		accIDs += accID
	}

	resStr, err = c.Get("/accounts/" + accIDs)
	if err != nil {
		logger.Println("FastSpring Error:", err)
		return
	}

	err = json.Unmarshal([]byte(resStr), &res)
	if err != nil {
		logger.Println("FastSpring Unmarshal Error:", err)
		return
	}

	logger.Println("FastSpring acc details:", resStr)

	return
}

func (c *Connection) Get(requestString string) (string, error) {
	var responseString string

	req, err := http.NewRequest("GET", "https://api.fastspring.com"+requestString, nil)
	req.SetBasicAuth(credentials.FastSpringUserName, credentials.FastSpringPassword)
	if err != nil {
		return "", err
	}

	response, err := c.client.Do(req)
	if err != nil {
		logger.Println("FastSpring http client error:", err)
	} else {
		content, _ := ioutil.ReadAll(response.Body)
		responseString = strings.TrimSpace(string(content))
		response.Body.Close()
	}

	return responseString, err
}

func (c *Connection) Post(requestString string, request []byte) (string, error) {
	var responseString string

	buf := bytes.NewBuffer(request)

	req, err := http.NewRequest("POST", "https://api.fastspring.com"+requestString, buf)
	req.SetBasicAuth(credentials.FastSpringUserName, credentials.FastSpringPassword)
	if err != nil {
		return "", err
	}

	response, err := c.client.Do(req)
	if err != nil {
		logger.Println("FastSpring http client error:", err)
	} else {
		content, _ := ioutil.ReadAll(response.Body)
		responseString = strings.TrimSpace(string(content))
		response.Body.Close()
	}

	return responseString, err
}
