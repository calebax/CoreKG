package nbgraph

import (
	"encoding/json"
	"fmt"
	"os"
)

type Input struct {
	Errors []struct {
		Code int `json:"code"`
	} `json:"errors"`
	Results []struct {
		Columns []string `json:"columns"`
		Data    []struct {
			Meta []struct {
				ID   interface{} `json:"id"` // To handle null values
				Type string      `json:"type"`
			} `json:"meta"`
			Row []struct {
				Clusters    string `json:"know_node.clusters"`
				Description string `json:"know_node.description"`
				SourceID    string `json:"know_node.source_id"`
				Type        string `json:"know_node.type"`
			} `json:"row"`
		} `json:"data"`
		Errors struct {
			Code int `json:"code"`
		} `json:"errors"`
		LatencyInUs int    `json:"latencyInUs"`
		SpaceName   string `json:"spaceName"`
	} `json:"results"`
}

type Output struct {
	Nodes []struct {
		ID   string `json:"id"`
		Data struct {
			Cluster     string `json:"cluster"`
			Description string `json:"description"`
			SourceID    string `json:"source_id"`
			Type        string `json:"type"`
		} `json:"data"`
	} `json:"nodes"`
	Edges []struct {
		Source string `json:"source"`
		Target string `json:"target"`
	} `json:"edges"`
}

func convert(input Input) Output {
	var output Output

	nodeIDMap := make(map[string]struct{}) // To avoid duplicate nodes

	// Process the nodes
	for _, result := range input.Results {
		for _, data := range result.Data {
			metaID := ""
			if data.Meta[0].ID != nil {
				metaID = fmt.Sprintf("%v", data.Meta[0].ID) // Handle dynamic types
			}

			// Avoid adding empty or duplicate nodes
			if metaID == "" || nodeIDMap[metaID] != (struct{}{}) {
				continue
			}

			nodeIDMap[metaID] = struct{}{} // Mark this ID as processed

			// Add to nodes
			node := struct {
				ID   string `json:"id"`
				Data struct {
					Cluster     string `json:"cluster"`
					Description string `json:"description"`
					SourceID    string `json:"source_id"`
					Type        string `json:"type"`
				} `json:"data"`
			}{
				ID: metaID,
			}

			if len(data.Row) > 0 {
				node.Data.Cluster = data.Row[0].Clusters
				node.Data.Description = data.Row[0].Description
				node.Data.SourceID = data.Row[0].SourceID
				node.Data.Type = data.Row[0].Type
			}

			output.Nodes = append(output.Nodes, node)
		}
	}

	// Process the edges
	for _, result := range input.Results {
		for _, data := range result.Data {
			if len(data.Meta) > 2 && data.Meta[1].ID != nil && data.Meta[2].ID != nil {
				edge := struct {
					Source string `json:"source"`
					Target string `json:"target"`
				}{
					Source: fmt.Sprintf("%v", data.Meta[0].ID),
					Target: fmt.Sprintf("%v", data.Meta[2].ID),
				}
				output.Edges = append(output.Edges, edge)
			}
		}
	}

	return output
}

func AA(a []byte) {
	var input Input
	err := json.Unmarshal(a, &input)
	if err != nil {
		fmt.Println("Error parsing input JSON:", err)
		return
	}
	output := convert(input)
	outputJSON, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Println("Error creating output JSON:", err)
		return
	}
	file, err := os.Create("output1111.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.Write(outputJSON)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}
