package httpclient

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/xzwsloser/TaskGo/pkg/logger"
)

type HttpClient struct {
	client	*http.Client
}

var _defaultHttpClient *HttpClient

const (
	GetMethod  = "GET"
	PostMethod = "POST"
)

var (
	ErrHttpClientNotInit = errors.New("Http Client Not Init")

	ErrNewRequest    = errors.New("Failed to new request")
	ErrReqSendFail   = errors.New("Failed to Send Req")
	ErrReqNotSuccess = errors.New("Status Code Not 200")
	ErrRespBodyRead	 = errors.New("Read From Resp Body Failed")
)

func NewHttpClient() *HttpClient {
	httpClient := &HttpClient{
		client: &http.Client{},
	}

	_defaultHttpClient = httpClient
	return httpClient
}

func (hc *HttpClient) Get(url string, timeout int64) (result string, err error) {
	req, err := http.NewRequest(GetMethod, url, nil)
	if err != nil {
		logger.GetLogger().Error("New Request Failed: " +  err.Error())
		err = ErrNewRequest
		return 
	}

	if timeout > 0 {
		hc.client.Timeout = time.Duration(timeout) * time.Second
	}

	resp, err := hc.client.Do(req)
	if err != nil {
		logger.GetLogger().Error("Send Get Http Request Failed: " + err.Error())
		err = ErrReqSendFail
		return 
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.GetLogger().Error("Http Request Status Code Not 200: " + err.Error())
		err = ErrReqNotSuccess
		return 
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.GetLogger().Error("Failed to read resp body: " + err.Error())
		err = ErrRespBodyRead
		return 
	}

	result = string(data)
	return 
}

func (hc *HttpClient) Post(url string, body string, timeout int64) (result string, err error) {
	buf := bytes.NewBufferString(body)
	req, err := http.NewRequest(PostMethod, url, buf)
	if err != nil {
		logger.GetLogger().Error("New Request Failed: " +  err.Error())
		err = ErrNewRequest
		return 
	}

	if timeout > 0 {
		hc.client.Timeout = time.Duration(timeout) * time.Second
	}

	resp, err := hc.client.Do(req)
	if err != nil {
		logger.GetLogger().Error("Send Post Http Request Failed: " + err.Error())
		err = ErrReqSendFail
		return 
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.GetLogger().Error("Http Request Status Code Not 200: " + err.Error())
		err = ErrReqNotSuccess
		return 
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.GetLogger().Error("Failed to read resp body: " + err.Error())
		err = ErrRespBodyRead
		return 
	}

	result = string(data)
	return 
}

func GetHttpClient() *HttpClient {
	return _defaultHttpClient
}


