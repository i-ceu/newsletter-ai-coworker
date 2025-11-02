package routes

import (
	"log"
	"net/http"
	"newsletter-ai-coworker/internal/config"
	"newsletter-ai-coworker/internal/controllers"
	"newsletter-ai-coworker/internal/services"
)

func RegisterRoutes() {
	cfg := config.Load()

	newsletterSvc := services.NewNewsletterService(cfg.GroqAPIKey)
	infographicSvc := services.NewInfographicService()
	agentSvc, err := services.NewAgentService(cfg.GroqAPIKey, newsletterSvc, infographicSvc)
	if err != nil {
		log.Fatalf("Failed to create agent service: %v", err)
	}

	controller := controllers.NewAgentController(agentSvc)

	http.Handle("/agent", controller)

	log.Printf("Newsletter Agent Server running on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
