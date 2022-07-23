package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/scylladb/gocqlx/qb"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/helpers"
	"github.com/zicops/zicops-user-manager/lib/db/bucket"
	"github.com/zicops/zicops-user-manager/lib/googleprojectlib"
)

func RegisterUsers(ctx context.Context, input []*model.UserInput, isZAdmin bool) ([]*model.User, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	roleValue := claims["email"]
	if !isZAdmin || strings.ToLower(roleValue.(string)) != "puneet@zicops.com" {
		return nil, fmt.Errorf("user is a not an admin: Unauthorized")
	}
	var outputUsers []*model.User
	var storageC *bucket.Client
	var photoBucket string
	var photoUrl string
	for _, user := range input {
		userID := base64.URLEncoding.EncodeToString([]byte(user.Email))
		if user.Photo != nil && user.PhotoURL == nil {
			if storageC == nil {
				storageC = bucket.NewStorageHandler()
				gproject := googleprojectlib.GetGoogleProjectID()
				err := storageC.InitializeStorageClient(ctx, gproject)
				if err != nil {
					return nil, err
				}
			}
			bucketPath := fmt.Sprintf("%s/%s/%s", "profiles", userID, user.Photo.Filename)
			writer, err := storageC.UploadToGCS(ctx, bucketPath)
			if err != nil {
				return nil, err
			}
			defer writer.Close()
			fileBuffer := bytes.NewBuffer(nil)
			if _, err := io.Copy(fileBuffer, user.Photo.File); err != nil {
				return nil, err
			}
			currentBytes := fileBuffer.Bytes()
			_, err = io.Copy(writer, bytes.NewReader(currentBytes))
			if err != nil {
				return nil, err
			}
			photoBucket = bucketPath
			photoUrl = storageC.GetSignedURLForObject(bucketPath)
		} else {
			photoBucket = ""
			if user.PhotoURL != nil {
				photoUrl = *user.PhotoURL
			}
		}
		_, err := global.IDP.RegisterUser(ctx, user.Email, user.FirstName, user.LastName, user.Phone)
		if err != nil {
			return nil, err
		}

		userCass := userz.User{
			ID:          userID,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			Gender:      user.Gender,
			Status:      user.Status,
			Role:        user.Role,
			IsVerified:  user.IsVerified,
			IsActive:    user.IsActive,
			CreatedBy:   user.CreatedBy,
			UpdatedBy:   user.UpdatedBy,
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
			PhotoBucket: photoBucket,
			PhotoURL:    photoUrl,
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
			PhotoURL:   &photoUrl,
		}
		passwordReset, err := global.IDP.GetResetPasswordURL(ctx, responseUser.Email)
		if err != nil {
			return nil, err
		}
		global.SGClient.SendJoinEmail(responseUser.Email, passwordReset, responseUser.FirstName+" "+responseUser.LastName)

		outputUsers = append(outputUsers, &responseUser)
	}
	return outputUsers, nil
}

func InviteUsers(ctx context.Context, emails []string) (*bool, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	registered := false
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
	if strings.ToLower(userCass.Role) != "admin" {
		return nil, fmt.Errorf("user is not an admin")
	}
	for _, email := range emails {
		userRecord, err := global.IDP.InviteUser(ctx, email)
		if err != nil {
			return &registered, err
		}
		passwordReset, err := global.IDP.GetResetPasswordURL(ctx, userRecord.Email)
		if err != nil {
			return &registered, err
		}
		userID := base64.URLEncoding.EncodeToString([]byte(email))
		userInput := model.UserInput{
			ID:         &userID,
			FirstName:  "",
			LastName:   "",
			Email:      userRecord.Email,
			Role:       "",
			Status:     "",
			IsVerified: false,
			IsActive:   false,
			CreatedBy:  email_creator,
			UpdatedBy:  email_creator,
			Photo:      nil,
			PhotoURL:   nil,
			Gender:     "",
			Phone:      "",
		}
		_, err = RegisterUsers(ctx, []*model.UserInput{&userInput}, true)
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
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if user.ID == nil {
		return nil, fmt.Errorf("user id is required")
	}
	canUpdate := false
	userID := base64.URLEncoding.EncodeToString([]byte(user.Email))
	if err != nil {
		return nil, err
	}
	token_email := claims["email"].(string)
	userId := base64.URLEncoding.EncodeToString([]byte(token_email))
	if userId == *user.ID {
		canUpdate = true
	}
	if !canUpdate {
		roleValue := claims["role"]
		if roleValue == nil || strings.ToLower(roleValue.(string)) != "admin" {
			return nil, fmt.Errorf("user is a not an admin: unauthorized")
		}
	}
	userCass := userz.User{
		ID: userID,
	}
	users := []userz.User{}
	getQuery := global.CassUserSession.Session.Query(userz.UserTable.Get()).BindMap(qb.M{"id": userCass.ID})
	if err := getQuery.SelectRelease(&users); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("exams not found")
	}
	userCass = users[0]
	updatedCols := []string{}
	emailUpdate := ""
	phoneUpdate := ""
	firstNameUpdate := ""
	lastNameUpdate := ""
	var storageC *bucket.Client
	var photoBucket string
	var photoUrl string
	fireUser, err := global.IDP.GetUserByEmail(ctx, userCass.Email)
	if err != nil {
		return nil, err
	}
	if user.Photo != nil && user.PhotoURL == nil {
		if storageC == nil {
			storageC = bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err := storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				return nil, err
			}
		}
		bucketPath := fmt.Sprintf("%s/%s/%s", "profiles", userCass.ID, user.Photo.Filename)
		writer, err := storageC.UploadToGCS(ctx, bucketPath)
		if err != nil {
			return nil, err
		}
		defer writer.Close()
		fileBuffer := bytes.NewBuffer(nil)
		if _, err := io.Copy(fileBuffer, user.Photo.File); err != nil {
			return nil, err
		}
		currentBytes := fileBuffer.Bytes()
		_, err = io.Copy(writer, bytes.NewReader(currentBytes))
		if err != nil {
			return nil, err
		}
		photoBucket = bucketPath
		photoUrl = storageC.GetSignedURLForObject(bucketPath)
	} else {
		photoBucket = ""
		photoUrl = *user.PhotoURL
	}
	if photoBucket != "" {
		userCass.PhotoBucket = photoBucket
		updatedCols = append(updatedCols, "photo_bucket")
	}
	if photoUrl != "" {
		userCass.PhotoURL = photoUrl
		updatedCols = append(updatedCols, "photo_url")
	}
	if user.Email != "" && user.Email != userCass.Email {
		userCass.Email = user.Email
		updatedCols = append(updatedCols, "email")
		emailUpdate = user.Email
	}
	if user.Phone != "" {
		phoneUpdate = user.Phone
	}
	if user.FirstName != "" && user.FirstName != userCass.FirstName {
		userCass.FirstName = user.FirstName
		updatedCols = append(updatedCols, "firstname")
		firstNameUpdate = user.FirstName
	}
	if user.LastName != "" && user.LastName != userCass.LastName {
		userCass.LastName = user.LastName
		updatedCols = append(updatedCols, "lastname")
		lastNameUpdate = user.LastName
	}
	if emailUpdate != "" || phoneUpdate != "" || firstNameUpdate != "" || lastNameUpdate != "" {
		_, err := global.IDP.UpdateUser(ctx, emailUpdate, firstNameUpdate, lastNameUpdate, phoneUpdate, fireUser.UID)
		if err != nil {
			return nil, err
		}
	}
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

func LoginUser(ctx context.Context) (*model.User, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userEmail := claims["email"].(string)
	if userEmail == "puneet@zicops.com" {
		return nil, fmt.Errorf("user is not allowed to proceed with zicops apis")
	}
	userID := base64.URLEncoding.EncodeToString([]byte(userEmail))
	userCass := userz.User{
		ID: userID,
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
	photoURL := userCass.PhotoURL
	if userCass.PhotoBucket != "" {
		storageC := bucket.NewStorageHandler()
		gproject := googleprojectlib.GetGoogleProjectID()
		err := storageC.InitializeStorageClient(ctx, gproject)
		if err != nil {
			return nil, err
		}
		photoURL = storageC.GetSignedURLForObject(userCass.PhotoBucket)
	}
	currentUser := model.User{
		ID:         &userCass.ID,
		FirstName:  userCass.FirstName,
		LastName:   userCass.LastName,
		Email:      userCass.Email,
		CreatedAt:  strconv.FormatInt(userCass.CreatedAt, 10),
		UpdatedAt:  strconv.FormatInt(userCass.UpdatedAt, 10),
		CreatedBy:  userCass.CreatedBy,
		UpdatedBy:  userCass.UpdatedBy,
		Role:       userCass.Role,
		Status:     userCass.Status,
		Gender:     userCass.Gender,
		IsVerified: userCass.IsVerified,
		IsActive:   userCass.IsActive,
		PhotoURL:   &photoURL,
	}
	return &currentUser, nil
}

func Logout(ctx context.Context) (*bool, error) {
	logoutSuccess := false
	email := ctx.Value("email").(string)
	logoutSuccess, err := global.IDP.LogoutUser(ctx, email)
	if err != nil {
		return &logoutSuccess, err
	}
	return &logoutSuccess, nil
}
