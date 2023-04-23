package orgs

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/lib/db/bucket"
	"github.com/zicops/zicops-user-manager/lib/googleprojectlib"
	"github.com/zicops/zicops-user-manager/lib/identity"
)

func AddOrganization(ctx context.Context, input model.OrganizationInput) (*model.Organization, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	roleValue := claims["email"]
	role := roleValue.(string)
	if strings.ToLower(role) != "puneet@zicops.com" {
		return nil, fmt.Errorf("user is a not an admin: Unauthorized")
	}
	logoUrl := ""
	logoBucket := ""
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		return nil, err
	}
	uniqueOrgId := input.Name + input.Website + input.Industry + input.Subdomain
	orgId := uuid.NewSHA1(uuid.NameSpaceURL, []byte(uniqueOrgId)).String()
	if input.Logo != nil {
		extension := strings.Split(input.Logo.Filename, ".")
		bucketPath := fmt.Sprintf("orgs/%s/%s/%s", "logos", orgId, base64.URLEncoding.EncodeToString([]byte(input.Logo.Filename)))
		if len(extension) > 1 {
			bucketPath += "." + extension[len(extension)-1]
		}
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
		logoUrl = storageC.GetSignedURLForObject(ctx, bucketPath)
	} else {
		logoBucket = ""
		if input.LogoURL != nil {
			logoUrl = *input.LogoURL
		}
	}
	linkedin := ""
	if input.LinkedinURL != nil {
		linkedin = *input.LinkedinURL
	}
	facebook := ""
	if input.FacebookURL != nil {
		facebook = *input.FacebookURL
	}
	twitter := ""
	if input.TwitterURL != nil {
		twitter = *input.TwitterURL
	}

	orgCass := userz.Organization{
		ID:              orgId,
		Name:            input.Name,
		Website:         input.Website,
		Industry:        input.Industry,
		LogoURL:         logoUrl,
		LogoBucket:      logoBucket,
		ZicopsSubdomain: input.Subdomain,
		Status:          input.Status,
		Facebook:        facebook,
		Twitter:         twitter,
		Linkedin:        linkedin,
		EmpCount:        fmt.Sprintf("%v", input.EmployeeCount),
		CreatedAt:       time.Now().Unix(),
		UpdatedAt:       time.Now().Unix(),
		CreatedBy:       role,
		UpdatedBy:       role,
		Type:            input.Type,
	}
	insertQuery := CassUserSession.Query(userz.OrganizationTable.Insert()).BindStruct(orgCass)
	if err := insertQuery.ExecRelease(); err != nil {
		return nil, err
	}
	created := strconv.FormatInt(orgCass.CreatedAt, 10)
	updated := strconv.FormatInt(orgCass.UpdatedAt, 10)
	org := &model.Organization{
		OrgID:         &orgCass.ID,
		Name:          orgCass.Name,
		Website:       orgCass.Website,
		Industry:      orgCass.Industry,
		LogoURL:       &orgCass.LogoURL,
		Subdomain:     orgCass.ZicopsSubdomain,
		Status:        orgCass.Status,
		FacebookURL:   &orgCass.Facebook,
		TwitterURL:    &orgCass.Twitter,
		LinkedinURL:   &orgCass.Linkedin,
		EmployeeCount: input.EmployeeCount,
		CreatedAt:     created,
		UpdatedAt:     updated,
		CreatedBy:     &orgCass.CreatedBy,
		UpdatedBy:     &orgCass.UpdatedBy,
		Type:          orgCass.Type,
	}
	return org, nil
}

func UpdateOrganization(ctx context.Context, input model.OrganizationInput) (*model.Organization, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if input.OrgID == nil {
		return nil, fmt.Errorf("org id is required")
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	email := claims["email"]
	role := email.(string)
	if strings.ToLower(role) != "puneet@zicops.com" {
		return nil, fmt.Errorf("user is a not zicops admin: Unauthorized")
	}
	orgCass := userz.Organization{
		ID: *input.OrgID,
	}
	org := []userz.Organization{}

	getQueryStr := fmt.Sprintf(`SELECT * from userz.organization where id='%s' `, orgCass.ID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
	if err := getQuery.SelectRelease(&org); err != nil {
		return nil, err
	}
	orgCass = org[0]
	updatedCols := []string{}
	if input.Name != orgCass.Name {
		orgCass.Name = input.Name
		updatedCols = append(updatedCols, "name")
	}
	if input.Website != orgCass.Website {
		orgCass.Website = input.Website
		updatedCols = append(updatedCols, "website")
	}
	if input.Industry != orgCass.Industry {
		orgCass.Industry = input.Industry
		updatedCols = append(updatedCols, "industry")
	}
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		return nil, err
	}
	if input.Logo != nil {
		extension := strings.Split(input.Logo.Filename, ".")
		bucketPath := fmt.Sprintf("orgs/%s/%s/%s", "logos", orgCass.ID, base64.URLEncoding.EncodeToString([]byte(input.Logo.Filename)))
		if len(extension) > 1 {
			bucketPath += "." + extension[len(extension)-1]
		}
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
		orgCass.LogoBucket = bucketPath
		orgCass.LogoURL = storageC.GetSignedURLForObject(ctx, bucketPath)
		updatedCols = append(updatedCols, "logo_url")
		updatedCols = append(updatedCols, "logo_bucket")
	}
	if input.LogoURL != nil && input.Logo == nil && *input.LogoURL != orgCass.LogoURL {
		orgCass.LogoURL = *input.LogoURL
		updatedCols = append(updatedCols, "logo_url")
	}
	if input.Subdomain != orgCass.ZicopsSubdomain {
		orgCass.ZicopsSubdomain = input.Subdomain
		updatedCols = append(updatedCols, "zicops_subdomain")
	}
	if input.Status != orgCass.Status {
		orgCass.Status = input.Status
		updatedCols = append(updatedCols, "status")
	}
	if input.FacebookURL != nil && *input.FacebookURL != orgCass.Facebook {
		orgCass.Facebook = *input.FacebookURL
		updatedCols = append(updatedCols, "facebook")
	}
	if input.TwitterURL != nil && *input.TwitterURL != orgCass.Twitter {
		orgCass.Twitter = *input.TwitterURL
		updatedCols = append(updatedCols, "twitter")
	}
	if input.LinkedinURL != nil && *input.LinkedinURL != orgCass.Linkedin {
		orgCass.Linkedin = *input.LinkedinURL
		updatedCols = append(updatedCols, "linkedin")
	}
	empCnt := fmt.Sprintf("%v", input.EmployeeCount)
	if empCnt != orgCass.EmpCount {
		orgCass.EmpCount = empCnt
		updatedCols = append(updatedCols, "emp_count")
	}
	if input.Type != orgCass.Type {
		orgCass.Type = input.Type
		updatedCols = append(updatedCols, "type")
	}
	if len(updatedCols) > 0 {
		orgCass.UpdatedAt = time.Now().Unix()
		orgCass.UpdatedBy = role
		updatedCols = append(updatedCols, "updated_at")
		upStms, uNames := userz.OrganizationTable.Update(updatedCols...)
		updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&orgCass)
		if err := updateQuery.ExecRelease(); err != nil {
			log.Errorf("error updating user: %v", err)
			return nil, err
		}
	}
	created := strconv.FormatInt(orgCass.CreatedAt, 10)
	updated := strconv.FormatInt(orgCass.UpdatedAt, 10)
	result := &model.Organization{
		OrgID:         &orgCass.ID,
		Name:          orgCass.Name,
		Website:       orgCass.Website,
		Industry:      orgCass.Industry,
		LogoURL:       &orgCass.LogoURL,
		Subdomain:     orgCass.ZicopsSubdomain,
		Status:        orgCass.Status,
		FacebookURL:   &orgCass.Facebook,
		TwitterURL:    &orgCass.Twitter,
		LinkedinURL:   &orgCass.Linkedin,
		EmployeeCount: input.EmployeeCount,
		CreatedAt:     created,
		CreatedBy:     &orgCass.CreatedBy,
		UpdatedAt:     updated,
		UpdatedBy:     &orgCass.UpdatedBy,
		Type:          orgCass.Type,
	}

	key := fmt.Sprintf("org:%s", orgCass.ZicopsSubdomain)
	redisBytes, err := json.Marshal(orgCass)
	if err == nil {
		redis.SetRedisValue(ctx, key, string(redisBytes))
	}
	return result, nil

}

func GetOrganizations(ctx context.Context, orgIds []*string) ([]*model.Organization, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	outputOrgs := make([]*model.Organization, len(orgIds))
	var wg sync.WaitGroup
	for i, orgid := range orgIds {
		if orgid == nil {
			continue
		}
		orgIDInput := orgid
		wg.Add(1)
		go func(i int, orgID *string) {
			if orgID == nil {
				return
			}
			orgCass := userz.Organization{}
			key := fmt.Sprintf("org:%s", *orgID)
			res, err := redis.GetRedisValue(ctx, key)
			if err == nil && res != "" {
				json.Unmarshal([]byte(res), &orgCass)
			}
			if orgCass.ID == "" {
				qryStr := fmt.Sprintf(`SELECT * from userz.organization where id='%s' ALLOW FILTERING`, *orgID)
				getOrgs := func() (users []userz.Organization, err error) {
					q := CassUserSession.Query(qryStr, nil)
					defer q.Release()
					iter := q.Iter()
					return users, iter.Select(&users)
				}
				orgs, err := getOrgs()
				if err != nil {
					return
				}
				if len(orgs) == 0 {
					return
				}
				orgCass = orgs[0]
			}
			created := strconv.FormatInt(orgCass.CreatedAt, 10)
			updated := strconv.FormatInt(orgCass.UpdatedAt, 10)
			emptCnt, _ := strconv.Atoi(orgCass.EmpCount)
			logoUrl := orgCass.LogoURL
			storageC := bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err = storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				return
			}
			if orgCass.LogoBucket != "" {
				logoUrl = storageC.GetSignedURLForObject(ctx, orgCass.LogoBucket)
			}

			result := &model.Organization{
				OrgID:         &orgCass.ID,
				Name:          orgCass.Name,
				Website:       orgCass.Website,
				Industry:      orgCass.Industry,
				LogoURL:       &logoUrl,
				Subdomain:     orgCass.ZicopsSubdomain,
				Status:        orgCass.Status,
				FacebookURL:   &orgCass.Facebook,
				TwitterURL:    &orgCass.Twitter,
				LinkedinURL:   &orgCass.Linkedin,
				EmployeeCount: emptCnt,
				CreatedAt:     created,
				CreatedBy:     &orgCass.CreatedBy,
				UpdatedAt:     updated,
				UpdatedBy:     &orgCass.UpdatedBy,
				Type:          orgCass.Type,
			}
			outputOrgs[i] = result
			redisBytes, err := json.Marshal(orgCass)
			if err == nil {
				redis.SetRedisValue(ctx, key, string(redisBytes))
				redis.SetTTL(ctx, key, 3600)
			}
			wg.Done()
		}(i, orgIDInput)
	}
	wg.Wait()
	return outputOrgs, nil
}

func GetOrganizationsByName(ctx context.Context, name *string, prevPageSnapShot string, pageSize int) ([]*model.Organization, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email := claims["email"].(string)
	if strings.ToLower(email) != "puneet@zicops.com" {
		return nil, fmt.Errorf("user is a not zicops admin: Unauthorized")
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}

	CassUserSession := session

	var qryStr string
	if name == nil || *name == "" {
		qryStr = `SELECT * from userz.organization`
	} else {
		n := strings.ToLower(*name)
		qryStr = fmt.Sprintf(`SELECT * from userz.organization where name='%s' `, n)
	}

	getOrgs := func() (users []userz.Organization, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return users, iter.Select(&users)
	}
	orgs, err := getOrgs()
	if err != nil {
		return nil, err
	}
	if len(orgs) == 0 {
		return nil, nil
	}
	//here will map data from backend to models interface and return it
	var result []*model.Organization
	for _, vv := range orgs {

		v := vv
		id := &v.ID
		empCount, _ := strconv.Atoi(v.EmpCount)
		lnkdIn := &v.Linkedin
		fb := &v.Facebook
		twtr := &v.Twitter
		createdAt := strconv.Itoa(int(v.CreatedAt))
		updatedAt := strconv.Itoa(int(v.UpdatedAt))
		cb := &v.CreatedBy
		ub := &v.UpdatedBy

		logoUrl := v.LogoURL
		storageC := bucket.NewStorageHandler()
		gproject := googleprojectlib.GetGoogleProjectID()
		err = storageC.InitializeStorageClient(ctx, gproject)
		if err != nil {
			return nil, nil
		}
		if v.LogoBucket != "" {
			logoUrl = storageC.GetSignedURLForObject(ctx, v.LogoBucket)
		}

		temp := &model.Organization{
			OrgID:         id,
			Name:          v.Name,
			LogoURL:       &logoUrl,
			Industry:      v.Industry,
			Type:          v.Type,
			Subdomain:     v.ZicopsSubdomain,
			EmployeeCount: empCount,
			Website:       v.Website,
			LinkedinURL:   lnkdIn,
			FacebookURL:   fb,
			TwitterURL:    twtr,
			Status:        v.Status,
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
			CreatedBy:     cb,
			UpdatedBy:     ub,
		}
		result = append(result, temp)
	}
	return result, nil
}

func GetOrganizationsByDomain(ctx context.Context, domain string) ([]*model.Organization, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	outputOrgs := make([]*model.Organization, 1)
	log.Errorf("domain: %v", domain)
	orgCass := userz.Organization{}
	key := fmt.Sprintf("org:%s", domain)
	res, err := redis.GetRedisValue(ctx, key)
	if err == nil && res != "" {
		json.Unmarshal([]byte(res), &orgCass)
	}
	if orgCass.ID == "" {
		qryStr := fmt.Sprintf(`SELECT * from userz.organization where zicops_subdomain='%s' ALLOW FILTERING`, domain)
		getOrgs := func() (users []userz.Organization, err error) {
			q := CassUserSession.Query(qryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return users, iter.Select(&users)
		}
		orgs, err := getOrgs()
		if err != nil {
			return nil, err
		}
		if len(orgs) == 0 {
			return nil, fmt.Errorf("no organization found")
		}
		orgCass = orgs[0]
	}
	created := strconv.FormatInt(orgCass.CreatedAt, 10)
	updated := strconv.FormatInt(orgCass.UpdatedAt, 10)
	emptCnt, _ := strconv.Atoi(orgCass.EmpCount)
	logoUrl := orgCass.LogoURL
	storageC := bucket.NewStorageHandler()
	gproject := googleprojectlib.GetGoogleProjectID()
	err = storageC.InitializeStorageClient(ctx, gproject)
	if err != nil {
		return nil, err
	}
	if orgCass.LogoBucket != "" {
		logoUrl = storageC.GetSignedURLForObject(ctx, orgCass.LogoBucket)
	}

	result := &model.Organization{
		OrgID:         &orgCass.ID,
		Name:          orgCass.Name,
		Website:       orgCass.Website,
		Industry:      orgCass.Industry,
		LogoURL:       &logoUrl,
		Subdomain:     orgCass.ZicopsSubdomain,
		Status:        orgCass.Status,
		FacebookURL:   &orgCass.Facebook,
		TwitterURL:    &orgCass.Twitter,
		LinkedinURL:   &orgCass.Linkedin,
		EmployeeCount: emptCnt,
		CreatedAt:     created,
		CreatedBy:     &orgCass.CreatedBy,
		UpdatedAt:     updated,
		UpdatedBy:     &orgCass.UpdatedBy,
		Type:          orgCass.Type,
	}
	outputOrgs[0] = result
	redisBytes, err := json.Marshal(orgCass)
	if err == nil {
		redis.SetRedisValue(ctx, key, string(redisBytes))
	}
	return outputOrgs, nil
}
