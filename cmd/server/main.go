// @title           CRM Backend API
// @version         1.0
// @description     Professional CRM API — Client & Interesting Client Management
// @termsOfService  http://swagger.io/terms/

// @contact.name  API Support
// @contact.email support@yourcrm.com

// @license.name Apache 2.0
// @license.url  http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT token.

package main

import (
	"log"
	"os"

	_ "crm-backend/docs" // swagger generated docs

	"crm-backend/config"
	"crm-backend/routes"
)

func main() {
	// Load config (.env → singleton)
	cfg := config.Load()

	// Ensure upload directory exists
	if err := os.MkdirAll(cfg.App.UploadDir, 0755); err != nil {
		log.Fatalf("failed to create upload dir: %v", err)
	}

	// Connect DB + AutoMigrate
	db := config.ConnectDB(cfg)

	// Build router
	r := routes.Setup(db, cfg)

	addr := ":" + cfg.App.Port
	log.Printf("🚀  CRM API  →  http://localhost%s", addr)
	log.Printf("📚  Swagger  →  http://localhost%s/swagger/index.html", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
