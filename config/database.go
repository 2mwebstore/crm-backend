package config

import (
	"fmt"
	"log"
	"time"

	"crm-backend/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type DBConfig struct {
	Host, Port, User, Password, Name, Charset string
}

func loadDB() DBConfig {
	return DBConfig{
		Host:     getEnv("DB_HOST", "127.0.0.1"),
		Port:     getEnv("DB_PORT", "3306"),
		User:     getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", ""),
		Name:     getEnv("DB_NAME", "crm_db"),
		Charset:  getEnv("DB_CHARSET", "utf8mb4"),
	}
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Asia%%2FPhnom_Penh",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Name,
		d.Charset,
	)
}

func ConnectDB(cfg *Config) *gorm.DB {
	logLevel := logger.Silent
	if cfg.App.Env == "development" {
		logLevel = logger.Info
	}

	cambodiaTZ, err := time.LoadLocation("Asia/Phnom_Penh")
	if err != nil {
		log.Fatalf("failed to load timezone: %v", err)
	}

	db, err := gorm.Open(mysql.Open(cfg.DB.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
		NowFunc: func() time.Time {
			return time.Now().In(cambodiaTZ)
		},
	})
	if err != nil {
		log.Fatalf("database: failed to connect — %v", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		_, _ = sqlDB.Exec("SET time_zone = '+07:00'")
	}

	if getEnv("DB_AUTO_MIGRATE", "true") == "true" {
		err = db.AutoMigrate(
			&models.Branch{},
			&models.CodeSequence{},
			&models.Permission{},
			&models.Role{},
			&models.User{},
			&models.Level{},
			&models.ContactSource{},
			&models.BankType{},
			&models.ProductType{},
			&models.BonusOptionType{},
			&models.CurrencyType{},
			&models.Client{},
			&models.ClientPhone{},
			&models.ClientBank{},
			&models.ClientProduct{},
			&models.ClientFollowUp{},
			&models.InterestingClient{},
			&models.InterestingClientPhone{},
			&models.Deposit{},
			&models.Withdrawal{},
			&models.TurnoverBet{},
			&models.CompanyBank{},
			&models.BalanceTransaction{},
			&models.DailyStartBalance{},
			&models.DailyStartBalanceDetail{},
			&models.AuditLog{},
		)
		if err != nil {
			log.Fatalf("database: migration failed — %v", err)
		}
		log.Println("✅ Database connected and migrated (Asia/Phnom_Penh)")
	} else {
		log.Println("✅ Database connected (Asia/Phnom_Penh) — AutoMigrate skipped (DB_AUTO_MIGRATE=false)")
	}

	return db
}
