package queries

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/scylladb/gocqlx/qb"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-cass-pool/redis"
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
	var outputResponse model.PaginatedUsers
	users := make([]userz.User, 0)
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
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	key := "GetUsersForAdmin" + emailCreatorID + string(newPage)
	result, err := redis.GetRedisValue(key)
	if err == nil {
		err = json.Unmarshal([]byte(result), &users)
		if err != nil {
			log.Errorf("error while unmarshalling redis value: %v", err)
		}
	}
	var newCursor string
	if len(users) <= 0 {
		userAdmin := userz.User{
			ID: emailCreatorID,
		}
		session, err := cassandra.GetCassSession("userz")
		if err != nil {
			return nil, err
		}
		CassUserSession := session

		getQuery := CassUserSession.Query(userz.UserTable.Get()).BindMap(qb.M{"id": userAdmin.ID})
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
		if pageSize == nil {
			pageSizeInt = 10
		} else {
			pageSizeInt = *pageSize
		}

		qryStr := fmt.Sprintf(`SELECT * from userz.users where created_by='%s' and updated_at <= %d  ALLOW FILTERING`, email_creator, *publishTime)
		getUsers := func(page []byte) (users []userz.User, nextPage []byte, err error) {
			q := CassUserSession.Query(qryStr, nil)
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
	}
	if len(newPage) != 0 {
		newCursor, err = global.CryptSession.EncryptAsString(newPage, nil)
		if err != nil {
			return nil, fmt.Errorf("error encrypting cursor: %v", err)
		}
		log.Infof("Users: %v", string(newCursor))

	}
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
	redisBytes, err := json.Marshal(users)
	if err == nil {
		redis.SetTTL(key, 3600)
		redis.SetRedisValue(key, string(redisBytes))
	}
	return &outputResponse, nil
}

func GetUserDetails(ctx context.Context, userIds []*string) ([]*model.User, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var outputResponse []*model.User
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		log.Errorf("Failed to upload image to course: %v", err.Error())
		return nil, err
	}
	for _, userID := range userIds {
		userCopy := userz.User{}
		key := "GetUserDetails" + *userID
		result, err := redis.GetRedisValue(key)
		if err == nil {
			json.Unmarshal([]byte(result), &userCopy)

		}
		if userCopy.ID == "" {
			qryStr := fmt.Sprintf(`SELECT * from userz.users where id='%s' ALLOW FILTERING`, *userID)
			getUsers := func() (users []userz.User, err error) {
				q := CassUserSession.Query(qryStr, nil)
				defer q.Release()

				iter := q.Iter()
				return users, iter.Select(&users)
			}
			users, err := getUsers()
			if err != nil {
				return nil, err
			}
			if len(users) == 0 {
				return nil, fmt.Errorf("user not found")
			}

			userCopy = users[0]
		}
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
		outputResponse = append(outputResponse, outputUser)
		redisBytes, err := json.Marshal(userCopy)
		if err == nil {
			redis.SetTTL(key, 3600)
			redis.SetRedisValue(key, string(redisBytes))
		}
	}

	return outputResponse, nil
}
