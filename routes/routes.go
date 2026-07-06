package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"crm-backend/config"
	"crm-backend/controllers"
	"crm-backend/middlewares"
	"crm-backend/migrations"
	"crm-backend/models"
	"crm-backend/repositories"
	"crm-backend/services"
)

func Setup(db *gorm.DB, cfg *config.Config) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(middlewares.Recovery(), middlewares.Logger(), middlewares.CORS())
	r.Static("/uploads", cfg.App.UploadDir)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo := repositories.NewUserRepository(db)
	permRepo := repositories.NewPermissionRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	levelRepo := repositories.NewLevelRepository(db)
	contactSourceRepo := repositories.NewContactSourceRepository(db)
	clientRepo := repositories.NewClientRepository(db)
	interestingRepo := repositories.NewInterestingClientRepository(db)
	bankTypeRepo := repositories.NewBankTypeRepository(db)
	productTypeRepo := repositories.NewProductTypeRepository(db)
	bonusOptionRepo := repositories.NewBonusOptionTypeRepository(db)
	currencyTypeRepo := repositories.NewCurrencyTypeRepository(db)
	branchRepo := repositories.NewBranchRepository(db)
	depositRepo := repositories.NewDepositRepository(db)
	withdrawalRepo := repositories.NewWithdrawalRepository(db)
	turnoverBetRepo := repositories.NewTurnoverBetRepository(db)
	followUpRepo := repositories.NewFollowUpRepository(db)

	// ── Seed permissions + system roles ───────────────────────────────────────
	// Always seed permissions — idempotent, adds any new ones added in code
	_ = permRepo.Seed(models.AllPermissions)
	allPerms, _ := permRepo.FindAll()
	// Always sync system roles — updates permissions to match current definitions
	_ = roleRepo.SeedSystemRoles(allPerms)
	migrations.SeedAll(db)

	// ── Services ──────────────────────────────────────────────────────────────
	authSvc := services.NewAuthService(userRepo, roleRepo)
	permSvc := services.NewPermissionService(permRepo)
	roleSvc := services.NewRoleService(roleRepo, permRepo, userRepo)
	userSvc := services.NewUserService(userRepo, roleRepo)
	levelSvc := services.NewLevelService(levelRepo)
	contactSourceSvc := services.NewContactSourceService(contactSourceRepo)
	bankTypeSvc := services.NewBankTypeService(bankTypeRepo)
	productTypeSvc := services.NewProductTypeService(productTypeRepo)
	bonusOptionSvc := services.NewBonusOptionTypeService(bonusOptionRepo)
	currencyTypeSvc := services.NewCurrencyTypeService(currencyTypeRepo)
	clientSvc := services.NewClientService(clientRepo, db)
	interestingSvc := services.NewInterestingClientService(interestingRepo, db)
	depositSvc := services.NewDepositService(depositRepo, clientRepo, db)
	withdrawalSvc := services.NewWithdrawalService(withdrawalRepo, db)
	turnoverBetSvc := services.NewTurnoverBetService(turnoverBetRepo, db)
	followUpSvc := services.NewFollowUpService(followUpRepo, db)

	// ── Controllers ───────────────────────────────────────────────────────────
	authCtrl := controllers.NewAuthController(authSvc)
	permCtrl := controllers.NewPermissionController(permSvc, userRepo)
	roleCtrl := controllers.NewRoleController(roleSvc, userSvc)
	userCtrl := controllers.NewUserController(userSvc)
	levelCtrl := controllers.NewLevelController(levelSvc, userSvc)
	contactSourceCtrl := controllers.NewContactSourceController(contactSourceSvc, userSvc)
	clientCtrl := controllers.NewClientController(clientSvc, userSvc, userRepo)
	interestingCtrl := controllers.NewInterestingClientController(interestingSvc, userSvc, clientRepo, userRepo, db)
	bankTypeCtrl := controllers.NewBankTypeController(bankTypeSvc, userSvc)
	productTypeCtrl := controllers.NewProductTypeController(productTypeSvc, userSvc)
	bonusOptionCtrl := controllers.NewBonusOptionTypeController(bonusOptionSvc, userSvc)
	currencyTypeCtrl := controllers.NewCurrencyTypeController(currencyTypeSvc, userSvc)
	branchSvc := services.NewBranchService(branchRepo)
	branchCtrl := controllers.NewBranchController(branchSvc)
	depositCtrl := controllers.NewDepositController(db, depositSvc, userSvc)
	turnoverBetCtrl := controllers.NewTurnoverBetController(db, turnoverBetSvc, userSvc)
	followUpCtrl := controllers.NewFollowUpController(db, followUpSvc, userSvc)
	reportCtrl := controllers.NewReportController(clientRepo, interestingRepo, depositRepo, withdrawalRepo, userRepo)
	withdrawalCtrl := controllers.NewWithdrawalController(db, withdrawalSvc, userSvc)

	// ── API v1 ────────────────────────────────────────────────────────────────
	// Inject db into context for controllers that need raw queries
	r.Use(func(c *gin.Context) { c.Set("db", db); c.Next() })
	v1 := r.Group("/api/v1")

	RegisterAuthRoutes(v1, authCtrl)
	RegisterRoleRoutes(v1, roleCtrl, permCtrl, userRepo)
	RegisterUserRoutes(v1, userCtrl, userRepo)
	RegisterClientRoutes(v1, clientCtrl, userRepo)
	RegisterInterestingClientRoutes(v1, interestingCtrl, userRepo)
	RegisterLevelRoutes(v1, levelCtrl, userRepo)
	RegisterContactSourceRoutes(v1, contactSourceCtrl, userRepo)
	RegisterLookupRoutes(v1, bankTypeCtrl, productTypeCtrl, bonusOptionCtrl, currencyTypeCtrl, userRepo)
	RegisterTransactionRoutes(v1, depositCtrl, withdrawalCtrl, userRepo)
	RegisterTurnoverBetRoutes(v1, turnoverBetCtrl, userRepo)
	RegisterFollowUpRoutes(v1, followUpCtrl, userRepo)
	RegisterReportRoutes(v1, reportCtrl, userRepo)
	RegisterBranchRoutes(v1, branchCtrl)

	return r
}
