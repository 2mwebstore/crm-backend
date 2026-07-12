# CRM Backend API

A professional **Go + Gin + GORM + MySQL** CRM REST API with Clean Architecture, JWT authentication, role-based access control, file upload, and Swagger documentation.

---

## Project Structure

```
crm-backend/
├── cmd/server/                  ← Entrypoint (main.go)
├── config/
│   ├── app.go                   ← Config singleton, env loading
│   ├── database.go              ← GORM connect + AutoMigrate
│   └── jwt.go                   ← JWT config loader
├── routes/
│   ├── routes.go                ← DI wiring + router setup
│   ├── auth_routes.go
│   ├── client_routes.go
│   ├── interesting_client_routes.go
│   ├── level_routes.go
│   └── contact_source_routes.go
├── controllers/                 ← HTTP layer (parse request → call service → respond)
├── services/                    ← Business logic
├── repositories/                ← Data access (GORM queries)
├── models/                      ← GORM models
│   ├── user.go
│   ├── level.go
│   ├── contact_source.go
│   ├── client.go
│   ├── client_phone.go
│   ├── interesting_client.go
│   └── interesting_client_phone.go
├── dto/
│   ├── auth/                    ← Auth request structs
│   ├── client/                  ← Client request/filter structs
│   ├── interesting_client/      ← Interesting client request/filter structs
│   └── common/                  ← Shared structs
├── middlewares/
│   ├── auth_middleware.go       ← JWT validation + role injection
│   ├── cors_middleware.go
│   ├── logger_middleware.go
│   └── recovery_middleware.go
├── utils/
│   ├── response.go              ← Standardised JSON envelope
│   ├── pagination.go            ← Page/PageSize parsing + GORM scope
│   ├── jwt.go                   ← Sign + parse JWT
│   ├── password.go              ← bcrypt helpers
│   ├── validator.go             ← Binding helpers
│   ├── helper.go                ← Param parsing, sort helpers, ptr utils
│   ├── uploader.go              ← Local file save + MIME check
│   └── code_generator.go        ← CLT-YYYYMMDD-XXXX code generation
├── migrations/
│   └── init.sql                 ← Seed admin user + lookup data
├── uploads/                     ← Runtime file storage
├── docs/                        ← Swagger spec
├── docker-compose.yml           ← MySQL 8 + phpMyAdmin
├── Makefile
└── .env.example
```

---

## Quick Start

### 1. Prerequisites
- Go 1.22+
- MySQL 8+ (or Docker)

### 2. Setup
```bash
git clone <repo> && cd crm-backend
make env          # creates .env from .env.example
# Edit .env with your DB credentials and JWT_SECRET
```

### 3. Start MySQL with Docker
```bash
make docker-up
# MySQL  → localhost:3306
# phpMyAdmin → http://localhost:8081  (root / secret)
```

### 4. Run
```bash
make run
# API    → http://localhost:8080/api/v1
# Swagger → http://localhost:8080/swagger/index.html
```

### 5. Build binary
```bash
make build   # → ./bin/crm-backend
```

---

## API Reference

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication
```
Authorization: Bearer <jwt_token>
```

---

### Auth
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/auth/register` | — | Register user |
| POST | `/auth/login` | — | Login → JWT |
| GET | `/auth/me` | ✅ | Current user profile |
| PUT | `/auth/profile` | ✅ | Update profile |
| POST | `/auth/change-password` | ✅ | Change password |

### Levels  *(client classification tiers)*
| Method | Endpoint | Roles | Description |
|--------|----------|-------|-------------|
| GET | `/levels` | All | List active levels |
| POST | `/levels` | admin/manager | Create |
| PUT | `/levels/:id` | admin/manager | Update |
| DELETE | `/levels/:id` | admin | Delete |

### Contact Sources  *(how clients were acquired)*
| Method | Endpoint | Roles | Description |
|--------|----------|-------|-------------|
| GET | `/contact-sources` | All | List active sources |
| POST | `/contact-sources` | admin/manager | Create |
| PUT | `/contact-sources/:id` | admin/manager | Update |
| DELETE | `/contact-sources/:id` | admin | Delete |

### Clients
| Method | Endpoint | Roles | Description |
|--------|----------|-------|-------------|
| GET | `/clients` | All | List (search, filter, paginate) |
| POST | `/clients` | All | Create |
| GET | `/clients/:id` | All | Get with relations |
| PUT | `/clients/:id` | All | Update (partial) |
| DELETE | `/clients/:id` | admin/manager | Delete |
| POST | `/clients/:id/logo` | All | Upload logo image |

#### Client Filter Query Params
| Param | Type | Example |
|-------|------|---------|
| `search` | string | `?search=acme` |
| `status` | string | `?status=active` |
| `type` | string | `?type=company` |
| `country` | string | `?country=Cambodia` |
| `industry` | string | `?industry=tech` |
| `level_id` | int | `?level_id=3` |
| `contact_source_id` | int | `?contact_source_id=1` |
| `assigned_to_id` | int | `?assigned_to_id=5` |
| `page` | int | `?page=2` |
| `page_size` | int | `?page_size=20` |
| `sort_by` | string | `?sort_by=annual_revenue` |
| `sort_dir` | string | `?sort_dir=desc` |

### Interesting Clients
| Method | Endpoint | Roles | Description |
|--------|----------|-------|-------------|
| GET | `/interesting-clients` | All | List (search, filter, paginate) |
| POST | `/interesting-clients` | All | Create |
| GET | `/interesting-clients/:id` | All | Get with relations |
| PUT | `/interesting-clients/:id` | All | Update (partial) |
| DELETE | `/interesting-clients/:id` | admin/manager | Delete |
| POST | `/interesting-clients/:id/convert` | All | Convert to real Client |

#### Convert endpoint behaviour
- Without body → creates a new `Client` from the Interesting Client's data, migrating phones
- With `{ "existing_client_id": 5 }` → links to an existing Client instead

---

## Data Models

### Client
```json
{
  "id": 1,
  "code": "CLT-20250627-A3X9",
  "name": "Acme Corp",
  "type": "company",
  "status": "active",
  "industry": "Technology",
  "email": "contact@acme.com",
  "website": "https://acme.com",
  "annual_revenue": 500000,
  "currency": "USD",
  "address": "123 Main St",
  "city": "Phnom Penh",
  "country": "Cambodia",
  "level": { "id": 3, "name": "Gold", "color": "#ffd700" },
  "contact_source": { "id": 1, "name": "Referral" },
  "assigned_to": { "id": 2, "name": "John Sales" },
  "phones": [
    { "phone": "+855123456789", "label": "primary", "is_primary": true }
  ]
}
```

### Interesting Client
```json
{
  "id": 1,
  "code": "INT-20250627-B7K2",
  "full_name": "Dara Sok",
  "company_name": "Sok Industries",
  "email": "dara@sok.com",
  "priority": "high",
  "interest_score": 85,
  "opportunity_value": 250000,
  "currency": "USD",
  "reason": "Met at TechSummit, strong budget signal",
  "next_follow_up_at": "2025-07-15T09:00:00Z",
  "is_converted": false,
  "phones": [
    { "phone": "+85512345678", "label": "mobile", "is_primary": true }
  ]
}
```

---

## Roles & Permissions

| Role | Can Do |
|------|--------|
| `admin` | Everything including deleting levels/sources |
| `manager` | Full CRUD on clients/interesting, manage levels/sources (no delete) |
| `sales` | Read + create + update clients/interesting. Cannot delete. |

---

## Default Seed Data

After `make docker-up` + `make run`:

**Admin login:**
```
Email:    admin@crm.local
Password: password   (bcrypt hash in init.sql)
```

**Levels seeded:** Bronze, Silver, Gold, Platinum

**Contact Sources seeded:** Referral, Cold Call, Website, Social Media, Event, Email Campaign, Walk-in, Partner

---

## Swagger UI

```
http://localhost:8080/swagger/index.html
```

To regenerate after modifying annotations:
```bash
make install-swag
make swag
```

---

## Environment Variables

| Variable | Default | Required |
|----------|---------|----------|
| `APP_ENV` | `development` | |
| `APP_PORT` | `8080` | |
| `BASE_URL` | `http://localhost:8080` | |
| `DB_HOST` | `127.0.0.1` | |
| `DB_PORT` | `3306` | |
| `DB_USER` | `root` | |
| `DB_PASSWORD` | *(empty)* | |
| `DB_NAME` | `crm_db` | |
| `JWT_SECRET` | — | ✅ |
| `JWT_EXPIRE_HOURS` | `24` | |
| `UPLOAD_DIR` | `./uploads` | |
| `MAX_UPLOAD_SIZE_MB` | `10` | |

---

## Production Deployment (VPS / DigitalOcean / Contabo)

```bash
# Build
make build   # → ./bin/crm-backend

# Systemd service
sudo nano /etc/systemd/system/crm.service
```
```ini
[Unit]
Description=CRM Backend API
After=network.target

[Service]
ExecStart=/var/www/crm/bin/crm-backend
WorkingDirectory=/var/www/crm
EnvironmentFile=/var/www/crm/.env
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```
```bash
sudo systemctl enable crm && sudo systemctl start crm

# Nginx reverse proxy
# location /api/ { proxy_pass http://127.0.0.1:8080/api/; }
# location /uploads/ { alias /var/www/crm/uploads/; }
# location /swagger/ { proxy_pass http://127.0.0.1:8080/swagger/; }
```

---

## Switching Uploads to Cloudflare R2 / S3

Only `utils/uploader.go` needs to change — replace the `SaveFile` function with an SDK call and return the CDN URL as `FileURL`. All callers remain untouched.



cd /var/www/crm-backend
cp .env.example .env
nano .env