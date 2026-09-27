// Package catalog owns provider-neutral catalog validation and identity resolution.
package catalog

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

const ContentType = "application/vnd.nibble.catalog.v1+json"

// Bundle is an internal ingest envelope. Provenance never appears in PublicCatalog.
type Bundle struct {
	Version   int                 `json:"version"`
	Providers map[string]Snapshot `json:"providers"`
}
type Snapshot struct {
	City        string  `json:"city"`
	Region      string  `json:"region"`
	RetrievedAt string  `json:"retrievedAt"`
	Stores      []Store `json:"stores"`
}
type Store struct {
	ID                  string      `json:"id"`
	Name                string      `json:"name"`
	URL                 string      `json:"url"`
	Address             string      `json:"address"`
	Rating              string      `json:"rating"`
	Image               string      `json:"imageUrl"`
	Categories          []string    `json:"categories"`
	Items               []MenuItem  `json:"menuItems"`
	ObservedAt          string      `json:"observedAt,omitempty"`
	RetrievedAt         string      `json:"retrievedAt,omitempty"`
	DirectoryObservedAt string      `json:"directoryObservedAt,omitempty"`
	SourceAge           string      `json:"sourceCrawlAgeAtRetrieval,omitempty"`
	FulfillmentMode     string      `json:"fulfillmentMode,omitempty"`
	Promotions          []Promotion `json:"promotions,omitempty"`
}
type MenuItem struct {
	Name       string      `json:"name"`
	Section    string      `json:"section,omitempty"`
	Amount     *int64      `json:"amountCents"`
	Currency   string      `json:"currency"`
	PriceKind  string      `json:"priceKind,omitempty"`
	PriceNote  string      `json:"priceNote,omitempty"`
	Promotions []Promotion `json:"promotions,omitempty"`
}
type Promotion struct {
	Title      string `json:"title"`
	Conditions string `json:"conditions,omitempty"`
}
type Repository interface {
	Import(context.Context, Bundle, []byte) error
	Load(context.Context) (Bundle, error)
}
type Service struct{ repo Repository }

func New(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) Import(ctx context.Context, raw []byte) error {
	var b Bundle
	if err := json.Unmarshal(raw, &b); err != nil {
		return fmt.Errorf("invalid catalog: %w", err)
	}
	if err := b.Validate(); err != nil {
		return err
	}
	return s.repo.Import(ctx, b, raw)
}
func (s *Service) Read(ctx context.Context) (PublicCatalog, error) {
	b, err := s.repo.Load(ctx)
	if err != nil {
		return PublicCatalog{}, err
	}
	return Build(b), nil
}
func StoreID(provider, rawURL string) string {
	u, _ := url.Parse(rawURL)
	if u == nil {
		return ""
	}
	part := strings.TrimRight(u.Path, "/")
	part = part[strings.LastIndex(part, "/")+1:]
	if provider == "Uber Eats" {
		return fmt.Sprintf("ss_ubereats_%x", sha256.Sum256([]byte(part)))[:36]
	}
	part = part[strings.LastIndex(part, "-")+1:]
	if !regexp.MustCompile(`^\d+$`).MatchString(part) {
		return ""
	}
	return "ss_doordash_" + part
}
func (b Bundle) Validate() error {
	if b.Version != 1 || len(b.Providers) == 0 || len(b.Providers) > 2 {
		return fmt.Errorf("invalid catalog version/providers")
	}
	for provider, snapshot := range b.Providers {
		host := map[string]string{"Uber Eats": "www.ubereats.com", "DoorDash": "www.doordash.com"}[provider]
		if host == "" || snapshot.City != "Fredericton" || snapshot.Region != "NB" || len(snapshot.Stores) == 0 {
			return fmt.Errorf("invalid provider or city")
		}
		if _, err := time.Parse(time.RFC3339Nano, snapshot.RetrievedAt); err != nil {
			return fmt.Errorf("invalid retrieval time")
		}
		seen := map[string]bool{}
		for _, st := range snapshot.Stores {
			u, err := url.Parse(st.URL)
			if err != nil || u.Scheme != "https" || u.Host != host || u.User != nil || st.ID != StoreID(provider, st.URL) || st.Name == "" || seen[st.ID] {
				return fmt.Errorf("invalid or duplicate source store: %s", st.ID)
			}
			seen[st.ID] = true
			if st.ObservedAt == "" && st.SourceAge == "" && st.DirectoryObservedAt == "" {
				return fmt.Errorf("missing provenance: %s", st.ID)
			}
			for _, ts := range []string{st.ObservedAt, st.RetrievedAt, st.DirectoryObservedAt} {
				if ts != "" {
					if _, err := time.Parse(time.RFC3339Nano, ts); err != nil {
						return fmt.Errorf("invalid store timestamp")
					}
				}
			}
			for _, it := range st.Items {
				if strings.TrimSpace(it.Name) == "" || it.Currency != "CAD" || (it.Amount != nil && *it.Amount < 0) || (it.Amount == nil && it.PriceNote == "") {
					return fmt.Errorf("invalid price: %s", it.Name)
				}
				if it.PriceKind != "" && it.PriceKind != "base" && it.PriceKind != "from" && it.PriceKind != "options" {
					return fmt.Errorf("invalid price kind")
				}
			}
		}
	}
	return nil
}

type PublicCatalog struct {
	City        string        `json:"city"`
	Restaurants []*Restaurant `json:"restaurants"`
}
type Restaurant struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Address    string   `json:"address"`
	Rating     string   `json:"rating,omitempty"`
	Image      string   `json:"image,omitempty"`
	ArtworkKey string   `json:"artworkKey"`
	Categories []string `json:"categories"`
	Providers  []string `json:"providers"`
	Sources    []Source `json:"sources"`
	Items      []*Item  `json:"items"`
}
type Source struct {
	Provider      string `json:"provider"`
	URL           string `json:"url"`
	Address       string `json:"address"`
	MenuAvailable bool   `json:"menuAvailable"`
}
type Item struct {
	matchKey string
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Section  string  `json:"section"`
	Offers   []Offer `json:"offers"`
}
type Offer struct {
	Provider        string      `json:"provider"`
	Amount          *int64      `json:"amountCents"`
	Currency        string      `json:"currency"`
	PriceKind       string      `json:"priceKind"`
	PriceNote       string      `json:"priceNote,omitempty"`
	Status          string      `json:"status"`
	SourceURL       string      `json:"sourceUrl"`
	FulfillmentMode string      `json:"fulfillmentMode"`
	Delivery        *int64      `json:"deliveryCents"`
	Service         *int64      `json:"serviceCents"`
	Promotions      []Promotion `json:"promotions"`
}

// Reviewed branch mappings live with resolution, not provider extraction code.
//
//go:embed branch_matches.json
var rules embed.FS
var calorieSuffix = regexp.MustCompile(`(?i)\s*\[[\d.\-– ]+ Cals\]\s*$`)

func ItemKey(name string) string {
	name = strings.NewReplacer("®", "", "™", "", "’", "'", "‘", "'", "\u00a0", " ").Replace(name)
	return strings.Join(strings.Fields(strings.ToLower(calorieSuffix.ReplaceAllString(name, ""))), " ")
}
func Build(b Bundle) PublicCatalog {
	var matches map[string]struct {
		Canonical string `json:"canonicalStoreId"`
	}
	raw, _ := rules.ReadFile("branch_matches.json")
	_ = json.Unmarshal(raw, &matches)
	present := map[string]bool{}
	for _, s := range b.Providers {
		for _, st := range s.Stores {
			present[st.ID] = true
		}
	}
	result := PublicCatalog{City: "Fredericton", Restaurants: []*Restaurant{}}
	restaurants := map[string]*Restaurant{}
	for _, provider := range []string{"Uber Eats", "DoorDash"} {
		for _, st := range b.Providers[provider].Stores {
			canonical := st.ID
			if m, ok := matches[st.ID]; ok && present[m.Canonical] {
				canonical = m.Canonical
			}
			rid := "catalog-" + canonical
			r := restaurants[rid]
			if r == nil {
				r = &Restaurant{ID: rid, Name: st.Name, Address: st.Address, Rating: st.Rating, Image: st.Image, Categories: append([]string{}, st.Categories...), Items: []*Item{}, ArtworkKey: strings.TrimRight(st.URL, "/")}
				r.ArtworkKey = r.ArtworkKey[strings.LastIndex(r.ArtworkKey, "/")+1:]
				restaurants[rid] = r
				result.Restaurants = append(result.Restaurants, r)
			}
			if r.Address == "" {
				r.Address = st.Address
			}
			if r.Image == "" {
				r.Image = st.Image
			}
			r.Providers = append(r.Providers, provider)
			r.Sources = append(r.Sources, Source{provider, st.URL, st.Address, len(st.Items) > 0})
			// Ambiguous same-name variants are not combined into a made-up price.
			counts := map[string]int{}
			for _, it := range st.Items {
				counts[ItemKey(it.Name)]++
			}
			for _, it := range st.Items {
				key := ItemKey(it.Name)
				var item *Item
				if counts[key] == 1 {
					for _, existing := range r.Items {
						if existing.matchKey == key {
							item = existing
							break
						}
					}
				}
				if item == nil {
					id := it.Name
					if counts[key] > 1 {
						id = fmt.Sprintf("%s/%s/%d", provider, it.Name, len(r.Items))
					}
					matchKey := key
					if counts[key] > 1 {
						matchKey = ""
					}
					item = &Item{matchKey: matchKey, ID: id, Name: it.Name, Section: it.Section, Offers: []Offer{}}
					r.Items = append(r.Items, item)
				}
				kind := it.PriceKind
				if kind == "" {
					kind = "base"
				}
				mode := st.FulfillmentMode
				if mode == "" {
					mode = "unspecified"
				}
				promos := append([]Promotion{}, st.Promotions...)
				promos = append(promos, it.Promotions...)
				status := "priced"
				if it.Amount == nil {
					status = "options_required"
				}
				item.Offers = append(item.Offers, Offer{Provider: provider, Amount: it.Amount, Currency: it.Currency, PriceKind: kind, PriceNote: it.PriceNote, Status: status, SourceURL: st.URL, FulfillmentMode: mode, Promotions: promos})
			}
		}
	}
	for _, r := range result.Restaurants {
		if r.Address == "" {
			r.Address = "Fredericton · address unavailable"
		}
		for _, it := range r.Items {
			for _, src := range r.Sources {
				found := false
				for _, o := range it.Offers {
					if o.Provider == src.Provider {
						found = true
						break
					}
				}
				if !found {
					it.Offers = append(it.Offers, Offer{Provider: src.Provider, Currency: "CAD", Status: "unavailable", PriceKind: "options", PriceNote: "Price unavailable", SourceURL: src.URL, FulfillmentMode: "unspecified", Promotions: []Promotion{}})
				}
			}
			sort.SliceStable(it.Offers, func(i, j int) bool { return it.Offers[i].Provider < it.Offers[j].Provider })
		}
	}
	return result
}
