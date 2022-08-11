package queries

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/scylladb/gocqlx/qb"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/helpers"
	"github.com/zicops/zicops-user-manager/lib/db/bucket"
	"github.com/zicops/zicops-user-manager/lib/googleprojectlib"
)

func GetUsersForAdmin(ctx context.Context, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedUsers, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	userAdmin := userz.User{
		ID: emailCreatorID,
	}
	users := []userz.User{}
	getQuery := global.CassUserSession.Session.Query(userz.UserTable.Get()).BindMap(qb.M{"id": userAdmin.ID})
	if err := getQuery.SelectRelease(&users); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	userAdmin = users[0]
	if strings.ToLower(userAdmin.Role) != "admin" {
		return nil, fmt.Errorf("user is not an admin")
	}
	var newPage []byte
	//var pageDirection string
	var pageSizeInt int
	if pageCursor != nil && *pageCursor != "" {
		page, err := global.CryptSession.DecryptString(*pageCursor, nil)
		if err != nil {
			return nil, fmt.Errorf("invalid page cursor: %v", err)
		}
		newPage = page
	}
	if pageSize == nil {
		pageSizeInt = 10
	} else {
		pageSizeInt = *pageSize
	}
	var newCursor string

	qryStr := fmt.Sprintf(`SELECT * from userz.users where created_by='%s' and updated_at <= %d  ALLOW FILTERING`, email_creator, *publishTime)
	getUsers := func(page []byte) (users []userz.User, nextPage []byte, err error) {
		q := global.CassUserSession.Session.Query(qryStr, nil)
		defer q.Release()
		q.PageState(page)
		q.PageSize(pageSizeInt)

		iter := q.Iter()
		return users, iter.PageState(), iter.Select(&users)
	}
	users, newPage, err = getUsers(newPage)
	if err != nil {
		return nil, err
	}
	if len(newPage) != 0 {
		newCursor, err = global.CryptSession.EncryptAsString(newPage, nil)
		if err != nil {
			return nil, fmt.Errorf("error encrypting cursor: %v", err)
		}
		log.Infof("Users: %v", string(newCursor))

	}
	var outputResponse model.PaginatedUsers
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		log.Errorf("Failed to upload image to course: %v", err.Error())
		return nil, err
	}
	allUsers := make([]*model.User, 0)
	for _, copiedUser := range users {
		userCopy := copiedUser
		createdAt := strconv.FormatInt(userCopy.CreatedAt, 10)
		updatedAt := strconv.FormatInt(userCopy.UpdatedAt, 10)
		photoUrl := ""
		if userCopy.PhotoBucket != "" {
			photoUrl = storageC.GetSignedURLForObject(userCopy.PhotoBucket)
		} else {
			photoUrl = userCopy.PhotoURL
		}
		fireBaseUser, err := global.IDP.GetUserByEmail(ctx, userCopy.Email)
		if err != nil {
			log.Errorf("Failed to get user from firebase: %v", err.Error())
		}
		currentUser := &model.User{
			ID:         &userCopy.ID,
			Email:      userCopy.Email,
			FirstName:  userCopy.FirstName,
			LastName:   userCopy.LastName,
			Role:       userCopy.Role,
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
			PhotoURL:   &photoUrl,
			IsVerified: userCopy.IsVerified,
			IsActive:   userCopy.IsActive,
			CreatedBy:  &userCopy.CreatedBy,
			UpdatedBy:  &userCopy.UpdatedBy,
			Status:     userCopy.Status,
			Gender:     userCopy.Gender,
			Phone:      fireBaseUser.PhoneNumber,
		}
		allUsers = append(allUsers, currentUser)
	}
	outputResponse.Users = allUsers
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction
	return &outputResponse, nil
}

func GetUserDetails(ctx context.Context, userID string) (*model.User, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	userAdmin := userz.User{
		ID: emailCreatorID,
	}
	users := []userz.User{}
	getQuery := global.CassUserSession.Session.Query(userz.UserTable.Get()).BindMap(qb.M{"id": userAdmin.ID})
	if err := getQuery.SelectRelease(&users); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	userAdmin = users[0]
	if strings.ToLower(userAdmin.Role) != "admin" {
		return nil, fmt.Errorf("user is not an admin")
	}
	qryStr := fmt.Sprintf(`SELECT * from userz.users where id='%s' ALLOW FILTERING`, userID)
	getUsers := func() (users []userz.User, err error) {
		q := global.CassUserSession.Session.Query(qryStr, nil)
		defer q.Release()

		iter := q.Iter()
		return users, iter.Select(&users)
	}
	users, err = getUsers()
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		log.Errorf("Failed to upload image to course: %v", err.Error())
		return nil, err
	}
	userCopy := users[0]
	createdAt := strconv.FormatInt(userCopy.CreatedAt, 10)
	updatedAt := strconv.FormatInt(userCopy.UpdatedAt, 10)
	photoUrl := ""
	if userCopy.PhotoBucket != "" {
		photoUrl = storageC.GetSignedURLForObject(userCopy.PhotoBucket)
	} else {
		photoUrl = userCopy.PhotoURL
	}
	fireBaseUser, err := global.IDP.GetUserByEmail(ctx, userCopy.Email)
	if err != nil {
		log.Errorf("Failed to get user from firebase: %v", err.Error())
	}
	outputUser := &model.User{
		ID:         &userCopy.ID,
		Email:      userCopy.Email,
		FirstName:  userCopy.FirstName,
		LastName:   userCopy.LastName,
		Role:       userCopy.Role,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		PhotoURL:   &photoUrl,
		IsVerified: userCopy.IsVerified,
		IsActive:   userCopy.IsActive,
		CreatedBy:  &userCopy.CreatedBy,
		UpdatedBy:  &userCopy.UpdatedBy,
		Status:     userCopy.Status,
		Gender:     userCopy.Gender,
		Phone:      fireBaseUser.PhoneNumber,
	}

	return outputUser, nil
}
