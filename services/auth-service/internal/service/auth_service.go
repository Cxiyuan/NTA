package service

import (
	"context"
	"errors"
	"time"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/Cxiyuan/NTA/services/auth-service/internal/repository"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo      *repository.AuthRepository
	logger    *logrus.Logger
	jwtSecret string
}

type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

func NewAuthService(repo *repository.AuthRepository, logger *logrus.Logger, jwtSecret string) *AuthService {
	return &AuthService{
		repo:      repo,
		logger:    logger,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, *models.User, []string, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		s.logger.Warnf("Login attempt for non-existent user: %s", username)
		return "", nil, nil, errors.New("invalid credentials")
	}

	if user.Status != models.StatusActive {
		s.logger.Warnf("Login attempt for inactive user: %s", username)
		return "", nil, nil, errors.New("user inactive")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logger.Warnf("Failed login attempt for user: %s", username)
		return "", nil, nil, errors.New("invalid credentials")
	}

	roles, err := s.repo.GetUserRoles(ctx, user.ID)
	if err != nil {
		s.logger.Errorf("Failed to get roles for user %s: %v", username, err)
		roles = []string{models.RoleViewer}
	}

	if len(roles) == 0 {
		roles = []string{models.RoleViewer}
	}

	token, err := s.generateToken(user.ID, user.Username, roles)
	if err != nil {
		s.logger.Errorf("Failed to generate token for user %s: %v", username, err)
		return "", nil, nil, errors.New("token generation failed")
	}

	s.logger.Infof("User %s logged in successfully", username)
	return token, user, roles, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *AuthService) GetUserByID(ctx context.Context, userID uint) (*models.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

func (s *AuthService) CreateUser(ctx context.Context, user *models.User, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedPassword)
	return s.repo.CreateUser(ctx, user)
}

func (s *AuthService) UpdateUser(ctx context.Context, userID uint, updates map[string]interface{}) error {
	return s.repo.UpdateUser(ctx, userID, updates)
}

func (s *AuthService) DeleteUser(ctx context.Context, userID uint) error {
	return s.repo.DeleteUser(ctx, userID)
}

func (s *AuthService) ResetPassword(ctx context.Context, userID uint, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.UpdateUser(ctx, userID, map[string]interface{}{
		"password_hash": string(hashedPassword),
	})
}

func (s *AuthService) generateToken(userID uint, username string, roles []string) (string, error) {
	claims := &Claims{
		UserID:   string(userID),
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
