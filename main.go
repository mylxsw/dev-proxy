package main

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/mylxsw/asteria/log"
	uuid "github.com/satori/go.uuid"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.NewV4().String()

		director := func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = "localhost"

			body, _ := ioutil.ReadAll(req.Body)
			_ = req.Body.Close()
			req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

			log.WithFields(log.Fields{
				"url":     req.URL,
				"headers": req.Header,
				"body":    body,
			}).Debugf("request: %s", requestID)
		}

		respModifier := func(resp *http.Response) error {
			body, _ := ioutil.ReadAll(resp.Body)
			_ = resp.Body.Close()
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

			// var reader io.ReadCloser
			// switch resp.Header.Get("Content-Encoding") {
			// case "gzip":
			// 	reader, _ = gzip.NewReader(bytes.NewBuffer(body))
			// 	defer reader.Close()
			// default:
			// 	reader = ioutil.NopCloser(bytes.NewBuffer(body))
			// }

			// body, _ = ioutil.ReadAll(reader)

			log.WithFields(log.Fields{
				"status":      resp.Status,
				"status_code": resp.StatusCode,
				"length":      resp.ContentLength,
				"headers":     resp.Header,
				"body":        string(body),
			}).Debugf("response: %s", requestID)

			return nil
		}

		transport := http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}

		proxy := httputil.ReverseProxy{Director: director, ModifyResponse: respModifier, Transport: &transport}
		proxy.ServeHTTP(w, r)
	})
	http.ListenAndServe(":9090", nil)
}
