package nbgraph

import (
	"encoding/json"
	"fmt"
	"os"

	nebula "github.com/vesoft-inc/nebula-go/v3"
)

func SearchAll() error {
	cli, err := NewNebulaCLI()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	defer cli.Release()
	resp, err := cli.ExecuteJson("USE know_6_77; MATCH (n) OPTIONAL MATCH (n)-[r]->(m) RETURN n, r, m;")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	// fmt.Println("Execute 结果", string(resp))
	var data interface{}
	err = json.Unmarshal(resp, &data)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return err
	}

	// 将数据转换为 JSON 格式
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return err
	}

	// 将 JSON 数据写入文件
	file, err := os.Create("output.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}

	fmt.Println("JSON data has been written to output.json")
	AA(resp)
	// record, _ := resp.GetRowValuesByIndex(0)
	// record.GetValueByColName("description")
	// fmt.Printf("The first row elements: %s\n", record.String())
	// 检查查询结果
	// if resp.GetErrorCode() != nebula.ErrorCode_SUCCEEDED {
	// 	fmt.Println("Error: ", resp.GetErrorMsg())
	// }

	// 处理查询结果
	// resultSet := resp.GetResultSet()
	// for _, row := range resultSet.GetRows() {
	// 	for _, col := range row.GetColumns() {
	// 		switch col.GetType().GetType() {
	// 		case nebula.ColumnType_STRING:
	// 			fmt.Println("Name: ", string(col.GetString()))
	// 		default:
	// 			fmt.Println("Unsupported type")
	// 		}
	// 	}
	// }
	return nil
}

// 解析查询结果
func parseResults(resp *nebula.ResultSet) []map[string]interface{} {
	var results []map[string]interface{}
	for _, row := range resp.GetRows() {
		result := make(map[string]interface{})
		for i, colName := range resp.GetColNames() {
			result[colName] = row.Values[i]
		}
		results = append(results, result)
	}
	return results
}
