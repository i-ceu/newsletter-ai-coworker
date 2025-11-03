package routes

import (
	"blogpost-ai-coworker/internal/config"
	"blogpost-ai-coworker/internal/controllers"
	"blogpost-ai-coworker/internal/services"
	"log"
	"net/http"
)

func RegisterRoutes() {
	cfg := config.Load()

	blogpostSvc := services.NewBlogPostService(cfg.GroqAPIKey)
	infographicSvc := services.NewInfographicService()
	agentSvc, err := services.NewAgentService(cfg.GroqAPIKey, blogpostSvc, infographicSvc)
	if err != nil {
		log.Fatalf("Failed to create agent service: %v", err)
	}

	controller := controllers.NewAgentController(agentSvc)

	http.Handle("/agent", controller)

	log.Printf("BlogPost Agent Server running on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
