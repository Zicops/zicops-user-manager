package queries

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/helpers"
)

func GetUserCourseProgressByMapID(ctx context.Context, userId string, userCourseIDs []string) ([]*model.UserCourseProgress, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	if userId != "" {
		emailCreatorID = userId
	}
	//key := "GetUserCourseProgressByMapID" + emailCreatorID + userCourseID
	//result, err := redis.GetRedisValue(key)
	//if err == nil {
	//	var outputResponse []*model.UserCourseProgress
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
	userCPsMap := make([]*model.UserCourseProgress, 0)
	for _, userCourseID := range userCourseIDs {
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
		for _, copiedCP := range userCPs {
			userCP := copiedCP
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
			userCPsMap = append(userCPsMap, currentUserCP)
		}
	}
	//redisBytes, err := json.Marshal(userCPsMap)
	//if err == nil {
	//	redis.SetTTL(key, 300)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
	return userCPsMap, nil
}

func GetUserCourseProgressByTopicID(ctx context.Context, userId string, topicID string) ([]*model.UserCourseProgress, error) {
	claims, err := helpers.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	email_creator := claims["email"].(string)
	emailCreatorID := base64.URLEncoding.EncodeToString([]byte(email_creator))
	if userId != "" {
		emailCreatorID = userId
	}
	//key := "GetUserCourseProgressByTopicID" + emailCreatorID + topicID
	//result, err := redis.GetRedisValue(key)
	//if err == nil {
	//	var outputResponse []*model.UserCourseProgress
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

	qryStr := fmt.Sprintf(`SELECT * from userz.user_course_progress where user_id='%s' and topic_id='%s'  ALLOW FILTERING`, emailCreatorID, topicID)
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
	userCPsMap := make([]*model.UserCourseProgress, 0)
	for _, copiedCP := range userCPs {
		userCP := copiedCP
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
		userCPsMap = append(userCPsMap, currentUserCP)
	}
	//redisBytes, err := json.Marshal(userCPsMap)
	//if err == nil {
	//	redis.SetTTL(key, 300)
	//	redis.SetRedisValue(key, string(redisBytes))
	//}
	return userCPsMap, nil
}
