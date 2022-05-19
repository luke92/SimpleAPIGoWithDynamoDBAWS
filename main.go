package main

import (
	//common
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	//GIN
	"net/http"

	"github.com/gin-gonic/gin"

	// Import godotenv
	"github.com/joho/godotenv"

	//AWS DynamoDB
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// album represents data about a record album.
type Album struct {
	ID     string      `json:"id"`
	Title  string      `json:"title"`
	Artist string      `json:"artist"`
	Price  json.Number `json:"price,omitempty"`
}

// albums slice to seed record album data.
var albums = []Album{}

var svc *dynamodb.Client
var tableName string

func main() {

	tableName = "albums"
	svc = initDynamoDB()

	createTableIfNotExists()
	getAllItems()
	//TEST of get item from dynamo
	getItem("ID", "1")
	initAPI()
}

func initDynamoDB() *dynamodb.Client {
	aws_access_key := goDotEnvVariable("AWS_ACCESS_KEY_ID")
	aws_secret_key := goDotEnvVariable("AWS_YOUR_SECRET_KEY")
	aws_token := goDotEnvVariable("AWS_TOKEN")
	aws_region := goDotEnvVariable("AWS_REGION")
	use_static_credentials := goDotEnvVariable("USE_STATIC_CREDENTIALS")

	var cfg aws.Config
	var err error
	if use_static_credentials == "TRUE" {
		//SETUP AWS with Static Credentials
		cfg, err = config.LoadDefaultConfig(
			context.TODO(),
			config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(aws_access_key, aws_secret_key, aws_token),
			),
			config.WithRegion(aws_region),
		)
	} else {
		//SETUP AWS with Credentials IAM or Lambda

		cfg, err = config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
			o.Region = aws_region
			return nil
		})
	}

	if err != nil {
		panic(err)
	}

	svc := dynamodb.NewFromConfig(cfg)
	return svc
}

func initAPI() {
	port := goDotEnvVariable("SERVER_PORT")

	//GIN FRAMEWORK REST API
	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", postAlbums)
	router.PUT("/albums/:id", putAlbum)
	router.DELETE("/albums/:id", deleteAlbum)

	//Default PORT is 8080
	router.Run("localhost:" + port)
}

// use godot package to load/read the .env file and
// return the value of the key
func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func listTables(svc *dynamodb.Client) []string {
	tableNames := []string{}

	p := dynamodb.NewListTablesPaginator(svc, nil, func(o *dynamodb.ListTablesPaginatorOptions) {
		o.StopOnDuplicateToken = true
	})

	for p.HasMorePages() {

		out, err := p.NextPage(context.TODO())
		if err != nil {
			panic(err)
		}

		for _, tn := range out.TableNames {
			fmt.Println(tn)
			tableNames = append(tableNames, tn)
		}
	}
	return tableNames
}

func createTableIfNotExists() {
	tableNames := listTables(svc)

	_, found := Find(tableNames, tableName)
	if !found {

		out, err := svc.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
			AttributeDefinitions: []types.AttributeDefinition{
				{
					AttributeName: aws.String("ID"),
					AttributeType: types.ScalarAttributeTypeS,
				},
				{
					AttributeName: aws.String("Title"),
					AttributeType: types.ScalarAttributeTypeS,
				},
			},
			KeySchema: []types.KeySchemaElement{
				{
					AttributeName: aws.String("ID"),
					KeyType:       types.KeyTypeHash,
				},
				{
					AttributeName: aws.String("Title"),
					KeyType:       types.KeyTypeRange,
				},
			},
			TableName:   aws.String(tableName),
			BillingMode: types.BillingModePayPerRequest,
		})
		if err != nil {
			panic(err)
		}

		fmt.Println(out)
	}
}

func getAllItems() {

	out, err := svc.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(out.Items)
	for _, item := range out.Items {
		fmt.Println(item)
		album := mapDynamoItemToAlbum(item)
		albums = append(albums, album)
	}

}

func getItem(attributeName string, id string) {
	//Hago de esta manera porque la tabla tiene KEY doble (HASH y RANGE)
	out, err := svc.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String(attributeName + " = :hashKey"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hashKey": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		panic(err)
	}

	for _, itemDynamo := range out.Items {
		fmt.Println(itemDynamo)
		album := mapDynamoItemToAlbum(itemDynamo)
		fmt.Println(album)
	}
}

func mapDynamoItemToAlbum(item map[string]types.AttributeValue) Album {
	var album Album
	err := attributevalue.UnmarshalMap(item, &album)
	if err != nil {
		log.Fatalf("unmarshal failed, %v", err)
	}
	return album
}

func putItem(album Album) {
	priceString := fmt.Sprintf("%v", album.Price)
	out, err := svc.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			"ID":     &types.AttributeValueMemberS{Value: album.ID},
			"Title":  &types.AttributeValueMemberS{Value: album.Title},
			"Artist": &types.AttributeValueMemberS{Value: album.Artist},
			"Price":  &types.AttributeValueMemberS{Value: priceString},
		},
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(out.Attributes)
}

func deleteItem(album Album) {
	out, err := svc.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"ID":    &types.AttributeValueMemberS{Value: album.ID},
			"Title": &types.AttributeValueMemberS{Value: album.Title},
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(out.Attributes)
}

// Find takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums)
}

// postAlbums adds an album from JSON received in the request body.
func postAlbums(c *gin.Context) {
	var newAlbum Album

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}

	_, err := getById(newAlbum.ID)
	if err != nil {
		// Add the new album to the slice.
		putItem(newAlbum)
		albums = append(albums, newAlbum)
		c.IndentedJSON(http.StatusCreated, newAlbum)
	} else {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Album " + newAlbum.ID + " exists"})
	}

}

func putAlbum(c *gin.Context) {
	id := c.Param("id")

	var newAlbum Album

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}

	if id != newAlbum.ID {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Id URI is not the same as JSON "})
		return
	}

	_, err := getById(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Album " + err.Error()})
	} else {
		// Add the new album to the slice.
		putItem(newAlbum)
		albums = updateAlbum(albums, newAlbum)
		c.IndentedJSON(http.StatusCreated, newAlbum)
	}
}

func updateAlbum(albums []Album, newAlbum Album) []Album {
	for i := 0; i < len(albums); i++ {
		if albums[i].ID == newAlbum.ID {
			albums[i] = newAlbum
			break
		}
	}
	return albums
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumByID(c *gin.Context) {
	id := c.Param("id")

	// Loop through the list of albums, looking for
	// an album whose ID value matches the parameter.

	album, err := getById(id)

	if err == nil {
		c.IndentedJSON(http.StatusOK, album)
		return
	}

	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Album " + err.Error()})
}

func deleteAlbum(c *gin.Context) {
	id := c.Param("id")

	index, album := getIndex(id)
	if index == -1 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Album " + ErrorNotFound.Error()})
	} else {
		// Add the new album to the slice.
		deleteItem(album)
		albums = RemoveIndex(albums, index)
		c.IndentedJSON(http.StatusOK, album)
	}
}

var (
	ErrorNotFound = errors.New("not found")
)

func getById(id string) (Album, error) {
	var album Album
	for _, a := range albums {
		if a.ID == id {
			return a, nil
		}
	}
	return album, ErrorNotFound
}

func getIndex(id string) (int, Album) {
	var album Album
	for i := 0; i < len(albums); i++ {
		if albums[i].ID == id {
			album = albums[i]
			return i, album
		}
	}
	return -1, album
}

func RemoveIndex(albums []Album, index int) []Album {
	return append(albums[:index], albums[index+1:]...)
}
