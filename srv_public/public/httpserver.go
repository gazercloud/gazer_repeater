package public

import (
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"github.com/gorilla/mux"
	"http-server.org/gazer/credentials"
	"http-server.org/gazer/logger"
	"http-server.org/gazer/traffic_control"
	"http-server.org/gazer/traffic_logger"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type HttpServer struct {
	srv      *http.Server
	r        *mux.Router
	api      IHttpApi
	rootPath string
}

type IHttpApi interface {
	RequestJson(function string, requestText []byte) ([]byte, error)
}

func CurrentExePath() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir
}

func NewHttpServer(api IHttpApi) *HttpServer {
	var c HttpServer
	c.api = api
	c.rootPath = CurrentExePath() + "/www"
	return &c
}

func (c *HttpServer) Start() {
	logger.Println("HttpServer start")
	c.r = mux.NewRouter()

	// API
	c.r.HandleFunc("/api/request", c.processApiRequest)
	c.r.HandleFunc("/channel/{[A-Za-z0-9]+}", c.processChannel)
	c.r.HandleFunc("/item/{[A-Za-z0-9]+}/{.*}", c.processItem)
	c.r.HandleFunc("/node/{[A-Za-z0-9]+}", c.processNode)
	c.r.HandleFunc("/void", c.processVoid)

	// Static files
	c.r.NotFoundHandler = http.HandlerFunc(c.processFile)

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

	go func() {
		if err := http.ListenAndServe(":80", http.HandlerFunc(c.redirectTLS)); err != nil {
			logger.Println("ListenAndServe (redirectTLS) error: %v", err)
		}
	}()
}

func (c *HttpServer) redirectTLS(w http.ResponseWriter, r *http.Request) {
	traffic_logger.Write(getRealAddr(r), r.RequestURI, "http->https")
	http.Redirect(w, r, "https://gazer.cloud"+r.RequestURI, http.StatusMovedPermanently)
}

func (c *HttpServer) thListen() {
	logger.Println("HttpServer thListen begin")
	err := c.srv.ListenAndServeTLS("", "")
	if err != nil {
		logger.Println("HttpServer thListen error: ", err)
	}
	logger.Println("HttpServer thListen end")
}

func (c *HttpServer) Stop() error {
	return c.srv.Close()
}

func (c *HttpServer) Request(requestText string) (string, error) {
	var err error
	var respBytes []byte

	type Request struct {
		Function string `json:"func"`
		Path     string `json:"path"`
		Layer    string `json:"layer"`
	}
	var req Request
	err = json.Unmarshal([]byte(requestText), &req)
	if err != nil {
		return "", err
	}

	type Response struct {
		Value    string `json:"v"`
		DateTime string `json:"t"`
		Error    string `json:"e"`
	}

	var resp Response
	resp.Value = "123"
	resp.DateTime = time.Now().Format("2006-01-02 15-04-05.999")
	resp.Error = "ok"

	respBytes, err = json.MarshalIndent(resp, "", " ")
	if err != nil {
		return "", err
	}

	return string(respBytes), nil
}

func (c *HttpServer) processApiRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://gazer.cloud")

	var err error

	var responseText []byte
	function := r.FormValue("fn")

	requestJson := r.FormValue("rj")

	if len(requestJson) < 1 {
		requestJson = "{}"
	}

	traffic_control.AddRcv(len(requestJson) + 100)

	if len(requestJson) > 0 {
		requestJsonBytes := []byte(requestJson)
		responseText, err = c.api.RequestJson(function, requestJsonBytes)
		traffic_control.AddSend(len(responseText) + 100)
	}

	if err != nil {
		type ErrorObject struct {
			Error string `json:"error"`
		}

		var errObj ErrorObject
		errObj.Error = err.Error()

		w.WriteHeader(500)
		b, _ := json.Marshal(errObj)
		_, _ = w.Write(b)
		traffic_control.AddSend(len(b) + 100)
		return
	} else {
		if function == "session_open" {
			type SessionOpenResponse struct {
				SessionToken string `json:"session_token"`
			}

			var sessionOpenResponse SessionOpenResponse
			errSessionOpenResp := json.Unmarshal(responseText, &sessionOpenResponse)
			if errSessionOpenResp == nil {
				expiration := time.Now().Add(365 * 24 * time.Hour)
				cookie := http.Cookie{Name: "session_token", Path: "/", Value: sessionOpenResponse.SessionToken, Expires: expiration}
				http.SetCookie(w, &cookie)
			}
		}

	}

	_, _ = w.Write([]byte(responseText))
}

func (c *HttpServer) processVoid(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("1"))
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

func (c *HttpServer) processFile(w http.ResponseWriter, r *http.Request) {
	c.file(w, r, r.URL.Path)
}

func (c *HttpServer) processChannel(w http.ResponseWriter, r *http.Request) {
	c.file(w, r, "/channel_files/index.html")
}

func (c *HttpServer) processNode(w http.ResponseWriter, r *http.Request) {
	logger.Println("processNode")
	c.file(w, r, "/index.html")
}

func (c *HttpServer) processItem(w http.ResponseWriter, r *http.Request) {
	c.file(w, r, "/item_files/index.html")
}

func (c *HttpServer) file(w http.ResponseWriter, r *http.Request, urlPath string) {
	var err error
	var fileContent []byte
	var writtenBytes int

	realIP := getRealAddr(r)

	logger.Println("Real IP: ", realIP)
	logger.Println("HttpServer processFile: ", r.URL.String())

	originalURL := urlPath

	if urlPath == "/" || urlPath == "" {
		urlPath = "/index.html"
	}

	filePath, err := c.fullPath(urlPath, r.Host)

	logger.Println("FullPath: " + filePath)

	traffic_logger.Write(realIP, originalURL, "")

	if err != nil {
		w.WriteHeader(404)
		return
	}

	fileContent, err = ioutil.ReadFile(filePath)

	ext := filepath.Ext(filePath)
	if ext == ".html" {
		fileContent = c.processTemplate(fileContent, r.Host)
	}

	if err == nil {
		w.Header().Set("Content-Type", c.contentTypeByExt(filepath.Ext(filePath)))
		writtenBytes, err = w.Write(fileContent)
		if err != nil {
			logger.Println("HttpServer sendError w.Write error:", err)
		}
		if writtenBytes != len(fileContent) {
			logger.Println("HttpServer sendError w.Write data size mismatch. (", writtenBytes, " / ", len(fileContent))
		}
	} else {
		logger.Println("HttpServer processFile error: ", err)
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

func (c *HttpServer) sendError(w http.ResponseWriter, errorToSend error) {
	var err error
	var writtenBytes int
	var b []byte
	w.WriteHeader(500)
	b, err = json.Marshal(errorToSend.Error())
	if err != nil {
		logger.Println("HttpServer sendError json.Marshal error:", err)
	}
	writtenBytes, err = w.Write(b)
	if err != nil {
		logger.Println("HttpServer sendError w.Write error:", err)
	}
	if writtenBytes != len(b) {
		logger.Println("HttpServer sendError w.Write data size mismatch. (", writtenBytes, " / ", len(b))
	}
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

func (c *HttpServer) variablesOfTemplate(tmp []byte) map[string]string {
	result := make(map[string]string)
	return result
}

func (c *HttpServer) processTemplate(tmp []byte, host string) []byte {
	tmpString := string(tmp)
	reInclude := regexp.MustCompile(`\{#.*?#\}`)
	reVariables := regexp.MustCompile(`\{%.*?%\}`)
	reVariablesValues := regexp.MustCompile(`\{@.*?@\}`)

	includes := reInclude.FindAllString(tmpString, 100)

	for _, reString := range includes {
		filePath := strings.ReplaceAll(reString, "{#", "")
		filePath = strings.ReplaceAll(filePath, "#}", "")
		url, err := c.fullPath(filePath, host)
		if err != nil {
			logger.Println("processTemplate - c.fullpath(filePath) - ", err)
			continue
		}
		fileContent, err := ioutil.ReadFile(url)
		if err != nil {
			fileContent = []byte("-")
		} else {
			fileContent = c.processTemplate(fileContent, host)
		}
		tmpString = strings.ReplaceAll(tmpString, reString, string(fileContent))
	}

	variables := reVariables.FindAllString(tmpString, 100)
	vars := make(map[string]string)

	for _, reString := range variables {
		varString := strings.ReplaceAll(reString, "{%", "")
		varString = strings.ReplaceAll(varString, "%}", "")

		parts := strings.Split(varString, "=")
		if len(parts) == 2 {
			vars[parts[0]] = parts[1]
		}

		tmpString = strings.ReplaceAll(tmpString, reString, "")
	}

	variablesValues := reVariablesValues.FindAllString(tmpString, 100)
	for _, reString := range variablesValues {
		varString := strings.ReplaceAll(reString, "{@", "")
		varString = strings.ReplaceAll(varString, "@}", "")

		if value, ok := vars[varString]; ok {
			tmpString = strings.ReplaceAll(tmpString, reString, value)
		}
	}

	return []byte(tmpString)
}
