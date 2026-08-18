package chatagent

import (
	"context"
	"testing"
)

func TestGetAgentDetail(t *testing.T) {
	// testutils.Initialize(testutils.AppNameKechat)
	// defer testutils.Close()
	_, err := GetAgentDetailByName(context.Background(), "XN1aGrl")
	if err != nil {
		// logs.Error(err)
		return
	}
	// logs.Infof("agent: %+v", agent)
}
