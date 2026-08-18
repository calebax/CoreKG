package nbgraph

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ygpkg/yg-go/logs"
)

// DeleteTag 删除节点类型，删除文件时使用
func DeleteTag(ctx context.Context, uin, forestID, fileID uint) error {
	cli, err := NewNebulaCLI()
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return err
	}
	defer cli.Release()
	_, err = cli.ExecuteAndCheck("USE " + GetNebulaSpace(uin, forestID) +
		fmt.Sprintf(";DROP TAG IF EXISTS doc_%v;", fileID))
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return err
	}
	return nil
}

// DeleteEdge 删除边，删除文件时使用
func DeleteEdge(ctx context.Context, uin, forestID, fileID uint) error {
	cli, err := NewNebulaCLI()
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return err
	}
	defer cli.Release()
	_, err = cli.ExecuteAndCheck("USE " + GetNebulaSpace(uin, forestID) +
		fmt.Sprintf(";DROP EDGE IF EXISTS doc_%v_edge;", fileID))
	if err != nil {
		logs.ErrorContextf(ctx, err.Error())
		return err
	}
	return nil
}

// DeleteForest will check if this forest has any file that has running substatus
func DeleteForest(ctx context.Context, forestID uint, space string) error {
	//do nebula's delete action
	cli, err := NewNebulaCLI()
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteForest: %v", err)
		return err
	}
	defer cli.Release()
	//delete edges
	deleteEdgeNql := fmt.Sprintf("USE %v; "+
		"LOOKUP ON relationships WHERE relationships.forest_id == %v "+
		"YIELD src(edge) AS src, dst(edge) AS dst, rank(edge) AS rank | DELETE EDGE relationships $-.src -> $-.dst @ $-.rank",
		space, forestID)
	if _, err = cli.ExecuteAndCheck(deleteEdgeNql); err != nil {
		logs.ErrorContextf(ctx, "DeleteForest: delete edges err: %v", err)
		return err
	}

	//delete nodes
	deleteNodeNql := fmt.Sprintf("USE %v; "+
		"LOOKUP ON entities WHERE entities.forest_id == %v "+
		"YIELD id(vertex) AS vid | DELETE VERTEX $-.vid WITH EDGE;", space, forestID)
	if _, err = cli.ExecuteAndCheck(deleteNodeNql); err != nil {
		logs.ErrorContextf(ctx, "DeleteForest: delete nodes err: %v", err)
		return err
	}

	return nil
}

// DeleteFiles will delete all reference data about a slice of files.
// This is the primary, more powerful function for handling batch deletions.
func DeleteFiles(ctx context.Context, forestID uint, fileIDs []uint, space string) error {
	var fileIDsToDelete []string
	for _, v := range fileIDs {
		fileIDsToDelete = append(fileIDsToDelete, strconv.FormatUint(uint64(v), 10))
	}
	if len(fileIDsToDelete) == 0 {
		logs.InfoContextf(ctx, "DeleteFiles: received an empty slice of file IDs, nothing to do.")
		return nil
	}

	cli, err := NewNebulaCLI()
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteFiles: failed to get nebula client: %v", err)
		return err
	}
	defer cli.Release()

	// --- STAGE 1: READ ----
	// 捞取 forest 和 uin 下的所有相关节点
	lookupNql := fmt.Sprintf("USE %s; LOOKUP ON entities WHERE entities.forest_id == %d "+
		"YIELD id(vertex) AS vid, properties(vertex).file_id AS file_id",
		space, forestID)

	resp, err := cli.ExecuteAndCheck(lookupNql)
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteFiles: failed to lookup nodes: %v", err)
		return err
	}
	if resp.IsEmpty() {
		return nil
	}

	// --- STAGE 2: MODIFY ----

	// **将待删除的 file IDs 放入一个 set (map) 中**
	filesToDeleteSet := make(map[string]struct{}, len(fileIDsToDelete))
	for _, id := range fileIDsToDelete {
		filesToDeleteSet[id] = struct{}{}
	}

	colNames := resp.GetColNames()
	nameIndexMap := make(map[string]int, len(colNames))
	for i, name := range colNames {
		nameIndexMap[name] = i
	}
	vidIndex := nameIndexMap["vid"]
	fileIDIndex := nameIndexMap["file_id"]

	var nodesToDelete []string
	nodesToUpdate := make(map[string]string)

	for _, row := range resp.GetRows() {
		vid := string(row.Values[vidIndex].SVal)
		currentFileID := string(row.Values[fileIDIndex].SVal)

		currentIDs := strings.Split(currentFileID, "&&&")
		remainingIDs := make([]string, 0, len(currentIDs))

		// 筛选出不应被删除的ID
		for _, id := range currentIDs {
			if _, found := filesToDeleteSet[id]; !found {
				remainingIDs = append(remainingIDs, id)
			}
		}

		// 根据剩余ID的数量来决定操作
		if len(remainingIDs) == len(currentIDs) {
			// 如果没有任何ID被移除，说明此节点与本次删除无关，跳过
			continue
		}

		if len(remainingIDs) == 0 {
			// 如果所有关联的ID都被移除了，那么此节点需要被删除
			nodesToDelete = append(nodesToDelete, vid)
		} else {
			// 如果还有剩余的ID，那么此节点需要被更新
			nodesToUpdate[vid] = strings.Join(remainingIDs, "&&&")
		}
	}

	// --- STAGE 3: WRITE ----

	// Execute Deletions
	if len(nodesToDelete) > 0 {
		var quotedVidsBuilder strings.Builder
		for i, vid := range nodesToDelete {
			if i > 0 {
				quotedVidsBuilder.WriteString(", ")
			}
			fmt.Fprintf(&quotedVidsBuilder, "\"%s\"", escapeVidForNQL(vid))
		}

		deleteNql := fmt.Sprintf("USE %s; DELETE VERTEX %s WITH EDGE;", space, quotedVidsBuilder.String())
		if _, err := cli.ExecuteAndCheck(deleteNql); err != nil {
			logs.ErrorContextf(ctx, "DeleteFile: failed to delete nodes: %v,nql: %v", err, deleteNql)
			return err
		}
		logs.InfoContextf(ctx, "DeleteFile: successfully deleted %d nodes.", len(nodesToDelete))
	}

	// Execute Updates
	if len(nodesToUpdate) > 0 {
		var multiStatementBuilder strings.Builder
		fmt.Fprintf(&multiStatementBuilder, "USE %s;", space)
		for vid, newFileID := range nodesToUpdate {
			fmt.Fprintf(&multiStatementBuilder, " UPDATE VERTEX \"%s\" SET entities.file_id = \"%s\";",
				escapeVidForNQL(vid), escapeVidForNQL(newFileID))
		}
		updateNql := multiStatementBuilder.String()
		if _, err := cli.ExecuteAndCheck(updateNql); err != nil {
			logs.ErrorContextf(ctx, "DeleteFile: failed to update nodes with batch statements: %v ,nql: %v", err, updateNql)
			return err
		}
		logs.InfoContextf(ctx, "DeleteFile: successfully sent batch update for %d nodes.", len(nodesToUpdate))
	}

	return nil
}

// escapeVidForNQL prepares a string to be safely embedded as a VID in an nGQL query.
// It escapes special characters like backslashes, quotes, and newlines.
func escapeVidForNQL(vid string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\", // 1. 先将 \ 替换为 \\
		"\"", "\\\"", // 2. 再将 " 替换为 \"
		"\n", "\\n", // 3. 将换行符 LF 替换为 \n 两个字符
		"\r", "\\r", // 4. 将回车符 CR 替换为 \r 两个字符
		"\t", "\\t", // 5. 将制表符 Tab 替换为 \t 两个字符
	)
	return replacer.Replace(vid)
}

// DeleteFilesEntAndRel will delete all reference data about a slice of files which own entities/relationship type.
// This is the primary, more powerful function for handling batch deletions.
func DeleteFilesEntAndRel(ctx context.Context, forestID uint, fileIDs []uint, space string, cli *NebulaCli) error {
	var fileIDsToDelete []string
	for _, v := range fileIDs {
		fileIDsToDelete = append(fileIDsToDelete, strconv.FormatUint(uint64(v), 10))
	}
	if len(fileIDsToDelete) == 0 {
		logs.InfoContextf(ctx, "DeleteFiles: received an empty slice of file IDs, nothing to do.")
		return nil
	}

	// --- STAGE 1: READ ----
	// 捞取 forest 和 uin 下的所有相关节点
	lookupNql := fmt.Sprintf("USE %s; LOOKUP ON entities WHERE entities.forest_id == %d "+
		"AND entities.type IN [\"entity\"] "+
		"YIELD id(vertex) AS vid, properties(vertex).file_id AS file_id",
		space, forestID)

	resp, err := cli.ExecuteAndCheck(lookupNql)
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteFiles: failed to lookup nodes: %v", err)
		return err
	}
	if resp.IsEmpty() {
		return nil
	}

	// --- STAGE 2: MODIFY ----
	// **将待删除的 file IDs 放入一个 set (map) 中**
	filesToDeleteSet := make(map[string]struct{}, len(fileIDsToDelete))
	for _, id := range fileIDsToDelete {
		filesToDeleteSet[id] = struct{}{}
	}

	colNames := resp.GetColNames()
	nameIndexMap := make(map[string]int, len(colNames))
	for i, name := range colNames {
		nameIndexMap[name] = i
	}
	vidIndex := nameIndexMap["vid"]
	fileIDIndex := nameIndexMap["file_id"]

	var nodesToDelete []string
	nodesToUpdate := make(map[string]string)

	for _, row := range resp.GetRows() {
		vid := string(row.Values[vidIndex].SVal)
		currentFileID := string(row.Values[fileIDIndex].SVal)

		currentIDs := strings.Split(currentFileID, "&&&")
		remainingIDs := make([]string, 0, len(currentIDs))

		// 筛选出不应被删除的ID
		for _, id := range currentIDs {
			if _, found := filesToDeleteSet[id]; !found {
				remainingIDs = append(remainingIDs, id)
			}
		}

		// 根据剩余ID的数量来决定操作
		if len(remainingIDs) == len(currentIDs) {
			// 如果没有任何ID被移除，说明此节点与本次删除无关，跳过
			continue
		}

		if len(remainingIDs) == 0 {
			// 如果所有关联的ID都被移除了，那么此节点需要被删除
			nodesToDelete = append(nodesToDelete, vid)
		} else {
			// 如果还有剩余的ID，那么此节点需要被更新
			nodesToUpdate[vid] = strings.Join(remainingIDs, "&&&")
		}
	}

	// --- STAGE 3: WRITE ----

	// Execute Deletions
	if len(nodesToDelete) > 0 {
		var quotedVidsBuilder strings.Builder
		for i, vid := range nodesToDelete {
			if i > 0 {
				quotedVidsBuilder.WriteString(", ")
			}
			fmt.Fprintf(&quotedVidsBuilder, "\"%s\"", escapeVidForNQL(vid))
		}

		deleteNql := fmt.Sprintf("USE %s; DELETE VERTEX %s WITH EDGE;", space, quotedVidsBuilder.String())
		if _, err := cli.ExecuteAndCheck(deleteNql); err != nil {
			logs.ErrorContextf(ctx, "DeleteFile: failed to delete nodes: %v,nql: %v", err, deleteNql)
			return err
		}
		logs.InfoContextf(ctx, "DeleteFile: successfully deleted %d nodes.", len(nodesToDelete))
	}

	// Execute Updates
	if len(nodesToUpdate) > 0 {
		var multiStatementBuilder strings.Builder
		fmt.Fprintf(&multiStatementBuilder, "USE %s;", space)
		for vid, newFileID := range nodesToUpdate {
			fmt.Fprintf(&multiStatementBuilder, " UPDATE VERTEX \"%s\" SET entities.file_id = \"%s\";",
				escapeVidForNQL(vid), escapeVidForNQL(newFileID))
		}
		updateNql := multiStatementBuilder.String()
		if _, err := cli.ExecuteAndCheck(updateNql); err != nil {
			logs.ErrorContextf(ctx, "DeleteFile: failed to update nodes with batch statements: %v ,nql: %v", err, updateNql)
			return err
		}
		logs.InfoContextf(ctx, "DeleteFile: successfully sent batch update for %d nodes.", len(nodesToUpdate))
	}

	return nil
}
