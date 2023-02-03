package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/contracts/vendorz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/helpers"
	"github.com/zicops/zicops-user-manager/lib/db/bucket"
	"github.com/zicops/zicops-user-manager/lib/googleprojectlib"
)

func AddVendor(ctx context.Context, input *model.VendorInput) (string, error) {
	//create vendor, map it to lsp id, thats all
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting claims: %v", err)
		return "", err
	}
	lspId := claims["lsp_id"].(string)
	if input.LspID != nil {
		lspId = *input.LspID
	}
	email := claims["email"].(string)
	id, _ := uuid.NewUUID()
	vendorId := id.String()
	createdAt := time.Now().Unix()

	session, err := cassandra.GetCassSession("vendorz")
	if err != nil {
		return "", err
	}
	CassUserSession := session

	//create vendor
	vendor := vendorz.Vendor{
		VendorId:  vendorId,
		Name:      *input.Name,
		Type:      *input.Type,
		CreatedAt: createdAt,
		CreatedBy: email,
	}
	if input.Address != nil {
		vendor.Address = *input.Address
	}
	if input.Website != nil {
		vendor.Website = *input.Website
	}
	if input.FacebookURL != nil {
		vendor.Facebook = *input.FacebookURL
	}
	if input.InstagramURL != nil {
		vendor.Instagram = *input.InstagramURL
	}
	if input.TwitterURL != nil {
		vendor.Twitter = *input.TwitterURL
	}
	if input.LinkedinURL != nil {
		vendor.LinkedIn = *input.LinkedinURL
	}
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		return "", err
	}
	if input.Photo != nil {
		bucketPath := fmt.Sprintf("%s/%s/%s", "vendor", vendor.Name, input.Photo.Filename)
		writer, err := storageC.UploadToGCS(ctx, bucketPath)
		if err != nil {
			return "", err
		}
		defer writer.Close()
		fileBuffer := bytes.NewBuffer(nil)
		if _, err := io.Copy(fileBuffer, input.Photo.File); err != nil {
			return "", err
		}
		currentBytes := fileBuffer.Bytes()
		_, err = io.Copy(writer, bytes.NewReader(currentBytes))
		if err != nil {
			return "", err
		}
		url := storageC.GetSignedURLForObject(bucketPath)
		vendor.PhotoBucket = bucketPath
		vendor.PhotoUrl = url
	}

	if input.Users != nil {
		//invite users
		users := changesStringType(input.Users)
		vendor.Users = users
	}
	if input.Status != nil {
		vendor.Status = *input.Status
	}

	insertQuery := CassUserSession.Query(vendorz.VendorTable.Insert()).BindStruct(vendor)
	if err = insertQuery.Exec(); err != nil {
		return "", err
	}

	//create its mapping to the specific LSP
	vendorLspMap := vendorz.VendorLspMap{
		VendorId:  vendorId,
		LspId:     lspId,
		CreatedAt: createdAt,
		CreatedBy: email,
	}
	insertQueryMap := CassUserSession.Query(vendorz.VendorLspMapTable.Insert()).BindStruct(vendorLspMap)
	if err = insertQueryMap.Exec(); err != nil {
		return "", err
	}

	return vendorId, nil
}

func UpdateVendor(ctx context.Context, input model.VendorInput) (*model.Vendor, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting claims from context: %v", err)
		return nil, nil
	}
	email := claims["email"].(string)
	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor WHERE id = %s`, *input.VendorID)
	session, err := cassandra.GetCassSession("vendorz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	var vendors []vendorz.Vendor
	getQuery := CassUserSession.Query(queryStr, nil)
	if err = getQuery.SelectRelease(&vendors); err != nil {
		return nil, err
	}
	if len(vendors) == 0 {
		return nil, nil
	}

	vendor := vendors[0]
	var updatedCols []string

	if input.Level != nil {
		updatedCols = append(updatedCols, "level")
		vendor.Level = *input.Level
	}
	if input.Address != nil {
		updatedCols = append(updatedCols, "address")
		vendor.Address = *input.Address
	}
	if input.Website != nil {
		updatedCols = append(updatedCols, "website")
		vendor.Website = *input.Website
	}
	if input.FacebookURL != nil {
		updatedCols = append(updatedCols, "facebook")
		vendor.Facebook = *input.FacebookURL
	}
	if input.InstagramURL != nil {
		updatedCols = append(updatedCols, "instagram")
		vendor.Instagram = *input.InstagramURL
	}
	if input.TwitterURL != nil {
		updatedCols = append(updatedCols, "twitter")
		vendor.Twitter = *input.TwitterURL
	}
	if input.LinkedinURL != nil {
		updatedCols = append(updatedCols, "linkedin")
		vendor.LinkedIn = *input.LinkedinURL
	}

	if input.Users != nil {
		//invite users
		updatedCols = append(updatedCols, "users")
		users := changesStringType(input.Users)
		vendor.Users = users
	}
	if input.Status != nil {
		updatedCols = append(updatedCols, "status")
		vendor.Status = *input.Status
	}

	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		return nil, err
	}
	if input.Photo != nil {
		bucketPath := fmt.Sprintf("%s/%s/%s", "vendor", vendor.Name, input.Photo.Filename)
		writer, err := storageC.UploadToGCS(ctx, bucketPath)
		if err != nil {
			return nil, err
		}
		defer writer.Close()
		fileBuffer := bytes.NewBuffer(nil)
		if _, err := io.Copy(fileBuffer, input.Photo.File); err != nil {
			return nil, err
		}
		currentBytes := fileBuffer.Bytes()
		_, err = io.Copy(writer, bytes.NewReader(currentBytes))
		if err != nil {
			return nil, err
		}
		url := storageC.GetSignedURLForObject(bucketPath)
		vendor.PhotoBucket = bucketPath
		vendor.PhotoUrl = url
		updatedCols = append(updatedCols, "photo_bucket")
		updatedCols = append(updatedCols, "photo_url")
	}
	if len(updatedCols) > 0 {
		updatedCols = append(updatedCols, "updated_by")
		vendor.UpdatedBy = email

		updatedCols = append(updatedCols, "updated_at")
		vendor.UpdatedAt = time.Now().Unix()
	}

	upStms, uNames := vendorz.VendorTable.Update(updatedCols...)
	updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(vendor)
	if err = updateQuery.ExecRelease(); err != nil {
		log.Printf("Error updating user: %v", err)
	}
	createdAt := strconv.Itoa(int(vendor.CreatedAt))
	updatedAt := strconv.Itoa(int(vendor.UpdatedAt))

	res := &model.Vendor{
		VendorID:     vendor.VendorId,
		Type:         vendor.Type,
		Level:        vendor.Level,
		Name:         vendor.Name,
		PhotoURL:     &vendor.PhotoUrl,
		Address:      &vendor.Address,
		Website:      &vendor.Website,
		FacebookURL:  &vendor.Facebook,
		InstagramURL: &vendor.Instagram,
		TwitterURL:   &vendor.Twitter,
		LinkedinURL:  &vendor.LinkedIn,
		CreatedAt:    &createdAt,
		CreatedBy:    &vendor.CreatedBy,
		UpdatedAt:    &updatedAt,
		UpdatedBy:    &email,
		Status:       vendor.Status,
	}

	return res, nil
}

func changesStringType(input []*string) []string {
	var res []string
	for _, vv := range input {
		v := vv
		res = append(res, *v)
	}
	return res
}

func CreateProfileVendor(ctx context.Context, input *model.VendorProfile) (string, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting claims: %v", err)
		return "", err
	}

	return "", nil
}

func CreateExperienceVendor(ctx context.Context, input model.ExperienceInput) (string, error) {

	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting claims: %v", err)
		return "", err
	}
	email_creator := claims["email"].(string)

	//employement type, location, location type
	description := fmt.Sprintf("%s %s %s", input.EmployementType, input.Location, input.LocationType)
	session, err := cassandra.GetCassSession("vendorz")
	if err != nil {
		return "", err
	}
	expIdUuid, _ := uuid.NewUUID()
	expId := expIdUuid.String()

	currentTime := time.Now().Unix()
	CassUserSession := session
	exp := vendorz.VendorExperience{
		ExpId:       expId,
		VendorId:    input.VendorID,
		PfId:        input.PfID,
		StartDate:   int64(input.StartDate),
		EndDate:     int64(*input.EndDate),
		Title:       input.Title,
		Company:     input.CompanyName,
		Description: description,
		CreatedAt:   currentTime,
		CreatedBy:   email_creator,
		UpdatedAt:   currentTime,
		UpdatedBy:   email_creator,
		Status:      input.Status,
	}

	insertQuery := CassUserSession.Query(vendorz.VendorExperienceTable.Insert()).BindStruct(exp)
	if err := insertQuery.ExecRelease(); err != nil {
		return "", err
	}

	return expId, nil
}

func InviteUserWithRole(ctx context.Context, emails []string, lspID string, role *string) ([]*model.InviteResponse, error) {
	roles := []string{"admin", "learner", "vendor"}
	isPresent := false
	for _, vv := range roles {
		v := vv
		if v == *role {
			isPresent = true
		}
	}
	if !isPresent {
		l := "learner"
		role = &l
	}
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	users := []userz.User{}
	var res []*model.InviteResponse
	getQueryStr := fmt.Sprintf(`SELECT * from userz.users where id='%s' `, emailCreatorID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
	if err := getQuery.SelectRelease(&users); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	for _, dirtyEmail := range emails {
		email := strings.TrimSpace(dirtyEmail)
		if email == email_creator {
			log.Printf("user %v is trying to invite himself", email_creator)
			tmp := &model.InviteResponse{
				Email:   &email,
				Message: "Inviting himself",
			}
			res = append(res, tmp)
			continue
		}
		users := []userz.User{}
		userID := base64.URLEncoding.EncodeToString([]byte(email))
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
			Role:       *role,
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
		if len(users) > 0 {
			tmp := &model.InviteResponse{
				Email:   &email,
				Message: "User already exists",
			}
			res = append(res, tmp)
		} else {
			tmp := &model.InviteResponse{
				Email:   &email,
				Message: "New user",
			}
			res = append(res, tmp)
		}
		_, lspMaps, err := RegisterUsers(ctx, []*model.UserInput{&userInput}, true, len(users) > 0)
		if err != nil {
			return nil, err
		}
		userRoleMap := &model.UserRoleInput{
			UserID:    userID,
			Role:      *role,
			UserLspID: *lspMaps[0].UserLspID,
			IsActive:  true,
			CreatedBy: &email_creator,
			UpdatedBy: &email_creator,
		}
		if lspMaps[0].Status == "" {
			_, err = AddUserRoles(ctx, []*model.UserRoleInput{userRoleMap})
			if err != nil {
				return nil, err
			}
		}
	}
	return res, nil
}

func GetVendors(ctx context.Context, lspID *string) ([]*model.Vendor, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error getting claims from context: %v", err)
		return nil, err
	}
	lsp := claims["lsp_id"].(string)
	if lspID != nil {
		lsp = *lspID
	}
	var res []*model.Vendor
	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_lsp_map WHERE lsp_id = '%s' ALLOW FILTERING`, lsp)
	session, err := cassandra.GetCassSession("vendorz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	getVendorIds := func() (vendorIds []vendorz.VendorLspMap, err error) {
		q := CassUserSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return vendorIds, iter.Select(&vendorIds)
	}
	vendorIds, err := getVendorIds()
	if err != nil {
		return nil, err
	}
	if len(vendorIds) == 0 {
		return nil, nil
	}

	var wg sync.WaitGroup
	for _, vv := range vendorIds {
		v := vv
		wg.Add(1)
		go func(vendorId string) {

			queryStr = fmt.Sprintf(`SELECT * FROM vendorz.vendor WHERE id = %s`, vendorId)
			getVendors := func() (vendors []vendorz.Vendor, err error) {
				q := CassUserSession.Query(queryStr, nil)
				defer q.Release()
				iter := q.Iter()
				return vendors, iter.Select(&vendors)
			}
			vendors, err := getVendors()
			if err != nil {
				return
			}
			if len(vendors) == 0 {
				return
			}
			vendor := vendors[0]

			createdAt := strconv.Itoa(int(vendor.CreatedAt))
			updatedAt := strconv.Itoa(int(vendor.UpdatedAt))
			vendorData := &model.Vendor{
				VendorID:     vendor.VendorId,
				Type:         vendor.Type,
				Level:        vendor.Level,
				Name:         vendor.Name,
				PhotoURL:     &vendor.PhotoUrl,
				Address:      &vendor.Address,
				FacebookURL:  &vendor.Facebook,
				InstagramURL: &vendor.Instagram,
				TwitterURL:   &vendor.Twitter,
				LinkedinURL:  &vendor.LinkedIn,
				CreatedAt:    &createdAt,
				CreatedBy:    &vendor.CreatedBy,
				UpdatedAt:    &updatedAt,
				Status:       vendor.Status,
			}
			res = append(res, vendorData)
			wg.Done()
		}(v.VendorId)
		wg.Wait()
	}
	return res, nil
}
