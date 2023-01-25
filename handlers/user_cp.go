package handlers

import (
	"context"
	"fmt"
	"math"
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
	for _, v := range input {
		input := *v
		spliVPorgress := strings.Split(input.TimeStamp, "-")
		needTime := true
		if len(spliVPorgress) < 2 {
			needTime = false
		}
		half1Int := 0.0
		half2Int := 0.0
		if needTime {
			half1 := spliVPorgress[0]
			half2 := spliVPorgress[1]
			//convert half1 to int from floating point string
			half1Int, err = strconv.ParseFloat(half1, 64)
			if err != nil {
				log.Errorf("error while converting half1 to int: %v", err)
			}
			//convert half2 to int from floating point string
			half2Int, err = strconv.ParseFloat(half2, 64)
			if err != nil {
				log.Errorf("error while converting half2 to int: %v", err)
			}
			// convert half1 to int
			half1Int = math.Floor(half1Int)
			// convert half2 to int
			half2Int = math.Floor(half2Int)
		}
		totalTimeDiff := int64(half2Int - half1Int)
		createdBy := userCass.Email
		updatedBy := userCass.Email
		if input.CreatedBy != nil {
			createdBy = *input.CreatedBy
		}
		if input.UpdatedBy != nil {
			updatedBy = *input.UpdatedBy
		}
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
			TimeStamp:     totalTimeDiff,
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
			TimeStamp:     input.TimeStamp,
			VideoProgress: input.VideoProgress,
			CreatedAt:     created,
			UpdatedAt:     updated,
			CreatedBy:     &userLspMap.CreatedBy,
			UpdatedBy:     &userLspMap.UpdatedBy,
		}
		userLspMaps = append(userLspMaps, userLspOutput)
		if totalTimeDiff > 0 {
			go helpers.AddUpdateCourseViews(*lspID, userLspOutput.UserID, totalTimeDiff, 0)
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
		spliVPorgress := strings.Split(input.TimeStamp, "-")
		half1 := spliVPorgress[0]
		half2 := spliVPorgress[1]
		//convert half1 to int from floating point string
		half1Int, err := strconv.ParseFloat(half1, 64)
		if err != nil {
			log.Errorf("error while converting half1 to int: %v", err)
		}
		//convert half2 to int from floating point string
		half2Int, err := strconv.ParseFloat(half2, 64)
		if err != nil {
			log.Errorf("error while converting half2 to int: %v", err)
		}
		// convert half1 to int
		half1Int = math.Floor(half1Int)
		// convert half2 to int
		half2Int = math.Floor(half2Int)
		totalTimeDiff := int64(half2Int - half1Int)
		oldTimeDiff := userLspMap.TimeStamp
		totalTimeDiff = totalTimeDiff
		userLspMap.TimeStamp = totalTimeDiff
		updatedCols = append(updatedCols, "time_stamp")
		if totalTimeDiff > 0 {
			go helpers.AddUpdateCourseViews(*lspID, userLspMap.UserID, totalTimeDiff, oldTimeDiff)
		}
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
	return userLspOutput, nil
}

func GetCourseViews(ctx context.Context, lspIds []string, startTime *string, endTime *string) ([]*model.CourseViews, error) {
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	_, _, err = GetUserFromCassWithLsp(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	startT, err := strconv.ParseInt(*startTime, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("start time is required")
	}
	endT, err := strconv.ParseInt(*endTime, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("end time is required")
	}

	startDate := time.Unix(startT, 0)
	endDate := time.Unix(endT, 0)
	startDateString := startDate.Format("2006-01-02")
	endDateString := endDate.Format("2006-01-02")
	output := []*model.CourseViews{}
	for _, lspID := range lspIds {
		if lspID == "" {
			return nil, fmt.Errorf("lsp id is required")
		}
		getQueryStr := fmt.Sprintf("SELECT * FROM coursez.course_views WHERE lsp_id='%s' AND date_value >= '%s' AND date_value <= '%s' ALLOW FILTERING", lspID, startDateString, endDateString)
		getQuery := CassUserSession.Query(getQueryStr, nil)
		var courseViews []model.CourseViews
		if err := getQuery.SelectRelease(&courseViews); err != nil {
			log.Errorf("error getting course views: %v", err)
			return nil, err
		}
		if len(courseViews) == 0 {
			continue
		}
		currentView := courseViews[0]
		output = append(output, &currentView)
	}
	return output, nil
}
