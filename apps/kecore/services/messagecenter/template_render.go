package messagecenter

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/ygpkg/yg-go/logs"
)

type TemplateRender struct {
}

func NewTemplateRender() *TemplateRender {
	return &TemplateRender{}
}

func (r *TemplateRender) Render(templateStr string, params map[string]string) (string, error) {
	if templateStr == "" {
		return "", nil
	}

	// 如果需要验证必需参数，设置 missingkey=error 选项
	// 这样在模板中访问不存在的字段时会返回错误
	tmpl := template.New("render").Option("missingkey=error")

	tmpl, err := tmpl.Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("template parse fail, template: %s, params: %s, err:%v", templateStr, logs.JSON(params), err)
	}

	// 执行模板渲染
	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("template execute fail, template: %s, params: %s, err:%v", templateStr, logs.JSON(params), err)
	}

	return buf.String(), nil
}
