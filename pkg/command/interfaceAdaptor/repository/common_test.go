package repository

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
)

func CreateContainer(ctx context.Context) (*localstack.LocalStackContainer, error) {
	container, err := localstack.RunContainer(
		ctx,
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "localstack/localstack:2.1.0",
				Env: map[string]string{
					"SERVICES":              "dynamodb",
					"DEFAULT_REGION":        "us-east-1",
					"EAGER_SERVICE_LOADING": "1",
					"DYNAMODB_SHARED_DB":    "1",
					"DYNAMODB_IN_MEMORY":    "1",
				},
			},
		}),
	)
	return container, err
}
