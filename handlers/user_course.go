package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
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
	CassUserSession := session

	/*
		isAllowed := false
		role := strings.ToLower(userCass.Role)
		if userCass.ID == input[0].UserID || role == "admin" || strings.Contains(role, "manager") {
			isAllowed = true
		}
		if !isAllowed {
			return nil, fmt.Errorf("user not allowed to create org mapping")
		}
	*/
	userLspMaps := make([]*model.UserCourse, 0)
	for _, input := range input {

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
			ID:           uuid.New().String(),
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
		insertQuery := CassUserSession.Query(userz.UserCourseTable.Insert()).BindStruct(userLspMap)
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
	role := strings.ToLower(userCass.Role)
	if userCass.ID == input.UserID || role == "admin" || strings.Contains(role, "manager") {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create org mapping")
	}
	if input.UserCourseID == nil {
		return nil, fmt.Errorf("user course id is required")
	}
	if input.UserID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userLspMap := userz.UserCourse{
		ID: *input.UserCourseID,
	}
	userLsps := []userz.UserCourse{}

	getQueryStr := fmt.Sprintf("SELECT * FROM userz.user_course_map WHERE id='%s' AND user_id='%s'  ", userLspMap.ID, input.UserID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
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

	if len(updatedCols) > 0 {
		updatedAt := time.Now().Unix()
		userLspMap.UpdatedAt = updatedAt
		updatedCols = append(updatedCols, "updated_at")
		upStms, uNames := userz.UserCourseTable.Update(updatedCols...)
		updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&userLspMap)
		if err := updateQuery.ExecRelease(); err != nil {
			log.Errorf("error updating user course: %v", err)
			return nil, err
		}
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
