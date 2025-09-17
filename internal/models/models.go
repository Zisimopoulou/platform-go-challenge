package models

import (
	"time"
)

type AssetType string

const (
	TypeChart    AssetType = "chart"
	TypeInsight  AssetType = "insight"
	TypeAudience AssetType = "audience"
)

type Chart struct {
	Title string `json:"title" validate:"required"`
	XAxis string `json:"xAxis" validate:"required"`
	YAxis string `json:"yAxis" validate:"required"`
	Data  []int  `json:"data" validate:"required,min=1"`
}

type Insight struct {
	Text string `json:"text" validate:"required,min=1,max=500"`
}

type Audience struct {
	Gender         string `json:"gender" validate:"required,oneof=Male Female Other"`
	BirthCountry   string `json:"birthCountry" validate:"required,min=2,max=100"`
	AgeGroup       string `json:"ageGroup" validate:"required,min=1,max=50"`
	HoursDaily     string `json:"hoursDaily" validate:"required,min=1,max=50"`
	PurchasesLastM string `json:"purchasesLastMonth" validate:"required,min=1,max=50"`
}

type RawAsset struct {
	ID          string      `json:"id"`
	Type        AssetType   `json:"type" validate:"required,oneof=chart insight audience"`
	Description string      `json:"description,omitempty" validate:"max=500"`
	CreatedAt   time.Time   `json:"createdAt"`
	Payload     interface{} `json:"payload" validate:"required"`
}

type Favorite struct {
	FavoriteID string   `json:"favoriteId"`
	Asset      RawAsset `json:"asset"`
}

type PaginatedFavorites struct {
	Favorites  []Favorite `json:"favorites"`
	TotalCount int        `json:"totalCount"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
	HasMore    bool       `json:"hasMore"`
}