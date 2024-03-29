package queries

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/lib/db/bucket"
	"github.com/zicops/zicops-user-manager/lib/googleprojectlib"
	"github.com/zicops/zicops-user-manager/lib/identity"
)

func GetLatestCohorts(ctx context.Context, userID *string, userLspID *string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedCohorts, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	if userID != nil {
		emailCreatorID = *userID
	}
	var outputResponse model.PaginatedCohorts

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
	//laspID := ""
	//if userLspID != nil {
	//	laspID = *userLspID
	//}
	//key := "GetLatestCohorts" + emailCreatorID + laspID + string(newPage)
	//result, err := redis.GetRedisValue(key)
	//if err == nil {
	//	err = json.Unmarshal([]byte(result), &outputResponse)
	//	if err == nil {
	//		return &outputResponse, nil
	//	}
	//}

	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	if pageSize == nil {
		pageSizeInt = 10
	} else {
		pageSizeInt = *pageSize
	}
	var newCursor string
	lspClause := ""
	if userLspID != nil {
		lspClause = fmt.Sprintf(" and user_lsp_id='%s'", *userLspID)
	}
	qryStr := fmt.Sprintf(`SELECT * from userz.user_cohort_map where user_id='%s' and created_at <= %d %s ALLOW FILTERING`, emailCreatorID, *publishTime, lspClause)
	getUsers := func(page []byte) (users []userz.UserCohort, nextPage []byte, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		q.PageState(page)
		q.PageSize(pageSizeInt)

		iter := q.Iter()
		return users, iter.PageState(), iter.Select(&users)
	}
	usersCohort, newPage, err := getUsers(newPage)
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
	allUsers := make([]*model.UserCohort, len(usersCohort))
	if len(usersCohort) <= 0 {
		outputResponse.Cohorts = allUsers
		return &outputResponse, nil
	}
	var wg sync.WaitGroup
	for i, cu := range usersCohort {
		cc := cu
		wg.Add(1)
		go func(i int, cohortCopy userz.UserCohort) {
			createdAt := strconv.FormatInt(cohortCopy.CreatedAt, 10)
			updatedAt := strconv.FormatInt(cohortCopy.UpdatedAt, 10)
			userCohort := &model.UserCohort{
				UserID:           cohortCopy.UserID,
				UserLspID:        cohortCopy.UserLspID,
				UserCohortID:     &cohortCopy.ID,
				CohortID:         cohortCopy.CohortID,
				CreatedAt:        createdAt,
				UpdatedAt:        updatedAt,
				CreatedBy:        &cohortCopy.CreatedBy,
				UpdatedBy:        &cohortCopy.UpdatedBy,
				AddedBy:          cohortCopy.AddedBy,
				MembershipStatus: cohortCopy.MembershipStatus,
				Role:             cohortCopy.Role,
			}
			allUsers[i] = userCohort
			wg.Done()
		}(i, cc)
	}
	wg.Wait()
	outputResponse.Cohorts = allUsers
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction
	//redisBytes, err := json.Marshal(outputResponse)
	//if err == nil {
	//	redis.SetTTL(key, 3600)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
	return &outputResponse, nil
}

func GetCohortUsers(ctx context.Context, cohortID string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedCohorts, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
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
	//key := "GetCohortUsers" + cohortID + string(newPage)
	//result, err := redis.GetRedisValue(key)
	//if err == nil {
	//	var outputResponse model.PaginatedCohorts
	//	err = json.Unmarshal([]byte(result), &outputResponse)
	//	if err == nil {
	//		return &outputResponse, nil
	//	}
	//}

	if pageSize == nil {
		pageSizeInt = 10
	} else {
		pageSizeInt = *pageSize
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	var newCursor string
	qryStr := fmt.Sprintf(`SELECT * from userz.user_cohort_map where cohort_id='%s' and created_at<=%d ALLOW FILTERING`, cohortID, *publishTime)
	getUsersCohort := func(page []byte) (users []userz.UserCohort, nextPage []byte, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		q.PageState(page)
		q.PageSize(pageSizeInt)
		iter := q.Iter()
		return users, iter.PageState(), iter.Select(&users)
	}
	userCohorts, newPage, err := getUsersCohort(newPage)
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
	if len(userCohorts) == 0 {
		return nil, fmt.Errorf("no users found")
	}
	cohortUsers := make([]*model.UserCohort, len(userCohorts))
	var outputResponse model.PaginatedCohorts
	if len(userCohorts) <= 0 {
		outputResponse.Cohorts = cohortUsers
		return &outputResponse, nil
	}
	var wg sync.WaitGroup
	for i, uo := range userCohorts {
		cc := uo
		wg.Add(1)
		go func(i int, cohortCopy userz.UserCohort) {
			createdAt := strconv.FormatInt(cohortCopy.CreatedAt, 10)
			updatedAt := strconv.FormatInt(cohortCopy.UpdatedAt, 10)
			userCohort := &model.UserCohort{
				UserID:           cohortCopy.UserID,
				UserLspID:        cohortCopy.UserLspID,
				UserCohortID:     &cohortCopy.ID,
				CohortID:         cohortCopy.CohortID,
				CreatedAt:        createdAt,
				UpdatedAt:        updatedAt,
				CreatedBy:        &cohortCopy.CreatedBy,
				UpdatedBy:        &cohortCopy.UpdatedBy,
				AddedBy:          cohortCopy.AddedBy,
				MembershipStatus: cohortCopy.MembershipStatus,
				Role:             cohortCopy.Role,
			}
			cohortUsers[i] = userCohort
			wg.Done()
		}(i, cc)
	}
	wg.Wait()
	outputResponse.Cohorts = cohortUsers
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction
	//redisBytes, err := json.Marshal(outputResponse)
	//if err == nil {
	//	redis.SetTTL(key, 3600)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
	return &outputResponse, nil
}

func AddCohortMain(ctx context.Context, input model.CohortMainInput) (*model.CohortMain, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var photoBucket string
	var photoUrl string

	cohortID := uuid.New().String()
	email_creator := claims["email"].(string)
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		return nil, err
	}
	if input.Image != nil && input.ImageURL == nil {
		extension := strings.Split(input.Image.Filename, ".")
		bucketPath := fmt.Sprintf("%s/%s/%s", "cohorts", cohortID, base64.URLEncoding.EncodeToString([]byte(input.Image.Filename)))
		if len(extension) >= 1 {
			bucketPath += "." + extension[len(extension)-1]
		}
		writer, err := storageC.UploadToGCS(ctx, bucketPath)
		if err != nil {
			return nil, err
		}
		defer writer.Close()
		fileBuffer := bytes.NewBuffer(nil)
		if _, err := io.Copy(fileBuffer, input.Image.File); err != nil {
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
		if input.ImageURL != nil {
			photoUrl = *input.ImageURL
		}
	}
	words := []string{}
	if input.Name != "" {
		name := strings.ToLower(input.Name)
		for _, word := range name {
			wordsLower := strings.ToLower(string(word))
			words = append(words, wordsLower)
		}
	}
	cohortMainTable := userz.Cohort{
		ID:          cohortID,
		Name:        input.Name,
		Words:       words,
		Description: input.Description,
		ImageBucket: photoBucket,
		ImageUrl:    photoUrl,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		CreatedBy:   email_creator,
		UpdatedBy:   email_creator,
		Code:        input.Code,
		Type:        input.Type,
		IsActive:    input.IsActive,
		Status:      input.Status,
		LspId:       input.LspID,
		Size:        input.Size,
	}
	insertQuery := CassUserSession.Query(userz.CohortTable.Insert()).BindStruct(cohortMainTable)
	if err := insertQuery.ExecRelease(); err != nil {
		return nil, err
	}
	created := strconv.FormatInt(cohortMainTable.CreatedAt, 10)
	updated := strconv.FormatInt(cohortMainTable.UpdatedAt, 10)
	outputCohort := &model.CohortMain{
		CohortID:    &cohortID,
		Name:        cohortMainTable.Name,
		Description: cohortMainTable.Description,
		ImageURL:    &cohortMainTable.ImageUrl,
		CreatedAt:   created,
		UpdatedAt:   updated,
		CreatedBy:   &cohortMainTable.CreatedBy,
		UpdatedBy:   &cohortMainTable.UpdatedBy,
		Code:        cohortMainTable.Code,
		Type:        cohortMainTable.Type,
		IsActive:    cohortMainTable.IsActive,
		Status:      cohortMainTable.Status,
		LspID:       cohortMainTable.LspId,
		Size:        cohortMainTable.Size,
	}

	return outputCohort, nil
}

func UpdateCohortMain(ctx context.Context, input model.CohortMainInput) (*model.CohortMain, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	lspID := claims["lsp_id"].(string)
	var photoBucket string
	var photoUrl string

	cohortID := uuid.New().String()
	if input.CohortID == nil {
		return nil, fmt.Errorf("cohort id is required")
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	currentCohort := userz.Cohort{
		ID: *input.CohortID,
	}
	cohorts := []userz.Cohort{}
	getQueryStr := fmt.Sprintf("SELECT * FROM userz.cohort_main WHERE id='%s' AND lsp_id='%s'  ", currentCohort.ID, lspID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
	if err := getQuery.SelectRelease(&cohorts); err != nil {
		return nil, err
	}
	if len(cohorts) == 0 {
		return nil, fmt.Errorf("cohorts not found")
	}
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		return nil, err
	}
	if input.Image != nil {
		extension := strings.Split(input.Image.Filename, ".")
		bucketPath := fmt.Sprintf("%s/%s/%s", "cohorts", cohortID, base64.URLEncoding.EncodeToString([]byte(input.Image.Filename)))
		if len(extension) >= 1 {
			bucketPath += "." + extension[len(extension)-1]
		}
		writer, err := storageC.UploadToGCS(ctx, bucketPath)
		if err != nil {
			return nil, err
		}
		defer writer.Close()
		fileBuffer := bytes.NewBuffer(nil)
		if _, err := io.Copy(fileBuffer, input.Image.File); err != nil {
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
		if input.ImageURL != nil {
			photoUrl = *input.ImageURL
		}
	}
	cohort := cohorts[0]
	updatedCols := []string{}
	if input.Name != "" {
		cohort.Name = input.Name
		updatedCols = append(updatedCols, "name")
		words := []string{}
		name := input.Name
		for _, word := range name {
			wordsLower := strings.ToLower(string(word))
			words = append(words, wordsLower)
		}
		cohort.Words = words
		updatedCols = append(updatedCols, "words")
	}
	if input.Description != "" {
		cohort.Description = input.Description
		updatedCols = append(updatedCols, "description")
	}
	if photoUrl != "" {
		cohort.ImageUrl = photoUrl
		updatedCols = append(updatedCols, "imageurl")
	}
	if input.Code != "" {
		cohort.Code = input.Code
		updatedCols = append(updatedCols, "code")
	}
	if input.Type != "" {
		cohort.Type = input.Type
		updatedCols = append(updatedCols, "type")
	}
	if photoBucket != "" {
		cohort.ImageBucket = photoBucket
		updatedCols = append(updatedCols, "imagebucket")
	}
	if input.Size > 0 {
		cohort.Size = input.Size
		updatedCols = append(updatedCols, "size")
	}
	if input.Status != "" {
		cohort.Status = input.Status
		updatedCols = append(updatedCols, "status")
	}
	if input.UpdatedBy != nil {
		cohort.UpdatedBy = *input.UpdatedBy
		updatedCols = append(updatedCols, "updated_by")
	}
	cohort.UpdatedAt = time.Now().Unix()
	updatedCols = append(updatedCols, "updated_at")
	upStms, uNames := userz.CohortTable.Update(updatedCols...)
	updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&cohort)
	if err := updateQuery.ExecRelease(); err != nil {
		log.Errorf("error updating cohort: %v", err)
		return nil, err
	}
	created := strconv.FormatInt(cohort.CreatedAt, 10)
	updated := strconv.FormatInt(cohort.UpdatedAt, 10)
	outputCohort := &model.CohortMain{
		CohortID:    &cohortID,
		Name:        cohort.Name,
		Description: cohort.Description,
		ImageURL:    &cohort.ImageUrl,
		CreatedAt:   created,
		UpdatedAt:   updated,
		CreatedBy:   &cohort.CreatedBy,
		UpdatedBy:   &cohort.UpdatedBy,
		Code:        cohort.Code,
		Type:        cohort.Type,
		IsActive:    cohort.IsActive,
		Status:      cohort.Status,
		LspID:       cohort.LspId,
		Size:        cohort.Size,
	}

	return outputCohort, nil
}

func GetCohortDetails(ctx context.Context, cohortID string) (*model.CohortMain, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	lspID := claims["lsp_id"].(string)
	//key := "GetCohortDetails" + cohortID
	//result, err := redis.GetRedisValue(key)
	cohort := userz.Cohort{}
	//if err == nil {
	//	json.Unmarshal([]byte(result), &cohort)
	//}
	var photoBucket string
	var photoUrl string
	if cohort.ID == "" {
		session, err := global.CassPool.GetSession(ctx, "userz")
		if err != nil {
			return nil, err
		}
		CassUserSession := session
		cohorts := []userz.Cohort{}

		getCohortQueryStr := fmt.Sprintf("SELECT * FROM userz.cohort_main WHERE id = '%s' AND lsp_id = '%s' ", cohortID, lspID)
		getQuery := CassUserSession.Query(getCohortQueryStr, nil)
		if err := getQuery.SelectRelease(&cohorts); err != nil {
			return nil, err
		}
		if len(cohorts) == 0 {
			return nil, fmt.Errorf("cohorts not found")
		}
		cohort = cohorts[0]
	}
	photoBucket = cohort.ImageBucket
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		return nil, err
	}
	if photoBucket != "" {
		photoUrl = storageC.GetSignedURLForObject(ctx, photoBucket)
	}
	created := strconv.FormatInt(cohort.CreatedAt, 10)
	updated := strconv.FormatInt(cohort.UpdatedAt, 10)
	outputCohort := &model.CohortMain{
		CohortID:    &cohortID,
		Name:        cohort.Name,
		Description: cohort.Description,
		ImageURL:    &photoUrl,
		CreatedAt:   created,
		UpdatedAt:   updated,
		CreatedBy:   &cohort.CreatedBy,
		UpdatedBy:   &cohort.UpdatedBy,
		Code:        cohort.Code,
		Type:        cohort.Type,
		IsActive:    cohort.IsActive,
		Status:      cohort.Status,
		LspID:       cohort.LspId,
		Size:        cohort.Size,
	}
	//redisBytes, err := json.Marshal(cohort)
	//if err == nil {
	//	redis.SetTTL(key, 3600)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
	return outputCohort, nil
}

func GetCohorts(ctx context.Context, cohortIds []*string) ([]*model.CohortMain, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	lspId := claims["lsp_id"].(string)

	res := make([]*model.CohortMain, len(cohortIds))
	var wg sync.WaitGroup
	for kk, vvv := range cohortIds {
		if vvv == nil {
			continue
		}
		vv := *vvv
		wg.Add(1)
		go func(k int, v string, lsp string) {

			session, err := global.CassPool.GetSession(ctx, "userz")
			if err != nil {
				return
			}
			CassUserSession := session
			queryStr := fmt.Sprintf(`SELECT * FROM userz.cohort_main WHERE id = '%s' AND lsp_id = '%s' ALLOW FILTERING`, v, lsp)
			getCohort := func() (cohort []userz.Cohort, err error) {
				q := CassUserSession.Query(queryStr, nil)
				defer q.Release()
				iter := q.Iter()
				return cohort, iter.Select(&cohort)
			}

			cohorts, err := getCohort()
			if err != nil {
				log.Printf("Error while getting cohorts: %v", err)
				wg.Done()
				return
			}
			if len(cohorts) == 0 {
				log.Println("Cohort not found")
				wg.Done()
				return
			}

			cohort := cohorts[0]

			var photoUrl string

			photoBucket := cohort.ImageBucket
			storageC := bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err = storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				log.Errorf("Got error : %v", err)
				return
			}
			if photoBucket != "" {
				photoUrl = storageC.GetSignedURLForObject(ctx, photoBucket)
			}
			created := strconv.FormatInt(cohort.CreatedAt, 10)
			updated := strconv.FormatInt(cohort.UpdatedAt, 10)
			outputCohort := &model.CohortMain{
				CohortID:    &v,
				Name:        cohort.Name,
				Description: cohort.Description,
				ImageURL:    &photoUrl,
				CreatedAt:   created,
				UpdatedAt:   updated,
				CreatedBy:   &cohort.CreatedBy,
				UpdatedBy:   &cohort.UpdatedBy,
				Code:        cohort.Code,
				Type:        cohort.Type,
				IsActive:    cohort.IsActive,
				Status:      cohort.Status,
				LspID:       cohort.LspId,
				Size:        cohort.Size,
			}
			res[k] = outputCohort

			wg.Done()
		}(kk, vv, lspId)
	}
	wg.Wait()

	return res, nil
}

func GetCohortMains(ctx context.Context, lspID string, publishTime *int, pageCursor *string, direction *string, pageSize *int, searchText *string) (*model.PaginatedCohortsMain, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
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
	//key := "GetCohortMains" + lspID + string(newPage)
	cohorts := make([]userz.Cohort, 0)
	//result, err := redis.GetRedisValue(key)
	//if err == nil {
	//	json.Unmarshal([]byte(result), &cohorts)
	//}
	var newCursor string

	if len(cohorts) <= 0 {
		session, err := global.CassPool.GetSession(ctx, "userz")
		if err != nil {
			return nil, err
		}
		CassUserSession := session
		// if strings.ToLower(userAdmin.Role) != "admin" {
		// 	return nil, fmt.Errorf("user is not an admin")
		// }

		if pageSize == nil {
			pageSizeInt = 10
		} else {
			pageSizeInt = *pageSize
		}
		whereClause := ""
		if searchText != nil && *searchText != "" {
			whereClause = fmt.Sprintf(" AND name CONTAINS '%s'", *searchText)
		}
		qryStr := fmt.Sprintf(`SELECT * from userz.cohort_main where lsp_id='%s' and created_at<=%d %s ALLOW FILTERING`, lspID, *publishTime, whereClause)
		getCohorts := func(page []byte) (users []userz.Cohort, nextPage []byte, err error) {
			q := CassUserSession.Query(qryStr, nil)
			defer q.Release()
			q.PageState(page)
			q.PageSize(pageSizeInt)
			iter := q.Iter()
			return users, iter.PageState(), iter.Select(&users)
		}
		cohorts, newPage, err = getCohorts(newPage)
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
	if len(cohorts) == 0 {
		return nil, fmt.Errorf("no cohorts found")
	}
	cohortUsers := make([]*model.CohortMain, len(cohorts))
	var outputResponse model.PaginatedCohortsMain
	if len(cohorts) <= 0 {
		outputResponse.Cohorts = cohortUsers
		return &outputResponse, nil
	}
	var wg sync.WaitGroup
	for i, uo := range cohorts {
		cc := uo
		wg.Add(1)
		go func(i int, cohortCopy userz.Cohort) {
			createdAt := strconv.FormatInt(cohortCopy.CreatedAt, 10)
			updatedAt := strconv.FormatInt(cohortCopy.UpdatedAt, 10)
			imageBucket := cohortCopy.ImageBucket
			var photoUrl string
			if imageBucket != "" {
				storageC := bucket.NewStorageHandler()
				gproject := googleprojectlib.GetGoogleProjectID()
				err := storageC.InitializeStorageClient(ctx, gproject)
				if err != nil {
					log.Errorf("error initializing storage client: %v", err)
				}
				photoUrl = storageC.GetSignedURLForObject(ctx, imageBucket)
			}
			userCohort := &model.CohortMain{
				LspID:       cohortCopy.LspId,
				CohortID:    &cohortCopy.ID,
				Name:        cohortCopy.Name,
				Description: cohortCopy.Description,
				ImageURL:    &photoUrl,
				Code:        cohortCopy.Code,
				Type:        cohortCopy.Type,
				IsActive:    cohortCopy.IsActive,
				Status:      cohortCopy.Status,
				Size:        cohortCopy.Size,
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
				CreatedBy:   &cohortCopy.CreatedBy,
				UpdatedBy:   &cohortCopy.UpdatedBy,
			}
			cohortUsers[i] = userCohort
			wg.Done()
		}(i, cc)
	}
	wg.Wait()
	outputResponse.Cohorts = cohortUsers
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction
	//redisBytes, err := json.Marshal(cohorts)
	//if err == nil {
	//	redis.SetTTL(key, 3600)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
	return &outputResponse, nil
}

func DeleteCohortImage(ctx context.Context, cohortID string, filename string) (*string, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting the claims: %v", err)
		return nil, err
	}
	lspId := claims["lsp_id"].(string)

	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	cohorts := []userz.Cohort{}
	queryStr := fmt.Sprintf(`SELECT * FROM userz.cohort_main WHERE id='%s' AND lsp_id='%s' `, cohortID, lspId)
	getQuery := CassUserSession.Query(queryStr, nil)
	if err := getQuery.SelectRelease(&cohorts); err != nil {
		return nil, err
	}
	if len(cohorts) == 0 {
		return nil, fmt.Errorf("cohorts not found")
	}
	cohort := cohorts[0]
	cohort.ImageUrl = ""

	upStms, uNames := userz.CohortTable.Update("imageurl")
	updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&cohort)
	if err := updateQuery.ExecRelease(); err != nil {
		log.Errorf("error updating cohort: %v", err)
		return nil, err
	}

	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		return nil, err
	}
	bucketPath := fmt.Sprintf("%s/%s/%s", "cohorts", cohortID, filename)
	//bucketPath := fmt.Sprintf("%s/%s", "cohorts", filename)
	res := storageC.DeleteObjectsFromBucket(ctx, bucketPath)

	if res != "success" {
		return nil, fmt.Errorf(res)
	}
	return &res, nil
}
