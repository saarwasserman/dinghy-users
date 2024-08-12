package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/saarwasserman/users/internal/data"
	"github.com/saarwasserman/users/protogen/users"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestRegisterUserFlow(t *testing.T) {
	// create users client
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.NewClient("localhost:40030", opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
		return
	}
	defer conn.Close()

	usersClient := users.NewUsersClient(conn)

	name := fmt.Sprintf("test_%d", time.Now().Unix())
	email := fmt.Sprintf("%s@dinghy.test", name)
	userRegistrationRequest := &users.UserRegisterRequest{
		Email:    email,
		Name:     name,
		Password: "somepassword"}

	// send register user call
	res, err := usersClient.RegisterUser(context.Background(), userRegistrationRequest)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	expectedRes := &users.UserDetailsResponse{
		Name:      name,
		Email:     email,
		Activated: false,
	}

	if res.Name != expectedRes.Name || res.Email != expectedRes.Email || res.Activated != expectedRes.Activated {
		t.Errorf("got %v and expected %v", res, expectedRes)
	}

	now := time.Now().UnixMilli()
	if now-res.CreatedAt > 10000 {
		t.Errorf("creation time might be wrong. diff between created %d and now %d", res.CreatedAt, now)
	}

}

func init() {

	// creates it for the auth service
	// TODO: move to common test initialization repo
	dsn := os.Getenv("USERS_DB_DSN")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	defer db.Close()

	models := data.NewModels(db)

	testUser := &data.User{
		Name:      "Test Auth User",
		Email:     "test_auth_user@dinghy.test",
		Activated: true,
	}

	err = models.Users.Insert(testUser)
	if err != nil {
		if !errors.Is(err, data.ErrDuplicateEmail) {
			log.Fatal(err.Error())
			return
		}
	}
}
