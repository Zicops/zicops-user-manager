package handlers

import (
	"context"

	"github.com/zicops/zicops-user-manager/global"
	"github.com/zicops/zicops-user-manager/graph/model"
)

func RegisterUser(ctx context.Context, input model.RegisterUser) (*bool, error) {
	registered := false
	userRecord, err := global.IDP.RegisterUser(ctx, *input.Email, *input.FirstName, *input.LastName, *input.Phone)
	if err != nil {
		return &registered, err
	}
	registered = true
	passwordReset, err := global.IDP.GetResetPasswordURL(ctx, userRecord.Email)
	if err != nil {
		return &registered, err
	}
	// send email with password reset link
	global.SGClient.SendJoinEmail(userRecord.Email, passwordReset, userRecord.DisplayName)
	return &registered, nil
}

func InviteUsers(ctx context.Context, emails []string) (*bool, error) {
	registered := false
	for _, email := range emails {
		userRecord, err := global.IDP.InviteUser(ctx, email)
		if err != nil {
			return &registered, err
		}
		passwordReset, err := global.IDP.GetResetPasswordURL(ctx, userRecord.Email)
		if err != nil {
			return &registered, err
		}
		// send email with password reset link
		global.SGClient.SendJoinEmail(userRecord.Email, passwordReset, userRecord.DisplayName)
	}
	registered = true
	return &registered, nil
}
