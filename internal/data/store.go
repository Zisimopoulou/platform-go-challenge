package data

import "github.com/Zisimopoulou/platform-go-challenge/internal/models"

type Store interface {
	Add(userID string, asset models.RawAsset) (string, error)
	List(userID string, limit, offset int) ([]models.Favorite, int, error)
	Delete(userID, favID string) error
	UpdateDescription(userID, favID, desc string) error
}