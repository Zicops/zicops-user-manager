package common

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/lib/db/bucket"
	"github.com/zicops/zicops-user-manager/lib/googleprojectlib"
	"github.com/zicops/zicops-user-manager/lib/identity"
	log "github.com/sirupsen/logrus"
)

func GetUserDetails(ctx context.Context, userIds []*string) ([]*model.User, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	roleRaw := claims["role"].(string)
	role := strings.ToLower(roleRaw)
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	outputResponse := make([]*model.User, len(userIds))
	var wg sync.WaitGroup
	for ii, id := range userIds {
		if id == nil {
			continue
		}
		copiedID := *id
		wg.Add(1)
		go func(i int, copyID string) {
			defer wg.Done()
			userCopy := userz.User{}
			userModel := model.User{}
			key := copyID
			result, err := redis.GetRedisValue(ctx, key)
			if err == nil && result != "" && role != "admin" {
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
			redisBytes, err := json.Marshal(outputUser)
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
