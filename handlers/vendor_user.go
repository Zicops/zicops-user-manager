package handlers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/zicops/contracts/vendorz"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
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
	if vendorID != nil {
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
	updatedQuery := CassSession.Query(stmt, names).BindStruct(vendorLsp)
	if err = updatedQuery.ExecRelease(); err != nil {
		return &res, err
	}

	res = true
	return &res, nil
}
