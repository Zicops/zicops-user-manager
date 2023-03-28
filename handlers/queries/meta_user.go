package queries

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/lib/identity"
)

func GetUserNotes(ctx context.Context, userID string, userLspID *string, courseID *string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedNotes, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	if userID != "" {
		emailCreatorID = userID
	}
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
	var outputResponse model.PaginatedNotes

	//key := "GetUserNotes" + userLspID + string(newPage)
	//result, err := redis.GetRedisValue(key)
	//if err == nil {
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
	whereClause := ""
	if userLspID != nil && *userLspID != "" {
		whereClause = fmt.Sprintf(`user_lsp_id='%s'`, *userLspID)
	}
	if courseID != nil && *courseID != "" {
		if whereClause != "" {
			whereClause = whereClause + fmt.Sprintf(` AND course_id='%s'`, *courseID)
		} else {
			whereClause = fmt.Sprintf(`course_id='%s'`, *courseID)
		}
	}
	qryStr := fmt.Sprintf(`SELECT * from userz.user_notes where user_id='%s' and created_at <= %d and %s ALLOW FILTERING`, emailCreatorID, *publishTime, whereClause)
	getUsers := func(page []byte) (courses []userz.UserNotes, nextPage []byte, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		q.PageState(page)
		q.PageSize(pageSizeInt)

		iter := q.Iter()
		return courses, iter.PageState(), iter.Select(&courses)
	}
	userNotes, newPage, err := getUsers(newPage)
	if err != nil {
		return nil, err
	}
	if len(userNotes) == 0 {
		return nil, fmt.Errorf("no user notes found")
	}
	if len(newPage) != 0 {
		newCursor, err = global.CryptSession.EncryptAsString(newPage, nil)
		if err != nil {
			return nil, fmt.Errorf("error encrypting cursor: %v", err)
		}
		log.Infof("Users: %v", string(newCursor))

	}
	allCourses := make([]*model.UserNotes, 0)
	for _, copiedCourse := range userNotes {
		courseCopy := copiedCourse
		createdAt := strconv.FormatInt(courseCopy.CreatedAt, 10)
		updatedAt := strconv.FormatInt(courseCopy.UpdatedAt, 10)
		currentCourse := &model.UserNotes{
			UserNotesID: &courseCopy.ID,
			UserID:      courseCopy.UserID,
			UserLspID:   courseCopy.UserLspID,
			CourseID:    courseCopy.CourseID,
			ModuleID:    courseCopy.ModuleID,
			TopicID:     courseCopy.TopicID,
			Sequence:    courseCopy.Sequence,
			Status:      courseCopy.Status,
			Details:     courseCopy.Details,
			IsActive:    courseCopy.IsActive,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
			CreatedBy:   &courseCopy.CreatedBy,
			UpdatedBy:   &courseCopy.UpdatedBy,
		}
		allCourses = append(allCourses, currentCourse)
	}
	outputResponse.Notes = allCourses
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction
	//redisBytes, err := json.Marshal(outputResponse)
	// if err == nil {
	// 	redis.SetTTL(key, 3600)
	// 	redis.SetRedisValue(key, string(redisBytes))
	// }
	return &outputResponse, nil
}

func GetUserBookmarks(ctx context.Context, userID string, userLspID *string, courseID *string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedBookmarks, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	if userID != "" {
		emailCreatorID = userID
	}
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
	//key := "GetUserBookmarks" + userLspID + string(newPage) + userID
	//result, err := redis.GetRedisValue(key)
	var outputResponse model.PaginatedBookmarks
	//if err == nil {
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
	whereClause := ""
	if userLspID != nil {
		whereClause = fmt.Sprintf(" and user_lsp_id='%s'", *userLspID)
	}
	if courseID != nil {
		whereClause = whereClause + fmt.Sprintf(" and course_id='%s'", *courseID)
	}
	qryStr := fmt.Sprintf(`SELECT * from userz.user_bookmarks where user_id='%s' and created_at<=%d %s ALLOW FILTERING`, emailCreatorID, *publishTime, whereClause)
	getUsers := func(page []byte) (courses []userz.UserBookmarks, nextPage []byte, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		q.PageState(page)
		q.PageSize(pageSizeInt)

		iter := q.Iter()
		return courses, iter.PageState(), iter.Select(&courses)
	}
	userNotes, newPage, err := getUsers(newPage)
	if err != nil {
		return nil, err
	}
	if len(userNotes) == 0 {
		return nil, fmt.Errorf("no user notes found")
	}
	if len(newPage) != 0 {
		newCursor, err = global.CryptSession.EncryptAsString(newPage, nil)
		if err != nil {
			return nil, fmt.Errorf("error encrypting cursor: %v", err)
		}
		log.Infof("Users: %v", string(newCursor))

	}
	allCourses := make([]*model.UserBookmark, len(userNotes))
	if len(userNotes) == 0 {
		return &outputResponse, nil
	}
	var wg sync.WaitGroup
	for i, ccc := range userNotes {
		cc := ccc
		wg.Add(1)
		go func(i int, courseCopy userz.UserBookmarks) {
			createdAt := strconv.FormatInt(courseCopy.CreatedAt, 10)
			updatedAt := strconv.FormatInt(courseCopy.UpdatedAt, 10)
			currentCourse := &model.UserBookmark{
				UserBmID:     &courseCopy.ID,
				UserCourseID: courseCopy.UserCPID,
				UserID:       courseCopy.UserID,
				UserLspID:    courseCopy.UserLspID,
				CourseID:     courseCopy.CourseID,
				ModuleID:     courseCopy.ModuleID,
				TopicID:      courseCopy.TopicID,
				IsActive:     courseCopy.IsActive,
				Name:         courseCopy.Name,
				TimeStamp:    courseCopy.TimeStamp,
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
	outputResponse.Bookmarks = allCourses
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction
	//redisBytes, err := json.Marshal(outputResponse)
	//if err == nil {
	//	redis.SetTTL(key, 3600)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
	return &outputResponse, nil
}

func GetUserExamAttempts(ctx context.Context, userID *string, examID string) ([]*model.UserExamAttempts, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	//key := "GetUserExamAttempts" + userLspID + userID
	//result, err := redis.GetRedisValue(key)
	//var outputResponse []*model.UserExamAttempts
	//if err == nil {
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
	whereClause := ""
	if userID != nil {
		whereClause = fmt.Sprintf(" and user_id='%s'", *userID)
	}
	qryStr := fmt.Sprintf(`SELECT * from userz.user_exam_attempts where exam_id='%s' %s ALLOW FILTERING`, examID, whereClause)
	getUserEA := func() (users []userz.UserExamAttempts, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return users, iter.Select(&users)
	}
	usersOrgs, err := getUserEA()
	if err != nil {
		return nil, err
	}
	if len(usersOrgs) == 0 {
		return nil, fmt.Errorf("no user ea found")
	}
	userOrgs := make([]*model.UserExamAttempts, len(usersOrgs))
	if len(userOrgs) == 0 {
		return userOrgs, nil
	}
	var wg sync.WaitGroup
	for i, uo := range usersOrgs {
		cc := uo
		wg.Add(1)
		go func(i int, copiedOrg userz.UserExamAttempts) {
			createdAt := strconv.FormatInt(copiedOrg.CreatedAt, 10)
			updatedAt := strconv.FormatInt(copiedOrg.UpdatedAt, 10)
			attemptStartTime := strconv.FormatInt(copiedOrg.AttemptStartTime, 10)
			currentUserOrg := &model.UserExamAttempts{
				UserEaID:         &copiedOrg.ID,
				UserID:           copiedOrg.UserID,
				UserLspID:        copiedOrg.UserLspID,
				UserCpID:         copiedOrg.UserCpID,
				UserCourseID:     copiedOrg.UserCmID,
				ExamID:           copiedOrg.ExamID,
				AttemptNo:        int(copiedOrg.AttemptNo),
				AttemptDuration:  copiedOrg.AttemptDuration,
				AttemptStatus:    copiedOrg.AttemptStatus,
				AttemptStartTime: attemptStartTime,
				CreatedBy:        &copiedOrg.CreatedBy,
				UpdatedBy:        &copiedOrg.UpdatedBy,
				CreatedAt:        createdAt,
				UpdatedAt:        updatedAt,
			}
			userOrgs[i] = currentUserOrg
			wg.Done()
		}(i, cc)
	}
	wg.Wait()
	//redisBytes, err := json.Marshal(userOrgs)
	//if err == nil {
	//	redis.SetTTL(key, 3600)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
	return userOrgs, nil
}

func GetUserExamAttemptsByExamIds(ctx context.Context, userID string, examIds []*string, filters *model.ExamAttemptsFilters) ([]*model.UserExamAttempts, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Errorf("Got error while getting claims : %v", err)
		return nil, err
	}

	examAttempts := []userz.UserExamAttempts{}

	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	//exap attempts, filter on the basis
	//user course map - course id

	CassUserSession := session
	for _, vv := range examIds {
		if vv == nil {
			continue
		}
		v := *vv
		queryStr := fmt.Sprintf(`SELECT * FROM userz.user_exam_attempts where user_id='%s' AND exam_id='%s'`, userID, v)
		if filters != nil && filters.AttemptStatus != nil {
			queryStr = queryStr + fmt.Sprintf(` AND attempt_status='%s'`, *filters.AttemptStatus)
		}
		queryStr = queryStr + "  ALLOW FILTERING"
		getExamAttempts := func() (attempts []userz.UserExamAttempts, err error) {
			q := CassUserSession.Query(queryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return attempts, iter.Select(&attempts)
		}

		attempts, err := getExamAttempts()
		if err != nil {
			log.Println("Got error while getting user exam attempts: ", err.Error())
			return nil, err
		}
		if len(attempts) == 0 {
			continue
		}

		examAttempts = append(examAttempts, attempts...)
	}

	res := make([]*model.UserExamAttempts, len(examAttempts))
	var wg sync.WaitGroup
	for kk, vvv := range examAttempts {
		vv := vvv
		wg.Add(1)
		go func(k int, v userz.UserExamAttempts) {

			qryStr := fmt.Sprintf(`SELECT * FROM userz.user_course_map WHERE id='%s' ALLOW FILTERING`, v.UserCmID)
			getUserCourse := func() (courses []userz.UserCourse, err error) {
				q := CassUserSession.Query(qryStr, nil)
				defer q.Release()
				iter := q.Iter()
				return courses, iter.Select(&courses)
			}
			courses, err := getUserCourse()
			if err != nil {
				log.Errorf(err.Error())
				return
			}
			if len(courses) == 0 {
				return
			}

			attemptStartTime := strconv.FormatInt(v.AttemptStartTime, 10)
			createdAt := strconv.FormatInt(v.CreatedAt, 10)
			updatedAt := strconv.FormatInt(v.UpdatedAt, 10)
			currentAttempt := &model.UserExamAttempts{
				UserEaID:         &v.ID,
				UserID:           v.UserID,
				UserLspID:        v.UserLspID,
				UserCpID:         v.UserCpID,
				UserCourseID:     v.UserCmID,
				CourseID:         &courses[0].CourseID,
				ExamID:           v.ExamID,
				AttemptNo:        int(v.AttemptNo),
				AttemptDuration:  v.AttemptDuration,
				AttemptStatus:    v.AttemptStatus,
				AttemptStartTime: attemptStartTime,
				CreatedAt:        createdAt,
				CreatedBy:        &v.CreatedBy,
				UpdatedAt:        updatedAt,
				UpdatedBy:        &v.UpdatedBy,
			}
			res[k] = currentAttempt
			wg.Done()
		}(kk, vv)
	}
	wg.Wait()

	return res, nil
}

func GetUserExamResults(ctx context.Context, userEaDetails []*model.UserExamResultDetails) ([]*model.UserExamResultInfo, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	//key := "GetUserExamResults" + userEaID + userID
	//result, err := redis.GetRedisValue(key)
	//var outputResponse model.UserExamResult
	//if err == nil {
	//	err = json.Unmarshal([]byte(result), &outputResponse)
	//	if err == nil {
	//		return &outputResponse, nil
	//	}
	//}

	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	userOrgs := make([]*model.UserExamResultInfo, 0)
	for _, userEacopiedEADetail := range userEaDetails {
		if userEacopiedEADetail == nil {
			continue
		}
		userEaDetail := userEacopiedEADetail
		userID := userEaDetail.UserID
		userEaID := userEaDetail.UserEaID
		qryStr := fmt.Sprintf(`SELECT * from userz.user_exam_results where user_id='%s' and user_ea_id='%s'  ALLOW FILTERING`, userID, userEaID)
		getUserEA := func() (users []userz.UserExamResults, err error) {
			q := CassUserSession.Query(qryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return users, iter.Select(&users)
		}
		usersOrgs, err := getUserEA()
		if err != nil {
			return nil, err
		}
		if len(usersOrgs) == 0 {
			return nil, fmt.Errorf("no user exam results found")
		}
		tmpUserOrgs := make([]*model.UserExamResult, len(usersOrgs))
		var wg sync.WaitGroup
		for i, userOrg := range usersOrgs {
			cUserOrg := userOrg
			wg.Add(1)
			go func(i int, copiedOrg userz.UserExamResults) {
				createdAt := strconv.FormatInt(copiedOrg.CreatedAt, 10)
				updatedAt := strconv.FormatInt(copiedOrg.UpdatedAt, 10)
				currentUserOrg := &model.UserExamResult{
					UserErID:       &copiedOrg.ID,
					UserID:         copiedOrg.UserID,
					UserEaID:       copiedOrg.UserEaID,
					UserScore:      int(copiedOrg.UserScore),
					CorrectAnswers: int(copiedOrg.CorrectAnswers),
					WrongAnswers:   int(copiedOrg.WrongAnswers),
					ResultStatus:   copiedOrg.ResultStatus,
					CreatedBy:      &copiedOrg.CreatedBy,
					UpdatedBy:      &copiedOrg.UpdatedBy,
					CreatedAt:      createdAt,
					UpdatedAt:      updatedAt,
				}
				tmpUserOrgs[i] = currentUserOrg
				wg.Done()
			}(i, cUserOrg)
		}
		wg.Wait()
		var result model.UserExamResultInfo
		result.UserEaID = userEaID
		result.UserID = userID
		result.Results = tmpUserOrgs
		userOrgs = append(userOrgs, &result)
	}
	//redisBytes, err := json.Marshal(userOrgs[0])
	//if err == nil {
	//	redis.SetTTL(key, 3600)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
	return userOrgs, nil
}

func GetUserExamProgress(ctx context.Context, userID string, userEaID string) ([]*model.UserExamProgress, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	//key := "GetUserExamProgress" + userEaID + userID
	//result, err := redis.GetRedisValue(key)
	//var outputResponse []*model.UserExamProgress
	//if err == nil {
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

	qryStr := fmt.Sprintf(`SELECT * from userz.user_exam_progress where user_id='%s' and user_ea_id='%s'  ALLOW FILTERING`, userID, userEaID)
	getUserEA := func() (users []userz.UserExamProgress, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return users, iter.Select(&users)
	}
	usersOrgs, err := getUserEA()
	if err != nil {
		return nil, err
	}
	if len(usersOrgs) == 0 {
		return nil, fmt.Errorf("no user ep found")
	}
	userOrgs := make([]*model.UserExamProgress, len(usersOrgs))
	if len(userOrgs) == 0 {
		return userOrgs, nil
	}
	var wg sync.WaitGroup
	for i, uo := range usersOrgs {
		cc := uo
		wg.Add(1)
		go func(i int, copiedOrg userz.UserExamProgress) {
			createdAt := strconv.FormatInt(copiedOrg.CreatedAt, 10)
			updatedAt := strconv.FormatInt(copiedOrg.UpdatedAt, 10)
			totalTimeSpent := strconv.FormatInt(copiedOrg.TotalTimeSpent, 10)
			currentUserOrg := &model.UserExamProgress{
				UserEpID:       &copiedOrg.ID,
				UserID:         copiedOrg.UserID,
				UserLspID:      copiedOrg.UserLspID,
				UserCpID:       copiedOrg.UserCpID,
				UserEaID:       copiedOrg.UserEaID,
				SrNo:           int(copiedOrg.SrNo),
				QuestionID:     copiedOrg.QuestionID,
				QuestionType:   copiedOrg.QuestionType,
				Answer:         copiedOrg.Answer,
				QAttemptStatus: copiedOrg.QAttemptStatus,
				TotalTimeSpent: totalTimeSpent,
				CorrectAnswer:  copiedOrg.CorrectAnswer,
				SectionID:      copiedOrg.SectionID,
				CreatedBy:      &copiedOrg.CreatedBy,
				UpdatedBy:      &copiedOrg.UpdatedBy,
				CreatedAt:      createdAt,
				UpdatedAt:      updatedAt,
			}
			userOrgs[i] = currentUserOrg
			wg.Done()
		}(i, cc)
	}
	wg.Wait()
	//redisBytes, err := json.Marshal(userOrgs)
	//if err == nil {
	//	redis.SetTTL(key, 60)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
	return userOrgs, nil
}

func GetUserQuizAttempts(ctx context.Context, userID string, topicID string) ([]*model.UserQuizAttempt, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	//key := "GetUserQuizAttempts" + topicID + userID
	//result, err := redis.GetRedisValue(key)
	//var outputResponse []*model.UserQuizAttempt
	//if err == nil {
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

	qryStr := fmt.Sprintf(`SELECT * from userz.user_quiz_attempts where user_id='%s' and topic_id='%s'  ALLOW FILTERING`, userID, topicID)
	getUserQA := func() (users []userz.UserQuizAttempts, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return users, iter.Select(&users)
	}
	usersOrgs, err := getUserQA()
	if err != nil {
		return nil, err
	}
	if len(usersOrgs) == 0 {
		return nil, fmt.Errorf("no user qa found")
	}
	userOrgs := make([]*model.UserQuizAttempt, 0)
	for _, userOrg := range usersOrgs {
		copiedOrg := userOrg
		createdAt := strconv.FormatInt(userOrg.CreatedAt, 10)
		updatedAt := strconv.FormatInt(userOrg.UpdatedAt, 10)
		currentUserOrg := &model.UserQuizAttempt{
			UserQaID:     &copiedOrg.ID,
			UserID:       copiedOrg.UserID,
			TopicID:      copiedOrg.TopicID,
			UserCpID:     copiedOrg.UserCpID,
			UserCourseID: copiedOrg.UserCmID,
			QuizID:       copiedOrg.QuizID,
			QuizAttempt:  int(copiedOrg.QuizAttempt),
			Result:       copiedOrg.Result,
			IsActive:     copiedOrg.IsActive,
			CreatedBy:    &copiedOrg.CreatedBy,
			UpdatedBy:    &copiedOrg.UpdatedBy,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		}
		userOrgs = append(userOrgs, currentUserOrg)
	}
	//redisBytes, err := json.Marshal(userOrgs)
	//if err == nil {
	//	redis.SetTTL(key, 60)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
	return userOrgs, nil
}
