# Auth + Tests Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add JWT authentication with role-based access (admin/auditor/viewer) and comprehensive tests (unit + integration) to the CyberRisk platform.

**Architecture:** Auth uses bcrypt password hashing + JWT access tokens (15 min TTL). Middleware chain: CORS -> JWT -> Role Check -> Handler. Public routes (/api/auth/*, /health) skip JWT. Tests use testify for assertions and testcontainers-go for PostgreSQL integration tests.

**Tech Stack:** golang-jwt/jwt/v5, golang.org/x/crypto/bcrypt, stretchr/testify, testcontainers-go

---

## File Structure

### New files:
| File | Responsibility |
|------|---------------|
| `internal/repository/user_repository.go` | User CRUD (Create, GetByUsername, GetByID) |
| `internal/service/auth/service.go` | Register, Login, ValidateToken, password hashing |
| `internal/transport/http/auth_handlers.go` | POST /api/auth/register, POST /api/auth/login, GET /api/auth/me |
| `internal/transport/http/middleware.go` | JWT extraction + validation middleware, role-based access middleware |
| `internal/service/risk/calculator_test.go` | Unit tests for risk calculator |
| `internal/service/asset/cia_calculator_test.go` | Unit tests for CIA calculator |
| `internal/service/auth/service_test.go` | Unit tests for auth service |
| `internal/repository/testhelper_test.go` | Testcontainers PostgreSQL setup + migrations |
| `internal/repository/asset_repository_test.go` | Integration tests for asset repository |
| `internal/repository/threat_repository_test.go` | Integration tests for threat repository |
| `internal/repository/vulnerability_repository_test.go` | Integration tests for vulnerability repository |
| `internal/repository/software_repository_test.go` | Integration tests for software repository |
| `internal/repository/user_repository_test.go` | Integration tests for user repository |

### Modified files:
| File | Changes |
|------|---------|
| `internal/config/config.go` | Add JWTSecret field |
| `internal/transport/http/server.go` | Wire auth service, handlers, middleware into router |
| `go.mod` | Add jwt, testify, testcontainers dependencies |

---

### Task 1: Add dependencies

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: Add required Go modules**

```bash
cd /Users/velvetway/Downloads/CyberRisk
go get github.com/golang-jwt/jwt/v5@latest
go get golang.org/x/crypto@latest
go get github.com/stretchr/testify@latest
go get github.com/testcontainers/testcontainers-go@latest
go get github.com/testcontainers/testcontainers-go/modules/postgres@latest
```

- [ ] **Step 2: Verify modules installed**

Run: `go mod tidy`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add jwt, bcrypt, testify, testcontainers"
```

---

### Task 2: Add JWTSecret to config

**Files:**
- Modify: `internal/config/config.go`

- [ ] **Step 1: Update Config struct and Load function**

```go
package config

import (
	"os"
)

type Config struct {
	DBDSN     string
	HTTPPort  string
	JWTSecret string
}

func Load() (*Config, error) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://app:app@localhost:5432/cyber_risk?sslmode=disable"
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8081"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-me"
	}

	return &Config{
		DBDSN:     dsn,
		HTTPPort:  port,
		JWTSecret: jwtSecret,
	}, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/config/config.go
git commit -m "feat: add JWT_SECRET to config"
```

---

### Task 3: Create user repository

**Files:**
- Create: `internal/repository/user_repository.go`

- [ ] **Step 1: Write the user repository**

```go
package repository

import (
	"context"
	"errors"
	"fmt"

	"Diplom/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (username, password_hash, role, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(ctx, query,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, password_hash, role, is_active, created_at, updated_at
		FROM users WHERE username = $1`

	var u domain.User
	err := r.db.QueryRow(ctx, query, username).Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.Role,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return &u, nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
		SELECT id, username, password_hash, role, is_active, created_at, updated_at
		FROM users WHERE id = $1`

	var u domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.Role,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/repository/...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/repository/user_repository.go
git commit -m "feat: add user repository for auth"
```

---

### Task 4: Create auth service

**Files:**
- Create: `internal/service/auth/service.go`

- [ ] **Step 1: Write the auth service**

```go
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"Diplom/internal/domain"
	"Diplom/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists       = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive     = errors.New("user is inactive")
)

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type Service interface {
	Register(ctx context.Context, username, password string, role domain.UserRole) (*domain.User, error)
	Login(ctx context.Context, username, password string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
}

type service struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
	tokenTTL  time.Duration
}

func NewService(userRepo repository.UserRepository, jwtSecret string) Service {
	return &service{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
		tokenTTL:  15 * time.Minute,
	}
}

func (s *service) Register(ctx context.Context, username, password string, role domain.UserRole) (*domain.User, error) {
	existing, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("check existing user: %w", err)
	}
	if existing != nil {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *service) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}
	if !user.IsActive {
		return "", ErrUserInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	now := time.Now()
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return tokenString, nil
}

func (s *service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/service/auth/...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/service/auth/service.go
git commit -m "feat: add auth service with JWT + bcrypt"
```

---

### Task 5: Create JWT middleware and role checker

**Files:**
- Create: `internal/transport/http/middleware.go`

- [ ] **Step 1: Write the middleware**

```go
package http

import (
	"strings"

	authService "Diplom/internal/service/auth"

	"github.com/gofiber/fiber/v2"
)

// JWTMiddleware extracts and validates JWT from Authorization header.
// On success, stores claims in c.Locals("claims").
func JWTMiddleware(authSvc authService.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		claims, err := authSvc.ValidateToken(parts[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		c.Locals("claims", claims)
		return c.Next()
	}
}

// RequireRole returns middleware that checks if the user has one of the allowed roles.
func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("claims").(*authService.Claims)
		if !ok || claims == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		if !allowed[claims.Role] {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "insufficient permissions",
			})
		}

		return c.Next()
	}
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/transport/http/...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/transport/http/middleware.go
git commit -m "feat: add JWT middleware and role checker"
```

---

### Task 6: Create auth handlers

**Files:**
- Create: `internal/transport/http/auth_handlers.go`

- [ ] **Step 1: Write auth handlers**

```go
package http

import (
	"Diplom/internal/domain"
	authService "Diplom/internal/service/auth"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authSvc authService.Service
}

func NewAuthHandler(authSvc authService.Service) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

func (h *AuthHandler) Register(g fiber.Router) {
	g.Post("/register", h.register)
	g.Post("/login", h.login)
}

// RegisterProtected registers routes that require JWT.
func (h *AuthHandler) RegisterProtected(g fiber.Router) {
	g.Get("/me", h.me)
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

type userResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

func (h *AuthHandler) register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username and password are required"})
	}

	if len(req.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "password must be at least 6 characters"})
	}

	role := domain.UserRoleViewer
	if req.Role != "" {
		switch domain.UserRole(req.Role) {
		case domain.UserRoleAdmin, domain.UserRoleAuditor, domain.UserRoleViewer:
			role = domain.UserRole(req.Role)
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid role"})
		}
	}

	user, err := h.authSvc.Register(c.Context(), req.Username, req.Password, role)
	if err != nil {
		if err == authService.ErrUserExists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "user already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(userResponse{
		ID:       user.ID,
		Username: user.Username,
		Role:     string(user.Role),
		IsActive: user.IsActive,
	})
}

func (h *AuthHandler) login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username and password are required"})
	}

	token, err := h.authSvc.Login(c.Context(), req.Username, req.Password)
	if err != nil {
		if err == authService.ErrInvalidCredentials || err == authService.ErrUserInactive {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(tokenResponse{Token: token})
}

func (h *AuthHandler) me(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*authService.Claims)
	if !ok || claims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	return c.JSON(userResponse{
		ID:       claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
		IsActive: true,
	})
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/transport/http/...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/transport/http/auth_handlers.go
git commit -m "feat: add auth HTTP handlers (register, login, me)"
```

---

### Task 7: Wire auth into server.go

**Files:**
- Modify: `internal/transport/http/server.go`

- [ ] **Step 1: Update NewServer to accept jwtSecret and wire auth**

Replace the entire `server.go` with:

```go
package http

import (
	"context"
	"log"

	"Diplom/internal/repository"
	assetService "Diplom/internal/service/asset"
	assetVulnService "Diplom/internal/service/asset_vulnerability"
	authService "Diplom/internal/service/auth"
	riskService "Diplom/internal/service/risk"
	softwareService "Diplom/internal/service/software"
	threatService "Diplom/internal/service/threat"
	vulnerabilityService "Diplom/internal/service/vulnerability"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewServer(_ context.Context, db *pgxpool.Pool, jwtSecret string) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware: CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Middleware: Recover from panics
	app.Use(recover.New())

	// ---------- health-check ----------
	app.Get("/health", func(c *fiber.Ctx) error {
		if db != nil {
			if err := db.Ping(c.Context()); err != nil {
				log.Printf("health: db ping error: %v", err)
				return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
					"status": "degraded",
					"error":  "db not available",
				})
			}
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// ---------- DI ----------

	// Auth
	userRepo := repository.NewUserRepository(db)
	authSvc := authService.NewService(userRepo, jwtSecret)
	authHandler := NewAuthHandler(authSvc)

	// Software
	softwareRepo := repository.NewSoftwareRepository(db)
	softwareSvc := softwareService.NewService(softwareRepo)
	softwareHandler := NewSoftwareHandler(softwareSvc)

	// Assets
	assetRepo := repository.NewAssetRepository(db)
	assetSvc := assetService.NewService(assetRepo)
	assetHandler := NewAssetHandler(assetSvc, softwareSvc)

	// Threats
	threatRepo := repository.NewThreatRepository(db)
	threatSvc := threatService.NewService(threatRepo)
	threatHandler := NewThreatHandler(threatSvc)

	// Vulnerabilities
	vulnRepo := repository.NewVulnerabilityRepository(db)
	vulnSvc := vulnerabilityService.NewService(vulnRepo)
	vulnHandler := NewVulnerabilityHandler(vulnSvc)

	// Asset-Vulnerability links
	assetVulnRepo := repository.NewAssetVulnerabilityRepository(db)
	assetVulnSvc := assetVulnService.NewService(assetVulnRepo)
	assetVulnHandler := NewAssetVulnerabilityHandler(assetVulnSvc)

	// Risk
	riskSvc := riskService.NewService(assetRepo, threatRepo, vulnRepo, assetVulnRepo)
	riskHandler := NewRiskHandler(riskSvc)

	// ---------- Public routes (no auth) ----------
	api := app.Group("/api")

	authGroup := api.Group("/auth")
	authHandler.Register(authGroup) // POST /api/auth/register, POST /api/auth/login

	// ---------- Protected routes (JWT required) ----------
	protected := api.Group("", JWTMiddleware(authSvc))

	// GET /api/auth/me
	authHandler.RegisterProtected(protected.Group("/auth"))

	// Read-only routes: viewer, auditor, admin
	readOnly := protected.Group("", RequireRole("viewer", "auditor", "admin"))

	// Write routes: auditor, admin
	write := protected.Group("", RequireRole("auditor", "admin"))

	// Delete routes: admin only
	adminOnly := protected.Group("", RequireRole("admin"))

	// Assets (actual method names from asset_handlers.go)
	assetsRead := readOnly.Group("/assets")
	assetsRead.Get("/", assetHandler.listAssets)
	assetsRead.Get("/:id", assetHandler.getAsset)
	assetsRead.Get("/:assetID/vulnerabilities", assetVulnHandler.listForAsset)
	assetsRead.Get("/:id/software/alternatives", assetHandler.assetSoftwareAlternatives)

	assetsWrite := write.Group("/assets")
	assetsWrite.Post("/", assetHandler.createAsset)
	assetsWrite.Put("/:id", assetHandler.updateAsset)
	assetsWrite.Post("/:assetID/vulnerabilities", assetVulnHandler.addToAsset)

	assetsAdmin := adminOnly.Group("/assets")
	assetsAdmin.Delete("/:id", assetHandler.deleteAsset)
	assetsAdmin.Delete("/:assetID/vulnerabilities/:vulnID", assetVulnHandler.removeFromAsset)

	// Threats (actual method names from threat_handlers.go)
	readOnly.Get("/threats", threatHandler.listThreats)
	readOnly.Get("/threats/:id", threatHandler.getThreat)
	write.Post("/threats", threatHandler.createThreat)
	write.Put("/threats/:id", threatHandler.updateThreat)
	adminOnly.Delete("/threats/:id", threatHandler.deleteThreat)

	// Vulnerabilities (actual method names from vulnerability_handlers.go)
	readOnly.Get("/vulnerabilities", vulnHandler.list)
	readOnly.Get("/vulnerabilities/:id", vulnHandler.get)
	write.Post("/vulnerabilities", vulnHandler.create)
	write.Put("/vulnerabilities/:id", vulnHandler.update)
	adminOnly.Delete("/vulnerabilities/:id", vulnHandler.delete)

	// Risk (actual method names from risk_handlers.go)
	readOnly.Get("/risk/overview", riskHandler.overview)
	readOnly.Get("/risk/asset/:id", riskHandler.assetRiskProfile)
	write.Post("/risk/preview", riskHandler.previewRisk)
	write.Post("/risk/report/pdf", riskHandler.GenerateRiskPDF)

	// Software (actual method names from software_handlers.go)
	readOnly.Get("/software", softwareHandler.listSoftware)
	readOnly.Get("/software/categories", softwareHandler.listCategories)
	readOnly.Get("/software/russian", softwareHandler.listRussianSoftware)
	readOnly.Get("/software/certified", softwareHandler.listCertifiedSoftware)
	readOnly.Get("/software/:id", softwareHandler.getSoftware)
	write.Post("/software", softwareHandler.createSoftware)
	write.Put("/software/:id", softwareHandler.updateSoftware)
	adminOnly.Delete("/software/:id", softwareHandler.deleteSoftware)

	// Alias
	readOnly.Get("/software/asset/:assetID/alternatives", softwareHandler.assetAlternatives)

	return app
}
```

Note: All handler methods are unexported but in the same `http` package as `server.go`, so they're directly accessible. Use exact method names from each handler file.

- [ ] **Step 3: Update main.go to pass jwtSecret**

In `cmd/server/main.go`, update the call to `NewServer`:

```go
// Before:
srv := httpTransport.NewServer(ctx, pool)

// After:
srv := httpTransport.NewServer(ctx, pool, cfg.JWTSecret)
```

- [ ] **Step 4: Verify it compiles**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add internal/transport/http/server.go internal/transport/http/*_handlers.go cmd/server/main.go
git commit -m "feat: wire auth into router with role-based access control"
```

---

### Task 8: Unit tests — Risk Calculator

**Files:**
- Create: `internal/service/risk/calculator_test.go`

- [ ] **Step 1: Write calculator tests**

```go
package risk

import (
	"testing"

	"Diplom/internal/domain"

	"github.com/stretchr/testify/assert"
)

func ptr[T any](v T) *T { return &v }

func baseAsset() *domain.Asset {
	return &domain.Asset{
		ID:                  1,
		Name:                "Test Server",
		BusinessCriticality: 3,
		Confidentiality:     3,
		Integrity:           3,
		Availability:        3,
		Environment:         domain.AssetEnvProd,
		HasInternetAccess:   true,
	}
}

func baseThreat() *domain.Threat {
	return &domain.Threat{
		ID:                    1,
		Name:                  "Test Threat",
		BaseLikelihood:        3,
		ImpactConfidentiality: false,
		ImpactIntegrity:       false,
		ImpactAvailability:    false,
	}
}

func TestCalculate_BasicRisk(t *testing.T) {
	calc := NewCalculator()
	asset := baseAsset()
	threat := baseThreat()
	vulns := []domain.Vulnerability{{Severity: 3}}

	result := calc.Calculate(asset, threat, vulns)

	assert.GreaterOrEqual(t, result.Impact, int16(1))
	assert.LessOrEqual(t, result.Impact, int16(5))
	assert.GreaterOrEqual(t, result.Likelihood, int16(1))
	assert.LessOrEqual(t, result.Likelihood, int16(5))
	assert.Equal(t, result.Impact*result.Likelihood, result.Score)
}

func TestCalculate_HighCriticalityIncreasesImpact(t *testing.T) {
	calc := NewCalculator()
	threat := baseThreat()
	vulns := []domain.Vulnerability{{Severity: 3}}

	lowCrit := baseAsset()
	lowCrit.BusinessCriticality = 1
	resultLow := calc.Calculate(lowCrit, threat, vulns)

	highCrit := baseAsset()
	highCrit.BusinessCriticality = 5
	resultHigh := calc.Calculate(highCrit, threat, vulns)

	assert.Greater(t, resultHigh.Impact, resultLow.Impact)
}

func TestCalculate_HighSeverityIncreasesImpact(t *testing.T) {
	calc := NewCalculator()
	asset := baseAsset()
	threat := baseThreat()

	lowSev := []domain.Vulnerability{{Severity: 1}}
	highSev := []domain.Vulnerability{{Severity: 5}}

	resultLow := calc.Calculate(asset, threat, lowSev)
	resultHigh := calc.Calculate(asset, threat, highSev)

	assert.GreaterOrEqual(t, resultHigh.Impact, resultLow.Impact)
}

func TestCalculate_ProdEnvIncreasesLikelihood(t *testing.T) {
	calc := NewCalculator()
	threat := baseThreat()
	vulns := []domain.Vulnerability{{Severity: 3}}

	devAsset := baseAsset()
	devAsset.Environment = domain.AssetEnvDev
	resultDev := calc.Calculate(devAsset, threat, vulns)

	prodAsset := baseAsset()
	prodAsset.Environment = domain.AssetEnvProd
	resultProd := calc.Calculate(prodAsset, threat, vulns)

	assert.GreaterOrEqual(t, resultProd.Likelihood, resultDev.Likelihood)
}

func TestCalculate_IsolatedReducesLikelihood(t *testing.T) {
	calc := NewCalculator()
	threat := baseThreat()
	threat.BaseLikelihood = 4
	vulns := []domain.Vulnerability{{Severity: 3}}

	open := baseAsset()
	open.IsIsolated = false
	resultOpen := calc.Calculate(open, threat, vulns)

	isolated := baseAsset()
	isolated.IsIsolated = true
	resultIsolated := calc.Calculate(isolated, threat, vulns)

	assert.LessOrEqual(t, resultIsolated.Likelihood, resultOpen.Likelihood)
}

func TestCalculate_ThreatCIABonus(t *testing.T) {
	calc := NewCalculator()
	vulns := []domain.Vulnerability{{Severity: 3}}

	asset := baseAsset()
	asset.Confidentiality = 5
	asset.Integrity = 5
	asset.Availability = 5

	noCIA := baseThreat()
	resultNo := calc.Calculate(asset, noCIA, vulns)

	fullCIA := baseThreat()
	fullCIA.ImpactConfidentiality = true
	fullCIA.ImpactIntegrity = true
	fullCIA.ImpactAvailability = true
	resultFull := calc.Calculate(asset, fullCIA, vulns)

	assert.Greater(t, resultFull.Impact, resultNo.Impact)
}

func TestCalculate_RiskLevels(t *testing.T) {
	tests := []struct {
		score int16
		level string
	}{
		{1, "Low"},
		{5, "Low"},
		{6, "Medium"},
		{10, "Medium"},
		{11, "High"},
		{15, "High"},
		{16, "Critical"},
		{25, "Critical"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.level, riskLevel(tt.score), "score=%d", tt.score)
	}
}

func TestCalculate_RegulatoryFactor_KII(t *testing.T) {
	calc := NewCalculator()
	asset := baseAsset()
	asset.KIICategory = ptr(domain.KIICategoryCat1)
	threat := baseThreat()
	vulns := []domain.Vulnerability{{Severity: 3}}

	result := calc.Calculate(asset, threat, vulns)

	assert.GreaterOrEqual(t, result.RegulatoryFactor, 2.0)
	assert.Greater(t, result.AdjustedScore, float64(result.Score))
}

func TestCalculate_RegulatoryFactor_StateSecret(t *testing.T) {
	calc := NewCalculator()
	asset := baseAsset()
	asset.DataCategory = ptr(domain.DataCategoryStateSecret)
	threat := baseThreat()
	vulns := []domain.Vulnerability{{Severity: 3}}

	result := calc.Calculate(asset, threat, vulns)

	assert.GreaterOrEqual(t, result.RegulatoryFactor, 2.5)
}

func TestCalculate_RegulatoryFactor_PersonalDataVolume(t *testing.T) {
	calc := NewCalculator()
	asset := baseAsset()
	asset.HasPersonalData = true
	asset.PersonalDataVolume = ptr(">100000")
	asset.DataCategory = ptr(domain.DataCategoryPersonalData)
	threat := baseThreat()
	vulns := []domain.Vulnerability{{Severity: 3}}

	result := calc.Calculate(asset, threat, vulns)

	// 1.3 (personal data) * 1.2 (volume > 100k) = 1.56
	assert.Greater(t, result.RegulatoryFactor, 1.5)
}

func TestCalculate_RegulatoryFactor_NoRegulatory(t *testing.T) {
	calc := NewCalculator()
	asset := baseAsset()
	threat := baseThreat()
	vulns := []domain.Vulnerability{{Severity: 3}}

	result := calc.Calculate(asset, threat, vulns)

	assert.Equal(t, 1.0, result.RegulatoryFactor)
	assert.Equal(t, float64(result.Score), result.AdjustedScore)
}

func TestCalculate_NoVulns(t *testing.T) {
	calc := NewCalculator()
	asset := baseAsset()
	threat := baseThreat()

	result := calc.Calculate(asset, threat, nil)

	assert.GreaterOrEqual(t, result.Impact, int16(1))
	assert.GreaterOrEqual(t, result.Likelihood, int16(1))
}

func TestCalculate_ClampValues(t *testing.T) {
	calc := NewCalculator()

	asset := baseAsset()
	asset.BusinessCriticality = 5
	asset.Confidentiality = 5
	asset.Integrity = 5
	asset.Availability = 5

	threat := baseThreat()
	threat.BaseLikelihood = 5
	threat.ImpactConfidentiality = true
	threat.ImpactIntegrity = true
	threat.ImpactAvailability = true

	vulns := []domain.Vulnerability{{Severity: 10}}

	result := calc.Calculate(asset, threat, vulns)

	assert.LessOrEqual(t, result.Impact, int16(5))
	assert.LessOrEqual(t, result.Likelihood, int16(5))
	assert.LessOrEqual(t, result.Score, int16(25))
}

func TestAdjustedRiskLevel(t *testing.T) {
	tests := []struct {
		score float64
		level string
	}{
		{5.0, "Low"},
		{11.9, "Low"},
		{12.0, "Medium"},
		{21.9, "Medium"},
		{22.0, "High"},
		{31.9, "High"},
		{32.0, "Critical"},
		{50.0, "Critical"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.level, AdjustedRiskLevel(tt.score), "score=%.1f", tt.score)
	}
}

func TestMaxVulnSeverity(t *testing.T) {
	assert.Equal(t, int16(1), maxVulnSeverity(nil))
	assert.Equal(t, int16(1), maxVulnSeverity([]domain.Vulnerability{}))
	assert.Equal(t, int16(7), maxVulnSeverity([]domain.Vulnerability{
		{Severity: 3}, {Severity: 7}, {Severity: 2},
	}))
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./internal/service/risk/ -v`
Expected: all PASS

- [ ] **Step 3: Commit**

```bash
git add internal/service/risk/calculator_test.go
git commit -m "test: add unit tests for risk calculator (15 tests)"
```

---

### Task 9: Unit tests — CIA Calculator

**Files:**
- Create: `internal/service/asset/cia_calculator_test.go`

- [ ] **Step 1: Write CIA calculator tests**

```go
package asset

import (
	"testing"

	"Diplom/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestCalculateCIA_Database(t *testing.T) {
	c, i, a := CalculateCIA("database", "prod", 3)
	// Database: high C and I, prod env adds +1
	assert.Equal(t, int16(5), c) // 4 + 1(prod)
	assert.Equal(t, int16(5), i) // 5 + 1(prod) clamped to 5
	assert.Equal(t, int16(4), a) // 3 + 1(prod)
}

func TestCalculateCIA_DevReducesValues(t *testing.T) {
	cProd, iProd, aProd := CalculateCIA("server", "prod", 3)
	cDev, iDev, aDev := CalculateCIA("server", "dev", 3)

	assert.Greater(t, cProd, cDev)
	assert.Greater(t, iProd, iDev)
	assert.Greater(t, aProd, aDev)
}

func TestCalculateCIA_HighCriticalityIncreasesValues(t *testing.T) {
	cLow, iLow, aLow := CalculateCIA("application", "prod", 1)
	cHigh, iHigh, aHigh := CalculateCIA("application", "prod", 5)

	assert.Greater(t, cHigh, cLow)
	assert.Greater(t, iHigh, iLow)
	assert.Greater(t, aHigh, aLow)
}

func TestCalculateCIA_ClampedTo1_5(t *testing.T) {
	// Extreme low: dev + criticality 1 on workstation (base 2,2,3)
	c, i, a := CalculateCIA("workstation", "dev", 1)
	assert.GreaterOrEqual(t, c, int16(1))
	assert.GreaterOrEqual(t, i, int16(1))
	assert.GreaterOrEqual(t, a, int16(1))

	// Extreme high: prod + criticality 5 on cloud (base 4,4,4)
	c, i, a = CalculateCIA("cloud", "prod", 5)
	assert.LessOrEqual(t, c, int16(5))
	assert.LessOrEqual(t, i, int16(5))
	assert.LessOrEqual(t, a, int16(5))
}

func TestCalculateCIA_UnknownType(t *testing.T) {
	c, i, a := CalculateCIA("unknown_type", "prod", 3)
	// Default base: 3,3,3 + prod(+1) = 4,4,4
	assert.Equal(t, int16(4), c)
	assert.Equal(t, int16(4), i)
	assert.Equal(t, int16(4), a)
}

func TestCalculateCriticality_StateSecret(t *testing.T) {
	cat := "state_secret"
	result := CalculateCriticality(&cat, nil, nil, false, nil, false, false, "prod")
	assert.Equal(t, int16(5), result) // 5.0 + 0.5(prod) clamped to 5
}

func TestCalculateCriticality_DevLowersScore(t *testing.T) {
	cat := "internal"
	prod := CalculateCriticality(&cat, nil, nil, false, nil, false, false, "prod")
	dev := CalculateCriticality(&cat, nil, nil, false, nil, false, false, "dev")
	assert.Greater(t, prod, dev)
}

func TestApplyCIA(t *testing.T) {
	assetType := "database"
	a := &domain.Asset{
		Type:                &assetType,
		Environment:         domain.AssetEnvProd,
		BusinessCriticality: 3,
	}
	ApplyCIA(a)

	assert.GreaterOrEqual(t, a.Confidentiality, int16(1))
	assert.LessOrEqual(t, a.Confidentiality, int16(5))
	assert.GreaterOrEqual(t, a.Integrity, int16(1))
	assert.GreaterOrEqual(t, a.Availability, int16(1))
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./internal/service/asset/ -v`
Expected: all PASS

- [ ] **Step 3: Commit**

```bash
git add internal/service/asset/cia_calculator_test.go
git commit -m "test: add unit tests for CIA calculator (8 tests)"
```

---

### Task 10: Unit tests — Auth Service

**Files:**
- Create: `internal/service/auth/service_test.go`

- [ ] **Step 1: Write auth service tests**

```go
package auth

import (
	"context"
	"testing"

	"Diplom/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeUserRepo implements repository.UserRepository in-memory for testing.
type fakeUserRepo struct {
	users  map[string]*domain.User
	nextID int64
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[string]*domain.User), nextID: 1}
}

func (r *fakeUserRepo) Create(_ context.Context, user *domain.User) error {
	user.ID = r.nextID
	r.nextID++
	r.users[user.Username] = user
	return nil
}

func (r *fakeUserRepo) GetByUsername(_ context.Context, username string) (*domain.User, error) {
	u, ok := r.users[username]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (r *fakeUserRepo) GetByID(_ context.Context, id int64) (*domain.User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

func TestRegister_Success(t *testing.T) {
	svc := NewService(newFakeUserRepo(), "test-secret")

	user, err := svc.Register(context.Background(), "admin", "password123", domain.UserRoleAdmin)

	require.NoError(t, err)
	assert.Equal(t, "admin", user.Username)
	assert.Equal(t, domain.UserRoleAdmin, user.Role)
	assert.True(t, user.IsActive)
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotEqual(t, "password123", user.PasswordHash) // must be hashed
}

func TestRegister_DuplicateUser(t *testing.T) {
	svc := NewService(newFakeUserRepo(), "test-secret")

	_, err := svc.Register(context.Background(), "admin", "password123", domain.UserRoleAdmin)
	require.NoError(t, err)

	_, err = svc.Register(context.Background(), "admin", "other", domain.UserRoleViewer)
	assert.ErrorIs(t, err, ErrUserExists)
}

func TestLogin_Success(t *testing.T) {
	svc := NewService(newFakeUserRepo(), "test-secret")

	_, err := svc.Register(context.Background(), "user1", "pass123", domain.UserRoleViewer)
	require.NoError(t, err)

	token, err := svc.Login(context.Background(), "user1", "pass123")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := NewService(newFakeUserRepo(), "test-secret")

	_, err := svc.Register(context.Background(), "user1", "pass123", domain.UserRoleViewer)
	require.NoError(t, err)

	_, err = svc.Login(context.Background(), "user1", "wrong")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLogin_NonExistentUser(t *testing.T) {
	svc := NewService(newFakeUserRepo(), "test-secret")

	_, err := svc.Login(context.Background(), "nobody", "pass")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestValidateToken_Success(t *testing.T) {
	svc := NewService(newFakeUserRepo(), "test-secret")

	_, err := svc.Register(context.Background(), "user1", "pass123", domain.UserRoleAuditor)
	require.NoError(t, err)

	token, err := svc.Login(context.Background(), "user1", "pass123")
	require.NoError(t, err)

	claims, err := svc.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user1", claims.Username)
	assert.Equal(t, "auditor", claims.Role)
	assert.Equal(t, int64(1), claims.UserID)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	svc := NewService(newFakeUserRepo(), "test-secret")

	_, err := svc.ValidateToken("invalid.token.here")
	assert.Error(t, err)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	svc1 := NewService(newFakeUserRepo(), "secret-1")
	repo := newFakeUserRepo()
	svc2 := NewService(repo, "secret-2")

	// Register and login with svc1's repo
	svc1Repo := newFakeUserRepo()
	svc1WithRepo := NewService(svc1Repo, "secret-1")
	_, _ = svc1WithRepo.Register(context.Background(), "user1", "pass", domain.UserRoleViewer)
	token, _ := svc1WithRepo.Login(context.Background(), "user1", "pass")

	// Validate with different secret
	_, err := svc2.ValidateToken(token)
	assert.Error(t, err)
	_ = svc1 // silence unused
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./internal/service/auth/ -v`
Expected: all PASS

- [ ] **Step 3: Commit**

```bash
git add internal/service/auth/service_test.go
git commit -m "test: add unit tests for auth service (8 tests)"
```

---

### Task 11: Integration test helper (testcontainers)

**Files:**
- Create: `internal/repository/testhelper_test.go`

- [ ] **Step 1: Write test helper with PostgreSQL container**

```go
package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupTestDB starts a PostgreSQL container and runs all migrations.
// Returns a connected pgxpool.Pool and a cleanup function.
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("test_cyber_risk"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("terminate container: %v", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	runMigrations(t, pool)

	return pool
}

func runMigrations(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	migrationsDir := findMigrationsDir(t)

	// Find all .up.sql files and sort them
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	var upFiles []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			if len(e.Name()) > 7 && e.Name()[len(e.Name())-7:] == ".up.sql" {
				upFiles = append(upFiles, e.Name())
			}
		}
	}
	sort.Strings(upFiles)

	for _, f := range upFiles {
		sql, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			t.Fatalf("read migration %s: %v", f, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			t.Fatalf("run migration %s: %v", f, err)
		}
	}
}

func findMigrationsDir(t *testing.T) string {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	// Go up from internal/repository/ to project root
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	dir := filepath.Join(projectRoot, "migrations")
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("migrations dir not found at %s: %v", dir, err)
	}
	return dir
}

// truncateAll clears data from all tables for test isolation.
func truncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	tables := []string{
		"risk_scenario_recommendations",
		"recommendation_templates",
		"risk_scenarios",
		"asset_controls",
		"controls",
		"asset_vulnerabilities",
		"asset_software",
		"software",
		"vulnerabilities",
		"threats",
		"assets",
		"users",
	}
	for _, table := range tables {
		_, err := pool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Logf("truncate %s: %v", table, err)
		}
	}
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/repository/...`
Expected: no errors (test files are not compiled by `go build`, but we verify no syntax issues)

Run: `go vet ./internal/repository/...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/repository/testhelper_test.go
git commit -m "test: add testcontainers PostgreSQL helper for integration tests"
```

---

### Task 12: Integration tests — Asset Repository

**Files:**
- Create: `internal/repository/asset_repository_test.go`

- [ ] **Step 1: Write asset repository integration tests**

```go
package repository

import (
	"context"
	"testing"

	"Diplom/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewAssetRepository(pool)
	ctx := context.Background()

	asset := &domain.Asset{
		Name:                "Test Server",
		BusinessCriticality: 3,
		Confidentiality:     4,
		Integrity:           3,
		Availability:        5,
		Environment:         domain.AssetEnvProd,
		HasInternetAccess:   true,
	}

	err := repo.Create(ctx, asset)
	require.NoError(t, err)
	assert.Greater(t, asset.ID, int64(0))
}

func TestAssetRepository_GetByID(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewAssetRepository(pool)
	ctx := context.Background()

	asset := &domain.Asset{
		Name:                "Test DB",
		BusinessCriticality: 4,
		Confidentiality:     5,
		Integrity:           4,
		Availability:        3,
		Environment:         domain.AssetEnvProd,
	}
	err := repo.Create(ctx, asset)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, asset.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "Test DB", found.Name)
	assert.Equal(t, int16(4), found.BusinessCriticality)
}

func TestAssetRepository_List(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewAssetRepository(pool)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		err := repo.Create(ctx, &domain.Asset{
			Name:                fmt.Sprintf("Asset %d", i),
			BusinessCriticality: 3,
			Confidentiality:     3,
			Integrity:           3,
			Availability:        3,
			Environment:         domain.AssetEnvDev,
		})
		require.NoError(t, err)
	}

	assets, err := repo.List(ctx, AssetFilter{Limit: 10, Offset: 0})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(assets), 3)
}

func TestAssetRepository_Update(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewAssetRepository(pool)
	ctx := context.Background()

	asset := &domain.Asset{
		Name:                "Original",
		BusinessCriticality: 2,
		Confidentiality:     2,
		Integrity:           2,
		Availability:        2,
		Environment:         domain.AssetEnvDev,
	}
	err := repo.Create(ctx, asset)
	require.NoError(t, err)

	asset.Name = "Updated"
	asset.BusinessCriticality = 5
	err = repo.Update(ctx, asset)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, asset.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated", found.Name)
	assert.Equal(t, int16(5), found.BusinessCriticality)
}

func TestAssetRepository_Delete(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewAssetRepository(pool)
	ctx := context.Background()

	asset := &domain.Asset{
		Name:                "ToDelete",
		BusinessCriticality: 1,
		Confidentiality:     1,
		Integrity:           1,
		Availability:        1,
		Environment:         domain.AssetEnvTest,
	}
	err := repo.Create(ctx, asset)
	require.NoError(t, err)

	err = repo.Delete(ctx, asset.ID)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, asset.ID)
	require.NoError(t, err)
	assert.Nil(t, found)
}
```

Note: Add `"fmt"` to imports for the `fmt.Sprintf` in `TestAssetRepository_List`.

- [ ] **Step 2: Run tests (requires Docker)**

Run: `go test ./internal/repository/ -run TestAssetRepository -v -timeout 120s`
Expected: all PASS (may take 15-30s for container startup)

- [ ] **Step 3: Commit**

```bash
git add internal/repository/asset_repository_test.go
git commit -m "test: add integration tests for asset repository (5 tests)"
```

---

### Task 13: Integration tests — Threat, Vulnerability, Software, User Repositories

**Files:**
- Create: `internal/repository/threat_repository_test.go`
- Create: `internal/repository/vulnerability_repository_test.go`
- Create: `internal/repository/software_repository_test.go`
- Create: `internal/repository/user_repository_test.go`

- [ ] **Step 1: Write threat repository tests**

```go
package repository

import (
	"context"
	"testing"

	"Diplom/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThreatRepository_CreateAndGet(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewThreatRepository(pool)
	ctx := context.Background()

	threat := &domain.Threat{
		Name:           "SQL Injection",
		SourceType:     domain.ThreatSourceExternal,
		BaseLikelihood: 4,
	}
	err := repo.Create(ctx, threat)
	require.NoError(t, err)
	assert.Greater(t, threat.ID, int64(0))

	found, err := repo.GetByID(ctx, threat.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "SQL Injection", found.Name)
}

func TestThreatRepository_List(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewThreatRepository(pool)
	ctx := context.Background()

	_ = repo.Create(ctx, &domain.Threat{Name: "T1", SourceType: domain.ThreatSourceExternal, BaseLikelihood: 3})
	_ = repo.Create(ctx, &domain.Threat{Name: "T2", SourceType: domain.ThreatSourceInternal, BaseLikelihood: 2})

	threats, err := repo.List(ctx, ThreatFilter{Limit: 10, Offset: 0})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(threats), 2)
}

func TestThreatRepository_Delete(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewThreatRepository(pool)
	ctx := context.Background()

	threat := &domain.Threat{Name: "ToDelete", SourceType: domain.ThreatSourceExternal, BaseLikelihood: 1}
	_ = repo.Create(ctx, threat)

	err := repo.Delete(ctx, threat.ID)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, threat.ID)
	require.NoError(t, err)
	assert.Nil(t, found)
}
```

- [ ] **Step 2: Write vulnerability repository tests**

```go
package repository

import (
	"context"
	"testing"

	"Diplom/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVulnerabilityRepository_CreateAndGet(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewVulnerabilityRepository(pool)
	ctx := context.Background()

	vuln := &domain.Vulnerability{
		Name:     "XSS",
		Severity: 7,
	}
	err := repo.Create(ctx, vuln)
	require.NoError(t, err)
	assert.Greater(t, vuln.ID, int64(0))

	found, err := repo.GetByID(ctx, vuln.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "XSS", found.Name)
	assert.Equal(t, int16(7), found.Severity)
}

func TestVulnerabilityRepository_ListByIDs(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewVulnerabilityRepository(pool)
	ctx := context.Background()

	v1 := &domain.Vulnerability{Name: "V1", Severity: 3}
	v2 := &domain.Vulnerability{Name: "V2", Severity: 5}
	_ = repo.Create(ctx, v1)
	_ = repo.Create(ctx, v2)

	vulns, err := repo.ListByIDs(ctx, []int64{v1.ID, v2.ID})
	require.NoError(t, err)
	assert.Len(t, vulns, 2)
}

func TestVulnerabilityRepository_Delete(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewVulnerabilityRepository(pool)
	ctx := context.Background()

	vuln := &domain.Vulnerability{Name: "ToDelete", Severity: 1}
	_ = repo.Create(ctx, vuln)

	err := repo.Delete(ctx, vuln.ID)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, vuln.ID)
	require.NoError(t, err)
	assert.Nil(t, found)
}
```

- [ ] **Step 3: Write software repository tests**

```go
package repository

import (
	"context"
	"testing"

	"Diplom/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSoftwareRepository_CreateAndGet(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewSoftwareRepository(pool)
	ctx := context.Background()

	sw := &domain.Software{
		Name:      "Kaspersky Endpoint Security",
		Vendor:    "Kaspersky Lab",
		IsRussian: true,
	}
	err := repo.Create(ctx, sw)
	require.NoError(t, err)
	assert.Greater(t, sw.ID, int64(0))

	found, err := repo.GetByID(ctx, sw.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "Kaspersky Endpoint Security", found.Name)
	assert.True(t, found.IsRussian)
}

func TestSoftwareRepository_ListWithFilter(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewSoftwareRepository(pool)
	ctx := context.Background()

	_ = repo.Create(ctx, &domain.Software{Name: "KES", Vendor: "Kaspersky", IsRussian: true})
	_ = repo.Create(ctx, &domain.Software{Name: "Norton", Vendor: "Symantec", IsRussian: false})

	russian := true
	list, err := repo.List(ctx, SoftwareFilter{Limit: 10, IsRussian: &russian})
	require.NoError(t, err)
	for _, sw := range list {
		assert.True(t, sw.IsRussian)
	}
}

func TestSoftwareRepository_Delete(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewSoftwareRepository(pool)
	ctx := context.Background()

	sw := &domain.Software{Name: "ToDelete", Vendor: "Test"}
	_ = repo.Create(ctx, sw)

	err := repo.Delete(ctx, sw.ID)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, sw.ID)
	require.NoError(t, err)
	assert.Nil(t, found)
}
```

- [ ] **Step 4: Write user repository tests**

```go
package repository

import (
	"context"
	"testing"

	"Diplom/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_CreateAndGetByUsername(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	user := &domain.User{
		Username:     "testadmin",
		PasswordHash: "$2a$10$fakehashhere",
		Role:         domain.UserRoleAdmin,
		IsActive:     true,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)
	assert.Greater(t, user.ID, int64(0))

	found, err := repo.GetByUsername(ctx, "testadmin")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, domain.UserRoleAdmin, found.Role)
}

func TestUserRepository_GetByID(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	user := &domain.User{
		Username:     "viewer1",
		PasswordHash: "$2a$10$fakehash",
		Role:         domain.UserRoleViewer,
		IsActive:     true,
	}
	_ = repo.Create(ctx, user)

	found, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "viewer1", found.Username)
}

func TestUserRepository_GetByUsername_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	truncateAll(t, pool)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	found, err := repo.GetByUsername(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, found)
}
```

- [ ] **Step 5: Run all integration tests**

Run: `go test ./internal/repository/ -v -timeout 120s`
Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add internal/repository/*_test.go
git commit -m "test: add integration tests for threat, vulnerability, software, user repositories"
```

---

### Task 14: Final verification

- [ ] **Step 1: Run all tests**

```bash
go test ./... -v -timeout 180s
```

Expected: all tests PASS

- [ ] **Step 2: Verify build**

```bash
go build ./...
```

Expected: no errors

- [ ] **Step 3: Final commit if any fixes needed**

```bash
git add -A
git commit -m "fix: resolve any test/build issues"
```
