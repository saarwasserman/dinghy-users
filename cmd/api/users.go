package main

import (
	"context"
	"errors"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/saarwasserman/users/internal/data"
	"github.com/saarwasserman/users/internal/validator"
	"github.com/saarwasserman/users/protogen/auth"
	"github.com/saarwasserman/users/protogen/notifications"
	"github.com/saarwasserman/users/protogen/users"
)

func (app *application) RegisterUser(ctx context.Context, req *users.UserRegisterRequest) (*users.UserDetailsResponse, error) {
	user := &data.User{
		Name:      req.Name,
		Email:     req.Email,
		Activated: false,
	}

	v := validator.New()

	data.ValidatePlaintextPassword(v, req.Password)
	data.ValidateUser(v, user)

	if !v.Valid() {
		return nil, status.Errorf(codes.InvalidArgument, "error %s", v.Errors)
	}

	err := app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			return nil, status.Errorf(codes.InvalidArgument, "error %s", v.Errors)
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	_, err = app.auth.SetPassword(ctx, &auth.SetPasswordRequest{
		UserId:   user.ID,
		Password: req.Password,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to set initial password")
	}

	// add initial permission
	_, err = app.auth.AddPermissionForUser(ctx, &auth.AddPermissionForUserRequest{
		UserId: user.ID,
		Codes:  []string{"movies:read"},
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	tokenResponse, err := app.auth.CreateToken(ctx, &auth.TokenCreationRequest{
		Scope:  data.ScopeActivation,
		UserId: user.ID,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = app.notifier.SendActivationEmail(context.Background(), &notifications.SendActivationEmailRequest{
		Recipient: user.Email,
		UserId:    strconv.FormatInt(user.ID, 10),
		Token:     tokenResponse.TokenPlaintext,
	})
	if err != nil {
		app.logger.PrintFatal(err, nil)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &users.UserDetailsResponse{
		Id:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.UnixMilli(),
		Activated: user.Activated,
	}, nil
}

func (app *application) ActivateUser(ctx context.Context, req *users.UserActivationRequest) (*users.UserDetailsResponse, error) {

	v := validator.New()

	if data.ValidateTokenPlaintext(v, req.TokenPlaintext); !v.Valid() {
		return nil, status.Errorf(codes.InvalidArgument, "error: %s", v.Errors)
	}

	authRes, err := app.auth.Authenticate(ctx, &auth.AuthenticationRequest{
		TokenScope:     data.ScopeActivation,
		TokenPlaintext: req.TokenPlaintext,
	})
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			return nil, status.Errorf(codes.InvalidArgument, "error: %s", v.Errors)
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	user, err := app.models.Users.GetByUserId(authRes.UserId)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			return nil, status.Errorf(codes.InvalidArgument, "error: %s", v.Errors)
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	user.Activated = true

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			return nil, status.Error(codes.InvalidArgument, "unable to update the record due to an edit conflict, please try again")
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	_, err = app.auth.DeleteAllTokensForUser(ctx, &auth.TokensDeletionRequest{Scope: data.ScopeActivation, UserId: user.ID})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &users.UserDetailsResponse{
		Id:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.UnixMilli(),
		Activated: user.Activated,
	}, nil
}

func (app *application) GetUser(ctx context.Context, req *users.UserDetailsRequest) (*users.UserDetailsResponse, error) {

	userId := app.contextGetUserId(ctx)

	user, err := app.models.Users.GetByUserId(userId)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return nil, status.Error(codes.Internal, "user not found")
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &users.UserDetailsResponse{
		Id:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.UnixMilli(),
		Activated: user.Activated,
	}, nil
}

func (app *application) Login(ctx context.Context, req *users.LoginRequest) (*users.LoginResponse, error) {
	user, err := app.models.Users.GetByEmail(req.Email)
	if err != nil {
		app.logger.PrintError(err, nil)
		return nil, err
	}

	tokenResponse, err := app.auth.CreateToken(ctx, &auth.TokenCreationRequest{
		UserId: user.ID,
		Scope:  data.ScopeAuthentication,
	})
	if err != nil {
		app.logger.PrintError(err, nil)
		return nil, err
	}

	return &users.LoginResponse{
		TokenPlaintext: tokenResponse.TokenPlaintext,
	}, nil
}

func (app *application) Logout(ctx context.Context, req *users.LogoutRequest) (*users.LogoutResponse, error) {

	userId := app.contextGetUserId(ctx)

	_, err := app.auth.DeleteAllTokensForUser(ctx, &auth.TokensDeletionRequest{
		Scope:  data.ScopeAuthentication,
		UserId: userId,
	})
	if err != nil {
		app.logger.PrintError(err, nil)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &users.LogoutResponse{}, nil
}
