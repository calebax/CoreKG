package locales

import (
	"os"

	"github.com/ygpkg/yg-go/i18n"
	"golang.org/x/text/language"
)

var I18nConfig = i18n.I18nConfig{
	SupportedLanguages: []language.Tag{
		language.SimplifiedChinese,  // zh-Hans
		language.TraditionalChinese, // zh-Hant
		language.AmericanEnglish,    // en-US
	},
	DefaultLanguage: language.SimplifiedChinese,
}

func init() {
	// 从环境变量获取默认语言
	lang := os.Getenv("LANG")
	if lang == "" {
		lang = "zh-Hans"
	}
	t, err := language.Parse(lang)
	if err != nil {
		return
	}
	I18nConfig.DefaultLanguage = t
}
