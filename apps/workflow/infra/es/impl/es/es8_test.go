/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package es

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"

	esmodel "github.com/insmtx/corekg/apps/workflow/infra/es"
)

func TestES8ClientSearchDecodesResponseBody(t *testing.T) {
	t.Parallel()

	typedClient, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{"http://elasticsearch.test"},
		Transport: roundTripperFunc(func(request *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type":      []string{"application/json"},
					"X-Elastic-Product": []string{"Elasticsearch"},
				},
				Body: io.NopCloser(strings.NewReader(`{
					"hits": {
						"hits": [{"_id": "project-1", "_source": {"name": "demo"}}],
						"total": {"relation": "eq", "value": 1}
					},
					"timed_out": false,
					"took": 1
				}`)),
				Request: request,
			}, nil
		}),
	})
	if err != nil {
		t.Fatalf("create elasticsearch client: %v", err)
	}

	client := &es8Client{
		esClient: typedClient,
		types:    &es8Types{},
	}

	response, err := client.Search(context.Background(), "projects", &esmodel.Request{})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(response.Hits.Hits) != 1 {
		t.Fatalf("expected one hit, got %d", len(response.Hits.Hits))
	}
	if response.Hits.Hits[0].Id_ == nil || *response.Hits.Hits[0].Id_ != "project-1" {
		t.Fatalf("unexpected hit id: %v", response.Hits.Hits[0].Id_)
	}
}

type roundTripperFunc func(request *http.Request) (*http.Response, error)

func (function roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}
