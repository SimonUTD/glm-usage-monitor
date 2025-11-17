package main

import (
	"context"
	"fmt"
	"log"
	"glm-usage-monitor/services"
)

// App struct
type App struct {
	ctx        context.Context
	database   *Database
	apiService *services.APIService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize database
	db, err := NewDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	a.database = db
	a.apiService = services.NewAPIService(db)
	log.Println("Database and API service initialized successfully")
}

// shutdown is called when the app is about to close
func (a *App) shutdown(ctx context.Context) {
	if a.database != nil {
		if err := a.database.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println("Database closed successfully")
		}
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetDatabase returns the database instance
func (a *App) GetDatabase() *Database {
	return a.database
}

// GetAPIService returns the API service instance
func (a *App) GetAPIService() *services.APIService {
	return a.apiService
}
