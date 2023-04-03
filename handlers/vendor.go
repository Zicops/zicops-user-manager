package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/contracts/vendorz"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/lib/db/bucket"
	"github.com/zicops/zicops-user-manager/lib/googleprojectlib"
	"github.com/zicops/zicops-user-manager/lib/identity"
	"github.com/zicops/zicops-user-manager/lib/utils"
)

func AddVendor(ctx context.Context, input *model.VendorInput) (*model.Vendor, error) {
	//create vendor, map it to lsp id, thats all
	claims, err := identity.GetClaimsFromContext(ctx)
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

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	check := checkName(ctx, *input.Name, lspId)
	if !check {
		return nil, errors.New("vendor with same name already exists")
	}

	vendorId := uuid.New().String()
	//create vendor
	vendor := vendorz.Vendor{
		VendorId:  vendorId,
		Name:      *input.Name,
		Type:      *input.Type,
		LspId:     lspId,
		CreatedAt: createdAt,
		CreatedBy: email,
		UpdatedAt: createdAt,
		UpdatedBy: email,
	}
	if input.Address != nil {
		vendor.Address = *input.Address
	}
	if input.Phone != nil {
		vendor.Phone = *input.Phone
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
		bucketPath := fmt.Sprintf("%s/%s/%s", "vendor", vendor.Name, base64.URLEncoding.EncodeToString([]byte(input.Photo.Filename)))
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
		url := storageC.GetSignedURLForObject(ctx, bucketPath)
		vendor.PhotoBucket = bucketPath
		vendor.PhotoUrl = url
	}

	if input.Users != nil {
		users := ChangesStringType(input.Users)
		// resp, err := MapVendorUser(ctx, vendorId, users, email, lspId)
		// if err != nil {
		// 	return nil, err
		// }
		// //check all users, and return appended users
		vendor.Users = users
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
	wordsRaw := strings.Fields(*input.Name)
	var words []string
	for _, v := range wordsRaw {
		words = append(words, strings.ToLower(v))
	}
	vendorLspMap := vendorz.VendorLspMap{
		VendorId:  vendorId,
		LspId:     lspId,
		CreatedAt: createdAt,
		CreatedBy: email,
		UpdatedAt: createdAt,
		UpdatedBy: email,
		Status:    *input.Status,
		Type:      *input.Type,
		Words:     words,
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
		Phone:        &vendor.Phone,
		LspID:        &vendor.LspId,
		CreatedAt:    &ca,
		CreatedBy:    &email,
		UpdatedAt:    &ca,
		UpdatedBy:    &email,
		Status:       input.Status,
	}

	return res, nil
}

func checkName(ctx context.Context, name string, lsp string) bool {
	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Printf("Got error while getting session: %v", err)
		return false
	}
	CassSession := session

	qryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_lsp_map WHERE lsp_id = '%s' ALLOW FILTERING`, lsp)
	getVendors := func() (vendorDetails []vendorz.VendorLspMap, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return vendorDetails, iter.Select(&vendorDetails)
	}

	vendorMaps, err := getVendors()
	if err != nil {
		log.Printf("Got error while getting vendor details: %v", err)
		return false
	}
	if len(vendorMaps) == 0 {
		return false
	}

	for _, vv := range vendorMaps {
		v := vv
		query := fmt.Sprintf(`SELECT * FROM vendorz.vendor WHERE id='%s' ALLOW FILTERING`, v.VendorId)
		getVendorDetails := func() (vendorData []vendorz.Vendor, err error) {
			q := CassSession.Query(query, nil)
			defer q.Release()
			iter := q.Iter()
			return vendorData, iter.Select(&vendorData)
		}

		vendors, err := getVendorDetails()
		if err != nil {
			log.Printf("Got error while getting vendor details: %v", err)
			return false
		}
		if len(vendors) == 0 {
			return false
		}

		vendor := vendors[0]
		if vendor.Name == name {
			return false
		}
	}
	return true
}

func UpdateVendor(ctx context.Context, input *model.VendorInput) (*model.Vendor, error) {
	if input.VendorID == nil {
		return nil, errors.New("please pass vendor id")
	}
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting claims from context: %v", err)
		return nil, nil
	}
	email := claims["email"].(string)
	lsp := claims["lsp_id"].(string)
	v_id := *input.VendorID
	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor WHERE id = '%s'`, v_id)
	session, err := global.CassPool.GetSession(ctx, "vendorz")
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
	updatedCols := []string{}

	if input.Level != nil {
		vendor.Level = *input.Level
		updatedCols = append(updatedCols, "level")
	}
	if input.Description != nil {
		vendor.Description = *input.Description
		updatedCols = append(updatedCols, "description")
	}
	if input.Address != nil {
		vendor.Address = *input.Address
		updatedCols = append(updatedCols, "address")
	}
	if input.Website != nil {
		vendor.Website = *input.Website
		updatedCols = append(updatedCols, "website")
	}
	if input.FacebookURL != nil {
		vendor.Facebook = *input.FacebookURL
		updatedCols = append(updatedCols, "facebook")
	}
	if input.InstagramURL != nil {
		vendor.Instagram = *input.InstagramURL
		updatedCols = append(updatedCols, "instagram")
	}
	if input.TwitterURL != nil {
		vendor.Twitter = *input.TwitterURL
		updatedCols = append(updatedCols, "twitter")
	}
	if input.LinkedinURL != nil {
		vendor.LinkedIn = *input.LinkedinURL
		updatedCols = append(updatedCols, "linkedin")
	}

	if input.Users != nil {
		users := ChangesStringType(input.Users)
		// resp, err := MapVendorUser(ctx, *input.VendorID, users, email, lsp)
		// if err != nil {
		// 	return nil, err
		// }
		vendor.Users = users
		updatedCols = append(updatedCols, "users")
	}
	if input.Phone != nil {
		vendor.Phone = *input.Phone
		updatedCols = append(updatedCols, "phone")
	}
	if input.Status != nil {
		vendor.Status = *input.Status
		updatedCols = append(updatedCols, "status")

		//updation here, cause created_At is a key
		err = updateMap(ctx, vendor, lsp, email)
		if err != nil {
			return nil, err
		}
		if *input.Status == "disable" {
			changeUserLspMapOfUsers(ctx, *input.VendorID, email, lsp)
		}
		if *input.Status == "active" && vendor.Status != "active" {
			changeUserLspMapOfUsersToActive(ctx, *input.VendorID, email, lsp)
		}
	}

	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		return nil, err
	}
	if input.Photo != nil {
		bucketPath := fmt.Sprintf("%s/%s/%s", "vendor", vendor.Name, base64.URLEncoding.EncodeToString([]byte(input.Photo.Filename)))
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
		url := storageC.GetSignedURLForObject(ctx, bucketPath)
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
		vendor.Name = *input.Name
		updatedCols = append(updatedCols, "name")
	}

	createdAt := strconv.Itoa(int(vendor.CreatedAt))
	updatedAt := time.Now().Unix()

	admins, err := GetVendorAdmins(ctx, *input.VendorID)
	if err != nil {
		log.Printf("Not able to get users: %v", err)
	}
	var adminNames []*string
	for _, v := range admins {
		if v == nil {
			continue
		}
		tmp := v.Email
		adminNames = append(adminNames, &tmp)
	}

	if len(updatedCols) > 0 {
		updatedCols = append(updatedCols, "updated_by")
		vendor.UpdatedBy = email

		updatedCols = append(updatedCols, "updated_at")
		vendor.UpdatedAt = updatedAt

		upStms, uNames := vendorz.VendorTable.Update(updatedCols...)
		updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&vendor)
		if err = updateQuery.ExecRelease(); err != nil {
			log.Printf("Error updating user: %v", err)
			return nil, err
		}
	}

	ua := strconv.Itoa(int(updatedAt))
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
		Phone:        &vendor.Phone,
		CreatedAt:    &createdAt,
		CreatedBy:    &vendor.CreatedBy,
		UpdatedAt:    &ua,
		UpdatedBy:    &email,
		Status:       &vendor.Status,
	}
	return res, nil
}

func changeUserLspMapOfUsersToActive(ctx context.Context, vendorId string, email string, lsp string) error {
	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return err
	}
	CassSession := session

	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_user_map WHERE vendor_id = '%s' AND status='active' ALLOW FILTERING`, vendorId)

	getUserIds := func() (vendorUserIds []vendorz.VendorUserMap, err error) {
		q := CassSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return vendorUserIds, iter.Select(&vendorUserIds)
	}
	userIds, err := getUserIds()
	if err != nil {
		return err
	}

	if len(userIds) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	for _, vvv := range userIds {
		vv := vvv
		wg.Add(1)
		//iterate over these userIds and return user details
		go func(userId string) {
			//return user data

			usersession, err := global.CassPool.GetSession(ctx, "userz")
			if err != nil {
				return
			}
			CassUserSession := usersession

			QueryStr := fmt.Sprintf(`SELECT * FROM userz.user_lsp_map WHERE user_id = '%s' AND lsp_id='%s' ALLOW FILTERING`, userId, lsp)
			getUserData := func() (users []userz.UserLsp, err error) {
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
			if user.Status == "invited_disable" {
				user.Status = ""
				user.UpdatedAt = time.Now().Unix()
				user.UpdatedBy = email

				updatedCols := []string{"status", "updated_at", "updated_by"}
				stmt, names := userz.UserLspTable.Update(updatedCols...)
				updateQuery := CassUserSession.Query(stmt, names).BindStruct(&user)
				if err = updateQuery.ExecRelease(); err != nil {
					log.Printf("Error: %v", err)
					return
				}
			} else if user.Status == "disable" {
				user.Status = "active"
				user.UpdatedAt = time.Now().Unix()
				user.UpdatedBy = email

				updatedCols := []string{"status", "updated_at", "updated_by"}
				stmt, names := userz.UserLspTable.Update(updatedCols...)
				updateQuery := CassUserSession.Query(stmt, names).BindStruct(&user)
				if err = updateQuery.ExecRelease(); err != nil {
					log.Printf("Error: %v", err)
					return
				}
			}

			wg.Done()
		}(vv.UserId)
	}
	wg.Wait()

	return nil
}

func changeUserLspMapOfUsers(ctx context.Context, vendorId string, email string, lsp string) error {

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return err
	}
	CassSession := session

	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_user_map WHERE vendor_id = '%s' ALLOW FILTERING`, vendorId)

	getUserIds := func() (vendorUserIds []vendorz.VendorUserMap, err error) {
		q := CassSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return vendorUserIds, iter.Select(&vendorUserIds)
	}
	userIds, err := getUserIds()
	if err != nil {
		return err
	}

	if len(userIds) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	for _, vvv := range userIds {
		vv := vvv
		wg.Add(1)
		//iterate over these userIds and return user details
		go func(userId string) {
			//return user data

			usersession, err := global.CassPool.GetSession(ctx, "userz")
			if err != nil {
				return
			}
			CassUserSession := usersession

			QueryStr := fmt.Sprintf(`SELECT * FROM userz.user_lsp_map WHERE user_id = '%s' ALLOW FILTERING`, userId)
			getUserData := func() (users []userz.UserLsp, err error) {
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
			for _, user := range users {
				if user.Status == "" {
					user.Status = "invited_disable"
					user.UpdatedAt = time.Now().Unix()
					user.UpdatedBy = email

					updatedCols := []string{"status", "updated_at", "updated_by"}
					stmt, names := userz.UserLspTable.Update(updatedCols...)
					updateQuery := CassUserSession.Query(stmt, names).BindStruct(&user)
					if err = updateQuery.ExecRelease(); err != nil {
						log.Printf("Error: %v", err)
						return
					}
				} else {
					user.Status = "disable"
					user.UpdatedAt = time.Now().Unix()
					user.UpdatedBy = email

					updatedCols := []string{"status", "updated_at", "updated_by"}
					stmt, names := userz.UserLspTable.Update(updatedCols...)
					updateQuery := CassUserSession.Query(stmt, names).BindStruct(&user)
					if err = updateQuery.ExecRelease(); err != nil {
						log.Printf("Error: %v", err)
						return
					}
				}

			}

			wg.Done()
		}(vv.UserId)
	}
	wg.Wait()

	return nil
}

func updateMap(ctx context.Context, vendor vendorz.Vendor, lsp string, email string) error {

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return err
	}
	CassUserSession := session
	qry := fmt.Sprintf(`SELECT * FROM vendorz.vendor_lsp_map where vendor_id='%s' AND lsp_id='%s' ALLOW FILTERING`, vendor.VendorId, lsp)
	getMaps := func() (vendorLspMaps []vendorz.VendorLspMap, err error) {
		q := CassUserSession.Query(qry, nil)
		defer q.Release()
		iter := q.Iter()
		return vendorLspMaps, iter.Select(&vendorLspMaps)
	}

	lspMaps, err := getMaps()
	if err != nil {
		return err
	}
	if len(lspMaps) == 0 {
		return nil
	}
	lspMap := lspMaps[0]
	lspMap.Status = vendor.Status
	lspMap.UpdatedAt = time.Now().Unix()
	lspMap.UpdatedBy = email
	stmt, names := vendorz.VendorLspMapTable.Update("status", "updated_at", "updated_by")
	updatedQuery := CassUserSession.Query(stmt, names).BindStruct(&lspMap)
	if err = updatedQuery.ExecRelease(); err != nil {
		return err
	}

	return nil
}

// func MapVendorUser(ctx context.Context, vendorId string, users []string, creator string, lsp string) ([]string, error) {

// 	sessionUser, err := global.CassPool.GetSession(ctx, "userz")
// 	if err != nil {
// 		return nil, err
// 	}
// 	CassSession := sessionUser

// 	session, err := global.CassPool.GetSession(ctx, "vendorz")
// 	if err != nil {
// 		return nil, err
// 	}
// 	CassUserSession := session

// 	//get all the emails already mapped with that vendor
// 	var mappedUsers []vendorz.VendorUserMap
// 	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_user_map WHERE vendor_id = '%s' ALLOW FILTERING`, vendorId)
// 	getUsers := func() (users []vendorz.VendorUserMap, err error) {
// 		q := CassUserSession.Query(queryStr, nil)
// 		defer q.Release()
// 		iter := q.Iter()
// 		return users, iter.Select(&users)
// 	}
// 	mappedUsers, err = getUsers()
// 	if err != nil {
// 		return nil, err
// 	}
// 	var resp []string
// 	for _, vv := range mappedUsers {
// 		v := vv
// 		email, err := base64.URLEncoding.DecodeString(v.UserId)
// 		if err != nil {
// 			return nil, err
// 		}

// 		//iterate over all the data from database, and compare with data sent from frontend
// 		//for now lets assume we are going to delete the value
// 		//if value is found, then ignore
// 		//if value not found in list sent from frontend, delete the value
// 		flagDel := true
// 		//check if value exists,
// 		for _, kk := range users {
// 			k := kk
// 			userId := base64.URLEncoding.EncodeToString([]byte(k))
// 			if userId == v.UserId {
// 				flagDel = false
// 			}
// 		}
// 		if flagDel {
// 			//disable the map
// 			err := changeStatusOfAllUsers(ctx, vendorId, v.UserId, lsp, creator, "disable")
// 			if err != nil {
// 				return nil, err
// 			}

// 		}

// 		resp = append(resp, string(email))
// 	}

// 	for _, dirtyEmail := range users {
// 		if !IsEmailValid(dirtyEmail) {
// 			continue
// 		}
// 		//check if already exists

// 		email := strings.ToLower(dirtyEmail)
// 		userId := base64.URLEncoding.EncodeToString([]byte(email))

// 		qryStr := fmt.Sprintf(`SELECT * FROM userz.user_lsp_map WHERE user_id='%s' AND lsp_id='%s' ALLOW FILTERING`, userId, lsp)
// 		getUserDetails := func() (userMap []userz.UserLsp, err error) {
// 			q := CassSession.Query(qryStr, nil)
// 			defer q.Release()
// 			iter := q.Iter()
// 			return userMap, iter.Select(&userMap)
// 		}
// 		users, err := getUserDetails()
// 		if err != nil {
// 			return nil, err
// 		}
// 		if len(users) != 0 {
// 			continue
// 		} else {
// 			var res []vendorz.VendorUserMap
// 			queryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_user_map WHERE vendor_id = '%s' AND user_id = '%s' ALLOW FILTERING`, vendorId, userId)
// 			getQuery := func() (maps []vendorz.VendorUserMap, err error) {
// 				q := CassUserSession.Query(queryStr, nil)
// 				defer q.Release()
// 				iter := q.Iter()
// 				return maps, iter.Select(&maps)
// 			}
// 			res, err = getQuery()
// 			if err != nil {
// 				return nil, err
// 			}
// 			if len(res) == 0 {
// 				createdAt := time.Now().Unix()
// 				vendorUserMap := vendorz.VendorUserMap{
// 					VendorId:  vendorId,
// 					UserId:    userId,
// 					CreatedAt: createdAt,
// 					CreatedBy: creator,
// 					Status:    "active",
// 				}
// 				insertVendorUserMap := CassUserSession.Query(vendorz.VendorUserMapTable.Insert()).BindStruct(vendorUserMap)
// 				if err = insertVendorUserMap.Exec(); err != nil {
// 					return nil, err
// 				}
// 				resp = append(resp, email)
// 			} else {
// 				continue
// 			}
// 		}
// 	}
// 	return resp, nil
// }

// func changeStatusOfAllUsers(ctx context.Context, vendorId string, userId string, lsp string, email string, status string) error {

// 	session, err := global.CassPool.GetSession(ctx, "vendorz")
// 	if err != nil {
// 		return err
// 	}
// 	CassSession := session
// 	qryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_user_map WHERE vendor_id='%s' AND user_id='%s' ALLOW FILTERING`, vendorId, userId)
// 	getDetails := func() (maps []vendorz.VendorUserMap, err error) {
// 		q := CassSession.Query(qryStr, nil)
// 		defer q.Release()
// 		iter := q.Iter()
// 		return maps, iter.Select(&maps)
// 	}
// 	vendorLspMaps, err := getDetails()
// 	if err != nil {
// 		return err
// 	}
// 	if len(vendorLspMaps) == 0 {
// 		return errors.New("map not found")
// 	}

// 	vendorLspMap := vendorLspMaps[0]
// 	vendorLspMap.Status = status
// 	vendorLspMap.UpdatedAt = time.Now().Unix()
// 	vendorLspMap.UpdatedBy = email
// 	updatedCols := []string{"status", "updated_at", "updated_by"}
// 	stmt, names := vendorz.VendorLspMapTable.Update(updatedCols...)
// 	updatedQuery := CassSession.Query(stmt, names).BindStruct(&vendorLspMap)
// 	if err = updatedQuery.ExecRelease(); err != nil {
// 		return err
// 	}

// 	session, err = global.CassPool.GetSession(ctx, "userz")
// 	if err != nil {
// 		return err
// 	}
// 	CassUserSession := session
// 	query := fmt.Sprintf(`SELECT * FROM userz.user_lsp_map WHERE user_id='%s' AND lsp_id='%s' ALLOW FILTERING`, userId, lsp)
// 	getUserData := func() (maps []userz.UserLsp, err error) {
// 		q := CassUserSession.Query(query, nil)
// 		defer q.Release()
// 		iter := q.Iter()
// 		return maps, iter.Select(&maps)
// 	}
// 	userLspMaps, err := getUserData()
// 	if err != nil {
// 		return err
// 	}
// 	if len(userLspMaps) == 0 {
// 		return fmt.Errorf("user does not exist: %v", userId)
// 	}
// 	userLspMap := userLspMaps[0]
// 	userLspMap.Status = status
// 	userLspMap.UpdatedAt = time.Now().Unix()
// 	userLspMap.UpdatedBy = email
// 	stmt, names = userz.UserLspTable.Update(updatedCols...)
// 	updateQuery := CassUserSession.Query(stmt, names).BindStruct(&userLspMap)
// 	if err = updateQuery.ExecRelease(); err != nil {
// 		return err
// 	}

// 	return nil
// }

func ChangesStringType(input []*string) []string {
	var res []string
	for _, vv := range input {
		if vv == nil {
			continue
		}
		v := vv
		res = append(res, *v)
	}
	return res
}

func CreateProfileVendor(ctx context.Context, input *model.VendorProfileInput) (*model.VendorProfile, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting claims: %v", err)
		return nil, err
	}
	createdBy := claims["email"].(string)
	lspId := claims["lsp_id"].(string)

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Printf("Got error while getting session: %v", err)
		return nil, err
	}
	CassSession := session
	var words []string

	email := strings.ToLower(input.Email)
	pfId := base64.URLEncoding.EncodeToString([]byte(email))

	verifyingQuery := fmt.Sprintf(`SELECT * FROM vendorz.profile WHERE pf_id = '%s' AND vendor_id = '%s' ALLOW FILTERING`, pfId, input.VendorID)
	getProfileDetail := func() (exp []vendorz.VendorProfile, err error) {
		q := CassSession.Query(verifyingQuery, nil)
		defer q.Release()
		iter := q.Iter()
		return exp, iter.Select(&exp)
	}

	profileDetails, err := getProfileDetail()

	if err != nil {
		log.Printf("Got error while getting data from profile experience: %v", err)
	}
	if len(profileDetails) != 0 {
		return nil, fmt.Errorf("email already in use")
	}

	profile := vendorz.VendorProfile{
		PfId:     pfId,
		VendorId: input.VendorID,
		Email:    email,
		LspId:    lspId,
	}
	if input.FirstName != nil {
		profile.FirstName = *input.FirstName
		firstName := strings.ToLower(*input.FirstName)
		nameArray := strings.Fields(firstName)
		words = append(words, nameArray...)
	}
	if input.LastName != nil {
		profile.LastName = *input.LastName
		lastName := strings.ToLower(*input.LastName)
		namesArray := strings.Fields(lastName)
		words = append(words, namesArray...)
	}
	profile.Name = words
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		return nil, err
	}
	if input.Photo != nil {
		bucketPath := fmt.Sprintf("%s/%s/%s/%s", "vendor", "profile", pfId, base64.URLEncoding.EncodeToString([]byte(input.Photo.Filename)))
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
		url := storageC.GetSignedURLForObject(ctx, bucketPath)
		profile.PhotoBucket = bucketPath
		profile.PhotoURL = url
	}
	if input.Description != nil {
		profile.Description = *input.Description
	}
	if input.Languages != nil {
		tmp := ChangesStringType(input.Languages)
		profile.Languages = tmp
	}
	if input.SmeExpertise != nil {
		tmp := ChangesStringType(input.SmeExpertise)
		profile.SMEExpertise = tmp
		profile.Sme = true
	}
	if input.ClassroomExpertise != nil {
		tmp := ChangesStringType(input.ClassroomExpertise)
		profile.ClassroomExpertise = tmp
		profile.Crt = true
	}
	if input.ContentDevelopment != nil {
		tmp := ChangesStringType(input.ContentDevelopment)
		profile.ContentDevelopment = tmp
		profile.Cd = true
	}
	if input.Experience != nil {
		tmp := ChangesStringType(input.Experience)
		profile.Experience = tmp
	}
	if input.ExperienceYears != nil {
		profile.ExperienceYears = *input.ExperienceYears
	}
	if input.IsSpeaker != nil {
		profile.IsSpeaker = *input.IsSpeaker
	}
	if input.Status != nil {
		profile.Status = *input.Status
	}
	if input.Phone != nil {
		profile.Phone = *input.Phone
	}
	profile.CreatedAt = time.Now().Unix()
	profile.CreatedBy = createdBy
	profile.UpdatedAt = time.Now().Unix()
	profile.UpdatedBy = email

	insertQuery := CassSession.Query(vendorz.VendorProfileTable.Insert()).BindStruct(profile)
	if err = insertQuery.Exec(); err != nil {
		log.Printf("Got error while inserting data: %v", err)
		return nil, err
	}

	createdAt := strconv.Itoa(int(time.Now().Unix()))
	sme := len(input.SmeExpertise) > 0
	cd := len(input.ContentDevelopment) > 0
	crt := len(input.ClassroomExpertise) > 0

	res := model.VendorProfile{
		PfID:               &profile.PfId,
		VendorID:           &profile.VendorId,
		FirstName:          &profile.FirstName,
		LastName:           &profile.LastName,
		Email:              &profile.Email,
		Phone:              &profile.Phone,
		PhotoURL:           &profile.PhotoURL,
		Description:        &profile.Description,
		Language:           input.Languages,
		SmeExpertise:       input.SmeExpertise,
		ClassroomExpertise: input.ClassroomExpertise,
		ContentDevelopment: input.ContentDevelopment,
		Sme:                &sme,
		Crt:                &crt,
		Cd:                 &cd,
		Experience:         input.Experience,
		ExperienceYears:    input.ExperienceYears,
		IsSpeaker:          &profile.IsSpeaker,
		LspID:              &lspId,
		CreatedAt:          &createdAt,
		CreatedBy:          &profile.CreatedBy,
		UpdatedAt:          &createdAt,
		UpdatedBy:          &profile.UpdatedBy,
		Status:             &profile.Status,
	}
	return &res, nil
}

func CreateExperienceVendor(ctx context.Context, input model.ExperienceInput) (*model.ExperienceVendor, error) {

	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting claims: %v", err)
		return nil, err
	}
	email_creator := claims["email"].(string)

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return nil, err
	}

	expId := uuid.New().String()

	email := strings.ToLower(input.Email)
	pfId := base64.URLEncoding.EncodeToString([]byte(email))
	currentTime := time.Now().Unix()
	CassUserSession := session

	exp := vendorz.VendorExperience{
		ExpId:     expId,
		VendorId:  *input.VendorID,
		PfId:      pfId,
		CreatedAt: currentTime,
		CreatedBy: email_creator,
		UpdatedAt: currentTime,
		UpdatedBy: email_creator,
	}
	if input.StartDate != nil {
		exp.StartDate = int64(*input.StartDate)
	}
	if input.EndDate != nil {
		exp.EndDate = int64(*input.EndDate)
	}
	if input.Title != nil {
		exp.Title = *input.Title
	}
	if input.Location != nil {
		exp.Location = *input.Location
	}
	if input.LocationType != nil {
		exp.LocationType = *input.LocationType
	}
	if input.EmployementType != nil {
		exp.EmployementType = *input.EmployementType
	}
	if input.CompanyName != nil {
		exp.Company = *input.CompanyName
	}
	if input.Status != nil {
		exp.Status = *input.Status
	}

	insertQuery := CassUserSession.Query(vendorz.VendorExperienceTable.Insert()).BindStruct(exp)
	if err := insertQuery.ExecRelease(); err != nil {
		return nil, err
	}
	ct := strconv.Itoa(int(currentTime))
	res := model.ExperienceVendor{
		ExpID:           expId,
		VendorID:        *input.VendorID,
		PfID:            pfId,
		StartDate:       input.StartDate,
		EndDate:         input.EndDate,
		Title:           input.Title,
		CompanyName:     input.CompanyName,
		Location:        input.Location,
		LocationType:    input.LocationType,
		EmployementType: input.EmployementType,
		CreatedAt:       &ct,
		CreatedBy:       &email_creator,
		UpdatedAt:       &ct,
		UpdatedBy:       &email_creator,
		Status:          input.Status,
	}

	return &res, nil
}

//user course map - mapping check

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
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	//log.Println(claims["origin"].(string))
	session, err := global.CassPool.GetSession(ctx, "userz")
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
		email = strings.ToLower(email)
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

		_, lspMaps, err := RegisterUsers(ctx, []*model.UserInput{&userInput}, true, len(users) > 0)
		if err != nil {
			return nil, err
		}

		if len(users) > 0 {
			tmp := &model.InviteResponse{
				Email:     &email,
				Message:   "User already exists",
				UserID:    &lspMaps[0].UserID,
				UserLspID: lspMaps[0].UserLspID,
			}
			res = append(res, tmp)
		} else {
			tmp := &model.InviteResponse{
				Email:     &email,
				Message:   "New user",
				UserID:    &lspMaps[0].UserID,
				UserLspID: lspMaps[0].UserLspID,
			}
			res = append(res, tmp)
		}

		//check if map exists, if yes, check is active, if false - update to true
		queryStr := fmt.Sprintf(`SELECT * FROM userz.user_role WHERE user_id = '%s' AND user_lsp_id = '%s' and role = '%s' ALLOW FILTERING`, userID, *lspMaps[0].UserLspID, *role)
		getUserRole := func() (userRoles []userz.UserRole, err error) {
			q := CassUserSession.Query(queryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return userRoles, iter.Select(&userRoles)
		}
		resp, err := getUserRole()
		if err != nil {
			log.Printf("Got error while getting user roles: %v", err)
		}
		if len(resp) == 0 {
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
	}
	return res, nil
}

func GetVendors(ctx context.Context, lspID *string, filters *model.VendorFilters) ([]*model.Vendor, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
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
	session, err := global.CassPool.GetSession(ctx, "vendorz")
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
		go func(vendorId string, status string) {
			storageC := bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err = storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				log.Printf("Failed to get images of vendors: %v", err.Error())
				return
			}

			queryStr = fmt.Sprintf(`SELECT * FROM vendorz.vendor WHERE id = '%s' `, vendorId)
			if filters != nil {
				if filters.Status != nil {
					queryStr = queryStr + fmt.Sprintf(` and status='%s' `, *filters.Status)
				}
			}
			queryStr = queryStr + ` ALLOW FILTERING`
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
			var usersEmail []*string
			admins, err := GetVendorAdminsEmails(ctx, vendorId)
			if err != nil {
				log.Printf("Got error while getting vendor Admins for %v: %v", vendorId, err)
			}
			if len(admins) != 0 {
				for _, vv := range admins {
					v := vv
					usersEmail = append(usersEmail, &v)
				}
			}

			//photo
			photoUrl := ""
			if vendor.PhotoBucket != "" {
				photoUrl = storageC.GetSignedURLForObject(ctx, vendor.PhotoBucket)
			} else {
				photoUrl = vendor.PhotoUrl
			}
			createdAt := strconv.Itoa(int(vendor.CreatedAt))
			updatedAt := strconv.Itoa(int(vendor.UpdatedAt))
			vendorData := &model.Vendor{
				VendorID:        vendor.VendorId,
				Type:            vendor.Type,
				Level:           vendor.Level,
				Name:            vendor.Name,
				PhotoURL:        &photoUrl,
				Description:     &vendor.Description,
				Website:         &vendor.Website,
				Address:         &vendor.Address,
				Users:           usersEmail,
				FacebookURL:     &vendor.Facebook,
				InstagramURL:    &vendor.Instagram,
				TwitterURL:      &vendor.Twitter,
				LinkedinURL:     &vendor.LinkedIn,
				CreatedAt:       &createdAt,
				CreatedBy:       &vendor.CreatedBy,
				UpdatedAt:       &updatedAt,
				UpdatedBy:       &vendor.UpdatedBy,
				Status:          &vendor.Status,
				VendorLspStatus: &status,
			}
			res = append(res, vendorData)
			wg.Done()
		}(v.VendorId, v.Status)
	}
	wg.Wait()
	return res, nil
}

func GetVendorAdminsEmails(ctx context.Context, vendorID string) ([]string, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var userIds []vendorz.VendorUserMap
	session, err := global.CassPool.GetSession(ctx, "vendorz")
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
	res := make([]string, len(userIds))

	for k, vv := range userIds {
		v := vv
		userId := v.UserId
		email, err := base64.URLEncoding.DecodeString(userId)
		if err != nil {
			return nil, err
		}

		if !IsEmailValid(string(email)) || string(email) == "" {
			return nil, err
		}
		res[k] = string(email)
	}

	return res, nil
}

func GetVendorAdmins(ctx context.Context, vendorID string) ([]*model.UserWithLspStatus, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	lspId := claims["lsp_id"].(string)

	var userIds []vendorz.VendorUserMap
	session, err := global.CassPool.GetSession(ctx, "vendorz")
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
	res := make([]*model.UserWithLspStatus, len(userIds))

	var wg sync.WaitGroup
	for kk, vvv := range userIds {
		vv := vvv
		wg.Add(1)
		//iterate over these userIds and return user details
		go func(userId string, k int, lsp string) {
			//return user data

			email, err := base64.URLEncoding.DecodeString(userId)
			if err != nil {
				return
			}

			if !IsEmailValid(string(email)) || string(email) == "" {
				return
			}

			usersession, err := global.CassPool.GetSession(ctx, "userz")
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
			//
			createdAt := strconv.Itoa(int(user.CreatedAt))
			updatedAt := strconv.Itoa(int(user.UpdatedAt))
			fireBaseUser, err := global.IDP.GetUserByEmail(ctx, user.Email)
			if err != nil {
				log.Printf("Failed to get user from firebase: %v", err.Error())
				return
			}
			phone := ""
			if fireBaseUser != nil {
				phone = fireBaseUser.PhoneNumber
			}

			var photoUrl string
			storageC := bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err = storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				log.Printf("Failed to upload image to course: %v", err.Error())
				return
			}
			if user.PhotoBucket != "" {
				photoUrl = storageC.GetSignedURLForObject(ctx, user.PhotoBucket)
			} else {
				photoUrl = user.PhotoURL
			}

			temp := &model.UserWithLspStatus{
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
				PhotoURL:   &photoUrl,
				Phone:      phone,
			}

			qry := fmt.Sprintf(`SELECT * FROM userz.user_lsp_map WHERE user_id='%s' AND lsp_id='%s' ALLOW FILTERING`, user.ID, lsp)
			getUserLsp := func() (userlsps []userz.UserLsp, err error) {
				q := CassUserSession.Query(qry, nil)
				defer q.Release()
				iter := q.Iter()
				return userlsps, iter.Select(&userlsps)
			}
			userLsps, err := getUserLsp()
			if err != nil {
				return
			}
			if len(userLsps) == 0 {
				return
			}
			temp.UserLspStatus = &userLsps[0].Status

			userData := temp
			res[k] = userData

			wg.Done()
		}(vv.UserId, kk, lspId)
	}
	wg.Wait()
	return res, nil
}

func IsEmailValid(e string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(e)
}

func GetVendorDetails(ctx context.Context, vendorID string) (*model.Vendor, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := global.CassPool.GetSession(ctx, "vendorz")
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
	var usersEmail []*string
	admins, err := GetVendorAdminsEmails(ctx, vendorID)
	if err != nil {
		log.Printf("Got error while getting vendor Admins for %v: %v", vendorID, err)
	}
	if len(admins) != 0 {
		for _, vv := range admins {
			v := vv
			usersEmail = append(usersEmail, &v)
		}
	}
	//photo
	photoUrl := ""
	if vendor.PhotoBucket != "" {
		photoUrl = storageC.GetSignedURLForObject(ctx, vendor.PhotoBucket)
	} else {
		photoUrl = vendor.PhotoUrl
	}

	res := &model.Vendor{
		VendorID:     vendor.VendorId,
		Type:         vendor.Type,
		Level:        vendor.Level,
		Name:         vendor.Name,
		LspID:        &vendor.LspId,
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
		Status:       &vendor.Status,
	}
	return res, nil
}

func GetPaginatedVendors(ctx context.Context, lspID *string, pageCursor *string, direction *string, pageSize *int, filters *model.VendorFilters) (*model.PaginatedVendors, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting context: %v", err)
		return nil, err
	}
	lsp := claims["lsp_id"].(string)
	email := claims["email"].(string)
	if lspID != nil {
		lsp = *lspID
	}
	var newPage []byte

	session, err := global.CassPool.GetSession(ctx, "vendorz")
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

	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_lsp_map where lsp_id = '%s' `, lsp)
	if filters != nil && filters.Status != nil {
		queryStr += fmt.Sprintf(` AND status='%s'`, *filters.Status)
	}
	if filters != nil && filters.Service != nil {
		service := strings.ToLower(*filters.Service)
		queryStr += fmt.Sprintf(` AND services contains '%s' `, service)
	}
	if filters != nil && filters.Type != nil {
		vendorType := strings.ToLower(*filters.Type)
		queryStr += fmt.Sprintf(` AND type='%s'`, vendorType)
	}
	if filters != nil && filters.Name != nil {
		name := strings.ToLower(*filters.Name)
		nameArray := strings.Fields(name)
		for _, vv := range nameArray {
			v := vv
			queryStr += fmt.Sprintf(` AND words contains '%s' `, v)
		}
	}
	queryStr += `ALLOW FILTERING`
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
	for kk, vvv := range vendorIds {
		vv := vvv
		wg.Add(1)
		go func(vendorLspMap vendorz.VendorLspMap, k int) {

			session, err := global.CassPool.GetSession(ctx, "vendorz")
			if err != nil {
				return
			}
			CassSession := session

			vendorId := vendorLspMap.VendorId
			status := vendorLspMap.Status

			storageC := bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err = storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				log.Printf("Failed to get images of vendors: %v", err.Error())
				return
			}

			queryStr = fmt.Sprintf(`SELECT * FROM vendorz.vendor WHERE id = '%s' ALLOW FILTERING`, vendorId)
			getVendors := func() (vendors []vendorz.Vendor, err error) {
				q := CassSession.Query(queryStr, nil)
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
			var usersEmail []*string
			admins, err := GetVendorAdminsEmails(ctx, vendorId)
			if err != nil {
				log.Printf("Got error while getting vendor Admins for %v: %v", vendorId, err)
				return
			}
			if len(admins) != 0 {
				for _, vv := range admins {
					v := vv
					usersEmail = append(usersEmail, &v)
				}
			}

			createdAt := strconv.Itoa(int(vendor.CreatedAt))
			updatedAt := strconv.Itoa(int(vendor.UpdatedAt))
			var services []*string
			for _, vv := range vendorLspMap.Services {
				v := vv
				services = append(services, &v)
			}

			var photoUrl string
			if vendor.PhotoUrl != "" {
				photoUrl = storageC.GetSignedURLForObject(ctx, vendor.PhotoBucket)
			} else {
				photoUrl = vendor.PhotoUrl
			}

			vendorData := &model.Vendor{
				VendorID:        vendor.VendorId,
				Type:            vendor.Type,
				Level:           vendor.Level,
				Name:            vendor.Name,
				PhotoURL:        &photoUrl,
				LspID:           &vendor.LspId,
				Description:     &vendor.Description,
				Website:         &vendor.Website,
				Address:         &vendor.Address,
				Users:           usersEmail,
				FacebookURL:     &vendor.Facebook,
				InstagramURL:    &vendor.Instagram,
				TwitterURL:      &vendor.Twitter,
				LinkedinURL:     &vendor.LinkedIn,
				Services:        services,
				CreatedAt:       &createdAt,
				CreatedBy:       &vendor.CreatedBy,
				UpdatedAt:       &updatedAt,
				UpdatedBy:       &email,
				Status:          &vendor.Status,
				VendorLspStatus: &status,
			}
			res[k] = vendorData
			wg.Done()

		}(vv, kk)
	}
	wg.Wait()
	outputResponse.Vendors = res
	outputResponse.Direction = direction
	outputResponse.PageSize = pageSize
	outputResponse.PageCursor = &newCursor
	return &outputResponse, nil
}

func GetVendorExperience(ctx context.Context, vendorID string, pfID string) ([]*model.ExperienceVendor, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting claims: %v", err)
		return nil, err
	}

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Printf("Got error while creating session: %v", err)
		return nil, err
	}
	CassSession := session

	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.experience WHERE vendor_id = '%s' AND pf_id = '%s' ALLOW FILTERING`, vendorID, pfID)
	getProfile := func() (vendorExp []vendorz.VendorExperience, err error) {
		q := CassSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return vendorExp, iter.Select(&vendorExp)
	}

	vendorProfileExp, err := getProfile()
	if err != nil {
		log.Printf("Got error while getting experience for a profile: %v", err)
		return nil, err
	}
	if len(vendorProfileExp) == 0 {
		return nil, nil
	}

	res := make([]*model.ExperienceVendor, len(vendorProfileExp))
	for k, vv := range vendorProfileExp {
		v := vv
		endDate := int(v.EndDate)

		ca := strconv.Itoa(int(v.CreatedAt))
		ua := strconv.Itoa(int(v.UpdatedAt))
		sd := int(v.StartDate)
		tmp := &model.ExperienceVendor{
			ExpID:           v.ExpId,
			VendorID:        v.VendorId,
			PfID:            v.PfId,
			StartDate:       &sd,
			EndDate:         &endDate,
			Title:           &v.Title,
			EmployementType: &v.EmployementType,
			Location:        &v.Location,
			LocationType:    &v.LocationType,
			CompanyName:     &v.Company,
			CreatedAt:       &ca,
			CreatedBy:       &v.CreatedBy,
			UpdatedAt:       &ua,
			UpdatedBy:       &v.UpdatedBy,
			Status:          &v.Status,
		}
		res[k] = tmp
	}
	return res, nil
}

func GetVendorExperienceDetails(ctx context.Context, vendorID string, pfID string, expID string) (*model.ExperienceVendor, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting claims: %v", err)
		return nil, err
	}
	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.experience WHERE vendor_id = '%s' AND pf_id = '%s' AND exp_id = '%s' ALLOW FILTERING`, vendorID, pfID, expID)
	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Printf("Got error while getting session of vendor: %v", err)
	}
	CassSession := session
	getProfileExperience := func() (exp []vendorz.VendorExperience, err error) {
		q := CassSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return exp, iter.Select(&exp)
	}

	profileExperience, err := getProfileExperience()

	if err != nil {
		log.Printf("Got error while getting data from profile experience: %v", err)
	}
	if len(profileExperience) == 0 {
		return nil, nil
	}
	pfe := profileExperience[0]

	ed := int(pfe.EndDate)
	ca := strconv.Itoa(int(pfe.CreatedAt))
	ua := strconv.Itoa(int(pfe.UpdatedAt))
	sd := int(pfe.StartDate)

	res := &model.ExperienceVendor{
		ExpID:           pfe.ExpId,
		VendorID:        pfe.VendorId,
		PfID:            pfe.PfId,
		StartDate:       &sd,
		EndDate:         &ed,
		Title:           &pfe.Title,
		Location:        &pfe.Location,
		LocationType:    &pfe.LocationType,
		EmployementType: &pfe.EmployementType,
		CompanyName:     &pfe.Company,
		CreatedAt:       &ca,
		CreatedBy:       &pfe.CreatedBy,
		UpdatedAt:       &ua,
		UpdatedBy:       &pfe.UpdatedBy,
		Status:          &pfe.Status,
	}
	return res, nil
}

func UpdateExperienceVendor(ctx context.Context, input model.ExperienceInput) (*model.ExperienceVendor, error) {
	if input.VendorID == nil || input.ExpID == nil {
		return nil, errors.New("please pass all of the following fields, vendorId, email, expId")
	}
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting claims : %v", err)
		return nil, nil
	}
	updatedBy := claims["email"].(string)

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Printf("Got error while getting session: %v", err)
		return nil, err
	}
	CassSession := session

	email := strings.ToLower(input.Email)
	pfId := base64.URLEncoding.EncodeToString([]byte(email))
	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.experience WHERE vendor_id = '%s' AND pf_id = '%s' AND exp_id = '%s' ALLOW FILTERING`, *input.VendorID, pfId, *input.ExpID)

	getExperienceVendor := func() (exp []vendorz.VendorExperience, err error) {
		q := CassSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return exp, iter.Select(&exp)
	}

	exp, err := getExperienceVendor()
	if err != nil {
		log.Printf("Got error while getting experience of the vendor: %v", err)
		return nil, err
	}

	if len(exp) == 0 {
		return nil, fmt.Errorf("experience not found: %v", err)
	}
	experienceVendor := exp[0]
	updatedCols := []string{}
	//update all the given values
	if input.CompanyName != nil {
		updatedCols = append(updatedCols, "company")
		experienceVendor.Company = *input.CompanyName
	}
	if input.StartDate != nil {
		tmp := *input.StartDate
		updatedCols = append(updatedCols, "start_date")
		experienceVendor.StartDate = int64(tmp)
	}
	if input.EndDate != nil {
		tmp := *input.EndDate
		updatedCols = append(updatedCols, "end_date")
		experienceVendor.EndDate = int64(tmp)
	}
	if input.Title != nil {
		updatedCols = append(updatedCols, "title")
		experienceVendor.Title = *input.Title
	}
	if input.Location != nil {
		updatedCols = append(updatedCols, "location")
		experienceVendor.Location = *input.Location
	}
	if input.LocationType != nil {
		updatedCols = append(updatedCols, "location_type")
		experienceVendor.LocationType = *input.LocationType
	}
	if input.EmployementType != nil {
		updatedCols = append(updatedCols, "employement_type")
		experienceVendor.EmployementType = *input.EmployementType
	}
	if input.Status != nil {
		updatedCols = append(updatedCols, "status")
		experienceVendor.Status = *input.Status
	}
	updatedAt := time.Now().Unix()
	if len(updatedCols) > 0 {
		updatedCols = append(updatedCols, "updated_by")
		updatedCols = append(updatedCols, "updated_at")
		experienceVendor.UpdatedAt = updatedAt
		experienceVendor.UpdatedBy = updatedBy

		upStms, uNames := vendorz.VendorExperienceTable.Update(updatedCols...)
		updateQuery := CassSession.Query(upStms, uNames).BindStruct(&experienceVendor)

		if err = updateQuery.ExecRelease(); err != nil {
			log.Printf("Error updating experience of vendor's profile: %v", err)
		}
	}
	endDate := int(experienceVendor.EndDate)
	ca := strconv.Itoa(int(experienceVendor.CreatedAt))
	ua := strconv.Itoa(int(updatedAt))
	sd := int(experienceVendor.StartDate)

	res := model.ExperienceVendor{
		ExpID:           *input.ExpID,
		VendorID:        *input.VendorID,
		PfID:            pfId,
		StartDate:       &sd,
		EndDate:         &endDate,
		Title:           &experienceVendor.Title,
		Location:        &experienceVendor.Location,
		LocationType:    &experienceVendor.LocationType,
		EmployementType: &experienceVendor.EmployementType,
		CompanyName:     &experienceVendor.Company,
		CreatedAt:       &ca,
		CreatedBy:       &experienceVendor.CreatedBy,
		UpdatedAt:       &ua,
		UpdatedBy:       &updatedBy,
		Status:          &experienceVendor.Status,
	}

	return &res, nil
}

func ViewProfileVendorDetails(ctx context.Context, vendorID string, email string) (*model.VendorProfile, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting claims: %v", err)
		return nil, err
	}
	email = strings.ToLower(email)
	pfId := base64.URLEncoding.EncodeToString([]byte(email))
	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.profile WHERE pf_id = '%s' AND vendor_id = '%s' ALLOW FILTERING`, pfId, vendorID)
	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Printf("Got error while getting session of vendor: %v", err)
	}
	CassSession := session
	getProfileDetail := func() (exp []vendorz.VendorProfile, err error) {
		q := CassSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return exp, iter.Select(&exp)
	}

	profileDetails, err := getProfileDetail()

	if err != nil {
		log.Printf("Got error while getting data from profile experience: %v", err)
	}
	if len(profileDetails) == 0 {
		return nil, nil
	}
	profile := profileDetails[0]

	//get photo url
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		log.Printf("Failed to upload image to course: %v", err.Error())
		return nil, err
	}
	photoUrl := ""
	if profile.PhotoBucket != "" {
		photoUrl = storageC.GetSignedURLForObject(ctx, profile.PhotoBucket)
	} else {
		photoUrl = profile.PhotoURL
	}

	languages := ChangeToPointerArray(profile.Languages)
	sme := ChangeToPointerArray(profile.SMEExpertise)
	crt := ChangeToPointerArray(profile.ClassroomExpertise)
	exp := ChangeToPointerArray(profile.Experience)
	cd := ChangeToPointerArray(profile.ContentDevelopment)
	createdAt := strconv.Itoa(int(profile.CreatedAt))
	updatedAt := strconv.Itoa(int(profile.UpdatedAt))

	res := model.VendorProfile{
		PfID:               &pfId,
		VendorID:           &vendorID,
		FirstName:          &profile.FirstName,
		LastName:           &profile.LastName,
		Email:              &profile.Email,
		Phone:              &profile.Phone,
		PhotoURL:           &photoUrl,
		Description:        &profile.Description,
		Language:           languages,
		SmeExpertise:       sme,
		ClassroomExpertise: crt,
		ContentDevelopment: cd,
		Experience:         exp,
		ExperienceYears:    &profile.ExperienceYears,
		IsSpeaker:          &profile.IsSpeaker,
		CreatedAt:          &createdAt,
		CreatedBy:          &profile.CreatedBy,
		UpdatedAt:          &updatedAt,
		UpdatedBy:          &profile.UpdatedBy,
		Status:             &profile.Status,
	}
	return &res, nil
}

func ChangeToPointerArray(input []string) []*string {
	var res []*string
	if len(input) > 0 {
		for _, vv := range input {
			v := vv
			res = append(res, &v)
		}
	}
	return res
}

func ViewAllProfiles(ctx context.Context, vendorID string, filter *string, name *string) ([]*model.VendorProfile, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting claims: %v", err)
		return nil, err
	}
	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.profile WHERE vendor_id = '%s' `, vendorID)
	if filter != nil && *filter != "" {
		queryStr = queryStr + fmt.Sprintf(`AND '%s' = true `, *filter)
	}
	if name != nil && *name != "" {
		names := strings.ToLower(*name)
		namesArray := strings.Fields(names)
		for _, vv := range namesArray {
			v := vv
			queryStr += fmt.Sprintf(` AND name CONTAINS '%s' `, v)
		}
	}
	queryStr = queryStr + "ALLOW FILTERING"

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Printf("Got error while getting session of vendor: %v", err)
	}
	CassSession := session

	getAllProfiles := func() (profiles []vendorz.VendorProfile, err error) {
		q := CassSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return profiles, iter.Select(&profiles)
	}

	profiles, err := getAllProfiles()
	if err != nil {
		log.Println("Got error while getting profiles ", err)
		return nil, err
	}
	if len(profiles) == 0 {
		return nil, nil
	}

	res := make([]*model.VendorProfile, len(profiles))
	var wg sync.WaitGroup
	for kk, vvv := range profiles {
		vv := vvv

		wg.Add(1)
		//get photo url
		go func(k int, v vendorz.VendorProfile) {
			storageC := bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err = storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				log.Printf("Failed to view all profiles: %v", err.Error())
				return
			}
			photoUrl := ""
			if v.PhotoBucket != "" {
				photoUrl = storageC.GetSignedURLForObject(ctx, v.PhotoBucket)
			} else {
				photoUrl = v.PhotoURL
			}

			languages := ChangeToPointerArray(v.Languages)
			sme := ChangeToPointerArray(v.SMEExpertise)
			crt := ChangeToPointerArray(v.ClassroomExpertise)
			exp := ChangeToPointerArray(v.Experience)
			cd := ChangeToPointerArray(v.ContentDevelopment)
			createdAt := strconv.Itoa(int(v.CreatedAt))
			updatedAt := strconv.Itoa(int(v.UpdatedAt))
			tmp := model.VendorProfile{
				PfID:               &v.PfId,
				VendorID:           &v.VendorId,
				FirstName:          &v.FirstName,
				LastName:           &v.LastName,
				Email:              &v.Email,
				Phone:              &v.Phone,
				PhotoURL:           &photoUrl,
				Description:        &v.Description,
				Language:           languages,
				SmeExpertise:       sme,
				ClassroomExpertise: crt,
				ContentDevelopment: cd,
				Experience:         exp,
				ExperienceYears:    &v.ExperienceYears,
				IsSpeaker:          &v.IsSpeaker,
				CreatedAt:          &createdAt,
				CreatedBy:          &v.CreatedBy,
				UpdatedAt:          &updatedAt,
				UpdatedBy:          &v.UpdatedBy,
				Status:             &v.Status,
			}
			res[k] = &tmp
			wg.Done()
		}(kk, vv)
	}
	wg.Wait()
	return res, nil
}

// status='active' in all vendors
func UpdateProfileVendor(ctx context.Context, input *model.VendorProfileInput) (*model.VendorProfile, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting claims: %v", err)
		return nil, err
	}
	email := claims["email"].(string)

	mail := strings.ToLower(input.Email)
	pfId := base64.URLEncoding.EncodeToString([]byte(mail))
	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.profile WHERE pf_id = '%s' AND vendor_id = '%s' ALLOW FILTERING`, pfId, input.VendorID)
	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Printf("Got error while getting session of vendor: %v", err)
	}
	CassSession := session

	getProfileDetails := func() (profiles []vendorz.VendorProfile, err error) {
		q := CassSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return profiles, iter.Select(&profiles)
	}
	profiles, err := getProfileDetails()
	if err != nil {
		log.Printf("Got error while getting profile data: %v", err)
		return nil, err
	}
	if len(profiles) == 0 {
		return nil, nil
	}
	profile := profiles[0]
	updatedCols := []string{}

	var crt bool
	if input.ClassroomExpertise != nil {
		tmp := ChangesStringType(input.ClassroomExpertise)
		profile.ClassroomExpertise = tmp
		updatedCols = append(updatedCols, "classroom_expertise")

		crt = len(profile.ClassroomExpertise) > 0
		profile.Crt = crt
		updatedCols = append(updatedCols, "crt")
	}
	if input.Description != nil {
		profile.Description = *input.Description
		updatedCols = append(updatedCols, "description")
	}
	if input.FirstName != nil {
		profile.FirstName = *input.FirstName
		updatedCols = append(updatedCols, "first_name")
	}
	if input.IsSpeaker != nil {
		profile.IsSpeaker = *input.IsSpeaker
		updatedCols = append(updatedCols, "is_speaker")
	}
	if input.Languages != nil {
		tmp := ChangesStringType(input.Languages)
		profile.Languages = tmp
		updatedCols = append(updatedCols, "languages")
	}
	if input.LastName != nil {
		profile.LastName = *input.LastName
		updatedCols = append(updatedCols, "last_name")
	}
	if input.Phone != nil {
		profile.Phone = *input.Phone
		updatedCols = append(updatedCols, "phone_number")
	}
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		return nil, err
	}
	if input.Photo != nil {
		bucketPath := fmt.Sprintf("%s/%s/%s/%s", "vendor", "profile", pfId, base64.URLEncoding.EncodeToString([]byte(input.Photo.Filename)))
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
		url := storageC.GetSignedURLForObject(ctx, bucketPath)
		profile.PhotoBucket = bucketPath
		profile.PhotoURL = url
		updatedCols = append(updatedCols, "photo_bucket", "photo_url")
	}
	var sme bool
	if input.SmeExpertise != nil {
		tmp := ChangesStringType(input.SmeExpertise)
		profile.SMEExpertise = tmp
		updatedCols = append(updatedCols, "sme_expertise")

		sme = len(profile.SMEExpertise) > 0
		profile.Sme = sme
		updatedCols = append(updatedCols, "sme")
	}
	var cd bool
	if input.ContentDevelopment != nil {
		tmp := ChangesStringType(input.ContentDevelopment)
		profile.ContentDevelopment = tmp
		updatedCols = append(updatedCols, "content_development")

		cd = len(profile.ContentDevelopment) > 0
		profile.Cd = cd
		updatedCols = append(updatedCols, "cd")
	}
	if input.Status != nil {
		profile.Status = *input.Status
		updatedCols = append(updatedCols, "status")
	}
	if input.ExperienceYears != nil {
		profile.ExperienceYears = *input.ExperienceYears
		updatedCols = append(updatedCols, "experience_years")
	}
	if input.Experience != nil {
		tmp := ChangesStringType(input.Experience)
		profile.Experience = tmp
		updatedCols = append(updatedCols, "experience")
	}

	if len(updatedCols) > 0 {
		profile.UpdatedBy = email
		profile.UpdatedAt = time.Now().Unix()
		updatedCols = append(updatedCols, "updated_by", "updated_at")
		upStms, uNames := vendorz.VendorProfileTable.Update(updatedCols...)
		updateQuery := CassSession.Query(upStms, uNames).BindStruct(&profile)
		if err = updateQuery.ExecRelease(); err != nil {
			log.Printf("Error updating profile: %v", err)
			return nil, err
		}
	}
	lang := ChangeToPointerArray(profile.Languages)
	smeExpertise := ChangeToPointerArray(profile.SMEExpertise)
	crtExpertise := ChangeToPointerArray(profile.ClassroomExpertise)
	exp := ChangeToPointerArray(profile.Experience)
	cdExpertise := ChangeToPointerArray(profile.ContentDevelopment)
	ca := strconv.Itoa(int(profile.CreatedAt))
	ua := strconv.Itoa(int(profile.UpdatedAt))

	res := model.VendorProfile{
		PfID:               &profile.PfId,
		VendorID:           &profile.VendorId,
		FirstName:          &profile.FirstName,
		LastName:           &profile.LastName,
		Email:              &profile.Email,
		Phone:              &profile.Phone,
		PhotoURL:           &profile.PhotoURL,
		Description:        &profile.Description,
		Language:           lang,
		SmeExpertise:       smeExpertise,
		ClassroomExpertise: crtExpertise,
		ContentDevelopment: cdExpertise,
		Sme:                &sme,
		Crt:                &crt,
		Cd:                 &cd,
		Experience:         exp,
		ExperienceYears:    &profile.ExperienceYears,
		IsSpeaker:          &profile.IsSpeaker,
		CreatedAt:          &ca,
		CreatedBy:          &profile.CreatedBy,
		UpdatedAt:          &ua,
		UpdatedBy:          &email,
		Status:             &profile.Status,
	}
	return &res, nil
}

func UploadSampleFile(ctx context.Context, input *model.SampleFileInput) (*model.SampleFile, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error in getting claims: %v", err)
	}
	email := claims["email"].(string)
	log.Println("Upload Sample File called")

	res := model.SampleFile{}
	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Printf("Got error while getting session of vendor: %v", err)
	}
	CassSession := session

	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		log.Printf("Failed to upload sample file: %v", err.Error())
		return &res, err
	}
	bucketPath := fmt.Sprintf("%s/%s/%s/%s", "vendor", input.VendorID, input.PType, input.Name)
	// writer, err := storageC.UploadToGCS(ctx, bucketPath)
	// if err != nil {
	// 	log.Printf("Failed to upload sample file: %v", err.Error())
	// 	return &res, nil
	// }
	// defer writer.Close()
	// fileBuffer := bytes.NewBuffer(nil)
	// if _, err := io.Copy(fileBuffer, input.File.File); err != nil {
	// 	return &res, nil
	// }
	// currentBytes := fileBuffer.Bytes()
	// _, err = io.Copy(writer, bytes.NewReader(currentBytes))
	// if err != nil {
	// 	return &res, nil
	// }

	utils.SenUploadRequestToQueue(ctx, &input.File, bucketPath)
	getUrl := storageC.GetSignedURLForObject(ctx, bucketPath)
	if getUrl == "" {
		return &res, fmt.Errorf("failed to upload sample file: %v", errors.New("failed to get URL"))
	}
	sfId := uuid.New().String()
	ca := time.Now().Unix()

	file := vendorz.SampleFile{
		SfId:           sfId,
		Name:           input.Name,
		Pricing:        input.Pricing,
		FileBucket:     bucketPath,
		VendorId:       input.VendorID,
		PType:          input.PType,
		ActualFileType: input.File.ContentType,
		CreatedAt:      ca,
		CreatedBy:      email,
		UpdatedAt:      ca,
		UpdatedBy:      email,
	}

	createdAt := strconv.Itoa(int(ca))
	res.SfID = sfId
	res.Name = &input.Name
	res.Price = &input.Pricing
	res.CreatedAt = &createdAt
	res.CreatedBy = &email
	res.UpdatedAt = &createdAt
	res.UpdatedBy = &email
	res.FileURL = &getUrl
	res.PType = &input.PType
	res.ActualFileType = &input.File.ContentType

	if input.Description != nil {
		file.Description = *input.Description
		res.Description = input.Description
	}
	if input.FileType != nil {
		file.FileType = *input.FileType
		res.FileType = input.FileType
	}
	if getUrl != "" {
		file.FileUrl = getUrl
	}
	if input.Status != nil {
		file.Status = *input.Status
		res.Status = input.Status
	}
	if input.Rate != nil {
		file.Rate = int64(*input.Rate)
		res.Rate = input.Rate
	}
	if input.Currency != nil {
		file.Currency = *input.Currency
		res.Currency = input.Currency
	}
	if input.Unit != nil {
		file.Unit = *input.Unit
		res.Unit = input.Unit
	}
	insertQueryMap := CassSession.Query(vendorz.SampleFileTable.Insert()).BindStruct(file)
	if err = insertQueryMap.Exec(); err != nil {
		return nil, err
	}

	return &res, nil
}

func GetSampleFiles(ctx context.Context, vendorID string, pType string) ([]*model.SampleFile, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Printf("Got error while getting claims of the user: %v", err)
		return nil, err
	}

	var res []*model.SampleFile

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Printf("Got error while getting session: %v", err)
		return nil, err
	}
	CassSession := session
	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.sample_file WHERE vendor_id = '%s' AND p_type = '%s' ALLOW FILTERING`, vendorID, pType)
	getFiles := func() (files []vendorz.SampleFile, err error) {
		q := CassSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return files, iter.Select(&files)
	}
	files, err := getFiles()
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}

	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		log.Printf("Failed to view sample files of course: %v", err.Error())
		return nil, err
	}
	for _, vv := range files {
		v := vv
		photoUrl := ""
		createdAt := strconv.Itoa(int(v.CreatedAt))
		updatedAt := strconv.Itoa(int(v.UpdatedAt))
		rate := int(v.Rate)
		//just map these to model.sample-files and return
		file := model.SampleFile{
			SfID:           v.SfId,
			Name:           &v.Name,
			FileType:       &v.FileType,
			Price:          &v.Pricing,
			CreatedAt:      &createdAt,
			CreatedBy:      &v.CreatedBy,
			UpdatedAt:      &updatedAt,
			UpdatedBy:      &v.UpdatedBy,
			Status:         &v.Status,
			Description:    &v.Description,
			Rate:           &rate,
			Currency:       &v.Currency,
			Unit:           &v.Unit,
			ActualFileType: &v.ActualFileType,
		}
		if v.FileBucket != "" {
			photoUrl = storageC.GetSignedURLForObject(ctx, v.FileBucket)
		} else {
			photoUrl = v.FileUrl
		}
		file.FileURL = &photoUrl

		res = append(res, &file)
	}
	return res, nil
}

func CreateSubjectMatterExpertise(ctx context.Context, input *model.SMEInput) (*model.Sme, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Errorf("Got error while getting claims: %v", err)
	}
	email := claims["email"].(string)
	lsp := claims["lsp_id"].(string)

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Errorf("Error while getting session: %v", err)
	}
	CassSession := session
	smeId := uuid.New().String()

	ca := time.Now().Unix()
	sme := vendorz.SME{
		SMEId:     smeId,
		VendorId:  input.VendorID,
		CreatedAt: ca,
		CreatedBy: email,
		UpdatedAt: ca,
		UpdatedBy: email,
	}
	if input.IsApplicable != nil {
		sme.IsApplicable = *input.IsApplicable
	}
	if input.Description != nil {
		sme.Description = *input.Description
	}
	if input.Expertise != nil {
		tmp := ChangesStringType(input.Expertise)
		sme.Expertise = tmp
	}
	if input.Languages != nil {
		tmp := ChangesStringType(input.Languages)
		sme.Languages = tmp
	}
	if input.OutputDeliveries != nil {
		tmp := ChangesStringType(input.OutputDeliveries)
		sme.OutputDeliveries = tmp
	}
	if input.SampleFiles != nil {
		tmp := ChangesStringType(input.SampleFiles)
		sme.SampleFiles = tmp
	}
	if input.Status != nil {
		sme.Status = *input.Status
	}
	if input.IsExpertiseOffline != nil {
		sme.IsExpertiseOffline = *input.IsExpertiseOffline
	}
	if input.IsExpertiseOnline != nil {
		sme.IsExpertiseOnline = *input.IsExpertiseOnline
	}
	insertQuery := CassSession.Query(vendorz.SMETable.Insert()).BindStruct(sme)
	if err = insertQuery.Exec(); err != nil {
		return nil, err
	}

	err = updateVendorLspMap(ctx, input.VendorID, lsp, "sme", true)
	if err != nil {
		return nil, err
	}

	createdAt := strconv.Itoa(int(ca))
	res := model.Sme{
		VendorID:           &input.VendorID,
		SmeID:              &smeId,
		Description:        input.Description,
		IsApplicable:       input.IsApplicable,
		Expertise:          input.Expertise,
		Languages:          input.Languages,
		OutputDeliveries:   input.OutputDeliveries,
		SampleFiles:        input.SampleFiles,
		IsExpertiseOnline:  input.IsExpertiseOnline,
		IsExpertiseOffline: input.IsExpertiseOffline,
		CreatedAt:          &createdAt,
		CreatedBy:          &email,
		UpdatedAt:          &createdAt,
		UpdatedBy:          &email,
		Status:             input.Status,
	}

	return &res, nil
}

func UpdateSubjectMatterExpertise(ctx context.Context, input *model.SMEInput) (*model.Sme, error) {
	if input.VendorID == "" || input.SmeID == nil {
		log.Errorf("Please pass both vendor id and sme Id")
	}
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Errorf("Got error while getting claims: %v", err)
	}
	email := claims["email"].(string)
	lsp := claims["lsp_id"].(string)

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Errorf("Error while getting session: %v", err)
	}
	CassSession := session

	//vendor_id, sme_id
	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.sme WHERE vendor_id = '%s' AND sme_id = '%s' ALLOW FILTERING`, input.VendorID, *input.SmeID)
	getSmeData := func() (smeData []vendorz.SME, err error) {
		q := CassSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return smeData, iter.Select(&smeData)
	}
	smeDatas, err := getSmeData()
	if err != nil {
		return nil, err
	}
	if len(smeDatas) == 0 {
		return nil, nil
	}

	smeData := smeDatas[0]
	updatedCols := []string{}

	if input.Description != nil {
		smeData.Description = *input.Description
		updatedCols = append(updatedCols, "description")
	}
	if input.Expertise != nil {
		tmp := ChangesStringType(input.Expertise)
		smeData.Expertise = tmp
		updatedCols = append(updatedCols, "expertise")
	}
	if input.IsApplicable != nil {

		//if input.IsApplicable is set to true, then add, otherwise delete
		err = updateVendorLspMap(ctx, input.VendorID, lsp, "sme", *input.IsApplicable)
		if err != nil {
			return nil, err
		}

		smeData.IsApplicable = *input.IsApplicable
		updatedCols = append(updatedCols, "is_applicable")
	}
	if input.Languages != nil {
		tmp := ChangesStringType(input.Languages)
		smeData.Languages = tmp
		updatedCols = append(updatedCols, "languages")
	}
	if input.OutputDeliveries != nil {
		tmp := ChangesStringType(input.OutputDeliveries)
		smeData.OutputDeliveries = tmp
		updatedCols = append(updatedCols, "output_deliveries")
	}
	if input.SampleFiles != nil {
		tmp := ChangesStringType(input.SampleFiles)
		smeData.SampleFiles = tmp
		updatedCols = append(updatedCols, "sample_files")
	}
	if input.Status != nil {
		smeData.Status = *input.Status
		updatedCols = append(updatedCols, "status")
	}
	if input.IsExpertiseOffline != nil {
		smeData.IsExpertiseOffline = *input.IsExpertiseOffline
		updatedCols = append(updatedCols, "is_expertise_offline")
	}
	if input.IsExpertiseOnline != nil {
		smeData.IsExpertiseOnline = *input.IsExpertiseOnline
		updatedCols = append(updatedCols, "is_expertise_online")
	}
	ua := time.Now().Unix()
	if len(updatedCols) > 0 {
		smeData.UpdatedAt = ua
		smeData.UpdatedBy = email
		updatedCols = append(updatedCols, "updated_at", "updated_by")

		utStms, uNames := vendorz.SMETable.Update(updatedCols...)
		updateQuery := CassSession.Query(utStms, uNames).BindStruct(&smeData)
		if err = updateQuery.ExecRelease(); err != nil {
			log.Errorf("Error while updating SME")
			return nil, err
		}
	}

	expertise := ChangeToPointerArray(smeData.Expertise)
	lan := ChangeToPointerArray(smeData.Languages)
	od := ChangeToPointerArray(smeData.OutputDeliveries)
	sf := ChangeToPointerArray(smeData.SampleFiles)
	ca := strconv.Itoa(int(smeData.CreatedAt))
	updatedAt := strconv.Itoa(int(ua))
	res := model.Sme{
		VendorID:           &input.VendorID,
		SmeID:              input.SmeID,
		Description:        &smeData.Description,
		IsApplicable:       &smeData.IsApplicable,
		Expertise:          expertise,
		Languages:          lan,
		OutputDeliveries:   od,
		SampleFiles:        sf,
		IsExpertiseOnline:  &smeData.IsExpertiseOnline,
		IsExpertiseOffline: &smeData.IsExpertiseOffline,
		CreatedAt:          &ca,
		CreatedBy:          &smeData.CreatedBy,
		UpdatedAt:          &updatedAt,
		UpdatedBy:          &smeData.UpdatedBy,
		Status:             &smeData.Status,
	}

	return &res, nil
}

func GetSmeDetails(ctx context.Context, vendorID string) (*model.Sme, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Errorf("Got error while getting claims: %v", err)
	}

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Errorf("Error while getting session: %v", err)
	}
	CassSession := session

	//vendor_id, sme_id
	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.sme WHERE vendor_id = '%s' ALLOW FILTERING`, vendorID)
	getSmeData := func() (smeData []vendorz.SME, err error) {
		q := CassSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return smeData, iter.Select(&smeData)
	}
	smeDatas, err := getSmeData()
	if err != nil {
		return nil, err
	}
	if len(smeDatas) == 0 {
		return nil, nil
	}

	smeData := smeDatas[0]
	expertise := ChangeToPointerArray(smeData.Expertise)
	lan := ChangeToPointerArray(smeData.Languages)
	od := ChangeToPointerArray(smeData.OutputDeliveries)
	sf := ChangeToPointerArray(smeData.SampleFiles)
	ca := strconv.Itoa(int(smeData.CreatedAt))
	ua := strconv.Itoa(int(smeData.UpdatedAt))
	res := model.Sme{
		VendorID:           &vendorID,
		SmeID:              &smeData.SMEId,
		Description:        &smeData.Description,
		IsApplicable:       &smeData.IsApplicable,
		Expertise:          expertise,
		Languages:          lan,
		OutputDeliveries:   od,
		SampleFiles:        sf,
		IsExpertiseOnline:  &smeData.IsExpertiseOnline,
		IsExpertiseOffline: &smeData.IsExpertiseOffline,
		CreatedAt:          &ca,
		CreatedBy:          &smeData.CreatedBy,
		UpdatedAt:          &ua,
		UpdatedBy:          &smeData.UpdatedBy,
		Status:             &smeData.Status,
	}
	return &res, nil
}

func CreateClassRoomTraining(ctx context.Context, input *model.CRTInput) (*model.Crt, error) {

	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Errorf("Got error while getting claims: %v", err)
	}
	email := claims["email"].(string)
	lsp := claims["lsp_id"].(string)

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Errorf("Error while getting session: %v", err)
	}
	CassSession := session

	crtId := uuid.New().String()
	createdAt := time.Now().Unix()
	crt := vendorz.CRT{
		CtId:      crtId,
		VendorId:  input.VendorID,
		CreatedAt: createdAt,
		CreatedBy: email,
		UpdatedAt: createdAt,
		UpdatedBy: email,
	}
	if input.Description != nil {
		crt.Description = *input.Description
	}
	if input.Expertise != nil {
		tmp := ChangesStringType(input.Expertise)
		crt.Expertise = tmp
	}
	if input.IsApplicable != nil {
		crt.IsApplicable = *input.IsApplicable
	}
	if input.Languages != nil {
		tmp := ChangesStringType(input.Languages)
		crt.Languages = tmp
	}
	if input.OutputDeliveries != nil {
		tmp := ChangesStringType(input.OutputDeliveries)
		crt.OutputDeliveries = tmp
	}
	if input.SampleFiles != nil {
		tmp := ChangesStringType(input.SampleFiles)
		crt.SampleFiles = tmp
	}
	if input.IsExpertiseOnline != nil {
		crt.IsExpertiseOnline = *input.IsExpertiseOnline
	}
	if input.Status != nil {
		crt.Status = *input.Status
	}
	if input.IsExpertiseOffline != nil {
		crt.IsExpertiseOffline = *input.IsExpertiseOffline
	}
	insertQuery := CassSession.Query(vendorz.ClassRoomTrainingTable.Insert()).BindStruct(crt)
	if err = insertQuery.Exec(); err != nil {
		return nil, err
	}

	err = updateVendorLspMap(ctx, input.VendorID, lsp, "crt", true)
	if err != nil {
		return nil, err
	}

	ca := strconv.Itoa(int(createdAt))
	res := model.Crt{
		CrtID:              &crtId,
		VendorID:           input.VendorID,
		Description:        input.Description,
		IsApplicable:       input.IsApplicable,
		Expertise:          input.Expertise,
		Languages:          input.Languages,
		SampleFiles:        input.SampleFiles,
		OutputDeliveries:   input.OutputDeliveries,
		IsExpertiseOnline:  input.IsExpertiseOnline,
		IsExpertiseOffline: input.IsExpertiseOffline,
		CreatedAt:          &ca,
		CreatedBy:          &email,
		UpdatedAt:          &ca,
		UpdatedBy:          &email,
		Status:             input.Status,
	}

	return &res, nil
}

func UpdateClassRoomTraining(ctx context.Context, input *model.CRTInput) (*model.Crt, error) {
	if input.VendorID == "" || input.CrtID == nil {
		return nil, fmt.Errorf("please provide both vendorId and CrtId")
	}

	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Errorf("Got error while getting claims: %v", err)
	}
	email := claims["email"].(string)
	lsp := claims["lsp_id"].(string)

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Errorf("Error while getting session: %v", err)
	}
	CassSession := session

	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.classroom_training WHERE vendor_id = '%s' ALLOW FILTERING`, input.VendorID)
	getCRT := func() (crtData []vendorz.CRT, err error) {
		q := CassSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return crtData, iter.Select(&crtData)
	}
	crtDatas, err := getCRT()
	if err != nil {
		return nil, err
	}
	if len(crtDatas) == 0 {
		return nil, nil
	}

	crt := crtDatas[0]
	updatedCols := []string{}

	if input.Description != nil {
		crt.Description = *input.Description
		updatedCols = append(updatedCols, "description")
	}
	if input.Expertise != nil {
		tmp := ChangesStringType(input.Expertise)
		crt.Expertise = tmp
		updatedCols = append(updatedCols, "expertise")
	}
	if input.IsApplicable != nil {

		err = updateVendorLspMap(ctx, input.VendorID, lsp, "crt", *input.IsApplicable)
		if err != nil {
			return nil, err
		}

		crt.IsApplicable = *input.IsApplicable
		updatedCols = append(updatedCols, "is_applicable")
	}
	if input.IsExpertiseOnline != nil {
		crt.IsExpertiseOnline = *input.IsExpertiseOnline
		updatedCols = append(updatedCols, "is_expertise_online")
	}
	if input.Languages != nil {
		tmp := ChangesStringType(input.Languages)
		crt.Languages = tmp
		updatedCols = append(updatedCols, "languages")
	}
	if input.OutputDeliveries != nil {
		tmp := ChangesStringType(input.OutputDeliveries)
		crt.OutputDeliveries = tmp
		updatedCols = append(updatedCols, "output_deliveries")
	}
	if input.SampleFiles != nil {
		tmp := ChangesStringType(input.SampleFiles)
		crt.SampleFiles = tmp
		updatedCols = append(updatedCols, "sample_files")
	}
	if input.Status != nil {
		crt.Status = *input.Status
		updatedCols = append(updatedCols, "status")
	}
	if input.IsExpertiseOffline != nil {
		crt.IsExpertiseOffline = *input.IsExpertiseOffline
		updatedCols = append(updatedCols, "is_expertise_offline")
	}
	updatedAt := time.Now().Unix()
	if len(updatedCols) > 0 {
		crt.UpdatedBy = email
		crt.UpdatedAt = updatedAt
		updatedCols = append(updatedCols, "updated_at", "updated_by")

		utStms, uNames := vendorz.ClassRoomTrainingTable.Update(updatedCols...)
		updateQuery := CassSession.Query(utStms, uNames).BindStruct(&crt)
		if err = updateQuery.ExecRelease(); err != nil {
			log.Errorf("Error while updating CRT")
			return nil, err
		}
	}

	exp := ChangeToPointerArray(crt.Expertise)
	lan := ChangeToPointerArray(crt.Languages)
	od := ChangeToPointerArray(crt.OutputDeliveries)
	sf := ChangeToPointerArray(crt.SampleFiles)
	ca := strconv.Itoa(int(crt.CreatedAt))
	ua := strconv.Itoa(int(crt.UpdatedAt))
	res := model.Crt{
		CrtID:              &crt.CtId,
		VendorID:           crt.VendorId,
		Description:        &crt.Description,
		IsApplicable:       &crt.IsApplicable,
		Expertise:          exp,
		Languages:          lan,
		OutputDeliveries:   od,
		SampleFiles:        sf,
		IsExpertiseOnline:  &crt.IsExpertiseOnline,
		IsExpertiseOffline: &crt.IsExpertiseOffline,
		CreatedAt:          &ca,
		CreatedBy:          &crt.CreatedBy,
		UpdatedAt:          &ua,
		UpdatedBy:          &email,
		Status:             &crt.Status,
	}
	return &res, nil
}

func GetClassRoomTraining(ctx context.Context, vendorID string) (*model.Crt, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Errorf("Got error while getting claims: %v", err)
	}

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Errorf("Error while getting session: %v", err)
	}
	CassSession := session

	//vendor_id, sme_id
	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.classroom_training WHERE vendor_id = '%s' ALLOW FILTERING`, vendorID)
	getCrtData := func() (crtData []vendorz.CRT, err error) {
		q := CassSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return crtData, iter.Select(&crtData)
	}
	crtDatas, err := getCrtData()
	if err != nil {
		return nil, err
	}
	if len(crtDatas) == 0 {
		return nil, nil
	}

	crt := crtDatas[0]

	exp := ChangeToPointerArray(crt.Expertise)
	lan := ChangeToPointerArray(crt.Languages)
	od := ChangeToPointerArray(crt.OutputDeliveries)
	sf := ChangeToPointerArray(crt.SampleFiles)
	ca := strconv.Itoa(int(crt.CreatedAt))
	ua := strconv.Itoa(int(crt.UpdatedAt))
	res := model.Crt{
		CrtID:              &crt.CtId,
		VendorID:           crt.VendorId,
		Description:        &crt.Description,
		IsApplicable:       &crt.IsApplicable,
		Expertise:          exp,
		Languages:          lan,
		OutputDeliveries:   od,
		SampleFiles:        sf,
		IsExpertiseOnline:  &crt.IsExpertiseOnline,
		IsExpertiseOffline: &crt.IsExpertiseOffline,
		CreatedAt:          &ca,
		CreatedBy:          &crt.CreatedBy,
		UpdatedAt:          &ua,
		UpdatedBy:          &crt.UpdatedBy,
		Status:             &crt.Status,
	}
	return &res, nil
}

func CreateContentDevelopment(ctx context.Context, input *model.ContentDevelopmentInput) (*model.ContentDevelopment, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Errorf("Got error while getting claims: %v", err)
	}
	email := claims["email"].(string)
	lsp := claims["lsp_id"].(string)

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Errorf("Error while getting session: %v", err)
	}
	CassSession := session

	cdId := uuid.New().String()
	createdAt := time.Now().Unix()

	cd := vendorz.ContentDevelopment{
		CdId:      cdId,
		VendorId:  input.VendorID,
		CreatedAt: createdAt,
		CreatedBy: email,
		UpdatedAt: createdAt,
		UpdatedBy: email,
	}
	if input.Description != nil {
		cd.Description = *input.Description
	}
	if input.Expertise != nil {
		tmp := ChangesStringType(input.Expertise)
		cd.Expertise = tmp
	}
	if input.IsApplicable != nil {
		cd.IsApplicable = *input.IsApplicable
	}
	if input.Languages != nil {
		tmp := ChangesStringType(input.Languages)
		cd.Languages = tmp
	}
	if input.OutputDeliveries != nil {
		tmp := ChangesStringType(input.OutputDeliveries)
		cd.OutputDeliveries = tmp
	}
	if input.SampleFiles != nil {
		tmp := ChangesStringType(input.SampleFiles)
		cd.SampleFiles = tmp
	}
	if input.Status != nil {
		cd.Status = *input.Status
	}
	if input.IsExpertiseOffline != nil {
		cd.IsExpertiseOffline = *input.IsExpertiseOffline
	}
	if input.IsExpertiseOnline != nil {
		cd.IsExpertiseOnline = *input.IsExpertiseOnline
	}
	insertQuery := CassSession.Query(vendorz.ContentDevelopmentTable.Insert()).BindStruct(cd)
	if err = insertQuery.Exec(); err != nil {
		return nil, err
	}

	err = updateVendorLspMap(ctx, input.VendorID, lsp, "cd", true)
	if err != nil {
		return nil, err
	}

	ca := strconv.Itoa(int(createdAt))
	res := model.ContentDevelopment{
		CdID:               &cdId,
		VendorID:           &cd.VendorId,
		Description:        &cd.Description,
		IsApplicable:       &cd.IsApplicable,
		Expertise:          input.Expertise,
		Languages:          input.Languages,
		OutputDeliveries:   input.OutputDeliveries,
		SampleFiles:        input.SampleFiles,
		IsExpertiseOnline:  input.IsExpertiseOnline,
		IsExpertiseOffline: input.IsExpertiseOffline,
		CreatedAt:          &ca,
		CreatedBy:          &email,
		UpdatedAt:          &ca,
		UpdatedBy:          &email,
		Status:             &cd.Status,
	}

	return &res, nil
}

func UpdateContentDevelopment(ctx context.Context, input *model.ContentDevelopmentInput) (*model.ContentDevelopment, error) {
	if input.VendorID == "" || input.CdID == nil {
		return nil, fmt.Errorf("please provide both vendorId and CrtId")
	}

	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Errorf("Got error while getting claims: %v", err)
	}
	email := claims["email"].(string)
	lsp := claims["lsp_id"].(string)

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Errorf("Error while getting session: %v", err)
	}
	CassSession := session

	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.content_development WHERE vendor_id = '%s' ALLOW FILTERING`, input.VendorID)
	getContentDevelopment := func() (cdData []vendorz.ContentDevelopment, err error) {
		q := CassSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return cdData, iter.Select(&cdData)
	}
	cdDatas, err := getContentDevelopment()
	if err != nil {
		return nil, err
	}
	if len(cdDatas) == 0 {
		return nil, nil
	}
	cd := cdDatas[0]
	updatedCols := []string{}

	if input.Description != nil {
		cd.Description = *input.Description
		updatedCols = append(updatedCols, "description")
	}
	if input.Expertise != nil {
		tmp := ChangesStringType(input.Expertise)
		cd.Expertise = tmp
		updatedCols = append(updatedCols, "expertise")
	}
	if input.IsApplicable != nil {

		err = updateVendorLspMap(ctx, input.VendorID, lsp, "cd", *input.IsApplicable)
		if err != nil {
			return nil, err
		}

		cd.IsApplicable = *input.IsApplicable
		updatedCols = append(updatedCols, "is_applicable")
	}
	if input.Languages != nil {
		tmp := ChangesStringType(input.Languages)
		cd.Languages = tmp
		updatedCols = append(updatedCols, "languages")
	}
	if input.OutputDeliveries != nil {
		tmp := ChangesStringType(input.OutputDeliveries)
		cd.OutputDeliveries = tmp
		updatedCols = append(updatedCols, "output_deliveries")
	}
	if input.Status != nil {
		cd.Status = *input.Status
		updatedCols = append(updatedCols, "status")
	}
	if input.SampleFiles != nil {
		tmp := ChangesStringType(input.SampleFiles)
		cd.SampleFiles = tmp
		updatedCols = append(updatedCols, "sample_files")
	}
	if input.IsExpertiseOffline != nil {
		cd.IsExpertiseOffline = *input.IsExpertiseOffline
		updatedCols = append(updatedCols, "is_expertise_offline")
	}
	if input.IsExpertiseOnline != nil {
		cd.IsExpertiseOnline = *input.IsExpertiseOnline
		updatedCols = append(updatedCols, "is_expertise_online")
	}
	updatedAt := time.Now().Unix()
	if len(updatedCols) > 0 {
		cd.UpdatedAt = updatedAt
		cd.UpdatedBy = email
		updatedCols = append(updatedCols, "updated_at", "updated_by")

		utStms, uNames := vendorz.ContentDevelopmentTable.Update(updatedCols...)
		updateQuery := CassSession.Query(utStms, uNames).BindStruct(&cd)
		if err = updateQuery.ExecRelease(); err != nil {
			log.Errorf("Got error while updating content development: %v", err)
			return nil, err
		}
	}

	exp := ChangeToPointerArray(cd.Expertise)
	lan := ChangeToPointerArray(cd.Languages)
	od := ChangeToPointerArray(cd.OutputDeliveries)
	sf := ChangeToPointerArray(cd.SampleFiles)
	ca := strconv.Itoa(int(cd.CreatedAt))
	ua := strconv.Itoa(int(updatedAt))
	res := &model.ContentDevelopment{
		CdID:               &cd.CdId,
		VendorID:           &cd.VendorId,
		Description:        &cd.Description,
		IsApplicable:       &cd.IsApplicable,
		Expertise:          exp,
		Languages:          lan,
		OutputDeliveries:   od,
		SampleFiles:        sf,
		IsExpertiseOnline:  &cd.IsExpertiseOnline,
		IsExpertiseOffline: &cd.IsExpertiseOffline,
		CreatedAt:          &ca,
		CreatedBy:          &cd.CreatedBy,
		UpdatedAt:          &ua,
		UpdatedBy:          &email,
		Status:             &cd.Status,
	}
	return res, nil
}

func GetContentDevelopment(ctx context.Context, vendorID string) (*model.ContentDevelopment, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Errorf("Got error while getting claims: %v", err)
	}

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Errorf("Error while getting session: %v", err)
	}
	CassSession := session

	//vendor_id, sme_id
	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.content_development WHERE vendor_id = '%s' ALLOW FILTERING`, vendorID)
	getContentDevelopmentData := func() (cdData []vendorz.ContentDevelopment, err error) {
		q := CassSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return cdData, iter.Select(&cdData)
	}
	cdDatas, err := getContentDevelopmentData()
	if err != nil {
		return nil, err
	}
	if len(cdDatas) == 0 {
		return nil, nil
	}

	cd := cdDatas[0]
	exp := ChangeToPointerArray(cd.Expertise)
	lan := ChangeToPointerArray(cd.Languages)
	od := ChangeToPointerArray(cd.OutputDeliveries)
	sf := ChangeToPointerArray(cd.SampleFiles)
	ua := strconv.Itoa(int(cd.UpdatedAt))
	ca := strconv.Itoa(int(cd.CreatedAt))
	res := &model.ContentDevelopment{
		CdID:               &cd.CdId,
		VendorID:           &cd.VendorId,
		Description:        &cd.Description,
		IsApplicable:       &cd.IsApplicable,
		Expertise:          exp,
		Languages:          lan,
		OutputDeliveries:   od,
		SampleFiles:        sf,
		IsExpertiseOnline:  &cd.IsExpertiseOnline,
		IsExpertiseOffline: &cd.IsExpertiseOffline,
		CreatedAt:          &ca,
		CreatedBy:          &cd.CreatedBy,
		UpdatedAt:          &ua,
		UpdatedBy:          &cd.UpdatedBy,
		Status:             &cd.Status,
	}
	return res, nil
}

func DeleteSampleFile(ctx context.Context, sfID string, vendorID string, pType string) (*bool, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Errorf("Got error while getting claims: %v", err)
		return nil, err
	}
	val := false

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Errorf("Got error while getting session: %v", err)
		return nil, err
	}
	CassSession := session

	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.sample_file WHERE sf_id = '%s' AND vendor_id = '%s' AND p_type = '%s'  ALLOW FILTERING`, sfID, vendorID, pType)
	getSampleFile := func() (files []vendorz.SampleFile, err error) {
		q := CassSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return files, iter.Select(&files)
	}
	filesData, err := getSampleFile()
	if err != nil {
		return nil, err
	}
	if len(filesData) == 0 {
		return nil, nil
	}
	file := filesData[0]

	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		log.Printf("Failed to view sample files of course: %v", err.Error())
		return nil, err
	}
	fileBucket := ""
	if file.FileBucket != "" {
		fileBucket = file.FileBucket
	}

	deleteStr := fmt.Sprintf(`DELETE FROM vendorz.sample_file WHERE sf_id = '%s' `, sfID)
	if err = CassSession.Query(deleteStr, nil).Exec(); err != nil {
		return &val, err
	}

	res := storageC.DeleteObjectsFromBucket(ctx, fileBucket)
	if res == "success" {
		val = true
		return &val, nil
	}

	return &val, fmt.Errorf(res)
}

func GetUserVendors(ctx context.Context, userID *string) ([]*model.Vendor, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Errorf("Got error while getting claims: %v", err)
		return nil, err
	}

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Errorf("Got error while getting session: %v", err)
		return nil, err
	}
	CassSession := session
	queryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_user_map WHERE user_id = '%s' ALLOW FILTERING`, *userID)
	getVendorMap := func() (vendorUserMap []vendorz.VendorUserMap, err error) {
		q := CassSession.Query(queryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return vendorUserMap, iter.Select(&vendorUserMap)
	}

	data, err := getVendorMap()
	if err != nil {
		log.Printf("Got error while getting vendor user map: %v", err)
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	res := make([]*model.Vendor, len(data))
	var wg sync.WaitGroup
	for kk, vvv := range data {
		vv := vvv
		wg.Add(1)
		go func(k int, v string) {

			storageC := bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err = storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				log.Printf("Failed to view all profiles: %v", err.Error())
				return
			}
			photoUrl := ""

			query := fmt.Sprintf(`SELECT * FROM vendorz.vendor WHERE id = '%s'`, v)
			getVendor := func() (vendorData []vendorz.Vendor, err error) {
				qq := CassSession.Query(query, nil)
				defer qq.Release()
				iter := qq.Iter()
				return vendorData, iter.Select(&vendorData)
			}
			vendors, err := getVendor()
			if err != nil {
				log.Printf("Got error while getting vendors data: %v", err)
				return
			}
			if len(vendors) == 0 {
				return
			}

			vendor := vendors[0]

			if vendor.PhotoBucket != "" {
				photoUrl = storageC.GetSignedURLForObject(ctx, vendor.PhotoBucket)
			} else {
				photoUrl = ""
			}
			users := ChangeToPointerArray(vendor.Users)
			ca := strconv.Itoa(int(vendor.CreatedAt))
			ua := strconv.Itoa(int(vendor.UpdatedAt))
			tmp := model.Vendor{
				VendorID:     vendor.VendorId,
				Type:         vendor.Type,
				Level:        vendor.Level,
				Name:         vendor.Name,
				Description:  &vendor.Description,
				PhotoURL:     &photoUrl,
				Address:      &vendor.Address,
				Users:        users,
				Website:      &vendor.Website,
				FacebookURL:  &vendor.Facebook,
				InstagramURL: &vendor.Instagram,
				TwitterURL:   &vendor.Twitter,
				LinkedinURL:  &vendor.LinkedIn,
				CreatedAt:    &ca,
				CreatedBy:    &vendor.CreatedBy,
				UpdatedAt:    &ua,
				UpdatedBy:    &vendor.UpdatedBy,
				Status:       &vendor.Status,
			}
			res[k] = &tmp

			wg.Done()
		}(kk, vv.VendorId)
	}
	wg.Wait()

	return res, nil
}

func GetVendorServices(ctx context.Context, vendorID *string) ([]*string, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Errorf("Got error while getting claims: %v", err)
		return nil, err
	}
	lspId := claims["lsp_id"].(string)
	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return nil, err
	}
	CassSession := session
	res := []string{}

	//check for sme data
	qryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_lsp_map WHERE vendor_id='%s' AND lsp_id='%s' ALLOW FILTERING`, *vendorID, lspId)
	getVendorDetails := func() (vendorMaps []vendorz.VendorLspMap, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return vendorMaps, iter.Select(&vendorMaps)
	}
	vendors, err := getVendorDetails()
	if err != nil {
		return nil, err
	}
	if len(vendors) == 0 {
		return nil, errors.New("vendor does not exist")
	}
	vendor := vendors[0]
	services := vendor.Services
	for _, vv := range services {
		v := vv
		res = append(res, v)
	}
	var tmp []*string
	for _, values := range res {
		v := values
		tmp = append(tmp, &v)
	}

	return tmp, nil
}
