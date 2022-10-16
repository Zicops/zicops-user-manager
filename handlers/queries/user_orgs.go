package queries

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/global"
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
	//key := "GetUserOrganizations" + emailCreatorID
	//result, err := redis.GetRedisValue(key)
	//if err == nil {
	//	var userOrgs []*model.UserOrganizationMap
	//	err = json.Unmarshal([]byte(result), &userOrgs)
	//	if err == nil {
	//		return userOrgs, nil
	//	}
	//}

	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	createdAt := time.Now().Unix()
	qryStr := fmt.Sprintf(`SELECT * from userz.user_org_map where user_id='%s' AND created_at < %d  ALLOW FILTERING`, emailCreatorID, createdAt)
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
	//redisBytes, err := json.Marshal(userOrgs)
	//if err == nil {
	//	redis.SetTTL(key, 3600)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
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
	//key := "GetUserPreferences" + emailCreatorID
	//result, err := redis.GetRedisValue(key)
	//if err == nil {
	//	var userPreferences []*model.UserPreference
	//	err = json.Unmarshal([]byte(result), &userPreferences)
	//	if err == nil {
	//		return userPreferences, nil
	//	}
	//}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	createdAt := time.Now().Unix()
	qryStr := fmt.Sprintf(`SELECT * from userz.user_preferences where user_id='%s' AND created_at < %d ALLOW FILTERING`, emailCreatorID, createdAt)
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
	//redisBytes, err := json.Marshal(userOrgs)
	//if err == nil {
	//	redis.SetTTL(key, 3600)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
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
	//key := "GetUserLsps" + emailCreatorID
	//result, err := redis.GetRedisValue(key)
	//if err == nil {
	//	var userLsps []*model.UserLspMap
	//	err = json.Unmarshal([]byte(result), &userLsps)
	//	if err == nil {
	//		return userLsps, nil
	//	}
	//}

	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	createdAt := time.Now().Unix()
	qryStr := fmt.Sprintf(`SELECT * from userz.user_lsp_map where user_id='%s' AND created_at < %d ALLOW FILTERING`, emailCreatorID, createdAt)
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
			LspID:     copiedOrg.LspId,
			Status:    copiedOrg.Status,
			CreatedBy: &copiedOrg.CreatedBy,
			UpdatedBy: &copiedOrg.UpdatedBy,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
		userOrgs = append(userOrgs, currentUserOrg)
	}
	//redisBytes, err := json.Marshal(userOrgs)
	//if err == nil {
	//	redis.SetTTL(key, 3600)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}

	return userOrgs, nil
}

func GetUserLspMapsByLspID(ctx context.Context, lspID string, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedUserLspMaps, error) {
	var newPage []byte
	//var pageDirection string
	var pageSizeInt int
	if pageCursor != nil && *pageCursor != "" {
		page, err := global.CryptSession.DecryptString(*pageCursor, nil)
		if err != nil {
			return nil, fmt.Errorf("invalid page cursor: %v", err)
		}
		newPage = page
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	createdAt := time.Now().Unix()
	qryStr := fmt.Sprintf(`SELECT * from userz.user_lsp_map where lsp_id='%s' AND created_at < %d ALLOW FILTERING`, lspID, createdAt)
	getUsers := func(page []byte) (users []userz.UserLsp, nextPage []byte, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		q.PageState(page)
		q.PageSize(pageSizeInt)

		iter := q.Iter()
		return users, iter.PageState(), iter.Select(&users)
	}
	usersOrgs, newPage, err := getUsers(newPage)
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
			LspID:     copiedOrg.LspId,
			Status:    copiedOrg.Status,
			CreatedBy: &copiedOrg.CreatedBy,
			UpdatedBy: &copiedOrg.UpdatedBy,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
		userOrgs = append(userOrgs, currentUserOrg)
	}
	var outputResponse model.PaginatedUserLspMaps
	var newCursor string
	if len(newPage) != 0 {
		newCursor, err = global.CryptSession.EncryptAsString(newPage, nil)
		if err != nil {
			return nil, fmt.Errorf("error encrypting cursor: %v", err)
		}
		log.Infof("Users: %v", string(newCursor))

	}
	outputResponse.UserLspMaps = userOrgs
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction

	return &outputResponse, nil
}

func GetUserOrgDetails(ctx context.Context, userID string, lspID string) (*model.UserOrganizationMap, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	//key := "GetUserOrgDetails" + userID + lspID
	//result, err := redis.GetRedisValue(key)
	//if err == nil {
	//	var userOrgDetails *model.UserOrganizationMap
	//	err = json.Unmarshal([]byte(result), &userOrgDetails)
	//	if err == nil {
	//		return userOrgDetails, nil
	//	}
	//}

	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	createdAt := time.Now().Unix()
	qryStr := fmt.Sprintf(`SELECT * from userz.user_org_map where user_id='%s' and user_lsp_id='%s'  AND created_at < %d ALLOW FILTERING`, userID, lspID, createdAt)
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
	//redisBytes, err := json.Marshal(userOrgs[0])
	//if err == nil {
	//	redis.SetTTL(key, 3600)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}

	return userOrgs[0], nil
}

func GetUserPreferenceForLsp(ctx context.Context, userID string, lspID string) (*model.UserPreference, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	//key := "GetUserPreferenceForLsp" + userID + lspID
	//result, err := redis.GetRedisValue(key)
	//if err == nil {
	//	var userPreference *model.UserPreference
	//	err = json.Unmarshal([]byte(result), &userPreference)
	//	if err == nil {
	//		return userPreference, nil
	//	}
	//}

	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	createdAt := time.Now().Unix()
	qryStr := fmt.Sprintf(`SELECT * from userz.user_preferences where user_id='%s' and user_lsp_id='%s' AND created_at < %d ALLOW FILTERING`, userID, lspID, createdAt)
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
	//redisBytes, err := json.Marshal(userOrgs[0])
	//if err == nil {
	//	redis.SetTTL(key, 3600)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
	return userOrgs[0], nil
}

func GetUserLspByLspID(ctx context.Context, userID string, lspID string) (*model.UserLspMap, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	//key := "GetUserLspByLspID" + userID + lspID
	//result, err := redis.GetRedisValue(key)
	//if err == nil {
	//	var userLsp *model.UserLspMap
	//	err = json.Unmarshal([]byte(result), &userLsp)
	//	if err == nil {
	//		return userLsp, nil
	//	}
	//}

	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	createdAt := time.Now().Unix()
	qryStr := fmt.Sprintf(`SELECT * from userz.user_lsp_map where user_id='%s' and lsp_id='%s' AND created_at < %d ALLOW FILTERING`, userID, lspID, createdAt)
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
			LspID:     copiedOrg.LspId,
			Status:    copiedOrg.Status,
			CreatedBy: &copiedOrg.CreatedBy,
			UpdatedBy: &copiedOrg.UpdatedBy,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
		userOrgs = append(userOrgs, currentUserOrg)
	}
	//redisBytes, err := json.Marshal(userOrgs[0])
	//if err == nil {
	//	redis.SetTTL(key, 3600)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}

	return userOrgs[0], nil
}
