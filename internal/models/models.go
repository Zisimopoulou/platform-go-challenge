package models

import "time"

type AssetType string

const (
	TypeChart    AssetType = "chart"
	TypeInsight  AssetType = "insight"
	TypeAudience AssetType = "audience"
)

type Chart struct {
	Title string `json:"title"`
	XAxis string `json:"xAxis"`
	YAxis string `json:"yAxis"`
	Data  []int  `json:"data"`
}

type Insight struct {
	Text string `json:"text"`
}

type Audience struct {
	Gender         string `json:"gender"`
	BirthCountry   string `json:"birthCountry"`
	AgeGroup       string `json:"ageGroup"`
	HoursDaily     string `json:"hoursDaily"`
	PurchasesLastM string `json:"purchasesLastMonth"`
}

type RawAsset struct {
	ID          string      `json:"id"`
	Type        AssetType   `json:"type"`
	Description string      `json:"description,omitempty"`
	CreatedAt   time.Time   `json:"createdAt"`
	Payload     interface{} `json:"payload"`
}

type Favorite struct {
	FavoriteID string   `json:"favoriteId"`
	Asset      RawAsset `json:"asset"`
}
