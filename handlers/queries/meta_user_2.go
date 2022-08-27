package queries

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/helpers"
)

func GetLatestCohorts(ctx context.Context, userID *string, userLspID *string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedCohorts, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	if userID != nil {
		emailCreatorID = *userID
	}
	var newPage []byte
	//var pageDirection string
	var pageSizeInt int
	if pageCursor != nil && *pageCursor != "" {
		page, err := global.CryptSession.DecryptString(*pageCursor, nil)
		if err != nil {
			return nil, fmt.Errorf("invalid page cursor: %v", err)
		}
		newPage = page
	}
	if pageSize == nil {
		pageSizeInt = 10
	} else {
		pageSizeInt = *pageSize
	}
	var newCursor string
	lspClause := ""
	if userLspID != nil {
		lspClause = fmt.Sprintf(" and user_lsp_id='%s'", *userLspID)
	}
	qryStr := fmt.Sprintf(`SELECT * from userz.user_cohort_map where user_id='%s' and updated_at <= %d %s ALLOW FILTERING`, emailCreatorID, *publishTime, lspClause)
	getUsers := func(page []byte) (users []userz.UserCohort, nextPage []byte, err error) {
		q := global.CassUserSession.Session.Query(qryStr, nil)
		defer q.Release()
		q.PageState(page)
		q.PageSize(pageSizeInt)

		iter := q.Iter()
		return users, iter.PageState(), iter.Select(&users)
	}
	usersCohort, newPage, err := getUsers(newPage)
	if err != nil {
		return nil, err
	}
	if len(newPage) != 0 {
		newCursor, err = global.CryptSession.EncryptAsString(newPage, nil)
		if err != nil {
			return nil, fmt.Errorf("error encrypting cursor: %v", err)
		}
		log.Infof("Users: %v", string(newCursor))

	}
	var outputResponse model.PaginatedCohorts
	allUsers := make([]*model.UserCohort, 0)
	for _, copiedUser := range usersCohort {
		cohortCopy := copiedUser
		createdAt := strconv.FormatInt(cohortCopy.CreatedAt, 10)
		updatedAt := strconv.FormatInt(cohortCopy.UpdatedAt, 10)
		userCohort := &model.UserCohort{
			UserID:           cohortCopy.UserID,
			UserLspID:        cohortCopy.UserLspID,
			UserCohortID:     &cohortCopy.ID,
			CohortID:         cohortCopy.CohortID,
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
			CreatedBy:        &cohortCopy.CreatedBy,
			UpdatedBy:        &cohortCopy.UpdatedBy,
			AddedBy:          cohortCopy.AddedBy,
			MembershipStatus: cohortCopy.MembershipStatus,
		}
		allUsers = append(allUsers, userCohort)
	}
	outputResponse.Cohorts = allUsers
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction
	return &outputResponse, nil
}
