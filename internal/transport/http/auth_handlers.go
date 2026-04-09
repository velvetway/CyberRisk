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

func (h *AuthHandler) RegisterRoutes(g fiber.Router) {
	g.Post("/register", h.register)
	g.Post("/login", h.login)
}

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
