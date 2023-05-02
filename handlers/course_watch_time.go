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

func AddUserTotalWatchTime(ctx context.Context, userID *string, courseID *string, time *int, date *string) (*bool, error) {
	if courseID == nil || date == nil {
		return nil, fmt.Errorf("please enter course id as well as date")
	}
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email := claims["email"].(string)
	userIdStr := base64.URLEncoding.EncodeToString([]byte(email))
	if userID != nil && *userID != "" {
		userIdStr = *userID
	}
	timeInt := 15
	if time != nil && *time != 0 {
		timeInt = *time
	}

	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	qryStr := fmt.Sprintf(`SELECT * FROM userz.user_course_views WHERE course_id='%s' AND date_value='%s' and users='%s' ALLOW FILTERING`, *courseID, *date, userIdStr)
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
			CourseId:  *courseID,
			DateValue: *date,
			Time:      int64(timeInt),
			CreatedAt: t.Now().Unix(),
			Users:     userIdStr,
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
	query := fmt.Sprintf(`SELECT * FROM userz.course_consumption_stats WHERE course_id='%s' ALLOW FILTERING`, *courseID)
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

//scren share enabled

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
			tmp := model.CourseWatchTime{
				CourseID:  &v.CourseId,
				Date:      &v.DateValue,
				Time:      &t,
				CreatedAt: &ca,
				User:      &v.Users,
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

	totalHours := float64(totalTime) / (60 * 60)

	return &totalHours, nil
}
