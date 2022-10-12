package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/scylladb/gocqlx/qb"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-cass-pool/redis"
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
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	roleValue := claims["email"]
	if !isZAdmin {
		if strings.ToLower(roleValue.(string)) != "puneet@zicops.com" {
			return nil, fmt.Errorf("user is a not an admin: Unauthorized")
		}
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
			log.Errorf("error while registering user: %v", err)
		}
		if user.CreatedBy == nil {
			user.CreatedBy = &user.Email
		}
		if user.UpdatedBy == nil {
			user.UpdatedBy = &user.Email
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
			CreatedBy:   *user.CreatedBy,
			UpdatedBy:   *user.UpdatedBy,
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
			PhotoBucket: photoBucket,
			PhotoURL:    photoUrl,
			Email:       user.Email,
		}
		insertQuery := CassUserSession.Query(userz.UserTable.Insert()).BindStruct(userCass)
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
		if isZAdmin {
			passwordReset, err := global.IDP.GetResetPasswordURL(ctx, responseUser.Email)
			if err != nil {
				return nil, err
			}
			global.SGClient.SendJoinEmail(responseUser.Email, passwordReset, responseUser.FirstName+" "+responseUser.LastName)
		}
		outputUsers = append(outputUsers, &responseUser)
	}
	return outputUsers, nil
}

func InviteUsers(ctx context.Context, emails []string, lspID string) (*bool, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	registered := false
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	userCass := userz.User{
		ID: emailCreatorID,
	}
	users := []userz.User{}
	getQuery := CassUserSession.Query(userz.UserTable.Get()).BindMap(qb.M{"id": userCass.ID})
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
		if email == email_creator {
			log.Errorf("user %v is trying to invite himself", email_creator)
			continue
		}
		userID := base64.URLEncoding.EncodeToString([]byte(email))
		userInput := model.UserInput{
			ID:         &userID,
			FirstName:  "",
			LastName:   "",
			Email:      email,
			Role:       "",
			Status:     "",
			IsVerified: false,
			IsActive:   false,
			CreatedBy:  &email_creator,
			UpdatedBy:  &email_creator,
			Photo:      nil,
			PhotoURL:   nil,
			Gender:     "",
			Phone:      "",
		}
		_, err = RegisterUsers(ctx, []*model.UserInput{&userInput}, true)
		if err != nil {
			return &registered, err
		}
		userLspMap := &model.UserLspMapInput{
			UserID: userID,
			LspID:  lspID,
			Status: "",
		}
		_, err = AddUserLspMap(ctx, []*model.UserLspMapInput{userLspMap})
		if err != nil {
			return &registered, err
		}
		// passwordReset, err := global.IDP.GetResetPasswordURL(ctx, email)
		// if err != nil {
		// 	return &registered, err
		// }
		// // send email with password reset link
		// global.SGClient.SendJoinEmail(email, passwordReset, userCass.FirstName+" "+userCass.LastName)
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
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

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
	getQuery := CassUserSession.Query(userz.UserTable.Get()).BindMap(qb.M{"id": userCass.ID})
	if err := getQuery.SelectRelease(&users); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("users not found")
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
		if user.PhotoURL != nil {
			photoUrl = *user.PhotoURL
		}
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
	if user.Phone != "" && user.Phone != fireUser.PhoneNumber {
		phoneUpdate = user.Phone
	}
	if user.FirstName != "" && user.FirstName != userCass.FirstName {
		userCass.FirstName = user.FirstName
		updatedCols = append(updatedCols, "first_name")
		firstNameUpdate = user.FirstName
	}
	if user.LastName != "" && user.LastName != userCass.LastName {
		userCass.LastName = user.LastName
		updatedCols = append(updatedCols, "last_name")
		lastNameUpdate = user.LastName
		if firstNameUpdate == "" {
			firstNameUpdate = user.FirstName
		}
	}
	if firstNameUpdate != "" || lastNameUpdate == "" {
		lastNameUpdate = user.LastName
	}

	if emailUpdate != "" || firstNameUpdate != "" || lastNameUpdate != "" || phoneUpdate != "" {
		_, err := global.IDP.UpdateUser(ctx, emailUpdate, firstNameUpdate, lastNameUpdate, phoneUpdate, fireUser.UID)
		if err != nil {
			return nil, err
		}
	}
	if user.Status != "" && user.Status != userCass.Status {
		userCass.Status = user.Status
		updatedCols = append(updatedCols, "status")
	}
	if user.Gender != "" && user.Gender != userCass.Gender {
		userCass.Gender = user.Gender
		updatedCols = append(updatedCols, "gender")
	}
	if user.CreatedBy != nil && userCass.CreatedBy != *user.CreatedBy {
		userCass.CreatedBy = *user.CreatedBy
		updatedCols = append(updatedCols, "created_by")
	}
	if user.UpdatedBy != nil && *user.UpdatedBy != userCass.UpdatedBy {
		userCass.UpdatedBy = *user.UpdatedBy
		updatedCols = append(updatedCols, "updated_by")
	}
	if user.Role != "" && user.Role != userCass.Role {
		userCass.Role = user.Role
		updatedCols = append(updatedCols, "role")
	}
	if user.IsActive != userCass.IsActive {
		userCass.IsActive = user.IsActive
		updatedCols = append(updatedCols, "is_active")
	}
	if user.IsVerified != userCass.IsVerified {
		userCass.IsVerified = user.IsVerified
		updatedCols = append(updatedCols, "is_verified")
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
	updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&userCass)
	if err := updateQuery.ExecRelease(); err != nil {
		log.Errorf("error updating user: %v", err)
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
		PhotoURL:   &photoUrl,
	}
	userBytes, err := json.Marshal(user)
	if err == nil {
		redis.SetRedisValue(*user.ID, string(userBytes))
		redis.SetTTL(*user.ID, 3600)
	}
	return &responseUser, nil
}

func LoginUser(ctx context.Context) (*model.User, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userEmail := claims["email"].(string)
	if userEmail == "puneet@zicops.com" {
		return nil, fmt.Errorf("user is not allowed to proceed with zicops apis")
	}
	currentUserIT, err := global.IDP.GetUserByEmail(ctx, userEmail)
	if err != nil {
		return nil, err
	}
	userID := base64.URLEncoding.EncodeToString([]byte(userEmail))
	userCass := userz.User{
		ID: userID,
	}
	users := []userz.User{}
	getQuery := CassUserSession.Query(userz.UserTable.Get()).BindMap(qb.M{"id": userCass.ID})
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
		CreatedBy:  &userCass.CreatedBy,
		UpdatedBy:  &userCass.UpdatedBy,
		Role:       userCass.Role,
		Status:     userCass.Status,
		Gender:     userCass.Gender,
		IsVerified: userCass.IsVerified,
		IsActive:   userCass.IsActive,
		PhotoURL:   &photoURL,
		Phone:      currentUserIT.PhoneNumber,
	}
	userBytes, err := json.Marshal(userCass)
	if err == nil {
		redis.SetRedisValue(userCass.ID, string(userBytes))
		redis.SetTTL(userCass.ID, 7200)
		log.Infof("user logged in: %v", userCass.ID)
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
	redis.DeleteRedisValue(base64.URLEncoding.EncodeToString([]byte(email)))
	return &logoutSuccess, nil
}
