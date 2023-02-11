package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/contracts/vendorz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/helpers"
	"github.com/zicops/zicops-user-manager/lib/db/bucket"
	"github.com/zicops/zicops-user-manager/lib/googleprojectlib"
)

func AddVendor(ctx context.Context, input *model.VendorInput) (*model.Vendor, error) {
	//create vendor, map it to lsp id, thats all
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting claims: %v", err)
		return nil, err
	}
	lspId := claims["lsp_id"].(string)
	if input.LspID != nil {
		lspId = *input.LspID
	}
	vendorType := "vendor"
	if input.Type == nil || *input.Type == "" {
		*input.Type = vendorType
	}
	email := claims["email"].(string)
	createdAt := time.Now().Unix()

	session, err := cassandra.GetCassSession("vendorz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	vendorId := uuid.New().String()
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
	}

	if input.Users != nil {
		users := changesStringType(input.Users)
		resp, err := MapVendorUser(ctx, vendorId, users, email)
		if err != nil {
			return nil, err
		}
		//check all users, and return appended users
		vendor.Users = resp
	}
	if input.Status != nil {
		vendor.Status = *input.Status
	}
	if input.Level != nil {
		vendor.Level = *input.Level
	}
	if input.Description != nil {
		vendor.Description = *input.Description
	}

	insertQuery := CassUserSession.Query(vendorz.VendorTable.Insert()).BindStruct(vendor)
	if err = insertQuery.Exec(); err != nil {
		return nil, err
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
		return nil, err
	}

	ca := strconv.Itoa(int(createdAt))
	res := &model.Vendor{
		VendorID:     vendorId,
		Type:         vendor.Type,
		Level:        vendor.Level,
		Name:         vendor.Name,
		Description:  &vendor.Description,
		PhotoURL:     &vendor.PhotoUrl,
		Users:        input.Users,
		Address:      &vendor.Address,
		Website:      &vendor.Website,
		FacebookURL:  &vendor.Facebook,
		InstagramURL: &vendor.Instagram,
		TwitterURL:   &vendor.Twitter,
		LinkedinURL:  &vendor.LinkedIn,
		CreatedAt:    &ca,
		CreatedBy:    &email,
		Status:       *input.Status,
	}

	return res, nil
}

func UpdateVendor(ctx context.Context, input *model.VendorInput) (*model.Vendor, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting claims from context: %v", err)
		return nil, nil
	}
	email := claims["email"].(string)
	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor WHERE id = '%s'`, *input.VendorID)
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
	if input.Description != nil {
		updatedCols = append(updatedCols, "description")
		vendor.Description = *input.Description
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
		updatedCols = append(updatedCols, "users")
		users := changesStringType(input.Users)
		resp, err := MapVendorUser(ctx, *input.VendorID, users, email)
		if err != nil {
			return nil, err
		}
		vendor.Users = resp
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
	if input.Type != nil {
		vendor.Type = *input.Type
		updatedCols = append(updatedCols, "type")
	}

	if input.Name != nil {
		var vendorsName []vendorz.Vendor
		queryName := fmt.Sprintf(`SELECT * FROM vendorz.vendor WHERE name = '%s' ALLOW FILTERING`, *input.Name)
		getQueryName := CassUserSession.Query(queryName, nil)
		if err = getQueryName.SelectRelease(&vendorsName); err != nil {
			return nil, err
		}
		if len(vendorsName) == 0 {
			vendor.Name = *input.Name
			updatedCols = append(updatedCols, "name")
			//if the name we have entered is different from vendor id's original name, means we are trying to update the name of vendor to something else
			//which is not unique
		} else {
			for _, vv := range vendorsName {
				v := vv
				if v.Name == *input.Name && input.VendorID != &v.VendorId {
					return nil, errors.New("name needs to be unique, cant be updated")
				}
			}
		}
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

	admins, err := GetVendorAdmins(ctx, *input.VendorID)
	if err != nil {
		log.Printf("Not able to get users: %v", err)
	}
	var adminNames []*string
	for _, v := range admins {
		tmp := v.Email
		adminNames = append(adminNames, &tmp)
	}

	res := &model.Vendor{
		VendorID:     vendor.VendorId,
		Type:         vendor.Type,
		Level:        vendor.Level,
		Name:         vendor.Name,
		Users:        adminNames,
		Description:  &vendor.Description,
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

func MapVendorUser(ctx context.Context, vendorId string, users []string, creator string) ([]string, error) {

	session, err := cassandra.GetCassSession("vendorz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	//get all the emails already mapped with that vendor
	var mappedUsers []vendorz.VendorUserMap
	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_user_map WHERE vendor_id = '%s' ALLOW FILTERING`, vendorId)
	query := CassUserSession.Query(queryStr, nil)
	if err = query.SelectRelease(&mappedUsers); err != nil {
		return nil, err
	}
	var resp []string
	for _, vv := range mappedUsers {
		v := vv
		email, err := base64.URLEncoding.DecodeString(v.UserId)
		if err != nil {
			return nil, err
		}
		resp = append(resp, string(email))
	}

	for _, email := range users {
		if !IsEmailValid(email) {
			continue
		}
		//check if already exists
		userId := base64.URLEncoding.EncodeToString([]byte(email))
		var res []vendorz.VendorUserMap
		queryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_user_map WHERE vendor_id = '%s' AND user_id = '%s' ALLOW FILTERING`, vendorId, userId)
		getQuery := CassUserSession.Query(queryStr, nil)
		if err = getQuery.SelectRelease(&res); err != nil {
			return nil, err
		}
		if len(res) == 0 {
			createdAt := time.Now().Unix()
			vendorUserMap := vendorz.VendorUserMap{
				VendorId:  vendorId,
				UserId:    userId,
				CreatedAt: createdAt,
				CreatedBy: creator,
				Status:    "",
			}
			insertVendorUserMap := CassUserSession.Query(vendorz.VendorUserMapTable.Insert()).BindStruct(vendorUserMap)
			if err = insertVendorUserMap.Exec(); err != nil {
				return nil, err
			}
			resp = append(resp, email)
		} else {
			continue
		}
	}
	return resp, nil
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
			storageC := bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err = storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				log.Printf("Failed to upload image to course: %v", err.Error())
				return
			}

			queryStr = fmt.Sprintf(`SELECT * FROM vendorz.vendor WHERE id = '%s' ALLOW FILTERING`, vendorId)
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

			//vendorAdmins
			admins, err := GetVendorAdmins(ctx, vendorId)
			if err != nil {
				log.Printf("Got error while getting vendor Admins for %v: %v", vendorId, err)
			}
			var usersEmail []*string
			for _, vv := range admins {
				v := vv
				usersEmail = append(usersEmail, &v.Email)
			}

			//photo
			photoUrl := ""
			if vendor.PhotoBucket != "" {
				photoUrl = storageC.GetSignedURLForObject(vendor.PhotoBucket)
			} else {
				photoUrl = vendor.PhotoUrl
			}
			createdAt := strconv.Itoa(int(vendor.CreatedAt))
			updatedAt := strconv.Itoa(int(vendor.UpdatedAt))
			vendorData := &model.Vendor{
				VendorID:     vendor.VendorId,
				Type:         vendor.Type,
				Level:        vendor.Level,
				Name:         vendor.Name,
				PhotoURL:     &photoUrl,
				Description:  &vendor.Description,
				Website:      &vendor.Website,
				Address:      &vendor.Address,
				Users:        usersEmail,
				FacebookURL:  &vendor.Facebook,
				InstagramURL: &vendor.Instagram,
				TwitterURL:   &vendor.Twitter,
				LinkedinURL:  &vendor.LinkedIn,
				CreatedAt:    &createdAt,
				CreatedBy:    &vendor.CreatedBy,
				UpdatedAt:    &updatedAt,
				UpdatedBy:    &vendor.UpdatedBy,
				Status:       vendor.Status,
			}
			res = append(res, vendorData)
			wg.Done()
		}(v.VendorId)
	}
	wg.Wait()
	return res, nil
}

func GetVendorAdmins(ctx context.Context, vendorID string) ([]*model.User, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var userIds []vendorz.VendorUserMap
	session, err := cassandra.GetCassSession("vendorz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_user_map WHERE vendor_id = '%s' ALLOW FILTERING`, vendorID)

	getUserIds := func() (vendorUserIds []vendorz.VendorUserMap, err error) {
		q := CassUserSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return vendorUserIds, iter.Select(&vendorUserIds)
	}
	userIds, err = getUserIds()
	if err != nil {
		return nil, err
	}

	if len(userIds) == 0 {
		return nil, nil
	}
	res := make([]*model.User, len(userIds))

	var wg sync.WaitGroup
	for k, vv := range userIds {
		v := vv
		wg.Add(1)
		//iterate over these userIds and return user details
		go func(userId string, k int) {
			//return user data

			email, err := base64.URLEncoding.DecodeString(userId)
			if err != nil {
				return
			}

			if !IsEmailValid(string(email)) || string(email) == "" {
				return
			}

			usersession, err := cassandra.GetCassSession("userz")
			if err != nil {
				return
			}
			CassUserSession := usersession

			QueryStr := fmt.Sprintf(`SELECT * FROM userz.users WHERE id = '%s'`, userId)
			getUserData := func() (users []userz.User, err error) {
				q := CassUserSession.Query(QueryStr, nil)
				defer q.Release()
				iter := q.Iter()
				return users, iter.Select(&users)
			}
			users, err := getUserData()
			if err != nil {
				return
			}

			if len(users) == 0 {
				return
			}
			user := users[0]

			createdAt := strconv.Itoa(int(user.CreatedAt))
			updatedAt := strconv.Itoa(int(user.UpdatedAt))
			temp := &model.User{
				ID:         &user.ID,
				FirstName:  user.FirstName,
				LastName:   user.LastName,
				Status:     user.Status,
				Role:       user.Role,
				IsVerified: user.IsVerified,
				IsActive:   user.IsActive,
				Gender:     user.Gender,
				CreatedAt:  createdAt,
				CreatedBy:  &user.CreatedBy,
				UpdatedAt:  updatedAt,
				UpdatedBy:  &user.UpdatedBy,
				Email:      user.Email,
				PhotoURL:   &user.PhotoURL,
			}
			userData := temp
			res[k] = userData

			wg.Done()
		}(v.UserId, k)
	}
	wg.Wait()
	return res, nil
}

func IsEmailValid(e string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(e)
}

func GetVendorDetails(ctx context.Context, vendorID string) (*model.Vendor, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := cassandra.GetCassSession("vendorz")
	if err != nil {
		return nil, err
	}
	CassSession := session
	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor WHERE id = '%s'`, vendorID)
	getVendorDetails := func() (vendors []vendorz.Vendor, err error) {
		q := CassSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return vendors, iter.Select(&vendors)
	}
	vendors, err := getVendorDetails()
	if err != nil {
		return nil, err
	}
	if len(vendors) == 0 {
		return nil, nil
	}
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		log.Printf("Failed to upload image to course: %v", err.Error())
		return nil, err
	}

	vendor := vendors[0]
	createdAt := strconv.Itoa(int(vendor.CreatedAt))
	updatedAt := strconv.Itoa(int(vendor.UpdatedAt))

	//vendorAdmins
	admins, err := GetVendorAdmins(ctx, vendorID)
	if err != nil {
		log.Printf("Got error while getting vendor Admins for %v: %v", vendorID, err)
	}
	var usersEmail []*string
	for _, vv := range admins {
		v := vv
		usersEmail = append(usersEmail, &v.Email)
	}

	//photo
	photoUrl := ""
	if vendor.PhotoBucket != "" {
		photoUrl = storageC.GetSignedURLForObject(vendor.PhotoBucket)
	} else {
		photoUrl = vendor.PhotoUrl
	}

	res := &model.Vendor{
		VendorID:     vendor.VendorId,
		Type:         vendor.Type,
		Level:        vendor.Level,
		Name:         vendor.Name,
		Description:  &vendor.Description,
		PhotoURL:     &photoUrl,
		Address:      &vendor.Address,
		Users:        usersEmail,
		Website:      &vendor.Website,
		FacebookURL:  &vendor.Facebook,
		InstagramURL: &vendor.Instagram,
		TwitterURL:   &vendor.Twitter,
		LinkedinURL:  &vendor.LinkedIn,
		CreatedAt:    &createdAt,
		CreatedBy:    &vendor.CreatedBy,
		UpdatedAt:    &updatedAt,
		UpdatedBy:    &vendor.UpdatedBy,
		Status:       vendor.Status,
	}
	return res, nil
}

func GetPaginatedVendors(ctx context.Context, lspID *string, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedVendors, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting context: %v", err)
		return nil, err
	}
	lsp := claims["lsp_id"].(string)
	if lspID != nil {
		lsp = *lspID
	}
	var newPage []byte

	session, err := cassandra.GetCassSession("vendorz")
	if err != nil {
		return nil, err
	}
	CassVendorSession := session

	if pageCursor != nil && *pageCursor != "" {
		page, err := global.CryptSession.DecryptString(*pageCursor, nil)
		if err != nil {
			return nil, err
		}
		newPage = page
	}

	var vendorIds []vendorz.VendorLspMap
	var newCursor string
	var pageSizeInt int
	if pageSize != nil {
		pageSizeInt = *pageSize
	} else {
		pageSizeInt = 10
	}

	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_lsp_map where lsp_id = '%s' ALLOW FILTERING`, lsp)
	getVendorIds := func(page []byte) (vendors []vendorz.VendorLspMap, nextPage []byte, err error) {
		q := CassVendorSession.Query(queryStr, nil)
		defer q.Release()
		q.PageState(page)
		q.PageSize(pageSizeInt)
		iter := q.Iter()
		return vendors, iter.PageState(), iter.Select(&vendors)
	}

	vendorIds, newPage, err = getVendorIds(newPage)
	if err != nil {
		return nil, err
	}
	if len(newPage) != 0 {
		newCursor, err = global.CryptSession.EncryptAsString(newPage, nil)
		if err != nil {
			return nil, err
		}
	}
	if len(vendorIds) == 0 {
		return nil, nil
	}
	res := make([]*model.Vendor, len(vendorIds))

	var outputResponse model.PaginatedVendors
	var wg sync.WaitGroup
	for k, vv := range vendorIds {
		v := vv
		wg.Add(1)
		go func(vendorId string, k int) {

			queryStr = fmt.Sprintf(`SELECT * FROM vendorz.vendor WHERE id = '%s' ALLOW FILTERING`, vendorId)
			getVendors := func() (vendors []vendorz.Vendor, err error) {
				q := CassVendorSession.Query(queryStr, nil)
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
				Website:      &vendor.Website,
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
			res[k] = vendorData
			wg.Done()
		}(v.VendorId, k)
	}
	wg.Wait()
	outputResponse.Vendors = res
	outputResponse.Direction = direction
	outputResponse.PageSize = pageSize
	outputResponse.PageCursor = &newCursor
	return &outputResponse, nil
}
