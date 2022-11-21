// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"github.com/99designs/gqlgen/graphql"
)

type CohortMain struct {
	CohortID    *string `json:"cohort_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	LspID       string  `json:"lsp_id"`
	Code        string  `json:"code"`
	Status      string  `json:"status"`
	Type        string  `json:"type"`
	IsActive    bool    `json:"is_active"`
	CreatedBy   *string `json:"created_by"`
	UpdatedBy   *string `json:"updated_by"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	Size        int     `json:"size"`
	ImageURL    *string `json:"imageUrl"`
}

type CohortMainInput struct {
	CohortID    *string         `json:"cohort_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	LspID       string          `json:"lsp_id"`
	Code        string          `json:"code"`
	Status      string          `json:"status"`
	Type        string          `json:"type"`
	IsActive    bool            `json:"is_active"`
	CreatedBy   *string         `json:"created_by"`
	UpdatedBy   *string         `json:"updated_by"`
	Size        int             `json:"size"`
	Image       *graphql.Upload `json:"image"`
	ImageURL    *string         `json:"imageUrl"`
}

type LearningSpace struct {
	LspID      *string   `json:"lsp_id"`
	OrgID      string    `json:"org_id"`
	OuID       string    `json:"ou_id"`
	Name       string    `json:"name"`
	LogoURL    *string   `json:"logo_url"`
	ProfileURL *string   `json:"profile_url"`
	NoUsers    int       `json:"no_users"`
	Owners     []*string `json:"owners"`
	IsDefault  bool      `json:"is_default"`
	Status     string    `json:"status"`
	CreatedAt  string    `json:"created_at"`
	UpdatedAt  string    `json:"updated_at"`
	CreatedBy  *string   `json:"created_by"`
	UpdatedBy  *string   `json:"updated_by"`
}

type LearningSpaceInput struct {
	LspID      *string         `json:"lsp_id"`
	OrgID      string          `json:"org_id"`
	OuID       string          `json:"ou_id"`
	Name       string          `json:"name"`
	LogoURL    *string         `json:"logo_url"`
	Logo       *graphql.Upload `json:"logo"`
	ProfileURL *string         `json:"profile_url"`
	Profile    *graphql.Upload `json:"profile"`
	NoUsers    int             `json:"no_users"`
	Owners     []*string       `json:"owners"`
	IsDefault  bool            `json:"is_default"`
	Status     string          `json:"status"`
	CreatedBy  *string         `json:"created_by"`
	UpdatedBy  *string         `json:"updated_by"`
}

type Organization struct {
	OrgID         *string `json:"org_id"`
	Name          string  `json:"name"`
	LogoURL       *string `json:"logo_url"`
	Industry      string  `json:"industry"`
	Type          string  `json:"type"`
	Subdomain     string  `json:"subdomain"`
	EmployeeCount int     `json:"employee_count"`
	Website       string  `json:"website"`
	LinkedinURL   *string `json:"linkedin_url"`
	FacebookURL   *string `json:"facebook_url"`
	TwitterURL    *string `json:"twitter_url"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	CreatedBy     *string `json:"created_by"`
	UpdatedBy     *string `json:"updated_by"`
}

type OrganizationInput struct {
	OrgID         *string         `json:"org_id"`
	Name          string          `json:"name"`
	LogoURL       *string         `json:"logo_url"`
	Industry      string          `json:"industry"`
	Type          string          `json:"type"`
	Subdomain     string          `json:"subdomain"`
	EmployeeCount int             `json:"employee_count"`
	Website       string          `json:"website"`
	LinkedinURL   *string         `json:"linkedin_url"`
	FacebookURL   *string         `json:"facebook_url"`
	TwitterURL    *string         `json:"twitter_url"`
	Status        string          `json:"status"`
	Logo          *graphql.Upload `json:"logo"`
}

type OrganizationUnit struct {
	OuID       *string `json:"ou_id"`
	OrgID      string  `json:"org_id"`
	EmpCount   int     `json:"emp_count"`
	Address    string  `json:"address"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	Country    string  `json:"country"`
	PostalCode string  `json:"postal_code"`
	Status     string  `json:"status"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
	CreatedBy  *string `json:"created_by"`
	UpdatedBy  *string `json:"updated_by"`
}

type OrganizationUnitInput struct {
	OuID       *string `json:"ou_id"`
	OrgID      string  `json:"org_id"`
	EmpCount   int     `json:"emp_count"`
	Address    string  `json:"address"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	Country    string  `json:"country"`
	PostalCode string  `json:"postal_code"`
	Status     string  `json:"status"`
	CreatedBy  *string `json:"created_by"`
	UpdatedBy  *string `json:"updated_by"`
}

type PaginatedBookmarks struct {
	Bookmarks  []*UserBookmark `json:"bookmarks"`
	PageCursor *string         `json:"pageCursor"`
	Direction  *string         `json:"direction"`
	PageSize   *int            `json:"pageSize"`
}

type PaginatedCohorts struct {
	Cohorts    []*UserCohort `json:"cohorts"`
	PageCursor *string       `json:"pageCursor"`
	Direction  *string       `json:"direction"`
	PageSize   *int          `json:"pageSize"`
}

type PaginatedCohortsMain struct {
	Cohorts    []*CohortMain `json:"cohorts"`
	PageCursor *string       `json:"pageCursor"`
	Direction  *string       `json:"direction"`
	PageSize   *int          `json:"pageSize"`
}

type PaginatedCourseMaps struct {
	UserCourses []*UserCourse `json:"user_courses"`
	PageCursor  *string       `json:"pageCursor"`
	Direction   *string       `json:"direction"`
	PageSize    *int          `json:"pageSize"`
}

type PaginatedNotes struct {
	Notes      []*UserNotes `json:"notes"`
	PageCursor *string      `json:"pageCursor"`
	Direction  *string      `json:"direction"`
	PageSize   *int         `json:"pageSize"`
}

type PaginatedUserLspMaps struct {
	UserLspMaps []*UserLspMap `json:"user_lsp_maps"`
	PageCursor  *string       `json:"pageCursor"`
	Direction   *string       `json:"direction"`
	PageSize    *int          `json:"pageSize"`
}

type PaginatedUsers struct {
	Users      []*User `json:"users"`
	PageCursor *string `json:"pageCursor"`
	Direction  *string `json:"direction"`
	PageSize   *int    `json:"pageSize"`
}

type User struct {
	ID         *string `json:"id"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	Status     string  `json:"status"`
	Role       string  `json:"role"`
	IsVerified bool    `json:"is_verified"`
	IsActive   bool    `json:"is_active"`
	Gender     string  `json:"gender"`
	CreatedBy  *string `json:"created_by"`
	UpdatedBy  *string `json:"updated_by"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
	Email      string  `json:"email"`
	Phone      string  `json:"phone"`
	PhotoURL   *string `json:"photo_url"`
}

type UserBookmark struct {
	UserBmID     *string `json:"user_bm_id"`
	UserID       string  `json:"user_id"`
	UserLspID    string  `json:"user_lsp_id"`
	UserCourseID string  `json:"user_course_id"`
	CourseID     string  `json:"course_id"`
	ModuleID     string  `json:"module_id"`
	TopicID      string  `json:"topic_id"`
	Name         string  `json:"name"`
	TimeStamp    string  `json:"time_stamp"`
	IsActive     bool    `json:"is_active"`
	CreatedBy    *string `json:"created_by"`
	UpdatedBy    *string `json:"updated_by"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type UserBookmarkInput struct {
	UserBmID     *string `json:"user_bm_id"`
	UserID       string  `json:"user_id"`
	UserLspID    string  `json:"user_lsp_id"`
	UserCourseID string  `json:"user_course_id"`
	CourseID     string  `json:"course_id"`
	ModuleID     string  `json:"module_id"`
	TopicID      string  `json:"topic_id"`
	Name         string  `json:"name"`
	TimeStamp    string  `json:"time_stamp"`
	IsActive     bool    `json:"is_active"`
	CreatedBy    *string `json:"created_by"`
	UpdatedBy    *string `json:"updated_by"`
}

type UserCohort struct {
	UserCohortID     *string `json:"user_cohort_id"`
	UserID           string  `json:"user_id"`
	UserLspID        string  `json:"user_lsp_id"`
	CohortID         string  `json:"cohort_id"`
	AddedBy          string  `json:"added_by"`
	MembershipStatus string  `json:"membership_status"`
	Role             string  `json:"role"`
	CreatedBy        *string `json:"created_by"`
	UpdatedBy        *string `json:"updated_by"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

type UserCohortInput struct {
	UserCohortID     *string `json:"user_cohort_id"`
	UserID           string  `json:"user_id"`
	UserLspID        string  `json:"user_lsp_id"`
	CohortID         string  `json:"cohort_id"`
	AddedBy          string  `json:"added_by"`
	MembershipStatus string  `json:"membership_status"`
	Role             string  `json:"role"`
	CreatedBy        *string `json:"created_by"`
	UpdatedBy        *string `json:"updated_by"`
}

type UserCourse struct {
	UserCourseID *string `json:"user_course_id"`
	UserID       string  `json:"user_id"`
	UserLspID    string  `json:"user_lsp_id"`
	CourseID     string  `json:"course_id"`
	CourseType   string  `json:"course_type"`
	AddedBy      string  `json:"added_by"`
	IsMandatory  bool    `json:"is_mandatory"`
	EndDate      *string `json:"end_date"`
	CourseStatus string  `json:"course_status"`
	CreatedBy    *string `json:"created_by"`
	UpdatedBy    *string `json:"updated_by"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type UserCourseInput struct {
	UserCourseID *string `json:"user_course_id"`
	UserID       string  `json:"user_id"`
	UserLspID    string  `json:"user_lsp_id"`
	CourseID     string  `json:"course_id"`
	CourseType   string  `json:"course_type"`
	AddedBy      string  `json:"added_by"`
	IsMandatory  bool    `json:"is_mandatory"`
	EndDate      *string `json:"end_date"`
	CourseStatus string  `json:"course_status"`
	CreatedBy    *string `json:"created_by"`
	UpdatedBy    *string `json:"updated_by"`
}

type UserCourseProgress struct {
	UserCpID      *string `json:"user_cp_id"`
	UserID        string  `json:"user_id"`
	UserCourseID  string  `json:"user_course_id"`
	TopicID       string  `json:"topic_id"`
	TopicType     string  `json:"topic_type"`
	Status        string  `json:"status"`
	VideoProgress string  `json:"video_progress"`
	TimeStamp     string  `json:"time_stamp"`
	CreatedBy     *string `json:"created_by"`
	UpdatedBy     *string `json:"updated_by"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type UserCourseProgressInput struct {
	UserCpID      *string `json:"user_cp_id"`
	UserID        string  `json:"user_id"`
	UserCourseID  string  `json:"user_course_id"`
	TopicID       string  `json:"topic_id"`
	TopicType     string  `json:"topic_type"`
	Status        string  `json:"status"`
	VideoProgress string  `json:"video_progress"`
	TimeStamp     string  `json:"time_stamp"`
	CreatedBy     *string `json:"created_by"`
	UpdatedBy     *string `json:"updated_by"`
}

type UserExamAttempts struct {
	UserEaID         *string `json:"user_ea_id"`
	UserID           string  `json:"user_id"`
	UserLspID        string  `json:"user_lsp_id"`
	UserCpID         string  `json:"user_cp_id"`
	UserCourseID     string  `json:"user_course_id"`
	ExamID           string  `json:"exam_id"`
	AttemptNo        int     `json:"attempt_no"`
	AttemptStatus    string  `json:"attempt_status"`
	AttemptStartTime string  `json:"attempt_start_time"`
	AttemptDuration  string  `json:"attempt_duration"`
	CreatedBy        *string `json:"created_by"`
	UpdatedBy        *string `json:"updated_by"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

type UserExamAttemptsInput struct {
	UserEaID         *string `json:"user_ea_id"`
	UserID           string  `json:"user_id"`
	UserLspID        string  `json:"user_lsp_id"`
	UserCpID         string  `json:"user_cp_id"`
	UserCourseID     string  `json:"user_course_id"`
	ExamID           string  `json:"exam_id"`
	AttemptNo        int     `json:"attempt_no"`
	AttemptStatus    string  `json:"attempt_status"`
	AttemptStartTime string  `json:"attempt_start_time"`
	AttemptDuration  string  `json:"attempt_duration"`
	CreatedBy        *string `json:"created_by"`
	UpdatedBy        *string `json:"updated_by"`
}

type UserExamProgress struct {
	UserEpID       *string `json:"user_ep_id"`
	UserID         string  `json:"user_id"`
	UserEaID       string  `json:"user_ea_id"`
	UserLspID      string  `json:"user_lsp_id"`
	UserCpID       string  `json:"user_cp_id"`
	SrNo           int     `json:"sr_no"`
	QuestionID     string  `json:"question_id"`
	QuestionType   string  `json:"question_type"`
	Answer         string  `json:"answer"`
	QAttemptStatus string  `json:"q_attempt_status"`
	TotalTimeSpent string  `json:"total_time_spent"`
	CorrectAnswer  string  `json:"correct_answer"`
	SectionID      string  `json:"section_id"`
	CreatedBy      *string `json:"created_by"`
	UpdatedBy      *string `json:"updated_by"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

type UserExamProgressInput struct {
	UserEpID       *string `json:"user_ep_id"`
	UserID         string  `json:"user_id"`
	UserEaID       string  `json:"user_ea_id"`
	UserLspID      string  `json:"user_lsp_id"`
	UserCpID       string  `json:"user_cp_id"`
	SrNo           int     `json:"sr_no"`
	QuestionID     string  `json:"question_id"`
	QuestionType   string  `json:"question_type"`
	Answer         string  `json:"answer"`
	QAttemptStatus string  `json:"q_attempt_status"`
	TotalTimeSpent string  `json:"total_time_spent"`
	CorrectAnswer  string  `json:"correct_answer"`
	SectionID      string  `json:"section_id"`
	CreatedBy      *string `json:"created_by"`
	UpdatedBy      *string `json:"updated_by"`
}

type UserExamResult struct {
	UserErID       *string `json:"user_er_id"`
	UserID         string  `json:"user_id"`
	UserEaID       string  `json:"user_ea_id"`
	UserScore      int     `json:"user_score"`
	CorrectAnswers int     `json:"correct_answers"`
	WrongAnswers   int     `json:"wrong_answers"`
	ResultStatus   string  `json:"result_status"`
	CreatedBy      *string `json:"created_by"`
	UpdatedBy      *string `json:"updated_by"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

type UserExamResultDetails struct {
	UserID   string `json:"user_id"`
	UserEaID string `json:"user_ea_id"`
}

type UserExamResultInfo struct {
	UserID   string            `json:"user_id"`
	UserEaID string            `json:"user_ea_id"`
	Results  []*UserExamResult `json:"results"`
}

type UserExamResultInput struct {
	UserErID       *string `json:"user_er_id"`
	UserID         string  `json:"user_id"`
	UserEaID       string  `json:"user_ea_id"`
	UserScore      int     `json:"user_score"`
	CorrectAnswers int     `json:"correct_answers"`
	WrongAnswers   int     `json:"wrong_answers"`
	ResultStatus   string  `json:"result_status"`
	CreatedBy      *string `json:"created_by"`
	UpdatedBy      *string `json:"updated_by"`
}

type UserFilters struct {
	Email      *string `json:"email"`
	NameSearch *string `json:"nameSearch"`
	Role       *string `json:"role"`
	Status     *string `json:"status"`
}

type UserInput struct {
	ID         *string         `json:"id"`
	FirstName  string          `json:"first_name"`
	LastName   string          `json:"last_name"`
	Status     string          `json:"status"`
	Role       string          `json:"role"`
	IsVerified bool            `json:"is_verified"`
	IsActive   bool            `json:"is_active"`
	Gender     string          `json:"gender"`
	CreatedBy  *string         `json:"created_by"`
	UpdatedBy  *string         `json:"updated_by"`
	Email      string          `json:"email"`
	Phone      string          `json:"phone"`
	Photo      *graphql.Upload `json:"Photo"`
	PhotoURL   *string         `json:"photo_url"`
}

type UserLanguageMap struct {
	UserLanguageID *string `json:"user_language_id"`
	UserID         string  `json:"user_id"`
	UserLspID      string  `json:"user_lsp_id"`
	Language       string  `json:"language"`
	IsBaseLanguage bool    `json:"is_base_language"`
	IsActive       bool    `json:"is_active"`
	CreatedBy      *string `json:"created_by"`
	UpdatedBy      *string `json:"updated_by"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

type UserLanguageMapInput struct {
	UserID         string  `json:"user_id"`
	UserLspID      string  `json:"user_lsp_id"`
	Language       string  `json:"language"`
	IsBaseLanguage bool    `json:"is_base_language"`
	IsActive       bool    `json:"is_active"`
	CreatedBy      *string `json:"created_by"`
	UpdatedBy      *string `json:"updated_by"`
}

type UserLspMap struct {
	UserLspID *string `json:"user_lsp_id"`
	UserID    string  `json:"user_id"`
	LspID     string  `json:"lsp_id"`
	Status    string  `json:"status"`
	CreatedBy *string `json:"created_by"`
	UpdatedBy *string `json:"updated_by"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type UserLspMapInput struct {
	UserLspID *string `json:"user_lsp_id"`
	UserID    string  `json:"user_id"`
	LspID     string  `json:"lsp_id"`
	Status    string  `json:"status"`
	CreatedBy *string `json:"created_by"`
	UpdatedBy *string `json:"updated_by"`
}

type UserNotes struct {
	UserNotesID *string `json:"user_notes_id"`
	UserID      string  `json:"user_id"`
	UserLspID   string  `json:"user_lsp_id"`
	CourseID    string  `json:"course_id"`
	ModuleID    string  `json:"module_id"`
	TopicID     string  `json:"topic_id"`
	Sequence    int     `json:"sequence"`
	Status      string  `json:"status"`
	Details     string  `json:"details"`
	IsActive    bool    `json:"is_active"`
	CreatedBy   *string `json:"created_by"`
	UpdatedBy   *string `json:"updated_by"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type UserNotesInput struct {
	UserNotesID *string `json:"user_notes_id"`
	UserID      string  `json:"user_id"`
	UserLspID   string  `json:"user_lsp_id"`
	CourseID    string  `json:"course_id"`
	ModuleID    string  `json:"module_id"`
	TopicID     string  `json:"topic_id"`
	Sequence    int     `json:"sequence"`
	Status      string  `json:"status"`
	Details     string  `json:"details"`
	IsActive    bool    `json:"is_active"`
	CreatedBy   *string `json:"created_by"`
	UpdatedBy   *string `json:"updated_by"`
}

type UserOrganizationMap struct {
	UserOrganizationID *string `json:"user_organization_id"`
	UserID             string  `json:"user_id"`
	UserLspID          string  `json:"user_lsp_id"`
	OrganizationID     string  `json:"organization_id"`
	OrganizationRole   string  `json:"organization_role"`
	IsActive           bool    `json:"is_active"`
	EmployeeID         string  `json:"employee_id"`
	CreatedBy          *string `json:"created_by"`
	UpdatedBy          *string `json:"updated_by"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

type UserOrganizationMapInput struct {
	UserOrganizationID *string `json:"user_organization_id"`
	UserID             string  `json:"user_id"`
	UserLspID          string  `json:"user_lsp_id"`
	OrganizationID     string  `json:"organization_id"`
	OrganizationRole   string  `json:"organization_role"`
	IsActive           bool    `json:"is_active"`
	EmployeeID         string  `json:"employee_id"`
	CreatedBy          *string `json:"created_by"`
	UpdatedBy          *string `json:"updated_by"`
}

type UserPreference struct {
	UserPreferenceID *string `json:"user_preference_id"`
	UserID           string  `json:"user_id"`
	UserLspID        string  `json:"user_lsp_id"`
	SubCategory      string  `json:"sub_category"`
	IsBase           bool    `json:"is_base"`
	IsActive         bool    `json:"is_active"`
	CreatedBy        *string `json:"created_by"`
	UpdatedBy        *string `json:"updated_by"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

type UserPreferenceInput struct {
	UserPreferenceID *string `json:"user_preference_id"`
	UserID           string  `json:"user_id"`
	UserLspID        string  `json:"user_lsp_id"`
	SubCategory      string  `json:"sub_category"`
	IsBase           bool    `json:"is_base"`
	IsActive         bool    `json:"is_active"`
	CreatedBy        *string `json:"created_by"`
	UpdatedBy        *string `json:"updated_by"`
}

type UserQuizAttempt struct {
	UserQaID     *string `json:"user_qa_id"`
	UserID       string  `json:"user_id"`
	UserCpID     string  `json:"user_cp_id"`
	UserCourseID string  `json:"user_course_id"`
	QuizID       string  `json:"quiz_id"`
	QuizAttempt  int     `json:"quiz_attempt"`
	TopicID      string  `json:"topic_id"`
	Result       string  `json:"result"`
	IsActive     bool    `json:"is_active"`
	CreatedBy    *string `json:"created_by"`
	UpdatedBy    *string `json:"updated_by"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type UserQuizAttemptInput struct {
	UserQaID     *string `json:"user_qa_id"`
	UserID       string  `json:"user_id"`
	UserCpID     string  `json:"user_cp_id"`
	UserCourseID string  `json:"user_course_id"`
	QuizID       string  `json:"quiz_id"`
	QuizAttempt  int     `json:"quiz_attempt"`
	TopicID      string  `json:"topic_id"`
	Result       string  `json:"result"`
	IsActive     bool    `json:"is_active"`
	CreatedBy    *string `json:"created_by"`
	UpdatedBy    *string `json:"updated_by"`
}

type UserRole struct {
	UserRoleID *string `json:"user_role_id"`
	UserID     string  `json:"user_id"`
	UserLspID  string  `json:"user_lsp_id"`
	Role       string  `json:"role"`
	IsActive   bool    `json:"is_active"`
	CreatedBy  *string `json:"created_by"`
	UpdatedBy  *string `json:"updated_by"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

type UserRoleInput struct {
	UserRoleID *string `json:"user_role_id"`
	UserID     string  `json:"user_id"`
	UserLspID  string  `json:"user_lsp_id"`
	Role       string  `json:"role"`
	IsActive   bool    `json:"is_active"`
	CreatedBy  *string `json:"created_by"`
	UpdatedBy  *string `json:"updated_by"`
}
