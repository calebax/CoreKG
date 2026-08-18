package employee

// // QueryPositionList 查询职位列表，待修改
// func QueryPositionList(opt apiobj.PageQuery, positionList *admintype.QueryPositionListResponse) error {
// 	query := dbutil.Ops().Table(opdb.TableNameOpPosition).Where("deleted_at is null")
// 	// Joins("INNER JOIN rbac_rel_prole_binding ON rbac_rel_prole_binding.user_id = core_user.id")

// 	for _, filter := range opt.Filters {
// 		switch filter.Field {
// 		case "name":
// 			query = query.Where("op_position.name LIKE ?", "%"+filter.Value[0]+"%")
// 		default:
// 			logs.Warnf("[opaccount][QueryPositionList] invalid filter field: %s", filter.Field)
// 			return fmt.Errorf("invalid filter field: %s", filter.Field)
// 		}
// 	}

// 	if err := query.Count(&positionList.Total).Error; err != nil {
// 		return err
// 	}
// 	if positionList.Total == 0 {
// 		return nil
// 	}

// 	if len(opt.OrderBy) > 0 {
// 		query = query.Order(strings.Join(opt.OrderBy, ","))
// 	}

// 	query = query.Offset(opt.Offset)
// 	if !opt.ListAll && opt.Limit > 0 {
// 		query = query.Limit(opt.Limit)
// 	}

// 	err := query.Find(&positionList.Data).Error
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
