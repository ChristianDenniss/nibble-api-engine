package home

import (
	"time"

	homeentity "github.com/ChristianDenniss/go-data-model/home/entity"
)

// Wire shapes match the generated HomeFeed types in nibble-web-platform (camelCase).

type sponsoredMarkWire struct {
	PlacementID    string `json:"placementId"`
	CampaignID     string `json:"campaignId"`
	AdvertiserName string `json:"advertiserName"`
	Label          string `json:"label"`
}

type dealBadgeWire struct {
	PromotionID     string `json:"promotionId"`
	ChannelID       string `json:"channelId"`
	Label           string `json:"label"`
	FulfillmentMode string `json:"fulfillmentMode"`
}

type bannerWire struct {
	ID           string             `json:"id"`
	Kind         string             `json:"kind"`
	Headline     string             `json:"headline"`
	Body         string             `json:"body"`
	ImageURL     string             `json:"imageURL"`
	CallToAction string             `json:"callToAction"`
	RestaurantID string             `json:"restaurantId"`
	Sponsored    *sponsoredMarkWire `json:"sponsored"`
	Deal         *dealBadgeWire     `json:"deal"`
}

type feedItemWire struct {
	RestaurantID string             `json:"restaurantId"`
	Reason       string             `json:"reason"`
	Sponsored    *sponsoredMarkWire `json:"sponsored"`
	Deal         *dealBadgeWire     `json:"deal"`
}

type sectionWire struct {
	Kind  string         `json:"kind"`
	Title string         `json:"title"`
	Items []feedItemWire `json:"items"`
}

type feedResponse struct {
	GeneratedAt string        `json:"generatedAt"`
	Banners     []bannerWire  `json:"banners"`
	Sections    []sectionWire `json:"sections"`
}

func toFeedResponse(f homeentity.Feed) feedResponse {
	out := feedResponse{
		GeneratedAt: f.GeneratedAt.UTC().Format(time.RFC3339),
		Banners:     make([]bannerWire, 0, len(f.Banners)),
		Sections:    make([]sectionWire, 0, len(f.Sections)),
	}
	for _, b := range f.Banners {
		out.Banners = append(out.Banners, bannerWire{
			ID:           b.ID,
			Kind:         b.Kind,
			Headline:     b.Headline,
			Body:         b.Body,
			ImageURL:     b.ImageURL,
			CallToAction: b.CallToAction,
			RestaurantID: b.RestaurantID,
			Sponsored:    toSponsored(b.Sponsored),
			Deal:         toDeal(b.Deal),
		})
	}
	for _, s := range f.Sections {
		items := make([]feedItemWire, 0, len(s.Items))
		for _, item := range s.Items {
			items = append(items, feedItemWire{
				RestaurantID: item.RestaurantID,
				Reason:       item.Reason,
				Sponsored:    toSponsored(item.Sponsored),
				Deal:         toDeal(item.Deal),
			})
		}
		out.Sections = append(out.Sections, sectionWire{Kind: s.Kind, Title: s.Title, Items: items})
	}
	return out
}

func toSponsored(m *homeentity.SponsoredMark) *sponsoredMarkWire {
	if m == nil {
		return nil
	}
	return &sponsoredMarkWire{PlacementID: m.PlacementID, CampaignID: m.CampaignID, AdvertiserName: m.AdvertiserName, Label: m.Label}
}

func toDeal(d *homeentity.DealBadge) *dealBadgeWire {
	if d == nil {
		return nil
	}
	return &dealBadgeWire{PromotionID: d.PromotionID, ChannelID: d.ChannelID, Label: d.Label, FulfillmentMode: d.FulfillmentMode}
}
