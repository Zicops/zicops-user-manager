package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/helpers"
)

func AddUserCourseProgress(ctx context.Context, input []*model.UserCourseProgressInput) ([]*model.UserCourseProgress, error) {
	userCass, lspID, err := GetUserFromCassWithLsp(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	isAllowed := false
	role := strings.ToLower(userCass.Role)
	if userCass.ID == input[0].UserID || role == "admin" || strings.Contains(role, "manager") {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create org mapping")
	}
	userLspMaps := make([]*model.UserCourseProgress, 0)
	for _, input := range input {

		createdBy := userCass.Email
		updatedBy := userCass.Email
		if input.CreatedBy != nil {
			createdBy = *input.CreatedBy
		}
		if input.UpdatedBy != nil {
			updatedBy = *input.UpdatedBy
		}
		timeStamp, _ := strconv.ParseInt(input.TimeStamp, 10, 64)
		videoProgress := ""
		if input.VideoProgress != "" {
			videoProgress = input.VideoProgress
		}
		userLspMap := userz.UserCourseProgress{
			ID:            uuid.New().String(),
			UserID:        input.UserID,
			UserCmID:      input.UserCourseID,
			TopicID:       input.TopicID,
			TopicType:     input.TopicType,
			Status:        input.Status,
			TimeStamp:     timeStamp,
			VideoProgress: videoProgress,
			CreatedAt:     time.Now().Unix(),
			UpdatedAt:     time.Now().Unix(),
			CreatedBy:     createdBy,
			UpdatedBy:     updatedBy,
		}

		insertQuery := CassUserSession.Query(userz.UserCourseProgressTable.Insert()).BindStruct(userLspMap)
		if err := insertQuery.ExecRelease(); err != nil {
			return nil, err
		}
		created := strconv.FormatInt(userLspMap.CreatedAt, 10)
		updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
		userLspOutput := &model.UserCourseProgress{
			UserCpID:      &userLspMap.ID,
			UserCourseID:  userLspMap.UserCmID,
			UserID:        userLspMap.UserID,
			TopicID:       userLspMap.TopicID,
			TopicType:     userLspMap.TopicType,
			Status:        userLspMap.Status,
			TimeStamp:     strconv.FormatInt(userLspMap.TimeStamp, 10),
			VideoProgress: input.VideoProgress,
			CreatedAt:     created,
			UpdatedAt:     updated,
			CreatedBy:     &userLspMap.CreatedBy,
			UpdatedBy:     &userLspMap.UpdatedBy,
		}
		userLspMaps = append(userLspMaps, userLspOutput)
		vProgressSeconds, err := strconv.ParseInt(input.VideoProgress, 10, 64)
		if err == nil {
			go helpers.AddUpdateCourseViews(ctx, *lspID, userLspMap.TopicID, userLspMap.UserID, vProgressSeconds)
		}
	}
	return userLspMaps, nil
}

func UpdateUserCourseProgress(ctx context.Context, input model.UserCourseProgressInput) (*model.UserCourseProgress, error) {
	userCass, lspID, err := GetUserFromCassWithLsp(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	isAllowed := false
	role := strings.ToLower(userCass.Role)
	if userCass.ID == input.UserID || role == "admin" || strings.Contains(role, "manager") {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create course progress mapping")
	}
	if input.UserCpID == nil {
		return nil, fmt.Errorf("user cp id is required")
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userLspMap := userz.UserCourseProgress{
		ID: *input.UserCpID,
	}
	userLsps := []userz.UserCourseProgress{}
	userID := userCass.ID
	if input.UserID != "" {
		userID = input.UserID
	}
	getQueryStr := fmt.Sprintf("SELECT * FROM userz.user_course_progress WHERE id='%s' AND user_id='%s'  ", userLspMap.ID, userID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
	if err := getQuery.SelectRelease(&userLsps); err != nil {
		return nil, err
	}
	if len(userLsps) == 0 {
		return nil, fmt.Errorf("users cp not found")
	}
	userLspMap = userLsps[0]
	updatedCols := []string{}
	if input.UserCourseID != "" && input.UserCourseID != userLspMap.UserCmID {
		userLspMap.UserCmID = input.UserCourseID
		updatedCols = append(updatedCols, "user_cm_id")
	}
	if input.VideoProgress != "" && input.VideoProgress != userLspMap.VideoProgress {
		userLspMap.VideoProgress = input.VideoProgress
		updatedCols = append(updatedCols, "video_progress")
	}
	if input.Status != "" && input.Status != userLspMap.Status {

		if input.Status == "in-progress" && userLspMap.Status == "open" {
			userLspMap.Status = "started"
		} else {
			userLspMap.Status = input.Status
		}
		updatedCols = append(updatedCols, "status")
	}
	if input.TopicID != "" && input.TopicID != userLspMap.TopicID {
		userLspMap.TopicID = input.TopicID
		updatedCols = append(updatedCols, "topic_id")
	}
	if input.TopicType != "" && input.TopicType != userLspMap.TopicType {
		userLspMap.TopicType = input.TopicType
		updatedCols = append(updatedCols, "topic_type")
	}
	if input.TimeStamp != "" && input.TimeStamp != strconv.FormatInt(userLspMap.TimeStamp, 10) {
		timeStamp, _ := strconv.ParseInt(input.TimeStamp, 10, 64)
		userLspMap.TimeStamp = timeStamp
		updatedCols = append(updatedCols, "time_stamp")
	}
	if input.UpdatedBy != nil {
		userLspMap.UpdatedBy = *input.UpdatedBy
		updatedCols = append(updatedCols, "updated_by")
	}

	if len(updatedCols) > 0 {
		updatedAt := time.Now().Unix()
		userLspMap.UpdatedAt = updatedAt
		updatedCols = append(updatedCols, "updated_at")
		upStms, uNames := userz.UserCourseProgressTable.Update(updatedCols...)
		updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&userLspMap)
		if err := updateQuery.ExecRelease(); err != nil {
			log.Errorf("error updating user course progress: %v", err)
			return nil, err
		}
	}
	created := strconv.FormatInt(userLspMap.CreatedAt, 10)
	updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
	userLspOutput := &model.UserCourseProgress{
		UserCpID:      &userLspMap.ID,
		UserCourseID:  userLspMap.UserCmID,
		UserID:        userLspMap.UserID,
		TopicID:       userLspMap.TopicID,
		TopicType:     userLspMap.TopicType,
		Status:        userLspMap.Status,
		TimeStamp:     strconv.FormatInt(userLspMap.TimeStamp, 10),
		VideoProgress: input.VideoProgress,
		CreatedAt:     created,
		UpdatedAt:     updated,
		CreatedBy:     &userLspMap.CreatedBy,
		UpdatedBy:     &userLspMap.UpdatedBy,
	}
	vProgressSeconds, err := strconv.ParseInt(input.VideoProgress, 10, 64)
	if err == nil {
		go helpers.AddUpdateCourseViews(ctx, *lspID, userLspMap.TopicID, userLspMap.UserID, vProgressSeconds)
	}
	return userLspOutput, nil
}
