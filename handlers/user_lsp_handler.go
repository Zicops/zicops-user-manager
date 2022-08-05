package handlers

import (
	"context"
	"encoding/base64"
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
	"github.com/zicops/zicops-user-manager/helpers"
)

func AddUserLspMap(ctx context.Context, input []*model.UserLspMapInput) ([]*model.UserLspMap, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	userCass := userz.User{
		ID: emailCreatorID,
	}
	users := []userz.User{}
	getQuery := global.CassUserSession.Session.Query(userz.UserTable.Get()).BindMap(qb.M{"id": userCass.ID})
	if err := getQuery.SelectRelease(&users); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	userCass = users[0]
	isAllowed := false
	if userCass.Email == email_creator || strings.ToLower(userCass.Role) == "admin" {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create lsp mapping")
	}
	userLspMaps := make([]*model.UserLspMap, 0)
	for _, input := range input {
		guid := xid.New()
		createdBy := email_creator
		updatedBy := email_creator
		if input.CreatedBy != nil {
			createdBy = *input.CreatedBy
		}
		if input.UpdatedBy != nil {
			updatedBy = *input.UpdatedBy
		}
		userLspMap := userz.UserLsp{
			ID:        guid.String(),
			UserID:    input.UserID,
			LspID:     input.LspID,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		}
		insertQuery := global.CassUserSession.Session.Query(userz.UserLspTable.Insert()).BindStruct(userLspMap)
		if err := insertQuery.ExecRelease(); err != nil {
			return nil, err
		}
		created := strconv.FormatInt(userLspMap.CreatedAt, 10)
		updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
		userLspOutput := &model.UserLspMap{
			UserLspID: &userLspMap.ID,
			UserID:    userLspMap.UserID,
			LspID:     userLspMap.LspID,
			CreatedAt: created,
			UpdatedAt: updated,
			CreatedBy: &userLspMap.CreatedBy,
			UpdatedBy: &userLspMap.UpdatedBy,
		}
		userLspMaps = append(userLspMaps, userLspOutput)
	}
	return userLspMaps, nil
}

func UpdateUserLspMap(ctx context.Context, input model.UserLspMapInput) (*model.UserLspMap, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	userCass := userz.User{
		ID: emailCreatorID,
	}
	users := []userz.User{}
	getQuery := global.CassUserSession.Session.Query(userz.UserTable.Get()).BindMap(qb.M{"id": userCass.ID})
	if err := getQuery.SelectRelease(&users); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	userCass = users[0]
	isAllowed := false
	if userCass.Email == email_creator || strings.ToLower(userCass.Role) == "admin" {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create lsp mapping")
	}
	if input.UserLspID == nil {
		return nil, fmt.Errorf("user lsp id is required")
	}
	userLspMap := userz.UserLsp{
		ID: *input.UserLspID,
	}
	userLsps := []userz.UserLsp{}
	getQuery = global.CassUserSession.Session.Query(userz.UserLspTable.Get()).BindMap(qb.M{"id": userLspMap.ID})
	if err := getQuery.SelectRelease(&userLsps); err != nil {
		return nil, err
	}
	if len(users) == 0 {
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
		userLspMap.LspID = input.LspID
		updatedCols = append(updatedCols, "lsp_id")
	}
	updatedAt := time.Now().Unix()
	userLspMap.UpdatedAt = updatedAt
	updatedCols = append(updatedCols, "updated_at")
	if len(updatedCols) == 0 {
		return nil, fmt.Errorf("nothing to update")
	}
	upStms, uNames := userz.UserLspTable.Update(updatedCols...)
	updateQuery := global.CassUserSession.Session.Query(upStms, uNames).BindStruct(&userLspMap)
	if err := updateQuery.ExecRelease(); err != nil {
		log.Errorf("error updating user lsp: %v", err)
		return nil, err
	}
	created := strconv.FormatInt(userLspMap.CreatedAt, 10)
	updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
	userLspOutput := &model.UserLspMap{
		UserLspID: &userLspMap.ID,
		UserID:    userLspMap.UserID,
		LspID:     userLspMap.LspID,
		CreatedAt: created,
		UpdatedAt: updated,
		CreatedBy: &userLspMap.CreatedBy,
		UpdatedBy: &userLspMap.UpdatedBy,
	}
	return userLspOutput, nil
}
