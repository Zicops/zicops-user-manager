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
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
)

func AddUserExamAttempts(ctx context.Context, input []*model.UserExamAttemptsInput) ([]*model.UserExamAttempts, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	isAllowed := false
	if userCass.ID == input[0].UserID || strings.ToLower(userCass.Role) == "admin" {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create exams mapping")
	}
	userLspMaps := make([]*model.UserExamAttempts, 0)
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
		examAttemptStart, _ := strconv.ParseInt(input.AttemptStartTime, 10, 64)
		examAttempDuration, _ := strconv.ParseInt(input.AttemptDuration, 10, 64)
		userLspMap := userz.UserExamAttempts{
			ID:               guid.String(),
			UserID:           input.UserID,
			UserLspID:        input.UserLspID,
			UserCpID:         input.UserCpID,
			UserCmID:         input.UserCourseID,
			ExamID:           input.ExamID,
			AttemptNo:        int64(input.AttemptNo),
			AttemptStatus:    input.AttemptStatus,
			AttemptStartTime: examAttemptStart,
			AttemptDuration:  examAttempDuration,
			CreatedAt:        time.Now().Unix(),
			UpdatedAt:        time.Now().Unix(),
			CreatedBy:        createdBy,
			UpdatedBy:        updatedBy,
		}
		insertQuery := global.CassUserSession.Session.Query(userz.UserExamAttemptsTable.Insert()).BindStruct(userLspMap)
		if err := insertQuery.ExecRelease(); err != nil {
			return nil, err
		}
		created := strconv.FormatInt(userLspMap.CreatedAt, 10)
		updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
		userLspOutput := &model.UserExamAttempts{
			UserEaID:         &userLspMap.ID,
			UserID:           userLspMap.UserID,
			UserLspID:        userLspMap.UserLspID,
			UserCourseID:     userLspMap.UserCmID,
			UserCpID:         userLspMap.UserCpID,
			ExamID:           userLspMap.ExamID,
			AttemptNo:        int(userLspMap.AttemptNo),
			AttemptStatus:    userLspMap.AttemptStatus,
			AttemptStartTime: input.AttemptStartTime,
			AttemptDuration:  input.AttemptDuration,
			CreatedAt:        created,
			UpdatedAt:        updated,
			CreatedBy:        &userLspMap.CreatedBy,
			UpdatedBy:        &userLspMap.UpdatedBy,
		}
		userLspMaps = append(userLspMaps, userLspOutput)
	}
	return userLspMaps, nil
}

func UpdateUserExamAttempts(ctx context.Context, input model.UserExamAttemptsInput) (*model.UserExamAttempts, error) {
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
	if input.UserEaID == nil {
		return nil, fmt.Errorf("user eq id is required")
	}
	userLspMap := userz.UserExamAttempts{
		ID: *input.UserEaID,
	}
	userLsps := []userz.UserExamAttempts{}
	getQuery := global.CassUserSession.Session.Query(userz.UserExamAttemptsTable.Get()).BindMap(qb.M{"id": userLspMap.ID})
	if err := getQuery.SelectRelease(&userLsps); err != nil {
		return nil, err
	}
	if len(userLsps) == 0 {
		return nil, fmt.Errorf("users exams not found")
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
	if input.ExamID != "" && input.ExamID != userLspMap.ExamID {
		userLspMap.ExamID = input.ExamID
		updatedCols = append(updatedCols, "exam_id")
	}
	if input.AttemptNo != 0 && int64(input.AttemptNo) != userLspMap.AttemptNo {
		userLspMap.AttemptNo = int64(input.AttemptNo)
		updatedCols = append(updatedCols, "attempt_no")
	}
	if input.AttemptStatus != "" && input.AttemptStatus != userLspMap.AttemptStatus {
		userLspMap.AttemptStatus = input.AttemptStatus
		updatedCols = append(updatedCols, "attempt_status")
	}
	attemptStartTime, _ := strconv.ParseInt(input.AttemptStartTime, 10, 64)
	attemptDuration, _ := strconv.ParseInt(input.AttemptDuration, 10, 64)
	if input.AttemptStartTime != "" && attemptStartTime != userLspMap.AttemptStartTime {
		userLspMap.AttemptStartTime = attemptStartTime
		updatedCols = append(updatedCols, "attempt_start_time")
	}
	if input.AttemptDuration != "" && attemptDuration != userLspMap.AttemptDuration {
		userLspMap.AttemptDuration = attemptDuration
		updatedCols = append(updatedCols, "attempt_duration")
	}
	if input.UserLspID != "" && input.UserLspID != userLspMap.UserLspID {
		userLspMap.UserLspID = input.UserLspID
		updatedCols = append(updatedCols, "user_lsp_id")
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
	upStms, uNames := userz.UserExamAttemptsTable.Update(updatedCols...)
	updateQuery := global.CassUserSession.Session.Query(upStms, uNames).BindStruct(&userLspMap)
	if err := updateQuery.ExecRelease(); err != nil {
		log.Errorf("error updating user exam attempts: %v", err)
		return nil, err
	}
	created := strconv.FormatInt(userLspMap.CreatedAt, 10)
	updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
	userLspOutput := &model.UserExamAttempts{
		UserEaID:         &userLspMap.ID,
		UserID:           userLspMap.UserID,
		UserLspID:        userLspMap.UserLspID,
		UserCourseID:     userLspMap.UserCmID,
		UserCpID:         userLspMap.UserCpID,
		ExamID:           userLspMap.ExamID,
		AttemptNo:        int(userLspMap.AttemptNo),
		AttemptStatus:    userLspMap.AttemptStatus,
		AttemptStartTime: input.AttemptStartTime,
		AttemptDuration:  input.AttemptDuration,
		CreatedAt:        created,
		UpdatedAt:        updated,
		CreatedBy:        &userLspMap.CreatedBy,
		UpdatedBy:        &userLspMap.UpdatedBy,
	}
	return userLspOutput, nil
}