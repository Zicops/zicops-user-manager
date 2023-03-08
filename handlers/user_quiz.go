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
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
)

func AddUserQuizAttempt(ctx context.Context, input []*model.UserQuizAttemptInput) ([]*model.UserQuizAttempt, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	isAllowed := false
	role := strings.ToLower(userCass.Role)
	if userCass.ID == input[0].UserID || role == "admin" || strings.Contains(role, "manager") {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create org mapping")
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userLspMaps := make([]*model.UserQuizAttempt, 0)
	for _, input := range input {

		if input == nil {
			continue
		}
		createdBy := userCass.Email
		updatedBy := userCass.Email
		if input.CreatedBy != nil {
			createdBy = *input.CreatedBy
		}
		if input.UpdatedBy != nil {
			updatedBy = *input.UpdatedBy
		}
		// convert input.QuizAttempt to int64
		quizAttempt := int64(input.QuizAttempt)
		userLspMap := userz.UserQuizAttempts{
			ID:          uuid.New().String(),
			UserID:      input.UserID,
			UserCmID:    input.UserCourseID,
			UserCpID:    input.UserCpID,
			QuizID:      input.QuizID,
			QuizAttempt: quizAttempt,
			Result:      input.Result,
			TopicID:     input.TopicID,
			IsActive:    input.IsActive,
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
			CreatedBy:   createdBy,
			UpdatedBy:   updatedBy,
		}
		insertQuery := CassUserSession.Query(userz.UserQuizAttemptsTable.Insert()).BindStruct(userLspMap)
		if err := insertQuery.ExecRelease(); err != nil {
			return nil, err
		}
		created := strconv.FormatInt(userLspMap.CreatedAt, 10)
		updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
		userLspOutput := &model.UserQuizAttempt{
			UserQaID:     &userLspMap.ID,
			UserID:       userLspMap.UserID,
			UserCourseID: userLspMap.UserCmID,
			UserCpID:     userLspMap.UserCpID,
			QuizID:       userLspMap.QuizID,
			QuizAttempt:  input.QuizAttempt,
			Result:       userLspMap.Result,
			IsActive:     userLspMap.IsActive,
			CreatedAt:    created,
			UpdatedAt:    updated,
			CreatedBy:    &userLspMap.CreatedBy,
			UpdatedBy:    &userLspMap.UpdatedBy,
			TopicID:      userLspMap.TopicID,
		}
		userLspMaps = append(userLspMaps, userLspOutput)
	}
	return userLspMaps, nil
}

func UpdateUserQuizAttempt(ctx context.Context, input model.UserQuizAttemptInput) (*model.UserQuizAttempt, error) {
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
	if input.UserQaID == nil {
		return nil, fmt.Errorf("user qa id is required")
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userLspMap := userz.UserQuizAttempts{
		ID: *input.UserQaID,
	}
	userLsps := []userz.UserQuizAttempts{}
	userID := userCass.ID
	if input.UserID != "" {
		userID = input.UserID
	}
	getQueryStr := fmt.Sprintf("SELECT * FROM userz.user_quiz_attempts WHERE id='%s' AND user_id='%s'  ", userLspMap.ID, userID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
	if err := getQuery.SelectRelease(&userLsps); err != nil {
		return nil, err
	}
	if len(userLsps) == 0 {
		return nil, fmt.Errorf("users orgs not found")
	}
	userLspMap = userLsps[0]
	updatedCols := []string{}
	if input.UserCpID != "" && input.UserCpID != userLspMap.UserCpID {
		userLspMap.UserCpID = input.UserCpID
		updatedCols = append(updatedCols, "user_cp_id")
	}
	if input.UserCourseID != "" && input.UserCourseID != userLspMap.UserCmID {
		userLspMap.UserCmID = input.UserCourseID
		updatedCols = append(updatedCols, "user_cm_id")
	}
	if input.QuizID != "" && input.QuizID != userLspMap.QuizID {
		userLspMap.QuizID = input.QuizID
		updatedCols = append(updatedCols, "quiz_id")
	}
	if input.QuizAttempt != 0 && int64(input.QuizAttempt) != userLspMap.QuizAttempt {
		userLspMap.QuizAttempt = int64(input.QuizAttempt)
		updatedCols = append(updatedCols, "quiz_attempt")
	}
	if input.Result != "" && input.Result != userLspMap.Result {
		userLspMap.Result = input.Result
		updatedCols = append(updatedCols, "result")
	}
	if input.TopicID != "" && input.TopicID != userLspMap.TopicID {
		userLspMap.TopicID = input.TopicID
		updatedCols = append(updatedCols, "topic_id")
	}
	if input.UpdatedBy != nil {
		userLspMap.UpdatedBy = *input.UpdatedBy
		updatedCols = append(updatedCols, "updated_by")
	}
	if input.IsActive != userLspMap.IsActive {
		userLspMap.IsActive = input.IsActive
		updatedCols = append(updatedCols, "is_active")
	}
	if len(updatedCols) > 0 {
		updatedAt := time.Now().Unix()
		userLspMap.UpdatedAt = updatedAt
		updatedCols = append(updatedCols, "updated_at")
		upStms, uNames := userz.UserQuizAttemptsTable.Update(updatedCols...)
		updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&userLspMap)
		if err := updateQuery.ExecRelease(); err != nil {
			log.Errorf("error updating user quiz attempts: %v", err)
			return nil, err
		}
	}
	created := strconv.FormatInt(userLspMap.CreatedAt, 10)
	updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
	userLspOutput := &model.UserQuizAttempt{
		UserQaID:     &userLspMap.ID,
		UserID:       userLspMap.UserID,
		UserCourseID: userLspMap.UserCmID,
		UserCpID:     userLspMap.UserCpID,
		QuizID:       userLspMap.QuizID,
		QuizAttempt:  input.QuizAttempt,
		Result:       userLspMap.Result,
		IsActive:     userLspMap.IsActive,
		CreatedAt:    created,
		UpdatedAt:    updated,
		CreatedBy:    &userLspMap.CreatedBy,
		UpdatedBy:    &userLspMap.UpdatedBy,
		TopicID:      userLspMap.TopicID,
	}
	return userLspOutput, nil
}
