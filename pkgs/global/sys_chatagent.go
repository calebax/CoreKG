package global

type AgentName string

const (
	// ChatAgentxxx key  原始值  新值
	ChatAgentxxx = "sys_agent"

	// ChatAgentxxx agent_excel_question_to_sql  4RyUuX7  sys_agent_excel_question_to_sql  // excel问题转sql
	ChatAgentExcelQuestionToSql AgentName = "sys_agent_excel_question_to_sql"

	// ChatAgentExcelSqlResultAnalysis agent_excel_sql_result_analysis  CJKGBXd  sys_agent_excel_sql_result_analysis // excel问答用
	ChatAgentExcelSqlResultAnalysis AgentName = "sys_agent_excel_sql_result_analysis"

	// ChatAgentExcelHeaderRowNumberListAnalysis 识别表头行号集合
	ChatAgentExcelHeaderRowNumberListAnalysis AgentName = "sys_agent_excel_header_row_number_list_analysis"

	// ChatAgentESChat agent_es_chat  3B2DqQT,RO1rf6I  sys_agent_es_chat  // es问答用
	ChatAgentESChat AgentName = "sys_agent_es_chat"

	// ChatAgentInetentionRecognition agent_intention_recognition  oSMfiUt  sys_agent_intention_recognition  // 意图识别
	ChatAgentInetentionRecognition AgentName = "sys_agent_intention_recognition"

	// ChatAgentSubQuestionChat agent_subquestion_chat  X9IIdTO  sys_agent_subquestion_chat
	ChatAgentSubQuestionChat AgentName = "sys_agent_subquestion_chat"

	// ChatAgentExcelIsFirstColumnRowTitle key  qzLGXdF  sys_agent_excel_is_first_column_row_title // 第一列是否是表头
	ChatAgentExcelIsFirstColumnRowTitle AgentName = "sys_agent_excel_is_first_column_row_title"

	// ChatAgentExcelHeaderRow key  CHzq0ri  sys_agent_excel_header_row
	ChatAgentExcelHeaderRow AgentName = "sys_agent_excel_header_row"
	// ChantAgentMysqlChoiceTableByQuestionKey mysql 表名意图识别
	ChantAgentMysqlChoiceTableByQuestionKey AgentName = "sys_agent_mysql_choice_table_by_question"
	// ChantAgentMysqlGenerateEcharts mysql 生成echarts
	ChantAgentMysqlGenerateEcharts AgentName = "sys_agent_mysql_generate_echarts"
	// ChatAgentMysqlIntentIsMeaningfulECharts 判断echarts图是否有意义
	ChatAgentMysqlIntentIsMeaningfulECharts AgentName = "sys_agent_mysql_intent_is_meaningful_echarts"
	// ChatAgentMysqlGenerateEchartsByChartType 根据图表类型生成echarts
	ChatAgentMysqlGenerateEchartsByChartType AgentName = "sys_agent_mysql_generate_echarts_by_chart_type"
	//ChatAgentAIWriteAbbreviation 缩写 npkCo26 sys_agent_write_abbreviation
	ChatAgentAIWriteAbbreviation AgentName = "sys_agent_write_abbreviation"
	//ChatAgentAIWriteExpansion 扩写 GeL4VTC sys_agent_write_expansion
	ChatAgentAIWriteExpansion AgentName = "sys_agent_write_expansion"
	//ChatAgentAIWriteEmbellishment 润色 KxnqohY sys_agent_write_embellishment
	ChatAgentAIWriteEmbellishment AgentName = "sys_agent_write_embellishment"
	//ChatAgentAIWriteProofreading 校对 rbAjHhW sys_agent_write_proofreading
	ChatAgentAIWriteProofreading AgentName = "sys_agent_write_proofreading"
	//ChatAgentAIWriteContinuation 续写 rbAjHhW sys_agent_write_continuation
	ChatAgentAIWriteContinuation AgentName = "sys_agent_write_continuation"
	//ChatAgentAIWriteCustom 自定义 somVdGP sys_agent_write_custom
	ChatAgentAIWriteCustom AgentName = "sys_agent_write_custom"

	//ChatAgentUserQueryRewrite 用户语义补充 rbAjHhW sys_agent_user_query_rewrite
	ChatAgentUserQueryRewrite AgentName = "sys_agent_user_query_rewrite"
	//ChatAgentQuestionAnswer 知识库问答 rbAjHhW sys_agent_question_answer
	// ChatAgentQuestionAnswer AgentName = "sys_agent_question_answer" // 旧版不带引用标识提示词
	ChatAgentQuestionAnswer            AgentName = "sys_agent_reference_question_answer" // 新版带引用标识提示词
	ChatAgentDescriptionQuestionAnswer AgentName = "sys_agent_question_answer"
)

// AgentNameI18nSupported 支持多语言的agent名称列表
var AgentNameI18nSupported = map[AgentName]struct{}{
	ChatAgentExcelQuestionToSql:              {},
	ChatAgentExcelSqlResultAnalysis:          {},
	ChatAgentESChat:                          {},
	ChatAgentInetentionRecognition:           {},
	ChatAgentSubQuestionChat:                 {},
	ChatAgentExcelIsFirstColumnRowTitle:      {},
	ChatAgentExcelHeaderRow:                  {},
	ChantAgentMysqlChoiceTableByQuestionKey:  {},
	ChantAgentMysqlGenerateEcharts:           {},
	ChatAgentMysqlGenerateEchartsByChartType: {},
	ChatAgentAIWriteAbbreviation:             {},
	ChatAgentAIWriteExpansion:                {},
	ChatAgentAIWriteEmbellishment:            {},
	ChatAgentAIWriteProofreading:             {},
	ChatAgentAIWriteCustom:                   {},
	ChatAgentQuestionAnswer:                  {},
	ChatAgentUserQueryRewrite:                {},
}

func (an AgentName) String() string {
	return string(an)
}

func (an AgentName) I18nSupported() bool {
	_, ok := AgentNameI18nSupported[an]
	return ok
}

// I18nName 多语言名称
func (an AgentName) I18nName(lang string) string {
	if an.I18nSupported() {
		return an.String() + "__" + lang
	}
	return an.String()
}

const (
	LLMURL     = "https://api.example.com/v3/llm.chat/chat/completions"
	LLMBaseURL = "https://api.example.com/v3/llm.chat"
)
