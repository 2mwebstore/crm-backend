package migrations

import (
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"crm-backend/models"
)

// SeedAll runs all seeders in dependency order. Safe to call on every startup —
// each function checks for existing data before inserting.
func SeedAll(db *gorm.DB) {
	SeedSuperAdmin(db)
	SeedLevels(db)
	SeedContactSources(db)
	SeedBankTypes(db)
	SeedProductTypes(db)
	SeedBonusOptions(db)
	SeedCurrencies(db)
	SeedCompanyBanks(db) // must run after SeedBankTypes + SeedCurrencies
}

// ── Super Admin ───────────────────────────────────────────────────────────────

func SeedSuperAdmin(db *gorm.DB) {
	const email = "admin@crm.local"
	var existing models.User
	if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
		return // already exists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("Admin@1234"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("seeder: bcrypt error: %v", err)
	}
	u := models.User{
		Name:         "Super Admin",
		Email:        email,
		PasswordHash: string(hash),
		IsActive:     true,
		IsSuperAdmin: true,
	}
	if err := db.Create(&u).Error; err != nil {
		log.Fatalf("seeder: create super admin: %v", err)
	}
	log.Printf("✅ Super Admin  →  %s  /  Admin@1234", email)
}

// ── Levels ────────────────────────────────────────────────────────────────────

func SeedLevels(db *gorm.DB) {
	var count int64
	db.Model(&models.Level{}).Count(&count)
	if count > 0 {
		return
	}
	levels := []models.Level{
		{Name: "blacklist", Color: "#ef4444", SortOrder: 1, IsActive: true},
		{Name: "suspend", Color: "#f97316", SortOrder: 2, IsActive: true},
		{Name: "warning", Color: "#eab308", SortOrder: 3, IsActive: true},
		{Name: "tracking", Color: "#3b82f6", SortOrder: 4, IsActive: true},
		{Name: "vip1", Color: "#8b5cf6", SortOrder: 5, IsActive: true},
		{Name: "vip2", Color: "#a855f7", SortOrder: 6, IsActive: true},
		{Name: "vip3", Color: "#938af4", SortOrder: 7, IsActive: true},
	}
	for i := range levels {
		levels[i].CreatedByID = 1
	}
	if err := db.Create(&levels).Error; err != nil {
		log.Printf("seeder: levels: %v", err)
		return
	}
	log.Println("✅ Levels seeded")
}

// ── Contact Sources ───────────────────────────────────────────────────────────

func SeedContactSources(db *gorm.DB) {
	var count int64
	db.Model(&models.ContactSource{}).Count(&count)
	if count > 0 {
		return
	}
	sources := []models.ContactSource{
		{Name: "tiktok", Icon: "tiktok", IsActive: true},
		{Name: "facebook", Icon: "facebook", IsActive: true},
		{Name: "telegram", Icon: "telegram", IsActive: true},
	}
	for i := range sources {
		sources[i].CreatedByID = 1
	}
	if err := db.Create(&sources).Error; err != nil {
		log.Printf("seeder: contact sources: %v", err)
		return
	}
	log.Println("✅ Contact Sources seeded")
}

// ── Bank Types ────────────────────────────────────────────────────────────────

func SeedBankTypes(db *gorm.DB) {
	var count int64
	db.Model(&models.BankType{}).Count(&count)
	if count > 0 {
		return
	}
	banks := []models.BankType{
		{Name: "ABA Bank", Code: "ABA", IsActive: true, SortOrder: 1},
		{Name: "ACLEDA Bank", Code: "ACLEDA", IsActive: true, SortOrder: 2},
		{Name: "Wing Bank", Code: "WING", IsActive: true, SortOrder: 3},
		{Name: "TRUE Money", Code: "TRUE", IsActive: true, SortOrder: 4},
	}
	for i := range banks {
		banks[i].CreatedByID = 1
	}
	if err := db.Create(&banks).Error; err != nil {
		log.Printf("seeder: bank types: %v", err)
		return
	}
	log.Println("✅ Bank Types seeded")
}

// ── Product Types ─────────────────────────────────────────────────────────────

func SeedProductTypes(db *gorm.DB) {
	var count int64
	db.Model(&models.ProductType{}).Count(&count)
	if count > 0 {
		return
	}
	products := []models.ProductType{
		{Name: "Motor", Code: "MOTOR", IsActive: true, SortOrder: 1},
		{Name: "Car", Code: "CAR", IsActive: true, SortOrder: 2},
		{Name: "Home", Code: "HOME", IsActive: true, SortOrder: 3},
	}
	for i := range products {
		products[i].CreatedByID = 1
	}
	if err := db.Create(&products).Error; err != nil {
		log.Printf("seeder: product types: %v", err)
		return
	}
	log.Println("✅ Product Types seeded")
}

// ── Bonus Options ─────────────────────────────────────────────────────────────

func SeedBonusOptions(db *gorm.DB) {
	var count int64
	db.Model(&models.BonusOptionType{}).Count(&count)
	if count > 0 {
		return
	}
	bonuses := []models.BonusOptionType{
		{Name: "Voucher", IsActive: true, SortOrder: 1},
		{Name: "Weekly", IsActive: true, SortOrder: 2},
		{Name: "Monthly", IsActive: true, SortOrder: 3},
		{Name: "10%", IsActive: true, SortOrder: 4},
	}
	for i := range bonuses {
		bonuses[i].CreatedByID = 1
	}
	if err := db.Create(&bonuses).Error; err != nil {
		log.Printf("seeder: bonus options: %v", err)
		return
	}
	log.Println("✅ Bonus Options seeded")
}

// ── Currencies ────────────────────────────────────────────────────────────────

func SeedCurrencies(db *gorm.DB) {
	var count int64
	db.Model(&models.CurrencyType{}).Count(&count)
	if count > 0 {
		return
	}
	currencies := []models.CurrencyType{
		{Code: "USD", Name: "US Dollar", Symbol: "$", IsBase: true, IsActive: true, SortOrder: 1},
		{Code: "KHR", Name: "Cambodian Riel", Symbol: "៛", IsBase: false, IsActive: true, SortOrder: 2},
	}
	for i := range currencies {
		currencies[i].CreatedByID = 1
	}
	if err := db.Create(&currencies).Error; err != nil {
		log.Printf("seeder: currencies: %v", err)
		return
	}
	log.Println("✅ Currencies seeded")
}

// ── Company Banks ─────────────────────────────────────────────────────────────

// SeedCompanyBanks inserts a couple of sample company bank accounts.
// Depends on SeedBankTypes and SeedCurrencies having already run — it looks
// up the bank/currency rows by their seeded `code` rather than hardcoding
// IDs, since IDs can shift depending on seed order/history.
func SeedCompanyBanks(db *gorm.DB) {
	var count int64
	db.Model(&models.CompanyBank{}).Count(&count)
	if count > 0 {
		return
	}

	var aba models.BankType
	if err := db.Where("code = ?", "ABA").First(&aba).Error; err != nil {
		log.Printf("seeder: company banks skipped — ABA bank type not found (run SeedBankTypes first)")
		return
	}
	var acleda models.BankType
	hasAcleda := db.Where("code = ?", "ACLEDA").First(&acleda).Error == nil

	var usd, khr models.CurrencyType
	hasUSD := db.Where("code = ?", "USD").First(&usd).Error == nil
	hasKHR := db.Where("code = ?", "KHR").First(&khr).Error == nil

	var usdID, khrID *uint
	if hasUSD {
		usdID = &usd.ID
	}
	if hasKHR {
		khrID = &khr.ID
	}

	banks := []models.CompanyBank{
		{
			BankTypeID:     aba.ID,
			AccountNumber:  "000 000 001",
			AccountName:    "COMPANY NAME LTD",
			CurrencyTypeID: usdID,
			IsActive:       true,
			SortOrder:      1,
		},
		{
			BankTypeID:     aba.ID,
			AccountNumber:  "000 000 002",
			AccountName:    "COMPANY NAME LTD",
			CurrencyTypeID: khrID,
			IsActive:       true,
			SortOrder:      2,
		},
	}
	if hasAcleda {
		banks = append(banks, models.CompanyBank{
			BankTypeID:     acleda.ID,
			AccountNumber:  "000 000 003",
			AccountName:    "COMPANY NAME LTD",
			CurrencyTypeID: usdID,
			IsActive:       true,
			SortOrder:      3,
		})
	}

	for i := range banks {
		banks[i].CreatedByID = 1
	}
	if err := db.Create(&banks).Error; err != nil {
		log.Printf("seeder: company banks: %v", err)
		return
	}
	log.Println("✅ Company Banks seeded")
}
