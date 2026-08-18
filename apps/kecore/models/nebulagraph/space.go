package nebulagraph

import (
	"context"
	"fmt"
	"time"

	nebula "github.com/vesoft-inc/nebula-go/v3"
	"github.com/ygpkg/yg-go/logs"
)

// CreateSpace 创建图数据库space 确保创建成功
func (cli *NebulaCli) CreateSpace(spaceName string) error {
	spaceConf := nebula.SpaceConf{
		Name:           spaceName,
		Partition:      1, // TODO 分区数
		Replica:        1,
		VidType:        "fixed_string(512)", // 民法典用例出现id过长的情况,512
		IgnoreIfExists: false,
		Comment:        "space for nebula graph",
	}

	//create space
	resp, err := cli.Session.CreateSpace(spaceConf)
	if err != nil {
		logs.ErrorContextf(cli.ctx, "fail to create space:space:%s err:%v", spaceName, err)
		return fmt.Errorf("fail to init create space:%s err:%v", spaceName, err)
	}
	logs.InfoContextf(cli.ctx, "create space:%s", logs.JSON(resp))

	return nil
}

func (cli *NebulaCli) CheckSpaceExists(ctx context.Context, spaceName string) error {
	for i := 0; i < 30; i++ {
		if _, err := cli.ExecuteAndCheck(fmt.Sprintf("USE %v;",
			spaceName)); err != nil {
			logs.WarnContextf(cli.ctx, "fail to use space:%v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		return nil
	}
	return fmt.Errorf("spaceName[%v] space no exist", spaceName)
}
