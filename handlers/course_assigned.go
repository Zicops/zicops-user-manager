package handlers

import (
	"context"
	"fmt"

	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/lib/identity"
)

func GetAssignedCourses(ctx context.Context, lspID *string, status string, typeArg string) (*model.CourseCountStats, error) {
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
	CassSession := session
	qryStr := fmt.Sprintf(`SELECT * FROM userz.user_course_map WHERE lsp_id='%s' AND status='%s' course_type='%s' ALLOW FILTERING`, lsp, status, typeArg)
	getCourseMaps := func() (maps []userz.UserCourse, err error) {
		q := CassSession.Query(qryStr, nil)
		defer q.Release()
		iter := q.Iter()
		return maps, iter.Select(&maps)
	}
	courseMaps, err := getCourseMaps()
	if err != nil {
		return nil, err
	}

	tmp := make(map[string]bool, 0)
	for _, vv := range courseMaps {
		v := vv
		if tmp[v.CourseID] {
			continue
		}
		tmp[v.CourseID] = true
	}
	num := len(tmp)
	res := model.CourseCountStats{
		LspID:        &lsp,
		CourseStatus: &status,
		CourseType:   &typeArg,
		Count:        &num,
	}

	return &res, nil
}
