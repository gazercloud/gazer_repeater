package srv_repeater

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

func HttpPostCallReCaptcha(host string, token string) (string, error) {
	var client *http.Client
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	client = &http.Client{Transport: tr}
	client.Timeout = 1 * time.Second

	var body bytes.Buffer
	var responseString string
	writer := multipart.NewWriter(&body)
	{
		fw, _ := writer.CreateFormField("secret")
		fw.Write([]byte("6LdySEMbAAAAAAlHUpcf0nm77Vz_jwIP5m-TDgcm"))
	}
	{
		fw, _ := writer.CreateFormField("response")
		fw.Write([]byte(token))
	}
	writer.Close()

	response, err := client.Post(host, writer.FormDataContentType(), &body)
	if err != nil {
		//logger.Println("http client error:", err)
	} else {
		content, _ := ioutil.ReadAll(response.Body)
		responseString = strings.TrimSpace(string(content))
		response.Body.Close()
	}

	return responseString, err
}
