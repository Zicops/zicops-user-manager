package orgs

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/helpers"
	"github.com/zicops/zicops-user-manager/lib/db/bucket"
	"github.com/zicops/zicops-user-manager/lib/googleprojectlib"
)

func AddLearningSpace(ctx context.Context, input model.LearningSpaceInput) (*model.LearningSpace, error) {
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
	role := roleValue.(string)
	uniqueOrgId := input.OrgID + input.OuID
	lspID := base64.URLEncoding.EncodeToString([]byte(uniqueOrgId))
	owners := []string{}
	if input.Owners != nil {
		for _, owner := range input.Owners {
			owners = append(owners, *owner)
		}
	} else {
		owners = append(owners, role)
	}
	logoUrl := ""
	logoBucket := ""
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		return nil, err
	}
	if input.Logo != nil && input.Logo == nil {

		bucketPath := fmt.Sprintf("%s/%s/%s", lspID, "logos", input.Logo.Filename)
		writer, err := storageC.UploadToGCS(ctx, bucketPath)
		if err != nil {
			return nil, err
		}
		defer writer.Close()
		fileBuffer := bytes.NewBuffer(nil)
		if _, err := io.Copy(fileBuffer, input.Logo.File); err != nil {
			return nil, err
		}
		currentBytes := fileBuffer.Bytes()
		_, err = io.Copy(writer, bytes.NewReader(currentBytes))
		if err != nil {
			return nil, err
		}
		logoBucket = bucketPath
		logoUrl = storageC.GetSignedURLForObject(bucketPath)
	} else {
		logoBucket = ""
		if input.LogoURL != nil {
			logoUrl = *input.LogoURL
		}
	}
	photoUrl := ""
	photoBucket := ""
	if input.Profile != nil && input.Profile == nil {
		bucketPath := fmt.Sprintf("%s/%s/%s", lspID, "photos", input.Profile.Filename)
		writer, err := storageC.UploadToGCS(ctx, bucketPath)
		if err != nil {
			return nil, err
		}
		defer writer.Close()
		fileBuffer := bytes.NewBuffer(nil)
		if _, err := io.Copy(fileBuffer, input.Profile.File); err != nil {
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
		if input.ProfileURL != nil {
			photoUrl = *input.ProfileURL
		}
	}

	lspYay := userz.Lsp{
		ID:                   lspID,
		OrgID:                input.OrgID,
		OrgUnitID:            input.OuID,
		Name:                 input.Name,
		NoOfUsers:            int64(input.NoUsers),
		Owners:               owners,
		IsDefault:            input.IsDefault,
		Status:               input.Status,
		CreatedAt:            time.Now().Unix(),
		UpdatedAt:            time.Now().Unix(),
		CreatedBy:            role,
		UpdatedBy:            role,
		LogoBucket:           logoBucket,
		LogoURL:              logoUrl,
		ProfilePictureBucket: photoBucket,
		ProfilePictureURL:    photoUrl,
	}
	insertQuery := CassUserSession.Query(userz.LspTable.Insert()).BindStruct(lspYay)
	if err := insertQuery.ExecRelease(); err != nil {
		return nil, err
	}
	created := strconv.FormatInt(lspYay.CreatedAt, 10)
	updated := strconv.FormatInt(lspYay.UpdatedAt, 10)
	org := &model.LearningSpace{
		LspID:      &lspYay.ID,
		OrgID:      lspYay.OrgID,
		OuID:       lspYay.OrgUnitID,
		Name:       lspYay.Name,
		NoUsers:    input.NoUsers,
		Owners:     input.Owners,
		IsDefault:  lspYay.IsDefault,
		Status:     lspYay.Status,
		CreatedAt:  created,
		UpdatedAt:  updated,
		CreatedBy:  &lspYay.CreatedBy,
		UpdatedBy:  &lspYay.UpdatedBy,
		LogoURL:    &lspYay.LogoURL,
		ProfileURL: &photoUrl,
	}
	return org, nil
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
func UpdateLearningSpace(ctx context.Context, input model.LearningSpaceInput) (*model.LearningSpace, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if input.LspID == nil {
		return nil, fmt.Errorf("lsp id is required")
	}
	if input.OrgID == "" {
		return nil, fmt.Errorf("org id is required")
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	email := claims["email"]
	role := email.(string)
	orgCass := userz.Lsp{
		ID:    *input.LspID,
		OrgID: input.OrgID,
	}
	org := []userz.Lsp{}

	getQueryStr := fmt.Sprintf(`SELECT * from userz.learning_space where id='%s' and org_id='%s' `, orgCass.ID, orgCass.OrgID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
	if err := getQuery.SelectRelease(&org); err != nil {
		return nil, err
	}
	orgCass = org[0]
	updatedCols := []string{}
	if int64(input.NoUsers) != orgCass.NoOfUsers {
		orgCass.NoOfUsers = int64(input.NoUsers)
		updatedCols = append(updatedCols, "no_of_users")
	}

	if input.Status != "" && input.Status != orgCass.Status {
		orgCass.Status = input.Status
		updatedCols = append(updatedCols, "status")
	}
	if input.Name != "" && input.Name != orgCass.Name {
		orgCass.Name = input.Name
		updatedCols = append(updatedCols, "name")
	}
	owners := orgCass.Owners
	lenOwners := len(owners)
	if input.Owners != nil && len(input.Owners) > 0 {
		for _, owner := range input.Owners {
			// check if owner is already present
			if Contains(owners, *owner) {
				continue
			}
			owners = append(owners, *owner)
		}
		if lenOwners != len(owners) {
			orgCass.Owners = owners
			updatedCols = append(updatedCols, "owners")
		}
	}
	if input.IsDefault != orgCass.IsDefault {
		orgCass.IsDefault = input.IsDefault
		updatedCols = append(updatedCols, "is_default")
	}
	if input.OuID != "" && input.OuID != orgCass.OrgUnitID {
		orgCass.OrgUnitID = input.OuID
		updatedCols = append(updatedCols, "org_unit_id")
	}
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		return nil, err
	}
	if input.Logo != nil {
		bucketPath := fmt.Sprintf("%s/%s/%s", orgCass.ID, "logos", input.Profile.Filename)
		writer, err := storageC.UploadToGCS(ctx, bucketPath)
		if err != nil {
			return nil, err
		}
		defer writer.Close()
		fileBuffer := bytes.NewBuffer(nil)
		if _, err := io.Copy(fileBuffer, input.Logo.File); err != nil {
			return nil, err
		}
		currentBytes := fileBuffer.Bytes()
		_, err = io.Copy(writer, bytes.NewReader(currentBytes))
		if err != nil {
			return nil, err
		}
		url := storageC.GetSignedURLForObject(bucketPath)
		orgCass.LogoBucket = bucketPath
		orgCass.LogoURL = url
		updatedCols = append(updatedCols, "logo_bucket")
		updatedCols = append(updatedCols, "logo_url")
	}
	if input.Profile != nil {
		bucketPath := fmt.Sprintf("%s/%s/%s", orgCass.ID, "profile", input.Profile.Filename)
		writer, err := storageC.UploadToGCS(ctx, bucketPath)
		if err != nil {
			return nil, err
		}
		defer writer.Close()
		fileBuffer := bytes.NewBuffer(nil)
		if _, err := io.Copy(fileBuffer, input.Profile.File); err != nil {
			return nil, err
		}
		currentBytes := fileBuffer.Bytes()
		_, err = io.Copy(writer, bytes.NewReader(currentBytes))
		if err != nil {
			return nil, err
		}
		url := storageC.GetSignedURLForObject(bucketPath)
		orgCass.ProfilePictureBucket = bucketPath
		orgCass.ProfilePictureURL = url
		updatedCols = append(updatedCols, "profile_picture_bucket")
		updatedCols = append(updatedCols, "profile_picture_url")
	}

	if len(updatedCols) > 0 {
		orgCass.UpdatedAt = time.Now().Unix()
		orgCass.UpdatedBy = role
		updatedCols = append(updatedCols, "updated_at")
		upStms, uNames := userz.LspTable.Update(updatedCols...)
		updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&orgCass)
		if err := updateQuery.ExecRelease(); err != nil {
			log.Errorf("error updating user: %v", err)
			return nil, err
		}
	}
	lspYay := orgCass
	created := strconv.FormatInt(lspYay.CreatedAt, 10)
	updated := strconv.FormatInt(lspYay.UpdatedAt, 10)
	orgLsp := &model.LearningSpace{
		LspID:      &lspYay.ID,
		OrgID:      lspYay.OrgID,
		OuID:       lspYay.OrgUnitID,
		Name:       lspYay.Name,
		NoUsers:    input.NoUsers,
		Owners:     input.Owners,
		IsDefault:  lspYay.IsDefault,
		Status:     lspYay.Status,
		CreatedAt:  created,
		UpdatedAt:  updated,
		CreatedBy:  &lspYay.CreatedBy,
		UpdatedBy:  &lspYay.UpdatedBy,
		LogoURL:    &lspYay.LogoURL,
		ProfileURL: &lspYay.ProfilePictureURL,
	}
	return orgLsp, nil
}

func GetLearningSpaceDetails(ctx context.Context, lspIds []*string) ([]*model.LearningSpace, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	outputOrgs := make([]*model.LearningSpace, len(lspIds))
	var wg sync.WaitGroup
	for i, orgID := range lspIds {
		wg.Add(1)
		go func(i int, orgID *string) {
			if orgID == nil {
				return
			}
			qryStr := fmt.Sprintf(`SELECT * from userz.learning_space where id='%s' ALLOW FILTERING `, *orgID)
			getOrgs := func() (users []userz.Lsp, err error) {
				q := CassUserSession.Query(qryStr, nil)
				defer q.Release()
				iter := q.Iter()
				return users, iter.Select(&users)
			}
			orgs, err := getOrgs()
			if err != nil {
				log.Errorf("error getting orgs: %v", err)
				return 
			}
			if len(orgs) == 0 {
				return
			}
			orgCass := orgs[0]
			created := strconv.FormatInt(orgCass.CreatedAt, 10)
			updated := strconv.FormatInt(orgCass.UpdatedAt, 10)
			emptCnt := int(orgCass.NoOfUsers)
			owners := []*string{}
			for _, owner := range orgCass.Owners {
				owners = append(owners, &owner)
			}
			logoUrl := orgCass.LogoURL
			profileUrl := orgCass.ProfilePictureURL
			storageC := bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err = storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				return
			}
			if orgCass.LogoBucket != "" {
				logoUrl = storageC.GetSignedURLForObject(orgCass.LogoBucket)
			}
			if orgCass.ProfilePictureBucket != "" {
				profileUrl = storageC.GetSignedURLForObject(orgCass.ProfilePictureBucket)
			}

			result := &model.LearningSpace{
				LspID:      &orgCass.ID,
				OrgID:      orgCass.OrgID,
				OuID:       orgCass.OrgUnitID,
				Name:       orgCass.Name,
				NoUsers:    emptCnt,
				Owners:     owners,
				IsDefault:  orgCass.IsDefault,
				Status:     orgCass.Status,
				CreatedAt:  created,
				UpdatedAt:  updated,
				CreatedBy:  &orgCass.CreatedBy,
				UpdatedBy:  &orgCass.UpdatedBy,
				LogoURL:    &logoUrl,
				ProfileURL: &profileUrl,
			}
			outputOrgs[i] = result
			wg.Done()
		}(i, orgID)
	}
	wg.Wait()
	return outputOrgs, nil
}

func GetLearningSpacesByOrgID(ctx context.Context, orgID string) ([]*model.LearningSpace, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	outputOrgs := make([]*model.LearningSpace, len(orgID))
	lspIDs := []string{orgID}
	var wg sync.WaitGroup
	for i, orgID := range lspIDs {
		wg.Add(1)
		go func(i int, orgID string) {
			qryStr := fmt.Sprintf(`SELECT * from userz.learning_space where org_id='%s' ALLOW FILTERING `, orgID)
			getOrgs := func() (users []userz.Lsp, err error) {
				q := CassUserSession.Query(qryStr, nil)
				defer q.Release()
				iter := q.Iter()
				return users, iter.Select(&users)
			}
			orgs, err := getOrgs()
			if err != nil {
				log.Errorf("error getting orgs: %v", err)
				return
			}
			if len(orgs) == 0 {
				return
			}
			orgCass := orgs[0]
			created := strconv.FormatInt(orgCass.CreatedAt, 10)
			updated := strconv.FormatInt(orgCass.UpdatedAt, 10)
			emptCnt := int(orgCass.NoOfUsers)
			owners := []*string{}
			for _, owner := range orgCass.Owners {
				owners = append(owners, &owner)
			}
			logoUrl := orgCass.LogoURL
			profileUrl := orgCass.ProfilePictureURL
			storageC := bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err = storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				log.Println("Error in initializing storage client", err)
				return
			}
			if orgCass.LogoBucket != "" {

				logoUrl = storageC.GetSignedURLForObject(orgCass.LogoBucket)
			}
			if orgCass.ProfilePictureBucket != "" {
				profileUrl = storageC.GetSignedURLForObject(orgCass.ProfilePictureBucket)
			}

			result := &model.LearningSpace{
				LspID:      &orgCass.ID,
				OrgID:      orgCass.OrgID,
				OuID:       orgCass.OrgUnitID,
				Name:       orgCass.Name,
				NoUsers:    emptCnt,
				Owners:     owners,
				IsDefault:  orgCass.IsDefault,
				Status:     orgCass.Status,
				CreatedAt:  created,
				UpdatedAt:  updated,
				CreatedBy:  &orgCass.CreatedBy,
				UpdatedBy:  &orgCass.UpdatedBy,
				LogoURL:    &logoUrl,
				ProfileURL: &profileUrl,
			}
			outputOrgs[i] = result
			wg.Done()
		}(i, orgID)
	}
	wg.Wait()
	return outputOrgs, nil
}

func GetLearningSpacesByOuID(ctx context.Context, ouID string, orgID string) ([]*model.LearningSpace, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	outputOrgs := make([]*model.LearningSpace, len(ouID))
	lspIDs := []string{orgID}
	var wg sync.WaitGroup
	for i, orgID := range lspIDs {
		wg.Add(1)
		go func(i int, orgID string) {
			qryStr := fmt.Sprintf(`SELECT * from userz.learning_space where org_unit_id='%s' AND org_id='%s' ALLOW FILTERING `, ouID, orgID)
			getOrgs := func() (users []userz.Lsp, err error) {
				q := CassUserSession.Query(qryStr, nil)
				defer q.Release()
				iter := q.Iter()
				return users, iter.Select(&users)
			}
			orgs, err := getOrgs()
			if err != nil {
				log.Errorf("error getting orgs: %v", err)
				return
			}
			if len(orgs) == 0 {
				return
			}
			orgCass := orgs[0]
			created := strconv.FormatInt(orgCass.CreatedAt, 10)
			updated := strconv.FormatInt(orgCass.UpdatedAt, 10)
			emptCnt := int(orgCass.NoOfUsers)
			owners := []*string{}
			for _, owner := range orgCass.Owners {
				owners = append(owners, &owner)
			}
			logoUrl := orgCass.LogoURL
			profileUrl := orgCass.ProfilePictureURL
			storageC := bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err = storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				log.Println("Error in initializing storage client", err)
				return
			}
			if orgCass.LogoBucket != "" {
				logoUrl = storageC.GetSignedURLForObject(orgCass.LogoBucket)
			}
			if orgCass.ProfilePictureBucket != "" {
				profileUrl = storageC.GetSignedURLForObject(orgCass.ProfilePictureBucket)
			}

			result := &model.LearningSpace{
				LspID:      &orgCass.ID,
				OrgID:      orgCass.OrgID,
				OuID:       orgCass.OrgUnitID,
				Name:       orgCass.Name,
				NoUsers:    emptCnt,
				Owners:     owners,
				IsDefault:  orgCass.IsDefault,
				Status:     orgCass.Status,
				CreatedAt:  created,
				UpdatedAt:  updated,
				CreatedBy:  &orgCass.CreatedBy,
				UpdatedBy:  &orgCass.UpdatedBy,
				LogoURL:    &logoUrl,
				ProfileURL: &profileUrl,
			}
			outputOrgs[i] = result
			wg.Done()
		}(i, orgID)
	}
	wg.Wait()
	return outputOrgs, nil
}