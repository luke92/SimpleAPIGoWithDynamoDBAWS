# SimpleAPIGoWithDynamoDBAWS
A simple API in GO using DynamoDB for save information

## How to create project with GIN
- Run `go mod init github.com/luke92/SimpleAPIGoWithDynamoDBAWS`
- Run `go get -u github.com/gin-gonic/gin`
- Create `main.go`

## Configuration
- Configure the `.env` file for custom port server, credentials aws and region

## Connect with AWS
- Configure your Access and Secret KEY if you need use static credentials

## Documentation
- https://go.dev/doc/tutorial/web-service-gin (Create Project GIN)
- https://dynobase.dev/dynamodb-golang-query-examples/ (Examples for DynamoDB)
- https://towardsdatascience.com/use-environment-variable-in-your-next-golang-project-39e17c3aaa66 (Configuration)