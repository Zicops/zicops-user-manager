package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"sync"
	t "time"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/lib/identity"
)

func AddUserTotalWatchTime(ctx context.Context, input *model.CourseWatchTimeInput) (*bool, error) {
	if input.CourseID == nil || input.Date == nil {
		return nil, fmt.Errorf("please enter course id as well as date")
	}
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email := claims["email"].(string)
	userIdStr := base64.URLEncoding.EncodeToString([]byte(email))
	if input.UserID != nil && *input.UserID != "" {
		userIdStr = *input.UserID
	}
	timeInt := 15
	if input.Time != nil && *input.Time != 0 {
		timeInt = *input.Time
	}

	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	qryStr := fmt.Sprintf(`SELECT * FROM userz.user_course_views WHERE course_id='%s' AND date_value='%s' and users='%s' ALLOW FILTERING`, *input.CourseID, *input.Date, userIdStr)
	getUserViewTime := func() (userCourse []userz.UserCourseViews, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return userCourse, iter.Select(&userCourse)
	}

	userViewTimes, err := getUserViewTime()
	if err != nil {
		return nil, err
	}
	if len(userViewTimes) == 0 {
		//create new mapping
		watchTime := userz.UserCourseViews{
			CourseId:  *input.CourseID,
			DateValue: *input.Date,
			Time:      int64(timeInt),
			CreatedAt: t.Now().Unix(),
			Users:     userIdStr,
		}
		if input.TopicID != nil {
			watchTime.TopicId = *input.TopicID
		}
		if input.Category != nil {
			watchTime.Category = *input.Category
		}
		if input.SubCategories != nil {
			var tmp []string
			for _, vv := range input.SubCategories {
				if vv == nil {
					continue
				}
				v := vv
				tmp = append(tmp, *v)
			}
			watchTime.SubCategories = tmp
		}

		insertQuery := CassUserSession.Query(userz.UserCourseViewsTable.Insert()).BindStruct(watchTime)
		if err = insertQuery.Exec(); err != nil {
			return nil, err
		}
	} else {
		//already existing map, add the time in that
		userViewTime := userViewTimes[0]
		userViewTime.Time = userViewTime.Time + int64(timeInt)
		stmt, names := userz.UserCourseViewsTable.Update("time")
		updatedQuery := CassUserSession.Query(stmt, names).BindStruct(&userViewTime)
		if err = updatedQuery.ExecRelease(); err != nil {
			log.Printf("Got error while updating watch time: %v", err)
			return nil, err
		}
	}

	//update course_consumption_stats' course id total time column
	query := fmt.Sprintf(`SELECT * FROM userz.course_consumption_stats WHERE course_id='%s' ALLOW FILTERING`, *input.CourseID)
	getCourseConsumption := func() (cc []userz.CCStats, err error) {
		q := CassUserSession.Query(query, nil)
		defer q.Release()
		iter := q.Iter()
		return cc, iter.Select(&cc)
	}

	ccStats, err := getCourseConsumption()
	if err != nil {
		return nil, err
	}
	if len(ccStats) == 0 {
		return nil, fmt.Errorf("course consumption stats for this course does not exist")
	}
	ccStat := ccStats[0]
	ccStat.TotalWatchTime = ccStat.TotalWatchTime + int64(timeInt)
	stmt, names := userz.CCTable.Update("total_time")
	updatedQuery := CassUserSession.Query(stmt, names).BindStruct(&ccStat)
	if err = updatedQuery.ExecRelease(); err != nil {
		return nil, err
	}

	res := true
	return &res, nil
}

func GetCourseWatchTime(ctx context.Context, courseID *string, startDate *string, endDate *string) ([]*model.CourseWatchTime, error) {
	if courseID == nil || startDate == nil || endDate == nil {
		return nil, fmt.Errorf("please mention course id, start date and end date")
	}
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	qryStr := fmt.Sprintf(`SELECT * FROM userz.user_course_views WHERE course_id='%s' AND date_value <= '%s' AND date_value >= '%s' ALLOW FILTERING`, *courseID, *endDate, *startDate)
	getTotalWatchTime := func() (watchTime []userz.UserCourseViews, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return watchTime, iter.Select(&watchTime)
	}

	totalWatchTime, err := getTotalWatchTime()
	if err != nil {
		return nil, err
	}

	if len(totalWatchTime) == 0 {
		return nil, nil
	}

	res := make([]*model.CourseWatchTime, len(totalWatchTime))
	var wg sync.WaitGroup
	for kk, vvv := range totalWatchTime {
		vv := vvv
		wg.Add(1)
		go func(k int, v userz.UserCourseViews) {
			defer wg.Done()
			t := int(v.Time)
			ca := strconv.Itoa(int(v.CreatedAt))
			var arr []*string
			for _, vv := range v.SubCategories {
				v := vv
				arr = append(arr, &v)
			}
			tmp := model.CourseWatchTime{
				CourseID:      &v.CourseId,
				Date:          &v.DateValue,
				Time:          &t,
				CreatedAt:     &ca,
				User:          &v.Users,
				Category:      &v.Category,
				TopicID:       &v.TopicId,
				SubCategories: arr,
			}

			res[k] = &tmp
		}(kk, vv)
	}
	wg.Wait()

	return res, nil
}

func GetCourseTotalWatchTime(ctx context.Context, courseID *string) (*float64, error) {
	if courseID == nil {
		return nil, fmt.Errorf("please enter course id")
	}
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	qryStr := fmt.Sprintf(`SELECT * FROM userz.course_consumption_stats WHERE course_id='%s' ALLOW FILTERING`, *courseID)
	getStats := func() (maps []userz.CCStats, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return maps, iter.Select(&maps)
	}

	courseStats, err := getStats()
	if err != nil {
		return nil, err
	}

	if len(courseStats) == 0 {
		return nil, nil
	}

	var totalTime int64
	for _, vv := range courseStats {
		v := vv
		totalTime = totalTime + v.TotalWatchTime
	}

	totalTimeF := float64(totalTime)
	return &totalTimeF, nil
}

func GetUserWatchTime(ctx context.Context, userID string, startDate *string, endDate *string) ([]*model.CourseWatchTime, error) {
	if startDate == nil || endDate == nil {
		return nil, fmt.Errorf("please enter start date and end date")
	}
	_, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// session, err := global.CassPool.GetSession(ctx, "userz")
	// if err != nil {
	// 	return nil, err
	// }
	// CassUserSession := session
	// qryStr := fmt.Sprintf(`SELECT * FROM userz.user_course_views WHERE users='%s' AND date_value <= '%s' AND date_value >= '%s' ALLOW FILTERING`, userID, *endDate, *startDate)
	// getData := func() (datas []userz.UserCourseViews, err error) {
	// 	q := CassUserSession.Query(qryStr, nil)
	// 	defer q.Release()
	// 	iter := q.Iter()
	// 	return datas, iter.Select(&datas)
	// }
	return nil, nil
}
