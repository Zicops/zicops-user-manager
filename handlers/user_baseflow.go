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

	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/lib/db/bucket"
	"github.com/zicops/zicops-user-manager/lib/googleprojectlib"
	"github.com/zicops/zicops-user-manager/lib/identity"
)

func RegisterUsers(ctx context.Context, input []*model.UserInput, isZAdmin bool, userExists bool) ([]*model.User, []*model.UserLspMap, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, nil, err
	}
	CassUserSession := session

	roleValue := claims["email"]
	//name := claims["name"].(string)
	lspId := claims["lsp_id"].(string)
	origin := claims["origin"].(string)
	if origin == "" {
		origin = "https://demo.zicops.com"
	}
	if !isZAdmin {
		if strings.ToLower(roleValue.(string)) != "puneet@zicops.com" {
			return nil, nil, fmt.Errorf("user is a not an admin: Unauthorized")
		}
	}
	var outputUsers []*model.User
	var usrLspMaps []*model.UserLspMap
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		return nil, nil, err
	}
	var photoBucket string
	var photoUrl string
	qryStr := fmt.Sprintf(`SELECT * from userz.learning_space where id='%s' ALLOW FILTERING `, lspId)
	getOrgs := func() (users []userz.Lsp, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return users, iter.Select(&users)
	}
	orgs, err := getOrgs()
	if err != nil {
		log.Errorf("error getting lsp: %v", err)
		return nil, nil, err
	}
	if len(orgs) == 0 {
		return nil, nil, fmt.Errorf("lsp not found")
	}
	lspToAdd := orgs[0]
	for _, user := range input {
		if user == nil {
			continue
		}
		emailLower := strings.ToLower(user.Email)
		userID := base64.URLEncoding.EncodeToString([]byte(emailLower))
		if user.Photo != nil && user.PhotoURL == nil {
			extension := strings.Split(user.Photo.Filename, ".")
			bucketPath := fmt.Sprintf("%s/%s/%s", "profiles", userID, base64.URLEncoding.EncodeToString([]byte(user.Photo.Filename)))
			if len(extension) >= 1 {
				bucketPath += "." + extension[len(extension)-1]
			}
			writer, err := storageC.UploadToGCS(ctx, bucketPath)
			if err != nil {
				return nil, nil, err
			}
			defer writer.Close()
			fileBuffer := bytes.NewBuffer(nil)
			if _, err := io.Copy(fileBuffer, user.Photo.File); err != nil {
				return nil, nil, err
			}
			currentBytes := fileBuffer.Bytes()
			_, err = io.Copy(writer, bytes.NewReader(currentBytes))
			if err != nil {
				return nil, nil, err
			}
			photoBucket = bucketPath
			photoUrl = storageC.GetSignedURLForObject(ctx, bucketPath)
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
		if !userExists {
			insertQuery := CassUserSession.Query(userz.UserTable.Insert()).BindStruct(userCass)
			if err := insertQuery.ExecRelease(); err != nil {
				return nil, nil, err
			}
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

		outputUsers = append(outputUsers, &responseUser)
		userLspMap := &model.UserLspMapInput{
			UserID:    userID,
			LspID:     lspId,
			Status:    "",
			CreatedBy: user.CreatedBy,
			UpdatedBy: user.UpdatedBy,
		}
		isAdminCall := true
		usrLspMap, err := AddUserLspMap(ctx, []*model.UserLspMapInput{userLspMap}, &isAdminCall)
		if err != nil {
			return nil, nil, err
		}
		shouldSendEmail := false
		for _, usrLsp := range usrLspMap {
			if usrLsp == nil {
				continue
			}
			if usrLsp.Status == "" {
				shouldSendEmail = true
			}
		}
		usrLspMaps = append(usrLspMaps, usrLspMap...)
		if shouldSendEmail {
			if isZAdmin && !userExists {
				passwordReset, err := global.IDP.GetResetPasswordURL(ctx, responseUser.Email, origin, origin)
				//log.Println("origin in new admin", origin)
				if err != nil {
					return nil, nil, err
				}
				global.SGClient.SendJoinEmail(responseUser.Email, passwordReset, responseUser.FirstName+" "+responseUser.LastName, lspToAdd.Name, lspToAdd.Name)
			} else if isZAdmin && userExists {
				err := global.SGClient.SendInviteToLspEmail(responseUser.Email, origin+"/login", lspToAdd.Name, lspToAdd.Name)
				//log.Println(origin)
				if err != nil {
					log.Printf("Got error while sending invite to lsp: %v", err)
				}
			}
		}

	}
	return outputUsers, usrLspMaps, nil
}

func InviteUsers(ctx context.Context, emails []string, lspID string) (*bool, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	registered := false
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	users := []userz.User{}

	getQueryStr := fmt.Sprintf(`SELECT * from userz.users where id='%s' `, emailCreatorID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
	if err := getQuery.SelectRelease(&users); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	for _, dirtyEmail := range emails {
		emailRaw := strings.TrimSpace(dirtyEmail)
		email := strings.ToLower(emailRaw)
		if email == email_creator {
			log.Errorf("user %v is trying to invite himself", email_creator)
			continue
		}
		users := []userz.User{}
		emailLower := strings.ToLower(email)
		userID := base64.URLEncoding.EncodeToString([]byte(emailLower))
		getQueryStr := fmt.Sprintf(`SELECT * from userz.users where id='%s' `, userID)
		getQuery := CassUserSession.Query(getQueryStr, nil)
		if err := getQuery.SelectRelease(&users); err != nil {
			return nil, err
		}
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
		_, lspMaps, err := RegisterUsers(ctx, []*model.UserInput{&userInput}, true, len(users) > 0)
		if err != nil {
			return &registered, err
		}
		userRoleMap := &model.UserRoleInput{
			UserID:    userID,
			Role:      "learner",
			UserLspID: *lspMaps[0].UserLspID,
			IsActive:  true,
			CreatedBy: &email_creator,
			UpdatedBy: &email_creator,
		}
		if lspMaps[0].Status == "" {
			_, err = AddUserRoles(ctx, []*model.UserRoleInput{userRoleMap})
			if err != nil {
				return &registered, err
			}
		}
	}
	registered = true
	return &registered, nil
}

func UpdateUser(ctx context.Context, user model.UserInput) (*model.User, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if user.ID == nil {
		return nil, fmt.Errorf("user id is required")
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	canUpdate := false
	emailLower := strings.ToLower(user.Email)
	userID := base64.URLEncoding.EncodeToString([]byte(emailLower))
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

	getQueryStr := fmt.Sprintf(`SELECT * from userz.users where id='%s' `, userID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
	if err := getQuery.SelectRelease(&users); err != nil {
		return nil, err
	}
	userCass = users[0]
	updatedCols := []string{}
	emailUpdate := ""
	phoneUpdate := ""
	firstNameUpdate := ""
	lastNameUpdate := ""
	var photoBucket string
	var photoUrl string
	fireUser, err := global.IDP.GetUserByEmail(ctx, userCass.Email)
	if err != nil {
		return nil, err
	}
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		return nil, err
	}
	if user.Photo != nil && user.PhotoURL == nil {
		extension := strings.Split(user.Photo.Filename, ".")
		bucketPath := fmt.Sprintf("%s/%s/%s", "profiles", *user.ID, base64.URLEncoding.EncodeToString([]byte(user.Photo.Filename)))
		if len(extension) >= 1 {
			bucketPath += "." + extension[len(extension)-1]
		}
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
		photoUrl = storageC.GetSignedURLForObject(ctx, bucketPath)
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
	if user.IsVerified != userCass.IsVerified {
		userCass.IsVerified = user.IsVerified
		updatedCols = append(updatedCols, "is_verified")
	}
	created := strconv.FormatInt(userCass.CreatedAt, 10)
	updated := strconv.FormatInt(userCass.UpdatedAt, 10)
	if len(updatedCols) > 0 {
		updatedAt := time.Now().Unix()
		userCass.UpdatedAt = updatedAt
		updatedCols = append(updatedCols, "updated_at")
		upStms, uNames := userz.UserTable.Update(updatedCols...)
		updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&userCass)
		if err := updateQuery.ExecRelease(); err != nil {
			log.Errorf("error updating user: %v", err)
			return nil, err
		}
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
		redis.SetRedisValue(ctx, *user.ID, string(userBytes))
		redis.SetTTL(ctx, *user.ID, 3600)
	}
	return &responseUser, nil
}

func LoginUser(ctx context.Context) (*model.User, error) {

	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userEmail := claims["email"].(string)
	if userEmail == "puneet@zicops.com" {
		return nil, fmt.Errorf("user is not allowed to proceed with zicops apis")
	}
	emailLower := strings.ToLower(userEmail)
	userID := base64.URLEncoding.EncodeToString([]byte(emailLower))
	userCass := userz.User{
		ID: userID,
	}
	result, err := redis.GetRedisValue(ctx, userID)
	if err == nil && result != "" {
		err = json.Unmarshal([]byte(result), &userCass)
	}
	phone := ""
	if err == nil && userCass.ID != "" {
		session, err := global.CassPool.GetSession(ctx, "userz")
		if err != nil {
			return nil, err
		}
		CassUserSession := session
		currentUserIT, err := global.IDP.GetUserByEmail(ctx, emailLower)
		if err != nil {
			return nil, err
		}
		phone = currentUserIT.PhoneNumber
		users := []userz.User{}

		getQueryStr := fmt.Sprintf(`SELECT * from userz.users where id='%s' `, userID)
		getQuery := CassUserSession.Query(getQueryStr, nil)
		if err := getQuery.SelectRelease(&users); err != nil {
			return nil, err
		}
		if len(users) == 0 {
			return nil, fmt.Errorf("user not found")
		}
		userCass = users[0]
	}
	photoURL := userCass.PhotoURL
	if userCass.PhotoBucket != "" {
		storageC := bucket.NewStorageHandler()
		gproject := googleprojectlib.GetGoogleProjectID()
		err := storageC.InitializeStorageClient(ctx, gproject)
		if err != nil {
			return nil, err
		}
		photoURL = storageC.GetSignedURLForObject(ctx, userCass.PhotoBucket)
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
		Phone:      phone,
	}

	userBytes, err := json.Marshal(userCass)
	if err == nil {
		redis.SetRedisValue(ctx, userCass.ID, string(userBytes))
		redis.SetTTL(ctx, userCass.ID, 7200)
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
	emailLower := strings.ToLower(email)
	redis.SetRedisValue(ctx, base64.URLEncoding.EncodeToString([]byte(emailLower)), "")
	return &logoutSuccess, nil
}
