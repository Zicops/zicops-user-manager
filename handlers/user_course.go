package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/coursez"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/lib/identity"
	"github.com/zicops/zicops-user-manager/lib/stats"
)

type AddedBy struct {
	UserId string `json:"userId"`
	Role   string `json:"role"`
}

func AddUserCourse(ctx context.Context, input []*model.UserCourseInput) ([]*model.UserCourse, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	/*
		isAllowed := false
		role := strings.ToLower(userCass.Role)
		if userCass.ID == input[0].UserID || role == "admin" || strings.Contains(role, "manager") {
			isAllowed = true
		}
		if !isAllowed {
			return nil, fmt.Errorf("user not allowed to create org mapping")
		}
	*/
	userLspMaps := make([]*model.UserCourse, 0)
	for _, input := range input {

		if input == nil {
			continue
		}
		queryStr := fmt.Sprintf(`SELECT * FROM user_course_map WHERE user_id='%s' AND course_id='%s' AND user_lsp_id='%s' ALLOW FILTERING`, input.UserID, input.CourseID, input.UserLspID)
		getUserCourseMap := func() (maps []userz.UserCourse, err error) {
			q := CassUserSession.Query(queryStr, nil)
			defer q.Release()
			iter := q.Iter()
			return maps, iter.Select(&maps)
		}
		userCourseMap, err := getUserCourseMap()
		if err != nil {
			return nil, err
		}
		if len(userCourseMap) > 0 {
			return nil, nil
		}

		createdBy := userCass.Email
		updatedBy := userCass.Email
		if input.CreatedBy != nil {
			createdBy = *input.CreatedBy
		}
		if input.UpdatedBy != nil {
			updatedBy = *input.UpdatedBy
		}
		var endDate int64
		if input.EndDate != nil {
			endDate, _ = strconv.ParseInt(*input.EndDate, 10, 64)
		}
		userLspMap := userz.UserCourse{
			ID:           uuid.New().String(),
			UserID:       input.UserID,
			LspID:        *input.LspID,
			UserLspID:    input.UserLspID,
			CourseID:     input.CourseID,
			CourseType:   input.CourseType,
			CourseStatus: input.CourseStatus,
			AddedBy:      input.AddedBy,
			IsMandatory:  input.IsMandatory,
			EndDate:      endDate,
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
			CreatedBy:    createdBy,
			UpdatedBy:    updatedBy,
		}
		insertQuery := CassUserSession.Query(userz.UserCourseTable.Insert()).BindStruct(userLspMap)
		if err := insertQuery.ExecRelease(); err != nil {
			return nil, err
		}
		//getcoursetopic - topics
		created := strconv.FormatInt(userLspMap.CreatedAt, 10)
		updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
		userLspOutput := &model.UserCourse{
			UserCourseID: &userLspMap.ID,
			UserLspID:    userLspMap.UserLspID,
			UserID:       userLspMap.UserID,
			CourseID:     userLspMap.CourseID,
			CourseType:   userLspMap.CourseType,
			CourseStatus: userLspMap.CourseStatus,
			AddedBy:      userLspMap.AddedBy,
			IsMandatory:  userLspMap.IsMandatory,
			EndDate:      input.EndDate,
			CreatedAt:    created,
			UpdatedAt:    updated,
			CreatedBy:    &userLspMap.CreatedBy,
			UpdatedBy:    &userLspMap.UpdatedBy,
		}
		userLspMaps = append(userLspMaps, userLspOutput)
		// create or update course consumption stats
		stats.UpdateCCStats(ctx, CassUserSession, *input.LspID, userLspOutput.CourseID, userLspOutput.UserID, userLspOutput.CourseStatus, true, userLspMap.UpdatedAt-userLspMap.CreatedAt, *userLspOutput.EndDate)

	}
	return userLspMaps, nil
}

func UpdateUserCourse(ctx context.Context, input model.UserCourseInput) (*model.UserCourse, error) {
	userCass, err := GetUserFromCass(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	isAllowed := true
	role := strings.ToLower(userCass.Role)
	if userCass.ID == input.UserID || role == "admin" || strings.Contains(role, "manager") {
		isAllowed = true
	}
	if !isAllowed {
		return nil, fmt.Errorf("user not allowed to create org mapping")
	}
	if input.UserCourseID == nil {
		return nil, fmt.Errorf("user course id is required")
	}
	if input.UserID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	userLspMap := userz.UserCourse{
		ID: *input.UserCourseID,
	}
	userLsps := []userz.UserCourse{}

	getQueryStr := fmt.Sprintf("SELECT * FROM userz.user_course_map WHERE id='%s' AND user_id='%s' ALLOW FILTERING", userLspMap.ID, input.UserID)
	getQuery := CassUserSession.Query(getQueryStr, nil)
	if err := getQuery.SelectRelease(&userLsps); err != nil {
		return nil, err
	}
	if len(userLsps) == 0 {
		return nil, fmt.Errorf("users orgs not found")
	}
	userLspMap = userLsps[0]
	updatedCols := []string{}
	if input.EndDate != nil {
		endDate, _ := strconv.ParseInt(*input.EndDate, 10, 64)
		userLspMap.EndDate = endDate
		updatedCols = append(updatedCols, "end_date")
	}
	if input.CourseID != "" && input.CourseID != userLspMap.CourseID {
		userLspMap.CourseID = input.CourseID
		updatedCols = append(updatedCols, "course_id")
	}
	if input.CourseType != "" && input.CourseType != userLspMap.CourseType {
		userLspMap.CourseType = input.CourseType
		updatedCols = append(updatedCols, "course_type")
	}
	if input.CourseStatus != "" && input.CourseStatus != userLspMap.CourseStatus {
		userLspMap.CourseStatus = input.CourseStatus

		updatedCols = append(updatedCols, "course_status")
	}
	if input.AddedBy != "" && input.AddedBy != userLspMap.AddedBy {
		userLspMap.AddedBy = input.AddedBy
		updatedCols = append(updatedCols, "added_by")
	}
	if input.IsMandatory != userLspMap.IsMandatory {
		userLspMap.IsMandatory = input.IsMandatory
		updatedCols = append(updatedCols, "is_mandatory")
	}
	if input.UpdatedBy != nil {
		userLspMap.UpdatedBy = *input.UpdatedBy
		updatedCols = append(updatedCols, "updated_by")
	}
	if input.UserLspID != "" {
		userLspMap.UserLspID = input.UserLspID
		updatedCols = append(updatedCols, "user_lsp_id")
	}
	if input.LspID != nil {
		userLspMap.LspID = *input.LspID
		updatedCols = append(updatedCols, "lsp_id")
	}

	if len(updatedCols) > 0 {
		updatedAt := time.Now().Unix()
		userLspMap.UpdatedAt = updatedAt
		updatedCols = append(updatedCols, "updated_at")
		upStms, uNames := userz.UserCourseTable.Update(updatedCols...)
		updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&userLspMap)
		if err := updateQuery.ExecRelease(); err != nil {
			log.Errorf("error updating user course: %v", err)
			return nil, err
		}
	}
	created := strconv.FormatInt(userLspMap.CreatedAt, 10)
	updated := strconv.FormatInt(userLspMap.UpdatedAt, 10)
	userLspOutput := &model.UserCourse{
		UserCourseID: &userLspMap.ID,
		UserLspID:    userLspMap.UserLspID,
		UserID:       userLspMap.UserID,
		LspID:        &userLspMap.LspID,
		CourseID:     userLspMap.CourseID,
		CourseType:   userLspMap.CourseType,
		CourseStatus: userLspMap.CourseStatus,
		AddedBy:      userLspMap.AddedBy,
		IsMandatory:  userLspMap.IsMandatory,
		EndDate:      input.EndDate,
		CreatedAt:    created,
		UpdatedAt:    updated,
		CreatedBy:    &userLspMap.CreatedBy,
		UpdatedBy:    &userLspMap.UpdatedBy,
	}
	// create or update course consumption stats
	stats.UpdateCCStats(ctx, CassUserSession, userLspMap.LspID, userLspOutput.CourseID, userLspOutput.UserID, userLspOutput.CourseStatus, false, userLspMap.UpdatedAt-userLspMap.CreatedAt, *userLspOutput.EndDate)
	return userLspOutput, nil
}

func checkStatusOfEachTopic(ctx context.Context, userId string, userCourseId string) bool {

	userCP, err := getUserCourseProgressByUserCourseID(ctx, userId, userCourseId)
	if err != nil {
		log.Errorf("Got error while checking course progress: %v", err)
	}

	for _, vv := range userCP {
		v := vv
		if v == nil {
			continue
		}
		res := *v
		if res.Status != "completed" {
			return false
		}
	}
	return true
}

func getUserCourseProgressByUserCourseID(ctx context.Context, userId string, userCourseID string) ([]*model.UserCourseProgress, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	if userId != "" {
		emailCreatorID = userId
	}

	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	userCPsMap := make([]*model.UserCourseProgress, 0)
	qryStr := fmt.Sprintf(`SELECT * from userz.user_course_progress where user_id='%s' and user_cm_id='%s'  ALLOW FILTERING`, emailCreatorID, userCourseID)
	getUsersCProgress := func() (users []userz.UserCourseProgress, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return users, iter.Select(&users)
	}
	userCPs, err := getUsersCProgress()
	if err != nil {
		return nil, err
	}
	userCPsMapCurrent := make([]*model.UserCourseProgress, len(userCPs))
	if len(userCPs) == 0 {
		return nil, nil
	}
	var wg sync.WaitGroup
	for i, cp := range userCPs {
		uc := cp
		wg.Add(1)
		go func(i int, userCP userz.UserCourseProgress) {
			createdAt := strconv.FormatInt(userCP.CreatedAt, 10)
			updatedAt := strconv.FormatInt(userCP.UpdatedAt, 10)
			timeStamp := strconv.FormatInt(userCP.TimeStamp, 10)
			currentUserCP := &model.UserCourseProgress{
				UserCpID:      &userCP.ID,
				UserID:        userCP.UserID,
				UserCourseID:  userCP.UserCmID,
				TopicID:       userCP.TopicID,
				TopicType:     userCP.TopicType,
				Status:        userCP.Status,
				VideoProgress: userCP.VideoProgress,
				TimeStamp:     timeStamp,
				CreatedBy:     &userCP.CreatedBy,
				UpdatedBy:     &userCP.UpdatedBy,
				CreatedAt:     createdAt,
				UpdatedAt:     updatedAt,
			}
			userCPsMapCurrent[i] = currentUserCP
			wg.Done()
		}(i, uc)
	}
	wg.Wait()
	userCPsMap = append(userCPsMap, userCPsMapCurrent...)
	return userCPsMap, nil
}

func AddUserCohortCourses(ctx context.Context, userIds []string, cohortID string) (*bool, error) {
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		log.Errorf("Got error while getting claims: %v", err)
		return nil, err
	}
	email := claims["email"].(string)
	id := base64.URLEncoding.EncodeToString([]byte(email))

	session, err := global.CassPool.GetSession(ctx, "coursez")
	if err != nil {
		return nil, err
	}
	CassSession := session
	qryStr := fmt.Sprintf(`SELECT * FROM coursez.course_cohort_mapping WHERE cohortid='%s' ALLOW FILTERING`, cohortID)
	getCCourses := func() (maps []coursez.CourseCohortMapping, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return maps, iter.Select(&maps)
	}

	cohortCourses, err := getCCourses()
	if err != nil {
		log.Println("Got error in getting cohort courses: ", err)
		return nil, err
	}
	if len(cohortCourses) == 0 {
		return nil, nil
	}

	session, err = global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	//we have list of courses to be mapped, and list of users with whom we have to map them
	//double loop, pass the parameters, send to goroutines
	//user course map - if exists check role in added by, if cohort/admin leave, if self, change to cohort, change the completion day in user course map
	// to course_cohort_map.expected time + user_course_map.time
	// not then create accordingly

	//if completed, then status = completed, leave as it is
	var res error
	var wg sync.WaitGroup
	for _, cccourses := range cohortCourses {
		ccourse := cccourses
		for _, uuusers := range userIds {
			uuser := uuusers
			//check for map

			wg.Add(1)
			go func(course coursez.CourseCohortMapping, userId string) {
				defer wg.Done()
				//if map does not exist, or added by role is self, then call course_
				queryStr := fmt.Sprintf(`SELECT * FROM userz.user_course_map WHERE user_id='%s' AND course_id='%s' ALLOW FILTERING`, userId, course.CourseID)
				getUserCourseMap := func() (maps []userz.UserCourse, err error) {
					q := CassUserSession.Query(queryStr, nil)
					defer q.Release()
					iter := q.Iter()
					return maps, iter.Select(&maps)
				}
				ucMaps, err := getUserCourseMap()
				if err != nil {
					log.Println("Got error while getting user course map: ", err.Error())
					res = err
					return
				}
				if len(ucMaps) == 0 {
					//create map

					qry := fmt.Sprintf(`SELECT * from userz.user_lsp_map where user_id='%s' and lsp_id='%s'  ALLOW FILTERING`, userId, course.LspId)
					getUsersOrgs := func() (users []userz.UserLsp, err error) {
						q := CassUserSession.Query(qry, nil)
						defer q.Release()
						iter := q.Iter()
						return users, iter.Select(&users)
					}
					usersOrgs, err := getUsersOrgs()
					if err != nil {
						res = err
						return
					}
					if len(usersOrgs) == 0 {
						res = errors.New("no user lsp map found")
						return
					}
					userLspId := usersOrgs[0].ID

					added := AddedBy{
						UserId: id,
						Role:   "cohort",
					}
					addedString, err := json.Marshal(added)
					if err != nil {
						res = err
						return
					}
					end := int64(course.ExpectedCompletionDays) * 24 * 60 * 60
					endString := strconv.Itoa(int(end))
					input := &model.UserCourseInput{
						UserID:       userId,
						LspID:        &course.LspId,
						UserLspID:    userLspId,
						CourseID:     course.CourseID,
						CourseType:   course.CourseType,
						AddedBy:      string(addedString),
						IsMandatory:  course.IsMandatory,
						EndDate:      &endString,
						CourseStatus: "open",
						CreatedBy:    &id,
						UpdatedBy:    &id,
					}

					inp := []*model.UserCourseInput{input}
					_, err = AddUserCourse(ctx, inp)
					if err != nil {
						res = err
						return
					}

				} else {
					ucMap := ucMaps[0]
					added := AddedBy{}
					err = json.Unmarshal([]byte(ucMap.AddedBy), &added)
					if err != nil {
						log.Printf("Got error while unmarshalling: %v", err)
						return
					}

					var updatedCols []string
					if added.Role == "self" {
						//change role to cohort
						added.Role = "cohort"
						addedString, _ := json.Marshal(added)
						ucMap.AddedBy = string(addedString)

						//change the expected completion day to
						ucMap.EndDate = int64(course.ExpectedCompletionDays)*24*60*60 + ucMap.EndDate

						ucMap.UpdatedAt = time.Now().Unix()
						ucMap.UpdatedBy = id
						updatedCols = []string{"added_by", "end_date", "updated_at", "updated_by"}

					}
					if ucMap.CourseStatus == "disable" {
						ucMap.CourseStatus = "open"
						updatedCols = append(updatedCols, "course_status")
					}

					upStms, uNames := userz.UserCourseTable.Update(updatedCols...)
					updateQuery := CassUserSession.Query(upStms, uNames).BindStruct(&ucMap)
					if err := updateQuery.ExecRelease(); err != nil {
						log.Errorf("error updating courses: %v", err)
						return
					}
				}

			}(ccourse, uuser)
		}
	}
	wg.Wait()
	tmp := false
	if res != nil {
		return &tmp, res
	}
	tmp = true

	return &tmp, nil
}
