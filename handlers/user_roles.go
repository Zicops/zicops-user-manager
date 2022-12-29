package handlers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/graph/model"
)

func AddUserRoles(ctx context.Context, input []*model.UserRoleInput) ([]*model.UserRole, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	isAllowed := true
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create org mapping")
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userLspMaps := make([]*model.UserRole, 0)
	for _, input := range input {

		createdBy := userCass.Email
		updatedBy := userCass.Email
		if input.CreatedBy != nil {
			createdBy = *input.CreatedBy
		}
		if input.UpdatedBy != nil {
			updatedBy = *input.UpdatedBy
		}
		userLspMap := userz.UserRole{
			ID:        uuid.New().String(),
			UserID:    input.UserID,
			UserLspID: input.UserLspID,
			Role:      input.Role,
			IsActive:  input.IsActive,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		}
		insertQuery := CassUserSession.Query(userz.UserRoleTable.Insert()).BindStruct(userLspMap)
		if err := insertQuery.ExecRelease(); err != nil {
			return nil, err
		}
		created := strconv.FormatInt(userLspMap.CreatedAt, 10)
		updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
		userLspOutput := &model.UserRole{
			UserRoleID: &userLspMap.ID,
			UserLspID:  userLspMap.UserLspID,
			UserID:     userLspMap.UserID,
			Role:       userLspMap.Role,
			IsActive:   userLspMap.IsActive,
			CreatedAt:  created,
			UpdatedAt:  updated,
			CreatedBy:  &userLspMap.CreatedBy,
			UpdatedBy:  &userLspMap.UpdatedBy,
		}
		userLspMaps = append(userLspMaps, userLspOutput)
	}
	return userLspMaps, nil
}

func UpdateUserRole(ctx context.Context, input model.UserRoleInput) (*model.UserRole, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	isAllowed := true
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create org mapping")
	}
	if input.UserRoleID == nil {
		return nil, fmt.Errorf("user role id is required")
	}
	userLspMap := userz.UserRole{
		ID: *input.UserRoleID,
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userLsps := []userz.UserRole{}
	userID := userCass.ID
	if input.UserID != "" {
		userID = input.UserID
	}
	getQueryStr := fmt.Sprintf("SELECT * FROM userz.user_role WHERE id='%s' AND user_id='%s'  ", userLspMap.ID, userID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
	if err := getQuery.SelectRelease(&userLsps); err != nil {
		return nil, err
	}
	if len(userLsps) == 0 {
		return nil, fmt.Errorf("users orgs not found")
	}
	userLspMap = userLsps[0]
	updatedCols := []string{}
	if input.Role != "" {
		userLspMap.Role = input.Role
		updatedCols = append(updatedCols, "role")
	}
	if input.UpdatedBy != nil {
		userLspMap.UpdatedBy = *input.UpdatedBy
		updatedCols = append(updatedCols, "updated_by")
	}
	if input.IsActive != userLspMap.IsActive {
		userLspMap.IsActive = input.IsActive
		updatedCols = append(updatedCols, "is_active")
	}
	if input.UserLspID != "" {
		userLspMap.UserLspID = input.UserLspID
		updatedCols = append(updatedCols, "user_lsp_id")
	}
	if len(updatedCols) > 0 {
		updatedAt := time.Now().Unix()
		userLspMap.UpdatedAt = updatedAt
		updatedCols = append(updatedCols, "updated_at")
		upStms, uNames := userz.UserRoleTable.Update(updatedCols...)
		updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&userLspMap)
		if err := updateQuery.ExecRelease(); err != nil {
			log.Errorf("error updating user org: %v", err)
			return nil, err
		}
	}
	created := strconv.FormatInt(userLspMap.CreatedAt, 10)
	updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
	userLspOutput := &model.UserRole{
		UserRoleID: &userLspMap.ID,
		UserLspID:  userLspMap.UserLspID,
		UserID:     userLspMap.UserID,
		Role:       userLspMap.Role,
		IsActive:   userLspMap.IsActive,
		CreatedAt:  created,
		UpdatedAt:  updated,
		CreatedBy:  &userLspMap.CreatedBy,
		UpdatedBy:  &userLspMap.UpdatedBy,
	}
	return userLspOutput, nil
}

func GetUserLspRoles(ctx context.Context, userID string, userLspIds []string) ([]*model.UserRole, error) {
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	userLspMaps := make([]*model.UserRole, 0)
	for _, input := range userLspIds {
		userLspMap := userz.UserRole{
			UserID:    userID,
			UserLspID: input,
		}
		userLsps := []userz.UserRole{}
		getQueryStr := fmt.Sprintf("SELECT * FROM userz.user_role WHERE user_id='%s' AND user_lsp_id='%s'  ALLOW FILTERING", userLspMap.UserID, userLspMap.UserLspID)
		getQuery := CassUserSession.Query(getQueryStr, nil)
		if err := getQuery.SelectRelease(&userLsps); err != nil {
			return nil, err
		}
		if len(userLsps) == 0 {
			continue
		}
		for _, usrLspRoleCopy := range userLsps {
			userLspMap := usrLspRoleCopy
			created := strconv.FormatInt(userLspMap.CreatedAt, 10)
			updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
			userLspOutput := &model.UserRole{
				UserRoleID: &userLspMap.ID,
				UserLspID:  userLspMap.UserLspID,
				UserID:     userLspMap.UserID,
				Role:       userLspMap.Role,
				IsActive:   userLspMap.IsActive,
				CreatedAt:  created,
				UpdatedAt:  updated,
				CreatedBy:  &userLspMap.CreatedBy,
				UpdatedBy:  &userLspMap.UpdatedBy,
			}
			userLspMaps = append(userLspMaps, userLspOutput)
		}
	}
	return userLspMaps, nil
}
