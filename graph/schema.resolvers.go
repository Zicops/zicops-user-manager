package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/zicops-user-manager/graph/generated"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/handlers"
)

func (r *mutationResolver) RegisterUsers(ctx context.Context, input []*model.UserInput) ([]*model.User, error) {
	result, err := handlers.RegisterUsers(ctx, input, true)
	if err != nil {
		log.Errorf("Error registering users: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) InviteUsers(ctx context.Context, emails []string) (*bool, error) {
	result, err := handlers.InviteUsers(ctx, emails)
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
	result, err := handlers.AddUserLspMap(ctx, input)
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
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpdateUserExamProgress(ctx context.Context, input model.UserExamProgressInput) (*model.UserExamProgress, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) AddUserExamResult(ctx context.Context, input []*model.UserExamResultInput) ([]*model.UserExamResult, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpdateUserExamResult(ctx context.Context, input model.UserExamResultInput) (*model.UserExamResult, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Logout(ctx context.Context) (*bool, error) {
	result, err := handlers.Logout(ctx)
	if err != nil {
		log.Errorf("Error logging out user: %v", err)
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
