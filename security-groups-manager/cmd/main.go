package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
)

var executionEnvironment = new(ExecutionEnvironment)

func debugf(format string, v ...interface{}) {
	if executionEnvironment.DoDebug {
		log.Printf(format, v...)
	}
}

func execute() (*Controller, error) {
	if !executionEnvironment.IsLambda {
		var err error

		executionEnvironment, err = NewExecutionEnvironment(false)
		if err != nil {
			return nil, err
		}
	}

	controller := NewController(executionEnvironment.Client)
	err := controller.InitAsIsSecurityGroups()
	if err != nil {
		return nil, err
	}
	controller.InitToBeSecurityGroups(executionEnvironment.Configuration)
	controller.CalculateSecurityGroupDeltas()
	controller.ProcessSecurityGroupDeltas()

	// Only needed by the Test functions
	return controller, nil
}

func handler() error {
	_, err := execute()
	if err != nil {
		return fmt.Errorf("execution failed")
	}

	return nil
}

func main() {
	if executionEnvironment.IsLambda {
		lambda.Start(handler)
	} else {
		handler()
	}
}
