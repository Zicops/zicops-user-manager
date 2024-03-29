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

func AddUserExamResult(ctx context.Context, input []*model.UserExamResultInput) ([]*model.UserExamResult, error) {
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
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userLspMaps := make([]*model.UserExamResult, 0)
	for _, res := range input {
		if res == nil {
			continue
		}
		input := res

		queryStr := fmt.Sprintf(`SELECT * FROM userz.user_exam_results WHERE user_id='%s' AND user_ea_id='%s' ALLOW FILTERING`, input.UserID, input.UserEaID)
		checkMapping := func() (examResults []userz.UserExamResults, err error) {
			q := CassUserSession.Query(queryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return examResults, iter.Select(&examResults)
		}
		userExamResults, err := checkMapping()
		if err != nil {
			return nil, err
		}
		if len(userExamResults) != 0 {
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
		userLspMap := userz.UserExamResults{
			ID:             uuid.New().String(),
			UserID:         input.UserID,
			UserEaID:       input.UserEaID,
			UserScore:      int64(input.UserScore),
			CorrectAnswers: int64(input.CorrectAnswers),
			WrongAnswers:   int64(input.WrongAnswers),
			ResultStatus:   input.ResultStatus,
			CreatedAt:      time.Now().Unix(),
			UpdatedAt:      time.Now().Unix(),
			CreatedBy:      createdBy,
			UpdatedBy:      updatedBy,
		}
		insertQuery := CassUserSession.Query(userz.UserExamResultsTable.Insert()).BindStruct(userLspMap)
		if err := insertQuery.ExecRelease(); err != nil {
			return nil, err
		}
		created := strconv.FormatInt(userLspMap.CreatedAt, 10)
		updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
		userLspOutput := &model.UserExamResult{
			UserErID:       &userLspMap.ID,
			UserID:         userLspMap.UserID,
			UserEaID:       userLspMap.UserEaID,
			UserScore:      input.UserScore,
			CorrectAnswers: input.CorrectAnswers,
			WrongAnswers:   input.WrongAnswers,
			ResultStatus:   userLspMap.ResultStatus,
			CreatedAt:      created,
			UpdatedAt:      updated,
			CreatedBy:      &userLspMap.CreatedBy,
			UpdatedBy:      &userLspMap.UpdatedBy,
		}
		userLspMaps = append(userLspMaps, userLspOutput)
	}
	return userLspMaps, nil
}

func UpdateUserExamResult(ctx context.Context, input model.UserExamResultInput) (*model.UserExamResult, error) {
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
	if input.UserErID == nil {
		return nil, fmt.Errorf("user er id is required")
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userLspMap := userz.UserExamResults{
		ID: *input.UserErID,
	}
	userLsps := []userz.UserExamResults{}
	userID := userCass.ID
	if input.UserID != "" {
		userID = input.UserID
	}
	getQueryStr := fmt.Sprintf("SELECT * FROM userz.user_exam_results WHERE id='%s' AND user_id='%s'  ", userLspMap.ID, userID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
	if err := getQuery.SelectRelease(&userLsps); err != nil {
		return nil, err
	}
	if len(userLsps) == 0 {
		return nil, fmt.Errorf("users exams not found")
	}
	userLspMap = userLsps[0]
	updatedCols := []string{}
	if input.UserEaID != "" && input.UserEaID != userLspMap.UserEaID {
		userLspMap.UserEaID = input.UserEaID
		updatedCols = append(updatedCols, "user_ea_id")
	}
	if int64(input.UserScore) != userLspMap.UserScore {
		userLspMap.UserScore = int64(input.UserScore)
		updatedCols = append(updatedCols, "user_score")
	}
	if int64(input.CorrectAnswers) != userLspMap.CorrectAnswers {
		userLspMap.CorrectAnswers = int64(input.CorrectAnswers)
		updatedCols = append(updatedCols, "correct_answers")
	}
	if int64(input.WrongAnswers) != userLspMap.WrongAnswers {
		userLspMap.WrongAnswers = int64(input.WrongAnswers)
		updatedCols = append(updatedCols, "wrong_answers")
	}
	if input.ResultStatus != "" && input.ResultStatus != userLspMap.ResultStatus {
		userLspMap.ResultStatus = input.ResultStatus
		updatedCols = append(updatedCols, "result_status")
	}
	if input.UpdatedBy != nil {
		userLspMap.UpdatedBy = *input.UpdatedBy
		updatedCols = append(updatedCols, "updated_by")
	}
	if len(updatedCols) > 0 {
		updatedAt := time.Now().Unix()
		userLspMap.UpdatedAt = updatedAt
		updatedCols = append(updatedCols, "updated_at")
		upStms, uNames := userz.UserExamResultsTable.Update(updatedCols...)
		updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&userLspMap)
		if err := updateQuery.ExecRelease(); err != nil {
			log.Errorf("error updating user exam results: %v", err)
			return nil, err
		}
	}
	created := strconv.FormatInt(userLspMap.CreatedAt, 10)
	updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
	userLspOutput := &model.UserExamResult{
		UserErID:       &userLspMap.ID,
		UserID:         userLspMap.UserID,
		UserEaID:       userLspMap.UserEaID,
		UserScore:      input.UserScore,
		CorrectAnswers: input.CorrectAnswers,
		WrongAnswers:   input.WrongAnswers,
		ResultStatus:   userLspMap.ResultStatus,
		CreatedAt:      created,
		UpdatedAt:      updated,
		CreatedBy:      &userLspMap.CreatedBy,
		UpdatedBy:      &userLspMap.UpdatedBy,
	}
	return userLspOutput, nil
}
