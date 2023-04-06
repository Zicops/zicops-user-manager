package handlers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/contracts/vendorz"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/lib/db/bucket"
	"github.com/zicops/zicops-user-manager/lib/googleprojectlib"
	"github.com/zicops/zicops-user-manager/lib/identity"
)

func CreateVendorUserMap(ctx context.Context, vendorID *string, userID *string, status *string) (*model.VendorUserMap, error) {
	if vendorID == nil || userID == nil || status == nil {
		return nil, errors.New("please enter all the features vendorId, userId, status")
	}
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err

	}
	email := claims["email"].(string)

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	ca := time.Now().Unix()
	vendorMap := vendorz.VendorUserMap{
		VendorId:  *vendorID,
		UserId:    *userID,
		CreatedAt: ca,
		CreatedBy: email,
		UpdatedAt: ca,
		UpdatedBy: email,
		Status:    *status,
	}

	insertQuery := CassSession.Query(vendorz.VendorUserMapTable.Insert()).BindStruct(vendorMap)
	if err = insertQuery.Exec(); err != nil {
		return nil, err
	}

	createdAt := strconv.Itoa(int(ca))
	res := model.VendorUserMap{
		VendorID:  vendorID,
		UserID:    userID,
		CreatedAt: &createdAt,
		CreatedBy: &email,
		UpdatedAt: &createdAt,
		UpdatedBy: &email,
		Status:    status,
	}
	return &res, nil
}

func UpdateVendorUserMap(ctx context.Context, vendorID *string, userID *string, status *string) (*model.VendorUserMap, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email := claims["email"].(string)

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	qryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_user_map WHERE vendor_id='%s' AND user_id='%s' ALLOW FILTERING`, *vendorID, *userID)
	getVendorUserMap := func() (maps []vendorz.VendorUserMap, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return maps, iter.Select(&maps)
	}

	maps, err := getVendorUserMap()
	if err != nil {
		return nil, err
	}

	if len(maps) == 0 {
		return nil, nil
	}

	vendorUserMap := maps[0]
	var updatedCols []string
	if status != nil {
		vendorUserMap.Status = *status
		updatedCols = append(updatedCols, "status")
	}
	ua := time.Now().Unix()
	if len(updatedCols) != 0 {
		updatedCols = append(updatedCols, "updated_at")
		updatedCols = append(updatedCols, "updated_by")
		vendorUserMap.UpdatedAt = ua
		vendorUserMap.UpdatedBy = email
		stmt, names := vendorz.VendorUserMapTable.Update(updatedCols...)
		updateQuery := CassSession.Query(stmt, names).BindStruct(&vendorUserMap)
		if err = updateQuery.ExecRelease(); err != nil {
			return nil, err
		}
	}

	ca := strconv.Itoa(int(vendorUserMap.CreatedAt))
	updatedAt := strconv.Itoa(int(ua))
	res := model.VendorUserMap{
		VendorID:  vendorID,
		UserID:    userID,
		CreatedAt: &ca,
		CreatedBy: &vendorUserMap.CreatedBy,
		UpdatedAt: &updatedAt,
		UpdatedBy: &email,
		Status:    &vendorUserMap.Status,
	}

	return &res, nil
}

func DeleteVendorUserMap(ctx context.Context, vendorID *string, userID *string) (*bool, error) {
	if vendorID == nil && userID == nil {
		return nil, fmt.Errorf("please enter both vendorId and userId")
	}
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	res := false
	deleteStr := fmt.Sprintf(`DELETE FROM vendorz.vendor_user_map WHERE vendor_id='%s' AND user_id='%s'`, *vendorID, *userID)
	if err = CassSession.Query(deleteStr, nil).Exec(); err != nil {
		return &res, err
	}
	res = true
	return &res, nil
}

func DisableVendorLspMap(ctx context.Context, vendorID *string, lspID *string) (*bool, error) {
	if vendorID == nil {
		return nil, fmt.Errorf("please enter vendorId")
	}
	res := false
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return &res, err
	}
	email := claims["email"].(string)
	lsp := claims["lsp_id"].(string)
	if lspID != nil {
		lsp = *lspID
	}

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return &res, err
	}
	CassSession := session
	qryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_lsp_map WHERE vendor_id='%s' AND lsp_id='%s' ALLOW FILTERING`, *vendorID, lsp)
	getVendorLsp := func() (maps []vendorz.VendorLspMap, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return maps, iter.Select(&maps)
	}

	vendorLspMaps, err := getVendorLsp()
	if err != nil {
		return &res, err
	}
	ua := time.Now().Unix()
	vendorLsp := vendorLspMaps[0]
	vendorLsp.Status = "disable"
	vendorLsp.UpdatedAt = ua
	vendorLsp.UpdatedBy = email
	updatedCols := []string{"status", "updated_at", "updated_by"}
	stmt, names := vendorz.VendorLspMapTable.Update(updatedCols...)
	updatedQuery := CassSession.Query(stmt, names).BindStruct(&vendorLsp)
	if err = updatedQuery.ExecRelease(); err != nil {
		return &res, err
	}

	err = disableUsersOfVendors(ctx, *vendorID, lsp, email)
	if err != nil {
		log.Printf("Got error while disabling users: %v", err)
		return &res, err
	}

	res = true
	return &res, nil
}

func disableUsersOfVendors(ctx context.Context, vendorId string, lsp string, email string) error {

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return err
	}
	CassSession := session
	qryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_user_map WHERE vendor_id='%s' ALLOW FILTERING`, vendorId)
	getUsers := func() (maps []vendorz.VendorUserMap, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Iter()
		iter := q.Iter()
		return maps, iter.Select(&maps)
	}
	vendorUserMaps, err := getUsers()
	if err != nil {
		return err
	}
	if len(vendorUserMaps) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	for _, vv := range vendorUserMaps {
		v := vv
		wg.Add(1)
		go func(userId string, lspId string, email string) {
			defer wg.Done()

			session, err := global.CassPool.GetSession(ctx, "userz")
			if err != nil {
				log.Printf("Got error while getting users data: %v", err)
				return
			}
			CassUserSession := session

			query := fmt.Sprintf(`SELECT * FROM userz.user_lsp_map WHERE user_id='%s' AND lsp_id='%s' ALLOW FILTERING`, userId, lspId)
			getUserData := func() (usersData []userz.UserLsp, err error) {
				q := CassUserSession.Query(query, nil)
				defer q.Release()
				iter := q.Iter()
				return usersData, iter.Select(&usersData)
			}

			users, err := getUserData()
			if err != nil {
				log.Printf("Got error while get users data: %v", err)
				return
			}
			if len(users) == 0 {
				return
			}

			user := users[0]
			user.Status = "disable"
			user.UpdatedAt = time.Now().Unix()
			user.UpdatedBy = email
			updatedCols := []string{"status", "updated_at", "updated_by"}

			stmt, names := userz.UserLspTable.Update(updatedCols...)
			updatedQuery := CassUserSession.Query(stmt, names).BindStruct(&user)
			if err = updatedQuery.ExecRelease(); err != nil {
				log.Printf("Got error while getting user info: %v", err)
				return
			}

		}(v.UserId, lsp, email)
	}
	wg.Wait()
	return nil
}

func GetAllVendors(ctx context.Context, vendorIds []*string) ([]*model.Vendor, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	res := make([]*model.Vendor, len(vendorIds))
	var wg sync.WaitGroup
	for kk, vv := range vendorIds {
		if vv == nil {
			return nil, err
		}
		v := *vv
		wg.Add(1)
		go func(k int, vendorId string) {
			defer wg.Done()
			qryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor where id='%s'`, vendorId)
			getVendors := func() (vendorDetails []vendorz.Vendor, err error) {
				q := CassSession.Query(qryStr, nil)
				defer q.Release()
				iter := q.Iter()
				return vendorDetails, iter.Select(&vendorDetails)
			}

			vendors, err := getVendors()
			if err != nil {
				log.Errorf("Got error while getting vendors details: %v", err)
				return
			}
			if len(vendors) == 0 {
				return
			}

			vendor := vendors[0]

			qry := fmt.Sprintf(`SELECT * FROM vendorz.vendor_lsp_map WHERE vendor_id='%s' ALLOW FILTERING`, vendorId)
			getMaps := func() (maps []vendorz.VendorLspMap, err error) {
				q := CassSession.Query(qry, nil)
				defer q.Release()
				iter := q.Iter()
				return maps, iter.Select(&maps)
			}
			vendorLspMaps, err := getMaps()
			if err != nil {
				log.Errorf("Got error while getting vendor lsp map: %v", err)
				return
			}
			if len(vendorLspMaps) == 0 {
				return
			}
			servicesVendor := vendorLspMaps[0].Services
			vendorLspStatus := vendorLspMaps[0].Status

			var users []*string
			for _, xx := range vendor.Users {
				x := xx
				users = append(users, &x)
			}
			var services []*string
			for _, xx := range servicesVendor {
				x := xx
				services = append(services, &x)
			}

			storageC := bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err = storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				log.Printf("Failed to upload image to course: %v", err.Error())
				return
			}
			photoUrl := ""
			if vendor.PhotoBucket != "" {
				photoUrl = storageC.GetSignedURLForObject(ctx, vendor.PhotoBucket)
			} else {
				photoUrl = vendor.PhotoUrl
			}
			ca := strconv.Itoa(int(vendor.CreatedAt))
			ua := strconv.Itoa(int(vendor.UpdatedAt))
			tmp := model.Vendor{
				VendorID:        vendor.VendorId,
				Type:            vendor.Type,
				Level:           vendor.Level,
				Name:            vendor.Name,
				Phone:           &vendor.Phone,
				LspID:           &vendor.LspId,
				Description:     &vendor.Description,
				PhotoURL:        &photoUrl,
				Address:         &vendor.Address,
				Users:           users,
				Website:         &vendor.Website,
				FacebookURL:     &vendor.Facebook,
				InstagramURL:    &vendor.Instagram,
				TwitterURL:      &vendor.Twitter,
				LinkedinURL:     &vendor.LinkedIn,
				Services:        services,
				CreatedAt:       &ca,
				CreatedBy:       &vendor.CreatedBy,
				UpdatedAt:       &ua,
				UpdatedBy:       &vendor.UpdatedBy,
				Status:          &vendor.Status,
				VendorLspStatus: &vendorLspStatus,
			}
			res[k] = &tmp

		}(kk, v)
	}
	wg.Wait()
	return res, nil
}
