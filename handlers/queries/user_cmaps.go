package queries

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/helpers"
)

func GetUserCourseMaps(ctx context.Context, userId string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedCourseMaps, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	if userId != "" {
		emailCreatorID = userId
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
	//key := "GetUserCourseMaps" + emailCreatorID + string(newPage)
	//result, err := redis.GetRedisValue(key)
	//if err == nil {
	//	var outputResponse model.PaginatedCourseMaps
	//	err = json.Unmarshal([]byte(result), &outputResponse)
	//	if err == nil {
	//		return &outputResponse, nil
	//	}
	//}
	if pageSize == nil {
		pageSizeInt = 10
	} else {
		pageSizeInt = *pageSize
	}
	var newCursor string

	qryStr := fmt.Sprintf(`SELECT * from userz.user_course_map where user_id='%s' and created_at <= %d  ALLOW FILTERING`, emailCreatorID, *publishTime)
	getUsers := func(page []byte) (courses []userz.UserCourse, nextPage []byte, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		q.PageState(page)
		q.PageSize(pageSizeInt)

		iter := q.Iter()
		return courses, iter.PageState(), iter.Select(&courses)
	}
	userCourses, newPage, err := getUsers(newPage)
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
	var outputResponse model.PaginatedCourseMaps
	allCourses := make([]*model.UserCourse, 0)
	for _, copiedCourse := range userCourses {
		courseCopy := copiedCourse
		endDate := strconv.FormatInt(courseCopy.EndDate, 10)
		createdAt := strconv.FormatInt(courseCopy.CreatedAt, 10)
		updatedAt := strconv.FormatInt(courseCopy.UpdatedAt, 10)
		currentCourse := &model.UserCourse{
			UserCourseID: &courseCopy.ID,
			UserID:       courseCopy.UserID,
			UserLspID:    courseCopy.UserLspID,
			CourseID:     courseCopy.CourseID,
			CourseType:   courseCopy.CourseType,
			AddedBy:      courseCopy.AddedBy,
			IsMandatory:  courseCopy.IsMandatory,
			EndDate:      &endDate,
			CourseStatus: courseCopy.CourseStatus,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
			CreatedBy:    &courseCopy.CreatedBy,
			UpdatedBy:    &courseCopy.UpdatedBy,
		}
		allCourses = append(allCourses, currentCourse)
	}
	outputResponse.UserCourses = allCourses
	outputResponse.PageCursor = &newCursor
	outputResponse.PageSize = &pageSizeInt
	outputResponse.Direction = direction
	//redisBytes, err := json.Marshal(outputResponse)
	//if err == nil {
	//	redis.SetTTL(key, 90)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
	return &outputResponse, nil
}

func GetUserCourseMapByCourseID(ctx context.Context, userId string, courseID string) ([]*model.UserCourse, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	if userId != "" {
		emailCreatorID = userId
	}
	//key := "GetUserCourseMapByCourseID" + emailCreatorID + courseID
	//result, err := redis.GetRedisValue(key)
	//if err == nil {
	//	var outputResponse []*model.UserCourse
	//	err = json.Unmarshal([]byte(result), &outputResponse)
	//	if err == nil {
	//		return outputResponse, nil
	//	}
	//}
	session, err := cassandra.GetCassSession("userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session
	createdAt := time.Now().Unix()
	qryStr := fmt.Sprintf(`SELECT * from userz.user_course_map where user_id='%s' and course_id='%s' AND created_at < %d ALLOW FILTERING`, emailCreatorID, courseID, createdAt)
	getUsers := func() (courses []userz.UserCourse, err error) {
		q := CassUserSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return courses, iter.Select(&courses)
	}
	userCourses, err := getUsers()
	if err != nil {
		return nil, err
	}
	if len(userCourses) == 0 {
		return nil, fmt.Errorf("no user course found with id %s", courseID)
	}
	allCourses := make([]*model.UserCourse, 0)
	for _, copiedCourse := range userCourses {
		courseCopy := copiedCourse
		endDate := strconv.FormatInt(courseCopy.EndDate, 10)
		createdAt := strconv.FormatInt(courseCopy.CreatedAt, 10)
		updatedAt := strconv.FormatInt(courseCopy.UpdatedAt, 10)
		currentCourse := &model.UserCourse{
			UserCourseID: &courseCopy.ID,
			UserID:       courseCopy.UserID,
			UserLspID:    courseCopy.UserLspID,
			CourseID:     courseCopy.CourseID,
			CourseType:   courseCopy.CourseType,
			AddedBy:      courseCopy.AddedBy,
			IsMandatory:  courseCopy.IsMandatory,
			EndDate:      &endDate,
			CourseStatus: courseCopy.CourseStatus,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
			CreatedBy:    &courseCopy.CreatedBy,
			UpdatedBy:    &courseCopy.UpdatedBy,
		}
		allCourses = append(allCourses, currentCourse)
	}
	//redisBytes, err := json.Marshal(allCourses)
	//if err == nil {
	//	redis.SetTTL(key, 90)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
	return allCourses, nil
}
