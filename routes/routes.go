package routes

import (
	"log"

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
	middlewares.InitAuth(db)
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(middlewares.Recovery(), middlewares.Logger(), middlewares.CORS(), middlewares.AuditLog())
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
	companyBankRepo := repositories.NewCompanyBankRepository(db)
	balanceTxRepo := repositories.NewBalanceTransactionRepository(db)
	dailyStartBalanceRepo := repositories.NewDailyStartBalanceRepository(db)
	auditLogRepo := repositories.NewAuditLogRepository(db)

	// ── Seed permissions + system roles + sample data ─────────────────────────
	// All three gated together behind DB_SEED — turn off once a production
	// DB already has its own real permissions, roles, and data configured.
	// Defaults to "true" so nothing changes unless DB_SEED is explicitly set.
	if config.GetEnv("DB_SEED", "true") == "true" {
		_ = permRepo.Seed(models.AllPermissions)
		allPerms, _ := permRepo.FindAll()
		_ = roleRepo.SeedSystemRoles(allPerms)
		migrations.SeedAll(db)
	} else {
		log.Println("⏭️  Seeder skipped (DB_SEED=false) — permissions, system roles, and sample data left untouched")
	}

	// ── Services ──────────────────────────────────────────────────────────────
	authSvc := services.NewAuthService(userRepo, roleRepo)
	permSvc := services.NewPermissionService(permRepo)
	roleSvc := services.NewRoleService(roleRepo, permRepo, userRepo)
	userSvc := services.NewUserService(userRepo, roleRepo)
	levelSvc := services.NewLevelService(levelRepo)
	contactSourceSvc := services.NewContactSourceService(contactSourceRepo)
	bankTypeSvc := services.NewBankTypeService(bankTypeRepo)
	CompanyBankSvc := services.NewCompanyBankService(companyBankRepo, dailyStartBalanceRepo)
	productTypeSvc := services.NewProductTypeService(productTypeRepo, dailyStartBalanceRepo)
	bonusOptionSvc := services.NewBonusOptionTypeService(bonusOptionRepo)
	currencyTypeSvc := services.NewCurrencyTypeService(currencyTypeRepo)
	clientSvc := services.NewClientService(clientRepo, db)
	interestingSvc := services.NewInterestingClientService(interestingRepo, db)
	depositSvc := services.NewDepositService(depositRepo, clientRepo, companyBankRepo, productTypeRepo, dailyStartBalanceRepo, branchRepo, db)
	withdrawalSvc := services.NewWithdrawalService(withdrawalRepo, clientRepo, companyBankRepo, productTypeRepo, dailyStartBalanceRepo, branchRepo, db)
	turnoverBetSvc := services.NewTurnoverBetService(turnoverBetRepo, db)
	followUpSvc := services.NewFollowUpService(followUpRepo, db)
	balanceTxSvc := services.NewBalanceTransactionService(balanceTxRepo)
	dailyStartBalanceSvc := services.NewDailyStartBalanceService(dailyStartBalanceRepo, userRepo, companyBankRepo, productTypeRepo, depositRepo, withdrawalRepo, balanceTxRepo)
	auditLogSvc := services.NewAuditLogService(auditLogRepo)

	// ── Controllers ───────────────────────────────────────────────────────────
	authCtrl := controllers.NewAuthController(authSvc)
	permCtrl := controllers.NewPermissionController(permSvc, userRepo)
	roleCtrl := controllers.NewRoleController(roleSvc, userSvc)
	userCtrl := controllers.NewUserController(userSvc)
	levelCtrl := controllers.NewLevelController(levelSvc, userSvc)
	contactSourceCtrl := controllers.NewContactSourceController(contactSourceSvc, userSvc)
	clientCtrl := controllers.NewClientController(clientSvc, userSvc, userRepo, db)
	interestingCtrl := controllers.NewInterestingClientController(interestingSvc, userSvc, clientRepo, userRepo, db)
	bankTypeCtrl := controllers.NewBankTypeController(bankTypeSvc, userSvc)
	companybankCtrl := controllers.NewCompanyBankController(CompanyBankSvc, userSvc)
	productTypeCtrl := controllers.NewProductTypeController(productTypeSvc, userSvc)
	bonusOptionCtrl := controllers.NewBonusOptionTypeController(bonusOptionSvc, userSvc)
	currencyTypeCtrl := controllers.NewCurrencyTypeController(currencyTypeSvc, userSvc)
	branchSvc := services.NewBranchService(branchRepo)
	branchCtrl := controllers.NewBranchController(branchSvc)
	depositCtrl := controllers.NewDepositController(db, depositSvc, userSvc)
	turnoverBetCtrl := controllers.NewTurnoverBetController(db, turnoverBetSvc, userSvc)
	followUpCtrl := controllers.NewFollowUpController(db, followUpSvc, userSvc)
	reportCtrl := controllers.NewReportController(clientRepo, interestingRepo, depositRepo, withdrawalRepo, userRepo, companyBankRepo, bankTypeRepo)
	withdrawalCtrl := controllers.NewWithdrawalController(db, withdrawalSvc, userSvc)
	balanceTxCtrl := controllers.NewBalanceTransactionController(balanceTxSvc)
	dailyStartBalanceCtrl := controllers.NewDailyStartBalanceController(dailyStartBalanceSvc)
	auditLogCtrl := controllers.NewAuditLogController(auditLogSvc)

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
	RegisterLookupRoutes(v1, bankTypeCtrl, companybankCtrl, productTypeCtrl, bonusOptionCtrl, currencyTypeCtrl, userRepo, balanceTxCtrl)
	RegisterTransactionRoutes(v1, depositCtrl, withdrawalCtrl, userRepo)
	RegisterTurnoverBetRoutes(v1, turnoverBetCtrl, userRepo)
	RegisterFollowUpRoutes(v1, followUpCtrl, userRepo)
	RegisterReportRoutes(v1, reportCtrl, userRepo)
	RegisterBranchRoutes(v1, branchCtrl)
	RegisterDailyStartBalanceRoutes(v1, dailyStartBalanceCtrl, userRepo)
	RegisterAuditLogRoutes(v1, auditLogCtrl, userRepo)

	return r
}
