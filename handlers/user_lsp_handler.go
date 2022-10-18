package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/graph/model"
)

func AddUserLspMap(ctx context.Context, input []*model.UserLspMapInput, isAdmin *bool) ([]*model.UserLspMap, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil && isAdmin == nil {
		return nil, fmt.Errorf("user not found")
	}
	isAllowed := false
	if isAdmin != nil && *isAdmin {
		isAllowed = *isAdmin
	}
	if !isAllowed {
		role := strings.ToLower(userCass.Role)
		if userCass.ID == input[0].UserID || role == "admin" || strings.Contains(role, "manager") {
			isAllowed = true
		}
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create lsp mapping")
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	userLspMaps := make([]*model.UserLspMap, 0)
	for _, input := range input {
		guid := xid.New()
		createdBy := ""
		updatedBy := ""
		if input.CreatedBy != nil {
			createdBy = *input.CreatedBy
		} else {
			createdBy = userCass.Email
		}
		if input.UpdatedBy != nil {
			updatedBy = *input.UpdatedBy
		} else {
			updatedBy = userCass.Email
		}
		userLspMap := userz.UserLsp{
			ID:        guid.String(),
			UserID:    input.UserID,
			LspId:     input.LspID,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
			Status:    input.Status,
		}
		insertQuery := CassUserSession.Query(userz.UserLspTable.Insert()).BindStruct(userLspMap)
		if err := insertQuery.ExecRelease(); err != nil {
			return nil, err
		}
		created := strconv.FormatInt(userLspMap.CreatedAt, 10)
		updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
		userLspOutput := &model.UserLspMap{
			UserLspID: &userLspMap.ID,
			UserID:    userLspMap.UserID,
			LspID:     userLspMap.LspId,
			CreatedAt: created,
			UpdatedAt: updated,
			CreatedBy: &userLspMap.CreatedBy,
			UpdatedBy: &userLspMap.UpdatedBy,
			Status:    userLspMap.Status,
		}
		userLspMaps = append(userLspMaps, userLspOutput)
	}
	return userLspMaps, nil
}

func UpdateUserLspMap(ctx context.Context, input model.UserLspMapInput) (*model.UserLspMap, error) {
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
		return nil, fmt.Errorf("user not allowed to create lsp mapping")
	}
	if input.UserLspID == nil {
		return nil, fmt.Errorf("user lsp id is required")
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userLspMap := userz.UserLsp{
		ID: *input.UserLspID,
	}
	userLsps := []userz.UserLsp{}

	getQueryStr := fmt.Sprintf("SELECT * FROM userz.user_lsp_map WHERE id='%s' AND user_id='%s'  ", userLspMap.ID, userCass.ID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
	if err := getQuery.SelectRelease(&userLsps); err != nil {
		return nil, err
	}
	if len(userLsps) == 0 {
		return nil, fmt.Errorf("users lsp not found")
	}
	userLspMap = userLsps[0]
	updatedCols := []string{}
	if input.Status != "" {
		userLspMap.Status = input.Status
		updatedCols = append(updatedCols, "status")
	}
	if input.UpdatedBy != nil {
		userLspMap.UpdatedBy = *input.UpdatedBy
		updatedCols = append(updatedCols, "updated_by")
	}
	if input.LspID != "" {
		userLspMap.LspId = input.LspID
		updatedCols = append(updatedCols, "lsp_id")
	}
	if len(updatedCols) > 0 {
		updatedAt := time.Now().Unix()
		userLspMap.UpdatedAt = updatedAt
		updatedCols = append(updatedCols, "updated_at")
		upStms, uNames := userz.UserLspTable.Update(updatedCols...)
		updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&userLspMap)
		if err := updateQuery.ExecRelease(); err != nil {
			log.Errorf("error updating user lsp: %v", err)
			return nil, err
		}
	}
	created := strconv.FormatInt(userLspMap.CreatedAt, 10)
	updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
	userLspOutput := &model.UserLspMap{
		UserLspID: &userLspMap.ID,
		UserID:    userLspMap.UserID,
		LspID:     userLspMap.LspId,
		CreatedAt: created,
		UpdatedAt: updated,
		CreatedBy: &userLspMap.CreatedBy,
		UpdatedBy: &userLspMap.UpdatedBy,
		Status:    userLspMap.Status,
	}
	return userLspOutput, nil
}
