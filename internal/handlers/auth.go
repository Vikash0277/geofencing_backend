package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"geofencing_backend/database"
	"geofencing_backend/internal/dto"
	"geofencing_backend/internal/models"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOauthConfig *oauth2.Config

func InitOauthConfig() {
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

func generateToken(user models.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default_secret"
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Provider: "local",
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user or email already exists"})
	}

	token, err := generateToken(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.AuthResponse{
		Token: token,
		User: fiber.Map{
			"id":       user.ID,
			"name":     user.Name,
			"email":    user.Email,
			"role":     user.Role,
			"provider": user.Provider,
		},
	})
}

func Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if user.Provider != "local" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Please login using " + user.Provider})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	token, err := generateToken(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.JSON(dto.AuthResponse{
		Token: token,
		User: fiber.Map{
			"id":       user.ID,
			"name":     user.Name,
			"email":    user.Email,
			"role":     user.Role,
			"provider": user.Provider,
		},
	})
}

func GoogleLogin(c *fiber.Ctx) error {
	if googleOauthConfig == nil {
		InitOauthConfig()
	}
	// Note: in a real app, generate a random state string and verify it in the callback
	url := googleOauthConfig.AuthCodeURL("random_state_string", oauth2.AccessTypeOffline)
	return c.Redirect(url)
}

func GoogleCallback(c *fiber.Ctx) error {
	if googleOauthConfig == nil {
		InitOauthConfig()
	}

	state := c.Query("state")
	if state != "random_state_string" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid OAuth state"})
	}

	code := c.Query("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Println("OAuth Exchange Error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to exchange token"})
	}

	client := googleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Println("Get UserInfo Error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get user info"})
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode user info"})
	}

	var user models.User
	err = database.DB.Where("email = ?", userInfo.Email).First(&user).Error
	if err != nil {
		// User does not exist, create a new one
		user = models.User{
			Name:       userInfo.Name,
			Email:      userInfo.Email,
			Provider:   "google",
			ProviderID: userInfo.ID,
		}
		if err := database.DB.Create(&user).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
		}
	} else {
		// User exists, check if provider is google
		if user.Provider != "google" {
			// They used local auth before, we can either link accounts or return an error.
			// Returning an error for simplicity.
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Email already registered with " + user.Provider + " login."})
		}
	}

	jwtToken, err := generateToken(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	// Build the redirect URL back to the frontend with auth data
	// Build the redirect URL back to the frontend with token and user data
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	// Encode user data as JSON and pass via URL query parameters
	userJSON, _ := json.Marshal(fiber.Map{
		"id":       user.ID,
		"name":     user.Name,
		"email":    user.Email,
		"role":     user.Role,
		"provider": user.Provider,
	})

	redirectURL := fmt.Sprintf("%s?token=%s&user=%s", frontendURL, url.QueryEscape(jwtToken), url.QueryEscape(string(userJSON)))
	return c.Redirect(redirectURL)
}
