package orgs

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/helpers"
)

func AddOrganizationUnit(ctx context.Context, input model.OrganizationUnitInput) (*model.OrganizationUnit, error) {
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
	uniqueOrgId := input.OrgID + input.Address + input.PostalCode + input.City + input.State + input.Country
	orgId := base64.URLEncoding.EncodeToString([]byte(uniqueOrgId))

	orgCass := userz.OrganizationUnits{
		ID:         orgId,
		OrgID:      input.OrgID,
		EmpCount:   fmt.Sprintf("%d", input.EmpCount),
		Address:    input.Address,
		PostalCode: input.PostalCode,
		City:       input.City,
		State:      input.State,
		Country:    input.Country,
		Status:     input.Status,
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
		CreatedBy:  role,
		UpdatedBy:  role,
	}
	insertQuery := CassUserSession.Query(userz.OrganizationUnitsTable.Insert()).BindStruct(orgCass)
	if err := insertQuery.ExecRelease(); err != nil {
		return nil, err
	}
	created := strconv.FormatInt(orgCass.CreatedAt, 10)
	updated := strconv.FormatInt(orgCass.UpdatedAt, 10)
	org := &model.OrganizationUnit{
		OuID:       &orgCass.ID,
		OrgID:      orgCass.OrgID,
		EmpCount:   input.EmpCount,
		State:      orgCass.State,
		City:       orgCass.City,
		Address:    orgCass.Address,
		PostalCode: orgCass.PostalCode,
		Country:    orgCass.Country,
		Status:     orgCass.Status,
		CreatedAt:  created,
		CreatedBy:  &orgCass.CreatedBy,
		UpdatedAt:  updated,
		UpdatedBy:  &orgCass.UpdatedBy,
	}
	return org, nil
}

func UpdateOrganizationUnit(ctx context.Context, input model.OrganizationUnitInput) (*model.OrganizationUnit, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if input.OuID == nil {
		return nil, fmt.Errorf("org unit id is required")
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
	if strings.ToLower(role) != "puneet@zicops.com" {
		return nil, fmt.Errorf("user is a not zicops admin: Unauthorized")
	}
	orgCass := userz.OrganizationUnits{
		ID:    *input.OuID,
		OrgID: input.OrgID,
	}
	org := []userz.OrganizationUnits{}

	getQueryStr := fmt.Sprintf(`SELECT * from userz.organization_units where id='%s' and org_id='%s' `, orgCass.ID, orgCass.OrgID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
	if err := getQuery.SelectRelease(&org); err != nil {
		return nil, err
	}
	orgCass = org[0]
	updatedCols := []string{}
	empCount := fmt.Sprintf("%d", input.EmpCount)
	if empCount != orgCass.EmpCount {
		orgCass.EmpCount = empCount
		updatedCols = append(updatedCols, "emp_count")
	}
	if input.Address != "" && input.Address != orgCass.Address {
		orgCass.Address = input.Address
		updatedCols = append(updatedCols, "address")
	}
	if input.PostalCode != "" && input.PostalCode != orgCass.PostalCode {
		orgCass.PostalCode = input.PostalCode
		updatedCols = append(updatedCols, "postal_code")
	}
	if input.City != "" && input.City != orgCass.City {
		orgCass.City = input.City
		updatedCols = append(updatedCols, "city")
	}
	if input.State != "" && input.State != orgCass.State {
		orgCass.State = input.State
		updatedCols = append(updatedCols, "state")
	}
	if input.Country != "" && input.Country != orgCass.Country {
		orgCass.Country = input.Country
		updatedCols = append(updatedCols, "country")
	}
	if input.Status != "" && input.Status != orgCass.Status {
		orgCass.Status = input.Status
		updatedCols = append(updatedCols, "status")
	}
	if len(updatedCols) > 0 {
		orgCass.UpdatedAt = time.Now().Unix()
		orgCass.UpdatedBy = role
		updatedCols = append(updatedCols, "updated_at")
		upStms, uNames := userz.OrganizationUnitsTable.Update(updatedCols...)
		updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&orgCass)
		if err := updateQuery.ExecRelease(); err != nil {
			log.Errorf("error updating user: %v", err)
			return nil, err
		}
	}
	created := strconv.FormatInt(orgCass.CreatedAt, 10)
	updated := strconv.FormatInt(orgCass.UpdatedAt, 10)
	result := &model.OrganizationUnit{
		OuID:       &orgCass.ID,
		OrgID:      orgCass.OrgID,
		EmpCount:   input.EmpCount,
		State:      orgCass.State,
		City:       orgCass.City,
		Address:    orgCass.Address,
		PostalCode: orgCass.PostalCode,
		Country:    orgCass.Country,
		Status:     orgCass.Status,
		CreatedAt:  created,
		CreatedBy:  &orgCass.CreatedBy,
		UpdatedAt:  updated,
		UpdatedBy:  &orgCass.UpdatedBy,
	}
	return result, nil

}

func GetOrganizationUnits(ctx context.Context, ouIds []*string) ([]*model.OrganizationUnit, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	outputOrgs := []*model.OrganizationUnit{}
	for _, orgID := range ouIds {
		if orgID == nil {
			continue
		}
		qryStr := fmt.Sprintf(`SELECT * from userz.organization_units where id='%s' ALLOW FILTERING `, *orgID)
		getOrgs := func() (users []userz.OrganizationUnits, err error) {
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
		result := &model.OrganizationUnit{
			OuID:       &orgCass.ID,
			OrgID:      orgCass.OrgID,
			EmpCount:   emptCnt,
			State:      orgCass.State,
			City:       orgCass.City,
			Address:    orgCass.Address,
			PostalCode: orgCass.PostalCode,
			Country:    orgCass.Country,
			Status:     orgCass.Status,
			CreatedAt:  created,
			CreatedBy:  &orgCass.CreatedBy,
			UpdatedAt:  updated,
			UpdatedBy:  &orgCass.UpdatedBy,
		}
		outputOrgs = append(outputOrgs, result)
	}
	return outputOrgs, nil
}

func GetUnitsByOrgID(ctx context.Context, orgID string) ([]*model.OrganizationUnit, error) {
	_, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	outputOrgs := []*model.OrganizationUnit{}
	qryStr := fmt.Sprintf(`SELECT * from userz.organization_units where org_id='%s' ALLOW FILTERING `, orgID)
	getOrgs := func() (users []userz.OrganizationUnits, err error) {
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
		return nil, fmt.Errorf("no org units found")
	}
	for _, orgCas := range orgs {
		orgCass := orgCas
		created := strconv.FormatInt(orgCass.CreatedAt, 10)
		updated := strconv.FormatInt(orgCass.UpdatedAt, 10)
		emptCnt, _ := strconv.Atoi(orgCass.EmpCount)
		result := &model.OrganizationUnit{
			OuID:       &orgCass.ID,
			OrgID:      orgCass.OrgID,
			EmpCount:   emptCnt,
			State:      orgCass.State,
			City:       orgCass.City,
			Address:    orgCass.Address,
			PostalCode: orgCass.PostalCode,
			Country:    orgCass.Country,
			Status:     orgCass.Status,
			CreatedAt:  created,
			CreatedBy:  &orgCass.CreatedBy,
			UpdatedAt:  updated,
			UpdatedBy:  &orgCass.UpdatedBy,
		}
		outputOrgs = append(outputOrgs, result)
	}
	return outputOrgs, nil
}
