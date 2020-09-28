package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gojek/merlin/client"
	"golang.org/x/oauth2/google"
)

func main() {
	ctx := context.Background()

	basePath := "http://merlin.dev/api/merlin/v1"
	if os.Getenv("MERLIN_API_BASEPATH") != "" {
		basePath = os.Getenv("MERLIN_API_BASEPATH")
	}

	// Create an HTTP client with Google default credential
	googleClient, err := google.DefaultClient(ctx, "https://www.googleapis.com/auth/userinfo.email")
	if err != nil {
		panic(err)
	}

	cfg := client.NewConfiguration()
	cfg.BasePath = basePath
	cfg.HTTPClient = googleClient

	apiClient := client.NewAPIClient(cfg)

	// Get all projects
	projects, _, err := apiClient.ProjectApi.ProjectsGet(ctx, nil)
	if err != nil {
		panic(err)
	}

	for _, project := range projects {
		fmt.Println()
		fmt.Println("---")
		fmt.Println()

		fmt.Println("Project:", project.Name)

		// Get all models in the given project
		models, _, err := apiClient.ModelsApi.ProjectsProjectIdModelsGet(ctx, project.Id, nil)
		if err != nil {
			panic(err)
		}

		for _, model := range models {
			fmt.Println()
			fmt.Println("- Model:", model.Name)

			// Get all model's endpoints
			modelEndpoints, _, err := apiClient.ModelEndpointsApi.ModelsModelIdEndpointsGet(ctx, model.Id)
			if err != nil {
				panic(err)
			}

			if len(modelEndpoints) == 0 {
				fmt.Println("  Model endpoints: not available")
				continue
			}

			fmt.Println("  Model endpoints:")
			for _, modelEndpoint := range modelEndpoints {
				fmt.Printf("  - %s: %s\n", modelEndpoint.EnvironmentName, modelEndpoint.Url)
			}
		}
	}
}
