// Copyright 2019 ScientiaMobile Inc.
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

// userAgentFeeder allow to set a User Agent per request reading from a io.ReadSeeker feed.

package requester

import (
	"bufio"
	"io"
	"net/http"
	"sync"
)

func newUserAgentFeeder(reader io.ReadSeeker) *userAgentFeeder {
	return &userAgentFeeder{
		reader:  reader,
		scanner: bufio.NewScanner(reader),
	}
}

type userAgentFeeder struct {
	reader  io.ReadSeeker
	scanner *bufio.Scanner
	mutex   sync.Mutex
}

// userAgent returns a User Agent from the feed. When all the feed is read
// it will restart to read from the beginning
func (uaf *userAgentFeeder) userAgent() (string, error) {
	uaf.mutex.Lock()
	if uaf.scanner.Scan() == true {
		ua := uaf.scanner.Text()
		uaf.mutex.Unlock()
		return ua, nil
	}
	if err := uaf.scanner.Err(); err != nil {
		uaf.mutex.Unlock()
		return "", err
	}
	uaf.mutex.Unlock()
	uaf.reader.Seek(0, io.SeekStart)
	uaf.scanner = bufio.NewScanner(uaf.reader)
	return uaf.userAgent()
}

// Feed sets the request User Agent header with the feed entry. If feed entry is
// an empty string the UserAgent won't be updated.
func (uaf *userAgentFeeder) Feed(req *http.Request) {
	userAgent, _ := uaf.userAgent()
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
}
