package handlers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/vendorz"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/lib/db/bucket"
	"github.com/zicops/zicops-user-manager/lib/googleprojectlib"
	"github.com/zicops/zicops-user-manager/lib/identity"
)

func AddOrder(ctx context.Context, input *model.VendorOrderInput) (*model.VendorOrder, error) {
	if input.VendorID == nil {
		return nil, errors.New("please enter vendor Id")
	}
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userEmail := claims["email"].(string)
	lspId := claims["lsp_id"].(string)
	if input.LspID != nil {
		lspId = *input.LspID
	}
	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	ca := time.Now().Unix()
	id := uuid.New().String()
	order := vendorz.VendorOrder{
		OrderId:   id,
		VendorId:  *input.VendorID,
		LspId:     lspId,
		CreatedAt: ca,
		CreatedBy: userEmail,
		UpdatedAt: ca,
		UpdatedBy: userEmail,
	}
	if input.Status != nil {
		order.Status = *input.Status
	}
	if input.Total != nil {
		order.Total = int64(*input.Total)
	}
	if input.Tax != nil {
		order.Tax = int64(*input.Tax)
	}
	if input.GrandTotal != nil {
		order.GrandTotal = int64(*input.GrandTotal)
	}
	insertQuery := CassSession.Query(vendorz.VendorOrderTable.Insert()).BindStruct(order)
	if err = insertQuery.Exec(); err != nil {
		return nil, err
	}
	createdAt := strconv.Itoa(int(order.CreatedAt))
	res := model.VendorOrder{
		ID:         &id,
		VendorID:   input.VendorID,
		LspID:      &lspId,
		Total:      input.Total,
		Tax:        input.Tax,
		GrandTotal: input.GrandTotal,
		CreatedAt:  &createdAt,
		CreatedBy:  &userEmail,
		UpdatedAt:  nil,
		UpdatedBy:  nil,
		Status:     input.Status,
	}

	return &res, nil
}

func UpdateOrder(ctx context.Context, input *model.VendorOrderInput) (*model.VendorOrder, error) {
	if input.VendorID == nil || input.ID == nil {
		return nil, errors.New("please enter vendor Id and orderId")
	}
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userEmail := claims["email"].(string)

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return nil, err
	}
	CassSession := session
	qryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_order WHERE id='%s' AND vendor_id='%s' ALLOW FILTERING`, *input.ID, *input.VendorID)
	getOrders := func() (vendorOrders []vendorz.VendorOrder, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return vendorOrders, iter.Select(&vendorOrders)
	}

	orders, err := getOrders()
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, nil
	}

	order := orders[0]
	updatedCols := []string{}
	if input.GrandTotal != nil {
		order.GrandTotal = int64(*input.GrandTotal)
		updatedCols = append(updatedCols, "grand_total")
	}

	if input.Status != nil {
		order.Status = *input.Status
		updatedCols = append(updatedCols, "status")
	}
	if input.Tax != nil {
		order.Tax = int64(*input.Tax)
		updatedCols = append(updatedCols, "tax")
	}
	if input.Total != nil {
		order.Total = int64(*input.Total)
		updatedCols = append(updatedCols, "total")
	}

	var updatedAt int64
	var updatedBy string
	if len(updatedCols) > 0 {
		updatedAt = time.Now().Unix()
		updatedBy = userEmail
		order.UpdatedAt = updatedAt
		order.UpdatedBy = updatedBy
		updatedCols = append(updatedCols, "updated_at", "updated_by")
		stmt, names := vendorz.VendorOrderTable.Update(updatedCols...)
		updatedQuery := CassSession.Query(stmt, names).BindStruct(&order)
		if err = updatedQuery.ExecRelease(); err != nil {
			return nil, err
		}
	}

	ca := strconv.Itoa(int(order.CreatedAt))
	total := int(order.Total)
	tax := int(order.Tax)
	grandTotal := int(order.GrandTotal)
	ua := strconv.Itoa(int(updatedAt))
	res := model.VendorOrder{
		ID:         &order.OrderId,
		VendorID:   &order.VendorId,
		LspID:      &order.LspId,
		Total:      &total,
		Tax:        &tax,
		GrandTotal: &grandTotal,
		CreatedAt:  &ca,
		CreatedBy:  &order.CreatedBy,
		UpdatedAt:  &ua,
		UpdatedBy:  &updatedBy,
		Status:     &order.Status,
	}
	return &res, nil
}

func AddOrderServies(ctx context.Context, input []*model.OrderServicesInput) ([]*model.OrderServices, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userEmail := claims["email"].(string)

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	var wg sync.WaitGroup
	res := make([]*model.OrderServices, len(input))
	for k, vvv := range input {
		wg.Add(1)
		v := vvv
		if v.OrderID == nil || v.ServiceType == nil {
			log.Println("enter both order Id and service id")
			continue
		}

		id := uuid.New().String()
		createdAt := time.Now().Unix()
		service := vendorz.OrderServices{
			ServiceId:   id,
			OrderId:     *v.OrderID,
			ServiceType: *v.ServiceType,
			CreatedAt:   createdAt,
			CreatedBy:   userEmail,
			UpdatedAt:   createdAt,
			UpdatedBy:   userEmail,
		}

		if v.Currency != nil {
			service.Currency = *v.Currency
		}
		if v.Description != nil {
			service.Description = *v.Description
		}
		if v.Quantity != nil {
			service.Quantity = int64(*v.Quantity)
		}
		if v.Rate != nil {
			service.Rate = int64(*v.Rate)
		}
		if v.Total != nil {
			service.Total = int64(*v.Total)
		}
		if v.Unit != nil {
			service.Unit = *v.Unit
		}
		if v.Status != nil {
			service.Status = *v.Status
		}
		insertQuery := CassSession.Query(vendorz.OrderServiesTable.Insert()).BindStruct(service)
		if err = insertQuery.Exec(); err != nil {
			log.Printf("Got error while entering data: %v", err)
			continue
		}

		ca := strconv.Itoa(int(createdAt))
		tmp := model.OrderServices{
			ServiceID:   &id,
			OrderID:     v.OrderID,
			ServiceType: v.ServiceType,
			Description: v.Description,
			Unit:        v.Unit,
			Currency:    v.Currency,
			Rate:        v.Rate,
			Quantity:    v.Quantity,
			Total:       v.Total,
			CreatedAt:   &ca,
			CreatedBy:   &userEmail,
			Status:      v.Status,
		}

		res[k] = &tmp

	}
	return res, nil
}

func UpdateOrderServices(ctx context.Context, input *model.OrderServicesInput) (*model.OrderServices, error) {
	if input.OrderID == nil || input.ServiceID == nil {
		return nil, fmt.Errorf("please pass both order id and service id")
	}
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userEmail := claims["email"].(string)

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	qryStr := fmt.Sprintf(`SELECT * FROM vendorz.order_services WHERE id='%s' AND order_id='%s' ALLOW FILTERING`, *input.ServiceID, *input.OrderID)
	getServices := func() (Orderservices []vendorz.OrderServices, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return Orderservices, iter.Select(&Orderservices)
	}
	services, err := getServices()
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, nil
	}

	service := services[0]
	updatedCols := []string{}

	if input.Currency != nil {
		service.Currency = *input.Currency
		updatedCols = append(updatedCols, "currency")
	}
	if input.Description != nil {
		service.Description = *input.Description
		updatedCols = append(updatedCols, "description")
	}
	if input.Quantity != nil {
		service.Quantity = int64(*input.Quantity)
		updatedCols = append(updatedCols, "quantity")
	}
	if input.Rate != nil {
		service.Rate = int64(*input.Rate)
		updatedCols = append(updatedCols, "rate")
	}
	if input.ServiceType != nil {
		service.ServiceType = *input.ServiceType
		updatedCols = append(updatedCols, "service_type")
	}
	if input.Status != nil {
		service.Status = *input.Status
		updatedCols = append(updatedCols, "status")
	}
	if input.Total != nil {
		service.Total = int64(*input.Total)
		updatedCols = append(updatedCols, "total")
	}
	if input.Unit != nil {
		service.Unit = *input.Unit
		updatedCols = append(updatedCols, "unit")
	}
	var updatedBy string
	var updatedAt int64
	if len(updatedCols) > 0 {
		updatedBy = userEmail
		updatedAt = time.Now().Unix()
		service.UpdatedAt = updatedAt
		service.UpdatedBy = updatedBy
		updatedCols = append(updatedCols, "updated_at", "updated_by")
		stmt, names := vendorz.OrderServiesTable.Update(updatedCols...)
		updatedQuery := CassSession.Query(stmt, names).BindStruct(&service)
		if err = updatedQuery.ExecRelease(); err != nil {
			return nil, err
		}
	}

	ca := strconv.Itoa(int(service.CreatedAt))
	ua := strconv.Itoa(int(updatedAt))
	rate := int(service.Rate)
	quantity := int(service.Quantity)
	total := int(service.Total)
	res := model.OrderServices{
		ServiceID:   &service.ServiceId,
		OrderID:     &service.OrderId,
		ServiceType: &service.ServiceType,
		Description: &service.Description,
		Unit:        &service.Unit,
		Currency:    &service.Currency,
		Rate:        &rate,
		Quantity:    &quantity,
		Total:       &total,
		CreatedAt:   &ca,
		CreatedBy:   &service.CreatedBy,
		UpdatedAt:   &ua,
		UpdatedBy:   &service.UpdatedBy,
		Status:      &service.Status,
	}
	return &res, nil
}

func GetAllOrders(ctx context.Context, lspID *string, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedVendorOrder, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	lsp := claims["lsp_id"].(string)
	if lspID != nil {
		lsp = *lspID
	}

	var newPage []byte
	if pageCursor != nil && *pageCursor != "" {
		page, err := global.CryptSession.DecryptString(*pageCursor, nil)
		if err != nil {
			return nil, err
		}
		newPage = page
	}

	var pageSizeInt int
	if pageSize != nil {
		pageSizeInt = *pageSize
	} else {
		pageSizeInt = 10
	}

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	qryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_order WHERE lsp_id='%s' ALLOW FILTERING`, lsp)
	getOrders := func(page []byte) (vendorOrders []vendorz.VendorOrder, newPage []byte, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		q.PageState(page)
		q.PageSize(pageSizeInt)
		iter := q.Iter()
		return vendorOrders, iter.PageState(), iter.Select(&vendorOrders)
	}

	orders, newPage, err := getOrders(newPage)
	if err != nil {
		return nil, err
	}

	var newCursor string
	if len(newPage) != 0 {
		newCursor, err = global.CryptSession.EncryptAsString(newPage, nil)
		if err != nil {
			return nil, err
		}
	}
	if len(orders) == 0 {
		return nil, nil
	}

	res := make([]*model.VendorOrder, len(orders))
	var wg sync.WaitGroup
	for kk, vv := range orders {
		v := vv
		wg.Add(1)
		go func(k int, order vendorz.VendorOrder) {
			total := int(order.Total)
			tax := int(order.Tax)
			grandTotal := int(order.Tax)
			ca := strconv.Itoa(int(order.CreatedAt))
			ua := strconv.Itoa(int(order.UpdatedAt))
			tmp := model.VendorOrder{
				ID:         &order.OrderId,
				VendorID:   &order.VendorId,
				LspID:      &order.LspId,
				Total:      &total,
				Tax:        &tax,
				GrandTotal: &grandTotal,
				CreatedAt:  &ca,
				CreatedBy:  &order.CreatedBy,
				UpdatedAt:  &ua,
				UpdatedBy:  &order.UpdatedBy,
				Status:     &order.Status,
			}

			res[k] = &tmp
			wg.Done()
		}(kk, v)
	}
	wg.Wait()

	resp := model.PaginatedVendorOrder{
		Orders:     res,
		PageCursor: &newCursor,
		Direction:  direction,
		PageSize:   &pageSizeInt,
	}
	return &resp, nil
}

func GetOrderServices(ctx context.Context, orderID []*string) ([]*model.OrderServices, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return nil, err
	}
	CassSession := session
	var orders []vendorz.OrderServices
	for _, vv := range orderID {
		v := *vv
		qryStr := fmt.Sprintf(`SELECT * FROM vendorz.order_services WHERE order_id = '%s' ALLOW FILTERING`, v)
		getServices := func() (services []vendorz.OrderServices, err error) {
			q := CassSession.Query(qryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return services, iter.Select(&services)
		}

		orderServices, err := getServices()
		if err != nil {
			log.Printf("Got error while getting services of an order: %v", err)
			return nil, err
		}
		if len(orderServices) == 0 {
			continue
		}
		orders = append(orders, orderServices...)
	}

	res := make([]*model.OrderServices, len(orders))
	var wg sync.WaitGroup
	for kk, vv := range orders {
		v := vv
		wg.Add(1)
		go func(k int, order vendorz.OrderServices) {

			rate := int(order.Rate)
			q := int(order.Quantity)
			total := int(order.Total)
			ca := strconv.Itoa(int(order.CreatedAt))
			ua := strconv.Itoa(int(order.UpdatedAt))
			tmp := model.OrderServices{
				ServiceID:   &order.ServiceId,
				OrderID:     &order.OrderId,
				ServiceType: &order.ServiceType,
				Description: &order.Description,
				Unit:        &order.Unit,
				Currency:    &order.Currency,
				Rate:        &rate,
				Quantity:    &q,
				Total:       &total,
				CreatedAt:   &ca,
				CreatedBy:   &order.CreatedBy,
				UpdatedAt:   &ua,
				UpdatedBy:   &order.UpdatedBy,
				Status:      &order.Status,
			}

			res[k] = &tmp

			wg.Done()
		}(kk, v)
	}
	wg.Wait()

	return res, nil
}

//getspeaker - speaker of that type =

func updateVendorLspMap(ctx context.Context, vendorId string, lsp string, service string, add bool) error {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Errorf("Got error while getting claims: %v", err)
	}
	email := claims["email"].(string)

	qryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_lsp_map WHERE vendor_id='%s' AND lsp_id='%s' ALLOW FILTERING`, vendorId, lsp)

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		log.Errorf("Error while getting session: %v", err)
	}
	CassSession := session

	getVendorLspMap := func() (maps []vendorz.VendorLspMap, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return maps, iter.Select(&maps)
	}

	maps, err := getVendorLspMap()
	if err != nil {
		return err
	}
	if len(maps) == 0 {
		return fmt.Errorf("map does not exist")
	}
	vendorLspMap := maps[0]
	services := vendorLspMap.Services
	// add is for checking whether to add or delete the value
	if add {

		//check if service already exists
		for _, v := range services {
			if v == service {
				return nil
			}
		}
		//if not then append
		services = append(services, service)

		vendorLspMap.Services = services
		vendorLspMap.UpdatedAt = time.Now().Unix()
		vendorLspMap.UpdatedBy = email

		updatedCols := []string{"services", "updated_by", "updated_at"}
		stmt, names := vendorz.VendorLspMapTable.Update(updatedCols...)
		updateQuery := CassSession.Query(stmt, names).BindStruct(&vendorLspMap)
		if err = updateQuery.ExecRelease(); err != nil {
			return err
		}

	} else {
		//it means delete the given service from the table
		var pos int
		//flag is used to see if service actually exists in the table or not
		flag := false
		for k, v := range services {
			if v == service {
				flag = true
				pos = k
			}
		}

		if flag {
			//we have the index of element in services array
			services = append(services[:pos], services[pos+1:]...)

			vendorLspMap.Services = services
			vendorLspMap.UpdatedAt = time.Now().Unix()
			vendorLspMap.UpdatedBy = email

			updatedCols := []string{"services", "updated_by", "updated_at"}
			stmt, names := vendorz.VendorLspMapTable.Update(updatedCols...)
			updateQuery := CassSession.Query(stmt, names).BindStruct(&vendorLspMap)
			if err = updateQuery.ExecRelease(); err != nil {
				return err
			}
		}

	}
	return nil
}

func GetSpeakers(ctx context.Context, lspID *string, service *string, name *string) ([]*model.VendorProfile, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	lsp := claims["lsp_id"].(string)
	if lspID != nil {
		lsp = *lspID
	}

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	qryStr := fmt.Sprintf(`SELECT * FROM vendorz.profile WHERE lsp_id='%s' AND is_speaker=true `, lsp)
	if service != nil && *service != "" {
		qryStr += fmt.Sprintf(` AND %s=true `, *service)
	}
	if name != nil {
		nameStr := strings.ToLower(*name)
		namesArray := strings.Fields(nameStr)
		for _, vv := range namesArray {
			v := vv
			qryStr += fmt.Sprintf(` AND name contains '%s' `, v)
		}
	}
	qryStr += ` ALLOW FILTERING`
	getProfiles := func() (maps []vendorz.VendorProfile, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return maps, iter.Select(&maps)
	}

	profiles, err := getProfiles()
	if err != nil {
		return nil, err
	}

	if len(profiles) == 0 {
		return nil, nil
	}

	var wg sync.WaitGroup
	res := make([]*model.VendorProfile, len(profiles))

	for kk, vvvv := range profiles {
		vvv := vvvv
		wg.Add(1)
		go func(k int, profile vendorz.VendorProfile) {

			storageC := bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err = storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				log.Errorf("Error initializing bucket: %v", err)
				return
			}

			var photoUrl string
			if profile.PhotoBucket != "" {
				photoUrl = storageC.GetSignedURLForObject(ctx, profile.PhotoBucket)
			} else {
				photoUrl = profile.PhotoURL
			}

			var lang []*string
			for _, vv := range profile.Languages {
				v := vv
				lang = append(lang, &v)
			}

			var sme []*string
			for _, vv := range profile.SMEExpertise {
				v := vv
				sme = append(sme, &v)
			}

			var cre []*string
			for _, vv := range profile.ClassroomExpertise {
				v := vv
				cre = append(cre, &v)
			}

			var cd []*string
			for _, vv := range profile.ContentDevelopment {
				v := vv
				cd = append(cd, &v)
			}

			var exp []*string
			for _, vv := range profile.Experience {
				v := vv
				exp = append(exp, &v)
			}

			ca := strconv.Itoa(int(profile.CreatedAt))
			ua := strconv.Itoa(int(profile.UpdatedAt))
			tmp := model.VendorProfile{
				PfID:               &profile.PfId,
				VendorID:           &profile.VendorId,
				FirstName:          &profile.FirstName,
				LastName:           &profile.LastName,
				Email:              &profile.Email,
				Phone:              &profile.Phone,
				PhotoURL:           &photoUrl,
				Description:        &profile.Description,
				Language:           lang,
				SmeExpertise:       sme,
				ClassroomExpertise: cre,
				ContentDevelopment: cd,
				Experience:         exp,
				ExperienceYears:    &profile.ExperienceYears,
				Sme:                &profile.Sme,
				Crt:                &profile.Crt,
				Cd:                 &profile.Cd,
				IsSpeaker:          &profile.IsSpeaker,
				LspID:              &profile.LspId,
				CreatedAt:          &ca,
				CreatedBy:          &profile.CreatedBy,
				UpdatedAt:          &ua,
				UpdatedBy:          &profile.UpdatedBy,
				Status:             &profile.Status,
			}

			res[k] = &tmp
			wg.Done()
		}(kk, vvv)
	}

	wg.Wait()
	return res, nil
}

func GetOrders(ctx context.Context, orderID []*string) ([]*model.VendorOrder, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	session, err := global.CassPool.GetSession(ctx, "vendorz")
	if err != nil {
		return nil, err
	}
	CassSession := session
	res := make([]*model.VendorOrder, len(orderID))
	var wg sync.WaitGroup
	for kk, vv := range orderID {
		v := vv
		if v == nil {
			continue
		}
		wg.Add(1)

		go func(k int, id *string) {
			qryStr := fmt.Sprintf(`SELECT * FROM vendorz.vendor_order where id='%s' ALLOW FILTERING`, *id)
			getOrder := func() (orderDetails []vendorz.VendorOrder, err error) {
				q := CassSession.Query(qryStr, nil)
				defer q.Release()
				iter := q.Iter()
				return orderDetails, iter.Select(&orderDetails)
			}
			orders, err := getOrder()
			if err != nil {
				log.Errorf("Got error while geting order details: %v", err)
				return
			}
			if len(orders) == 0 {
				return
			}
			order := orders[0]
			total := int(order.Total)
			tax := int(order.Tax)
			grandTotal := int(order.GrandTotal)
			ca := strconv.Itoa(int(order.CreatedAt))
			ua := strconv.Itoa(int(order.UpdatedAt))
			tmp := model.VendorOrder{
				ID:         &order.OrderId,
				VendorID:   &order.VendorId,
				LspID:      &order.LspId,
				Total:      &total,
				Tax:        &tax,
				GrandTotal: &grandTotal,
				CreatedAt:  &ca,
				CreatedBy:  &order.CreatedBy,
				UpdatedAt:  &ua,
				UpdatedBy:  &order.UpdatedBy,
				Status:     &order.Status,
			}
			res[k] = &tmp
			wg.Done()
		}(kk, v)
	}
	wg.Wait()

	return res, nil
}
