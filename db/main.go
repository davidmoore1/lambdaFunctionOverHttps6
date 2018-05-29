package main

import (
	"strings"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"encoding/json"
	"net/http"
	"os"
	"fmt"
//	"strings"
	"log"
)
type ItemInfo struct {
    Name string`json:"Name"`
    City string`json:"City"`
    State string`json:"State"`
    Age int`json:"Age"`
}

// API call responses have to provide CORS headers manually
var DefaultResponseCorsHeaders = map[string]string{
	"Access-Control-Allow-Origin":      "*",
	"Access-Control-Allow-Credentials": "true",
}

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("db function called: " + req.HTTPMethod)
	userTable := os.Getenv("USERS_TABLE")
	log.Println("Table name: " + userTable)
	switch req.HTTPMethod {
	case "GET":
		var item ItemInfo
		recKey := req.PathParameters["id"]
		recKey = strings.Replace(recKey, "%20", " ", -1)
		if recKey != "" {
			log.Println("Getting record with name: " + recKey)
		}
		err := getRecord(userTable, recKey, &item)
		if err != nil {
			return ServerError(err)
		}

		if item.Name == "" {
			return ClientError(http.StatusNotFound, "Record not found: "+req.HTTPMethod)
		}
		js, err := json.Marshal(item)
		if err != nil {
			return ServerError(err)
		}
		
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       string(js),
			Headers:    DefaultResponseCorsHeaders,
		}, nil
	case "POST":
		var item ItemInfo
		err := json.Unmarshal([]byte(req.Body), &item)
		if err != nil {
			return ServerError(err)
		}
		log.Println("Item Name: " + item.Name + " City: " + item.City + " State: " + item.State )

		// Update the node in the database
		err = PutItem(userTable, item)
		if err != nil {
			return ServerError(err)
		}

		// Return the updated node as json
		js, err := json.Marshal(item)
		if err != nil {
			return ServerError(err)
		}

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       string(js),
			Headers:    DefaultResponseCorsHeaders,
		}, nil

	default:
		return ClientError(http.StatusMethodNotAllowed, "Bad request method: "+req.HTTPMethod)
	}
}

func getRecord(tableName string, recKey string, itemObj interface{}) error {
	var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))
	// Prepare the input for the query.
	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Name": {
				S: aws.String(recKey),
			},
		},
	}

	// Retrieve the item from DynamoDB. If no matching item is found
	// return nil.
	result, err := db.GetItem(input)
	if err != nil {
		return err
	}
	if result.Item == nil {
		log.Println("No item returned for getItem")
		return nil
	}

	// The result.Item object returned has the underlying type
	// map[string]*AttributeValue. We can use the UnmarshalMap helper
	// to parse this straight into the fields of a struct. Note:
	// UnmarshalListOfMaps also exists if you are working with multiple
	// items.
	err = dynamodbattribute.UnmarshalMap(result.Item, itemObj)
	if err != nil {
		return err
	}
	log.Println("Successfully retrieved record")
	return nil
}
func PutItem(tableName string, item interface{}) error {
	var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))
	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		ServerError(fmt.Errorf("failed to DynamoDB marshal Record, %v", err))
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	}

	_, err = db.PutItem(input)
	return err
}
// Similarly add a helper for send responses relating to client errors.
func ClientError(status int, body string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       body,
		Headers:    DefaultResponseCorsHeaders,
	}, nil
}
// Add a helper for handling errors. This logs any error to os.Stderr
// and returns a 500 Internal Server Error response that the AWS API
// Gateway understands.
func ServerError(err error) (events.APIGatewayProxyResponse, error) {
	log.Println(err.Error())

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
		Headers:    DefaultResponseCorsHeaders,
	}, nil
}
func main() {
	lambda.Start(router)
}
