package main

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/codecommit"
	//	"strings"
	"log"
)

type ItemInfo struct {
	Project string `json:"Project"`
	Source  string `json:"Source"`
	Branch  string `json:"Branch"`
}

// API call responses have to provide CORS headers manually
var DefaultResponseCorsHeaders = map[string]string{
	"Access-Control-Allow-Origin":      "*",
	"Access-Control-Allow-Credentials": "true",
}

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("triggerBuild function called: " + req.HTTPMethod)
	switch req.HTTPMethod {
	case "POST":
		var item ItemInfo
		err := json.Unmarshal([]byte(req.Body), &item)
		if err != nil {
			return ServerError(err)
		}
		log.Println("Item Project: " + item.Project + " Branch: " + item.Branch + " Source: " + item.Source)

		// Update the node in the database
		buildOutput, err := startBuild(item)
		if err != nil {
			return ServerError(err)
		}

		// Return the updated node as json
		/*		js, err := json.Marshal(item)
				if err != nil {
					return ServerError(err)
				}
		*/
		var buildArn = buildOutput.Build.Arn
		var buildSource = buildOutput.Build.Source.Location
		var buildSourceVersion = buildOutput.Build.SourceVersion
		var outString = "{ Build: { Arn: \"" + *buildArn + "\", Source: \"" + *buildSource + "\", SourceVersion: \"" + *buildSourceVersion + "\" } }"
		log.Println(outString)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       outString,
			Headers:    DefaultResponseCorsHeaders,
		}, nil

	default:
		return ClientError(http.StatusMethodNotAllowed, "Bad request method: "+req.HTTPMethod)
	}
}
func startBuild(item ItemInfo) (*codebuild.StartBuildOutput, error) {
	cc := codecommit.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))
	var repositoryInput codecommit.GetRepositoryInput
	repositoryInput.SetRepositoryName(item.Source)
	repositoryInfoOutput, repositoryOutputErr := cc.GetRepository(&repositoryInput)
	if repositoryOutputErr != nil {
		log.Println("Repository Info Error")
		return nil, repositoryOutputErr
	}

	var branchInput codecommit.GetBranchInput
	branchInput.SetBranchName(item.Branch)
	branchInput.SetRepositoryName(item.Source)
	branchInfoOutput, branchOutputErr := cc.GetBranch(&branchInput)
	if branchOutputErr != nil {
		log.Println("Branch Info Error")
		return nil, branchOutputErr
	}
	commitId := branchInfoOutput.Branch.CommitId
	sourceURL := repositoryInfoOutput.RepositoryMetadata.CloneUrlHttp

	var cb = codebuild.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))
	var buildInput codebuild.StartBuildInput
	buildInput.SetProjectName(item.Project)
	buildInput.SetSourceVersion(*commitId)
	buildInput.SetSourceLocationOverride(*sourceURL)
	buildOutput, err := cb.StartBuild(&buildInput)
	if err != nil {
		return nil, err
	}
	return buildOutput, nil
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
