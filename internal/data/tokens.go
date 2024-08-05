package data

import "github.com/saarwasserman/users/internal/validator"


const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}
