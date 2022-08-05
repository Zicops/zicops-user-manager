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

func AddUserOrganizationMap(ctx context.Context, input []*model.UserOrganizationMapInput) ([]*model.UserOrganizationMap, error) {
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
	userLspMaps := make([]*model.UserOrganizationMap, 0)
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
		userLspMap := userz.UserOrg{
			ID:        guid.String(),
			UserID:    input.UserID,
			UserLspID: input.UserLspID,
			OrgID:     input.OrganizationID,
			OrgRole:   input.OrganizationRole,
			IsActive:  input.IsActive,
			EmpID:     input.EmployeeID,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		}
		insertQuery := global.CassUserSession.Session.Query(userz.UserOrgTable.Insert()).BindStruct(userLspMap)
		if err := insertQuery.ExecRelease(); err != nil {
			return nil, err
		}
		created := strconv.FormatInt(userLspMap.CreatedAt, 10)
		updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
		userLspOutput := &model.UserOrganizationMap{
			UserOrganizationID: &userLspMap.ID,
			UserLspID:          userLspMap.UserLspID,
			UserID:             userLspMap.UserID,
			OrganizationID:     userLspMap.OrgID,
			OrganizationRole:   userLspMap.OrgRole,
			IsActive:           userLspMap.IsActive,
			EmployeeID:         userLspMap.EmpID,
			CreatedAt:          created,
			UpdatedAt:          updated,
			CreatedBy:          &userLspMap.CreatedBy,
			UpdatedBy:          &userLspMap.UpdatedBy,
		}
		userLspMaps = append(userLspMaps, userLspOutput)
	}
	return userLspMaps, nil
}

func UpdateUserOrganizationMap(ctx context.Context, input model.UserOrganizationMapInput) (*model.UserOrganizationMap, error) {
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
	if input.UserOrganizationID == nil {
		return nil, fmt.Errorf("user org id is required")
	}
	userLspMap := userz.UserOrg{
		ID: *input.UserOrganizationID,
	}
	userLsps := []userz.UserOrg{}
	getQuery := global.CassUserSession.Session.Query(userz.UserOrgTable.Get()).BindMap(qb.M{"id": userLspMap.ID})
	if err := getQuery.SelectRelease(&userLsps); err != nil {
		return nil, err
	}
	if len(userLsps) == 0 {
		return nil, fmt.Errorf("users orgs not found")
	}
	userLspMap = userLsps[0]
	updatedCols := []string{}
	if input.EmployeeID != "" {
		userLspMap.EmpID = input.EmployeeID
		updatedCols = append(updatedCols, "emp_id")
	}
	if input.UpdatedBy != nil {
		userLspMap.UpdatedBy = *input.UpdatedBy
		updatedCols = append(updatedCols, "updated_by")
	}
	if input.OrganizationID != "" {
		userLspMap.OrgID = input.OrganizationID
		updatedCols = append(updatedCols, "org_id")
	}
	if input.OrganizationRole != "" {
		userLspMap.OrgRole = input.OrganizationRole
		updatedCols = append(updatedCols, "org_role")
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
	upStms, uNames := userz.UserOrgTable.Update(updatedCols...)
	updateQuery := global.CassUserSession.Session.Query(upStms, uNames).BindStruct(&userLspMap)
	if err := updateQuery.ExecRelease(); err != nil {
		log.Errorf("error updating user org: %v", err)
		return nil, err
	}
	created := strconv.FormatInt(userLspMap.CreatedAt, 10)
	updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
	userLspOutput := &model.UserOrganizationMap{
		UserOrganizationID: &userLspMap.ID,
		UserLspID:          userLspMap.UserLspID,
		UserID:             userLspMap.UserID,
		OrganizationID:     userLspMap.OrgID,
		OrganizationRole:   userLspMap.OrgRole,
		IsActive:           userLspMap.IsActive,
		EmployeeID:         userLspMap.EmpID,
		CreatedAt:          created,
		UpdatedAt:          updated,
		CreatedBy:          &userLspMap.CreatedBy,
		UpdatedBy:          &userLspMap.UpdatedBy,
	}
	return userLspOutput, nil
}
