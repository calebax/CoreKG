package wx

import (
	"testing"

	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"gopkg.in/yaml.v3"
)

func TestWechatConfig(t *testing.T) {
	txt := `name: wechat_123
appid: 2341
appsecret: 12435
`
	cfg := &offConfig.Config{}
	err := yaml.Unmarshal([]byte(txt), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", cfg)

}
