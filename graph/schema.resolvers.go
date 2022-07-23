package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

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
