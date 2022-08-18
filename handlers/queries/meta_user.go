package queries

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
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

	qryStr := fmt.Sprintf(`SELECT * from userz.user_notes where user_id='%s' and updated_at <= %d and user_lsp_id='%s' ALLOW FILTERING`, emailCreatorID, *publishTime, userLspID)
	getUsers := func(page []byte) (courses []userz.UserNotes, nextPage []byte, err error) {
		q := global.CassUserSession.Session.Query(qryStr, nil)
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
	var outputResponse model.PaginatedNotes
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
	return &outputResponse, nil
}

func GetUserBookmarks(ctx context.Context, userID string, userLspID string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedBookmarks, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	if userID != "" {
		emailCreatorID = userID
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
	if pageSize == nil {
		pageSizeInt = 10
	} else {
		pageSizeInt = *pageSize
	}
	var newCursor string

	qryStr := fmt.Sprintf(`SELECT * from userz.user_bookmarks where user_id='%s' and updated_at <= %d and user_lsp_id='%s' ALLOW FILTERING`, emailCreatorID, *publishTime, userLspID)
	getUsers := func(page []byte) (courses []userz.UserBookmarks, nextPage []byte, err error) {
		q := global.CassUserSession.Session.Query(qryStr, nil)
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
	var outputResponse model.PaginatedBookmarks
	allCourses := make([]*model.UserBookmark, 0)
	for _, copiedCourse := range userNotes {
		courseCopy := copiedCourse
		createdAt := strconv.FormatInt(courseCopy.CreatedAt, 10)
		updatedAt := strconv.FormatInt(courseCopy.UpdatedAt, 10)
		timeStamp := strconv.FormatInt(courseCopy.TimeStamp, 10)
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
			TimeStamp:    timeStamp,
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
	return &outputResponse, nil
}

func GetUserExamAttempts(ctx context.Context, userID string, userLspID string) ([]*model.UserExamAttempts, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	qryStr := fmt.Sprintf(`SELECT * from userz.user_exam_attempts where user_id='%s' and lsp_id='%s' ALLOW FILTERING`, userID, userLspID)
	getUserEA := func() (users []userz.UserExamAttempts, err error) {
		q := global.CassUserSession.Session.Query(qryStr, nil)
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
		attemptDuration := strconv.FormatInt(userOrg.AttemptDuration, 10)
		attemptStartTime := strconv.FormatInt(userOrg.AttemptStartTime, 10)
		currentUserOrg := &model.UserExamAttempts{
			UserEaID:         &copiedOrg.ID,
			UserID:           copiedOrg.UserID,
			UserLspID:        copiedOrg.UserLspID,
			UserCpID:         copiedOrg.UserCpID,
			UserCourseID:     copiedOrg.UserCmID,
			ExamID:           copiedOrg.ExamID,
			AttemptNo:        int(copiedOrg.AttemptNo),
			AttemptDuration:  attemptDuration,
			AttemptStatus:    copiedOrg.AttemptStatus,
			AttemptStartTime: attemptStartTime,
			CreatedBy:        &copiedOrg.CreatedBy,
			UpdatedBy:        &copiedOrg.UpdatedBy,
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
		}
		userOrgs = append(userOrgs, currentUserOrg)
	}
	return userOrgs, nil
}

func GetUserExamResults(ctx context.Context, userID string, userEaID string) (*model.UserExamResult, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	qryStr := fmt.Sprintf(`SELECT * from userz.user_exam_results where user_id='%s' and user_ea_id='%s' ALLOW FILTERING`, userID, userEaID)
	getUserEA := func() (users []userz.UserExamResults, err error) {
		q := global.CassUserSession.Session.Query(qryStr, nil)
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
	return userOrgs[0], nil
}
