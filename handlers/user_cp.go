package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rs/xid"
	"github.com/scylladb/gocqlx/qb"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
)

func AddUserCourseProgress(ctx context.Context, input []*model.UserCourseProgressInput) ([]*model.UserCourseProgress, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	global.CassUserSession = session
	defer global.CassUserSession.Close()
	isAllowed := false
	if userCass.ID == input[0].UserID || strings.ToLower(userCass.Role) == "admin" {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create org mapping")
	}
	userLspMaps := make([]*model.UserCourseProgress, 0)
	for _, input := range input {
		guid := xid.New()
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
			ID:            guid.String(),
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

		insertQuery := global.CassUserSession.Query(userz.UserCourseProgressTable.Insert()).BindStruct(userLspMap)
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
	}
	return userLspMaps, nil
}

func UpdateUserCourseProgress(ctx context.Context, input model.UserCourseProgressInput) (*model.UserCourseProgress, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	isAllowed := false
	if userCass.ID == input.UserID || strings.ToLower(userCass.Role) == "admin" {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create course progress mapping")
	}
	if input.UserCpID == nil {
		return nil, fmt.Errorf("user cp id is required")
	}
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	global.CassUserSession = session
	defer global.CassUserSession.Close()
	userLspMap := userz.UserCourseProgress{
		ID: *input.UserCpID,
	}
	userLsps := []userz.UserCourseProgress{}
	getQuery := global.CassUserSession.Query(userz.UserCourseProgressTable.Get()).BindMap(qb.M{"id": userLspMap.ID})
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
		userLspMap.Status = input.Status
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
	updatedAt := time.Now().Unix()
	userLspMap.UpdatedAt = updatedAt
	updatedCols = append(updatedCols, "updated_at")
	if len(updatedCols) == 0 {
		return nil, fmt.Errorf("nothing to update")
	}
	upStms, uNames := userz.UserCourseProgressTable.Update(updatedCols...)
	updateQuery := global.CassUserSession.Query(upStms, uNames).BindStruct(&userLspMap)
	if err := updateQuery.ExecRelease(); err != nil {
		log.Errorf("error updating user course progress: %v", err)
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
	return userLspOutput, nil
}
