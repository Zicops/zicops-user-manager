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

func AddUserCohort(ctx context.Context, input []*model.UserCohortInput) ([]*model.UserCohort, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	isAllowed := true
	role := strings.ToLower(userCass.Role)
	if userCass.ID == input[0].UserID || role == "admin" || strings.Contains(role, "manager") {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create org mapping")
	}
	userLspMaps := make([]*model.UserCohort, 0)
	for _, input := range input {

		if input == nil {
			continue
		}
		checkQuery := fmt.Sprintf(`SELECT * FROM userz.user_cohort_map WHERE user_id = '%s' AND cohort_id='%s' ALLOW FILTERING`, input.UserID, input.CohortID)
		checkMapping := func() (cohortMaps []userz.UserCohort, err error) {
			q := CassUserSession.Query(checkQuery, nil)
			defer q.Release()
			iter := q.Iter()
			return cohortMaps, iter.Select(&cohortMaps)
		}
		cohorts, err := checkMapping()
		if err != nil {
			return nil, err
		}
		if len(cohorts) != 0 {
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
		userLspMap := userz.UserCohort{
			ID:               uuid.New().String(),
			UserID:           input.UserID,
			UserLspID:        input.UserLspID,
			CohortID:         input.CohortID,
			AddedBy:          input.AddedBy,
			MembershipStatus: input.MembershipStatus,
			Role:             input.Role,
			CreatedAt:        time.Now().Unix(),
			UpdatedAt:        time.Now().Unix(),
			CreatedBy:        createdBy,
			UpdatedBy:        updatedBy,
		}
		insertQuery := CassUserSession.Query(userz.UserCohortTable.Insert()).BindStruct(userLspMap)
		if err := insertQuery.ExecRelease(); err != nil {
			return nil, err
		}
		created := strconv.FormatInt(userLspMap.CreatedAt, 10)
		updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
		userLspOutput := &model.UserCohort{
			UserCohortID:     &userLspMap.ID,
			UserLspID:        userLspMap.UserLspID,
			UserID:           userLspMap.UserID,
			MembershipStatus: userLspMap.MembershipStatus,
			Role:             userLspMap.Role,
			CohortID:         userLspMap.CohortID,
			AddedBy:          userLspMap.AddedBy,
			CreatedAt:        created,
			UpdatedAt:        updated,
			CreatedBy:        &userLspMap.CreatedBy,
			UpdatedBy:        &userLspMap.UpdatedBy,
		}
		userLspMaps = append(userLspMaps, userLspOutput)
	}
	return userLspMaps, nil
}

func UpdateUserCohort(ctx context.Context, input model.UserCohortInput) (*model.UserCohort, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	isAllowed := true
	role := strings.ToLower(userCass.Role)
	if userCass.ID == input.UserID || role == "admin" || strings.Contains(role, "manager") {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create cohort mapping")
	}
	if input.UserCohortID == nil {
		return nil, fmt.Errorf("user cohort id is required")
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userLspMap := userz.UserCohort{
		ID: *input.UserCohortID,
	}
	userLsps := []userz.UserCohort{}

	getQueryStr := fmt.Sprintf("SELECT * FROM userz.user_cohort_map WHERE id='%s' AND user_id='%s'  ", *input.UserCohortID, input.UserID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
	if err := getQuery.SelectRelease(&userLsps); err != nil {
		return nil, err
	}
	if len(userLsps) == 0 {
		return nil, fmt.Errorf("users cohort not found")
	}
	userLspMap = userLsps[0]
	updatedCols := []string{}
	if input.CohortID != "" {
		userLspMap.CohortID = input.CohortID
		updatedCols = append(updatedCols, "cohort_id")
	}
	if input.UpdatedBy != nil {
		userLspMap.UpdatedBy = *input.UpdatedBy
		updatedCols = append(updatedCols, "updated_by")
	}
	if input.AddedBy != "" {
		userLspMap.AddedBy = input.AddedBy
		updatedCols = append(updatedCols, "added_by")
	}
	if input.MembershipStatus != "" {
		userLspMap.MembershipStatus = input.MembershipStatus
		updatedCols = append(updatedCols, "membership_status")
	}
	if input.UserLspID != "" {
		userLspMap.UserLspID = input.UserLspID
		updatedCols = append(updatedCols, "user_lsp_id")
	}
	if input.Role != "" && userLspMap.Role != input.Role {
		userLspMap.Role = input.Role
		updatedCols = append(updatedCols, "role")
	}
	if len(updatedCols) > 0 {
		updatedAt := time.Now().Unix()
		userLspMap.UpdatedAt = updatedAt
		updatedCols = append(updatedCols, "updated_at")

		upStms, uNames := userz.UserCohortTable.Update(updatedCols...)
		updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&userLspMap)
		if err := updateQuery.ExecRelease(); err != nil {
			log.Errorf("error updating user org: %v", err)
			return nil, err
		}
	}
	created := strconv.FormatInt(userLspMap.CreatedAt, 10)
	updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
	userLspOutput := &model.UserCohort{
		UserCohortID:     &userLspMap.ID,
		UserLspID:        userLspMap.UserLspID,
		UserID:           userLspMap.UserID,
		MembershipStatus: userLspMap.MembershipStatus,
		Role:             userLspMap.Role,
		CohortID:         userLspMap.CohortID,
		AddedBy:          userLspMap.AddedBy,
		CreatedAt:        created,
		UpdatedAt:        updated,
		CreatedBy:        &userLspMap.CreatedBy,
		UpdatedBy:        &userLspMap.UpdatedBy,
	}
	return userLspOutput, nil
}
