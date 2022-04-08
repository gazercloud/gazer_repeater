package srv_repeater

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gazercloud/gazer_repeater/credentials"
	"github.com/gazercloud/gazer_repeater/logger"
	"github.com/gazercloud/gazer_repeater/storage"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type HttpServer struct {
	srv             *http.Server
	r               *mux.Router
	api             IHttpApi
	storage         *storage.Storage
	rootPath        string
	noAuthFunctions map[string]bool
}

type IHttpApi interface {
	RequestJson(function string, requestText []byte, session *storage.Session, nodeId string, addr string) ([]byte, error)
}

func CurrentExePath() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir
}

func NewHttpServer(api IHttpApi, storage *storage.Storage) *HttpServer {
	var c HttpServer
	c.api = api
	c.storage = storage
	c.rootPath = CurrentExePath() + "/www"

	c.noAuthFunctions = make(map[string]bool)
	c.noAuthFunctions["session_open"] = true
	c.noAuthFunctions["s-registration"] = true
	c.noAuthFunctions["s-confirm-registration"] = true
	c.noAuthFunctions["s-active-nodes"] = true
	c.noAuthFunctions["s-where-node"] = true
	c.noAuthFunctions["s-repeater-for-node"] = true
	c.noAuthFunctions["s-restore-password"] = true
	c.noAuthFunctions["s-reset-password"] = true
	c.noAuthFunctions["s-state"] = true
	c.noAuthFunctions["session_activate"] = true

	return &c
}

func (c *HttpServer) Start() {
	logger.Println("HttpServer start")
	c.r = mux.NewRouter()

	// API
	c.r.HandleFunc("/api/request", c.processApiRequest)
	c.r.HandleFunc("/api/fastspring-licence", c.fastSpringLicence)
	c.r.HandleFunc("/api/fastspring-hook", c.fastSpringHook)

	// Static files
	c.r.NotFoundHandler = http.HandlerFunc(c.processFileLocal)

	bsBundle, err := hex.DecodeString(credentials.HttpsBundle)
	if err != nil {
		logger.Println("[HttpServer]", "Start error(HttpsBundle):", err)
		return
	}
	bsPrivate, err := hex.DecodeString(credentials.HttpsPrivate)
	if err != nil {
		logger.Println("[HttpServer]", "Start error(HttpsPrivate):", err)
		return
	}

	cert, err := tls.X509KeyPair(bsBundle, bsPrivate)
	if err != nil {
		logger.Println("[HttpServer]", "Start error(X509KeyPair):", err)
		return
	}
	c.srv = &http.Server{
		Addr: ":443",
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}
	c.srv.Handler = c.r
	go c.thListen()
}

func (c *HttpServer) redirectTLS(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://gazer.cloud"+r.RequestURI, http.StatusMovedPermanently)
}

func (c *HttpServer) thListen() {
	logger.Println("[HttpServer]", "HttpServer thListen begin")
	err := c.srv.ListenAndServeTLS("", "")
	if err != nil {
		logger.Println("[HttpServer]", "[error]", "HttpServer thListen error: ", err)
	}
	logger.Println("[HttpServer]", "HttpServer thListen end")
}

func (c *HttpServer) Stop() error {
	return c.srv.Close()
}

func (c *HttpServer) processApiRequest(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	nodeId := ""
	var sessionToken string

	/*{
		// getting nodeCode
		originProcessing := strings.ReplaceAll(origin, "https://", "")
		indexOfDomain := strings.Index(originProcessing, "-n.gazer.cloud")
		if indexOfDomain >= 0 {
			nodeCode := originProcessing[:indexOfDomain]
			nodeId = nodeCode
		}
	}*/

	{
		// CORS
		if strings.HasSuffix(origin, "gazer.cloud") && strings.HasPrefix(origin, "https://") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
	}

	var err error
	var session *storage.Session

	var responseText []byte
	function := r.FormValue("fn")
	requestType := r.FormValue("rt")
	requestJson := r.FormValue("rj")
	requestJsonZ := r.FormValue("rjz")
	sessionToken = r.FormValue("s")
	if nodeId == "" {
		nodeId = r.FormValue("n")
	}

	// logger.Println("processApiRequest", function)

	if requestType == "z" {
		requestJson = requestJsonZ
	}

	if len(requestJson) < 1 {
		requestJson = "{}"
	}

	sessionTokenCookie, errSessionToken := r.Cookie("session_token")

	if errSessionToken == nil {
		if sessionToken == "" {
			sessionToken = sessionTokenCookie.Value
		}
	}

	if _, isNoAuthFunction := c.noAuthFunctions[function]; !isNoAuthFunction {
		session, err = c.storage.SessionCheck(sessionToken)
		if err != nil {
			// even if no session - remove cookie
			if function == "session_close" {
				expiration := time.Now().Add(-365 * 24 * time.Hour)
				cookie := http.Cookie{Name: "session_token", Path: "/", Value: "", Expires: expiration, Domain: "gazer.cloud"}
				http.SetCookie(w, &cookie)
				logger.Println("[HttpServer]", "--- session_close cookie is set:", cookie)
			}
			logger.Println("[HttpServer]", "Session Token error: ", err, "Token:", sessionToken)
		}
	}

	// Execution of request
	requestJsonBytes := []byte(requestJson)
	if err == nil {
		if len(requestJson) > 0 {
			clearIPAddr := r.RemoteAddr
			indexIfColumn := strings.Index(clearIPAddr, ":")
			if indexIfColumn >= 0 {
				if indexIfColumn < len(clearIPAddr) {
					clearIPAddr = clearIPAddr[:indexIfColumn]
				}
			}
			responseText, err = c.api.RequestJson(function, requestJsonBytes, session, nodeId, clearIPAddr)
		}
	}

	if function == "s-confirm-registration" {
		if err == nil {
			c.redirect(w, r, "https://home.gazer.cloud/#form=confirmation_ok")
		} else {
			c.redirect(w, r, "https://home.gazer.cloud/#form=confirmation_error")
		}
		return
	}

	if err != nil {
		logger.Println("[HttpServer]", "[error]", "processApiRequest error", function, err)
		// Send error
		/*type ErrorObject struct {
			Error string `json:"error"`
		}
		var errObj ErrorObject
		errObj.Error = err.Error()*/
		w.WriteHeader(500)
		b := []byte(err.Error())
		_, _ = w.Write(b)
		return
	}

	//logger.Println("processApiRequest ok", function)

	if function == "session_open" || function == "session_activate" {
		// Set cookie
		type SessionOpenResponse struct {
			SessionToken string `json:"session_token"`
		}

		var sessionOpenResponse SessionOpenResponse
		errSessionOpenResp := json.Unmarshal(responseText, &sessionOpenResponse)
		if errSessionOpenResp == nil {
			expiration := time.Now().Add(365 * 24 * time.Hour)
			cookie := http.Cookie{Name: "session_token", Path: "/", Value: sessionOpenResponse.SessionToken, Expires: expiration, Domain: "gazer.cloud"}
			http.SetCookie(w, &cookie)
		}
	}

	if function == "session_close" {
		logger.Println("[HttpServer]", "session_close begin")
		// Set cookie
		type SessionOpenRequest struct {
			Key string `json:"key"`
		}

		logger.Println("[HttpServer]", "session_close key req:", string(requestJsonBytes))

		var sessionCloseRequest SessionOpenRequest
		errSessionOpenResp := json.Unmarshal(requestJsonBytes, &sessionCloseRequest)
		if errSessionOpenResp == nil {
			logger.Println("[HttpServer]", "--- session_close key:", sessionCloseRequest.Key)
			if session != nil {
				logger.Println("[HttpServer]", "--- session_close sessionKey:", session.Key)
			}
			if session != nil && sessionCloseRequest.Key == session.Key {
				expiration := time.Now().Add(-365 * 24 * time.Hour)
				cookie := http.Cookie{Name: "session_token", Path: "/", Value: "", Expires: expiration, Domain: "gazer.cloud"}
				http.SetCookie(w, &cookie)
				logger.Println("[HttpServer]", "--- session_close cookie is set:", cookie)
			}
		}
		logger.Println("[HttpServer]", "--- session_close end")
	}

	// Send normal response
	_, _ = w.Write(responseText)
}

func (c *HttpServer) contentTypeByExt(ext string) string {
	var builtinTypesLower = map[string]string{
		".css":  "text/css; charset=utf-8",
		".gif":  "image/gif",
		".htm":  "text/html; charset=utf-8",
		".html": "text/html; charset=utf-8",
		".jpeg": "image/jpeg",
		".jpg":  "image/jpeg",
		".js":   "text/javascript; charset=utf-8",
		".mjs":  "text/javascript; charset=utf-8",
		".pdf":  "application/pdf",
		".png":  "image/png",
		".svg":  "image/svg+xml",
		".wasm": "application/wasm",
		".webp": "image/webp",
		".xml":  "text/xml; charset=utf-8",
	}

	logger.Println("Ext: ", ext)

	if ct, ok := builtinTypesLower[ext]; ok {
		return ct
	}
	return "text/plain"
}

func (c *HttpServer) processFileLocal(w http.ResponseWriter, r *http.Request) {
	var err error
	var fileContent []byte
	var writtenBytes int

	urlPath := r.URL.Path
	//realIP := getRealAddr(r)

	//logger.Println("Real IP: ", realIP)
	//logger.Println("HttpServer processFile: ", r.URL.String())

	if strings.Contains(urlPath, "..") {
		logger.Println("HttpServer [ERROR] .. ", urlPath)
		return
	}

	if urlPath == "/" || urlPath == "" {
		urlPath = "/index.html"
	}

	var filePath string

	if strings.HasSuffix(r.Host, "-n.gazer.cloud") {
		filePath = "www/node" + urlPath
	}

	if strings.HasSuffix(r.Host, "client.gazer.cloud") {
		filePath = "www/node" + urlPath
	}

	if r.Host == "home.gazer.cloud" {
		filePath = "www/home" + urlPath
	}

	logger.Println("[HttpServer]", "getting file: ", urlPath, "filePath:", filePath)

	//res, err := webapp.GetAsset(filePath)
	res, err := ioutil.ReadFile(filePath)
	if err == nil {
		_, _ = w.Write(res)
	} else {
		logger.Println("[HttpServer]", "[error]", "getting file: ", urlPath, err)
		w.WriteHeader(404)
	}

	if err == nil {
		w.Header().Set("Content-Type", c.contentTypeByExt(filepath.Ext(filePath)))
		writtenBytes, err = w.Write(fileContent)
		if err != nil {
			logger.Println("[HttpServer]", "[error]", "sendError w.Write error:", err)
		}
		if writtenBytes != len(fileContent) {
			logger.Println("[HttpServer]", "[error]", "sendError w.Write data size mismatch. (", writtenBytes, " / ", len(fileContent))
		}
	} else {
		logger.Println("[HttpServer]", "[error]", "HttpServer processFile error: ", err)
		w.WriteHeader(404)
	}
}

func getRealAddr(r *http.Request) string {

	remoteIP := ""
	// the default is the originating ip. but we try to find better options because this is almost
	// never the right IP
	if parts := strings.Split(r.RemoteAddr, ":"); len(parts) == 2 {
		remoteIP = parts[0]
	}
	// If we have a forwarded-for header, take the address from there
	if xff := strings.Trim(r.Header.Get("X-Forwarded-For"), ","); len(xff) > 0 {
		addrs := strings.Split(xff, ",")
		lastFwd := addrs[len(addrs)-1]
		if ip := net.ParseIP(lastFwd); ip != nil {
			remoteIP = ip.String()
		}
		// parse X-Real-Ip header
	} else if xri := r.Header.Get("X-Real-Ip"); len(xri) > 0 {
		if ip := net.ParseIP(xri); ip != nil {
			remoteIP = ip.String()
		}
	}

	return remoteIP
}

func (c *HttpServer) fullPath(url string, host string) (string, error) {
	result := ""

	result = c.rootPath + "/" + url

	if strings.HasSuffix(host, "-n.gazer.cloud") {
		result = c.rootPath + "/node/" + url
	}

	if host == "home.gazer.cloud" {
		result = c.rootPath + "/home/" + url
	}

	fi, err := os.Stat(result)
	if err == nil {
		if fi.IsDir() {
			result += "/index.html"
		}
	}

	return result, err
}

func (c *HttpServer) redirect(w http.ResponseWriter, r *http.Request, url string) {
	w.Header().Set("Cache-Control", "no-cache, private, max-age=0")
	w.Header().Set("Expires", time.Unix(0, 0).Format(http.TimeFormat))
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("X-Accel-Expires", "0")
	http.Redirect(w, r, url, 307)
}

func (c *HttpServer) fastSpringLicence(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	resultStringForHash := ""
	type FormParam struct {
		Key   string
		Value string
	}
	formValues := make([]FormParam, 0)
	for key, values := range r.Form {
		if len(values) > 0 {
			var p FormParam
			p.Key = key
			p.Value = values[0]
			if key != "security_request_hash" {
				formValues = append(formValues, p)
			}
		}
	}

	sort.Slice(formValues, func(i, j int) bool {
		return formValues[i].Key < formValues[j].Key
	})

	for _, p := range formValues {
		resultStringForHash += p.Value
	}

	email := r.FormValue("email")
	product := r.FormValue("product")

	privateKey := "b6c2038e18e1502324dd83d1d7a0710e"
	resultStringForHash += privateKey
	md5Res := md5.Sum([]byte(resultStringForHash))
	md5Hex := hex.EncodeToString(md5Res[:])

	if md5Hex != r.Form.Get("security_request_hash") {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	res := ""
	res += "EMail: " + email + "\r\n"
	res += "Product: " + product + "\r\n"

	_, _ = w.Write([]byte(res))
}

func (c *HttpServer) fastSpringHook(w http.ResponseWriter, r *http.Request) {
	logger.Println("[HttpServer]", "fastSpringHook begin")

	c.storage.Log("i", "fastspring hook")

	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Println("[HttpServer]", "[error]", "fastSpringHook - ioutil.ReadAll error:", err)
		w.WriteHeader(500)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	type OrderCompleteEventItemDataAccountContact struct {
		First   string `json:"first"`
		Last    string `json:"last"`
		Email   string `json:"email"`
		Company string `json:"company"`
		Phone   string `json:"phone"`
	}

	type OrderCompleteEventItemDataAccount struct {
		Id      string                                   `json:"id"`
		Account string                                   `json:"account"`
		Contact OrderCompleteEventItemDataAccountContact `json:"contact"`
	}

	type OrderCompleteEventItemDataItem struct {
		Product  string  `json:"product"`
		Quantity int64   `json:"quantity"`
		Subtotal float64 `json:"subtotal"`
	}

	type OrderCompleteEventItemData struct {
		Order     string                            `json:"order"`
		Id        string                            `json:"id"`
		Reference string                            `json:"reference"`
		Live      bool                              `json:"live"`
		Account   OrderCompleteEventItemDataAccount `json:"account"`
		Items     []OrderCompleteEventItemDataItem  `json:"items"`
	}

	type OrderCompleteEventItem struct {
		Id        string                     `json:"id"`
		Processed bool                       `json:"processed"`
		Created   int64                      `json:"created"`
		Type      string                     `json:"type"`
		Live      bool                       `json:"live"`
		Data      OrderCompleteEventItemData `json:"data"`
	}

	type OrderCompleteEvent struct {
		Events []OrderCompleteEventItem `json:"events"`
	}

	var ev OrderCompleteEvent

	err = json.Unmarshal(bs, &ev)
	if err != nil {
		logger.Println("[HttpServer]", "[error]", "fastSpringHook - Unmarshal error:", err)
		w.WriteHeader(500)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if ev.Events != nil {
		for _, e1 := range ev.Events {
			email := e1.Data.Account.Contact.Email
			var accInfo storage.AccountInfo
			accInfo, err = c.storage.GetAccountInfoByEMail(email)
			if err != nil {
				logger.Println("[HttpServer]", "[error]", "fastSpringHook - GetAccountInfoByEMail error:", err)
				w.WriteHeader(500)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			if e1.Data.Items != nil {
				for _, e2 := range e1.Data.Items {
					_, err = c.storage.AddOrder(accInfo.Id, "FS-"+e1.Data.Id+"-"+e1.Data.Reference, e1.Data.Reference, true, e2.Product, e2.Quantity, e2.Subtotal, string(bs))
					if err != nil {
						logger.Println("[HttpServer]", "[error]", "fastSpringHook - AddOrder error:", err)
						c.storage.Log("e", "fastspring hook err email: "+email+" qty: "+fmt.Sprint(e2.Quantity)+" sum: "+fmt.Sprint(e2.Subtotal))
						w.WriteHeader(500)
						_, _ = w.Write([]byte(err.Error()))
						return
					}
					err = c.storage.UpdateMaxNodesCount(accInfo.Id)
					if err != nil {
						logger.Println("[HttpServer]", "[error]", "fastSpringHook - UpdateMaxNodesCount error:", err)
						c.storage.Log("e", "fastspring hook err email: "+email+" qty: "+fmt.Sprint(e2.Quantity)+" sum: "+fmt.Sprint(e2.Subtotal))
						w.WriteHeader(500)
						_, _ = w.Write([]byte(err.Error()))
						return
					}
					c.storage.Log("i", "fastspring hook ok email: "+email+" qty: "+fmt.Sprint(e2.Quantity)+" sum: "+fmt.Sprint(e2.Subtotal))
				}
			}
		}
	}
}
