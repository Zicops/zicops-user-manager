// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"github.com/99designs/gqlgen/graphql"
)

type Crt struct {
	CrtID             *string   `json:"crt_id"`
	VendorID          string    `json:"vendor_id"`
	Description       *string   `json:"description"`
	IsApplicable      *bool     `json:"is_applicable"`
	Expertise         []*string `json:"expertise"`
	Languages         []*string `json:"languages"`
	OutputDeliveries  []*string `json:"output_deliveries"`
	SampleFiles       []*string `json:"sample_files"`
	Profiles          []*string `json:"profiles"`
	IsExpertiseOnline *bool     `json:"is_expertise_online"`
	CreatedAt         *string   `json:"created_at"`
	CreatedBy         *string   `json:"created_by"`
	UpdatedAt         *string   `json:"updated_at"`
	UpdatedBy         *string   `json:"updated_by"`
	Status            *string   `json:"status"`
}

type CRTInput struct {
	CrtID             *string   `json:"crt_id"`
	VendorID          string    `json:"vendor_id"`
	Description       *string   `json:"description"`
	IsApplicable      *bool     `json:"is_applicable"`
	Expertise         []*string `json:"expertise"`
	Languages         []*string `json:"languages"`
	OutputDeliveries  []*string `json:"output_deliveries"`
	SampleFiles       []*string `json:"sample_files"`
	Profiles          []*string `json:"profiles"`
	IsExpertiseOnline *bool     `json:"is_expertise_online"`
	Status            *string   `json:"status"`
}

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

type ContentDevelopment struct {
	CdID             *string   `json:"cd_id"`
	VendorID         *string   `json:"vendor_id"`
	Description      *string   `json:"description"`
	IsApplicable     *bool     `json:"is_applicable"`
	Expertise        []*string `json:"expertise"`
	Languages        []*string `json:"languages"`
	OutputDeliveries []*string `json:"output_deliveries"`
	SampleFiles      []*string `json:"sample_files"`
	CreatedAt        *string   `json:"created_at"`
	CreatedBy        *string   `json:"created_by"`
	UpdatedAt        *string   `json:"updated_at"`
	UpdatedBy        *string   `json:"updated_by"`
	Status           *string   `json:"status"`
}

type ContentDevelopmentInput struct {
	CdID             *string   `json:"cd_id"`
	VendorID         string    `json:"vendor_id"`
	Description      *string   `json:"description"`
	IsApplicable     *bool     `json:"is_applicable"`
	Expertise        []*string `json:"expertise"`
	Languages        []*string `json:"languages"`
	OutputDeliveries []*string `json:"output_deliveries"`
	SampleFiles      []*string `json:"sample_files"`
	Status           *string   `json:"status"`
}

type Count struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type CourseConsumptionStats struct {
	ID                     *string `json:"ID"`
	LspID                  *string `json:"LspId"`
	CourseID               *string `json:"CourseId"`
	Category               *string `json:"Category"`
	SubCategory            *string `json:"SubCategory"`
	Owner                  *string `json:"Owner"`
	Duration               *int    `json:"Duration"`
	TotalLearners          *int    `json:"TotalLearners"`
	ActiveLearners         *int    `json:"ActiveLearners"`
	CompletedLearners      *int    `json:"CompletedLearners"`
	ExpectedCompletionTime *int    `json:"ExpectedCompletionTime"`
	AverageCompletionTime  *int    `json:"AverageCompletionTime"`
	AverageComplianceScore *int    `json:"AverageComplianceScore"`
	CreatedAt              *int    `json:"CreatedAt"`
	UpdatedAt              *int    `json:"UpdatedAt"`
	CreatedBy              *string `json:"CreatedBy"`
	UpdatedBy              *string `json:"UpdatedBy"`
}

type CourseMapFilters struct {
	LspID       []*string `json:"lsp_id"`
	IsMandatory *bool     `json:"is_mandatory"`
	Status      *string   `json:"status"`
	Type        *string   `json:"type"`
}

type CourseViews struct {
	Seconds    *int      `json:"seconds"`
	CreatedAt  *string   `json:"created_at"`
	LspID      *string   `json:"lsp_id"`
	UserIds    []*string `json:"user_ids"`
	DateString *string   `json:"date_string"`
}

type ExperienceInput struct {
	ExpID           *string `json:"exp_id"`
	VendorID        *string `json:"vendor_id"`
	Email           string  `json:"email"`
	Title           *string `json:"title"`
	CompanyName     *string `json:"company_name"`
	EmployementType *string `json:"employement_type"`
	Location        *string `json:"location"`
	LocationType    *string `json:"location_type"`
	StartDate       *int    `json:"start_date"`
	EndDate         *int    `json:"end_date"`
	Status          *string `json:"status"`
}

type ExperienceVendor struct {
	ExpID           string  `json:"ExpId"`
	VendorID        string  `json:"VendorId"`
	PfID            string  `json:"PfId"`
	StartDate       *int    `json:"StartDate"`
	EndDate         *int    `json:"EndDate"`
	Title           *string `json:"Title"`
	Location        *string `json:"Location"`
	LocationType    *string `json:"LocationType"`
	EmployementType *string `json:"EmployementType"`
	CompanyName     *string `json:"CompanyName"`
	CreatedAt       *string `json:"CreatedAt"`
	CreatedBy       *string `json:"CreatedBy"`
	UpdatedAt       *string `json:"UpdatedAt"`
	UpdatedBy       *string `json:"UpdatedBy"`
	Status          *string `json:"Status"`
}

type InviteResponse struct {
	Email   *string `json:"email"`
	Message string  `json:"message"`
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
	Type       string          `json:"type"`
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

type PaginatedCCStats struct {
	Stats      []*CourseConsumptionStats `json:"stats"`
	PageCursor *string                   `json:"pageCursor"`
	Direction  *string                   `json:"direction"`
	PageSize   *int                      `json:"pageSize"`
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

type PaginatedVendors struct {
	Vendors    []*Vendor `json:"vendors"`
	PageCursor *string   `json:"pageCursor"`
	Direction  *string   `json:"direction"`
	PageSize   *int      `json:"pageSize"`
}

type Sme struct {
	VendorID         *string   `json:"vendor_id"`
	SmeID            *string   `json:"sme_id"`
	Description      *string   `json:"description"`
	IsApplicable     *bool     `json:"is_applicable"`
	Expertise        []*string `json:"expertise"`
	Languages        []*string `json:"languages"`
	OutputDeliveries []*string `json:"output_deliveries"`
	SampleFiles      []*string `json:"sample_files"`
	Profiles         []*string `json:"profiles"`
	CreatedAt        *string   `json:"created_at"`
	CreatedBy        *string   `json:"created_by"`
	UpdatedAt        *string   `json:"updated_at"`
	UpdatedBy        *string   `json:"updated_by"`
	Status           *string   `json:"status"`
}

type SMEInput struct {
	VendorID         string    `json:"vendor_id"`
	SmeID            *string   `json:"sme_id"`
	Description      *string   `json:"description"`
	IsApplicable     *bool     `json:"is_applicable"`
	Expertise        []*string `json:"expertise"`
	Languages        []*string `json:"languages"`
	OutputDeliveries []*string `json:"output_deliveries"`
	SampleFiles      []*string `json:"sample_files"`
	Profiles         []*string `json:"profiles"`
	Status           *string   `json:"Status"`
}

type SampleFile struct {
	SfID      string  `json:"sf_id"`
	Name      *string `json:"name"`
	FileType  *string `json:"fileType"`
	Price     *string `json:"price"`
	FileURL   *string `json:"file_url"`
	CreatedAt *string `json:"created_at"`
	CreatedBy *string `json:"created_by"`
	UpdatedAt *string `json:"updated_at"`
	UpdatedBy *string `json:"updated_by"`
	Status    *string `json:"status"`
}

type SampleFileInput struct {
	File        graphql.Upload `json:"file"`
	Name        string         `json:"name"`
	Description *string        `json:"description"`
	Pricing     string         `json:"pricing"`
	FileType    *string        `json:"fileType"`
	Status      *string        `json:"status"`
	VendorID    string         `json:"vendorId"`
	PType       string         `json:"p_type"`
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
	LspID        *string `json:"lsp_id"`
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
	LspID        *string `json:"lsp_id"`
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

type UserCourseMapStats struct {
	LspID       *string  `json:"lsp_id"`
	UserID      *string  `json:"user_id"`
	TypeStats   []*Count `json:"type_stats"`
	StatusStats []*Count `json:"status_stats"`
}

type UserCourseMapStatsInput struct {
	LspID        *string   `json:"lsp_id"`
	UserID       *string   `json:"user_id"`
	CourseType   []*string `json:"course_type"`
	CourseStatus []*string `json:"course_status"`
	IsMandatory  *bool     `json:"is_mandatory"`
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

type Vendor struct {
	VendorID     string    `json:"vendorId"`
	Type         string    `json:"type"`
	Level        string    `json:"level"`
	Name         string    `json:"name"`
	Description  *string   `json:"description"`
	PhotoURL     *string   `json:"photo_url"`
	Address      *string   `json:"address"`
	Users        []*string `json:"users"`
	Website      *string   `json:"website"`
	FacebookURL  *string   `json:"facebook_url"`
	InstagramURL *string   `json:"instagram_url"`
	TwitterURL   *string   `json:"twitter_url"`
	LinkedinURL  *string   `json:"linkedin_url"`
	CreatedAt    *string   `json:"created_at"`
	CreatedBy    *string   `json:"created_by"`
	UpdatedAt    *string   `json:"updated_at"`
	UpdatedBy    *string   `json:"updated_by"`
	Status       *string   `json:"status"`
}

type VendorInput struct {
	LspID        *string         `json:"lsp_id"`
	Name         *string         `json:"name"`
	Level        *string         `json:"level"`
	VendorID     *string         `json:"vendor_id"`
	Type         *string         `json:"type"`
	Photo        *graphql.Upload `json:"photo"`
	Address      *string         `json:"address"`
	Website      *string         `json:"website"`
	FacebookURL  *string         `json:"facebook_url"`
	InstagramURL *string         `json:"instagram_url"`
	TwitterURL   *string         `json:"twitter_url"`
	LinkedinURL  *string         `json:"linkedin_url"`
	Users        []*string       `json:"users"`
	Description  *string         `json:"description"`
	Status       *string         `json:"status"`
}

type VendorProfile struct {
	PfID               *string   `json:"pf_id"`
	VendorID           *string   `json:"vendor_id"`
	FirstName          *string   `json:"first_name"`
	LastName           *string   `json:"last_name"`
	Email              *string   `json:"email"`
	Phone              *string   `json:"phone"`
	PhotoURL           *string   `json:"photo_url"`
	Description        *string   `json:"description"`
	Language           []*string `json:"language"`
	SmeExpertise       []*string `json:"sme_expertise"`
	ClassroomExpertise []*string `json:"classroom_expertise"`
	Experience         []*string `json:"experience"`
	ExperienceYears    *string   `json:"experience_years"`
	IsSpeaker          *bool     `json:"is_speaker"`
	CreatedAt          *string   `json:"created_at"`
	CreatedBy          *string   `json:"created_by"`
	UpdatedAt          *string   `json:"updated_at"`
	UpdatedBy          *string   `json:"updated_by"`
	Status             *string   `json:"status"`
}

type VendorProfileInput struct {
	VendorID           string          `json:"vendor_id"`
	FirstName          *string         `json:"first_name"`
	LastName           *string         `json:"last_name"`
	Email              string          `json:"email"`
	Phone              *string         `json:"phone"`
	Photo              *graphql.Upload `json:"photo"`
	Description        *string         `json:"description"`
	Languages          []*string       `json:"languages"`
	SmeExpertise       []*string       `json:"SME_expertise"`
	ClassroomExpertise []*string       `json:"Classroom_expertise"`
	Experience         []*string       `json:"experience"`
	ExperienceYears    *string         `json:"experience_years"`
	IsSpeaker          *bool           `json:"is_speaker"`
	Status             *string         `json:"status"`
}
