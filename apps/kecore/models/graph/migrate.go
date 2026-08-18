package graph

import (
	"context"

	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
)

func MigrateTNodeToNode(ctx context.Context) error {
	batch := 5000

	for {
		res := dbutil.Knownow().Exec(`
        UPDATE ke_graph_tag_node t
        JOIN ke_graph_node n ON t.node_id = n.id
        SET 
            t.name = n.name,
            t.company_id = n.company_id,
			t.uin = n.uin
        WHERE t.id IN (
            SELECT id FROM (
                SELECT t2.id
                FROM ke_graph_tag_node t2
                WHERE t2.name = ?
                LIMIT ?
            ) AS tmp
        );
    `, "", batch)

		if res.Error != nil {
			logs.ErrorContext(ctx, "MigrateTNodeToNode err:%v", res.Error)
			return res.Error
		}

		if res.RowsAffected == 0 {
			break
		}

		logs.InfoContextf(ctx, "Updated rows: %d", res.RowsAffected)
	}
	return nil
}
