package global

const (
	CoreTableNamePrefix = "core_"
)

const (
	DBInstanceSystemSettingKey         = "mysql_excel_instance"
	DBInstanceSystemReadonlySettingKey = "mysql_excel_instance_readonly"
)

const (
	DeployModeOnPremise   = "on_premise"
	DeployModeOpenPO      = "openpo"
	DeployModeTencentFree = "tencent_free"
)

const (
	EsIndexKGDefault = "ke_0"
)

const (
	SettingGroupKnowledge = "knowledge"
	SettingGroupCoreKG    = "corekg"

	SettingKeyLlmRoleName                 = "llm_role_name"
	SettingKeySystemLlmAPIKey             = "system_llm_api_key"
	SettingKeyLlmImageParse               = "llm_image_parse"
	SettingKeyAgentEsChat                 = "agent_es_chat"
	SettingKeyAgentExcelQuestionToSQL     = "agent_excel_question_to_sql"
	SettingKeyAgentExcelSQLResultAnalysis = "agent_excel_sql_result_analysis"
	SettingKeyAgentIntentionRecognition   = "agent_intention_recognition"
)

const (
	DefaultLlmRoleName = "言小古"
)
