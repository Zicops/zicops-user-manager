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

func GetLatestCohortDetails(ctx context.Context, lspID string, userID *string, publishTime *int, pageCursor *string, direction *string, pageSize *int) ([]*model.CohortMain, error) {
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

	if userID == nil {

		if strings.ToLower(userAdmin.Role) != "admin" {
			return nil, fmt.Errorf("user is not an admin")
		} else {
			//here userID is nill, hence we only have lspID, so we need to return cohortMain based on lspID
			qryStr := fmt.Sprintf(`SELECT * from userz.cohort_main where lsp_id='%s' ALLOW FILTERING`, lspID)

			getCohorts := func() (users []userz.Cohort, err error) {
				q := global.CassUserSession.Session.Query(qryStr, nil)
				defer q.Release()
				iter := q.Iter()
				return users, iter.Select(&users)
			}

			userCohort, err := getCohorts()
			if err != nil {
				return nil, err
			}

			allUsers := make([]*model.CohortMain, 0)

			for _, copiedUser := range userCohort {
				cohortCopy := copiedUser
				createdAt := strconv.FormatInt(cohortCopy.CreatedAt, 10)
				updatedAt := strconv.FormatInt(cohortCopy.UpdatedAt, 10)

				userCohort := &model.CohortMain{
					CohortID:    &cohortCopy.ID,
					Name:        cohortCopy.Name,
					Description: cohortCopy.Description,
					LspID:       cohortCopy.LspID,
					Code:        cohortCopy.Code,
					Status:      cohortCopy.Status,
					Type:        cohortCopy.Type,
					IsActive:    cohortCopy.IsActive,
					CreatedBy:   &cohortCopy.CreatedBy,
					UpdatedBy:   &cohortCopy.UpdatedBy,
					CreatedAt:   createdAt,
					UpdatedAt:   updatedAt,
					Size:        cohortCopy.Size,
					ImageURL:    &cohortCopy.ImageUrl,
				}
				allUsers = append(allUsers, userCohort)
			}

			return allUsers, nil
		}
	} else {
		//here we have both user_id and lsp_id
		//pass user_id to get cohort_map
		qryStr := fmt.Sprintf(`SELECT cohort_id  from userz.user_cohort_map where user_id = '%s ALLOW FILTERING`, *userID)

		getCohortId := func() (cohortId []string, err error) {
			q := global.CassUserSession.Session.Query(qryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return cohortId, iter.Select(&cohortId)
		}

		cohortIds, err := getCohortId()

		if err != nil {
			return nil, err
		}
		allUsers := make([]*model.CohortMain, 0)

		//now we have an array of cohortIds, just pass it to getCohortDetails, and store that in another array and pass it
		for _, cohortId := range cohortIds {
			userCohort, err := GetCohortDetails(ctx, cohortId)
			if err != nil {
				return nil, err
			}

			allUsers = append(allUsers, userCohort)
		}
		return allUsers, nil
	}
}
