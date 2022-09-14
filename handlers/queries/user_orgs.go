package queries

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/helpers"
)

func GetUserOrganizations(ctx context.Context, userId string) ([]*model.UserOrganizationMap, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	if userId != "" {
		emailCreatorID = userId
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	qryStr := fmt.Sprintf(`SELECT * from userz.user_org_map where user_id='%s' ALLOW FILTERING`, emailCreatorID)
	getUsersOrgs := func() (users []userz.UserOrg, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return users, iter.Select(&users)
	}
	usersOrgs, err := getUsersOrgs()
	if err != nil {
		return nil, err
	}
	userOrgs := make([]*model.UserOrganizationMap, 0)
	for _, userOrg := range usersOrgs {
		copiedOrg := userOrg
		createdAt := strconv.FormatInt(userOrg.CreatedAt, 10)
		updatedAt := strconv.FormatInt(userOrg.UpdatedAt, 10)
		currentUserOrg := &model.UserOrganizationMap{
			UserID:             copiedOrg.UserID,
			UserOrganizationID: &copiedOrg.ID,
			OrganizationID:     copiedOrg.OrgID,
			UserLspID:          copiedOrg.UserLspID,
			OrganizationRole:   copiedOrg.OrgRole,
			IsActive:           copiedOrg.IsActive,
			EmployeeID:         copiedOrg.EmpID,
			CreatedBy:          &copiedOrg.CreatedBy,
			UpdatedBy:          &copiedOrg.UpdatedBy,
			CreatedAt:          createdAt,
			UpdatedAt:          updatedAt,
		}
		userOrgs = append(userOrgs, currentUserOrg)
	}
	return userOrgs, nil
}

func GetUserPreferences(ctx context.Context, userId string) ([]*model.UserPreference, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	if userId != "" {
		emailCreatorID = userId
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	qryStr := fmt.Sprintf(`SELECT * from userz.user_preferences where user_id='%s' ALLOW FILTERING`, emailCreatorID)
	getUsersOrgs := func() (users []userz.UserPreferences, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return users, iter.Select(&users)
	}
	usersOrgs, err := getUsersOrgs()
	if err != nil {
		return nil, err
	}
	userOrgs := make([]*model.UserPreference, 0)
	for _, userOrg := range usersOrgs {
		copiedOrg := userOrg
		createdAt := strconv.FormatInt(userOrg.CreatedAt, 10)
		updatedAt := strconv.FormatInt(userOrg.UpdatedAt, 10)
		currentUserOrg := &model.UserPreference{
			UserPreferenceID: &copiedOrg.ID,
			UserID:           copiedOrg.UserID,
			UserLspID:        copiedOrg.UserLspID,
			IsActive:         copiedOrg.IsActive,
			CreatedBy:        &copiedOrg.CreatedBy,
			UpdatedBy:        &copiedOrg.UpdatedBy,
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
			SubCategory:      copiedOrg.SubCategory,
			IsBase:           copiedOrg.IsBase,
		}
		userOrgs = append(userOrgs, currentUserOrg)
	}
	return userOrgs, nil
}

func GetUserLsps(ctx context.Context, userId string) ([]*model.UserLspMap, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	if userId != "" {
		emailCreatorID = userId
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	qryStr := fmt.Sprintf(`SELECT * from userz.user_lsp_map where user_id='%s' ALLOW FILTERING`, emailCreatorID)
	getUsersOrgs := func() (users []userz.UserLsp, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return users, iter.Select(&users)
	}
	usersOrgs, err := getUsersOrgs()
	if err != nil {
		return nil, err
	}
	userOrgs := make([]*model.UserLspMap, 0)
	for _, userOrg := range usersOrgs {
		copiedOrg := userOrg
		createdAt := strconv.FormatInt(userOrg.CreatedAt, 10)
		updatedAt := strconv.FormatInt(userOrg.UpdatedAt, 10)
		currentUserOrg := &model.UserLspMap{
			UserLspID: &copiedOrg.ID,
			UserID:    copiedOrg.UserID,
			LspID:     copiedOrg.LspID,
			Status:    copiedOrg.Status,
			CreatedBy: &copiedOrg.CreatedBy,
			UpdatedBy: &copiedOrg.UpdatedBy,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
		userOrgs = append(userOrgs, currentUserOrg)
	}
	return userOrgs, nil
}

func GetUserOrgDetails(ctx context.Context, userID string, lspID string) (*model.UserOrganizationMap, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	qryStr := fmt.Sprintf(`SELECT * from userz.user_org_map where user_id='%s' and user_lsp_id='%s'  ALLOW FILTERING`, userID, lspID)
	getUsersOrgs := func() (users []userz.UserOrg, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return users, iter.Select(&users)
	}
	usersOrgs, err := getUsersOrgs()
	if err != nil {
		return nil, err
	}
	if len(usersOrgs) == 0 {
		return nil, fmt.Errorf("no user org found")
	}
	userOrgs := make([]*model.UserOrganizationMap, 0)
	for _, userOrg := range usersOrgs {
		copiedOrg := userOrg
		createdAt := strconv.FormatInt(userOrg.CreatedAt, 10)
		updatedAt := strconv.FormatInt(userOrg.UpdatedAt, 10)
		currentUserOrg := &model.UserOrganizationMap{
			UserID:             copiedOrg.UserID,
			UserOrganizationID: &copiedOrg.ID,
			OrganizationID:     copiedOrg.OrgID,
			UserLspID:          copiedOrg.UserLspID,
			OrganizationRole:   copiedOrg.OrgRole,
			IsActive:           copiedOrg.IsActive,
			EmployeeID:         copiedOrg.EmpID,
			CreatedBy:          &copiedOrg.CreatedBy,
			UpdatedBy:          &copiedOrg.UpdatedBy,
			CreatedAt:          createdAt,
			UpdatedAt:          updatedAt,
		}
		userOrgs = append(userOrgs, currentUserOrg)
	}
	return userOrgs[0], nil
}

func GetUserPreferenceForLsp(ctx context.Context, userID string, lspID string) (*model.UserPreference, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	qryStr := fmt.Sprintf(`SELECT * from userz.user_preferences where user_id='%s' and user_lsp_id='%s' ALLOW FILTERING`, userID, lspID)
	getUsersOrgs := func() (users []userz.UserPreferences, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return users, iter.Select(&users)
	}
	usersOrgs, err := getUsersOrgs()
	if err != nil {
		return nil, err
	}
	if len(usersOrgs) == 0 {
		return nil, fmt.Errorf("no user lsp preference found")
	}
	userOrgs := make([]*model.UserPreference, 0)
	for _, userOrg := range usersOrgs {
		copiedOrg := userOrg
		createdAt := strconv.FormatInt(userOrg.CreatedAt, 10)
		updatedAt := strconv.FormatInt(userOrg.UpdatedAt, 10)
		currentUserOrg := &model.UserPreference{
			UserPreferenceID: &copiedOrg.ID,
			UserID:           copiedOrg.UserID,
			UserLspID:        copiedOrg.UserLspID,
			IsActive:         copiedOrg.IsActive,
			CreatedBy:        &copiedOrg.CreatedBy,
			UpdatedBy:        &copiedOrg.UpdatedBy,
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
			SubCategory:      copiedOrg.SubCategory,
			IsBase:           copiedOrg.IsBase,
		}
		userOrgs = append(userOrgs, currentUserOrg)
	}
	return userOrgs[0], nil
}

func GetUserLspByLspID(ctx context.Context, userID string, lspID string) (*model.UserLspMap, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	qryStr := fmt.Sprintf(`SELECT * from userz.user_lsp_map where user_id='%s' and lsp_id='%s' ALLOW FILTERING`, userID, lspID)
	getUsersOrgs := func() (users []userz.UserLsp, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return users, iter.Select(&users)
	}
	usersOrgs, err := getUsersOrgs()
	if err != nil {
		return nil, err
	}
	if len(usersOrgs) == 0 {
		return nil, fmt.Errorf("no user lsp found")
	}
	userOrgs := make([]*model.UserLspMap, 0)
	for _, userOrg := range usersOrgs {
		copiedOrg := userOrg
		createdAt := strconv.FormatInt(userOrg.CreatedAt, 10)
		updatedAt := strconv.FormatInt(userOrg.UpdatedAt, 10)
		currentUserOrg := &model.UserLspMap{
			UserLspID: &copiedOrg.ID,
			UserID:    copiedOrg.UserID,
			LspID:     copiedOrg.LspID,
			Status:    copiedOrg.Status,
			CreatedBy: &copiedOrg.CreatedBy,
			UpdatedBy: &copiedOrg.UpdatedBy,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
		userOrgs = append(userOrgs, currentUserOrg)
	}
	return userOrgs[0], nil
}
