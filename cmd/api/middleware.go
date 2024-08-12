package main

import (
	"context"
	"slices"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	interceptorsAuth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/saarwasserman/users/internal/data"
	"github.com/saarwasserman/users/protogen/auth"
)

func (app *application) Authenticator(ctx context.Context) (context.Context, error) {
	token_plaintext, err := interceptorsAuth.AuthFromMD(ctx, "bearer")
	if err != nil {
		app.logger.PrintError(err, nil)
		return nil, status.Error(codes.Unauthenticated, "missing bearer token")
	}

	authResponse, err := app.auth.Authenticate(ctx, &auth.AuthenticationRequest{
		TokenScope:     data.ScopeAuthentication,
		TokenPlaintext: token_plaintext,
	})
	if err != nil {
		app.logger.PrintError(err, nil)
		return ctx, status.Error(codes.Unauthenticated, err.Error())
	}

	ctx = app.contextSetUserId(ctx, authResponse.UserId)

	return ctx, nil
}

func (app *application) AuthMatcher(ctx context.Context, callMeta interceptors.CallMeta) bool {
	methods := []string{"GetUser", "Logout"}
	return slices.Contains(methods, callMeta.Method)
}
