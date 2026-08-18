package chatmodel

import (
	"fmt"
	"testing"
)

func TestValidModel(t *testing.T) {
	testCases := []string{
		"deepseek/deepseek-v3", // 命名空间风格
		"qwen2.5",              // 含点
		"llama3:8b",            // [新增支持] 含冒号，Ollama/Docker 风格
		"gemma:latest",         // [新增支持] 含冒号
		"my-model_v1.0:test",   // 混合使用
		":invalid",             // 非法：以冒号开头
		"invalid:",             // 合法（根据当前正则，以冒号结尾是被允许的，如果不想允许结尾带符号需调整）
	}

	for _, name := range testCases {
		fmt.Printf("Model: %-25s Valid: %v\n", name, ValidModel(name))
	}
}
