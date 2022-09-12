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
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
)

func AddUserNotes(ctx context.Context, input []*model.UserNotesInput) ([]*model.UserNotes, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	isAllowed := false
	if userCass.ID == input[0].UserID || strings.ToLower(userCass.Role) == "admin" {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create notes")
	}
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	global.CassUserSession = session
	defer global.CassUserSession.Close()
	userLspMaps := make([]*model.UserNotes, 0)
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
		userLspMap := userz.UserNotes{
			ID:        guid.String(),
			UserID:    input.UserID,
			UserLspID: input.UserLspID,
			CourseID:  input.CourseID,
			ModuleID:  input.ModuleID,
			TopicID:   input.TopicID,
			Details:   input.Details,
			Status:    input.Status,
			Sequence:  input.Sequence,
			IsActive:  input.IsActive,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			CreatedBy: createdBy,
			UpdatedBy: updatedBy,
		}
		insertQuery := global.CassUserSession.Query(userz.UserNotesTable.Insert()).BindStruct(userLspMap)
		if err := insertQuery.ExecRelease(); err != nil {
			return nil, err
		}
		created := strconv.FormatInt(userLspMap.CreatedAt, 10)
		updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
		userLspOutput := &model.UserNotes{
			UserNotesID: &userLspMap.ID,
			UserLspID:   userLspMap.UserLspID,
			UserID:      userLspMap.UserID,
			CourseID:    userLspMap.CourseID,
			ModuleID:    userLspMap.ModuleID,
			TopicID:     userLspMap.TopicID,
			Details:     userLspMap.Details,
			Status:      userLspMap.Status,
			Sequence:    userLspMap.Sequence,
			IsActive:    userLspMap.IsActive,
			CreatedAt:   created,
			UpdatedAt:   updated,
			CreatedBy:   &userLspMap.CreatedBy,
			UpdatedBy:   &userLspMap.UpdatedBy,
		}
		userLspMaps = append(userLspMaps, userLspOutput)
	}
	return userLspMaps, nil
}

func UpdateUserNotes(ctx context.Context, input model.UserNotesInput) (*model.UserNotes, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	isAllowed := false
	if userCass.ID == input.UserID || strings.ToLower(userCass.Role) == "admin" {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create org mapping")
	}
	if input.UserNotesID == nil {
		return nil, fmt.Errorf("user notes id is required")
	}
	session, err := cassandra.GetCassSession("coursez")
	if err != nil {
		return nil, err
	}
	global.CassUserSession = session
	defer global.CassUserSession.Close()
	userLspMap := userz.UserNotes{
		ID: *input.UserNotesID,
	}
	userLsps := []userz.UserNotes{}
	getQuery := global.CassUserSession.Query(userz.UserNotesTable.Get()).BindMap(qb.M{"id": userLspMap.ID})
	if err := getQuery.SelectRelease(&userLsps); err != nil {
		return nil, err
	}
	if len(userLsps) == 0 {
		return nil, fmt.Errorf("users notes not found")
	}
	userLspMap = userLsps[0]
	updatedCols := []string{}
	if input.Details != "" {
		userLspMap.Details = input.Details
		updatedCols = append(updatedCols, "details")
	}
	if input.UpdatedBy != nil {
		userLspMap.UpdatedBy = *input.UpdatedBy
		updatedCols = append(updatedCols, "updated_by")
	}
	if input.Status != "" && input.Status != userLspMap.Status {
		userLspMap.Status = input.Status
		updatedCols = append(updatedCols, "status")
	}
	if input.Sequence != 0 && input.Sequence != userLspMap.Sequence {
		userLspMap.Sequence = input.Sequence
		updatedCols = append(updatedCols, "sequence")
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
	if input.UserLspID != "" && input.UserLspID != userLspMap.UserLspID {
		userLspMap.UserLspID = input.UserLspID
		updatedCols = append(updatedCols, "user_lsp_id")
	}
	if input.IsActive != userLspMap.IsActive {
		userLspMap.IsActive = input.IsActive
		updatedCols = append(updatedCols, "is_active")
	}
	updatedAt := time.Now().Unix()
	userLspMap.UpdatedAt = updatedAt
	updatedCols = append(updatedCols, "updated_at")
	if len(updatedCols) == 0 {
		return nil, fmt.Errorf("nothing to update")
	}
	upStms, uNames := userz.UserNotesTable.Update(updatedCols...)
	updateQuery := global.CassUserSession.Query(upStms, uNames).BindStruct(&userLspMap)
	if err := updateQuery.ExecRelease(); err != nil {
		log.Errorf("error updating user notes: %v", err)
		return nil, err
	}
	created := strconv.FormatInt(userLspMap.CreatedAt, 10)
	updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
	userLspOutput := &model.UserNotes{
		UserNotesID: &userLspMap.ID,
		UserLspID:   userLspMap.UserLspID,
		UserID:      userLspMap.UserID,
		CourseID:    userLspMap.CourseID,
		ModuleID:    userLspMap.ModuleID,
		TopicID:     userLspMap.TopicID,
		Details:     userLspMap.Details,
		Status:      userLspMap.Status,
		Sequence:    userLspMap.Sequence,
		IsActive:    userLspMap.IsActive,
		CreatedAt:   created,
		UpdatedAt:   updated,
		CreatedBy:   &userLspMap.CreatedBy,
		UpdatedBy:   &userLspMap.UpdatedBy,
	}
	return userLspOutput, nil
}
