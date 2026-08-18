package mysql

import (
	"bufio"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// ExecuteSQLFile 执行sql文件
func ExecuteSQLFile(gCtx context.Context, db *gorm.DB, fpath string) error {
	filename := filepath.Base(fpath)
	ctx := logs.WithContextFields(gCtx, "filename", filename)

	rcd := &InitExecRecourd{
		FileName:   filepath.Base(fpath),
		StartTime:  time.Now(),
		ExecStatus: ExecStatusFailed,
	}
	defer func() {
		rcd.EndTime = time.Now()
		rcd.ExecTime = rcd.EndTime.Sub(rcd.StartTime).Seconds()
		if err := db.Save(rcd).Error; err != nil {
			logs.ErrorContextf(ctx, "save init exec record error: %v", err)
		}
	}()

	logs.InfoContextf(ctx, "start execute sql file: %s", fpath)
	file, err := os.Open(fpath)
	if err != nil {
		logs.ErrorContextf(ctx, "open file %s failed, %s", fpath, err)
		return err
	}
	defer file.Close()

	runStartAt := 0
	{
		last, err := LastExecRecourd(db, filename)
		if err != nil {
			logs.ErrorContextf(ctx, "get last exec record error: %v", err)
			return err
		}
		if last != nil {
			runStartAt = last.FailLineAt
		}
	}

	lineNumber, err := ExecuteSQLReader(ctx, db, file, runStartAt)
	if err != nil {
		rcd.FailLineAt = lineNumber
		rcd.ErrorMessage = err.Error()
		logs.ErrorContextf(ctx, "execute sql file %s failed, %s", fpath, err)
		return err
	}
	logs.InfoContextf(ctx, "end execute sql file: %s", fpath)
	rcd.ExecStatus = ExecStatusSuccess
	return nil
}

// ExecuteSQLReader 执行sql文件
func ExecuteSQLReader(ctx context.Context, db *gorm.DB, r io.Reader, runAt int) (int, error) {
	statements, err := parseSQLReader(r)
	if err != nil {
		logs.ErrorContextf(ctx, "parse sql failed, %s", err)
		return 0, err
	}

	for _, stmt := range statements {
		if stmt.Number < runAt {
			logs.DebugContextf(ctx, "skip sql(%v): %s", stmt.Number, stmt.Line)
			continue
		}
		logs.DebugContextf(ctx, "execute sql(%v): %s", stmt.Number, stmt.Line)
		if err := db.Exec(stmt.Line).Error; err != nil {
			logs.ErrorContextf(ctx, "execute sql(%v) %s failed, %s", stmt.Number, stmt.Line, err)
			return stmt.Number, err
		}
	}

	return 0, nil
}

// SQLLine sql语句
type SQLLine struct {
	Line   string
	Number int
}

func parseSQLReader(file io.Reader) ([]SQLLine, error) {
	statements := []SQLLine{}
	var currentStatement strings.Builder
	scanner := bufio.NewScanner(file)

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "#") {
			continue
		}

		// 跳过/* */注释
		if strings.HasPrefix(line, "/*") && strings.HasSuffix(line, "*/") {
			continue
		}

		currentStatement.WriteString(line)
		currentStatement.WriteString(" ")

		// 检查是否为语句结束
		if strings.HasSuffix(line, ";") {
			stmt := strings.TrimSpace(currentStatement.String())
			if stmt != "" {
				statements = append(statements, SQLLine{stmt, lineNumber})
			}
			currentStatement.Reset()
		}
	}

	// 处理最后一个语句（如果没有分号结尾）
	if currentStatement.Len() > 0 {
		stmt := strings.TrimSpace(currentStatement.String())
		if stmt != "" {
			statements = append(statements, SQLLine{stmt, lineNumber})
		}
	}

	return statements, scanner.Err()
}
