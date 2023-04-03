package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/zicops-user-manager/graph/generated"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/handlers"
	"github.com/zicops/zicops-user-manager/handlers/orgs"
	"github.com/zicops/zicops-user-manager/handlers/queries"
)

// RegisterUsers is the resolver for the registerUsers field.
func (r *mutationResolver) RegisterUsers(ctx context.Context, input []*model.UserInput) ([]*model.User, error) {
	result, _, err := handlers.RegisterUsers(ctx, input, true, false)
	if err != nil {
		log.Errorf("Error registering users: %v", err)
		return nil, err
	}
	return result, nil
}

// InviteUsers is the resolver for the inviteUsers field.
func (r *mutationResolver) InviteUsers(ctx context.Context, emails []string, lspID *string) (*bool, error) {
	lspIDStr := ""
	if lspID != nil {
		lspIDStr = *lspID
	}

	result, err := handlers.InviteUsers(ctx, emails, lspIDStr)
	if err != nil {
		log.Errorf("Error inviting users: %v", err)
		return nil, err
	}
	return result, nil
}

// InviteUsersWithRole is the resolver for the inviteUsersWithRole field.
func (r *mutationResolver) InviteUsersWithRole(ctx context.Context, emails []string, lspID *string, role *string) ([]*model.InviteResponse, error) {
	res, err := handlers.InviteUserWithRole(ctx, emails, *lspID, role)
	if err != nil {
		log.Println("Error while Inviting users with roles: %v", err)
		return nil, err
	}
	return res, nil
}

// UpdateUser is the resolver for the updateUser field.
func (r *mutationResolver) UpdateUser(ctx context.Context, input model.UserInput) (*model.User, error) {
	result, err := handlers.UpdateUser(ctx, input)
	if err != nil {
		log.Errorf("Error updating user: %v", err)
		return nil, err
	}
	return result, nil
}

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context) (*model.User, error) {
	result, err := handlers.LoginUser(ctx)
	if err != nil {
		log.Errorf("Error logging in user: %v", err)
		return nil, err
	}
	return result, nil
}

// AddUserLspMap is the resolver for the addUserLspMap field.
func (r *mutationResolver) AddUserLspMap(ctx context.Context, input []*model.UserLspMapInput) ([]*model.UserLspMap, error) {
	isAdmin := false
	result, err := handlers.AddUserLspMap(ctx, input, &isAdmin)
	if err != nil {
		log.Errorf("Error adding lsp map for user: %v", err)
		return nil, err
	}
	return result, nil
}

// UpdateUserLspMap is the resolver for the updateUserLspMap field.
func (r *mutationResolver) UpdateUserLspMap(ctx context.Context, input model.UserLspMapInput) (*model.UserLspMap, error) {
	result, err := handlers.UpdateUserLspMap(ctx, input)
	if err != nil {
		log.Errorf("Error updating lsp map for user: %v", err)
		return nil, err
	}
	return result, nil
}

// AddUserOrganizationMap is the resolver for the addUserOrganizationMap field.
func (r *mutationResolver) AddUserOrganizationMap(ctx context.Context, input []*model.UserOrganizationMapInput) ([]*model.UserOrganizationMap, error) {
	result, err := handlers.AddUserOrganizationMap(ctx, input)
	if err != nil {
		log.Errorf("Error adding org map for user: %v", err)
		return nil, err
	}
	return result, nil
}

// UpdateUserOrganizationMap is the resolver for the updateUserOrganizationMap field.
func (r *mutationResolver) UpdateUserOrganizationMap(ctx context.Context, input model.UserOrganizationMapInput) (*model.UserOrganizationMap, error) {
	result, err := handlers.UpdateUserOrganizationMap(ctx, input)
	if err != nil {
		log.Errorf("Error updating org map for user: %v", err)
		return nil, err
	}
	return result, nil
}

// AddUserLanguageMap is the resolver for the addUserLanguageMap field.
func (r *mutationResolver) AddUserLanguageMap(ctx context.Context, input []*model.UserLanguageMapInput) ([]*model.UserLanguageMap, error) {
	result, err := handlers.AddUserLanguageMap(ctx, input)
	if err != nil {
		log.Errorf("Error adding lang map for user: %v", err)
		return nil, err
	}
	return result, nil
}

// AddUserPreference is the resolver for the addUserPreference field.
func (r *mutationResolver) AddUserPreference(ctx context.Context, input []*model.UserPreferenceInput) ([]*model.UserPreference, error) {
	result, err := handlers.AddUserPreference(ctx, input)
	if err != nil {
		log.Errorf("Error adding preference map for user: %v", err)
		return nil, err
	}
	return result, nil
}

// UpdateUserPreference is the resolver for the updateUserPreference field.
func (r *mutationResolver) UpdateUserPreference(ctx context.Context, input model.UserPreferenceInput) (*model.UserPreference, error) {
	result, err := handlers.UpdateUserPreference(ctx, input)
	if err != nil {
		log.Errorf("Error updating preference map for user: %v", err)
		return nil, err
	}
	return result, nil
}

// AddUserRoles is the resolver for the addUserRoles field.
func (r *mutationResolver) AddUserRoles(ctx context.Context, input []*model.UserRoleInput) ([]*model.UserRole, error) {
	result, err := handlers.AddUserRoles(ctx, input)
	if err != nil {
		log.Errorf("Error adding roles map for user: %v", err)
		return nil, err
	}
	return result, nil
}

// UpdateUserRole is the resolver for the updateUserRole field.
func (r *mutationResolver) UpdateUserRole(ctx context.Context, input model.UserRoleInput) (*model.UserRole, error) {
	result, err := handlers.UpdateUserRole(ctx, input)
	if err != nil {
		log.Errorf("Error updating roles map for user: %v", err)
		return nil, err
	}
	return result, nil
}

// AddUserCohort is the resolver for the addUserCohort field.
func (r *mutationResolver) AddUserCohort(ctx context.Context, input []*model.UserCohortInput) ([]*model.UserCohort, error) {
	result, err := handlers.AddUserCohort(ctx, input)
	if err != nil {
		log.Errorf("Error adding cohort map for user: %v", err)
		return nil, err
	}
	return result, nil
}

// UpdateUserCohort is the resolver for the updateUserCohort field.
func (r *mutationResolver) UpdateUserCohort(ctx context.Context, input model.UserCohortInput) (*model.UserCohort, error) {
	result, err := handlers.UpdateUserCohort(ctx, input)
	if err != nil {
		log.Errorf("Error updating cohort map for user: %v", err)
		return nil, err
	}
	return result, nil
}

// AddUserCourse is the resolver for the addUserCourse field.
func (r *mutationResolver) AddUserCourse(ctx context.Context, input []*model.UserCourseInput) ([]*model.UserCourse, error) {
	result, err := handlers.AddUserCourse(ctx, input)
	if err != nil {
		log.Errorf("Error adding course map for user: %v", err)
		return nil, err
	}
	return result, nil
}

// AddUserCohortCourses is the resolver for the addUserCohortCourses field.
func (r *mutationResolver) AddUserCohortCourses(ctx context.Context, userIds []string, cohortID string) (*bool, error) {
	result, err := handlers.AddUserCohortCourses(ctx, userIds, cohortID)
	if err != nil {
		log.Errorf("Error adding course map for all the users: %v", err)
		return nil, err
	}
	return result, nil
}

// UpdateUserCourse is the resolver for the updateUserCourse field.
func (r *mutationResolver) UpdateUserCourse(ctx context.Context, input model.UserCourseInput) (*model.UserCourse, error) {
	result, err := handlers.UpdateUserCourse(ctx, input)
	if err != nil {
		log.Errorf("Error updating course map for user: %v", err)
		return nil, err
	}
	return result, nil
}

// AddUserCourseProgress is the resolver for the addUserCourseProgress field.
func (r *mutationResolver) AddUserCourseProgress(ctx context.Context, input []*model.UserCourseProgressInput) ([]*model.UserCourseProgress, error) {
	result, err := handlers.AddUserCourseProgress(ctx, input)
	if err != nil {
		log.Errorf("Error adding course progress for user: %v", err)
		return nil, err
	}
	return result, nil
}

// UpdateUserCourseProgress is the resolver for the updateUserCourseProgress field.
func (r *mutationResolver) UpdateUserCourseProgress(ctx context.Context, input model.UserCourseProgressInput) (*model.UserCourseProgress, error) {
	result, err := handlers.UpdateUserCourseProgress(ctx, input)
	if err != nil {
		log.Errorf("Error updating course progress for user: %v", err)
		return nil, err
	}
	return result, nil
}

// AddUserQuizAttempt is the resolver for the addUserQuizAttempt field.
func (r *mutationResolver) AddUserQuizAttempt(ctx context.Context, input []*model.UserQuizAttemptInput) ([]*model.UserQuizAttempt, error) {
	result, err := handlers.AddUserQuizAttempt(ctx, input)
	if err != nil {
		log.Errorf("Error adding quiz attempt for user: %v", err)
		return nil, err
	}
	return result, nil
}

// UpdateUserQuizAttempt is the resolver for the updateUserQuizAttempt field.
func (r *mutationResolver) UpdateUserQuizAttempt(ctx context.Context, input model.UserQuizAttemptInput) (*model.UserQuizAttempt, error) {
	result, err := handlers.UpdateUserQuizAttempt(ctx, input)
	if err != nil {
		log.Errorf("Error updating quiz attempt for user: %v", err)
		return nil, err
	}
	return result, nil
}

// AddUserBookmark is the resolver for the addUserBookmark field.
func (r *mutationResolver) AddUserBookmark(ctx context.Context, input []*model.UserBookmarkInput) ([]*model.UserBookmark, error) {
	result, err := handlers.AddUserBookmark(ctx, input)
	if err != nil {
		log.Errorf("Error adding bookmark for user: %v", err)
		return nil, err
	}
	return result, nil
}

// UpdateUserBookmark is the resolver for the updateUserBookmark field.
func (r *mutationResolver) UpdateUserBookmark(ctx context.Context, input model.UserBookmarkInput) (*model.UserBookmark, error) {
	result, err := handlers.UpdateUserBookmark(ctx, input)
	if err != nil {
		log.Errorf("Error updating bookmark for user: %v", err)
		return nil, err
	}
	return result, nil
}

// AddUserNotes is the resolver for the addUserNotes field.
func (r *mutationResolver) AddUserNotes(ctx context.Context, input []*model.UserNotesInput) ([]*model.UserNotes, error) {
	result, err := handlers.AddUserNotes(ctx, input)
	if err != nil {
		log.Errorf("Error adding notes for user: %v", err)
		return nil, err
	}
	return result, nil
}

// UpdateUserNotes is the resolver for the updateUserNotes field.
func (r *mutationResolver) UpdateUserNotes(ctx context.Context, input model.UserNotesInput) (*model.UserNotes, error) {
	result, err := handlers.UpdateUserNotes(ctx, input)
	if err != nil {
		log.Errorf("Error updating notes for user: %v", err)
		return nil, err
	}
	return result, nil
}

// AddUserExamAttempts is the resolver for the addUserExamAttempts field.
func (r *mutationResolver) AddUserExamAttempts(ctx context.Context, input []*model.UserExamAttemptsInput) ([]*model.UserExamAttempts, error) {
	result, err := handlers.AddUserExamAttempts(ctx, input)
	if err != nil {
		log.Errorf("Error adding exams for user: %v", err)
		return nil, err
	}
	return result, nil
}

// UpdateUserExamAttempts is the resolver for the updateUserExamAttempts field.
func (r *mutationResolver) UpdateUserExamAttempts(ctx context.Context, input model.UserExamAttemptsInput) (*model.UserExamAttempts, error) {
	result, err := handlers.UpdateUserExamAttempts(ctx, input)
	if err != nil {
		log.Errorf("Error updating exams for user: %v", err)
		return nil, err
	}
	return result, nil
}

// AddUserExamProgress is the resolver for the addUserExamProgress field.
func (r *mutationResolver) AddUserExamProgress(ctx context.Context, input []*model.UserExamProgressInput) ([]*model.UserExamProgress, error) {
	result, err := handlers.AddUserExamProgress(ctx, input)
	if err != nil {
		log.Errorf("Error adding exam progress for user: %v", err)
		return nil, err
	}
	return result, nil
}

// UpdateUserExamProgress is the resolver for the updateUserExamProgress field.
func (r *mutationResolver) UpdateUserExamProgress(ctx context.Context, input model.UserExamProgressInput) (*model.UserExamProgress, error) {
	result, err := handlers.UpdateUserExamProgress(ctx, input)
	if err != nil {
		log.Errorf("Error updating exam progress for user: %v", err)
		return nil, err
	}
	return result, nil
}

// AddUserExamResult is the resolver for the addUserExamResult field.
func (r *mutationResolver) AddUserExamResult(ctx context.Context, input []*model.UserExamResultInput) ([]*model.UserExamResult, error) {
	result, err := handlers.AddUserExamResult(ctx, input)
	if err != nil {
		log.Errorf("Error adding exam result for user: %v", err)
		return nil, err
	}
	return result, nil
}

// UpdateUserExamResult is the resolver for the updateUserExamResult field.
func (r *mutationResolver) UpdateUserExamResult(ctx context.Context, input model.UserExamResultInput) (*model.UserExamResult, error) {
	result, err := handlers.UpdateUserExamResult(ctx, input)
	if err != nil {
		log.Errorf("Error updating exam result for user: %v", err)
		return nil, err
	}
	return result, nil
}

// AddCohortMain is the resolver for the addCohortMain field.
func (r *mutationResolver) AddCohortMain(ctx context.Context, input model.CohortMainInput) (*model.CohortMain, error) {
	result, err := queries.AddCohortMain(ctx, input)
	if err != nil {
		log.Errorf("Error adding cohort: %v", err)
		return nil, err
	}
	return result, nil
}

// UpdateCohortMain is the resolver for the updateCohortMain field.
func (r *mutationResolver) UpdateCohortMain(ctx context.Context, input model.CohortMainInput) (*model.CohortMain, error) {
	result, err := queries.UpdateCohortMain(ctx, input)
	if err != nil {
		log.Errorf("Error updating cohort: %v", err)
		return nil, err
	}
	return result, nil
}

// AddOrganization is the resolver for the addOrganization field.
func (r *mutationResolver) AddOrganization(ctx context.Context, input model.OrganizationInput) (*model.Organization, error) {
	result, err := orgs.AddOrganization(ctx, input)
	if err != nil {
		log.Errorf("Error adding organization: %v", err)
		return nil, err
	}
	return result, nil
}

// UpdateOrganization is the resolver for the updateOrganization field.
func (r *mutationResolver) UpdateOrganization(ctx context.Context, input model.OrganizationInput) (*model.Organization, error) {
	result, err := orgs.UpdateOrganization(ctx, input)
	if err != nil {
		log.Errorf("Error updating organization: %v", err)
		return nil, err
	}
	return result, nil
}

// AddOrganizationUnit is the resolver for the addOrganizationUnit field.
func (r *mutationResolver) AddOrganizationUnit(ctx context.Context, input model.OrganizationUnitInput) (*model.OrganizationUnit, error) {
	result, err := orgs.AddOrganizationUnit(ctx, input)
	if err != nil {
		log.Errorf("Error adding organization unit: %v", err)
		return nil, err
	}
	return result, nil
}

// UpdateOrganizationUnit is the resolver for the updateOrganizationUnit field.
func (r *mutationResolver) UpdateOrganizationUnit(ctx context.Context, input model.OrganizationUnitInput) (*model.OrganizationUnit, error) {
	result, err := orgs.UpdateOrganizationUnit(ctx, input)
	if err != nil {
		log.Errorf("Error updating organization unit: %v", err)
		return nil, err
	}
	return result, nil
}

// AddLearningSpace is the resolver for the addLearningSpace field.
func (r *mutationResolver) AddLearningSpace(ctx context.Context, input model.LearningSpaceInput) (*model.LearningSpace, error) {
	result, err := orgs.AddLearningSpace(ctx, input)
	if err != nil {
		log.Errorf("Error adding learning space: %v", err)
		return nil, err
	}
	return result, nil
}

// UpdateLearningSpace is the resolver for the updateLearningSpace field.
func (r *mutationResolver) UpdateLearningSpace(ctx context.Context, input model.LearningSpaceInput) (*model.LearningSpace, error) {
	result, err := orgs.UpdateLearningSpace(ctx, input)
	if err != nil {
		log.Errorf("Error updating learning space: %v", err)
		return nil, err
	}
	return result, nil
}

// DeleteCohortImage is the resolver for the deleteCohortImage field.
func (r *mutationResolver) DeleteCohortImage(ctx context.Context, cohortID string, filename string) (*string, error) {
	result, err := queries.DeleteCohortImage(ctx, cohortID, filename)
	if err != nil {
		log.Errorf("Error while deleting the image of cohort: %v", err)
		return nil, err
	}
	return result, nil
}

// AddVendor is the resolver for the addVendor field.
func (r *mutationResolver) AddVendor(ctx context.Context, input *model.VendorInput) (*model.Vendor, error) {
	res, err := handlers.AddVendor(ctx, input)
	if err != nil {
		log.Errorf("Got error while creating vedor: %v", err)
		return nil, err
	}
	return res, nil
}

// UpdateVendor is the resolver for the updateVendor field.
func (r *mutationResolver) UpdateVendor(ctx context.Context, input *model.VendorInput) (*model.Vendor, error) {
	resp, err := handlers.UpdateVendor(ctx, input)
	if err != nil {
		log.Errorf("Got error while updating vendor: %v", err)
		return nil, err
	}
	return resp, nil
}

// CreateProfileVendor is the resolver for the createProfileVendor field.
func (r *mutationResolver) CreateProfileVendor(ctx context.Context, input *model.VendorProfileInput) (*model.VendorProfile, error) {
	resp, err := handlers.CreateProfileVendor(ctx, input)
	if err != nil {
		log.Println("Got error while creating profiles of vendor: %v", err)
		return nil, err
	}
	return resp, err
}

// CreateExperienceVendor is the resolver for the createExperienceVendor field.
func (r *mutationResolver) CreateExperienceVendor(ctx context.Context, input model.ExperienceInput) (*model.ExperienceVendor, error) {
	resp, err := handlers.CreateExperienceVendor(ctx, input)
	if err != nil {
		log.Println("Got error while creating experience of vendor: %v", err)
		return nil, err
	}
	return resp, nil
}

// UpdateExperienceVendor is the resolver for the updateExperienceVendor field.
func (r *mutationResolver) UpdateExperienceVendor(ctx context.Context, input model.ExperienceInput) (*model.ExperienceVendor, error) {
	res, err := handlers.UpdateExperienceVendor(ctx, input)
	if err != nil {
		log.Printf("Error updating experience of the vendor: %v", err)
		return nil, err
	}
	return res, nil
}

// UploadSampleFile is the resolver for the uploadSampleFile field.
func (r *mutationResolver) UploadSampleFile(ctx context.Context, input *model.SampleFileInput) (*model.SampleFile, error) {
	res, err := handlers.UploadSampleFile(ctx, input)
	if err != nil {
		log.Printf("Error uploading sample file: %v", err)
		return nil, err
	}
	return res, nil
}

// DeleteSampleFile is the resolver for the deleteSampleFile field.
func (r *mutationResolver) DeleteSampleFile(ctx context.Context, sfID string, vendorID string, pType string) (*bool, error) {
	res, err := handlers.DeleteSampleFile(ctx, sfID, vendorID, pType)
	if err != nil {
		log.Printf("Error deleting sample files: %v", err)
		return nil, err
	}
	return res, nil
}

// UpdateProfileVendor is the resolver for the updateProfileVendor field.
func (r *mutationResolver) UpdateProfileVendor(ctx context.Context, input *model.VendorProfileInput) (*model.VendorProfile, error) {
	res, err := handlers.UpdateProfileVendor(ctx, input)
	if err != nil {
		log.Printf("Error updating profile of the vendor: %v", err)
		return nil, err
	}
	return res, nil
}

// CreateSubjectMatterExpertise is the resolver for the createSubjectMatterExpertise field.
func (r *mutationResolver) CreateSubjectMatterExpertise(ctx context.Context, input *model.SMEInput) (*model.Sme, error) {
	resp, err := handlers.CreateSubjectMatterExpertise(ctx, input)
	if err != nil {
		log.Printf("Got error while creating subject matter expertise: %v", err)
		return nil, err
	}
	return resp, nil
}

// UpdateSubjectMatterExpertise is the resolver for the updateSubjectMatterExpertise field.
func (r *mutationResolver) UpdateSubjectMatterExpertise(ctx context.Context, input *model.SMEInput) (*model.Sme, error) {
	resp, err := handlers.UpdateSubjectMatterExpertise(ctx, input)
	if err != nil {
		log.Printf("Got error while updating subject matter expertise: %v", err)
		return nil, err
	}
	return resp, nil
}

// CreateClassRoomTraining is the resolver for the createClassRoomTraining field.
func (r *mutationResolver) CreateClassRoomTraining(ctx context.Context, input *model.CRTInput) (*model.Crt, error) {
	resp, err := handlers.CreateClassRoomTraining(ctx, input)
	if err != nil {
		log.Printf("Got error while creating classroom training: %v", err)
		return nil, err
	}
	return resp, nil
}

// UpdateClassRoomTraining is the resolver for the updateClassRoomTraining field.
func (r *mutationResolver) UpdateClassRoomTraining(ctx context.Context, input *model.CRTInput) (*model.Crt, error) {
	resp, err := handlers.UpdateClassRoomTraining(ctx, input)
	if err != nil {
		log.Printf("Got error while updating classroom training: %v", err)
		return nil, err
	}
	return resp, nil
}

// CreateContentDevelopment is the resolver for the createContentDevelopment field.
func (r *mutationResolver) CreateContentDevelopment(ctx context.Context, input *model.ContentDevelopmentInput) (*model.ContentDevelopment, error) {
	resp, err := handlers.CreateContentDevelopment(ctx, input)
	if err != nil {
		log.Printf("Got error while creating content development: %v", err)
		return nil, err
	}
	return resp, nil
}

// UpdateContentDevelopment is the resolver for the updateContentDevelopment field.
func (r *mutationResolver) UpdateContentDevelopment(ctx context.Context, input *model.ContentDevelopmentInput) (*model.ContentDevelopment, error) {
	resp, err := handlers.UpdateContentDevelopment(ctx, input)
	if err != nil {
		log.Printf("Got error while updating content development: %v", err)
		return nil, err
	}
	return resp, nil
}

// AddOrder is the resolver for the addOrder field.
func (r *mutationResolver) AddOrder(ctx context.Context, input *model.VendorOrderInput) (*model.VendorOrder, error) {
	resp, err := handlers.AddOrder(ctx, input)
	if err != nil {
		log.Printf("Got error while placing an order: %v", err)
		return nil, err
	}
	return resp, nil
}

// UpdateOrder is the resolver for the updateOrder field.
func (r *mutationResolver) UpdateOrder(ctx context.Context, input *model.VendorOrderInput) (*model.VendorOrder, error) {
	resp, err := handlers.UpdateOrder(ctx, input)
	if err != nil {
		log.Printf("Got error while editing an order: %v", err)
		return nil, err
	}
	return resp, nil
}

// AddOrderServies is the resolver for the addOrderServies field.
func (r *mutationResolver) AddOrderServies(ctx context.Context, input []*model.OrderServicesInput) ([]*model.OrderServices, error) {
	resp, err := handlers.AddOrderServies(ctx, input)
	if err != nil {
		log.Printf("Got error while adding services of an order: %v", err)
		return nil, err
	}
	return resp, nil
}

// UpdateOrderServices is the resolver for the updateOrderServices field.
func (r *mutationResolver) UpdateOrderServices(ctx context.Context, input *model.OrderServicesInput) (*model.OrderServices, error) {
	resp, err := handlers.UpdateOrderServices(ctx, input)
	if err != nil {
		log.Printf("Got error while adding services of an order: %v", err)
		return nil, err
	}
	return resp, nil
}

// CreateVendorUserMap is the resolver for the createVendorUserMap field.
func (r *mutationResolver) CreateVendorUserMap(ctx context.Context, vendorID *string, userID *string, status *string) (*model.VendorUserMap, error) {
	resp, err := handlers.CreateVendorUserMap(ctx, vendorID, userID, status)
	if err != nil {
		log.Printf("Got error while adding vendor user map: %v", err)
		return nil, err
	}
	return resp, nil
}

// UpdateVendorUserMap is the resolver for the updateVendorUserMap field.
func (r *mutationResolver) UpdateVendorUserMap(ctx context.Context, vendorID *string, userID *string, status *string) (*model.VendorUserMap, error) {
	resp, err := handlers.UpdateVendorUserMap(ctx, vendorID, userID, status)
	if err != nil {
		log.Printf("Got error while updating vendor user map: %v", err)
		return nil, err
	}
	return resp, nil
}

// DeleteVendorUserMap is the resolver for the deleteVendorUserMap field.
func (r *mutationResolver) DeleteVendorUserMap(ctx context.Context, vendorID *string, userID *string) (*bool, error) {
	resp, err := handlers.DeleteVendorUserMap(ctx, vendorID, userID)
	if err != nil {
		log.Printf("Got error while updating vendor user map: %v", err)
		return nil, err
	}
	return resp, nil
}

// DisableVendorLspMap is the resolver for the disableVendorLspMap field.
func (r *mutationResolver) DisableVendorLspMap(ctx context.Context, vendorID *string, lspID *string) (*bool, error) {
	resp, err := handlers.DisableVendorLspMap(ctx, vendorID, lspID)
	if err != nil {
		log.Printf("Got error while disabling vendor lsp map: %v", err)
		return nil, err
	}
	return resp, nil
}

// Logout is the resolver for the logout field.
func (r *queryResolver) Logout(ctx context.Context) (*bool, error) {
	result, err := handlers.Logout(ctx)
	if err != nil {
		log.Errorf("Error logging out user: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserLspMapsByLspID is the resolver for the getUserLspMapsByLspId field.
func (r *queryResolver) GetUserLspMapsByLspID(ctx context.Context, lspID string, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedUserLspMaps, error) {
	result, err := queries.GetUserLspMapsByLspID(ctx, lspID, pageCursor, direction, pageSize)
	if err != nil {
		log.Errorf("Error logging out user: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUsersForAdmin is the resolver for the getUsersForAdmin field.
func (r *queryResolver) GetUsersForAdmin(ctx context.Context, publishTime *int, pageCursor *string, direction *string, pageSize *int, filters *model.UserFilters) (*model.PaginatedUsers, error) {
	result, err := queries.GetUsersForAdmin(ctx, publishTime, pageCursor, direction, pageSize, filters)
	if err != nil {
		log.Errorf("Error getting users of an admin: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserDetails is the resolver for the getUserDetails field.
func (r *queryResolver) GetUserDetails(ctx context.Context, userIds []*string) ([]*model.User, error) {
	result, err := queries.GetUserDetails(ctx, userIds)
	if err != nil {
		log.Errorf("Error getting user of an admin: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserOrganizations is the resolver for the getUserOrganizations field.
func (r *queryResolver) GetUserOrganizations(ctx context.Context, userID string) ([]*model.UserOrganizationMap, error) {
	result, err := queries.GetUserOrganizations(ctx, userID)
	if err != nil {
		log.Errorf("Error getting orgs of a user: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserOrgDetails is the resolver for the getUserOrgDetails field.
func (r *queryResolver) GetUserOrgDetails(ctx context.Context, userID string, userLspID string) (*model.UserOrganizationMap, error) {
	result, err := queries.GetUserOrgDetails(ctx, userID, userLspID)
	if err != nil {
		log.Errorf("Error getting orgs of a user: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserPreferences is the resolver for the getUserPreferences field.
func (r *queryResolver) GetUserPreferences(ctx context.Context, userID string) ([]*model.UserPreference, error) {
	result, err := queries.GetUserPreferences(ctx, userID)
	if err != nil {
		log.Errorf("Error getting prefs of a user: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserPreferenceForLsp is the resolver for the getUserPreferenceForLsp field.
func (r *queryResolver) GetUserPreferenceForLsp(ctx context.Context, userID string, userLspID string) (*model.UserPreference, error) {
	result, err := queries.GetUserPreferenceForLsp(ctx, userID, userLspID)
	if err != nil {
		log.Errorf("Error getting prefs of a user: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserLsps is the resolver for the getUserLsps field.
func (r *queryResolver) GetUserLsps(ctx context.Context, userID string) ([]*model.UserLspMap, error) {
	result, err := queries.GetUserLsps(ctx, userID)
	if err != nil {
		log.Errorf("Error getting lsps of a user: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserLspByLspID is the resolver for the getUserLspByLspId field.
func (r *queryResolver) GetUserLspByLspID(ctx context.Context, userID string, lspID string) (*model.UserLspMap, error) {
	result, err := queries.GetUserLspByLspID(ctx, userID, lspID)
	if err != nil {
		log.Errorf("Error getting lsps of a user: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserCourseMaps is the resolver for the getUserCourseMaps field.
func (r *queryResolver) GetUserCourseMaps(ctx context.Context, userID string, publishTime *int, pageCursor *string, direction *string, pageSize *int, filters *model.CourseMapFilters) (*model.PaginatedCourseMaps, error) {
	result, err := queries.GetUserCourseMaps(ctx, userID, publishTime, pageCursor, direction, pageSize, filters)
	if err != nil {
		log.Errorf("Error getting courses of a user: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserCourseMapStats is the resolver for the getUserCourseMapStats field.
func (r *queryResolver) GetUserCourseMapStats(ctx context.Context, input model.UserCourseMapStatsInput) (*model.UserCourseMapStats, error) {
	result, err := queries.GetUserCourseMapStats(ctx, input)
	if err != nil {
		log.Errorf("Error getting course map statistics: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserCourseMapByCourseID is the resolver for the getUserCourseMapByCourseID field.
func (r *queryResolver) GetUserCourseMapByCourseID(ctx context.Context, userID string, courseID string, lspID *string) ([]*model.UserCourse, error) {
	result, err := queries.GetUserCourseMapByCourseID(ctx, userID, courseID)
	if err != nil {
		log.Errorf("Error getting course of a user: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserCourseProgressByMapID is the resolver for the getUserCourseProgressByMapId field.
func (r *queryResolver) GetUserCourseProgressByMapID(ctx context.Context, userID string, userCourseID []string) ([]*model.UserCourseProgress, error) {
	result, err := queries.GetUserCourseProgressByMapID(ctx, userID, userCourseID)
	if err != nil {
		log.Errorf("Error getting course progress of a user: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserCourseProgressByTopicID is the resolver for the getUserCourseProgressByTopicId field.
func (r *queryResolver) GetUserCourseProgressByTopicID(ctx context.Context, userID string, topicID string) ([]*model.UserCourseProgress, error) {
	result, err := queries.GetUserCourseProgressByTopicID(ctx, userID, topicID)
	if err != nil {
		log.Errorf("Error getting course progress of a user by topic id: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserNotes is the resolver for the getUserNotes field.
func (r *queryResolver) GetUserNotes(ctx context.Context, userID string, userLspID *string, courseID *string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedNotes, error) {
	result, err := queries.GetUserNotes(ctx, userID, userLspID, courseID, publishTime, pageCursor, direction, pageSize)
	if err != nil {
		log.Errorf("Error getting notes of a user: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserBookmarks is the resolver for the getUserBookmarks field.
func (r *queryResolver) GetUserBookmarks(ctx context.Context, userID string, userLspID *string, courseID *string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedBookmarks, error) {
	result, err := queries.GetUserBookmarks(ctx, userID, userLspID, courseID, publishTime, pageCursor, direction, pageSize)
	if err != nil {
		log.Errorf("Error getting bookmarks of a user: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserExamAttempts is the resolver for the getUserExamAttempts field.
func (r *queryResolver) GetUserExamAttempts(ctx context.Context, userID *string, examID string) ([]*model.UserExamAttempts, error) {
	result, err := queries.GetUserExamAttempts(ctx, userID, examID)
	if err != nil {
		log.Errorf("Error getting exam attempts of a user : %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserExamAttemptsByExamIds is the resolver for the getUserExamAttemptsByExamIds field.
func (r *queryResolver) GetUserExamAttemptsByExamIds(ctx context.Context, userID string, examIds []*string, filters *model.ExamAttemptsFilters) ([]*model.UserExamAttempts, error) {
	result, err := queries.GetUserExamAttemptsByExamIds(ctx, userID, examIds, filters)
	if err != nil {
		log.Errorf("Error getting exam attempts of a user : %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserExamResults is the resolver for the getUserExamResults field.
func (r *queryResolver) GetUserExamResults(ctx context.Context, userEaDetails []*model.UserExamResultDetails) ([]*model.UserExamResultInfo, error) {
	result, err := queries.GetUserExamResults(ctx, userEaDetails)
	if err != nil {
		log.Errorf("Error getting exam results of a user : %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserExamProgress is the resolver for the getUserExamProgress field.
func (r *queryResolver) GetUserExamProgress(ctx context.Context, userID string, userEaID string) ([]*model.UserExamProgress, error) {
	result, err := queries.GetUserExamProgress(ctx, userID, userEaID)
	if err != nil {
		log.Errorf("Error getting exam progress of a user : %v", err)
		return nil, err
	}
	return result, nil
}

// GetLatestCohorts is the resolver for the getLatestCohorts field.
func (r *queryResolver) GetLatestCohorts(ctx context.Context, userID *string, userLspID *string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedCohorts, error) {
	result, err := queries.GetLatestCohorts(ctx, userID, userLspID, publishTime, pageCursor, direction, pageSize)
	if err != nil {
		log.Errorf("Error getting users cohorts: %v", err)
		return nil, err
	}
	return result, nil
}

// GetCohortUsers is the resolver for the getCohortUsers field.
func (r *queryResolver) GetCohortUsers(ctx context.Context, cohortID string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedCohorts, error) {
	result, err := queries.GetCohortUsers(ctx, cohortID, publishTime, pageCursor, direction, pageSize)
	if err != nil {
		log.Errorf("Error getting users cohorts: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserQuizAttempts is the resolver for the getUserQuizAttempts field.
func (r *queryResolver) GetUserQuizAttempts(ctx context.Context, userID string, topicID string) ([]*model.UserQuizAttempt, error) {
	result, err := queries.GetUserQuizAttempts(ctx, userID, topicID)
	if err != nil {
		log.Errorf("Error getting quiz attempts of a user : %v", err)
		return nil, err
	}
	return result, nil
}

// GetCohortDetails is the resolver for the getCohortDetails field.
func (r *queryResolver) GetCohortDetails(ctx context.Context, cohortID string) (*model.CohortMain, error) {
	result, err := queries.GetCohortDetails(ctx, cohortID)
	if err != nil {
		log.Errorf("Error getting cohort main : %v", err)
		return nil, err
	}
	return result, nil
}

// GetCohorts is the resolver for the getCohorts field.
func (r *queryResolver) GetCohorts(ctx context.Context, cohortIds []*string) ([]*model.CohortMain, error) {
	result, err := queries.GetCohorts(ctx, cohortIds)
	if err != nil {
		log.Errorf("Error getting cohort main : %v", err)
		return nil, err
	}
	return result, nil
}

// GetCohortMains is the resolver for the getCohortMains field.
func (r *queryResolver) GetCohortMains(ctx context.Context, lspID string, publishTime *int, pageCursor *string, direction *string, pageSize *int, searchText *string) (*model.PaginatedCohortsMain, error) {
	result, err := queries.GetCohortMains(ctx, lspID, publishTime, pageCursor, direction, pageSize, searchText)
	if err != nil {
		log.Errorf("Error getting cohorts: %v", err)
		return nil, err
	}
	return result, nil
}

// GetOrganizations is the resolver for the getOrganizations field.
func (r *queryResolver) GetOrganizations(ctx context.Context, orgIds []*string) ([]*model.Organization, error) {
	result, err := orgs.GetOrganizations(ctx, orgIds)
	if err != nil {
		log.Errorf("Error getting organizations: %v", err)
		return nil, err
	}
	return result, nil
}

// GetOrganizationsByName is the resolver for the getOrganizationsByName field.
func (r *queryResolver) GetOrganizationsByName(ctx context.Context, name *string, prevPageSnapShot string, pageSize int) ([]*model.Organization, error) {
	result, err := orgs.GetOrganizationsByName(ctx, name, prevPageSnapShot, pageSize)
	if err != nil {
		log.Errorf("Error getting organizations: %v", err)
		return nil, err
	}
	return result, nil
}

// GetOrganizationUnits is the resolver for the getOrganizationUnits field.
func (r *queryResolver) GetOrganizationUnits(ctx context.Context, ouIds []*string) ([]*model.OrganizationUnit, error) {
	result, err := orgs.GetOrganizationUnits(ctx, ouIds)
	if err != nil {
		log.Errorf("Error getting organization units: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUnitsByOrgID is the resolver for the getUnitsByOrgId field.
func (r *queryResolver) GetUnitsByOrgID(ctx context.Context, orgID string) ([]*model.OrganizationUnit, error) {
	result, err := orgs.GetUnitsByOrgID(ctx, orgID)
	if err != nil {
		log.Errorf("Error getting organization units: %v", err)
		return nil, err
	}
	return result, nil
}

// GetLearningSpacesByOrgID is the resolver for the getLearningSpacesByOrgId field.
func (r *queryResolver) GetLearningSpacesByOrgID(ctx context.Context, orgID string) ([]*model.LearningSpace, error) {
	result, err := orgs.GetLearningSpacesByOrgID(ctx, orgID)
	if err != nil {
		log.Errorf("Error getting learning spaces: %v", err)
		return nil, err
	}
	return result, nil
}

// GetLearningSpacesByOuID is the resolver for the getLearningSpacesByOuId field.
func (r *queryResolver) GetLearningSpacesByOuID(ctx context.Context, ouID string, orgID string) ([]*model.LearningSpace, error) {
	result, err := orgs.GetLearningSpacesByOuID(ctx, ouID, orgID)
	if err != nil {
		log.Errorf("Error getting learning spaces: %v", err)
		return nil, err
	}
	return result, nil
}

// GetLearningSpaceDetails is the resolver for the getLearningSpaceDetails field.
func (r *queryResolver) GetLearningSpaceDetails(ctx context.Context, lspIds []*string) ([]*model.LearningSpace, error) {
	result, err := orgs.GetLearningSpaceDetails(ctx, lspIds)
	if err != nil {
		log.Errorf("Error getting learning spaces: %v", err)
		return nil, err
	}
	return result, nil
}

// GetUserLspRoles is the resolver for the getUserLspRoles field.
func (r *queryResolver) GetUserLspRoles(ctx context.Context, userID string, userLspIds []string) ([]*model.UserRole, error) {
	result, err := handlers.GetUserLspRoles(ctx, userID, userLspIds)
	if err != nil {
		log.Errorf("Error getting learning spaces roles for user: %v", err)
		return nil, err
	}
	return result, nil
}

// GetCourseConsumptionStats is the resolver for the getCourseConsumptionStats field.
func (r *queryResolver) GetCourseConsumptionStats(ctx context.Context, lspID string, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedCCStats, error) {
	result, err := queries.GetCourseConsumptionStats(ctx, lspID, pageCursor, direction, pageSize)
	if err != nil {
		log.Errorf("Error getting learning spaces roles for user: %v", err)
		return nil, err
	}
	return result, nil
}

// GetCourseViews is the resolver for the getCourseViews field.
func (r *queryResolver) GetCourseViews(ctx context.Context, lspIds []string, startTime *string, endTime *string) ([]*model.CourseViews, error) {
	result, err := handlers.GetCourseViews(ctx, lspIds, startTime, endTime)
	if err != nil {
		log.Errorf("Error getting learning spaces roles for user: %v", err)
		return nil, err
	}
	return result, nil
}

// GetVendorExperience is the resolver for the getVendorExperience field.
func (r *queryResolver) GetVendorExperience(ctx context.Context, vendorID string, pfID string) ([]*model.ExperienceVendor, error) {
	res, err := handlers.GetVendorExperience(ctx, vendorID, pfID)
	if err != nil {
		log.Println("Got error while getting vendor's experience: %v", err)
		return nil, err
	}
	return res, nil
}

// GetVendorExperienceDetails is the resolver for the getVendorExperienceDetails field.
func (r *queryResolver) GetVendorExperienceDetails(ctx context.Context, vendorID string, pfID string, expID string) (*model.ExperienceVendor, error) {
	res, err := handlers.GetVendorExperienceDetails(ctx, vendorID, pfID, expID)
	if err != nil {
		log.Printf("Got error while getting vendor experience details: %v", err)
		return nil, err
	}
	return res, nil
}

// GetVendors is the resolver for the getVendors field.
func (r *queryResolver) GetVendors(ctx context.Context, lspID *string, filters *model.VendorFilters) ([]*model.Vendor, error) {
	res, err := handlers.GetVendors(ctx, lspID, filters)
	if err != nil {
		log.Println("Error getting vendors list: %v", err)
		return nil, err
	}
	return res, err
}

// GetPaginatedVendors is the resolver for the getPaginatedVendors field.
func (r *queryResolver) GetPaginatedVendors(ctx context.Context, lspID *string, pageCursor *string, direction *string, pageSize *int, filters *model.VendorFilters) (*model.PaginatedVendors, error) {
	res, err := handlers.GetPaginatedVendors(ctx, lspID, pageCursor, direction, pageSize, filters)
	if err != nil {
		log.Printf("Got error while getting paginated vendors: %v", err)
		return nil, err
	}
	return res, nil
}

// GetVendorAdmins is the resolver for the getVendorAdmins field.
func (r *queryResolver) GetVendorAdmins(ctx context.Context, vendorID string) ([]*model.UserWithLspStatus, error) {
	resp, err := handlers.GetVendorAdmins(ctx, vendorID)
	if err != nil {
		log.Printf("Got error while getting Vendor Admins: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetVendorDetails is the resolver for the getVendorDetails field.
func (r *queryResolver) GetVendorDetails(ctx context.Context, vendorID string) (*model.Vendor, error) {
	res, err := handlers.GetVendorDetails(ctx, vendorID)
	if err != nil {
		log.Println("Got error while getting vendor details: %v", err)
		return nil, err
	}
	return res, nil
}

// ViewProfileVendorDetails is the resolver for the viewProfileVendorDetails field.
func (r *queryResolver) ViewProfileVendorDetails(ctx context.Context, vendorID string, email string) (*model.VendorProfile, error) {
	res, err := handlers.ViewProfileVendorDetails(ctx, vendorID, email)
	if err != nil {
		log.Printf("Got error while getting details of the vendor: %v", err)
		return nil, err
	}
	return res, nil
}

// ViewAllProfiles is the resolver for the viewAllProfiles field.
func (r *queryResolver) ViewAllProfiles(ctx context.Context, vendorID string, filter *string, name *string) ([]*model.VendorProfile, error) {
	res, err := handlers.ViewAllProfiles(ctx, vendorID, filter, name)
	if err != nil {
		log.Printf("Got error while getting details of the vendor: %v", err)
		return nil, err
	}
	return res, nil
}

// GetSampleFiles is the resolver for the getSampleFiles field.
func (r *queryResolver) GetSampleFiles(ctx context.Context, vendorID string, pType string) ([]*model.SampleFile, error) {
	res, err := handlers.GetSampleFiles(ctx, vendorID, pType)
	if err != nil {
		log.Printf("error while getting sample files for vendor: %v", err)
		return nil, err
	}
	return res, nil
}

// GetSmeDetails is the resolver for the getSmeDetails field.
func (r *queryResolver) GetSmeDetails(ctx context.Context, vendorID string) (*model.Sme, error) {
	res, err := handlers.GetSmeDetails(ctx, vendorID)
	if err != nil {
		log.Printf("error while getting SME details for vendor: %v", err)
		return nil, err
	}
	return res, nil
}

// GetClassRoomTraining is the resolver for the getClassRoomTraining field.
func (r *queryResolver) GetClassRoomTraining(ctx context.Context, vendorID string) (*model.Crt, error) {
	res, err := handlers.GetClassRoomTraining(ctx, vendorID)
	if err != nil {
		log.Printf("error while getting classroom training data vendor: %v", err)
		return nil, err
	}
	return res, nil
}

// GetContentDevelopment is the resolver for the getContentDevelopment field.
func (r *queryResolver) GetContentDevelopment(ctx context.Context, vendorID string) (*model.ContentDevelopment, error) {
	res, err := handlers.GetContentDevelopment(ctx, vendorID)
	if err != nil {
		log.Printf("error while getting content development for vendor: %v", err)
		return nil, err
	}
	return res, nil
}

// GetUserVendor is the resolver for the getUserVendor field.
func (r *queryResolver) GetUserVendor(ctx context.Context, userID *string) ([]*model.Vendor, error) {
	res, err := handlers.GetUserVendors(ctx, userID)
	if err != nil {
		log.Printf("error while getting users for vendor: %v", err)
		return nil, err
	}
	return res, nil
}

// GetVendorServices is the resolver for the getVendorServices field.
func (r *queryResolver) GetVendorServices(ctx context.Context, vendorID *string) ([]*string, error) {
	res, err := handlers.GetVendorServices(ctx, vendorID)
	if err != nil {
		log.Printf("error while getting services of vendor: %v", err)
		return nil, err
	}
	return res, nil
}

// GetLspUsersRoles is the resolver for the getLspUsersRoles field.
func (r *queryResolver) GetLspUsersRoles(ctx context.Context, lspID string, userID []*string, userLspID []*string) ([]*model.UserDetailsRole, error) {
	res, err := handlers.GetLspUsersRoles(ctx, lspID, userID, userLspID)
	if err != nil {
		log.Printf("error getting user details with roles: %v", err)
		return nil, err
	}
	return res, nil
}

// GetPaginatedLspUsersWithRoles is the resolver for the getPaginatedLspUsersWithRoles field.
func (r *queryResolver) GetPaginatedLspUsersWithRoles(ctx context.Context, lspID string, role []*string, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedUserDetailsWithRole, error) {
	res, err := handlers.GetPaginatedLspUsersWithRoles(ctx, lspID, role, pageCursor, direction, pageSize)
	if err != nil {
		log.Printf("error getting user details with roles: %v", err)
		return nil, err
	}
	return res, nil
}

// GetAllOrders is the resolver for the getAllOrders field.
func (r *queryResolver) GetAllOrders(ctx context.Context, lspID *string, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedVendorOrder, error) {
	res, err := handlers.GetAllOrders(ctx, lspID, pageCursor, direction, pageSize)
	if err != nil {
		log.Printf("error getting orders of LSP: %v", err)
		return nil, err
	}
	return res, nil
}

// GetOrderServices is the resolver for the getOrderServices field.
func (r *queryResolver) GetOrderServices(ctx context.Context, orderID []*string) ([]*model.OrderServices, error) {
	res, err := handlers.GetOrderServices(ctx, orderID)
	if err != nil {
		log.Printf("error getting services of order: %v", err)
		return nil, err
	}
	return res, nil
}

// GetSpeakers is the resolver for the getSpeakers field.
func (r *queryResolver) GetSpeakers(ctx context.Context, lspID *string, service *string, name *string) ([]*model.VendorProfile, error) {
	res, err := handlers.GetSpeakers(ctx, lspID, service, name)
	if err != nil {
		log.Printf("error getting profiles: %v", err)
		return nil, err
	}
	return res, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
