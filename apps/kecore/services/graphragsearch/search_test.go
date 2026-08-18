package graphragsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/kecore/models/nebulagraph"
	"github.com/insmtx/corekg/apps/kecore/services/graphragsearch/agent"
	"github.com/insmtx/corekg/apps/kecore/services/graphragsearch/data"
	"github.com/insmtx/corekg/apps/kecore/services/graphragsearch/search"
	"github.com/insmtx/corekg/pkgs/global"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

func TestDianlu(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
	})
	ctx := context.Background()
	err := nebulagraph.InitNebulaConf(ctx)
	if err != nil {
		t.Fatalf("init nebula config error: %v", err)
	}
	cli, err := nebulagraph.NewNebulaCLI(ctx, "a_car_test")
	if err != nil {
		t.Fatalf("new nebula cli error: %v", err)
	}
	// err = cli.CreateSpace("a_car_test")
	// if err != nil {
	// 	t.Fatalf("create space error: %v", err)
	// }
	// _, err = cli.ExecuteAndCheck("CREATE TAG `目录`(title string, page int);")
	// if err != nil {
	// 	t.Fatalf("create ExecuteAndCheck error: %v", err)
	// }
	// _, err = cli.ExecuteAndCheck("CREATE Edge `包含`();")
	// if err != nil {
	// 	t.Fatalf("create ExecuteAndCheck error: %v", err)
	// }
	// return
	_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT VERTEX `目录` (title) VALUES \"%s\":(\"%s\");", "电路图", "电路图"))
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}
	// 读取json文件映射
	dataBytes, err := os.ReadFile("./data/电路图目录.json")
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}
	var dianlu data.Dianlu
	err = json.Unmarshal(dataBytes, &dianlu)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	for _, v := range dianlu {
		_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT VERTEX `目录` (title) VALUES \"%s\":(\"%s\");", v.Chapter, v.Chapter))
		if err != nil {
			t.Fatalf("execute and check error: %v", err)
		}
		_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT EDGE `包含` () VALUES \"%s\"->\"%s\":();", "电路图", v.Chapter))
		if err != nil {
			t.Fatalf("execute and check error: %v", err)
		}
		for _, vv := range v.Sections {
			_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT VERTEX `目录` (title, page) VALUES \"%s\":(\"%s\",%d);", vv.Title, vv.Title, vv.Page))
			if err != nil {
				t.Fatalf("execute and check error: %v", err)
			}
			_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT EDGE `包含` () VALUES \"%s\"->\"%s\":();", v.Chapter, vv.Title))
			if err != nil {
				t.Fatalf("execute and check error: %v", err)
			}
		}
	}

}

func TestWeixiu(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
	})
	ctx := context.Background()
	err := nebulagraph.InitNebulaConf(ctx)
	if err != nil {
		t.Fatalf("init nebula config error: %v", err)
	}
	cli, err := nebulagraph.NewNebulaCLI(ctx, "a_car_test")
	if err != nil {
		t.Fatalf("new nebula cli error: %v", err)
	}
	_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT VERTEX `目录` (title) VALUES \"%s\":(\"%s\");", "维修手册", "维修手册"))
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}
	// 读取json文件映射
	dataBytes, err := os.ReadFile("./data/维修手册目录.json")
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}
	var weixiu data.Weixiu
	err = json.Unmarshal(dataBytes, &weixiu)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	for _, v := range weixiu {
		_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT VERTEX `目录` (title) VALUES \"%s\":(\"%s\");", v.ID+"_"+v.Title, v.Title))
		if err != nil {
			t.Fatalf("execute and check error: %v", err)
		}
		_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT EDGE `包含` () VALUES \"%s\"->\"%s\":();", "维修手册", v.ID+"_"+v.Title))
		if err != nil {
			t.Fatalf("execute and check error: %v", err)
		}
		for _, vv := range v.Sections {
			_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT VERTEX `目录` (title) VALUES \"%s\":(\"%s\");", vv.ID+"_"+vv.Title, vv.Title))
			if err != nil {
				t.Fatalf("execute and check error: %v", err)
			}
			_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT EDGE `包含` () VALUES \"%s\"->\"%s\":();", v.ID+"_"+v.Title, vv.ID+"_"+vv.Title))
			if err != nil {
				t.Fatalf("execute and check error: %v", err)
			}
			for _, vvv := range vv.Subsections {
				_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT VERTEX `目录` (title, page) VALUES \"%s\":(\"%s\",%d);", vvv.ID+"_"+vvv.Title, vvv.Title, vvv.Page))
				if err != nil {
					t.Fatalf("execute and check error: %v", err)
				}
				_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT EDGE `包含` () VALUES \"%s\"->\"%s\":();", vv.ID+"_"+vvv.Title, vvv.ID+"_"+vvv.Title))
				if err != nil {
					t.Fatalf("execute and check error: %v", err)
				}
			}
		}
		for _, vv := range v.Subsections {
			_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT VERTEX `目录` (title, page) VALUES \"%s\":(\"%s\",%d);", vv.ID+"_"+vv.Title, vv.Title, vv.Page))
			if err != nil {
				t.Fatalf("execute and check error: %v", err)
			}
			_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT EDGE `包含` () VALUES \"%s\"->\"%s\":();", v.ID+"_"+v.Title, vv.ID+"_"+vv.Title))
			if err != nil {
				t.Fatalf("execute and check error: %v", err)
			}
		}
	}

}

func TestZhenduan(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
	})
	ctx := context.Background()
	err := nebulagraph.InitNebulaConf(ctx)
	if err != nil {
		t.Fatalf("init nebula config error: %v", err)
	}
	cli, err := nebulagraph.NewNebulaCLI(ctx, "a_car_test")
	if err != nil {
		t.Fatalf("new nebula cli error: %v", err)
	}
	_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT VERTEX `目录` (title) VALUES \"%s\":(\"%s\");", "诊断手册", "诊断手册"))
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}
	// 读取json文件映射
	dataBytes, err := os.ReadFile("./data/诊断手册目录.json")
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}
	var zhenduan data.Zhenduan
	err = json.Unmarshal(dataBytes, &zhenduan)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	for _, v := range zhenduan {
		_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT VERTEX `目录` (title) VALUES \"%s\":(\"%s\");", v.System, v.System))
		if err != nil {
			t.Fatalf("execute and check error: %v", err)
		}
		_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT EDGE `包含` () VALUES \"%s\"->\"%s\":();", "诊断手册", v.System))
		if err != nil {
			t.Fatalf("execute and check error: %v", err)
		}
		for _, vv := range v.Entries {
			_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT VERTEX `目录` (title, page) VALUES \"%s\":(\"%s\",%d);", vv.Text, vv.Text, vv.Page))
			if err != nil {
				t.Fatalf("execute and check error: %v", err)
			}
			_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT EDGE `包含` () VALUES \"%s\"->\"%s\":();", v.System, vv.Text))
			if err != nil {
				t.Fatalf("execute and check error: %v", err)
			}
			for _, vvv := range vv.DTCs {
				_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT VERTEX `DTC` (title) VALUES \"%s\":(\"%s\");", vvv, vvv))
				if err != nil {
					t.Fatalf("execute and check error: %v", err)
				}
				_, err = cli.ExecuteAndCheck(fmt.Sprintf("INSERT EDGE `包含DTC` () VALUES \"%s\"->\"%s\":();", vv.Text, vvv))
				if err != nil {
					t.Fatalf("execute and check error: %v", err)
				}
			}
		}

	}

}

func TestGetTitle(t *testing.T) {
	ctx := context.Background()
	err := nebulagraph.InitNebulaConf(ctx)
	if err != nil {
		t.Fatalf("init nebula config error: %v", err)
	}
	cli, err := nebulagraph.NewNebulaCLI(ctx, "a_car_test")
	if err != nil {
		t.Fatalf("new nebula cli error: %v", err)
	}
	res, _ := search.GetTitleList(ctx, cli, "诊断手册")
	fmt.Println(res)
}

func TestAgent(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
	})
	ctx := context.Background()
	err := nebulagraph.InitNebulaConf(ctx)
	if err != nil {
		t.Fatalf("init nebula config error: %v", err)
	}
	// cli, err := nebulagraph.NewNebulaCLI(ctx, "a_car_test")
	// if err != nil {
	// 	t.Fatalf("new nebula cli error: %v", err)
	// }

	sysAPIKey, err := settings.GetText(global.SettingGroupKnowledge, global.SettingKeySystemLlmAPIKey)
	if err != nil {
		t.Fatalf("get system api key from setting fail, err: %w", err)
	}
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  sysAPIKey,
		Model:   "deepseek/deepseek-v3",
		BaseURL: "https://api.example.com/v3/llm.chat",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}
	res, err := agent.ExecuteCatalogueAgent(ctx, chatModel, "车辆同时报出故障码 P16E016（动力电池故障）和 U011287（与混合动力电池传感器模块失去通信），且车辆无法上电。请分析可能的根本原因，并给出Top-3最可能的故障模式，说明其推理路径。")
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}
	fmt.Println(res)
	// zhenduantitle, err := search.GetTitleList(ctx, cli, "诊断手册")
	// if err != nil {
	// 	t.Fatalf("get title list error: %v", err)
	// }
	// zhenduan := agent.NewCatalpgueAgent(ctx, chatModel, zhenduantitle)
	// ag, _ := agent.NewCatalpgueParallelAgent(ctx, chatModel)
	// // 创建 Runner
	// runner := adk.NewRunner(ctx, adk.RunnerConfig{
	// 	Agent: ag,
	// })

	// iter := runner.Query(ctx, "车辆同时报出故障码 P16E016（动力电池故障）和 U011287（与混合动力电池传感器模块失去通信），且车辆无法上电。请分析可能的根本原因，并给出Top-3最可能的故障模式，说明其推理路径。")
	// for {
	// 	event, ok := iter.Next()
	// 	if !ok {
	// 		break
	// 	}
	// 	if event.Err != nil {
	// 		log.Printf("分析过程中出现错误: %v", event.Err)
	// 		continue
	// 	}
	// 	Event(event)

	// }
}

func TestAAgent2(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
	})
	ctx := context.Background()
	err := nebulagraph.InitNebulaConf(ctx)
	if err != nil {
		t.Fatalf("init nebula config error: %v", err)
	}
	// cli, err := nebulagraph.NewNebulaCLI(ctx, "a_car_test")
	// if err != nil {
	// 	t.Fatalf("new nebula cli error: %v", err)
	// }
	fs.InitForestStorage()

	sysAPIKey, err := settings.GetText(global.SettingGroupKnowledge, global.SettingKeySystemLlmAPIKey)
	if err != nil {
		t.Fatalf("get system api key from setting fail, err: %w", err)
	}
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  sysAPIKey,
		Model:   "deepseek/deepseek-v3",
		BaseURL: "https://api.example.com/v3/llm.chat",
	})
	if err != nil {
		log.Fatalf("创建 ChatModel 失败: %v", err)
	}

	ag, _ := agent.NewAnalystAgent(ctx, chatModel, nil)
	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           ag,
		EnableStreaming: true,
	})

	iter := runner.Query(ctx, "车辆同时报出故障码 P16E016（动力电池故障）和 U011287（与混合动力电池传感器模块失去通信），且车辆无法上电。请分析可能的根本原因，并给出Top-3最可能的故障模式，说明其推理路径。")
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			log.Printf("分析过程中出现错误: %v", event.Err)
			continue
		}
		Event(event)

	}
}

func Event(event *adk.AgentEvent) {
	fmt.Printf("name: %s\npath: %s", event.AgentName, event.RunPath)
	if event.Output != nil && event.Output.MessageOutput != nil {
		if m := event.Output.MessageOutput.Message; m != nil {
			if len(m.Content) > 0 {
				if m.Role == schema.Tool {
					fmt.Printf("\ntool response: %s", m.Content)
				} else {
					fmt.Printf("\nanswer: %s", m.Content)
				}
			}
			if len(m.ToolCalls) > 0 {
				for _, tc := range m.ToolCalls {
					fmt.Printf("\ntool name: %s", tc.Function.Name)
					fmt.Printf("\narguments: %s", tc.Function.Arguments)
				}
			}
		} else if s := event.Output.MessageOutput.MessageStream; s != nil {
			toolMap := map[int][]*schema.Message{}
			var contentStart bool
			charNumOfOneRow := 0
			maxCharNumOfOneRow := 120
			for {
				chunk, err := s.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}
					fmt.Printf("error: %v", err)
					return
				}
				if chunk.Content != "" {
					if !contentStart {
						contentStart = true
						if chunk.Role == schema.Tool {
							fmt.Printf("\ntool response: ")
						} else {
							fmt.Printf("\nanswer: ")
						}
					}

					charNumOfOneRow += len(chunk.Content)
					if strings.Contains(chunk.Content, "\n") {
						charNumOfOneRow = 0
					} else if charNumOfOneRow >= maxCharNumOfOneRow {
						fmt.Printf("\n")
						charNumOfOneRow = 0
					}
					fmt.Printf("%v", chunk.Content)
				}

				if len(chunk.ToolCalls) > 0 {
					for _, tc := range chunk.ToolCalls {
						index := tc.Index
						if index == nil {
							log.Fatalf("index is nil")
						}
						toolMap[*index] = append(toolMap[*index], &schema.Message{
							Role: chunk.Role,
							ToolCalls: []schema.ToolCall{
								{
									ID:    tc.ID,
									Type:  tc.Type,
									Index: tc.Index,
									Function: schema.FunctionCall{
										Name:      tc.Function.Name,
										Arguments: tc.Function.Arguments,
									},
								},
							},
						})
					}
				}
			}

			for _, msgs := range toolMap {
				m, err := schema.ConcatMessages(msgs)
				if err != nil {
					log.Fatalf("ConcatMessage failed: %v", err)
					return
				}
				fmt.Printf("\ntool name: %s", m.ToolCalls[0].Function.Name)
				fmt.Printf("\narguments: %s", m.ToolCalls[0].Function.Arguments)
			}
		}
	}
	if event.Action != nil {
		if event.Action.TransferToAgent != nil {
			fmt.Printf("\naction: transfer to %v", event.Action.TransferToAgent.DestAgentName)
		}
		if event.Action.Interrupted != nil {
			for _, ic := range event.Action.Interrupted.InterruptContexts {
				str, ok := ic.Info.(fmt.Stringer)
				if ok {
					fmt.Printf("\n%s", str.String())
				} else {
					fmt.Printf("\n%v", ic.Info)
				}
			}
		}
		if event.Action.Exit {
			fmt.Printf("\naction: exit")
		}
	}
	if event.Err != nil {
		fmt.Printf("\nerror: %v", event.Err)
	}
	fmt.Println()
	fmt.Println()
}

func TestEntity(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
	})
	ctx := context.Background()
	err := nebulagraph.InitNebulaConf(ctx)
	if err != nil {
		t.Fatalf("init nebula config error: %v", err)
	}
	cli, err := nebulagraph.NewNebulaCLI(ctx, "a_car_test")
	if err != nil {
		t.Fatalf("new nebula cli error: %v", err)
	}
	// 读取json文件映射
	dataBytes, err := os.ReadFile("./data/本体实例(1).json")
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}
	var data data.BenTiEntity
	err = json.Unmarshal(dataBytes, &data)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	// 插入节点
	for _, node := range data.Nodes {
		var keys []string
		var values []string
		for k, v := range node.Properties {
			keys = append(keys, k)
			values = append(values, fmt.Sprintf("%v", v))
		}
		keystr := strings.Join(keys, "`, `")
		valuestr := strings.Join(values, "\", \"")
		str := fmt.Sprintf("INSERT VERTEX `%s` (`%s`) VALUES \"%s\":(\"%s\");", node.Lables[0], keystr, node.ID, valuestr)
		logs.Infof("node insert str: %s", str)
		_, err = cli.ExecuteAndCheck(str)
		if err != nil {
			t.Fatalf("execute and check error: %v", err)
		}
	}

	for _, edge := range data.Edges {
		str := fmt.Sprintf("INSERT EDGE %s () VALUES \"%s\"->\"%s\":(); ", edge.Type, edge.Src, edge.Dst)
		logs.Infof("edge insert str: %s", str)
		_, err = cli.ExecuteAndCheck(str)
		if err != nil {
			t.Fatalf("execute and check error: %v", err)
		}
	}
}

func TestEntities(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://root:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:31306/yygu_db?charset=utf8mb4&parseTime=true&loc=Local",
	})
	ctx := context.Background()
	err := nebulagraph.InitNebulaConf(ctx)
	if err != nil {
		t.Fatalf("init nebula config error: %v", err)
	}
	cli, err := nebulagraph.NewNebulaCLI(ctx, "a_car_test")
	if err != nil {
		t.Fatalf("new nebula cli error: %v", err)
	}
	// 读取json文件映射
	dataBytes, err := os.ReadFile("./data/10个测试问题部分实例.json")
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}
	var data []data.BenTiEntity
	err = json.Unmarshal(dataBytes, &data)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	for _, item := range data {
		// 插入节点
		for _, node := range item.Nodes {
			var keys []string
			var values []string
			for k, v := range node.Properties {
				keys = append(keys, k)
				values = append(values, fmt.Sprintf("%v", v))
			}
			keystr := strings.Join(keys, "`, `")
			valuestr := strings.Join(values, "\", \"")
			str := fmt.Sprintf("INSERT VERTEX `%s` (`%s`) VALUES \"%s\":(\"%s\");", node.Lables[0], keystr, node.ID, valuestr)
			logs.Infof("node insert str: %s", str)
			_, err = cli.ExecuteAndCheck(str)
			if err != nil {
				t.Fatalf("execute and check error: %v", err)
			}
		}

		for _, edge := range item.Edges {
			str := fmt.Sprintf("INSERT EDGE %s () VALUES \"%s\"->\"%s\":(); ", edge.Type, edge.Src, edge.Dst)
			logs.Infof("edge insert str: %s", str)
			_, err = cli.ExecuteAndCheck(str)
			if err != nil {
				t.Fatalf("execute and check error: %v", err)
			}
		}
	}
}
