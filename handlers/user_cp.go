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
	"github.com/zicops/contracts/coursez"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/lib/stats"
)

func AddUserCourseProgress(ctx context.Context, input []*model.UserCourseProgressInput) ([]*model.UserCourseProgress, error) {
	userCass, lspID, err := GetUserFromCassWithLsp(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
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
		if v == nil {
			continue
		}
		input := *v
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
			TimeStamp:     0,
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
		go stats.AddUpdateCourseViews(*lspID, userLspOutput.UserID, 0, 0)

	}
	return userLspMaps, nil
}

// if all topic status is complete and course status != complete then course status- complete
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
	session, err := global.CassPool.GetSession(ctx, "userz")
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
			_, err = UpdateUserCourse(ctx, model.UserCourseInput{
				UserCourseID: &userLspMap.UserCmID,
				UserID:       userLspMap.UserID,
				CourseStatus: "started",
			})

			if err != nil {
				return nil, err
			}
		}
		if input.Status == "completed" && userLspMap.Status != "completed" {
			res := checkStatusOfEachTopic(ctx, input.UserID, input.UserCourseID)
			if res {
				//update course to be completed
				_, err = UpdateUserCourse(ctx, model.UserCourseInput{
					UserCourseID: &userLspMap.UserCmID,
					UserID:       userLspMap.UserID,
					CourseStatus: "completed",
				})

				if err != nil {
					return nil, err
				}
			}
		}
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
		spliVPorgress := strings.Split(input.TimeStamp, "-")
		half1 := spliVPorgress[0]
		//convert half1 to int from floating point string
		half1Int, err := strconv.ParseFloat(half1, 64)
		if err != nil {
			log.Errorf("error while converting half1 to int: %v", err)
		}
		// convert half1 to int
		half1Int = math.Floor(half1Int)
		oldTimeStamp := userLspMap.TimeStamp
		userLspMap.TimeStamp = int64(half1Int)
		updatedCols = append(updatedCols, "time_stamp")
		diff := userLspMap.TimeStamp - oldTimeStamp
		if diff > 0 {
			go stats.AddUpdateCourseViews(*lspID, userLspMap.UserID, diff, 0)
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
	session, err := global.CassPool.GetSession(ctx, "coursez")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	_, _, err = GetUserFromCassWithLsp(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	if startTime == nil || endTime == nil {
		return nil, fmt.Errorf("start and end time are required")
	}
	startDateString := *startTime
	endDateString := *endTime
	output := []*model.CourseViews{}
	for _, lspID := range lspIds {
		if lspID == "" {
			return nil, fmt.Errorf("lsp id is required")
		}
		getQueryStr := fmt.Sprintf("SELECT * FROM coursez.course_views WHERE lsp_id='%s' AND date_value >= '%s' AND date_value <= '%s' ALLOW FILTERING", lspID, startDateString, endDateString)
		getQuery := CassUserSession.Query(getQueryStr, nil)
		var courseViews []coursez.CourseView
		if err := getQuery.SelectRelease(&courseViews); err != nil {
			log.Errorf("error getting course views: %v", err)
			return nil, err
		}
		for _, cv := range courseViews {
			currentView := cv
			seconds := int(currentView.Hours)
			createdAt := strconv.Itoa(int(currentView.CreatedAt))
			var userIds []*string
			for _, vv := range currentView.Users {
				v := vv
				userIds = append(userIds, &v)
			}
			res := model.CourseViews{
				Seconds:    &seconds,
				CreatedAt:  &createdAt,
				LspID:      &currentView.LspId,
				UserIds:    userIds,
				DateString: &currentView.DateValue,
			}
			output = append(output, &res)
		}
	}
	return output, nil
}

//usercourse map
//user course progress
//input - completed, map not completed - check status of each topic of user course progresss - if yes, map - compeleted
//input - in progress, map open, user course map started

//input, output
//input in progress,
