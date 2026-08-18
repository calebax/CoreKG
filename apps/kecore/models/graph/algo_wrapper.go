package graph

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/nebulagraph"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type ParseAlgoWrapper struct {
	ctx     context.Context
	resault *GraphAlgoResp
	graph   *foresttype.ForestGraphInfo
	cli     *nebulagraph.NebulaCli
	tMap    map[string]*foresttype.GraphTag
	eMap    map[string]*foresttype.GraphTag
	// copiedManualNodeKeys 记录从上一版本复制到当前版本的手动节点 key（nodeName:tagName）
	// 用于后续算法节点插入时跳过同名同tag的节点，避免重复创建
	copiedManualNodeKeys map[string]struct{}
}

// NewParseAlgoWrapper 初始化算法解析
func NewParseAlgoWrapper(ctx context.Context, resault *GraphAlgoResp) (*ParseAlgoWrapper, error) {
	graph, err := GetGraph(ctx, resault.GraphID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logs.WarnContextf(ctx, "[NewParseAlgoWrapper] graph is not found, id:%d", resault.GraphID)
		} else {
			logs.ErrorContextf(ctx, "NewParseAlgoWrapper GetGraph err: %v", err)
		}
		return nil, err
	}
	cli, err := nebulagraph.NewNebulaCLI(ctx, graph.SpaceName)
	if err != nil {
		logs.ErrorContextf(ctx, "NewParseAlgoWrapper NewNebulaCLI error: %v", err)
		return nil, err
	}
	tMap, err := GetTagNameMapByGraphID(ctx, resault.GraphID, graph.VersionID)
	if err != nil {
		logs.ErrorContextf(ctx, "NewParseAlgoWrapper GetTagNameMapByGraphID err: %v", err)
		return nil, err
	}
	eMap, err := GetEdgeNameMapByGraphID(ctx, resault.GraphID, graph.VersionID)
	if err != nil {
		logs.ErrorContextf(ctx, "NewParseAlgoWrapper GetEdgeNameMapByGraphID err: %v", err)
		return nil, err
	}
	wrapper := &ParseAlgoWrapper{
		ctx:     ctx,
		resault: resault,
		graph:   graph,
		cli:     cli,
		tMap:    tMap,
		eMap:    eMap,
		copiedManualNodeKeys: make(map[string]struct{}),
	}
	if graph.ParseMode == foresttype.ParseModeAuto {
		err = wrapper.upsertTag()
		if err != nil {
			logs.ErrorContextf(ctx, "NewParseAlgoWrapper upsertTag err: %v", err)
			return nil, err
		}
		err = wrapper.upsertEdge()
		if err != nil {
			logs.ErrorContextf(ctx, "NewParseAlgoWrapper upsertEdge err: %v", err)
			return nil, err
		}
		// 等待两个心跳周期，防止报错
		time.Sleep(time.Second * 22)
	}
	return wrapper, nil
}

func (p *ParseAlgoWrapper) Close() {
	p.cli.Release()
}

const algoKey = "graph:task:graph_%d"

// GetLock 获取锁，一次一张图只允许一个进程插入图标，防止并发插入数据错误，10分钟超时，防止死锁
func GetLock(ctx context.Context, graphID uint) error {
	return redispool.Lock(fmt.Sprintf(algoKey, graphID), time.Minute*10)
}

// UnLock 释放锁
func UnLock(ctx context.Context, graphID uint) error {
	return redispool.UnLock(fmt.Sprintf(algoKey, graphID))
}
