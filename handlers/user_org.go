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
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/graph/model"
)

func AddUserOrganizationMap(ctx context.Context, input []*model.UserOrganizationMapInput) ([]*model.UserOrganizationMap, error) {
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
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userLspMaps := make([]*model.UserOrganizationMap, 0)
	for _, input := range input {
		queryStr := fmt.Sprintf(`SELECT * FROM user_org_map WHERE user_id='%s' AND org_id='%s' AND user_lsp_id='%s' ALLOW FILTERING`, input.UserID, input.OrganizationID, input.UserLspID)
		getUserOrgMap := func() (maps []userz.UserOrg, err error) {
			q := CassUserSession.Query(queryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return maps, iter.Select(&maps)
		}
		userOrgMaps, err := getUserOrgMap()
		if err != nil {
			log.Println("Got error while getting user org maps: ", err.Error())
		}
		if len(userOrgMaps) > 0 {
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
		userLspMap := userz.UserOrg{
			ID:        uuid.New().String(),
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
		insertQuery := CassUserSession.Query(userz.UserOrgTable.Insert()).BindStruct(userLspMap)
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
	role := strings.ToLower(userCass.Role)
	if userCass.ID == input.UserID || role == "admin" || strings.Contains(role, "manager") {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create org mapping")
	}
	if input.UserOrganizationID == nil {
		return nil, fmt.Errorf("user org id is required")
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userLspMap := userz.UserOrg{
		ID: *input.UserOrganizationID,
	}
	userLsps := []userz.UserOrg{}
	userID := userCass.ID
	if input.UserID != "" {
		userID = input.UserID
	}
	getQueryStr := fmt.Sprintf("SELECT * FROM userz.user_org_map WHERE id='%s' AND user_id='%s'  ", userLspMap.ID, userID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
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

	if len(updatedCols) > 0 {
		updatedAt := time.Now().Unix()
		userLspMap.UpdatedAt = updatedAt
		updatedCols = append(updatedCols, "updated_at")
		upStms, uNames := userz.UserOrgTable.Update(updatedCols...)
		updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&userLspMap)
		if err := updateQuery.ExecRelease(); err != nil {
			log.Errorf("error updating user org: %v", err)
			return nil, err
		}
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
