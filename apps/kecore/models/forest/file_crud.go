package forest

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/types"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetForestFileByID 查询知识森林file文件
func GetForestFileByID(fileID uint) (*foresttype.KnownowForestFile, error) {
	out := &foresttype.KnownowForestFile{}
	if err := dbutil.Knownow().
		First(out, fileID).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// GetForestFileNameByIDs 根据fileIDs查询知识森林文件名列表
func GetForestFileNameByIDs(fileIDs []uint) ([]string, error) {
	var out []string
	if err := dbutil.Knownow().
		Table(foresttype.TableNameKnownowForestFile).
		Where("id in (?)", fileIDs).
		Pluck("name", &out).
		Error; err != nil {
		return nil, err
	}
	return out, nil
}

// GetForestFileByIDs 查询知识森林file文件们
func GetForestFileByIDs(fileIDs []uint) ([]*foresttype.KnownowForestFile, error) {
	var out []*foresttype.KnownowForestFile
	if err := dbutil.Knownow().
		Where("id IN ?", fileIDs).
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func IsDirPath(p string) bool {
	return strings.HasSuffix(p, "/") && len(p) > 1
}

// GetForestDirByParentAndName 按名称查询知识森林父级目录下的file
func GetForestDirByParentAndName(forestID, parentID uint, name string) (*foresttype.KnownowForestFile, error) {
	out := &foresttype.KnownowForestFile{}
	err := dbutil.Knownow().
		Where("forest_id = ? "+
			"AND parent_id = ? "+
			//"AND is_dir = 1 "+
			"AND name = ?", forestID, parentID, name).
		Where("status = ?", foresttype.FileStatusNormal).
		First(out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetForestFileByName 查询知识森林file文件
func GetForestFileByName(forestID uint, parentID uint, name string) (*foresttype.KnownowForestFile, error) {
	out := &foresttype.KnownowForestFile{}
	if err := dbutil.Knownow().
		Where("forest_id = ?", forestID).
		Where("parent_id = ?", parentID).
		Where("name = ?", name).
		Where("status = ?", foresttype.FileStatusNormal).
		First(out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// IsExistForestFile 该目录下文件名是否已经存在 parentID=0是根目录
func IsExistForestFile(forestID, parentID uint, name string) (bool, error) {
	_, err := GetForestFileByName(forestID, parentID, name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ListAllForestFile 查询知识森林文件列表
func ListAllForestFile(forestID uint) ([]*foresttype.KnownowForestFile, error) {
	out := make([]*foresttype.KnownowForestFile, 0)
	err := dbutil.Knownow().Table(foresttype.TableNameKnownowForestFile).
		Where(foresttype.TableNameKnownowForestFile+".`forest_id` = ?", forestID).
		Where("deleted_at is null").
		Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func ListForestFile(fileids []uint) ([]*foresttype.KnownowForestFile, error) {
	out := make([]*foresttype.KnownowForestFile, 0)
	err := dbutil.Knownow().Table(foresttype.TableNameKnownowForestFile).
		Where("id IN (?)", fileids).
		Where("deleted_at is null").
		Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func ListForestEnableFile(fileIDs []uint) ([]*foresttype.KnownowForestFile, error) {
	out := make([]*foresttype.KnownowForestFile, 0)
	err := dbutil.Knownow().Table(foresttype.TableNameKnownowForestFile).
		Where("id IN (?)", fileIDs).
		Where("enable = ?", 1).
		Where("deleted_at IS NULL").
		Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetForestFilesByMDStatus 获取知识森林下的所有已解析出md文件的文件列表
func GetForestFilesByMDStatus(forestID uint) ([]foresttype.KnownowForestFile, error) {
	files := make([]foresttype.KnownowForestFile, 0)
	if err := dbutil.Knownow().
		Where("forest_id = ?", forestID).
		Where("is_dir = ?", types.False).
		Where("parse_status = ?", foresttype.TaskStatusSuccess).
		Find(&files).
		Error; err != nil {
		return nil, err
	}
	return files, nil
}

// GetFileCount 获取用户文件数量
func GetFileCount(company_id uint) (int64, error) {
	var count int64
	err := dbutil.Knownow().Table(foresttype.TableNameKnownowForestFile).
		Where(foresttype.TableNameKnownowForestFile+".`company_id` = ?", company_id).
		Where("deleted_at is null").
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetForestCount 获取用户知识森林数量
func GetForestCount(company_id uint) (int64, error) {
	var count int64
	err := dbutil.Knownow().Table(foresttype.TableNameKnownowForest).
		Where(foresttype.TableNameKnownowForest+".`company_id` = ?", company_id).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetRecentlyForest 根据更新时间拿最近的4个知识森林
func GetRecentlyForest(uin uint) ([]*foresttype.KnownowForest, error) {
	var forests []*foresttype.KnownowForest
	err := dbutil.Knownow().Table(foresttype.TableNameKnownowForest).
		Where(foresttype.TableNameKnownowForest+".`uin` = ?", uin).
		Order("updated_at DESC").
		Limit(4).
		Find(&forests).Error
	if err != nil {
		return nil, err
	}
	return forests, nil
}

var (
	ValidParsAbleExtSlice = []string{
		".png", ".jpg", ".jpeg",
		".txt", ".md",
		".mp4",
		".pdf",
		".ppt", ".pptx",
		".doc", ".docx",
		".ofd",
	}
)

const (
	MB200 = 200 << 20
	MB50  = 50 << 20
)

// PreViewAble check if the file could be previewed
func PreViewAble(f *foresttype.KnownowForestFile) bool {
	//check parse valid
	for _, ext := range ValidParsAbleExtSlice {
		if ext == f.Ext {
			return true
		}
	}
	return false
}

func ParsAble(f *foresttype.KnownowForestFile) bool {
	if !PreViewAble(f) {
		return false
	} else if f.Ext != ".mp4" && PreViewAble(f) && f.Size > MB50 {
		return false
	} else if f.Ext == ".mp4" && f.Size > MB200 {
		return false
	}
	return true
}

// GetPathString will return a path string slice about f
func GetPathString(f *foresttype.KnownowForestFile) ([]uint, []string, error) {
	var paths []string
	var parentIds []uint
	ipMap := make(map[uint64]string, 1)
	pids := strings.Split(f.ParentIDs, "/")
	for _, v := range pids {
		if v == "" {
			continue
		}
		parentId, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, nil, err
		}
		if p, ok := ipMap[parentId]; ok {
			paths = append(paths, p)
			parentIds = append(parentIds, uint(parentId))
			continue
		}

		pf, err := GetForestFileByID(uint(parentId))
		if err != nil {
			return nil, nil, err
		}
		parentIds = append(parentIds, uint(parentId))
		paths = append(paths, pf.Name)

		ipMap[parentId] = pf.Name
	}
	return parentIds, paths, nil
}

var (
	ErrHasRunningTask = errors.New("this forest has a running")
)

func DeleteForestStatusCheck(ctx context.Context, forestID uint) error {
	// check if any file under this forest has a running task
	var count int64
	if err := dbutil.Knownow().Table(foresttype.TableNameKnownowForestFile).
		Where("deleted_at IS NULL").
		Where("forest_id = ?", forestID).
		Where("(parse_status = ? OR mindmap_status = ? OR analysis_status = ? OR knowledge_status = ?)", foresttype.TaskStatusRunning, foresttype.TaskStatusRunning, foresttype.TaskStatusRunning, foresttype.TaskStatusRunning).
		Count(&count).
		Error; err != nil {
		logs.ErrorContextf(ctx, "DeleteForestStatusCheck: check forest status err: %v", err)
		return err
	}
	if count > 0 {
		logs.WarnContextf(ctx, "DeleteForestStatusCheck: forest %v has %v no done files", forestID, count)
		return ErrHasRunningTask
	}
	return nil
}

func DeleteFilesStatusCheck(ctx context.Context, fIDs []uint) error {
	// check if any file under this forest has a running task
	var count int64
	if err := dbutil.Knownow().Table(foresttype.TableNameKnownowForestFile).
		Where("deleted_at IS NULL").
		Where("id in (?)", fIDs).
		Where("is_dir = -1").
		Where("(parse_status = ? OR mindmap_status = ? OR analysis_status = ? OR knowledge_status = ?)", foresttype.TaskStatusRunning, foresttype.TaskStatusRunning, foresttype.TaskStatusRunning, foresttype.TaskStatusRunning).
		Count(&count).
		Error; err != nil {
		logs.ErrorContextf(ctx, "DeleteFilesStatusCheck: check files status err: %v", err)
		return err
	}
	if count > 0 {
		logs.WarnContextf(ctx, "DeleteFilesStatusCheck: Files %v has %v no done files", fIDs, count)
		return ErrHasRunningTask
	}
	return nil
}

// GetDirFileByIDs 获取文件夹
func GetDirsFileByIDs(ctx context.Context, ids []uint) ([]*foresttype.KnownowForestFile, error) {
	if len(ids) == 0 {
		return []*foresttype.KnownowForestFile{}, nil
	}
	dirs, err := GetForestFileByIDs(ids)
	if err != nil {
		logs.ErrorContextf(ctx, "GetDirsFileByIDs GetForestFileByIDs err: %v", err)
		return nil, err
	}

	filelist := []*foresttype.KnownowForestFile{}
	sql := dbutil.Knownow().Table(foresttype.TableNameKnownowForestFile).
		Where("deleted_at IS NULL").
		Where("parent_ids LIKE ?", fmt.Sprintf("%s%d", dirs[0].ParentIDs, dirs[0].ID)+"/%")
	for _, f := range dirs {
		sql = sql.Or("parent_ids LIKE ?", fmt.Sprintf("%s%d", f.ParentIDs, f.ID)+"/%")
	}
	err = sql.Find(&filelist).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find children files or directories: %v", err)
	}
	filelist = append(filelist, dirs...)
	return filelist, nil
}

// GetDirFiles will get all files(include itself and its son_file/son_dir)
func GetDirFiles(ctx context.Context, fileID uint) (res []*foresttype.KnownowForestFile, err error) {
	var (
		target *foresttype.KnownowForestFile
	)
	//get target dir
	if err = dbutil.Knownow().First(&target, fileID).Error; err != nil {
		return nil, err
	}
	//if target file is not dir
	if target.IsDir == -1 {
		res = append(res, target)
		return res, nil
	}

	if result := dbutil.Knownow().
		WithContext(ctx).
		Where("parent_ids like ?", fmt.Sprintf("%v%v/%%", target.ParentIDs, fileID)).
		Find(&res); result.Error != nil {
		return nil, err
	}
	//add target dir
	res = append([]*foresttype.KnownowForestFile{target}, res...)

	return
}

// GetDirsFiles will get all files (include itself and its son_file/son_dir) for multiple given file IDs.
func GetDirsFiles(ctx context.Context, fileIDs []uint) ([]*foresttype.KnownowForestFile, error) {
	if len(fileIDs) == 0 {
		return []*foresttype.KnownowForestFile{}, nil
	}

	var (
		targets, res []*foresttype.KnownowForestFile
		err          error
	)
	// 1. 获取所有目标文件夹的信息
	if err = dbutil.Knownow().WithContext(ctx).Where("id IN ?", fileIDs).Find(&targets).Error; err != nil {
		return nil, err
	}

	var (
		conditions  []clause.Expression
		nonDirFiles []*foresttype.KnownowForestFile
	)

	for _, target := range targets {
		if target.IsDir == -1 {
			nonDirFiles = append(nonDirFiles, target)
			continue
		}

		conditions = append(conditions, clause.Eq{Column: "id", Value: target.ID})

		conditions = append(conditions, clause.Like{Column: "parent_ids", Value: fmt.Sprintf("%v%v/", target.ParentIDs, target.ID) + "%"})
	}

	if len(conditions) == 0 {
		return nonDirFiles, nil
	}

	if err = dbutil.Knownow().WithContext(ctx).Clauses(clause.Where{
		Exprs: []clause.Expression{
			clause.Or(conditions...),
		},
	}).Find(&res).Error; err != nil {
		return nil, err
	}

	res = append(res, nonDirFiles...)

	return res, nil

}

func GetFilesSizeByCompanyID(ctx context.Context, companyID uint) (int64, error) {
	var res *int64
	if err := dbutil.Knownow().WithContext(ctx).
		Table(foresttype.TableNameKnownowForestFile).
		Where("company_id = ?", companyID).
		Where("size is not null").
		Select("SUM(size) as disk_storage").
		Scan(&res).
		Error; err != nil {
		return 0, err
	}
	if res == nil {
		return 0, nil
	}
	return *res, nil
}

type ListParseHistoryResponse struct {
	apiobj.QueryResponse
	Data []*File
}

func QueryParseHistory(ctx context.Context, opt apiobj.PageQuery, resp *ListParseHistoryResponse) error {
	query := dbutil.Knownow().Table(foresttype.TableNameKnownowForestFile+" f ").
		Unscoped().
		Select("f.*,"+
			"frs.forest_type as forest_type,"+
			"frs.data_source_type as data_source_type,"+
			"frs.data_source_subtype as data_source_subtype").
		Where("f.deleted_at IS NULL").
		Where("f.company_id = ?", opt.CompanyID).
		Where("f.is_dir = ?", -1).
		Where("f.parse_status != ?", foresttype.TaskStatusUnsupported).
		Where("f.status = ?", foresttype.FileStatusNormal).
		Joins("LEFT JOIN " + foresttype.TableNameKnownowForest + " frs ON f.forest_id = frs.id AND frs.deleted_at IS NULL")

	for _, filter := range opt.Filters {
		switch filter.Field {
		case "recent_days":
			query = query.Where(fmt.Sprintf("f.created_at >= DATE_SUB(NOW(), INTERVAL %v DAY)", filter.Value[0]))
		default:
			logs.ErrorContextf(ctx, "QueryParseHistory invalid filtering field: %s", filter.Field)
			return fmt.Errorf("invalid filtering field: %s", filter.Field)
		}
	}

	if err := query.Count(&resp.Total).Error; err != nil {
		logs.ErrorContextf(ctx, "QueryParseHistory calculate count failed: %v", err)
		return err
	}
	if resp.Total == 0 {
		return nil
	}
	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	}
	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	}
	if err := query.Find(&resp.Data).Error; err != nil {
		logs.ErrorContextf(ctx, "QueryParseHistory get files record failed: %v", err)
		return err
	}
	resp.Limit = opt.Limit
	resp.Offset = opt.Offset
	for _, v := range resp.Data {
		status, pro := v.KnownowForestFile.CalculateProgress()
		v.FileStatus = status
		v.FileProgress = pro
	}
	return nil
}

type Forest struct {
	foresttype.KnownowForest
	Files []*FileItem `json:"files"`
}

type FileItem struct {
	// ID is forest file id
	ID uint `json:"id"`
	// Name is forest file name
	Name string `json:"name"`
	// PublicUrl is forest file public url
	PublicUrl string `json:"public_url"`
	// FileID is not forest file id, it's core_upload_files' id
	FileID uint `json:"file_id"`
	//	ForestID is forest id
	ForestID uint `json:"forest_id"`
}

func GetForestFilePublicUrls(ctx context.Context, forestIDs []uint) (res []*Forest, err error) {
	if len(forestIDs) == 0 {
		logs.WarnContextf(ctx, "[GetForestFilePublicUrls] forestIDs is empty")
		return nil, nil
	}
	res = make([]*Forest, 0, len(forestIDs))
	if err = dbutil.Knownow().WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("id IN ?", forestIDs).
		Find(&res).Error; err != nil {
		logs.ErrorContextf(ctx, "[GetForestFilePublicUrls] get forests(%v) failed: %v", forestIDs, err)
		return nil, err
	}

	frsMap := utils.ToMap(res, func(f *Forest) uint {
		return f.ID
	})

	fs := make([]FileItem, 0, len(res))
	if err = dbutil.Knownow().WithContext(ctx).
		Table(foresttype.TableNameKnownowForestFile).
		Select("id,file_id,name,forest_id").
		Where("deleted_at IS NULL").
		Where("forest_id IN ?", forestIDs).
		Where("is_dir = ?", types.False).
		Find(&fs).
		Error; err != nil {
		logs.ErrorContextf(ctx, "[GetForestFilePublicUrls] get files(forest_ids:%v) failed: %v", forestIDs, err)
		return nil, err
	}

	fIDs := make([]uint, 0, len(fs))

	for _, v := range fs {
		fIDs = append(fIDs, v.FileID)
	}

	// 2. 查询文件的公共 URL
	type IDPubLicUrl struct {
		ID        uint   `json:"id"`
		PublicUrl string `json:"public_url"`
	}

	ufs := make([]*IDPubLicUrl, 0, len(fIDs))

	if err := dbutil.Core().WithContext(ctx).
		Table("core_upload_files").
		Select("id,public_url").
		Where("id IN ?", fIDs).
		Find(&ufs).Error; err != nil {
		logs.ErrorContextf(ctx, "[GetForestFilePublicUrls] get core_upload_files(ids:%v) failed: %v", fIDs, err)
		return nil, err
	}

	for _, v := range fs {
		v.PublicUrl = utils.ToMap(ufs, func(u *IDPubLicUrl) uint {
			return u.ID
		})[v.FileID].PublicUrl
		frsMap[v.ForestID].Files = append(frsMap[v.ForestID].Files, &v)
	}

	return res, nil
}

func GetForestFileByFileID(ctx context.Context, coreFileID uint) (*foresttype.KnownowForestFile, error) {
	var f foresttype.KnownowForestFile
	if err := dbutil.Knownow().WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("file_id = ?", coreFileID).
		First(&f).
		Error; err != nil {
		logs.ErrorContextf(ctx, "GetForestFileByFileID(coreFileID:%v) faild err: %v", coreFileID, err)
		return nil, err
	}
	return &f, nil
}
