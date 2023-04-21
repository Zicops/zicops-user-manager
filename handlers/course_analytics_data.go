package handlers

import (
	"context"
	"fmt"


	// log "github.com/sirupsen/logrus"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/lib/identity"
)

func GetCourseAnalyticsDataById(ctx context.Context, courseID *string, status *string) (*model.CourseAnalyticsFacts, error){
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
	qryStr += fmt.Sprintf(` ALLOW FILTERING`)

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
		CourseID:     courseID,
		Status:       status,
		Count:        &count,
	}

	return res, nil
}