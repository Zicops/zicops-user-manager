package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rs/xid"
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
