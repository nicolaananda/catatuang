package service

import (
	"context"
	"fmt"
	"time"

	"github.com/nicolaananda/catatuang/internal/domain"
	"github.com/nicolaananda/catatuang/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetOrCreateUser(ctx context.Context, msisdn string) (*domain.User, bool, error) {
	// Try to get existing user
	user, err := s.userRepo.GetByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get user: %w", err)
	}

	if user != nil {
		return user, false, nil
	}

	// Create new user with NEW_USER state
	newUser := &domain.User{
		MSISDN:      msisdn,
		Plan:        domain.PlanFree,
		FreeTxCount: 0,
		IsBlocked:   false,
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, false, fmt.Errorf("failed to create user: %w", err)
	}

	return newUser, true, nil
}

func (s *UserService) UpgradeToPremium(ctx context.Context, msisdn string, startDate time.Time, months int) error {
	user, err := s.userRepo.GetByMSISDN(ctx, msisdn)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	premiumUntil := startDate.AddDate(0, months, 0)
	user.Plan = domain.PlanPremium
	user.PremiumUntil = &premiumUntil

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to upgrade user: %w", err)
	}

	return nil
}

func (s *UserService) DowngradeToFree(ctx context.Context, msisdn string) error {
	user, err := s.userRepo.GetByMSISDN(ctx, msisdn)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	user.Plan = domain.PlanFree
	user.PremiumUntil = nil

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to downgrade user: %w", err)
	}

	return nil
}

func (s *UserService) BlockUser(ctx context.Context, msisdn string) error {
	user, err := s.userRepo.GetByMSISDN(ctx, msisdn)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	user.IsBlocked = true

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to block user: %w", err)
	}

	return nil
}

func (s *UserService) UnblockUser(ctx context.Context, msisdn string) error {
	user, err := s.userRepo.GetByMSISDN(ctx, msisdn)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	user.IsBlocked = false

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to unblock user: %w", err)
	}

	return nil
}

func (s *UserService) GetUserStatus(ctx context.Context, msisdn string) (*domain.User, error) {
	user, err := s.userRepo.GetByMSISDN(ctx, msisdn)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	return s.userRepo.GetAll(ctx)
}
