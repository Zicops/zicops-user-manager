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

func AddUserLanguageMap(ctx context.Context, input []*model.UserLanguageMapInput) ([]*model.UserLanguageMap, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	isAllowed := false
	if userCass.ID == input[0].UserID || strings.ToLower(userCass.Role) == "admin" {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create lang mapping")
	}
	userLspMaps := make([]*model.UserLanguageMap, 0)
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
		userLspMap := userz.UserLang{
			ID:        guid.String(),
			UserID:    input.UserID,
			UserLspID: input.UserLspID,
			Language:  input.Language,
			IsBase:    input.IsBaseLanguage,
			IsActive:  input.IsActive,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		}
		insertQuery := global.CassUserSession.Session.Query(userz.UserLangTable.Insert()).BindStruct(userLspMap)
		if err := insertQuery.ExecRelease(); err != nil {
			return nil, err
		}
		created := strconv.FormatInt(userLspMap.CreatedAt, 10)
		updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
		userLspOutput := &model.UserLanguageMap{
			UserLanguageID: &userLspMap.ID,
			UserLspID:      userLspMap.UserLspID,
			UserID:         userLspMap.UserID,
			IsActive:       userLspMap.IsActive,
			IsBaseLanguage: userLspMap.IsBase,
			Language:       userLspMap.Language,
			CreatedAt:      created,
			UpdatedAt:      updated,
			CreatedBy:      &userLspMap.CreatedBy,
			UpdatedBy:      &userLspMap.UpdatedBy,
		}
		userLspMaps = append(userLspMaps, userLspOutput)
	}
	return userLspMaps, nil
}

func AddUserPreference(ctx context.Context, input []*model.UserPreferenceInput) ([]*model.UserPreference, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	isAllowed := false
	if userCass.ID == input[0].UserID || strings.ToLower(userCass.Role) == "admin" {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create lang mapping")
	}
	userLspMaps := make([]*model.UserPreference, 0)
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
		userLspMap := userz.UserPreferences{
			ID:          guid.String(),
			UserID:      input.UserID,
			UserLspID:   input.UserLspID,
			SubCategory: input.SubCategory,
			IsBase:      input.IsBase,
			IsActive:    input.IsActive,
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
			CreatedBy:   createdBy,
			UpdatedBy:   updatedBy,
		}
		insertQuery := global.CassUserSession.Session.Query(userz.UserPreferencesTable.Insert()).BindStruct(userLspMap)
		if err := insertQuery.ExecRelease(); err != nil {
			return nil, err
		}
		created := strconv.FormatInt(userLspMap.CreatedAt, 10)
		updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
		userLspOutput := &model.UserPreference{
			UserPreferenceID: &userLspMap.ID,
			UserLspID:        userLspMap.UserLspID,
			UserID:           userLspMap.UserID,
			IsActive:         userLspMap.IsActive,
			IsBase:           userLspMap.IsBase,
			SubCategory:      userLspMap.SubCategory,
			CreatedAt:        created,
			UpdatedAt:        updated,
			CreatedBy:        &userLspMap.CreatedBy,
			UpdatedBy:        &userLspMap.UpdatedBy,
		}
		userLspMaps = append(userLspMaps, userLspOutput)
	}
	return userLspMaps, nil
}

func UpdateUserPreference(ctx context.Context, input model.UserPreferenceInput) (*model.UserPreference, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	isAllowed := false
	if userCass.ID == input.UserID || strings.ToLower(userCass.Role) == "admin" {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to update pref mapping")
	}
	if input.UserPreferenceID == nil {
		return nil, fmt.Errorf("user preference id is required")
	}
	userLspMap := userz.UserPreferences{
		ID: *input.UserPreferenceID,
	}
	userLsps := []userz.UserPreferences{}
	getQuery := global.CassUserSession.Session.Query(userz.UserPreferencesTable.Get()).BindMap(qb.M{"id": userLspMap.ID})
	if err := getQuery.SelectRelease(&userLsps); err != nil {
		return nil, err
	}
	if len(userLsps) == 0 {
		return nil, fmt.Errorf("users prefs not found")
	}
	userLspMap = userLsps[0]
	updatedCols := []string{}
	if input.SubCategory != "" && input.SubCategory != userLspMap.SubCategory {
		userLspMap.SubCategory = input.SubCategory
		updatedCols = append(updatedCols, "sub_category")
	}
	if input.IsBase != userLspMap.IsBase {
		userLspMap.IsBase = input.IsBase
		updatedCols = append(updatedCols, "is_base")
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
	upStms, uNames := userz.UserPreferencesTable.Update(updatedCols...)
	updateQuery := global.CassUserSession.Session.Query(upStms, uNames).BindStruct(&userLspMap)
	if err := updateQuery.ExecRelease(); err != nil {
		log.Errorf("error updating user pref: %v", err)
		return nil, err
	}
	created := strconv.FormatInt(userLspMap.CreatedAt, 10)
	updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
	userLspOutput := &model.UserPreference{
		UserPreferenceID: &userLspMap.ID,
		UserLspID:        userLspMap.UserLspID,
		UserID:           userLspMap.UserID,
		SubCategory:      userLspMap.SubCategory,
		IsBase:           userLspMap.IsBase,
		IsActive:         userLspMap.IsActive,
		CreatedAt:        created,
		UpdatedAt:        updated,
		CreatedBy:        &userLspMap.CreatedBy,
		UpdatedBy:        &userLspMap.UpdatedBy,
	}
	return userLspOutput, nil
}
