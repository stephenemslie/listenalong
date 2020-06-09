package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/stephenemslie/listenalong/api"
)

var gorillaLambda *gorillamux.GorillaMuxAdapter

func init() {
	router := api.NewRouter()
	gorillaLambda = gorillamux.New(router)
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return gorillaLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(handler)
}
