// Package identity that handles signup, reset password, verification of email etc.
// This is an admin package be careful while using these functions .....
package identity

import (
	"context"
	"fmt"
	"os"
	"strings"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/zicops/zicops-user-manager/constants"
	"github.com/zicops/zicops-user-manager/helpers"
	"google.golang.org/api/option"
)

// IDP ... client managing email/password related authorization
type IDP struct {
	projectID string
	client    *auth.Client
	tClient   *auth.TenantClient
	fireApp   *firebase.App
}

// NewIDPEP ..... intializes firebase auth which will do al sorts of authn/authz
func NewIDPEP(ctx context.Context, projectID string) (*IDP, error) {
	serviceAccountSD := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if serviceAccountSD == "" {
		return nil, fmt.Errorf("Missing service account file for backend server")
	}
	targetScopes := []string{
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/userinfo.email",
	}
	currentCreds, _, err := helpers.ReadCredentialsFile(ctx, serviceAccountSD, targetScopes)
	opt := option.WithCredentials(currentCreds)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize identity client")
	}
	currentClient, err := app.Auth(ctx)
	var tenantClient *auth.TenantClient
	if false {
		tenantClient, err = currentClient.TenantManager.AuthForTenant("firebaseprodtenant")
		if err != nil {
			return nil, fmt.Errorf("Failed to initialize identity client")
		}
	} else {
		tenantClient = nil
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize identity client")
	}
	return &IDP{projectID: projectID, client: currentClient, fireApp: app, tClient: tenantClient}, nil
}

// ResetUserPassword ...
func (id *IDP) ResetUserPassword(ctx context.Context, email string) error {
	currentResetLink := ""
	var err error
	if id.tClient != nil {
		currentResetLink, err = id.tClient.PasswordResetLink(ctx, email)
		if err != nil {
			return err
		}
	} else {
		currentResetLink, err = id.client.PasswordResetLink(ctx, email)
		if err != nil {
			return err
		}
	}
	// TODO: Implement SMTP server from GSuite/Others to send out custom emails
	// Would also need HTML template for the same
	fmt.Println(currentResetLink)
	return nil
}

// GetUserByEmail ...
func (id *IDP) GetUserByEmail(ctx context.Context, email string) (*auth.UserRecord, error) {
	var err error
	var currentUser *auth.UserRecord
	if id.tClient != nil {
		currentUser, err = id.tClient.GetUserByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
	} else {
		currentUser, err = id.client.GetUserByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
	}
	return currentUser, nil
}

// DeleteAnonymousUser ...
func (id *IDP) DeleteAnonymousUser(ctx context.Context, userID string) error {
	var err error
	if id.tClient != nil {
		err = id.tClient.DeleteUser(ctx, userID)
		if err != nil {
			return err
		}
	} else {
		err = id.client.DeleteUser(ctx, userID)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetVerificationURL ...
func (id *IDP) GetVerificationURL(ctx context.Context, email string) (string, error) {
	contURL := constants.CONTINUE_URL
	actionCode := auth.ActionCodeSettings{
		URL: contURL,
	}
	verifyLink := ""
	var err error
	if id.tClient != nil {
		verifyLink, err = id.tClient.EmailVerificationLinkWithSettings(ctx, email, &actionCode)
		if err != nil {
			return "", err
		}

	} else {
		verifyLink, err = id.client.EmailVerificationLinkWithSettings(ctx, email, &actionCode)
		if err != nil {
			return "", err
		}
	}
	return verifyLink, nil
}

// VerifyInFirebase ...
func (id *IDP) VerifyInFirebase(ctx context.Context, email string, uid string) error {
	userToUpdate := (&auth.UserToUpdate{}).
		EmailVerified(true)

	if id.tClient != nil {
		_, err := id.tClient.UpdateUser(ctx, uid, userToUpdate)
		if err != nil {
			return err
		}

	} else {
		_, err := id.client.UpdateUser(ctx, uid, userToUpdate)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetResetPasswordURL ...
func (id *IDP) GetResetPasswordURL(ctx context.Context, email string) (string, error) {
	contURL := constants.CONTINUE_URL
	actionCode := auth.ActionCodeSettings{
		URL: contURL,
	}
	verifyLink := ""
	var err error
	if id.tClient != nil {
		verifyLink, err = id.tClient.PasswordResetLinkWithSettings(ctx, email, &actionCode)
		if err != nil {
			return "", err
		}
	} else {
		verifyLink, err = id.client.PasswordResetLinkWithSettings(ctx, email, &actionCode)
		if err != nil {
			return "", err
		}
	}
	zicopReplace := strings.Replace(verifyLink, "https://zicops-one.firebaseapp.com/__/auth/action?", "https://zicops.com/reset-password?", 1)
	return zicopReplace, nil
}

// RegisterUser ...
func (id *IDP) RegisterUser(ctx context.Context, email string, firstName string, lastName string, phone string) (*auth.UserRecord, error) {
	var err error
	var currentUser *auth.UserRecord
	params := (&auth.UserToCreate{}).
		Email(email).
		EmailVerified(false).
		DisplayName(firstName + " " + lastName)
	if phone != "" {
		params = params.PhoneNumber(phone)
	}
	if id.tClient != nil {

		currentUser, err = id.tClient.CreateUser(ctx, params)
		if err != nil {
			return nil, err
		}

	} else {

		currentUser, err = id.client.CreateUser(ctx, params)
		if err != nil {
			return nil, err
		}

	}
	return currentUser, nil
}

// InviteUser ...
func (id *IDP) InviteUser(ctx context.Context, email string) (*auth.UserRecord, error) {
	var err error
	var currentUser *auth.UserRecord
	params := (&auth.UserToCreate{}).
		Email(email).
		EmailVerified(false)
	if id.tClient != nil {
		currentUser, err = id.tClient.CreateUser(ctx, params)
		if err != nil {
			return nil, err
		}
	} else {
		currentUser, err = id.client.CreateUser(ctx, params)
		if err != nil {
			return nil, err
		}
	}
	return currentUser, nil
}

// UpdateUser ...
func (id *IDP) UpdateUser(ctx context.Context, email string, firstName string, lastName string, phone string, userId string) (*auth.UserRecord, error) {
	var err error
	var currentUser *auth.UserRecord
	params := (&auth.UserToUpdate{})
	currentUpdate := false
	if firstName != "" && lastName != "" {
		params = params.DisplayName(firstName + " " + lastName)
		currentUpdate = true
	}
	if phone != "" {
		params = params.PhoneNumber(phone)
		currentUpdate = true
	}
	if email != "" {
		params = params.Email(email)
		currentUpdate = true
	}
	if currentUpdate {
		if id.tClient != nil {

			currentUser, err = id.tClient.UpdateUser(ctx, userId, params)
			if err != nil {
				return nil, err
			}

		} else {

			currentUser, err = id.client.UpdateUser(ctx, userId, params)
			if err != nil {
				return nil, err
			}

		}
	}
	return currentUser, nil
}

func (id *IDP) LoginUser(ctx context.Context, email string) (*auth.UserRecord, string, error) {
	var err error
	var currentUser *auth.UserRecord
	var tokenAccess string
	if id.tClient != nil {
		currentUser, err = id.tClient.GetUserByEmail(ctx, email)
		if err != nil {
			return nil, tokenAccess, err
		}

	} else {
		currentUser, err = id.client.GetUserByEmail(ctx, email)
		if err != nil {
			return nil, tokenAccess, err
		}
	}
	return currentUser, tokenAccess, nil
}

// GetUserByEmail ...
func (id *IDP) LogoutUser(ctx context.Context, email string) (bool, error) {
	currentUser, err := id.GetUserByEmail(ctx, email)
	if err != nil {
		return false, err
	}
	if id.tClient != nil {
		err = id.tClient.RevokeRefreshTokens(ctx, currentUser.UID)
		if err != nil {
			return false, err
		}
	} else {
		err = id.client.RevokeRefreshTokens(ctx, currentUser.UID)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

// VerifyUserToken ....
func (id *IDP) VerifyUserToken(ctx context.Context, idToken string) (*auth.Token, error) {
	verificationOutput, err := id.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}
	return verificationOutput, nil
}
