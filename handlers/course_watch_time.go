package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	t "time"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-user-manager/global"
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

	qryStr := fmt.Sprintf(`SELECT * FROM userz.user_course_views WHERE course_id='%s' AND date_value='%s' ALLOW FILTERING`, *courseID, *date)
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
			Users:     []string{userIdStr},
		}

		insertQuery := CassUserSession.Query(userz.UserCourseViewsTable.Insert()).BindStruct(watchTime)
		if err = insertQuery.Exec(); err != nil {
			return nil, err
		}
	} else {
		//already existing map, add the time in that
		userViewTime := userViewTimes[0]
		users := addUsers(userViewTime.Users, userIdStr)
		userViewTime.Users = users
		userViewTime.Time = userViewTime.Time + int64(timeInt)
		updatedCols := []string{"users", "time"}
		stmt, names := userz.UserCourseViewsTable.Update(updatedCols...)
		updatedQuery := CassUserSession.Query(stmt, names).BindStruct(&userViewTime)
		if err = updatedQuery.ExecRelease(); err != nil {
			log.Printf("Got error while updating watch time: %v", err)
			return nil, err
		}
	}

	res := true
	return &res, nil
}

func addUsers(users []string, newUser string) []string {
	isPresent := false
	for _, vv := range users {
		v := vv
		if v == newUser {
			isPresent = true
		}
	}
	if isPresent {
		return users
	}
	users = append(users, newUser)
	return users
}
