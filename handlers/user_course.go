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

func AddUserCourse(ctx context.Context, input []*model.UserCourseInput) ([]*model.UserCourse, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	global.CassUserSession = session
	defer global.CassUserSession.Close()
	isAllowed := false
	if userCass.ID == input[0].UserID || strings.ToLower(userCass.Role) == "admin" {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create org mapping")
	}
	userLspMaps := make([]*model.UserCourse, 0)
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
		var endDate int64
		if input.EndDate != nil {
			endDate, _ = strconv.ParseInt(*input.EndDate, 10, 64)
		}
		userLspMap := userz.UserCourse{
			ID:           guid.String(),
			UserID:       input.UserID,
			UserLspID:    input.UserLspID,
			CourseID:     input.CourseID,
			CourseType:   input.CourseType,
			CourseStatus: input.CourseStatus,
			AddedBy:      input.AddedBy,
			IsMandatory:  input.IsMandatory,
			EndDate:      endDate,
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
			CreatedBy:    createdBy,
			UpdatedBy:    updatedBy,
		}
		insertQuery := global.CassUserSession.Query(userz.UserCourseTable.Insert()).BindStruct(userLspMap)
		if err := insertQuery.ExecRelease(); err != nil {
			return nil, err
		}
		created := strconv.FormatInt(userLspMap.CreatedAt, 10)
		updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
		userLspOutput := &model.UserCourse{
			UserCourseID: &userLspMap.ID,
			UserLspID:    userLspMap.UserLspID,
			UserID:       userLspMap.UserID,
			CourseID:     userLspMap.CourseID,
			CourseType:   userLspMap.CourseType,
			CourseStatus: userLspMap.CourseStatus,
			AddedBy:      userLspMap.AddedBy,
			IsMandatory:  userLspMap.IsMandatory,
			EndDate:      input.EndDate,
			CreatedAt:    created,
			UpdatedAt:    updated,
			CreatedBy:    &userLspMap.CreatedBy,
			UpdatedBy:    &userLspMap.UpdatedBy,
		}
		userLspMaps = append(userLspMaps, userLspOutput)
	}
	return userLspMaps, nil
}

func UpdateUserCourse(ctx context.Context, input model.UserCourseInput) (*model.UserCourse, error) {
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
	if input.UserCourseID == nil {
		return nil, fmt.Errorf("user course id is required")
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	global.CassUserSession = session
	defer global.CassUserSession.Close()
	userLspMap := userz.UserCourse{
		ID: *input.UserCourseID,
	}
	userLsps := []userz.UserCourse{}
	getQuery := global.CassUserSession.Query(userz.UserCourseTable.Get()).BindMap(qb.M{"id": userLspMap.ID})
	if err := getQuery.SelectRelease(&userLsps); err != nil {
		return nil, err
	}
	if len(userLsps) == 0 {
		return nil, fmt.Errorf("users orgs not found")
	}
	userLspMap = userLsps[0]
	updatedCols := []string{}
	if input.EndDate != nil {
		endDate, _ := strconv.ParseInt(*input.EndDate, 10, 64)
		userLspMap.EndDate = endDate
		updatedCols = append(updatedCols, "end_date")
	}
	if input.CourseID != "" && input.CourseID != userLspMap.CourseID {
		userLspMap.CourseID = input.CourseID
		updatedCols = append(updatedCols, "course_id")
	}
	if input.CourseType != "" && input.CourseType != userLspMap.CourseType {
		userLspMap.CourseType = input.CourseType
		updatedCols = append(updatedCols, "course_type")
	}
	if input.CourseStatus != "" && input.CourseStatus != userLspMap.CourseStatus {
		userLspMap.CourseStatus = input.CourseStatus
		updatedCols = append(updatedCols, "course_status")
	}
	if input.AddedBy != "" && input.AddedBy != userLspMap.AddedBy {
		userLspMap.AddedBy = input.AddedBy
		updatedCols = append(updatedCols, "added_by")
	}
	if input.IsMandatory != userLspMap.IsMandatory {
		userLspMap.IsMandatory = input.IsMandatory
		updatedCols = append(updatedCols, "is_mandatory")
	}
	if input.UpdatedBy != nil {
		userLspMap.UpdatedBy = *input.UpdatedBy
		updatedCols = append(updatedCols, "updated_by")
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
	upStms, uNames := userz.UserCourseTable.Update(updatedCols...)
	updateQuery := global.CassUserSession.Query(upStms, uNames).BindStruct(&userLspMap)
	if err := updateQuery.ExecRelease(); err != nil {
		log.Errorf("error updating user course: %v", err)
		return nil, err
	}
	created := strconv.FormatInt(userLspMap.CreatedAt, 10)
	updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
	userLspOutput := &model.UserCourse{
		UserCourseID: &userLspMap.ID,
		UserLspID:    userLspMap.UserLspID,
		UserID:       userLspMap.UserID,
		CourseID:     userLspMap.CourseID,
		CourseType:   userLspMap.CourseType,
		CourseStatus: userLspMap.CourseStatus,
		AddedBy:      userLspMap.AddedBy,
		IsMandatory:  userLspMap.IsMandatory,
		EndDate:      input.EndDate,
		CreatedAt:    created,
		UpdatedAt:    updated,
		CreatedBy:    &userLspMap.CreatedBy,
		UpdatedBy:    &userLspMap.UpdatedBy,
	}
	return userLspOutput, nil
}
