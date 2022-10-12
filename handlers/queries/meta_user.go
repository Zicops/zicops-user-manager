package queries

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/helpers"
)

func GetUserNotes(ctx context.Context, userID string, userLspID string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedNotes, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	if userID != "" {
		emailCreatorID = userID
	}
	session, err := cassandra.GetCassSession("userz")
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

	qryStr := fmt.Sprintf(`SELECT * from userz.user_notes where user_id='%s' and updated_at <= %d and user_lsp_id='%s' ALLOW FILTERING`, emailCreatorID, *publishTime, userLspID)
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

func GetUserBookmarks(ctx context.Context, userID string, userLspID *string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedBookmarks, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	if userID != "" {
		emailCreatorID = userID
	}
	session, err := cassandra.GetCassSession("userz")
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
	qryStr := fmt.Sprintf(`SELECT * from userz.user_bookmarks where user_id='%s' and updated_at<=%d %s ALLOW FILTERING`, emailCreatorID, *publishTime, whereClause)
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
	allCourses := make([]*model.UserBookmark, 0)
	for _, copiedCourse := range userNotes {
		courseCopy := copiedCourse
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
		allCourses = append(allCourses, currentCourse)
	}
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

func GetUserExamAttempts(ctx context.Context, userID string, userLspID string) ([]*model.UserExamAttempts, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
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
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	qryStr := fmt.Sprintf(`SELECT * from userz.user_exam_attempts where user_id='%s' and user_lsp_id='%s' ALLOW FILTERING`, userID, userLspID)
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
	userOrgs := make([]*model.UserExamAttempts, 0)
	for _, userOrg := range usersOrgs {
		copiedOrg := userOrg
		createdAt := strconv.FormatInt(userOrg.CreatedAt, 10)
		updatedAt := strconv.FormatInt(userOrg.UpdatedAt, 10)
		attemptStartTime := strconv.FormatInt(userOrg.AttemptStartTime, 10)
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
		userOrgs = append(userOrgs, currentUserOrg)
	}
	//redisBytes, err := json.Marshal(userOrgs)
	//if err == nil {
	//	redis.SetTTL(key, 3600)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
	return userOrgs, nil
}

func GetUserExamResults(ctx context.Context, userID string, userEaID string) (*model.UserExamResult, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
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

	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	qryStr := fmt.Sprintf(`SELECT * from userz.user_exam_results where user_id='%s' and user_ea_id='%s' ALLOW FILTERING`, userID, userEaID)
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
	userOrgs := make([]*model.UserExamResult, 0)
	for _, userOrg := range usersOrgs {
		copiedOrg := userOrg
		createdAt := strconv.FormatInt(userOrg.CreatedAt, 10)
		updatedAt := strconv.FormatInt(userOrg.UpdatedAt, 10)
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
		userOrgs = append(userOrgs, currentUserOrg)
	}
	//redisBytes, err := json.Marshal(userOrgs[0])
	//if err == nil {
	//	redis.SetTTL(key, 3600)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
	return userOrgs[0], nil
}

func GetUserExamProgress(ctx context.Context, userID string, userEaID string) ([]*model.UserExamProgress, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
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
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	qryStr := fmt.Sprintf(`SELECT * from userz.user_exam_progress where user_id='%s' and user_ea_id='%s' ALLOW FILTERING`, userID, userEaID)
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
	userOrgs := make([]*model.UserExamProgress, 0)
	for _, userOrg := range usersOrgs {
		copiedOrg := userOrg
		createdAt := strconv.FormatInt(userOrg.CreatedAt, 10)
		updatedAt := strconv.FormatInt(userOrg.UpdatedAt, 10)
		totalTimeSpent := strconv.FormatInt(userOrg.TotalTimeSpent, 10)
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
		userOrgs = append(userOrgs, currentUserOrg)
	}
	//redisBytes, err := json.Marshal(userOrgs)
	//if err == nil {
	//	redis.SetTTL(key, 60)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
	return userOrgs, nil
}

func GetUserQuizAttempts(ctx context.Context, userID string, topicID string) ([]*model.UserQuizAttempt, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
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
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	qryStr := fmt.Sprintf(`SELECT * from userz.user_quiz_attempts where user_id='%s' and topic_id='%s' ALLOW FILTERING`, userID, topicID)
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
