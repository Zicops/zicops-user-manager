package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/graph/model"
)

func AddUserExamProgress(ctx context.Context, input []*model.UserExamProgressInput) ([]*model.UserExamProgress, error) {
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
		return nil, fmt.Errorf("user not allowed to create exams mapping")
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userLspMaps := make([]*model.UserExamProgress, 0)
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
		totalTimeSpent, _ := strconv.ParseInt(input.TotalTimeSpent, 10, 64)
		userLspMap := userz.UserExamProgress{
			ID:             guid.String(),
			UserID:         input.UserID,
			UserLspID:      input.UserLspID,
			UserCpID:       input.UserCpID,
			UserEaID:       input.UserEaID,
			SrNo:           int64(input.SrNo),
			QuestionID:     input.QuestionID,
			QuestionType:   input.QuestionType,
			Answer:         input.Answer,
			QAttemptStatus: input.QAttemptStatus,
			TotalTimeSpent: totalTimeSpent,
			CorrectAnswer:  input.CorrectAnswer,
			SectionID:      input.SectionID,
			CreatedAt:      time.Now().Unix(),
			UpdatedAt:      time.Now().Unix(),
			CreatedBy:      createdBy,
			UpdatedBy:      updatedBy,
		}
		insertQuery := CassUserSession.Query(userz.UserExamProgressTable.Insert()).BindStruct(userLspMap)
		if err := insertQuery.ExecRelease(); err != nil {
			return nil, err
		}
		created := strconv.FormatInt(userLspMap.CreatedAt, 10)
		updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
		userLspOutput := &model.UserExamProgress{
			UserEpID:       &userLspMap.ID,
			UserID:         userLspMap.UserID,
			UserLspID:      userLspMap.UserLspID,
			UserCpID:       userLspMap.UserCpID,
			UserEaID:       userLspMap.UserEaID,
			SrNo:           input.SrNo,
			QuestionID:     userLspMap.QuestionID,
			QuestionType:   userLspMap.QuestionType,
			Answer:         userLspMap.Answer,
			QAttemptStatus: userLspMap.QAttemptStatus,
			TotalTimeSpent: input.TotalTimeSpent,
			CorrectAnswer:  userLspMap.CorrectAnswer,
			SectionID:      userLspMap.SectionID,
			CreatedAt:      created,
			UpdatedAt:      updated,
			CreatedBy:      &userLspMap.CreatedBy,
			UpdatedBy:      &userLspMap.UpdatedBy,
		}
		userLspMaps = append(userLspMaps, userLspOutput)
	}
	return userLspMaps, nil
}

func UpdateUserExamProgress(ctx context.Context, input model.UserExamProgressInput) (*model.UserExamProgress, error) {
	userCass, err := GetUserFromCass(ctx)
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
	if userCass.ID == input.UserID || role == "admin" || strings.Contains(role, "manager") {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create org mapping")
	}
	if input.UserEpID == nil {
		return nil, fmt.Errorf("user ep id is required")
	}
	userLspMap := userz.UserExamProgress{
		ID: *input.UserEpID,
	}
	userLsps := []userz.UserExamProgress{}
	userID := userCass.ID
	if input.UserID != "" {
		userID = input.UserID
	}
	getQueryStr := fmt.Sprintf("SELECT * FROM userz.user_exam_progress WHERE id='%s' AND user_id='%s'  ", userLspMap.ID, userID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
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
	if input.UserEaID != "" && input.UserEaID != userLspMap.UserEaID {
		userLspMap.UserEaID = input.UserEaID
		updatedCols = append(updatedCols, "user_ea_id")
	}
	if input.SrNo != 0 && int64(input.SrNo) != userLspMap.SrNo {
		userLspMap.SrNo = int64(input.SrNo)
		updatedCols = append(updatedCols, "sr_no")
	}
	if input.QuestionID != "" && input.QuestionID != userLspMap.QuestionID {
		userLspMap.QuestionID = input.QuestionID
		updatedCols = append(updatedCols, "question_id")
	}
	if input.QuestionType != "" && input.QuestionType != userLspMap.QuestionType {
		userLspMap.QuestionType = input.QuestionType
		updatedCols = append(updatedCols, "question_type")
	}
	if input.Answer != "" && input.Answer != userLspMap.Answer {
		userLspMap.Answer = input.Answer
		updatedCols = append(updatedCols, "answer")
	}
	if input.QAttemptStatus != "" && input.QAttemptStatus != userLspMap.QAttemptStatus {
		userLspMap.QAttemptStatus = input.QAttemptStatus
		updatedCols = append(updatedCols, "q_attempt_status")
	}
	totalTimeSpent, _ := strconv.ParseInt(input.TotalTimeSpent, 10, 64)
	if input.TotalTimeSpent != "" && totalTimeSpent != userLspMap.TotalTimeSpent {
		userLspMap.TotalTimeSpent = totalTimeSpent
		updatedCols = append(updatedCols, "total_time_spent")
	}
	if input.CorrectAnswer != "" && input.CorrectAnswer != userLspMap.CorrectAnswer {
		userLspMap.CorrectAnswer = input.CorrectAnswer
		updatedCols = append(updatedCols, "correct_answer")
	}
	if input.SectionID != "" && input.SectionID != userLspMap.SectionID {
		userLspMap.SectionID = input.SectionID
		updatedCols = append(updatedCols, "section_id")
	}
	if input.UserLspID != "" && input.UserLspID != userLspMap.UserLspID {
		userLspMap.UserLspID = input.UserLspID
		updatedCols = append(updatedCols, "user_lsp_id")
	}
	if input.UpdatedBy != nil {
		userLspMap.UpdatedBy = *input.UpdatedBy
		updatedCols = append(updatedCols, "updated_by")
	}
	if len(updatedCols) > 0 {
		updatedAt := time.Now().Unix()
		userLspMap.UpdatedAt = updatedAt
		updatedCols = append(updatedCols, "updated_at")
		upStms, uNames := userz.UserExamProgressTable.Update(updatedCols...)
		updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&userLspMap)
		if err := updateQuery.ExecRelease(); err != nil {
			log.Errorf("error updating user exam progress: %v", err)
			return nil, err
		}
	}
	created := strconv.FormatInt(userLspMap.CreatedAt, 10)
	updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
	userLspOutput := &model.UserExamProgress{
		UserEpID:       &userLspMap.ID,
		UserID:         userLspMap.UserID,
		UserLspID:      userLspMap.UserLspID,
		UserCpID:       userLspMap.UserCpID,
		UserEaID:       userLspMap.UserEaID,
		SrNo:           input.SrNo,
		QuestionID:     userLspMap.QuestionID,
		QuestionType:   userLspMap.QuestionType,
		Answer:         userLspMap.Answer,
		QAttemptStatus: userLspMap.QAttemptStatus,
		TotalTimeSpent: input.TotalTimeSpent,
		CorrectAnswer:  userLspMap.CorrectAnswer,
		SectionID:      userLspMap.SectionID,
		CreatedAt:      created,
		UpdatedAt:      updated,
		CreatedBy:      &userLspMap.CreatedBy,
		UpdatedBy:      &userLspMap.UpdatedBy,
	}
	return userLspOutput, nil
}
