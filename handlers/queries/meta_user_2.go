package queries

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/rs/xid"
	"github.com/scylladb/gocqlx/qb"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/helpers"
	"github.com/zicops/zicops-user-manager/lib/db/bucket"
	"github.com/zicops/zicops-user-manager/lib/googleprojectlib"
)

func GetLatestCohorts(ctx context.Context, userID *string, userLspID *string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedCohorts, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	if userID != nil {
		emailCreatorID = *userID
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
	lspClause := ""
	if userLspID != nil {
		lspClause = fmt.Sprintf(" and user_lsp_id='%s'", *userLspID)
	}
	qryStr := fmt.Sprintf(`SELECT * from userz.user_cohort_map where user_id='%s' and updated_at <= %d %s ALLOW FILTERING`, emailCreatorID, *publishTime, lspClause)
	getUsers := func(page []byte) (users []userz.UserCohort, nextPage []byte, err error) {
		q := global.CassUserSession.Session.Query(qryStr, nil)
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
	var outputResponse model.PaginatedCohorts
	allUsers := make([]*model.UserCohort, 0)
	for _, copiedUser := range usersCohort {
		cohortCopy := copiedUser
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
		allUsers = append(allUsers, userCohort)
	}
	outputResponse.Cohorts = allUsers
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction
	return &outputResponse, nil
}

func GetCohortUsers(ctx context.Context, cohortID string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedCohorts, error) {
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
	qryStr := fmt.Sprintf(`SELECT * from userz.user_cohort_map where cohort_id='%s' and updated_at<=%d ALLOW FILTERING`, cohortID, *publishTime)
	getUsersCohort := func(page []byte) (users []userz.UserCohort, nextPage []byte, err error) {
		q := global.CassUserSession.Session.Query(qryStr, nil)
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
	cohortUsers := make([]*model.UserCohort, 0)
	for _, userOrg := range userCohorts {
		cohortCopy := userOrg
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
		cohortUsers = append(cohortUsers, userCohort)
	}
	var outputResponse model.PaginatedCohorts
	outputResponse.Cohorts = cohortUsers
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction
	return &outputResponse, nil
}

func AddCohortMain(ctx context.Context, input model.CohortMainInput) (*model.CohortMain, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var storageC *bucket.Client
	var photoBucket string
	var photoUrl string
	guid := xid.New()
	cohortID := guid.String()
	email_creator := claims["email"].(string)
	if input.Image != nil && input.ImageURL == nil {
		if storageC == nil {
			storageC = bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err := storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				return nil, err
			}
		}
		bucketPath := fmt.Sprintf("%s/%s/%s", "cohorts", cohortID, input.Image.Filename)
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
		photoUrl = storageC.GetSignedURLForObject(bucketPath)
	} else {
		photoBucket = ""
		if input.ImageURL != nil {
			photoUrl = *input.ImageURL
		}
	}
	cohortMainTable := userz.Cohort{
		ID:          cohortID,
		Name:        input.Name,
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
		LspID:       input.LspID,
		Size:        input.Size,
	}
	insertQuery := global.CassUserSession.Session.Query(userz.CohortTable.Insert()).BindStruct(cohortMainTable)
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
		LspID:       cohortMainTable.LspID,
		Size:        cohortMainTable.Size,
	}

	return outputCohort, nil
}

func UpdateCohortMain(ctx context.Context, input model.CohortMainInput) (*model.CohortMain, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var storageC *bucket.Client
	var photoBucket string
	var photoUrl string
	guid := xid.New()
	cohortID := guid.String()
	if input.CohortID == nil {
		return nil, fmt.Errorf("cohort id is required")
	}
	currentCohort := userz.Cohort{
		ID: *input.CohortID,
	}
	cohorts := []userz.Cohort{}
	getQuery := global.CassUserSession.Session.Query(userz.CohortTable.Get()).BindMap(qb.M{"id": currentCohort.ID})
	if err := getQuery.SelectRelease(&cohorts); err != nil {
		return nil, err
	}
	if len(cohorts) == 0 {
		return nil, fmt.Errorf("cohorts not found")
	}

	if input.Image != nil && input.ImageURL == nil {
		if storageC == nil {
			storageC = bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err := storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				return nil, err
			}
		}
		bucketPath := fmt.Sprintf("%s/%s/%s", "cohorts", cohortID, input.Image.Filename)
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
		photoUrl = storageC.GetSignedURLForObject(bucketPath)
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
	}
	if input.Description != "" {
		cohort.Description = input.Description
		updatedCols = append(updatedCols, "description")
	}
	if photoUrl != "" {
		cohort.ImageUrl = *input.ImageURL
		updatedCols = append(updatedCols, "imageUrl")
	}
	if input.Code != "" {
		cohort.Code = input.Code
		updatedCols = append(updatedCols, "code")
	}
	if input.Type != "" {
		cohort.Type = input.Type
		updatedCols = append(updatedCols, "type")
	}
	if input.IsActive != cohort.IsActive {
		cohort.IsActive = input.IsActive
		updatedCols = append(updatedCols, "is_active")
	}
	if photoBucket != "" {
		cohort.ImageBucket = photoBucket
		updatedCols = append(updatedCols, "imageBucket")
	}
	if input.Size > 0 {
		cohort.Size = input.Size
		updatedCols = append(updatedCols, "size")
	}
	if input.Status != "" {
		cohort.Status = input.Status
		updatedCols = append(updatedCols, "status")
	}
	if input.LspID != "" {
		cohort.LspID = input.LspID
		updatedCols = append(updatedCols, "lsp_id")
	}
	if input.UpdatedBy != nil {
		cohort.UpdatedBy = *input.UpdatedBy
		updatedCols = append(updatedCols, "updated_by")
	}
	cohort.UpdatedAt = time.Now().Unix()
	updatedCols = append(updatedCols, "updated_at")
	upStms, uNames := userz.CohortTable.Update(updatedCols...)
	updateQuery := global.CassUserSession.Session.Query(upStms, uNames).BindStruct(&cohort)
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
		LspID:       cohort.LspID,
		Size:        cohort.Size,
	}

	return outputCohort, nil
}

func GetCohortDetails(ctx context.Context, cohortID string) (*model.CohortMain, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var storageC *bucket.Client
	var photoBucket string
	var photoUrl string
	currentCohort := userz.Cohort{
		ID: cohortID,
	}
	cohorts := []userz.Cohort{}
	getQuery := global.CassUserSession.Session.Query(userz.CohortTable.Get()).BindMap(qb.M{"id": currentCohort.ID})
	if err := getQuery.SelectRelease(&cohorts); err != nil {
		return nil, err
	}
	if len(cohorts) == 0 {
		return nil, fmt.Errorf("cohorts not found")
	}
	cohort := cohorts[0]
	photoBucket = cohort.ImageBucket
	if photoBucket != "" {
		if storageC == nil {
			storageC = bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err := storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				return nil, err
			}
		}
		photoUrl = storageC.GetSignedURLForObject(photoBucket)
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
		LspID:       cohort.LspID,
		Size:        cohort.Size,
	}

	return outputCohort, nil
}
