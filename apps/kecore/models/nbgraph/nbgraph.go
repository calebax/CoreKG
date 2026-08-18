package nbgraph

import (
	"context"
	"fmt"
	"time"

	nebula "github.com/vesoft-inc/nebula-go/v3"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

type NebulaConf struct {
	Address  string `yaml:"address"`
	Port     int    `yaml:"port"`
	UserName string `yaml:"username"`
	Password string `yaml:"password"`
	Prefix   string `yaml:"prefix"`
}

var nebulaConf *NebulaConf

type NebulaCli struct {
	*nebula.Session
}

func InitNebulaConf(ctx context.Context) error {
	conf := &NebulaConf{}
	err := settings.GetYaml("knowledge", "nebula", conf)
	if err != nil {
		logs.ErrorContextf(ctx, "get nebual config error: %v", err)
		return err
	}
	nebulaConf = conf
	return nil
}

// GetNebulaSpace 获取及空间名
func GetNebulaSpace(uin, forestID uint) string {
	return fmt.Sprintf("%s_%v_%v", nebulaConf.Prefix, uin, forestID)
}

// NewNebulaCLI new nebula session
func NewNebulaCLI() (*NebulaCli, error) {

	hostAddress := nebula.HostAddress{Host: nebulaConf.Address, Port: nebulaConf.Port}
	hostList := []nebula.HostAddress{hostAddress}
	// Create configs for connection pool using default values
	testPoolConfig := nebula.GetDefaultConf()
	//fmt.Println(testPoolConfig)
	// Initialize connection pool
	pool, err := nebula.NewConnectionPool(hostList, testPoolConfig, nebula.DefaultLogger{})
	if err != nil {
		return nil, fmt.Errorf("fail to initialize the connection pool, "+
			"host: %s, port: %d, %s", nebulaConf.Address, nebulaConf.Port, err.Error())
	}

	// Create session
	session, err := pool.GetSession(nebulaConf.UserName, nebulaConf.Password)
	if err != nil {
		return nil, fmt.Errorf("fail to create a new session from connection pool, "+
			"username: %s, password: %s, %s", nebulaConf.UserName, nebulaConf.Password, err.Error())
	}

	return &NebulaCli{
		session,
	}, nil
}

// InitSpaceSchema init space and schema
func (cli *NebulaCli) InitSpaceSchema(ctx context.Context, uin, forestID, fileID uint) error {
	spaceConf := nebula.SpaceConf{
		Name:           GetNebulaSpace(uin, forestID),
		Partition:      1,
		Replica:        1,
		VidType:        "fixed_string(500)", // 民法典用例出现id过长的情况,暂改为500，新知识森林生效
		IgnoreIfExists: true,
		Comment:        "space for knownow forest",
	}

	//create space
	_, err := cli.CreateSpace(spaceConf)
	if err != nil {
		return fmt.Errorf("fail to init session space:%v", err)
	}
	//use space and loop to detect storage sync
	for {
		// if _, err = cli.ExecuteAndCheck(fmt.Sprintf("CLEAR SPACE %v;USE %v;",
		if _, err = cli.ExecuteAndCheck(fmt.Sprintf("USE %v;",
			spaceConf.Name)); err != nil {
			logs.WarnContextf(ctx, "fail to use space:%v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		break
	}

	if err = cli.CreateTag(ctx, fmt.Sprintf("doc_%v", fileID)+KnowNodeTagDFString); err != nil {
		return fmt.Errorf("fail to create tag:%v", err)
	}

	if err = cli.CreateEdge(ctx, fmt.Sprintf("doc_%v", fileID)+KnowEdgeDFString); err != nil {
		return fmt.Errorf("fail to create edge:%v", err)
	}

	return nil
}

// CreateTag create tag by a tagString
func (cli *NebulaCli) CreateTag(ctx context.Context, tagString string) error {
	logs.InfoContextf(ctx, "[nebula] create tag:%v", tagString)
	nGql := CREATE + TAG + IFNOTEXISTS + tagString
	_, err := cli.ExecuteAndCheck(nGql)
	if err != nil {
		return fmt.Errorf("%v,%v", nGql, err)
	}
	return nil
}

// CreateEdge create edge by an edgeString
func (cli *NebulaCli) CreateEdge(ctx context.Context, edgeString string) error {
	logs.InfoContextf(ctx, "[nebula] create edge:%v", edgeString)
	nGql := CREATE + EDGE + IFNOTEXISTS + edgeString
	_, err := cli.ExecuteAndCheck(nGql)
	if err != nil {
		return fmt.Errorf("%v,%v", nGql, err)
	}
	return nil
}

// InsertNode  insert a specific tag node
func (cli *NebulaCli) InsertNode(ctx context.Context, tagString string, nodeString string) error {
	logs.WarnContextf(ctx, "[nebula] start InsertNode :%v", tagString)
	nGql := INSERT + VERTEX + tagString + VALUES + nodeString
	_, err := cli.ExecuteAndCheck(INSERT + VERTEX + tagString + VALUES + nodeString)
	if err != nil {
		return fmt.Errorf("%v,%v", nGql, err)
	}
	return nil
}

// InsertEdge  insert a edge
func (cli *NebulaCli) InsertEdge(ctx context.Context, edgeDFString string, edgeString string) error {
	logs.InfoContextf(ctx, "[nebula] start InsertEdge :%v", edgeDFString)
	nGql := INSERT + EDGE + edgeDFString + VALUES + edgeString
	_, err := cli.ExecuteAndCheck(nGql)
	if err != nil {
		return fmt.Errorf("%v,%v", nGql, err)
	}
	return nil
}

func (cli *NebulaCli) Use(uin, forestID uint) error {
	nGql := fmt.Sprintf(USE+"know_%v_%v", uin, forestID)
	if _, err := cli.ExecuteAndCheck(nGql); err != nil {
		return fmt.Errorf("%v,%v", nGql, err)
	}
	return nil
}
