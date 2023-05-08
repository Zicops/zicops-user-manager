package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/scylladb/gocqlx/v2"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/lib/identity"
)

func GetCourseAnalyticsDataById(ctx context.Context, courseID *string, status *string) (*model.CourseAnalyticsFacts, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassSession := session

	qryStr := fmt.Sprintf(`SELECT COUNT(*) FROM userz.user_course_map WHERE course_id='%s' `, *courseID)
	if status != nil {
		qryStr += fmt.Sprintf(`  and course_status='%s' `, *status)
	}
	qryStr += ` ALLOW FILTERING`

	getCourseAnalytics := func() (count int, success bool) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return count, iter.Scan(&count)
	}
	count, success := getCourseAnalytics()
	if !success {
		return nil, fmt.Errorf("got error while searching for course analytics")
	}

	res := &model.CourseAnalyticsFacts{
		CourseID: courseID,
		Status:   status,
		Count:    &count,
	}

	return res, nil
}

func GetLearnerDetails(ctx context.Context, courseID *string, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedUserCourseAnalytics, error) {
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if courseID == nil {
		return nil, fmt.Errorf("please send courseId")
	}

	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	var newPage []byte
	if pageCursor != nil && *pageCursor != "" {
		page, err := global.CryptSession.DecryptString(*pageCursor, nil)
		if err != nil {
			return nil, err
		}
		newPage = page
	}
	var newCursor string
	var pageSizeInt int
	if pageSize != nil {
		pageSizeInt = *pageSize
	} else {
		pageSizeInt = 10
	}

	qryStr := fmt.Sprintf(`SELECT * FROM userz.user_course_map WHERE course_id='%s' ALLOW FILTERING`, *courseID)

	getUserCourseMaps := func(page []byte) (userCourseMaps []userz.UserCourse, nextPage []byte, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		q.PageState(page)
		q.PageSize(pageSizeInt)
		iter := q.Iter()
		return userCourseMaps, iter.PageState(), iter.Select(&userCourseMaps)
	}

	userCourseMaps, newPage, err := getUserCourseMaps(newPage)
	if err != nil {
		return nil, err
	}
	if len(newPage) != 0 {
		newCursor, err = global.CryptSession.EncryptAsString(newPage, nil)
		if err != nil {
			return nil, err
		}
	}
	if len(userCourseMaps) == 0 {
		return nil, nil
	}

	res := make([]*model.UserCourseAnalytics, len(userCourseMaps))
	var wg sync.WaitGroup
	for kk, vvv := range userCourseMaps {
		k := kk
		vv := vvv
		wg.Add(1)
		go func(i int, v userz.UserCourse) {
			defer wg.Done()
			// name  - user course map - users
			// email - user course map - users
			name, email, err := getUserDetail(ctx, CassUserSession, v.UserID)
			if err != nil {
				log.Printf("found err in GetLearnerDetails: %v", err)
				return
			}

			//completion - all topics iteration and decide
			completion, err := checkCompletionOfTopics(ctx, CassUserSession, v.ID)
			if err != nil {
				log.Printf("found error in checkCompletionOfTopics: %v", err)
				return
			}

			// assigned_by - user course map
			// assigned_on - user course map
			var added AddedBy
			err = json.Unmarshal([]byte(v.AddedBy), &added)
			if err != nil {
				log.Printf("got error while unmarshalling data: %v", err)
				return
			}
			ao := strconv.Itoa(int(v.CreatedAt))

			//time_taken - user course map = created at - current time or updated atf
			var timeTaken int64
			if v.CourseStatus == "completed" {
				timeTaken = v.UpdatedAt - v.CreatedAt
			} else {
				timeTaken = time.Now().Unix() - v.CreatedAt
			}
			tt := int(timeTaken)

			//timeline_complaint - user course map - end date and status
			var timelineComplaint string
			//if completed then check if user has completed within time line
			if v.CourseStatus != "completed" && v.UpdatedAt <= v.EndDate {
				timelineComplaint = "Yes"
			} else if v.CourseStatus == "completed" {
				//if not completed then check if time is left for user to complete
				if time.Now().Unix()-v.EndDate > 0 {
					//end date has passed
					timelineComplaint = "No"
				} else {
					timelineComplaint = "Yes"
				}
			}

			courseAnalytics := model.UserCourseAnalytics{
				Name:              &name,
				Email:             &email,
				Status:            &v.CourseStatus,
				Completion:        &completion,
				AssignedBy:        &added.Role,
				AssignedOn:        &ao,
				TimeTaken:         &tt,
				TimelineComplaint: &timelineComplaint,
			}
			res[i] = &courseAnalytics
		}(k, vv)
	}
	wg.Wait()
	outputRes := model.PaginatedUserCourseAnalytics{
		Data:       res,
		PageCursor: &newCursor,
		Direction:  direction,
		PageSize:   pageSize,
	}
	return &outputRes, nil
}

func getUserDetail(ctx context.Context, CassUserSession *gocqlx.Session, userId string) (string, string, error) {
	qryStr := fmt.Sprintf(`SELECT * FROM userz.users WHERE id='%s'`, userId)
	getUser := func() (userDetails []userz.User, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return userDetails, iter.Select(&userDetails)
	}
	user, err := getUser()
	if err != nil {
		return "", "", err
	}
	if len(user) == 0 {
		return "", "", fmt.Errorf("no user found for given userId")
	}
	name := user[0].FirstName + " " + user[0].LastName
	return name, user[0].Email, nil
}

func checkCompletionOfTopics(ctx context.Context, CassUserSession *gocqlx.Session, userCourseId string) (int, error) {
	qryStr := fmt.Sprintf(`SELECT * FROM userz.user_course_progress WHERE user_cm_id='%s' ALLOW FILTERING`, userCourseId)
	getTopics := func() (courseProgress []userz.UserCourseProgress, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return courseProgress, iter.Select(&courseProgress)
	}
	progressMap, err := getTopics()
	if err != nil {
		return 0, err
	}
	if len(progressMap) == 0 {
		return 0, nil
	}

	var total float32
	for _, vv := range progressMap {
		v := vv
		if v.Status == "completed" {
			total = total + 100
		}
	}
	l := float32(len(progressMap))
	res := total / l

	return int(res), nil
}
