package orgs

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/helpers"
	"github.com/zicops/zicops-user-manager/lib/db/bucket"
	"github.com/zicops/zicops-user-manager/lib/googleprojectlib"
)

func AddOrganization(ctx context.Context, input model.OrganizationInput) (*model.Organization, error) {
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
	if strings.ToLower(role) != "puneet@zicops.com" {
		return nil, fmt.Errorf("user is a not an admin: Unauthorized")
	}
	logoUrl := ""
	logoBucket := ""
	var storageC *bucket.Client
	uniqueOrgId := input.Name + input.Website + input.Industry
	orgId := base64.URLEncoding.EncodeToString([]byte(uniqueOrgId))
	if input.Logo != nil {
		if storageC == nil {
			storageC = bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err := storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				return nil, err
			}
		}
		bucketPath := fmt.Sprintf("orgs/%s/%s/%s", "logos", orgId, input.Logo.Filename)
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
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if input.OrgID == nil {
		return nil, fmt.Errorf("org id is required")
	}
	session, err := cassandra.GetCassSession("userz")
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
	if input.Logo != nil {
		if storageC == nil {
			storageC = bucket.NewStorageHandler()
			gproject := googleprojectlib.GetGoogleProjectID()
			err := storageC.InitializeStorageClient(ctx, gproject)
			if err != nil {
				return nil, err
			}
		}
		bucketPath := fmt.Sprintf("orgs/%s/%s/%s", "logos", orgCass.ID, input.Logo.Filename)
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
		orgCass.LogoURL = storageC.GetSignedURLForObject(bucketPath)
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
	return result, nil

}

func GetOrganizations(ctx context.Context, orgIds []*string) ([]*model.Organization, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	outputOrgs := []*model.Organization{}
	for _, orgID := range orgIds {
		if orgID == nil {
			continue
		}
		qryStr := fmt.Sprintf(`SELECT * from userz.organization where id='%s' `, *orgID)
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
			continue
		}
		orgCass := orgs[0]
		created := strconv.FormatInt(orgCass.CreatedAt, 10)
		updated := strconv.FormatInt(orgCass.UpdatedAt, 10)
		emptCnt, _ := strconv.Atoi(orgCass.EmpCount)
		logoUrl := orgCass.LogoURL
		if orgCass.LogoBucket != "" {
			storageC := bucket.NewStorageHandler()
			if storageC == nil {
				storageC = bucket.NewStorageHandler()
				gproject := googleprojectlib.GetGoogleProjectID()
				err := storageC.InitializeStorageClient(ctx, gproject)
				if err != nil {
					return nil, err
				}
			}
			logoUrl = storageC.GetSignedURLForObject(orgCass.LogoBucket)
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
		outputOrgs = append(outputOrgs, result)
	}
	return outputOrgs, nil
}
