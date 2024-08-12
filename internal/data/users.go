package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/saarwasserman/users/internal/validator"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

var AnonymousUser = &User{}

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePlaintextPassword(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)

	// TODO: check password using auth service request
}

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) Insert(user *User) error {

	query := `
		INSERT INTO users (name, email, activated)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{user.Name, user.Email, user.Activated}

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {

	query := `
		SELECT id, created_at, name, email, activated, version
		FROM users
		WHERE email = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Activated,
		&user.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) GetByUserId(userId int64) (*User, error) {

	query := `
		SELECT id, created_at, name, email, activated, version
		FROM users
		WHERE id = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, userId).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Activated,
		&user.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) Update(user *User) error {
	query := `
		UPDATE users
		SET name = $1, email = $2, activated = $3, version = version + 1
		WHERE id = $4 AND version = $5
		RETURNING version`

	args := []any{
		user.Name,
		user.Email,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (int64, error) {

	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
		SELECT user_id
		FROM tokens
		WHERE hash = $1
		AND tokens.scope = $2
		AND tokens.expiry > $3`

	args := []any{tokenHash[:], tokenScope, time.Now()}

	var userId int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&userId)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return -1, ErrRecordNotFound
		default:
			return -1, err
		}
	}

	return userId, nil
}
