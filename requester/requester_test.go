// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package requester

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestN(t *testing.T) {
	var count int64
	handler := func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&count, int64(1))
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	w := &Work{
		Request: req,
		N:       20,
		C:       2,
	}
	w.Run()
	if count != 20 {
		t.Errorf("Expected to send 20 requests, found %v", count)
	}
}

func TestQps(t *testing.T) {
	var wg sync.WaitGroup
	var count int64
	handler := func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&count, int64(1))
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	w := &Work{
		Request: req,
		N:       20,
		C:       2,
		QPS:     1,
	}
	wg.Add(1)
	time.AfterFunc(time.Second, func() {
		if count > 2 {
			t.Errorf("Expected to work at most 2 times, found %v", count)
		}
		wg.Done()
	})
	go w.Run()
	wg.Wait()
}

func TestRequest(t *testing.T) {
	var uri, contentType, some, method, auth string
	handler := func(w http.ResponseWriter, r *http.Request) {
		uri = r.RequestURI
		method = r.Method
		contentType = r.Header.Get("Content-type")
		some = r.Header.Get("X-some")
		auth = r.Header.Get("Authorization")
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	header := make(http.Header)
	header.Add("Content-type", "text/html")
	header.Add("X-some", "value")
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header = header
	req.SetBasicAuth("username", "password")
	w := &Work{
		Request: req,
		N:       1,
		C:       1,
	}
	w.Run()
	if uri != "/" {
		t.Errorf("Uri is expected to be /, %v is found", uri)
	}
	if method != "GET" {
		t.Errorf("Method is expected to be GET, %v is found", uri)
	}
	if contentType != "text/html" {
		t.Errorf("Content type is expected to be text/html, %v is found", contentType)
	}
	if some != "value" {
		t.Errorf("X-some header is expected to be value, %v is found", some)
	}
	if auth != "Basic dXNlcm5hbWU6cGFzc3dvcmQ=" {
		t.Errorf("Basic authorization is not properly set")
	}
}

func TestBody(t *testing.T) {
	var count int64
	handler := func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		if string(body) == "Body" {
			atomic.AddInt64(&count, 1)
		}
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	req, _ := http.NewRequest("POST", server.URL, bytes.NewBuffer([]byte("Body")))
	w := &Work{
		Request:     req,
		RequestBody: []byte("Body"),
		N:           10,
		C:           1,
	}
	w.Run()
	if count != 10 {
		t.Errorf("Expected to work 10 times, found %v", count)
	}
}

func TestRaceConditionDNSLookup(t *testing.T) {
	var count int64
	handler := func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&count, int64(1))
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()
	u, _ := url.Parse(server.URL)

	req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:%s", u.Port()), nil)
	w := &Work{
		Request: req,
		N:       5000,
		C:       20,
	}
	w.Run()
	if count != 5000 {
		t.Errorf("Expected to send 5000 requests, found %v", count)
	}
}

func TestUserAgentFeeder(t *testing.T) {
	count := 0
	userAgentCount := make(map[string]int)
	mutex := &sync.Mutex{}
	handler := func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		count++
		userAgentCount[r.UserAgent()]++
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	// userAgentFeed contains also an empty feed that should fallback to
	// userAgentDefault
	userAgentFeed := []string{"hey/1.0.0", "", "hey/1.1.0"}
	userAgentDefault := "hey/0.0.1"

	uaf := strings.NewReader(strings.Join(userAgentFeed, "\n"))

	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("User-Agent", userAgentDefault)
	w := &Work{
		Request:       req,
		N:             6,
		C:             2,
		UserAgentFeed: uaf,
	}
	w.Run()
	if count != 6 {
		t.Errorf("Expected to send 6 requests, found %v", count)
	}

	for _, ua := range userAgentFeed {
		if ua == "" {
			ua = userAgentDefault
		}
		v, ok := userAgentCount[ua]
		if !ok {
			t.Errorf("Expected to send requests with user agent %s but no one was sent", ua)
		}
		if v != 2 {
			t.Errorf("Expected to send 2 requests with user agent %s, found %v", ua, v)
		}
	}
}
