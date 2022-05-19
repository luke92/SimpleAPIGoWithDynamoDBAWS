package main

import (
	//common
	"context"
	"errors"
	"fmt"

	//GIN
	"net/http"

	"github.com/gin-gonic/gin"

	//AWS DynamoDB
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// album represents data about a record album.
type Album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// albums slice to seed record album data.
var albums = []Album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

func main() {

	//SETUP AWS with Credentials IAM or Lambda
	/*
		cfg, err := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
			o.Region = "us-east-1"
			return nil
		})
	*/

	//SETUP AWS with Static Credentials
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("YOUR_ACCESS_KEY_ID", "YOUR_SECRET_KEY", ""),
		),
		config.WithRegion("us-east-1"),
	)

	if err != nil {
		panic(err)
	}

	svc := dynamodb.NewFromConfig(cfg)

	tableNames := listTables(svc)
	createTableIfNotExists(svc, tableNames, "albums")

	//GIN FRAMEWORK REST API
	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", postAlbums)

	router.Run("localhost:8080")
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

func createTableIfNotExists(svc *dynamodb.Client, tableNames []string, tableName string) {
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

	// Add the new album to the slice.
	albums = append(albums, newAlbum)
	c.IndentedJSON(http.StatusCreated, newAlbum)
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
