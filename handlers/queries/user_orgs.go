package queries

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/handlers/orgs"
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

	qryStr := fmt.Sprintf(`SELECT * from userz.user_org_map where user_id='%s'   ALLOW FILTERING`, emailCreatorID)
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

	qryStr := fmt.Sprintf(`SELECT * from userz.user_preferences where user_id='%s'  ALLOW FILTERING`, emailCreatorID)
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
	userOrgs := make([]*model.UserLspMap, 0)

	key := fmt.Sprintf("zicops_user_lsp_%s", userId)
	res, err := redis.GetRedisValue(ctx, key)
	if err == nil && res != ""{
		err = json.Unmarshal([]byte(res), &userOrgs)
		if err == nil {
			return userOrgs, nil
		}
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	origin := claims["origin"].(string)
	mobile := claims["mobile"].(string)
	// remove https://
	origin = strings.Replace(origin, "https://", "", 1)
	// replace www. with empty string
	origin = strings.Replace(origin, "www.", "", 1)
	// replace /
	origin = strings.Replace(origin, "/", "", 1)
	returnAll := false

	if origin == "zicops.com" || origin == "demo.zicops.com" || mobile != "" {
		returnAll = true
	}

	var orgsFromDomain []*model.Organization
	if !returnAll {
		orgsFromDomain, err = orgs.GetOrganizationsByDomain(ctx, origin)
		if err != nil {
			return nil, err
		}
		log.Errorf("orgsFromDomain: %v", *(orgsFromDomain[0].OrgID))
		currentOrgID := orgsFromDomain[0].OrgID
		lspMaps, err := orgs.GetLearningSpacesByOrgID(ctx, *currentOrgID)
		if err != nil {
			return nil, err
		}
		usrLspMapLocal := make([]*model.UserLspMap, len(lspMaps))
		var wg sync.WaitGroup
		for i, lm := range lspMaps {
			lcp := lm
			wg.Add(1)
			go func(i int, lspCopy *model.LearningSpace) {
				lspIdCurrent := lspCopy.LspID
				usrLspMap, err := GetUserLspMapsByLspIDOne(ctx, *lspIdCurrent, userId)
				if err != nil {
					log.Errorf("Error in GetUserLsps: %v", err)
				}
				if usrLspMap.LspID != "" {
					usrLspMapLocal[i] = &usrLspMap
				}
				wg.Done()
			}(i, lcp)
		}
		wg.Wait()
		for _, usrLspMap := range usrLspMapLocal {
			if usrLspMap != nil {
				userOrgs = append(userOrgs, usrLspMap)
			}
		}
	} else {
		qryStr := fmt.Sprintf(`SELECT * from userz.user_lsp_map where user_id='%s'  ALLOW FILTERING`, emailCreatorID)
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
	}
	redisBytes, err := json.Marshal(userOrgs)
	if err == nil {
		redis.SetTTL(ctx, key, 3600)
		redis.SetRedisValue(ctx, key, string(redisBytes))
	}
	return userOrgs, nil
}

func GetUserLspMapsByLspIDOne(ctx context.Context, lspID string, userID string) (model.UserLspMap, error) {
	session, err := cassandra.GetCassSession("userz")
	var userLspMap model.UserLspMap
	if err != nil {
		return userLspMap, err
	}
	CassUserSession := session

	qryStr := fmt.Sprintf(`SELECT * from userz.user_lsp_map where lsp_id='%s' AND user_id='%s'  ALLOW FILTERING`, lspID, userID)
	getUsers := func() (users []userz.UserLsp, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return users, iter.Select(&users)
	}
	usersOrgs, err := getUsers()
	if err != nil {
		return userLspMap, err
	}
	userOrgs := make([]*model.UserLspMap, len(usersOrgs))
	var outputResponse model.PaginatedUserLspMaps

	if len(usersOrgs) <= 0 {
		outputResponse.UserLspMaps = userOrgs
		return userLspMap, nil
	}

	copiedOrg := usersOrgs[0]

	createdAt := strconv.FormatInt(copiedOrg.CreatedAt, 10)
	updatedAt := strconv.FormatInt(copiedOrg.UpdatedAt, 10)
	userLspMap = model.UserLspMap{
		UserLspID: &copiedOrg.ID,
		UserID:    copiedOrg.UserID,
		LspID:     copiedOrg.LspId,
		Status:    copiedOrg.Status,
		CreatedBy: &copiedOrg.CreatedBy,
		UpdatedBy: &copiedOrg.UpdatedBy,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	return userLspMap, nil
}

func GetUserLspMapsByLspID(ctx context.Context, lspID string, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedUserLspMaps, error) {
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	_, err = helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var newPage []byte
	//var pageDirection string
	var pageSizeInt int
	var newCursor string
	if pageCursor != nil && *pageCursor != "" {
		page, err := global.CryptSession.DecryptString(*pageCursor, nil)
		if err != nil {
			return nil, fmt.Errorf("invalid page cursor: %v", err)
		}
		newPage = page
	}
	qryStr := fmt.Sprintf(`SELECT * from userz.user_lsp_map where lsp_id='%s'  ALLOW FILTERING`, lspID)
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
	if len(newPage) != 0 {
		newCursor, err = global.CryptSession.EncryptAsString(newPage, nil)
		if err != nil {
			return nil, fmt.Errorf("error encrypting cursor: %v", err)
		}
		log.Infof("Users: %v", string(newCursor))

	}
	userOrgs := make([]*model.UserLspMap, len(usersOrgs))
	var outputResponse model.PaginatedUserLspMaps

	if len(usersOrgs) <= 0 {
		outputResponse.UserLspMaps = userOrgs
		return nil, nil
	}
	newOrgs := make([]*model.UserLspMap, 0)
	for _, userOrg := range usersOrgs {
		copiedOrg := userOrg
		if copiedOrg.ID == "" {
			continue
		}
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
		newOrgs = append(newOrgs, currentUserOrg)
	}
	outputResponse.UserLspMaps = newOrgs
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

	qryStr := fmt.Sprintf(`SELECT * from userz.user_org_map where user_id='%s' and user_lsp_id='%s'   ALLOW FILTERING`, userID, lspID)
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
	key := fmt.Sprintf("zicops_user_preference_%s_%s", userID, lspID)
	result, err := redis.GetRedisValue(ctx, key)
	if err == nil {
		var userPreference *model.UserPreference
		err = json.Unmarshal([]byte(result), &userPreference)
		if err == nil {
			return userPreference, nil
		}
	}

	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	qryStr := fmt.Sprintf(`SELECT * from userz.user_preferences where user_id='%s' and user_lsp_id='%s'  ALLOW FILTERING`, userID, lspID)
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
	redisBytes, err := json.Marshal(userOrgs[0])
	if err == nil {
		redis.SetRedisValue(ctx, key, string(redisBytes))
		redis.SetTTL(ctx, key, 3600)
	}
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

	qryStr := fmt.Sprintf(`SELECT * from userz.user_lsp_map where user_id='%s' and lsp_id='%s'  ALLOW FILTERING`, userID, lspID)
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
