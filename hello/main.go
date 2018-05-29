package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type CodebuildStateArtifact struct {
	Location string `json:"Location"`
}
type CodeBuildStateAdditionalInfo struct {
	Artifact  CodebuildStateArtifact `json:"Artifact"`
	Initiator string                 `json:"Initiator"`
}
type CodebuildStateChangeDetail struct {
	BuildId               string                       `json:"Build-id"`
	ProjectName           string                       `json:"Project-name"`
	AdditionalInformation CodeBuildStateAdditionalInfo `json:"Additional-information"`
}

// API call responses have to provide CORS headers manually
var DefaultResponseCorsHeaders = map[string]string{
	"Access-Control-Allow-Origin":      "*",
	"Access-Control-Allow-Credentials": "true",
}

func Handler(req events.CloudWatchEvent) (events.APIGatewayProxyResponse, error) {
	log.Println("Handler called")
	var outString = "{ Event: { Source: \"" + req.Source + "\", ID: \"" + req.ID + "\", DetailType: \"" + req.DetailType + "\"} } "
	log.Println(outString)
	var detailRecord CodebuildStateChangeDetail
	err := json.Unmarshal([]byte(req.Detail), &detailRecord)

	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       outString,
			Headers:    DefaultResponseCorsHeaders,
		}, nil
	}
	log.Println("Detail: ID:" + detailRecord.BuildId + " Project: " + detailRecord.ProjectName + " Initiator: " + detailRecord.AdditionalInformation.Initiator + " Location: " + detailRecord.AdditionalInformation.Artifact.Location)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       outString,
		Headers:    DefaultResponseCorsHeaders,
	}, nil
}

func main() {
	log.Println("Hello called")
	lambda.Start(Handler)
}
