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

func AddUserBookmark(ctx context.Context, input []*model.UserBookmarkInput) ([]*model.UserBookmark, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	isAllowed := false
	role := strings.ToLower(userCass.Role)
	if userCass.ID == input[0].UserID || role == "admin" || strings.Contains(role, "manager") {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create bookmarks")
	}
	userLspMaps := make([]*model.UserBookmark, 0)
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
		userLspMap := userz.UserBookmarks{
			ID:        guid.String(),
			UserID:    input.UserID,
			UserLspID: input.UserLspID,
			CourseID:  input.CourseID,
			ModuleID:  input.ModuleID,
			TopicID:   input.TopicID,
			UserCPID:  input.UserCourseID,
			Name:      input.Name,
			TimeStamp: input.TimeStamp,
			IsActive:  input.IsActive,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		}
		insertQuery := CassUserSession.Query(userz.UserBookmarksTable.Insert()).BindStruct(userLspMap)
		if err := insertQuery.ExecRelease(); err != nil {
			return nil, err
		}
		created := strconv.FormatInt(userLspMap.CreatedAt, 10)
		updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
		userLspOutput := &model.UserBookmark{
			UserBmID:     &userLspMap.ID,
			UserLspID:    userLspMap.UserLspID,
			UserID:       userLspMap.UserID,
			CourseID:     userLspMap.CourseID,
			ModuleID:     userLspMap.ModuleID,
			TopicID:      userLspMap.TopicID,
			UserCourseID: userLspMap.UserCPID,
			Name:         userLspMap.Name,
			TimeStamp:    input.TimeStamp,
			IsActive:     userLspMap.IsActive,
			CreatedAt:    created,
			UpdatedAt:    updated,
			CreatedBy:    &userLspMap.CreatedBy,
			UpdatedBy:    &userLspMap.UpdatedBy,
		}
		userLspMaps = append(userLspMaps, userLspOutput)
	}
	return userLspMaps, nil
}

func UpdateUserBookmark(ctx context.Context, input model.UserBookmarkInput) (*model.UserBookmark, error) {
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
	if input.UserBmID == nil {
		return nil, fmt.Errorf("user bookmark id is required")
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userLspMap := userz.UserBookmarks{
		ID: *input.UserBmID,
	}
	userLsps := []userz.UserBookmarks{}

	getQueryStr := fmt.Sprintf("SELECT * FROM userz.user_bookmarks WHERE id='%s' AND user_id='%s'  ", userLspMap.ID, userCass.ID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
	if err := getQuery.SelectRelease(&userLsps); err != nil {
		return nil, err
	}
	if len(userLsps) == 0 {
		return nil, fmt.Errorf("users orgs not found")
	}
	userLspMap = userLsps[0]
	updatedCols := []string{}
	if input.TimeStamp != "" {
		userLspMap.TimeStamp = input.TimeStamp
		updatedCols = append(updatedCols, "time_stamp")
	}
	if input.UpdatedBy != nil {
		userLspMap.UpdatedBy = *input.UpdatedBy
		updatedCols = append(updatedCols, "updated_by")
	}
	if input.Name != "" && input.Name != userLspMap.Name {
		userLspMap.Name = input.Name
		updatedCols = append(updatedCols, "name")
	}
	if input.TopicID != "" && input.TopicID != userLspMap.TopicID {
		userLspMap.TopicID = input.TopicID
		updatedCols = append(updatedCols, "topic_id")
	}
	if input.ModuleID != "" && input.ModuleID != userLspMap.ModuleID {
		userLspMap.ModuleID = input.ModuleID
		updatedCols = append(updatedCols, "module_id")
	}
	if input.CourseID != "" && input.CourseID != userLspMap.CourseID {
		userLspMap.CourseID = input.CourseID
		updatedCols = append(updatedCols, "course_id")
	}
	if input.UserCourseID != "" && input.UserCourseID != userLspMap.UserCPID {
		userLspMap.UserCPID = input.UserCourseID
		updatedCols = append(updatedCols, "user_cp_id")
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
	upStms, uNames := userz.UserBookmarksTable.Update(updatedCols...)
	updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&userLspMap)
	if err := updateQuery.ExecRelease(); err != nil {
		log.Errorf("error updating user bookmark: %v", err)
		return nil, err
	}
	created := strconv.FormatInt(userLspMap.CreatedAt, 10)
	updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
	userLspOutput := &model.UserBookmark{
		UserBmID:     &userLspMap.ID,
		UserLspID:    userLspMap.UserLspID,
		UserID:       userLspMap.UserID,
		CourseID:     userLspMap.CourseID,
		ModuleID:     userLspMap.ModuleID,
		TopicID:      userLspMap.TopicID,
		UserCourseID: userLspMap.UserCPID,
		Name:         userLspMap.Name,
		TimeStamp:    input.TimeStamp,
		IsActive:     userLspMap.IsActive,
		CreatedAt:    created,
		UpdatedAt:    updated,
		CreatedBy:    &userLspMap.CreatedBy,
		UpdatedBy:    &userLspMap.UpdatedBy,
	}
	return userLspOutput, nil
}
