// Copyright 2019 ScientiaMobile Inc. All Rights Reserved.
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
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func Test_userAgentFeeder_userAgent(t *testing.T) {
	userAgent1 := "hey/1.0"
	userAgent2 := "hey/1.1"

	reader := strings.NewReader(fmt.Sprintf("%s\n%s\n", userAgent1, userAgent2))
	uaf := newUserAgentFeeder(reader)

	got, err := uaf.userAgent()
	if err != nil {
		t.Errorf("userAgentFeeder.userAgent() error = %v, want nil", err)
	}
	expected := userAgent1
	if got != expected {
		t.Errorf("userAgentFeeder.userAgent() = %v, want %v", got, expected)
	}

	got, err = uaf.userAgent()
	if err != nil {
		t.Errorf("userAgentFeeder.userAgent() error = %v, want nil", err)
	}
	expected = userAgent2
	if got != expected {
		t.Errorf("userAgentFeeder.userAgent() = %v, want %v", got, expected)
	}

	// Lines ended should start from beginning
	got, err = uaf.userAgent()
	if err != nil {
		t.Errorf("userAgentFeeder.userAgent() error = %v, want nil", err)
	}
	expected = userAgent1
	if got != expected {
		t.Errorf("userAgentFeeder.userAgent() = %v, want %v", got, expected)
	}
}

func Test_userAgentFeeder_Feed(t *testing.T) {
	// userAgentFeed contains also an empty feed that should fallback to
	// userAgentDefault
	userAgentFeed := []string{"hey/1.0.0", "", "hey/1.1.0"}
	userAgentDefault := "hey/0.0.1"

	reader := strings.NewReader(strings.Join(userAgentFeed, "\n"))
	uaf := newUserAgentFeeder(reader)

	req, _ := http.NewRequest("GET", "", nil)
	req.Header.Set("User-Agent", userAgentDefault)

	for _, ua := range userAgentFeed {
		r := cloneRequest(req, nil)
		uaf.Feed(r)
		if ua == "" {
			ua = userAgentDefault
		}
		if ua != r.UserAgent() {
			t.Errorf("userAgentFeeder.Feed(req) = %v, want %v", r.UserAgent(), ua)
		}
	}
}
