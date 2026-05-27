package service

import (
	"context"
	"errors"

	"Hertz/biz/model"
	"Hertz/infra/database"
	"Hertz/pkg/auth"
)

var ErrInvalidCredential = errors.New("invalid email or password")

type AuthService struct {
	db         UserRepository
	jwtManager *auth.JWTManager
}

type UserRepository interface {
	FindUserByEmail(ctx context.Context, email string) (database.User, error)
}

func NewAuthService(db UserRepository, jwtManager *auth.JWTManager) *AuthService {
	return &AuthService{
		db:         db,
		jwtManager: jwtManager,
	}
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (model.LoginResponse, error) {
	user, err := s.db.FindUserByEmail(ctx, req.Email)
	if err != nil || user.PasswordHash != req.Password {
		return model.LoginResponse{}, ErrInvalidCredential
	}

	token, claims, err := s.jwtManager.Sign(user.ID, user.Name, user.Email)
	if err != nil {
		return model.LoginResponse{}, err
	}

	return model.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresAt:   claims.Expires,
		User: model.UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		},
	}, nil
}
