package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bitik/backend/internal/config"
	opensearch "github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

func Check(ctx context.Context, cfg config.SearchConfig) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.URL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	return nil
}

type Client struct {
	os            *opensearch.Client
	productsIndex string
}

func NewClient(cfg config.SearchConfig) (*Client, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return nil, fmt.Errorf("search url is empty")
	}
	endpoint := cfg.URL
	if !strings.Contains(endpoint, "://") {
		endpoint = "http://" + endpoint
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	osCfg := opensearch.Config{
		Addresses: []string{u.String()},
	}
	if cfg.Username != "" || cfg.Password != "" {
		osCfg.Username = cfg.Username
		osCfg.Password = cfg.Password
	}
	client, err := opensearch.NewClient(osCfg)
	if err != nil {
		return nil, err
	}
	index := strings.TrimSpace(cfg.ProductsIndex)
	if index == "" {
		index = "products_v1"
	}
	return &Client{os: client, productsIndex: index}, nil
}

func (c *Client) ProductsIndex() string { return c.productsIndex }

// EnsureProductsIndex creates the products index with mapping if missing.
func (c *Client) EnsureProductsIndex(ctx context.Context) error {
	// HEAD /{index}
	head := opensearchapi.IndicesExistsReq{Indices: []string{c.productsIndex}}
	resp, err := c.os.Do(ctx, head, nil)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	if resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("indices exists returned status %d", resp.StatusCode)
	}

	body := map[string]any{
		"settings": map[string]any{
			"index": map[string]any{
				"number_of_shards":   1,
				"number_of_replicas": 0,
			},
			"analysis": map[string]any{
				"analyzer": map[string]any{
					"folding": map[string]any{
						"type":      "custom",
						"tokenizer": "standard",
						"filter":    []string{"lowercase", "asciifolding"},
					},
				},
			},
		},
		"mappings": map[string]any{
			"dynamic": "false",
			"properties": map[string]any{
				"id":              map[string]any{"type": "keyword"},
				"seller_id":       map[string]any{"type": "keyword"},
				"category_id":     map[string]any{"type": "keyword"},
				"brand_id":        map[string]any{"type": "keyword"},
				"name":            map[string]any{"type": "text", "analyzer": "folding"},
				"slug":            map[string]any{"type": "keyword"},
				"description":     map[string]any{"type": "text", "analyzer": "folding"},
				"min_price_cents": map[string]any{"type": "long"},
				"max_price_cents": map[string]any{"type": "long"},
				"currency":        map[string]any{"type": "keyword"},
				"total_sold":      map[string]any{"type": "long"},
				"rating":          map[string]any{"type": "double"},
				"review_count":    map[string]any{"type": "long"},
				"published_at":    map[string]any{"type": "date"},
				"updated_at":      map[string]any{"type": "date"},
				"image_url":       map[string]any{"type": "keyword"},
			},
		},
	}
	raw, _ := json.Marshal(body)
	req := opensearchapi.IndicesCreateReq{
		Index: c.productsIndex,
		Body:  strings.NewReader(string(raw)),
	}
	createResp, err := c.os.Do(ctx, req, nil)
	if err != nil {
		return err
	}
	defer createResp.Body.Close()
	if createResp.StatusCode >= http.StatusBadRequest {
		b, _ := io.ReadAll(createResp.Body)
		return fmt.Errorf("create index failed: status=%d body=%s", createResp.StatusCode, strings.TrimSpace(string(b)))
	}
	return nil
}

type ProductDocument struct {
	ID            string     `json:"id"`
	SellerID      string     `json:"seller_id"`
	CategoryID    string     `json:"category_id,omitempty"`
	BrandID       string     `json:"brand_id,omitempty"`
	Name          string     `json:"name"`
	Slug          string     `json:"slug"`
	Description   string     `json:"description,omitempty"`
	MinPriceCents int64      `json:"min_price_cents"`
	MaxPriceCents int64      `json:"max_price_cents"`
	Currency      string     `json:"currency"`
	TotalSold     int64      `json:"total_sold"`
	Rating        float64    `json:"rating"`
	ReviewCount   int64      `json:"review_count"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
	UpdatedAt     time.Time  `json:"updated_at"`
	ImageURL      string     `json:"image_url,omitempty"`
}

func (c *Client) IndexProduct(ctx context.Context, doc ProductDocument) error {
	if err := c.EnsureProductsIndex(ctx); err != nil {
		return err
	}
	raw, _ := json.Marshal(doc)
	req := opensearchapi.IndexReq{
		Index:      c.productsIndex,
		DocumentID: doc.ID,
		Body:       strings.NewReader(string(raw)),
		Params:     opensearchapi.IndexParams{Refresh: "false"},
	}
	resp, err := c.os.Do(ctx, req, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("index product failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return nil
}

type SearchRequest struct {
	Query         string
	CategoryID    string
	BrandID       string
	SellerID      string
	MinPriceCents *int64
	MaxPriceCents *int64
	Sort          string
	From          int
	Size          int
}

type SearchHit struct {
	ID            string  `json:"id"`
	SellerID      string  `json:"seller_id"`
	CategoryID    string  `json:"category_id"`
	BrandID       string  `json:"brand_id"`
	Name          string  `json:"name"`
	Slug          string  `json:"slug"`
	MinPriceCents int64   `json:"min_price_cents"`
	MaxPriceCents int64   `json:"max_price_cents"`
	Currency      string  `json:"currency"`
	TotalSold     int64   `json:"total_sold"`
	Rating        float64 `json:"rating"`
	ReviewCount   int64   `json:"review_count"`
	ImageURL      string  `json:"image_url"`
}

type SearchResponse struct {
	Total int64
	Hits  []SearchHit
}

func (c *Client) SearchProducts(ctx context.Context, req SearchRequest) (SearchResponse, error) {
	if err := c.EnsureProductsIndex(ctx); err != nil {
		return SearchResponse{}, err
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	if req.Size > 100 {
		req.Size = 100
	}
	if req.From < 0 {
		req.From = 0
	}

	filters := make([]any, 0, 8)
	if strings.TrimSpace(req.CategoryID) != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"category_id": req.CategoryID}})
	}
	if strings.TrimSpace(req.BrandID) != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"brand_id": req.BrandID}})
	}
	if strings.TrimSpace(req.SellerID) != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"seller_id": req.SellerID}})
	}
	if req.MinPriceCents != nil || req.MaxPriceCents != nil {
		rng := map[string]any{}
		if req.MinPriceCents != nil {
			rng["gte"] = *req.MinPriceCents
		}
		if req.MaxPriceCents != nil {
			rng["lte"] = *req.MaxPriceCents
		}
		filters = append(filters, map[string]any{"range": map[string]any{"min_price_cents": rng}})
	}

	query := map[string]any{"match_all": map[string]any{}}
	q := strings.TrimSpace(req.Query)
	if q != "" {
		query = map[string]any{
			"multi_match": map[string]any{
				"query":  q,
				"fields": []string{"name^3", "description"},
				"type":   "best_fields",
			},
		}
	}

	boolQ := map[string]any{"must": query}
	if len(filters) > 0 {
		boolQ = map[string]any{"must": query, "filter": filters}
	}

	sort := []any{}
	switch req.Sort {
	case "price_asc":
		sort = append(sort, map[string]any{"min_price_cents": map[string]any{"order": "asc"}})
	case "price_desc":
		sort = append(sort, map[string]any{"min_price_cents": map[string]any{"order": "desc"}})
	case "popular":
		sort = append(sort, map[string]any{"total_sold": map[string]any{"order": "desc"}})
	case "latest", "newest":
		sort = append(sort, map[string]any{"published_at": map[string]any{"order": "desc", "missing": "_last"}})
	case "rating":
		sort = append(sort, map[string]any{"rating": map[string]any{"order": "desc"}})
	default:
		sort = append(sort, map[string]any{"_score": map[string]any{"order": "desc"}})
	}
	sort = append(sort, map[string]any{"updated_at": map[string]any{"order": "desc"}})

	body := map[string]any{
		"from": req.From,
		"size": req.Size,
		"query": map[string]any{
			"bool": boolQ,
		},
		"sort": sort,
		"_source": []string{
			"id", "seller_id", "category_id", "brand_id", "name", "slug",
			"min_price_cents", "max_price_cents", "currency", "total_sold",
			"rating", "review_count", "image_url",
		},
	}
	raw, _ := json.Marshal(body)

	searchReq := opensearchapi.SearchReq{
		Indices: []string{c.productsIndex},
		Body:  strings.NewReader(string(raw)),
	}
	resp, err := c.os.Do(ctx, searchReq, nil)
	if err != nil {
		return SearchResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		return SearchResponse{}, fmt.Errorf("opensearch search failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	var parsed struct {
		Hits struct {
			Total any `json:"total"`
			Hits  []struct {
				Source SearchHit `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return SearchResponse{}, err
	}
	total := int64(0)
	switch v := parsed.Hits.Total.(type) {
	case map[string]any:
		if n, ok := v["value"].(float64); ok {
			total = int64(n)
		}
	case float64:
		total = int64(v)
	}
	out := make([]SearchHit, 0, len(parsed.Hits.Hits))
	for _, h := range parsed.Hits.Hits {
		out = append(out, h.Source)
	}
	return SearchResponse{Total: total, Hits: out}, nil
}
