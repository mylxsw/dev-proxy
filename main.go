package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/mylxsw/asteria/log"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Proxy  []Proxy `yaml:"proxy" json:"proxy"`
	Listen string  `yaml:"listen" json:"listen"`
}

type Proxy struct {
	Location    string `yaml:"location" json:"location"`
	StripPrefix bool   `yaml:"strip_prefix" json:"strip_prefix"`
	Host        string `yaml:"host" json:"host"`
}

func main() {

	confBytes, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Errorf("read config file failed: %v", err)
	}

	var conf Config
	if err := yaml.Unmarshal(confBytes, &conf); err != nil {
		log.Errorf("parse config file failed: %v", err)
	}

	log.WithFields(log.Fields{
		"config": conf,
	}).Infof("server started, listen %s", conf.Listen)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.NewV4().String()

		director := func(req *http.Request) {
			req.URL.Scheme = "http"

			for _, proxy := range conf.Proxy {
				if strings.HasPrefix(req.URL.Path, proxy.Location) {
					req.URL.Host = proxy.Host
					req.Host = proxy.Host
					if proxy.StripPrefix {
						req.URL.Path = strings.TrimPrefix(req.RequestURI, proxy.Location)
					}

					break
				}
			}

			message := resolveRequest(req)
			log.Debugf("request %s\n%s", requestID, message)
		}

		respModifier := func(resp *http.Response) error {
			message := resolveResponse(resp)
			log.Debugf("response: %s\n%s", requestID, message)
			resp.Header.Add("Dev-Proxy-ID", requestID)
			return nil
		}

		proxy := httputil.ReverseProxy{Director: director, ModifyResponse: respModifier}
		proxy.ServeHTTP(w, r)
	})

	err = http.ListenAndServe(conf.Listen, nil)
	if err != nil {
		log.Errorf("http server failed: %v", err)
	}
}

func resolveRequest(req *http.Request) string {
	var body []byte
	if req.Body != nil {
		body, _ = ioutil.ReadAll(req.Body)
		_ = req.Body.Close()
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}

	message := fmt.Sprintf("%s %s %s\n", req.Method, req.URL.Path, req.Proto)
	message += fmt.Sprintf("Host: %s\n", req.URL.Host)
	for k, v := range req.Header {
		for _, vv := range v {
			message += fmt.Sprintf("%s: %v\n", k, vv)
		}
	}

	message += "\n"
	message += string(body)
	return message
}

func resolveResponse(resp *http.Response) string {
	body, _ := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	message := fmt.Sprintf("%s %s\n", resp.Proto, resp.Status)
	for k, v := range resp.Header {
		for _, vv := range v {
			message += fmt.Sprintf("%s: %s\n", k, vv)
		}
	}
	if resp.ContentLength > 0 {
		message += fmt.Sprintf("Content-Length: %d\n", resp.ContentLength)
	}
	if len(resp.TransferEncoding) > 0 {
		message += fmt.Sprintf("Transfer-Encoding: %s\n", strings.Join(resp.TransferEncoding, ", "))
	}
	message += "\n"
	message += string(body)
	return message
}
