# Users Service

A service to handle user access and profile

## Protocol Buffers

Under the users dir in  <b>dinghy-protobuffs</b> [repo](https://github.com/saarwasserman/dinghy-protobuffs)

## Build (bin, docker)

See Makefile's -build- commands


## Deploy (k8s)

This service should be connected to dinghy-auth-api and dinghy-notifications-api which is deployed spearately.

See Makefile's -deploy- command

Note: check the deploy yaml files and set the required secrets and env vars


## Databases

<b>PostgreSQL<b/>

`database: users`

`user: dinghy-users`

Contains the user informaion.

See Makefile's -db- commands to run migration and access db (use .envrc for the connection string)

<b>Redis (In Progress)<b> 

## Related Services

1. [dinghy-auth-api](https://github.com/saarwasserman/dinghy-auth) - authentication and authorization
2. [dinghy-notifications-api](https://github.com/saarwasserman/dinghy-notifications
) - email notifications
