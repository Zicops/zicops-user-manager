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

func AddUserRoles(ctx context.Context, input []*model.UserRoleInput) ([]*model.UserRole, error) {
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
	userLspMaps := make([]*model.UserRole, 0)
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
		userLspMap := userz.UserRole{
			ID:        guid.String(),
			UserID:    input.UserID,
			UserLspID: input.UserLspID,
			Role:      input.Role,
			IsActive:  input.IsActive,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		}
		insertQuery := global.CassUserSession.Session.Query(userz.UserRoleTable.Insert()).BindStruct(userLspMap)
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
	isAllowed := false
	if userCass.ID == input.UserID || strings.ToLower(userCass.Role) == "admin" {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create org mapping")
	}
	if input.UserRoleID == nil {
		return nil, fmt.Errorf("user role id is required")
	}
	userLspMap := userz.UserRole{
		ID: *input.UserRoleID,
	}
	userLsps := []userz.UserRole{}
	getQuery := global.CassUserSession.Session.Query(userz.UserRoleTable.Get()).BindMap(qb.M{"id": userLspMap.ID})
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
	updatedAt := time.Now().Unix()
	userLspMap.UpdatedAt = updatedAt
	updatedCols = append(updatedCols, "updated_at")
	if len(updatedCols) == 0 {
		return nil, fmt.Errorf("nothing to update")
	}
	upStms, uNames := userz.UserRoleTable.Update(updatedCols...)
	updateQuery := global.CassUserSession.Session.Query(upStms, uNames).BindStruct(&userLspMap)
	if err := updateQuery.ExecRelease(); err != nil {
		log.Errorf("error updating user org: %v", err)
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
	return userLspOutput, nil
}
