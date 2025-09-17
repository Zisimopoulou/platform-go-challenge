 package core

import (
	"errors"

	"github.com/Zisimopoulou/platform-go-challenge/internal/data"
	"github.com/Zisimopoulou/platform-go-challenge/internal/models"
	"github.com/Zisimopoulou/platform-go-challenge/internal/validation"
)

 type Service struct {
	store data.Store
}

func NewService(s data.Store) *Service {
	return &Service{store: s}
}

func (s *Service) AddFavorite(userID string, asset models.RawAsset) (string, error) {
 	if err := validation.ValidateAsset(&asset); err != nil {
		return "", err
	}

	return s.store.Add(userID, asset)
}

func (s *Service) ListFavorites(userID string) ([]models.Favorite, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	return s.store.List(userID)
}

func (s *Service) DeleteFavorite(userID, favID string) error {
	if userID == "" {
		return errors.New("user ID cannot be empty")
	}
	if favID == "" {
		return errors.New("favorite ID cannot be empty")
	}

	return s.store.Delete(userID, favID)
}

func (s *Service) UpdateDescription(userID, favID, desc string) error {
	if userID == "" {
		return errors.New("user ID cannot be empty")
	}
	if favID == "" {
		return errors.New("favorite ID cannot be empty")
	}
	if desc == "" {
		return errors.New("description cannot be empty")
	}

	if len(desc) > 500 {
		return errors.New("description cannot exceed 500 characters")
	}

	return s.store.UpdateDescription(userID, favID, desc)
}
