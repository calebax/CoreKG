package nebulagraph

import (
	"context"
	"fmt"

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

var (
	nebulaConf *NebulaConf
	pool       *nebula.ConnectionPool
)

func InitNebulaConf(ctx context.Context) error {
	conf := &NebulaConf{}
	err := settings.GetYaml("knowledge", "nebula", conf)
	if err != nil {
		logs.ErrorContextf(ctx, "get nebual config error: %v", err)
		return err
	}
	nebulaConf = conf
	hostAddress := nebula.HostAddress{Host: nebulaConf.Address, Port: nebulaConf.Port}
	hostList := []nebula.HostAddress{hostAddress}
	// Create configs for connection pool using default values
	testPoolConfig := nebula.GetDefaultConf()
	//fmt.Println(testPoolConfig)
	// Initialize connection pool
	pool, err = nebula.NewConnectionPool(hostList, testPoolConfig, nebula.DefaultLogger{})
	if err != nil {
		return fmt.Errorf("fail to initialize the connection pool, "+
			"host: %s, port: %d, %s", nebulaConf.Address, nebulaConf.Port, err.Error())
	}
	return nil
}

type NebulaCli struct {
	ctx context.Context
	*nebula.Session
}

// NewNebulaCLI new nebula session
func NewNebulaCLI(ctx context.Context, spaceName string) (*NebulaCli, error) {
	// Create session
	session, err := pool.GetSession(nebulaConf.UserName, nebulaConf.Password)
	if err != nil {
		return nil, fmt.Errorf("fail to create a new session from connection pool, "+
			"username: %s, password: %s, %s", nebulaConf.UserName, nebulaConf.Password, err.Error())
	}
	cli := &NebulaCli{
		ctx:     ctx,
		Session: session,
	}
	if spaceName != "" {
		err = cli.UseSpace(spaceName)
		if err != nil {
			logs.ErrorContextf(ctx, "fail to use space:%s, err:%v", spaceName, err)
			return nil, err
		}
	}

	return cli, nil
}

func (cli *NebulaCli) UseSpace(spaceName string) error {
	res, err := cli.ExecuteAndCheck(fmt.Sprintf("USE %s;", spaceName))
	if err != nil {
		logs.ErrorContextf(cli.ctx, "fail to use space:%s, res:%s", spaceName, logs.JSON(res))
		return err
	}
	return nil
}
