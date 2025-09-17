package core


import (
"errors"


"github.com/Zisimopoulou/platform-go-challenge/internal/data"
"github.com/Zisimopoulou/platform-go-challenge/internal/models"
)


type Service struct {
store data.Store
}


func NewService(s data.Store) *Service {
return &Service{store: s}
}


func (s *Service) AddFavorite(userID string, asset models.RawAsset) (string, error) {
if asset.Type == "" {
return "", errors.New("asset type required")
}
return s.store.Add(userID, asset)
}


func (s *Service) ListFavorites(userID string) ([]models.Favorite, error) {
return s.store.List(userID)
}


func (s *Service) DeleteFavorite(userID, favID string) error {
return s.store.Delete(userID, favID)
}


func (s *Service) UpdateDescription(userID, favID, desc string) error {
return s.store.UpdateDescription(userID, favID, desc)
}