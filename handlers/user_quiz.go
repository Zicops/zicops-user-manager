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
	"github.com/zicops/zicops-user-manager/graph/model"
)

func AddUserQuizAttempt(ctx context.Context, input []*model.UserQuizAttemptInput) ([]*model.UserQuizAttempt, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	isAllowed := false
	if userCass.ID == input[0].UserID || strings.ToLower(userCass.Role) == "admin" {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create org mapping")
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userLspMaps := make([]*model.UserQuizAttempt, 0)
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
		// convert input.QuizAttempt to int64
		quizAttempt := int64(input.QuizAttempt)
		userLspMap := userz.UserQuizAttempts{
			ID:          guid.String(),
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
	if userCass.ID == input.UserID || strings.ToLower(userCass.Role) == "admin" {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create org mapping")
	}
	if input.UserQaID == nil {
		return nil, fmt.Errorf("user qa id is required")
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userLspMap := userz.UserQuizAttempts{
		ID: *input.UserQaID,
	}
	userLsps := []userz.UserQuizAttempts{}
	getQuery := CassUserSession.Query(userz.UserQuizAttemptsTable.Get()).BindMap(qb.M{"id": userLspMap.ID, "user_id": userCass.ID})
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
	updatedAt := time.Now().Unix()
	userLspMap.UpdatedAt = updatedAt
	updatedCols = append(updatedCols, "updated_at")
	if len(updatedCols) == 0 {
		return nil, fmt.Errorf("nothing to update")
	}
	upStms, uNames := userz.UserQuizAttemptsTable.Update(updatedCols...)
	updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&userLspMap)
	if err := updateQuery.ExecRelease(); err != nil {
		log.Errorf("error updating user quiz attempts: %v", err)
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
	return userLspOutput, nil
}
