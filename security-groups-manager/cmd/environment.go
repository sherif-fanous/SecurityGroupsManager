package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

const awsLambdaFunctionNameEnvironmentVariableName = "AWS_LAMBDA_FUNCTION_NAME"
const configurationEnvironmentVariableName = "CONFIGURATION"
const debugEnvironmentVariableName = "DEBUG"

type ExecutionEnvironment struct {
	Client        *ec2.Client
	Configuration *Configuration
	DoDebug       bool
	IsLambda      bool
}

func NewExecutionEnvironment(isLambda bool) (*ExecutionEnvironment, error) {
	var err error

	executionEnvironment := new(ExecutionEnvironment)

	executionEnvironment.Client, err = initClient()
	if err != nil {
		return nil, err
	}
	executionEnvironment.Configuration, err = initConfiguration()
	if err != nil {
		return nil, err
	}
	executionEnvironment.DoDebug = initDoDebug()
	executionEnvironment.IsLambda = isLambda

	return executionEnvironment, nil
}

func init() {
	if _, ok := os.LookupEnv(awsLambdaFunctionNameEnvironmentVariableName); ok {
		var err error

		executionEnvironment, err = NewExecutionEnvironment(true)
		if err != nil {
			os.Exit(1)
		}
	}
}

func initClient() (*ec2.Client, error) {
	awsConfiguration, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Printf("Unable to load SDK config: %v", err)

		return nil, err
	}

	return ec2.NewFromConfig(awsConfiguration), nil
}

func initConfiguration() (*Configuration, error) {
	configurationEnvironmentVariableValue := lookupEnvironmentVariable(configurationEnvironmentVariableName)

	return NewConfiguration(configurationEnvironmentVariableValue)
}

func initDoDebug() bool {
	debugEnvironmentVariableValue := lookupEnvironmentVariable(debugEnvironmentVariableName)

	doDebug, err := strconv.ParseBool(debugEnvironmentVariableValue)
	if err != nil {
		log.Printf("Unable to parse %s environment variable: %v", debugEnvironmentVariableName, err)
	}

	return doDebug
}

func lookupEnvironmentVariable(environmentVariableName string) string {
	environmentVariableValue, ok := os.LookupEnv(environmentVariableName)
	if !ok {
		log.Printf("Unable to lookup %s environment variable", environmentVariableName)
	}

	return environmentVariableValue
}
