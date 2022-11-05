package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/zicops-user-manager/graph/generated"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/handlers"
	"github.com/zicops/zicops-user-manager/handlers/queries"
)

func (r *mutationResolver) RegisterUsers(ctx context.Context, input []*model.UserInput) ([]*model.User, error) {
	result, err := handlers.RegisterUsers(ctx, input, true, false)
	if err != nil {
		log.Errorf("Error registering users: %v", err)
		return nil, err
	}
	return result, nil
}

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

func (r *mutationResolver) UpdateUser(ctx context.Context, input model.UserInput) (*model.User, error) {
	result, err := handlers.UpdateUser(ctx, input)
	if err != nil {
		log.Errorf("Error updating user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) Login(ctx context.Context) (*model.User, error) {
	result, err := handlers.LoginUser(ctx)
	if err != nil {
		log.Errorf("Error logging in user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) AddUserLspMap(ctx context.Context, input []*model.UserLspMapInput) ([]*model.UserLspMap, error) {
	isAdmin := false
	result, err := handlers.AddUserLspMap(ctx, input, &isAdmin)
	if err != nil {
		log.Errorf("Error adding lsp map for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) UpdateUserLspMap(ctx context.Context, input model.UserLspMapInput) (*model.UserLspMap, error) {
	result, err := handlers.UpdateUserLspMap(ctx, input)
	if err != nil {
		log.Errorf("Error updating lsp map for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) AddUserOrganizationMap(ctx context.Context, input []*model.UserOrganizationMapInput) ([]*model.UserOrganizationMap, error) {
	result, err := handlers.AddUserOrganizationMap(ctx, input)
	if err != nil {
		log.Errorf("Error adding org map for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) UpdateUserOrganizationMap(ctx context.Context, input model.UserOrganizationMapInput) (*model.UserOrganizationMap, error) {
	result, err := handlers.UpdateUserOrganizationMap(ctx, input)
	if err != nil {
		log.Errorf("Error updating org map for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) AddUserLanguageMap(ctx context.Context, input []*model.UserLanguageMapInput) ([]*model.UserLanguageMap, error) {
	result, err := handlers.AddUserLanguageMap(ctx, input)
	if err != nil {
		log.Errorf("Error adding lang map for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) AddUserPreference(ctx context.Context, input []*model.UserPreferenceInput) ([]*model.UserPreference, error) {
	result, err := handlers.AddUserPreference(ctx, input)
	if err != nil {
		log.Errorf("Error adding preference map for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) UpdateUserPreference(ctx context.Context, input model.UserPreferenceInput) (*model.UserPreference, error) {
	result, err := handlers.UpdateUserPreference(ctx, input)
	if err != nil {
		log.Errorf("Error updating preference map for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) AddUserRoles(ctx context.Context, input []*model.UserRoleInput) ([]*model.UserRole, error) {
	result, err := handlers.AddUserRoles(ctx, input)
	if err != nil {
		log.Errorf("Error adding roles map for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) UpdateUserRole(ctx context.Context, input model.UserRoleInput) (*model.UserRole, error) {
	result, err := handlers.UpdateUserRole(ctx, input)
	if err != nil {
		log.Errorf("Error updating roles map for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) AddUserCohort(ctx context.Context, input []*model.UserCohortInput) ([]*model.UserCohort, error) {
	result, err := handlers.AddUserCohort(ctx, input)
	if err != nil {
		log.Errorf("Error adding cohort map for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) UpdateUserCohort(ctx context.Context, input model.UserCohortInput) (*model.UserCohort, error) {
	result, err := handlers.UpdateUserCohort(ctx, input)
	if err != nil {
		log.Errorf("Error updating cohort map for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) AddUserCourse(ctx context.Context, input []*model.UserCourseInput) ([]*model.UserCourse, error) {
	result, err := handlers.AddUserCourse(ctx, input)
	if err != nil {
		log.Errorf("Error adding course map for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) UpdateUserCourse(ctx context.Context, input model.UserCourseInput) (*model.UserCourse, error) {
	result, err := handlers.UpdateUserCourse(ctx, input)
	if err != nil {
		log.Errorf("Error updating course map for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) AddUserCourseProgress(ctx context.Context, input []*model.UserCourseProgressInput) ([]*model.UserCourseProgress, error) {
	result, err := handlers.AddUserCourseProgress(ctx, input)
	if err != nil {
		log.Errorf("Error adding course progress for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) UpdateUserCourseProgress(ctx context.Context, input model.UserCourseProgressInput) (*model.UserCourseProgress, error) {
	result, err := handlers.UpdateUserCourseProgress(ctx, input)
	if err != nil {
		log.Errorf("Error updating course progress for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) AddUserQuizAttempt(ctx context.Context, input []*model.UserQuizAttemptInput) ([]*model.UserQuizAttempt, error) {
	result, err := handlers.AddUserQuizAttempt(ctx, input)
	if err != nil {
		log.Errorf("Error adding quiz attempt for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) UpdateUserQuizAttempt(ctx context.Context, input model.UserQuizAttemptInput) (*model.UserQuizAttempt, error) {
	result, err := handlers.UpdateUserQuizAttempt(ctx, input)
	if err != nil {
		log.Errorf("Error updating quiz attempt for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) AddUserBookmark(ctx context.Context, input []*model.UserBookmarkInput) ([]*model.UserBookmark, error) {
	result, err := handlers.AddUserBookmark(ctx, input)
	if err != nil {
		log.Errorf("Error adding bookmark for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) UpdateUserBookmark(ctx context.Context, input model.UserBookmarkInput) (*model.UserBookmark, error) {
	result, err := handlers.UpdateUserBookmark(ctx, input)
	if err != nil {
		log.Errorf("Error updating bookmark for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) AddUserNotes(ctx context.Context, input []*model.UserNotesInput) ([]*model.UserNotes, error) {
	result, err := handlers.AddUserNotes(ctx, input)
	if err != nil {
		log.Errorf("Error adding notes for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) UpdateUserNotes(ctx context.Context, input model.UserNotesInput) (*model.UserNotes, error) {
	result, err := handlers.UpdateUserNotes(ctx, input)
	if err != nil {
		log.Errorf("Error updating notes for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) AddUserExamAttempts(ctx context.Context, input []*model.UserExamAttemptsInput) ([]*model.UserExamAttempts, error) {
	result, err := handlers.AddUserExamAttempts(ctx, input)
	if err != nil {
		log.Errorf("Error adding exams for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) UpdateUserExamAttempts(ctx context.Context, input model.UserExamAttemptsInput) (*model.UserExamAttempts, error) {
	result, err := handlers.UpdateUserExamAttempts(ctx, input)
	if err != nil {
		log.Errorf("Error updating exams for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) AddUserExamProgress(ctx context.Context, input []*model.UserExamProgressInput) ([]*model.UserExamProgress, error) {
	result, err := handlers.AddUserExamProgress(ctx, input)
	if err != nil {
		log.Errorf("Error adding exam progress for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) UpdateUserExamProgress(ctx context.Context, input model.UserExamProgressInput) (*model.UserExamProgress, error) {
	result, err := handlers.UpdateUserExamProgress(ctx, input)
	if err != nil {
		log.Errorf("Error updating exam progress for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) AddUserExamResult(ctx context.Context, input []*model.UserExamResultInput) ([]*model.UserExamResult, error) {
	result, err := handlers.AddUserExamResult(ctx, input)
	if err != nil {
		log.Errorf("Error adding exam result for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) UpdateUserExamResult(ctx context.Context, input model.UserExamResultInput) (*model.UserExamResult, error) {
	result, err := handlers.UpdateUserExamResult(ctx, input)
	if err != nil {
		log.Errorf("Error updating exam result for user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) AddCohortMain(ctx context.Context, input model.CohortMainInput) (*model.CohortMain, error) {
	result, err := queries.AddCohortMain(ctx, input)
	if err != nil {
		log.Errorf("Error adding cohort: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) UpdateCohortMain(ctx context.Context, input model.CohortMainInput) (*model.CohortMain, error) {
	result, err := queries.UpdateCohortMain(ctx, input)
	if err != nil {
		log.Errorf("Error updating cohort: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) Logout(ctx context.Context) (*bool, error) {
	result, err := handlers.Logout(ctx)
	if err != nil {
		log.Errorf("Error logging out user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetUserLspMapsByLspID(ctx context.Context, lspID string, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedUserLspMaps, error) {
	result, err := queries.GetUserLspMapsByLspID(ctx, lspID, pageCursor, direction, pageSize)
	if err != nil {
		log.Errorf("Error getting lsps of users: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetUsersForAdmin(ctx context.Context, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedUsers, error) {
	result, err := queries.GetUsersForAdmin(ctx, publishTime, pageCursor, direction, pageSize)
	if err != nil {
		log.Errorf("Error getting users of an admin: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetUserDetails(ctx context.Context, userIds []*string) ([]*model.User, error) {
	result, err := queries.GetUserDetails(ctx, userIds)
	if err != nil {
		log.Errorf("Error getting user of an admin: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetUserOrganizations(ctx context.Context, userID string) ([]*model.UserOrganizationMap, error) {
	result, err := queries.GetUserOrganizations(ctx, userID)
	if err != nil {
		log.Errorf("Error getting orgs of a user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetUserOrgDetails(ctx context.Context, userID string, userLspID string) (*model.UserOrganizationMap, error) {
	result, err := queries.GetUserOrgDetails(ctx, userID, userLspID)
	if err != nil {
		log.Errorf("Error getting orgs of a user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetUserPreferences(ctx context.Context, userID string) ([]*model.UserPreference, error) {
	result, err := queries.GetUserPreferences(ctx, userID)
	if err != nil {
		log.Errorf("Error getting prefs of a user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetUserPreferenceForLsp(ctx context.Context, userID string, userLspID string) (*model.UserPreference, error) {
	result, err := queries.GetUserPreferenceForLsp(ctx, userID, userLspID)
	if err != nil {
		log.Errorf("Error getting prefs of a user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetUserLsps(ctx context.Context, userID string) ([]*model.UserLspMap, error) {
	result, err := queries.GetUserLsps(ctx, userID)
	if err != nil {
		log.Errorf("Error getting lsps of a user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetUserLspByLspID(ctx context.Context, userID string, lspID string) (*model.UserLspMap, error) {
	result, err := queries.GetUserLspByLspID(ctx, userID, lspID)
	if err != nil {
		log.Errorf("Error getting lsps of a user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetUserCourseMaps(ctx context.Context, userID string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedCourseMaps, error) {
	result, err := queries.GetUserCourseMaps(ctx, userID, publishTime, pageCursor, direction, pageSize)
	if err != nil {
		log.Errorf("Error getting courses of a user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetUserCourseMapByCourseID(ctx context.Context, userID string, courseID string) ([]*model.UserCourse, error) {
	result, err := queries.GetUserCourseMapByCourseID(ctx, userID, courseID)
	if err != nil {
		log.Errorf("Error getting course of a user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetUserCourseProgressByMapID(ctx context.Context, userID string, userCourseID []string) ([]*model.UserCourseProgress, error) {
	result, err := queries.GetUserCourseProgressByMapID(ctx, userID, userCourseID)
	if err != nil {
		log.Errorf("Error getting course progress of a user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetUserCourseProgressByTopicID(ctx context.Context, userID string, topicID string) ([]*model.UserCourseProgress, error) {
	result, err := queries.GetUserCourseProgressByTopicID(ctx, userID, topicID)
	if err != nil {
		log.Errorf("Error getting course progress of a user by topic id: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetUserNotes(ctx context.Context, userID string, userLspID *string, courseID *string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedNotes, error) {
	result, err := queries.GetUserNotes(ctx, userID, userLspID, courseID, publishTime, pageCursor, direction, pageSize)
	if err != nil {
		log.Errorf("Error getting notes of a user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetUserBookmarks(ctx context.Context, userID string, userLspID *string, courseID *string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedBookmarks, error) {
	result, err := queries.GetUserBookmarks(ctx, userID, userLspID, courseID, publishTime, pageCursor, direction, pageSize)
	if err != nil {
		log.Errorf("Error getting bookmarks of a user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetUserExamAttempts(ctx context.Context, userID string, userLspID string) ([]*model.UserExamAttempts, error) {
	result, err := queries.GetUserExamAttempts(ctx, userID, userLspID)
	if err != nil {
		log.Errorf("Error getting exam attempts of a user : %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetUserExamResults(ctx context.Context, userID string, userEaID string) (*model.UserExamResult, error) {
	result, err := queries.GetUserExamResults(ctx, userID, userEaID)
	if err != nil {
		log.Errorf("Error getting exam results of a user : %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetUserExamProgress(ctx context.Context, userID string, userEaID string) ([]*model.UserExamProgress, error) {
	result, err := queries.GetUserExamProgress(ctx, userID, userEaID)
	if err != nil {
		log.Errorf("Error getting exam progress of a user : %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetLatestCohorts(ctx context.Context, userID *string, userLspID *string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedCohorts, error) {
	result, err := queries.GetLatestCohorts(ctx, userID, userLspID, publishTime, pageCursor, direction, pageSize)
	if err != nil {
		log.Errorf("Error getting users cohorts: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetCohortUsers(ctx context.Context, cohortID string, publishTime *int, pageCursor *string, direction *string, pageSize *int) (*model.PaginatedCohorts, error) {
	result, err := queries.GetCohortUsers(ctx, cohortID, publishTime, pageCursor, direction, pageSize)
	if err != nil {
		log.Errorf("Error getting users cohorts: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetUserQuizAttempts(ctx context.Context, userID string, topicID string) ([]*model.UserQuizAttempt, error) {
	result, err := queries.GetUserQuizAttempts(ctx, userID, topicID)
	if err != nil {
		log.Errorf("Error getting quiz attempts of a user : %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetCohortDetails(ctx context.Context, cohortID string) (*model.CohortMain, error) {
	result, err := queries.GetCohortDetails(ctx, cohortID)
	if err != nil {
		log.Errorf("Error getting cohort main : %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) GetCohortMains(ctx context.Context, lspID string, publishTime *int, pageCursor *string, direction *string, pageSize *int, searchText *string) (*model.PaginatedCohortsMain, error) {
	result, err := queries.GetCohortMains(ctx, lspID, publishTime, pageCursor, direction, pageSize, searchText)
	if err != nil {
		log.Errorf("Error getting cohorts: %v", err)
		return nil, err
	}
	return result, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
