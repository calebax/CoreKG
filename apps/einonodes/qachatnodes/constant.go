package qachatnodes

const (
	NodeQAPairDataLoader                   = "qa_pair_data_loader"
	NodeExcelDDLDataLoader                 = "excel_ddl_data_loader"
	NodeChunkChatReferencesChunkDataLoader = "chunk_chat_references_data_loader"
	NodeChunkChatIntentRecognizer          = "chunk_chat_intent_recognizer"
	NodeMysqlChartIsMeaningful             = "mysql_chart_is_meaningful"
	NodeChunkChatLLMExecutor               = "chunk_chat_llm_executor"
	NodeExcelRunSQLExecutor                = "excel_run_sql_executor"
	NodeExcelTransferSQLToAnswerExecutor   = "excel_transfer_sql_to_answer_executor"
	NodeBlackExecutor                      = "black_executor"
	NodeMysqlGenerateEChartsByChartType    = "mysql_generate_echarts_by_chart_type"
	NodeNoDataReporter                     = "no_data_reporter"
	NodeEChartsReporter                    = "echarts_repsorter"
	NodeCheckStatementRes                  = "check_statement_res"
	NodeMysqlGenerateECharts               = "mysql_generate_echarts"

	NodeGetSessionTools    = "get_session_tools"
	NodeExternalSearchNode = "external_search_node"
	NodeExternalChat       = "external_chat_node"
	NodeGetESKeyWords      = "get_es_key_words"
	NodeGmailSearch        = "gmail_search_node"
	NodeSlackSearch        = "slack_search_node"
	NodeGoogleDriveSearch  = "google_drive_search_node"
	NodeConfluenceSearch   = "confluence_search_node"
)

const (
	RecordKeyForestID          = "forest_id"
	RecordKeyMySQLTables       = "mysql_tables"
	RecordKeyQueryStatement    = "query_statement"
	RecordKeyMySQLTableDDLMap  = "mysql_table_ddl_map"
	RecordKeyQueryStatementRes = "query_statement_res"
	RecordKeyNoDataCaseName    = "no_data_case_name"
	RecordKeyNextNode          = "next_node"
	RecordKeyChartType         = "chart_type"
	RecordKeyTextContent       = "text_content"
	RecordKeyChartContent      = "chart_content"
	RecordKeyChartIsMeaningful = "chart_is_meaningful"

	RecordKeyUseGmail        = "use_gmail"
	RecordKeyGmailData       = "gmail_data"
	RecordKeyUseTest         = "use_test"
	RecordKeyESKeyWords      = "es_key_words"
	RecordKeyUseSlack        = "use_slack"
	RecordKeySlackData       = "slack_data"
	RecordKeyUseConfluence   = "use_confluence"
	RecordKeyConfluenceData  = "confluence_data"
	RecordKeyUseGoolgeDrive  = "use_google_drive"
	RecordKeyGoogleDriveData = "google_drive_data"
)

const (
	NoDataCaseNameNoExcelData = "data_case_no_excel_data"
)

var NoDataCaseNameTextMap = map[string]string{
	NoDataCaseNameNoExcelData: "kechat_mysql_chat_no_rows",
}

const (
	TrueStr  = "true"
	FalseStr = "false"
)
