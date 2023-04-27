package handlers

import (
	"context"
	"fmt"

	"github.com/zicops/contracts/coursez"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/lib/identity"
)

func GetAssignedCourses(ctx context.Context, lspID *string, typeArg string) (*model.CourseCountStats, error) {
	//coursez.course - all courses , status check - published
	//all courses - pass to user course map and check if they are assigned to anyone
	claims, err := identity.GetClaimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	lsp := claims["lsp_id"].(string)
	if lspID != nil {
		lsp = *lspID
	}

	session, err := global.CassPool.GetSession(ctx, "userz")
	if err != nil {
		return nil, err
	}
	CassUserSession := session

	sessionC, err := global.CassPool.GetSession(ctx, "coursez")
	if err != nil {
		return nil, err
	}
	CassCourseSession := sessionC

	qryStr := fmt.Sprintf(`SELECT * FROM coursez.course WHERE lsp_id='%s' AND type='%s' ALLOW FILTERING`, lsp, typeArg)
	getCourses := func() (coursesData []coursez.Course, err error) {
		q := CassCourseSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return coursesData, iter.Select(&coursesData)
	}

	courses, err := getCourses()
	if err != nil {
		return nil, err
	}
	if len(courses) == 0 {
		return nil, nil
	}

	total := 0
	for _, vv := range courses {
		v := vv
		if v.Status != "PUBLISHED" {
			continue
		}
		query := fmt.Sprintf(`SELECT * FROM userz.user_course_map WHERE course_id='%s' ALLOW FILTERING`, v.ID)
		getUserCourseMap := func() (maps []userz.UserCourse, err error) {
			q := CassUserSession.Query(query, nil)
			defer q.Release()
			iter := q.Iter()
			return maps, iter.Select(&maps)
		}
		userCourse, err := getUserCourseMap()
		if err != nil {
			return nil, err
		}
		if len(userCourse) == 0 {
			return nil, nil
		} else {
			total = total + 1
		}

	}

	res := model.CourseCountStats{
		LspID:      &lsp,
		CourseType: &typeArg,
		Count:      &total,
	}

	return &res, nil
}
