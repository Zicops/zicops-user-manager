package queries

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

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

func GetUsersForAdmin(ctx context.Context, publishTime *int, pageCursor *string, direction *string, pageSize *int, filters *model.UserFilters) (*model.PaginatedUsers, error) {
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
	//key := "GetUsersForAdmin" + emailCreatorID + string(newPage)
	// result, err := redis.GetRedisValue(key)
	// if err == nil {
	// 	err = json.Unmarshal([]byte(result), &users)
	// 	if err != nil {
	// 		log.Errorf("error while unmarshalling redis value: %v", err)
	// 	}
	// }
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

		getUserQString := fmt.Sprintf("SELECT * FROM userz.users WHERE id = '%s' ", userAdmin.ID)
		getQuery := CassUserSession.Query(getUserQString, nil)
		if err := getQuery.SelectRelease(&users); err != nil {
			return nil, err
		}
		if len(users) == 0 {
			return nil, fmt.Errorf("user not found")
		}
		userAdmin = users[0]
		//TODO revisit this
		// if strings.ToLower(userAdmin.Role) != "admin" {
		// 	return nil, fmt.Errorf("user is not an admin")
		// }
		if pageSize == nil {
			pageSizeInt = 10
		} else {
			pageSizeInt = *pageSize
		}
		whereClause := fmt.Sprintf("WHERE created_by = '%s' AND created_at <= %d ", emailCreatorID, *publishTime)
		if filters != nil {
			if filters.Email != nil {
				whereClause += fmt.Sprintf("AND email = '%s' ", *filters.Email)
			}
			if filters.NameSearch != nil {
				whereClause += fmt.Sprintf("AND first_name = '%s' ", *filters.NameSearch)
			}
			if filters.Role != nil {
				whereClause += fmt.Sprintf("AND role = '%s' ", *filters.Role)
			}
			if filters.Status != nil {
				whereClause += fmt.Sprintf("AND status = '%s' ", *filters.Status)
			}
		}
		qryStr := fmt.Sprintf(`SELECT * from userz.users %s ALLOW FILTERING`, whereClause)
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
	allUsers := make([]*model.User, len(users))
	if len(users) <= 0 {
		return &outputResponse, nil
	}
	var wg sync.WaitGroup
	for i, cu := range users {
		u := cu
		wg.Add(1)
		go func(i int, userCopy userz.User) {
			createdAt := strconv.FormatInt(userCopy.CreatedAt, 10)
			updatedAt := strconv.FormatInt(userCopy.UpdatedAt, 10)
			photoUrl := ""
			storageC := bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err = storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				log.Errorf("Failed to upload image to course: %v", err.Error())
			}
			if userCopy.PhotoBucket != "" {
				photoUrl = storageC.GetSignedURLForObject(ctx, userCopy.PhotoBucket)
			} else {
				photoUrl = userCopy.PhotoURL
			}
			fireBaseUser, err := global.IDP.GetUserByEmail(ctx, userCopy.Email)
			if err != nil {
				log.Errorf("Failed to get user from firebase: %v", err.Error())
			}
			phone := ""
			if fireBaseUser != nil {
				phone = fireBaseUser.PhoneNumber
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
				Phone:      phone,
			}
			allUsers[i] = currentUser
			wg.Done()
		}(i, u)
	}
	wg.Wait()
	outputResponse.Users = allUsers
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction
	// redisBytes, err := json.Marshal(users)
	// if err == nil {
	// 	redis.SetTTL(key, 3600)
	// 	redis.SetRedisValue(key, string(redisBytes))
	// }
	return &outputResponse, nil
}

func GetUserDetails(ctx context.Context, userIds []*string) ([]*model.User, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	outputResponse := make([]*model.User, len(userIds))
	var wg sync.WaitGroup
	for ii, id := range userIds {
		copiedID := *id
		wg.Add(1)
		go func(i int, copyID string) {
			defer wg.Done()
			userCopy := userz.User{}
			userModel := model.User{}
			key := copyID
			result, err := redis.GetRedisValue(ctx, key)
			if err == nil && result != "" {
				err = json.Unmarshal([]byte(result), &userModel)
				if err == nil && userModel.ID != nil && *userModel.ID == copyID {
					outputResponse[i] = &userModel
					return
				}
			}
			storageC := bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err = storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				log.Errorf("Failed to upload image to course: %v", err.Error())
				return
			}
			if userCopy.ID == "" {
				qryStr := fmt.Sprintf(`SELECT * from userz.users where id='%s' ALLOW FILTERING`, copyID)
				getUsers := func() (users []userz.User, err error) {
					q := CassUserSession.Query(qryStr, nil)
					defer q.Release()
					iter := q.Iter()
					return users, iter.Select(&users)
				}
				users, err := getUsers()
				if err != nil {
					log.Errorf("Failed to get user from cassandra: %v", err.Error())
					return
				}
				if len(users) == 0 {
					log.Errorf("Failed to get user from cassandra: not found")
					return
				}
				userCopy = users[0]
			}
			if userCopy.ID == "" {
				return
			}
			createdAt := strconv.FormatInt(userCopy.CreatedAt, 10)
			updatedAt := strconv.FormatInt(userCopy.UpdatedAt, 10)
			photoUrl := ""
			if userCopy.PhotoBucket != "" {
				photoUrl = storageC.GetSignedURLForObject(ctx, userCopy.PhotoBucket)
			} else {
				photoUrl = userCopy.PhotoURL
			}
			fireBaseUser, err := global.IDP.GetUserByEmail(ctx, userCopy.Email)
			if err != nil {
				log.Errorf("Failed to get user from firebase: %v", err.Error())
				return
			}
			phone := ""
			if fireBaseUser != nil {
				phone = fireBaseUser.PhoneNumber
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
				Phone:      phone,
			}
			redisBytes, err := json.Marshal(userCopy)
			if err == nil {
				redis.SetRedisValue(ctx, key, string(redisBytes))
				redis.SetTTL(ctx, key, 3600)
			}
			outputResponse[i] = outputUser
		}(ii, copiedID)
	}
	wg.Wait()
	// get clean user details and remove nulls
	newResponse := make([]*model.User, 0)
	for i, user := range outputResponse {
		if user == nil || user.ID == nil || *user.ID == "" {
			continue
		}
		newResponse = append(newResponse, outputResponse[i])
	}
	return newResponse, nil
}
