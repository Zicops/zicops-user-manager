package handlers

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/handlers/common"
	"github.com/zicops/zicops-user-manager/lib/identity"
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
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userLspMaps := make([]*model.UserRole, 0)
	for _, input := range input {

		if input == nil {
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
	session, err := global.CassPool.GetSession(ctx, "userz")
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
	session, err := global.CassPool.GetSession(ctx, "userz")
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

func GetLspUsersRoles(ctx context.Context, lspID string, role []*string) ([]*model.UserDetailsRole, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting context: %v", err)
		return nil, err
	}
	var res []*model.UserDetailsRole
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	queryStr := fmt.Sprintf(`SELECT * FROM userz.user_lsp_map WHERE lsp_id = '%s' ALLOW FILTERING`, lspID)
	getLspUsers := func() (maps []userz.UserLsp, err error) {
		q := CassUserSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return maps, iter.Select(&maps)
	}
	userLspMaps, err := getLspUsers()
	if err != nil {
		log.Printf("Got error getting users of a lsp: %v", err)
		return nil, err
	}
	if len(userLspMaps) == 0 {
		return nil, nil
	}
	users := make([]*string, 0)
	userIdLspIdMap := make(map[string]string)
	for _, v := range userLspMaps {
		users = append(users, &v.UserID)
		userIdLspIdMap[v.UserID] = v.ID
	}
	//get all users details
	userDetails, err := common.GetUserDetails(ctx, users)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	res = make([]*model.UserDetailsRole, len(userLspMaps))
	for i, ud := range userDetails {
		wg.Add(1)
		if ud == nil || ud.ID == nil {
			continue
		}
		go func(i int, ud *model.User) {
			defer wg.Done()
			userLspId := userIdLspIdMap[*ud.ID]
			//got all roles information for a user, with filter of a role
			qryStr := fmt.Sprintf(`SELECT * FROM userz.user_role WHERE user_id='%s' AND user_lsp_id='%s' `, *ud.ID, userLspId)
			if role != nil {
				qryStr = qryStr + " and role in ("
				for _, r := range role {
					if r == nil {
						continue
					}
					qryStr = qryStr + fmt.Sprintf(`'%s', `, *r)
				}
				//remove the extra comma and space which we have, plus add the bracket
				qryStr = qryStr[:len(qryStr)-2] + ")"
			}
			queryStr += " ALLOW FILTERING"
			getUserRoles := func() (maps []userz.UserRole, err error) {
				q := CassUserSession.Query(qryStr, nil)
				defer q.Release()
				iter := q.Iter()
				return maps, iter.Select(&maps)
			}
			userRoles, err := getUserRoles()
			if err != nil {
				log.Printf("Got error getting user roles: %v", err)
				return
			}
			roles := make([]*model.RoleData, 0)
			for _, vv := range userRoles {
				v := vv
				tmp := model.RoleData{
					UserRoleID: &v.ID,
					Role:       &v.Role,
				}
				roles = append(roles, &tmp)
			}
			res[i] = &model.UserDetailsRole{
				User:  ud,
				Roles: roles,
			}
		}(i, ud)
	}
	wg.Wait()
	return res, nil
}

func GetPaginatedLspUsersWithRoles(ctx context.Context, lspID string, role []*string, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedUserDetailsWithRole, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting context: %v", err)
		return nil, err
	}

	var newPage []byte

	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	if pageCursor != nil && *pageCursor != "" {
		page, err := global.CryptSession.DecryptString(*pageCursor, nil)
		if err != nil {
			return nil, err
		}
		newPage = page
	}

	var newCursor string
	var pageSizeInt int
	var outputResponse model.PaginatedUserDetailsWithRole
	if pageSize != nil {
		pageSizeInt = *pageSize
	} else {
		pageSizeInt = 10
	}

	queryStr := fmt.Sprintf(`SELECT * FROM userz.user_lsp_map WHERE lsp_id = '%s' ALLOW FILTERING`, lspID)
	getLspUsers := func(page []byte) (maps []userz.UserLsp, nextPage []byte, err error) {
		q := CassUserSession.Query(queryStr, nil)
		defer q.Release()
		q.PageState(page)
		q.PageSize(pageSizeInt)
		iter := q.Iter()
		return maps, iter.PageState(), iter.Select(&maps)
	}
	userLspMaps, newPage, err := getLspUsers(newPage)
	if err != nil {
		log.Printf("Got error getting users of a lsp: %v", err)
		return nil, err
	}
	if len(newPage) != 0 {
		newCursor, err = global.CryptSession.EncryptAsString(newPage, nil)
		if err != nil {
			return nil, err
		}
	}
	if len(userLspMaps) == 0 {
		return nil, nil
	}
	users := make([]*string, 0)
	userIdLspIdMap := make(map[string]string)
	for _, vv := range userLspMaps {
		v := vv
		users = append(users, &v.UserID)
		userIdLspIdMap[v.UserID] = v.ID
	}
	//get all users details
	userDetails, err := common.GetUserDetails(ctx, users)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	res := make([]*model.UserDetailsRole, len(userLspMaps))
	for i, ud := range userDetails {
		wg.Add(1)
		if ud == nil || ud.ID == nil {
			continue
		}
		go func(i int, ud *model.User) {
			defer wg.Done()
			userLspId := userIdLspIdMap[*ud.ID]
			//got all roles information for a user, with filter of a role
			qryStr := fmt.Sprintf(`SELECT * FROM userz.user_role WHERE user_id='%s' AND user_lsp_id='%s' `, *ud.ID, userLspId)
			if role != nil {
				qryStr = qryStr + " and role in ("
				for _, r := range role {
					if r == nil {
						continue
					}
					qryStr = qryStr + fmt.Sprintf(`'%s', `, *r)
				}
				//remove the extra comma and space which we have, plus add the bracket
				qryStr = qryStr[:len(qryStr)-2] + ")"
			}
			qryStr += " ALLOW FILTERING"
			getUserRoles := func() (maps []userz.UserRole, err error) {
				q := CassUserSession.Query(qryStr, nil)
				defer q.Release()
				iter := q.Iter()
				return maps, iter.Select(&maps)
			}
			userRoles, err := getUserRoles()
			if err != nil {
				log.Printf("Got error getting user roles: %v", err)
				return
			}
			roles := make([]*model.RoleData, 0)
			for _, vv := range userRoles {
				v := vv
				tmp := model.RoleData{
					UserRoleID: &v.ID,
					Role:       &v.Role,
				}
				roles = append(roles, &tmp)
			}
			res[i] = &model.UserDetailsRole{
				User:  ud,
				Roles: roles,
			}
		}(i, ud)
	}
	wg.Wait()
	outputResponse.Data = res
	outputResponse.Direction = direction
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	return &outputResponse, nil

}
