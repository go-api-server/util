package httpclient

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	DEF_TIMEOUT               = 1 * time.Minute
	DEF_DIAL_TIMEOUT          = 30 * time.Second
	DEF_TLS_HANDSHAKE_TIMEOUT = 10 * time.Second
	DEF_KEEPLIVE_TIMEOUT      = 30 * time.Second

	DEF_IDLE_CONNS = 50
)
var slowWarnTimeout = time.Second

type Part struct {
	FieldName string
	FileName  string
	Data      []byte
}

type HttpClient struct {
	cli *http.Client
}

func NewDefClient() *HttpClient {
	return NewClient(DEF_TIMEOUT, DEF_DIAL_TIMEOUT, DEF_KEEPLIVE_TIMEOUT, DEF_TLS_HANDSHAKE_TIMEOUT, DEF_IDLE_CONNS, false)
}

func NewClient(timeout, dialTimeout, keepaliveTimeout, tlsHandshakeTimeout time.Duration, idleConnCnt int, skipVerify bool) *HttpClient {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   dialTimeout,
			KeepAlive: keepaliveTimeout,
		}).Dial,
		TLSHandshakeTimeout: tlsHandshakeTimeout,
		MaxIdleConnsPerHost: idleConnCnt,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: skipVerify},
	}

	cli := &http.Client{Transport: tr, Timeout: timeout}
	return &HttpClient{cli}
}

func (self *HttpClient) PostMultipart(url string, fields url.Values, files []*Part) (*http.Response, []byte, error) {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	for k, values := range fields {
		for _, v := range values {
			fieldWriter, err := writer.CreateFormField(k)
			if err != nil {
				return nil, nil, err
			}

			_, err = fieldWriter.Write([]byte(v))
			if err != nil {
				return nil, nil, err
			}
		}
	}

	for _, file := range files {
		fileWriter, err := writer.CreateFormFile(file.FieldName, file.FileName)
		if err != nil {
			return nil, nil, err
		}

		_, err = fileWriter.Write(file.Data)
		if err != nil {
			return nil, nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, nil, err
	}

	header := make(http.Header)
	header.Set("Content-Type", writer.FormDataContentType())

	return self.Post(url, header, &buffer)
}

func (self *HttpClient) PostJson(url string, data interface{}) (*http.Response, []byte, error) {
	header := make(http.Header)
	header.Set("Content-Type", "application/json")

	var buff bytes.Buffer
	encoder := json.NewEncoder(&buff)
	encoder.SetEscapeHTML(false) // 必须设置为False， 不允许把字符转成\uxxxx表示
	if err := encoder.Encode(data); err != nil {
		panic(fmt.Errorf("[HTTP_RPC] marshal data for Postjson Error :%s", err.Error()))
	}
	if buff.Len() > 1000 {
		v := buff.Bytes()
		log.Printf("[HTTP_RPC] POST-JSON: %s ... %s [Data too long. length=%d]", v[:500], v[len(v)-500:], len(v))
	} else {
		log.Printf("[HTTP_RPC] POST-JSON: %s", buff.String())
	}

	return self.Post(url, header, &buff)
}

func (self *HttpClient) PutJson(url string, data interface{}) (*http.Response, []byte, error) {
	header := make(http.Header)
	header.Set("Content-Type", "application/json")

	var buff bytes.Buffer
	encoder := json.NewEncoder(&buff)
	encoder.SetEscapeHTML(false) // 必须设置为False， 不允许把字符转成\uxxxx表示
	if err := encoder.Encode(data); err != nil {
		panic(fmt.Errorf("[HTTP_RPC] marshal data for PutJson Error :%s", err.Error()))
	}
	if buff.Len() > 1000 {
		v := buff.Bytes()
		log.Printf("[HTTP_RPC] PUT-JSON: %s ... %s [Data too long. length=%d]", v[:500], v[len(v)-500:], len(v))
	} else {
		log.Printf("[HTTP_RPC] PUT-JSON: %s", buff.String())
	}
	return self.Put(url, header, &buff)
}

func (self *HttpClient) PostForm(url string, data url.Values) (*http.Response, []byte, error) {
	header := make(http.Header)
	header.Set("Content-Type", "application/x-www-form-urlencoded")
	v := data.Encode()
	if len(v) > 1000 {
		log.Printf("[HTTP_RPC] POST-FORM: %s ... %s [Data too long. length=%d]", v[:500], v[len(v)-500:], len(v))
	} else {
		log.Printf("[HTTP_RPC] POST-FORM: %s", v)
	}
	return self.Post(url, header, strings.NewReader(v))
}

func (self *HttpClient) PutForm(url string, data url.Values) (*http.Response, []byte, error) {
	header := make(http.Header)
	header.Set("Content-Type", "application/x-www-form-urlencoded")
	v := data.Encode()
	if len(v) > 1000 {
		log.Printf("[HTTP_RPC] PUT-FORM: %s ... %s [Data too long. length=%d]", v[:500], v[len(v)-500:], len(v))
	} else {
		log.Printf("[HTTP_RPC] PUT-FORM: %s", v)
	}
	return self.Put(url, header, strings.NewReader(v))
}

func (self *HttpClient) Get(url string, header http.Header) (*http.Response, []byte, error) {
	return self.Request("GET", url, header, nil)
}

func (self *HttpClient) Post(url string, header http.Header, body io.Reader) (*http.Response, []byte, error) {
	return self.Request("POST", url, header, body)
}

func (self *HttpClient) Put(url string, header http.Header, body io.Reader) (*http.Response, []byte, error) {
	return self.Request("PUT", url, header, body)
}

func (self *HttpClient) Delete(url string, header http.Header) (*http.Response, []byte, error) {
	return self.Request("DELETE", url, header, nil)
}

func (self *HttpClient) Request(method, url string, header http.Header, body io.Reader) (*http.Response, []byte, error) {
	return self.RequestWithCookie(method, url, header, nil, body)
}

func (self *HttpClient) RequestWithCookie(method, url string, header http.Header, cookies []*http.Cookie, body io.Reader) (*http.Response, []byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, nil, err
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	for k, v := range header {
		if len(v) == 0 {
			continue
		}
		if len(v) > 1 {
			for _, h := range v {
				req.Header.Add(k, h)
			}
		} else {
			req.Header.Set(k, v[0])
		}
	}

	return self.Do(req)
}

func (self *HttpClient) Do(req *http.Request) (*http.Response, []byte, error) {
	startTime := time.Now()
	res, err := self.cli.Do(req)
	if err != nil {
		return nil, nil, err
	}
	var buf []byte
	duration := time.Since(startTime)
	defer func() {
		tmp := strings.Split(req.URL.String(), "?")
		logURL := tmp[0]
		if res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusBadRequest {
			// log.Printf("[HTTP_RPC] OK \"%s %s\" HEADER:[%v] - %d %.2fms",
			//	req.Method, logURL, req.Header, res.StatusCode, float64(duration)/float64(time.Millisecond))
		} else if res.StatusCode >= http.StatusInternalServerError {
			log.Printf("[HTTP_RPC] SERVER ERROR \"%s %s\" HEADER:[%v] RESPONSE:[%s] - %d %.2fms",
				req.Method, logURL, req.Header, buf, res.StatusCode, float64(duration)/float64(time.Millisecond))
		} else {
			log.Printf("[HTTP_RPC] FAIL \"%s %s\" HEADER:[%v] RESPONSE:[%s] - %d %.2fms",
				req.Method, logURL, req.Header, buf, res.StatusCode, float64(duration)/float64(time.Millisecond))
		}

		if duration >= slowWarnTimeout {
			log.Printf("[HTTP_RPC] [%s %s] too slow, use time: %d Sec", req.Method, logURL, duration/time.Second)
		}
		if res != nil {
			res.Body.Close()
		}
	}()

	buf, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}
	return res, buf, nil
}

func (self *HttpClient) GetClient() *http.Client {
	return self.cli
}
