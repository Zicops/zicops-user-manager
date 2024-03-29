package stats

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/scylladb/gocqlx/v2"
	"github.com/sirupsen/logrus"
	"github.com/zicops/contracts/coursez"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-user-manager/global"
)

func UpdateCCStats(ctx context.Context, session *gocqlx.Session, lspId string, courseId string, userId string, status string, newAdd bool, completionTime int64, expectedCompletion string) {
	cSessionLocal, err := global.CassPool.GetSession(ctx, "coursez")
	if err != nil {
		fmt.Println("error getting cass session", err)
		return
	}
	courses := []coursez.Course{}
	qryStrGetCourse := fmt.Sprintf("SELECT * from coursez.course WHERE id='%s' ALLOW FILTERING ", courseId)
	qryGetCourse := cSessionLocal.Query(qryStrGetCourse, nil)
	if err := qryGetCourse.SelectRelease(&courses); err != nil {
		fmt.Println("error getting course", err)
		return
	}
	if len(courses) == 0 {
		fmt.Println("course not found")
		return
	}
	course := courses[0]
	qryStrGet := fmt.Sprintf("SELECT * from userz.course_consumption_stats WHERE lsp_id='%s' AND course_id='%s' ", lspId, courseId)
	qryGet := session.Query(qryStrGet, nil)
	ccStats := []userz.CCStats{}
	if err := qryGet.SelectRelease(&ccStats); err != nil {
		fmt.Println("error getting cc stats", err)
		return
	}
	if len(ccStats) == 0 {
		//expected completion time different
		// create new record
		expectCompletionInt, _ := strconv.ParseInt(expectedCompletion, 10, 64)
		tmp := expectCompletionInt - time.Now().Unix()
		ccStats := userz.CCStats{
			ID:                     uuid.New().String(),
			LspId:                  lspId,
			CourseId:               courseId,
			ExpectedCompletionTime: tmp,
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

		//compare expected completion time, and if the difference is more than one day, then change the expected completion time to current calculated
		// expected completion time
		expectCompletionInt, _ := strconv.ParseInt(expectedCompletion, 10, 64)
		tmp := expectCompletionInt - time.Now().Unix()
		current_dif := ccStats.ExpectedCompletionTime - tmp
		if math.Abs(float64(current_dif)) > 24*60*60 {
			ccStats.ExpectedCompletionTime = tmp
		}

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
			expectedCompletitionDuration := ccStats.ExpectedCompletionTime
			compliance_score := 100 - ((completionTime - expectedCompletitionDuration) / expectedCompletitionDuration)
			if compliance_score > 100 {
				compliance_score = 100
			}
			ccStats.AverageComplianceScore = compliance_score
		}
		if status == "disable" {
			ccStats.ActiveLearners = ccStats.ActiveLearners - 1
		}
		ccStats.TotalLearners = ccStats.CompletedLearners + ccStats.ActiveLearners
		deleteQry := fmt.Sprintf("DELETE FROM userz.course_consumption_stats WHERE lsp_id='%s' AND course_id='%s' ", lspId, courseId)
		if err := session.Query(deleteQry, nil).ExecRelease(); err != nil {
			fmt.Println("error deleting cc stats", err)
			return
		}
		insertQuery := session.Query(userz.CCTable.Insert()).BindStruct(ccStats)
		if err := insertQuery.ExecRelease(); err != nil {
			fmt.Println("error inserting cc stats", err)
			return
		}

	}
}

func AddUpdateCourseViews(lspId string, userId string, secs int64, oldSecs int64) {
	ctx := context.Background()
	cSessionLocal, err := global.CassPool.GetSession(ctx, "coursez")
	if err != nil {
		logrus.Error("error getting cass session", err)
		return
	}
	currentDateString := time.Now().Format("2006-01-02")
	qryStrGet := fmt.Sprintf("SELECT * from coursez.course_views WHERE lsp_id='%s' AND date_value='%s' ALLOW FILTERING", lspId, currentDateString)
	qryGet := cSessionLocal.Query(qryStrGet, nil)
	courseViews := []coursez.CourseView{}
	if err := qryGet.SelectRelease(&courseViews); err != nil {
		logrus.Error("error getting course views", err)
		return
	}
	if len(courseViews) == 0 {
		// create new record
		courseViews := coursez.CourseView{
			LspId:     lspId,
			DateValue: currentDateString,
			Users:     []string{userId},
			Hours:     secs,
			CreatedAt: time.Now().Unix(),
		}
		insertQuery := cSessionLocal.Query(coursez.CVTable.Insert()).BindStruct(courseViews)
		if err := insertQuery.ExecRelease(); err != nil {
			return
		}
	} else {
		// update existing record
		courseViews := courseViews[0]
		// add userId if not present
		foundUser := false
		for _, user := range courseViews.Users {
			if user == userId {
				foundUser = true
				break
			}
		}
		if !foundUser {
			courseViews.Users = append(courseViews.Users, userId)
		}
		courseViews.Hours = courseViews.Hours + secs - oldSecs
		insertQuery := cSessionLocal.Query(coursez.CVTable.Insert()).BindStruct(courseViews)
		if err := insertQuery.ExecRelease(); err != nil {
			return
		}
	}
}
