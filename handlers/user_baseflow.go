package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/scylladb/gocqlx/qb"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
)

func RegisterUsers(ctx context.Context, input []*model.UserInput) ([]*model.User, error) {
	var outputUsers []*model.User
	for _, user := range input {
		_, err := global.IDP.RegisterUser(ctx, user.Email, user.FirstName, user.LastName, user.Phone)
		if err != nil {
			return nil, err
		}
		userID := base64.URLEncoding.EncodeToString([]byte(user.Email))
		userCass := userz.User{
			ID:         userID,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			Gender:     user.Gender,
			Status:     user.Status,
			Role:       user.Role,
			IsVerified: user.IsVerified,
			IsActive:   user.IsActive,
			CreatedBy:  user.CreatedBy,
			UpdatedBy:  user.UpdatedBy,
			CreatedAt:  time.Now().Unix(),
			UpdatedAt:  time.Now().Unix(),
		}
		insertQuery := global.CassUserSession.Session.Query(userz.UserTable.Insert()).BindStruct(userCass)
		if err := insertQuery.ExecRelease(); err != nil {
			return nil, err
		}
		created := strconv.FormatInt(userCass.CreatedAt, 10)
		updated := strconv.FormatInt(userCass.UpdatedAt, 10)
		responseUser := model.User{
			ID:         &userID,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			Email:      user.Email,
			Phone:      user.Phone,
			CreatedAt:  created,
			UpdatedAt:  updated,
			CreatedBy:  user.CreatedBy,
			UpdatedBy:  user.UpdatedBy,
			Gender:     user.Gender,
			IsVerified: user.IsVerified,
			IsActive:   user.IsActive,
			Role:       user.Role,
			Status:     user.Status,
		}
		outputUsers = append(outputUsers, &responseUser)
	}
	return outputUsers, nil
}

func InviteUsers(ctx context.Context, emails []string) (*bool, error) {
	registered := false
	for _, email := range emails {
		userRecord, err := global.IDP.InviteUser(ctx, email)
		if err != nil {
			return &registered, err
		}
		passwordReset, err := global.IDP.GetResetPasswordURL(ctx, userRecord.Email)
		if err != nil {
			return &registered, err
		}
		// send email with password reset link
		global.SGClient.SendJoinEmail(userRecord.Email, passwordReset, userRecord.DisplayName)
	}
	registered = true
	return &registered, nil
}

func UpdateUser(ctx context.Context, user model.UserInput) (*model.User, error) {
	userID := user.ID
	if userID == nil {
		return nil, fmt.Errorf("userID is empty")
	}
	userCass := userz.User{
		ID: *userID,
	}
	banks := []userz.User{}
	getQuery := global.CassUserSession.Session.Query(userz.UserTable.Get()).BindMap(qb.M{"id": userCass.ID})
	if err := getQuery.SelectRelease(&banks); err != nil {
		return nil, err
	}
	if len(banks) == 0 {
		return nil, fmt.Errorf("exams not found")
	}
	userCass = banks[0]
	updatedCols := []string{}
	if user.Status != "" {
		userCass.Status = user.Status
		updatedCols = append(updatedCols, "status")
	}
	if user.Gender != "" {
		userCass.Gender = user.Gender
		updatedCols = append(updatedCols, "gender")
	}
	userCass.IsActive = user.IsActive
	userCass.IsVerified = user.IsVerified
	if user.CreatedBy != "" {
		userCass.CreatedBy = user.CreatedBy
		updatedCols = append(updatedCols, "created_by")
	}
	if user.UpdatedBy != "" {
		userCass.UpdatedBy = user.UpdatedBy
		updatedCols = append(updatedCols, "updated_by")
	}
	if user.Role != "" {
		userCass.Role = user.Role
		updatedCols = append(updatedCols, "role")
	}
	updatedAt := time.Now().Unix()
	userCass.UpdatedAt = updatedAt
	updatedCols = append(updatedCols, "updated_at")
	created := strconv.FormatInt(userCass.CreatedAt, 10)
	updated := strconv.FormatInt(userCass.UpdatedAt, 10)
	if len(updatedCols) == 0 {
		return nil, fmt.Errorf("nothing to update")
	}
	upStms, uNames := userz.UserTable.Update(updatedCols...)
	updateQuery := global.CassUserSession.Session.Query(upStms, uNames).BindStruct(&userCass)
	if err := updateQuery.ExecRelease(); err != nil {
		return nil, err
	}
	responseUser := model.User{
		ID:         &userCass.ID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Email:      user.Email,
		Phone:      user.Phone,
		CreatedAt:  created,
		UpdatedAt:  updated,
		CreatedBy:  user.CreatedBy,
		UpdatedBy:  user.UpdatedBy,
		Gender:     user.Gender,
		IsVerified: user.IsVerified,
		IsActive:   user.IsActive,
		Role:       user.Role,
		Status:     user.Status,
	}
	return &responseUser, nil
}

func UpdateUserEmail(ctx context.Context, userID string, email string) (*bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func UpdateUserPhone(ctx context.Context, userID string, phone string) (*bool, error) {
	panic(fmt.Errorf("not implemented"))
}
