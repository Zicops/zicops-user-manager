package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/helpers"
)

func AddUserCourse(ctx context.Context, input []*model.UserCourseInput) ([]*model.UserCourse, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	/*
		isAllowed := true
		role := strings.ToLower(userCass.Role)
		if userCass.ID == input[0].UserID || role == "admin" || strings.Contains(role, "manager") {
			isAllowed = true
		}
		if !isAllowed {
			return nil, fmt.Errorf("user not allowed to create org mapping")
		}
	*/

	userLspMaps := make([]*model.UserCourse, 0)
	for _, input := range input {

		createdBy := userCass.Email
		updatedBy := userCass.Email
		if input.CreatedBy != nil {
			createdBy = *input.CreatedBy
		}
		if input.UpdatedBy != nil {
			updatedBy = *input.UpdatedBy
		}
		var endDate int64
		if input.EndDate != nil {
			endDate, _ = strconv.ParseInt(*input.EndDate, 10, 64)
		}
		userLspMap := userz.UserCourse{
			ID:           uuid.New().String(),
			UserID:       input.UserID,
			LspID:        *input.LspID,
			UserLspID:    input.UserLspID,
			CourseID:     input.CourseID,
			CourseType:   input.CourseType,
			CourseStatus: input.CourseStatus,
			AddedBy:      input.AddedBy,
			IsMandatory:  input.IsMandatory,
			EndDate:      endDate,
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
			CreatedBy:    createdBy,
			UpdatedBy:    updatedBy,
		}
		insertQuery := CassUserSession.Query(userz.UserCourseTable.Insert()).BindStruct(userLspMap)
		if err := insertQuery.ExecRelease(); err != nil {
			return nil, err
		}
		//getcoursetopic - topics
		created := strconv.FormatInt(userLspMap.CreatedAt, 10)
		updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
		userLspOutput := &model.UserCourse{
			UserCourseID: &userLspMap.ID,
			UserLspID:    userLspMap.UserLspID,
			UserID:       userLspMap.UserID,
			CourseID:     userLspMap.CourseID,
			CourseType:   userLspMap.CourseType,
			CourseStatus: userLspMap.CourseStatus,
			AddedBy:      userLspMap.AddedBy,
			IsMandatory:  userLspMap.IsMandatory,
			EndDate:      input.EndDate,
			CreatedAt:    created,
			UpdatedAt:    updated,
			CreatedBy:    &userLspMap.CreatedBy,
			UpdatedBy:    &userLspMap.UpdatedBy,
		}
		userLspMaps = append(userLspMaps, userLspOutput)
	}
	return userLspMaps, nil
}

func UpdateUserCourse(ctx context.Context, input model.UserCourseInput) (*model.UserCourse, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	isAllowed := false
	role := strings.ToLower(userCass.Role)
	if userCass.ID == input.UserID || role == "admin" || strings.Contains(role, "manager") {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create org mapping")
	}
	if input.UserCourseID == nil {
		return nil, fmt.Errorf("user course id is required")
	}
	if input.UserID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userLspMap := userz.UserCourse{
		ID: *input.UserCourseID,
	}
	userLsps := []userz.UserCourse{}

	getQueryStr := fmt.Sprintf("SELECT * FROM userz.user_course_map WHERE id='%s' AND user_id='%s' ALLOW FILTERING", userLspMap.ID, input.UserID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
	if err := getQuery.SelectRelease(&userLsps); err != nil {
		return nil, err
	}
	if len(userLsps) == 0 {
		return nil, fmt.Errorf("users orgs not found")
	}
	userLspMap = userLsps[0]
	updatedCols := []string{}
	if input.EndDate != nil {
		endDate, _ := strconv.ParseInt(*input.EndDate, 10, 64)
		userLspMap.EndDate = endDate
		updatedCols = append(updatedCols, "end_date")
	}
	if input.CourseID != "" && input.CourseID != userLspMap.CourseID {
		userLspMap.CourseID = input.CourseID
		updatedCols = append(updatedCols, "course_id")
	}
	if input.CourseType != "" && input.CourseType != userLspMap.CourseType {
		userLspMap.CourseType = input.CourseType
		updatedCols = append(updatedCols, "course_type")
	}
	if input.CourseStatus != "" && input.CourseStatus != userLspMap.CourseStatus {
		if input.CourseStatus == "in-progress" && userLspMap.CourseStatus == "open" {
			userLspMap.CourseStatus = "started"
		} else if input.CourseStatus == "completed" && userLspMap.CourseStatus != "completed" {
			res := checkStatusOfEachTopic(ctx, input.UserID, userLspMap.ID)
			if res {
				userLspMap.CourseStatus = "completed"
			}
		} else {
			userLspMap.CourseStatus = input.CourseStatus
		}

		updatedCols = append(updatedCols, "course_status")
	}
	if input.AddedBy != "" && input.AddedBy != userLspMap.AddedBy {
		userLspMap.AddedBy = input.AddedBy
		updatedCols = append(updatedCols, "added_by")
	}
	if input.IsMandatory != userLspMap.IsMandatory {
		userLspMap.IsMandatory = input.IsMandatory
		updatedCols = append(updatedCols, "is_mandatory")
	}
	if input.UpdatedBy != nil {
		userLspMap.UpdatedBy = *input.UpdatedBy
		updatedCols = append(updatedCols, "updated_by")
	}
	if input.UserLspID != "" {
		userLspMap.UserLspID = input.UserLspID
		updatedCols = append(updatedCols, "user_lsp_id")
	}
	if input.LspID != nil {
		userLspMap.LspID = *input.LspID
		updatedCols = append(updatedCols, "lsp_id")
	}

	if len(updatedCols) > 0 {
		updatedAt := time.Now().Unix()
		userLspMap.UpdatedAt = updatedAt
		updatedCols = append(updatedCols, "updated_at")
		upStms, uNames := userz.UserCourseTable.Update(updatedCols...)
		updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&userLspMap)
		if err := updateQuery.ExecRelease(); err != nil {
			log.Errorf("error updating user course: %v", err)
			return nil, err
		}
	}
	created := strconv.FormatInt(userLspMap.CreatedAt, 10)
	updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
	userLspOutput := &model.UserCourse{
		UserCourseID: &userLspMap.ID,
		UserLspID:    userLspMap.UserLspID,
		UserID:       userLspMap.UserID,
		LspID:        &userLspMap.LspID,
		CourseID:     userLspMap.CourseID,
		CourseType:   userLspMap.CourseType,
		CourseStatus: userLspMap.CourseStatus,
		AddedBy:      userLspMap.AddedBy,
		IsMandatory:  userLspMap.IsMandatory,
		EndDate:      input.EndDate,
		CreatedAt:    created,
		UpdatedAt:    updated,
		CreatedBy:    &userLspMap.CreatedBy,
		UpdatedBy:    &userLspMap.UpdatedBy,
	}
	return userLspOutput, nil
}

func checkStatusOfEachTopic(ctx context.Context, userId string, userCourseId string) bool {

	userCP, err := getUserCourseProgressByUserCourseID(ctx, userId, userCourseId)
	if err != nil {
		log.Errorf("Got error while checking course progress: %v", err)
	}

	for _, vv := range userCP {
		v := vv
		res := *v
		if res.Status != "completed" {
			return false
		}
	}
	return true
}

func getUserCourseProgressByUserCourseID(ctx context.Context, userId string, userCourseID string) ([]*model.UserCourseProgress, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	if userId != "" {
		emailCreatorID = userId
	}

	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	userCPsMap := make([]*model.UserCourseProgress, 0)
	qryStr := fmt.Sprintf(`SELECT * from userz.user_course_progress where user_id='%s' and user_cm_id='%s'  ALLOW FILTERING`, emailCreatorID, userCourseID)
	getUsersCProgress := func() (users []userz.UserCourseProgress, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return users, iter.Select(&users)
	}
	userCPs, err := getUsersCProgress()
	if err != nil {
		return nil, err
	}
	userCPsMapCurrent := make([]*model.UserCourseProgress, len(userCPs))
	if len(userCPs) == 0 {
		return nil, nil
	}
	var wg sync.WaitGroup
	for i, copiedCP := range userCPs {
		userCP := copiedCP
		wg.Add(1)
		go func(i int, userCP userz.UserCourseProgress) {
			createdAt := strconv.FormatInt(userCP.CreatedAt, 10)
			updatedAt := strconv.FormatInt(userCP.UpdatedAt, 10)
			timeStamp := strconv.FormatInt(userCP.TimeStamp, 10)
			currentUserCP := &model.UserCourseProgress{
				UserCpID:      &userCP.ID,
				UserID:        userCP.UserID,
				UserCourseID:  userCP.UserCmID,
				TopicID:       userCP.TopicID,
				TopicType:     userCP.TopicType,
				Status:        userCP.Status,
				VideoProgress: userCP.VideoProgress,
				TimeStamp:     timeStamp,
				CreatedBy:     &userCP.CreatedBy,
				UpdatedBy:     &userCP.UpdatedBy,
				CreatedAt:     createdAt,
				UpdatedAt:     updatedAt,
			}
			userCPsMapCurrent[i] = currentUserCP
			wg.Done()
		}(i, userCP)
	}
	wg.Wait()
	userCPsMap = append(userCPsMap, userCPsMapCurrent...)
	return userCPsMap, nil
}
