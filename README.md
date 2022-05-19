# SimpleAPIGoWithDynamoDBAWS
A simple API in GO using DynamoDB for save information

## How to create project with GIN
- Run `go mod init github.com/luke92/SimpleAPIGoWithDynamoDBAWS`
- Run `go get -u github.com/gin-gonic/gin`
- Create `main.go`

## Endpoints
- GET ALL (GET http://localhost:8080/albums)
- GET BY ID (GET http://localhost:8080/albums/1)
- ADD ALBUM (POST http://localhost:8080/albums)
- UPDATE ALBUM (PUT http://localhost:8080/albums/1)
- DELETE ALBUM (DELETE http://localhost:8080/albums/1)

## Structure of JSON ALBUM
```
{
    "id": "4",
    "title": "Vivir asi es morir de amor",
    "artist": "Nathy Peluso",
    "price": 99.99
}
```

## Configuration
- Configure the `.env` file for custom port server, credentials aws and region

## Connect with AWS
- Configure your Access and Secret KEY if you need use static credentials

## Documentation
- https://go.dev/doc/tutorial/web-service-gin (Create Project GIN)
- https://dynobase.dev/dynamodb-golang-query-examples/ (Examples for DynamoDB)
- https://towardsdatascience.com/use-environment-variable-in-your-next-golang-project-39e17c3aaa66 (Configuration)