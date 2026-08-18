package data

type Dianlu []struct {
	Chapter  string `json:"chapter"`
	FileID   int    `json:"file_id"`
	Sections []struct {
		Title  string `json:"title"`
		FileID int    `json:"file_id"`
		Page   int    `json:"page"`
	} `json:"sections"`
}

type Weixiu []struct {
	ID       string `json:"id"`
	FileID   int    `json:"file_id"`
	Title    string `json:"title"`
	Sections []struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		FileID      int    `json:"file_id"`
		Subsections []struct {
			ID     string `json:"id"`
			Title  string `json:"title"`
			Page   int    `json:"page"`
			FileID int    `json:"file_id"`
		} `json:"subsections"`
	}
	Subsections []struct {
		ID     string `json:"id"`
		Title  string `json:"title"`
		Page   int    `json:"page"`
		FileID int    `json:"file_id"`
	} `json:"subsections"`
}

type Zhenduan []struct {
	System  string `json:"system"`
	FileID  int    `json:"file_id"`
	Entries []struct {
		Text   string   `json:"text"`
		FileID int      `json:"file_id"`
		Page   int      `json:"page"`
		DTCs   []string `json:"dtcs"`
	} `json:"entries"`
}

type BenTiEntity struct {
	Nodes []struct {
		ID         string                 `json:"id"`
		Lables     []string               `json:"labels"`
		Properties map[string]interface{} `json:"properties"`
	} `json:"nodes"`
	Edges []struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Src  string `json:"startNode"`
		Dst  string `json:"endNode"`
	} `json:"relations"`
}
