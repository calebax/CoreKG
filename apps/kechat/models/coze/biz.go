package coze

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CozeResponse struct {
	Code       int         `json:"code"`
	Msg        string      `json:"msg"`
	Data       interface{} `json:"data"`
	SessionKey string      `json:"session_key"`
}

type GetListResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		BotSpaceList []struct {
			SpaceID string `json:"id"`
		} `json:"bot_space_list"`
	} `json:"data"`
}

type CreatePluginResponse struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	PluginID string `json:"plugin_id"`
}

type CreateCozeAPIResponse struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	APIID string `json:"api_id"`
}

type UpdateCozeAPIResponse struct {
	Code         int    `json:"code"`
	Msg          string `json:"msg"`
	EeditVersion int    `json:"edit_version"`
}

type DebugCozeAPIResponse struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Success bool   `json:"success"`
}

type publishPluginResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	VersionTs string `json:"version_ts"`
}

type CreateKnowledgeAPIResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	DatasetID string `json:"dataset_id"`
}

type GetCozeSourceTypeResponse struct {
	Code         int    `json:"code"`
	Msg          string `json:"msg"`
	ResourceList []struct {
		ResID      string `json:"res_id"`
		SourceType string `json:"source_type"`
	} `json:"resource_list"`
}

type CreateCozeWorkflowAPIResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		WorkflowID string `json:"workflow_id"`
	} `json:"data"`
}

type DeleteCozeWorkflowResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Status int `json:"status"`
	} `json:"data"`
}

type PublicAgentExternalTokenData struct {
	AgentID    string `json:"agent_id"`
	CozeApiKey string `json:"auth_token"`
	ExpireAt   int64  `json:"expire_at"`
}
