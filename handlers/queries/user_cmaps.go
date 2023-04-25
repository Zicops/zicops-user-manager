package queries

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/lib/identity"
)

func GetUserCourseMaps(ctx context.Context, userId string, publishTime *int, pageCursor *string, direction *string, pageSize *int, filters *model.CourseMapFilters) (*model.PaginatedCourseMaps, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	if userId != "" {
		emailCreatorID = userId
	}
	var newPage []byte
	//var pageDirection string
	var pageSizeInt int
	if pageCursor != nil && *pageCursor != "" {
		page, err := global.CryptSession.DecryptString(*pageCursor, nil)
		if err != nil {
			return nil, fmt.Errorf("invalid page cursor: %v", err)
		}
		newPage = page
	}
	//key := "GetUserCourseMaps" + emailCreatorID + string(newPage)
	//result, err := redis.GetRedisValue(key)
	//if err == nil {
	//	var outputResponse model.PaginatedCourseMaps
	//	err = json.Unmarshal([]byte(result), &outputResponse)
	//	if err == nil {
	//		return &outputResponse, nil
	//	}
	//}
	if pageSize == nil {
		pageSizeInt = 10
	} else {
		pageSizeInt = *pageSize
	}
	var newCursor string

	qryStr := fmt.Sprintf(`SELECT * from userz.user_course_map where user_id='%s' and created_at <= %d  `, emailCreatorID, *publishTime)
	if filters != nil {
		if len(filters.LspID) > 0 {
			// cassandra contains clauses using lspIds
			lspIds := filters.LspID
			for _, lspId := range lspIds {
				if lspId == nil || *lspId == "" {
					continue
				}
				qryStr = qryStr + fmt.Sprintf(` and lsp_id='%s'`, *lspId)
			}
		}
		if filters.IsMandatory != nil {
			qryStr = qryStr + fmt.Sprintf(` and is_mandatory = %t`, *filters.IsMandatory)
		}
		if filters.Status != nil {
			qryStr = qryStr + fmt.Sprintf(` and course_status = '%s'`, *filters.Status)
		}
		if filters.Type != nil {
			qryStr = qryStr + fmt.Sprintf(` and course_type = '%s'`, *filters.Type)
		}
	}
	qryStr = qryStr + `ALLOW FILTERING`
	getUsers := func(page []byte) (courses []userz.UserCourse, nextPage []byte, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		q.PageState(page)
		q.PageSize(pageSizeInt)

		iter := q.Iter()
		return courses, iter.PageState(), iter.Select(&courses)
	}
	userCourses, newPage, err := getUsers(newPage)
	if err != nil {
		return nil, err
	}
	if len(newPage) != 0 {
		newCursor, err = global.CryptSession.EncryptAsString(newPage, nil)
		if err != nil {
			return nil, fmt.Errorf("error encrypting cursor: %v", err)
		}
		log.Infof("Users: %v", string(newCursor))

	}
	var outputResponse model.PaginatedCourseMaps
	allCourses := make([]*model.UserCourse, len(userCourses))
	if len(userCourses) == 0 {
		outputResponse.UserCourses = allCourses
		return &outputResponse, nil
	}
	var wg sync.WaitGroup
	for i, ccc := range userCourses {
		cc := ccc
		wg.Add(1)
		go func(i int, courseCopy userz.UserCourse) {
			endDate := strconv.FormatInt(courseCopy.EndDate, 10)
			createdAt := strconv.FormatInt(courseCopy.CreatedAt, 10)
			updatedAt := strconv.FormatInt(courseCopy.UpdatedAt, 10)
			currentCourse := &model.UserCourse{
				UserCourseID: &courseCopy.ID,
				UserID:       courseCopy.UserID,
				LspID:        &courseCopy.LspID,
				UserLspID:    courseCopy.UserLspID,
				CourseID:     courseCopy.CourseID,
				CourseType:   courseCopy.CourseType,
				AddedBy:      courseCopy.AddedBy,
				IsMandatory:  courseCopy.IsMandatory,
				EndDate:      &endDate,
				CourseStatus: courseCopy.CourseStatus,
				CreatedAt:    createdAt,
				UpdatedAt:    updatedAt,
				CreatedBy:    &courseCopy.CreatedBy,
				UpdatedBy:    &courseCopy.UpdatedBy,
			}
			allCourses[i] = currentCourse
			wg.Done()
		}(i, cc)
	}
	wg.Wait()
	outputResponse.UserCourses = allCourses
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction
	//redisBytes, err := json.Marshal(outputResponse)
	//if err == nil {
	//	redis.SetTTL(key, 90)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
	return &outputResponse, nil
}

func GetUserCourseMapByCourseID(ctx context.Context, userId string, courseID string) ([]*model.UserCourse, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	if userId != "" {
		emailCreatorID = userId
	}
	//key := "GetUserCourseMapByCourseID" + emailCreatorID + courseID
	//result, err := redis.GetRedisValue(key)
	//if err == nil {
	//	var outputResponse []*model.UserCourse
	//	err = json.Unmarshal([]byte(result), &outputResponse)
	//	if err == nil {
	//		return outputResponse, nil
	//	}
	//}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	qryStr := fmt.Sprintf(`SELECT * from userz.user_course_map where user_id='%s' and course_id='%s'  ALLOW FILTERING`, emailCreatorID, courseID)
	getUsers := func() (courses []userz.UserCourse, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return courses, iter.Select(&courses)
	}
	userCourses, err := getUsers()
	if err != nil {
		return nil, err
	}
	if len(userCourses) == 0 {
		return nil, fmt.Errorf("no user course found with id %s", courseID)
	}
	allCourses := make([]*model.UserCourse, 0)
	for _, copiedCourse := range userCourses {
		courseCopy := copiedCourse
		endDate := strconv.FormatInt(courseCopy.EndDate, 10)
		createdAt := strconv.FormatInt(courseCopy.CreatedAt, 10)
		updatedAt := strconv.FormatInt(courseCopy.UpdatedAt, 10)
		currentCourse := &model.UserCourse{
			UserCourseID: &courseCopy.ID,
			UserID:       courseCopy.UserID,
			UserLspID:    courseCopy.UserLspID,
			CourseID:     courseCopy.CourseID,
			CourseType:   courseCopy.CourseType,
			AddedBy:      courseCopy.AddedBy,
			IsMandatory:  courseCopy.IsMandatory,
			EndDate:      &endDate,
			CourseStatus: courseCopy.CourseStatus,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
			CreatedBy:    &courseCopy.CreatedBy,
			UpdatedBy:    &courseCopy.UpdatedBy,
		}
		allCourses = append(allCourses, currentCourse)
	}
	//redisBytes, err := json.Marshal(allCourses)
	//if err == nil {
	//	redis.SetTTL(key, 90)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
	return allCourses, nil
}

func GetUserCourseMapStats(ctx context.Context, input model.UserCourseMapStatsInput) (*model.UserCourseMapStats, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	whereClause := ""
	if input.LspID != nil {
		whereClause = fmt.Sprintf(`where lsp_id='%s'`, *input.LspID)
		if input.UserID != nil {
			whereClause = fmt.Sprintf(`where lsp_id='%s' and user_id='%s'`, *input.LspID, *input.UserID)
		}
	} else if input.UserID != nil {
		whereClause = fmt.Sprintf(`where user_id='%s'`, *input.UserID)
	}
	if input.IsMandatory != nil {
		if whereClause == "" {
			whereClause = fmt.Sprintf(`where is_mandatory=%t`, *input.IsMandatory)
		} else {
			whereClause = fmt.Sprintf(`%s and is_mandatory=%t`, whereClause, *input.IsMandatory)
		}
	}
	filteringRequired := false
	if input.CourseStatus != nil && input.CourseType != nil {
		if len(input.CourseStatus) > 1 && len(input.CourseType) > 1 {
			return nil, fmt.Errorf("course status and course type can only be used if one only one of each is provided")
		} else {
			filteringRequired = true
			whereClause = fmt.Sprintf(`where course_status='%s' and course_type='%s'`, *input.CourseStatus[0], *input.CourseType[0])
		}
	}
	var wg sync.WaitGroup
	statsStatus := make([]*model.Count, 0)
	if !filteringRequired && input.CourseStatus != nil && len(input.CourseStatus) > 0 {
		statsStatus = make([]*model.Count, len(input.CourseStatus))
		for i, s := range input.CourseStatus {
			if s == nil {
				continue
			}
			wg.Add(1)
			ss := *s
			tt := whereClause
			go func(i int, status string, tempClause string) {
				if i == 0 && tempClause == "" {
					tempClause = fmt.Sprintf(`where course_status='%s'`, status)
				} else {
					tempClause = whereClause + fmt.Sprintf(` AND course_status='%s'`, status)
				}
				qryStr := fmt.Sprintf(`SELECT count(*) from userz.user_course_map %s ALLOW FILTERING`, tempClause)
				getCSCount := func() (count int, success bool) {
					q := CassUserSession.Query(qryStr, nil)
					defer q.Release()
					iter := q.Iter()
					return count, iter.Scan(&count)
				}
				count, success := getCSCount()
				if !success {
					return
				}
				currentStatus := &model.Count{
					Name:  status,
					Count: count,
				}
				statsStatus[i] = currentStatus
				wg.Done()
			}(i, ss, tt)
		}
	}
	statsType := make([]*model.Count, 0)
	if !filteringRequired && input.CourseType != nil && len(input.CourseType) > 0 {
		statsType = make([]*model.Count, len(input.CourseType))
		for i, s := range input.CourseType {
			if s == nil {
				continue
			}
			wg.Add(1)
			tt := whereClause
			ss := *s
			go func(i int, status string, tempClause string) {
				if i == 0 && tempClause == "" {
					tempClause = fmt.Sprintf(`where course_type='%s'`, status)
				} else {
					tempClause = whereClause + fmt.Sprintf(` AND course_type='%s'`, status)
				}
				qryStr := fmt.Sprintf(`SELECT count(*) from userz.user_course_map %s ALLOW FILTERING`, tempClause)
				getCSCount := func() (count int, success bool) {
					q := CassUserSession.Query(qryStr, nil)
					defer q.Release()
					iter := q.Iter()
					return count, iter.Scan(&count)
				}
				count, success := getCSCount()
				if !success {
					return
				}
				currentStatus := &model.Count{
					Name:  status,
					Count: count,
				}
				statsType[i] = currentStatus
				wg.Done()
			}(i, ss, tt)
		}
	}
	wg.Wait()
	if filteringRequired {
		qryStr := fmt.Sprintf(`SELECT count(*) from userz.user_course_map %s ALLOW FILTERING`, whereClause)
		getCSCount := func() (count int, success bool) {
			q := CassUserSession.Query(qryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return count, iter.Scan(&count)
		}
		count, success := getCSCount()
		if !success {
			return nil, fmt.Errorf("error while getting count for course type %s", *input.CourseType[0])
		}
		currentStatus := &model.Count{
			Name:  *input.CourseStatus[0],
			Count: count,
		}
		statsStatus = append(statsStatus, currentStatus)
		currentType := &model.Count{
			Name:  *input.CourseType[0],
			Count: count,
		}
		statsType = append(statsType, currentType)
	}
	var currentOutput model.UserCourseMapStats
	currentOutput.LspID = input.LspID
	currentOutput.UserID = input.UserID
	currentOutput.StatusStats = statsStatus
	currentOutput.TypeStats = statsType
	return &currentOutput, nil
}

func GetCourseConsumptionStats(ctx context.Context, lspID string, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedCCStats, error) {
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	var newPage []byte
	//var pageDirection string
	var pageSizeInt int
	if pageCursor != nil && *pageCursor != "" {
		page, err := global.CryptSession.DecryptString(*pageCursor, nil)
		if err != nil {
			return nil, fmt.Errorf("invalid page cursor: %v", err)
		}
		newPage = page
	}
	if pageSize == nil {
		pageSizeInt = 10
	} else {
		pageSizeInt = *pageSize
	}
	var newCursor string
	qryStr := fmt.Sprintf(`SELECT * from userz.course_consumption_stats where lsp_id='%s' ALLOW FILTERING`, lspID)
	getCourseConsumptionStats := func(page []byte) (courseConsumptionStats []userz.CCStats, nextPage []byte, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		q.PageState(page)
		q.PageSize(pageSizeInt)
		iter := q.Iter()
		return courseConsumptionStats, nextPage, iter.Select(&courseConsumptionStats)
	}
	courseConsumptionStats, newPage, err := getCourseConsumptionStats(newPage)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	outputResponse := make([]*model.CourseConsumptionStats, len(courseConsumptionStats))
	for i, cc := range courseConsumptionStats {
		wg.Add(1)
		cStat := cc
		go func(i int, courseConsumptionStat userz.CCStats) {
			expectCompletiontime := int(courseConsumptionStat.ExpectedCompletionTime)
			avgCompletiontime := int(courseConsumptionStat.AverageCompletionTime)
			totalLearners := int(courseConsumptionStat.TotalLearners)
			activeLearners := int(courseConsumptionStat.ActiveLearners)
			completedLearners := int(courseConsumptionStat.CompletedLearners)
			avgScore := int(courseConsumptionStat.AverageComplianceScore)
			createdAt := int(courseConsumptionStat.CreatedAt)
			updatedAt := int(courseConsumptionStat.UpdatedAt)
			duration := int(courseConsumptionStat.Duration)
			currentCourseConsumptionStat := model.CourseConsumptionStats{
				ID:                     &courseConsumptionStat.ID,
				LspID:                  &courseConsumptionStat.LspId,
				CourseID:               &courseConsumptionStat.CourseId,
				ExpectedCompletionTime: &expectCompletiontime,
				AverageCompletionTime:  &avgCompletiontime,
				AverageComplianceScore: &avgScore,
				Category:               &courseConsumptionStat.Category,
				SubCategory:            &courseConsumptionStat.SubCategory,
				Owner:                  &courseConsumptionStat.Owner,
				Duration:               &duration,
				TotalLearners:          &totalLearners,
				ActiveLearners:         &activeLearners,
				CompletedLearners:      &completedLearners,
				CreatedAt:              &createdAt,
				UpdatedAt:              &updatedAt,
				CreatedBy:              &courseConsumptionStat.CreatedBy,
				UpdatedBy:              &courseConsumptionStat.UpdatedBy,
			}
			outputResponse[i] = &currentCourseConsumptionStat
			wg.Done()
		}(i, cStat)

	}
	wg.Wait()
	if len(newPage) != 0 {
		newCursor, err = global.CryptSession.EncryptAsString(newPage, nil)
		if err != nil {
			return nil, fmt.Errorf("error encrypting cursor: %v", err)
		}
		log.Infof("Users: %v", string(newCursor))

	}
	var resp model.PaginatedCCStats
	resp.PageCursor = &newCursor
	resp.Direction = direction
	resp.PageSize = pageSize
	resp.Stats = outputResponse
	return &resp, nil
}

func GetMostLeastAssignedCourse(ctx context.Context, lspID *string, input *string) (*model.CourseConsumptionStats, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	lsp := claims["lsp_id"].(string)
	if lspID != nil {
		lsp = *lspID
	}

	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	qryStr := fmt.Sprintf(`SELECT * FROM userz.course_consumption_stats WHERE lsp_id='%s' ALLOW FILTERING`, lsp)
	getDetails := func() (maps []userz.CCStats, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return maps, iter.Select(&maps)
	}
	ccStats, err := getDetails()
	if err != nil {
		return nil, err
	}

	if len(ccStats) == 0 {
		return nil, nil
	}
	//sort it on active learners
	sort.Slice(ccStats, func(i, j int) bool {
		return ccStats[i].ActiveLearners < ccStats[j].ActiveLearners
	})

	if *input == "least" {
		stat := ccStats[0]

		duration := int(stat.Duration)
		totalLearners := int(stat.TotalLearners)
		activeLearners := int(stat.ActiveLearners)
		completedLearners := int(stat.CompletedLearners)
		expectedCompletionTime := int(stat.ExpectedCompletionTime)
		averageCompletionTime := int(stat.AverageCompletionTime)
		averageComplianceScore := int(stat.AverageComplianceScore)
		createdAt := int(stat.CreatedAt)
		updatedAt := int(stat.UpdatedAt)
		res := model.CourseConsumptionStats{
			ID:                     &stat.ID,
			LspID:                  &stat.LspId,
			CourseID:               &stat.CourseId,
			Category:               &stat.Category,
			SubCategory:            &stat.SubCategory,
			Owner:                  &stat.Owner,
			Duration:               &duration,
			TotalLearners:          &totalLearners,
			ActiveLearners:         &activeLearners,
			CompletedLearners:      &completedLearners,
			ExpectedCompletionTime: &expectedCompletionTime,
			AverageCompletionTime:  &averageCompletionTime,
			AverageComplianceScore: &averageComplianceScore,
			CreatedAt:              &createdAt,
			UpdatedAt:              &updatedAt,
			CreatedBy:              &stat.CreatedBy,
			UpdatedBy:              &stat.UpdatedBy,
		}
		return &res, nil
	}
	stat := ccStats[len(ccStats)-1]
	duration := int(stat.Duration)
	totalLearners := int(stat.TotalLearners)
	activeLearners := int(stat.ActiveLearners)
	completedLearners := int(stat.CompletedLearners)
	expectedCompletionTime := int(stat.ExpectedCompletionTime)
	averageCompletionTime := int(stat.AverageCompletionTime)
	averageComplianceScore := int(stat.AverageComplianceScore)
	createdAt := int(stat.CreatedAt)
	updatedAt := int(stat.UpdatedAt)
	res := model.CourseConsumptionStats{
		ID:                     &stat.ID,
		LspID:                  &stat.LspId,
		CourseID:               &stat.CourseId,
		Category:               &stat.Category,
		SubCategory:            &stat.SubCategory,
		Owner:                  &stat.Owner,
		Duration:               &duration,
		TotalLearners:          &totalLearners,
		ActiveLearners:         &activeLearners,
		CompletedLearners:      &completedLearners,
		ExpectedCompletionTime: &expectedCompletionTime,
		AverageCompletionTime:  &averageCompletionTime,
		AverageComplianceScore: &averageComplianceScore,
		CreatedAt:              &createdAt,
		UpdatedAt:              &updatedAt,
		CreatedBy:              &stat.CreatedBy,
		UpdatedBy:              &stat.UpdatedBy,
	}
	return &res, nil
}
