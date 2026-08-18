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

package requestyygu

import (
	"context"
	"os"
	"testing"

	"github.com/ygpkg/yg-go/logs"
)

func TestDetailPersonalCenterResponse(t *testing.T) {
	if os.Getenv("TEST_INTEGRATION") != "true" {
		t.Skip("skipping integration test; set TEST_INTEGRATION=true to run")
	}
	res, err := DetailPersonalCenter(context.Background())
	if err != nil {
		logs.Errorf("failed to get personal center details: %v", err)
	}
	logs.Infof("personal center details: %+v", res)
}
