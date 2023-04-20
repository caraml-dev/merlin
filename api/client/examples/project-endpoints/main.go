package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/caraml-dev/merlin/client"
	"github.com/caraml-dev/mlp/api/pkg/auth"
)

func main() {
	ctx := context.Background()

	basePath := "http://merlin.dev/api/merlin/v1"
	if os.Getenv("MERLIN_API_BASEPATH") != "" {
		basePath = os.Getenv("MERLIN_API_BASEPATH")
	}

	// Create an HTTP client with Google default credential
	httpClient := http.DefaultClient
	googleClient, err := auth.InitGoogleClient(context.Background())
	if err == nil {
		httpClient = googleClient
	} else {
		log.Println("Google default credential not found. Fallback to HTTP default client")
	}

	cfg := client.NewConfiguration()
	cfg.BasePath = basePath
	cfg.HTTPClient = httpClient

	apiClient := client.NewAPIClient(cfg)

	// Get all projects
	projects, _, err := apiClient.ProjectApi.ProjectsGet(ctx, nil)
	if err != nil {
		panic(err)
	}
	log.Println("Projects:", projects)

	for _, project := range projects {
		log.Println()
		log.Println("---")
		log.Println()

		log.Println("Project:", project.Name)

		// Get all model endpoints in the given project
		modelEndpoints, _, err := apiClient.ModelEndpointsApi.ProjectsProjectIdModelEndpointsGet(ctx, project.Id, nil)
		if err != nil {
			panic(err)
		}
		log.Println("Model Endpoints:", modelEndpoints)

		if len(modelEndpoints) == 0 {
			log.Println("  Model endpoints: not available")
			continue
		}

		log.Println("Model endpoints:")
		for _, modelEndpoint := range modelEndpoints {
			log.Printf("- %s in %s: %s\n", modelEndpoint.Model.Name, modelEndpoint.EnvironmentName, modelEndpoint.Url)
		}
	}
}
