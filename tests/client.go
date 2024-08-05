package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/saarwasserman/users/protogen/users"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {

	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.NewClient("localhost:40030", opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
		return
	}
	defer conn.Close()


	// create new user
	// client := users.NewUsersClient(conn)
	// res, err := client.RegisterUser(context.Background(), &users.UserRegisterRequest{ Email: "test15@test.com",
	// Name: "test1", Password: "somepassword",})
	// if err != nil {
	// 	log.Fatal(err.Error())
	// 	return
	// }
	// fmt.Println(res)

	// activate user
	// client := users.NewUsersClient(conn)
	// res, err := client.ActivateUser(context.Background(), &users.UserActivationRequest{ TokenPlaintext: "ZAIA55XXTUHNHCVTQSNHXF7LAE"})
	// if err != nil {
	// 	log.Fatal("couldn't activate ", err.Error())
	// 	return
	// }
	// fmt.Println(res)

	// login
	// client := users.NewUsersClient(conn)
	// res, err := client.Login(context.Background(), &users.LoginRequest{
	// 	Email: "test15@test.com", Password: "somepassword",
	// })
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }
	// fmt.Println(res)

	// get user details
	// client := users.NewUsersClient(conn)
	// md := metadata.Pairs("authorization", "bearer VISIAMIDA5YZ4Y26N5TPLFLR44")
	// ctx := metadata.NewOutgoingContext(context.Background(), md)
	// res, err := client.GetUser(ctx, &users.UserDetailsRequest{})
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }
	// fmt.Println(res)

	// Logout
	client := users.NewUsersClient(conn)
	md := metadata.Pairs("authorization", "bearer VISIAMIDA5YZ4Y26N5TPLFLR44")
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	res, err := client.Logout(ctx, &users.LogoutRequest{})
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(res)

}
