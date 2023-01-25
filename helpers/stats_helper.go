package helpers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/scylladb/gocqlx/v2"
	"github.com/zicops/contracts/coursez"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/cassandra"
)

func UpdateCCStats(ctx context.Context, session *gocqlx.Session, courseId string, userId string, status string, newAdd bool, completionTime int64) {
	cSessionLocal, err := cassandra.GetCassSession("coursez")
	if err != nil {
		fmt.Println("error getting cass session", err)
		return
	}
	course := coursez.Course{}
	qryStrGetCourse := fmt.Sprintf("SELECT * from coursez.course WHERE id='%s' ALLOW FILTERING ", courseId)
	qryGetCourse := cSessionLocal.Query(qryStrGetCourse, nil)
	if err := qryGetCourse.SelectRelease(&course); err != nil {
		fmt.Println("error getting course", err)
		return
	}
	if course.ID == "" {
		fmt.Println("course not found")
		return
	}
	lspId := course.LspId
	qryStrGet := fmt.Sprintf("SELECT * from userz.course_consumption_stats WHERE lsp_id='%s' AND course_id='%s' ", lspId, courseId)
	qryGet := session.Query(qryStrGet, nil)
	ccStats := []userz.CCStats{}
	if err := qryGet.SelectRelease(&ccStats); err != nil {
		fmt.Println("error getting cc stats", err)
		return
	}
	if len(ccStats) == 0 {
		// create new record
		expectCompletionInt, _ := strconv.ParseInt(course.ExpectedCompletion, 10, 64)
		ccStats := userz.CCStats{
			ID:                     uuid.New().String(),
			LspId:                  lspId,
			CourseId:               courseId,
			ExpectedCompletionTime: expectCompletionInt,
			AverageComplianceScore: 0,
			AverageCompletionTime:  0,
			Duration:               int64(course.Duration),
			Owner:                  course.Owner,
			Category:               course.Category,
			SubCategory:            course.SubCategory,
			TotalLearners:          0,
			ActiveLearners:         0,
			CompletedLearners:      0,
			CreatedAt:              time.Now().Unix(),
			UpdatedAt:              time.Now().Unix(),
			CreatedBy:              userId,
			UpdatedBy:              userId,
		}
		if status == "completed" {
			ccStats.CompletedLearners = ccStats.CompletedLearners + 1
			ccStats.ActiveLearners = ccStats.ActiveLearners - 1
		} else {
			ccStats.ActiveLearners = ccStats.ActiveLearners + 1
		}
		ccStats.TotalLearners = ccStats.CompletedLearners + ccStats.ActiveLearners
		insertQuery := session.Query(userz.CCTable.Insert()).BindStruct(ccStats)
		if err := insertQuery.ExecRelease(); err != nil {
			return
		}
	} else {
		// update existing record
		ccStats := ccStats[0]
		ccStats.UpdatedBy = userId
		ccStats.UpdatedAt = time.Now().Unix()
		isCompleted := false
		if status == "completed" {
			isCompleted = true
			ccStats.CompletedLearners = ccStats.CompletedLearners + 1
			ccStats.ActiveLearners = ccStats.ActiveLearners - 1
		}
		if newAdd {
			ccStats.ActiveLearners = ccStats.ActiveLearners + 1
		}
		if isCompleted {
			ccStats.AverageCompletionTime = (ccStats.AverageCompletionTime + completionTime) / ccStats.CompletedLearners
			compliance_score := 100 - (ccStats.AverageCompletionTime-ccStats.ExpectedCompletionTime)/ccStats.ExpectedCompletionTime
			ccStats.AverageComplianceScore = compliance_score
		}
		ccStats.TotalLearners = ccStats.CompletedLearners + ccStats.ActiveLearners
		insertQuery := session.Query(userz.CCTable.Insert()).BindStruct(ccStats)
		if err := insertQuery.ExecRelease(); err != nil {
			return
		}
	}
}
