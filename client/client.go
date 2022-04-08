package client

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

var client *http.Client

func init() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	client = &http.Client{Transport: tr}
	client.Timeout = 1 * time.Second
}

func Call(host string, function string, request []byte) (string, error) {
	var body bytes.Buffer
	var responseString string
	writer := multipart.NewWriter(&body)
	{
		fw, _ := writer.CreateFormField("fn")
		fw.Write([]byte(function))
	}
	{
		fw, _ := writer.CreateFormField("rj")
		if request == nil {
			fw.Write([]byte("{}"))
		} else {
			fw.Write(request)
		}

	}
	writer.Close()

	response, err := client.Post("https://"+host+"/api/request", writer.FormDataContentType(), &body)
	if err != nil {
		//logger.Println("http client error:", err)
	} else {
		content, _ := ioutil.ReadAll(response.Body)
		responseString = strings.TrimSpace(string(content))
		response.Body.Close()
	}

	return responseString, err
}
