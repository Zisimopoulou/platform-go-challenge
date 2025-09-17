package data


import "github.com/Zisimopoulou/platform-go-challenge/internal/models"


type Store interface {
Add(userID string, asset models.RawAsset) (string, error)
List(userID string) ([]models.Favorite, error)
Delete(userID, favID string) error
UpdateDescription(userID, favID, desc string) error
}