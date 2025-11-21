package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lunartools/lunar-go-sdk/src/types"
)

type Client struct {
	clientID    string
	accessToken string
	baseURL     string
	httpClient  *http.Client
}

func NewClient(config types.Config) *Client {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://www.lunartools.co"
	}

	return &Client{
		clientID:    config.ClientID,
		accessToken: config.AccessToken,
		baseURL:     baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) AddProduct(product types.AddProduct) error {
	if strings.TrimSpace(product.Name) == "" {
		return fmt.Errorf("product name is required")
	}
	if strings.TrimSpace(product.SKU) == "" {
		return fmt.Errorf("product SKU is required")
	}
	if product.Qty < 0 {
		return fmt.Errorf("product quantity must be a non-negative number")
	}
	if product.Value != nil && *product.Value < 0 {
		return fmt.Errorf("product value must be a non-negative number")
	}
	if product.Spent != nil && *product.Spent < 0 {
		return fmt.Errorf("product spent must be a non-negative number")
	}

	payload := map[string]interface{}{
		"clientId":    c.clientID,
		"accessToken": c.accessToken,
		"name":        product.Name,
		"sku":         product.SKU,
		"qty":         product.Qty,
	}

	if product.Size != nil {
		payload["size"] = *product.Size
	}
	if product.Store != nil {
		payload["store"] = *product.Store
	}
	if product.Value != nil {
		payload["value"] = *product.Value
	}
	if product.Spent != nil {
		payload["spent"] = *product.Spent
	}

	return c.makeRequest("POST", "/sdk/add-product", payload, nil)
}

func (c *Client) AddOrder(order types.AddOrder) error {
	if strings.TrimSpace(order.Name) == "" {
		return fmt.Errorf("order name is required")
	}
	if strings.TrimSpace(order.Status) == "" {
		return fmt.Errorf("order status is required")
	}
	if strings.TrimSpace(order.OrderNumber) == "" {
		return fmt.Errorf("order number is required")
	}

	payload := map[string]interface{}{
		"clientId":    c.clientID,
		"accessToken": c.accessToken,
		"name":        order.Name,
		"status":      order.Status,
		"orderNumber": order.OrderNumber,
	}

	if order.Image != nil {
		payload["image"] = *order.Image
	}
	if order.Tracking != nil {
		payload["tracking"] = *order.Tracking
	}
	if order.Date != nil {
		payload["date"] = *order.Date
	}
	if order.Qty != nil {
		payload["qty"] = *order.Qty
	}
	if order.Price != nil {
		payload["price"] = *order.Price
	}
	if order.OrderTotal != nil {
		payload["orderTotal"] = *order.OrderTotal
	}
	if order.Account != nil {
		payload["account"] = *order.Account
	}
	if order.Retailer != nil {
		payload["retailer"] = *order.Retailer
	}
	if order.Tags != nil {
		payload["tags"] = *order.Tags
	}

	return c.makeRequest("POST", "/sdk/add-order", payload, nil)
}

func (c *Client) Webhook(webhookURL string, payload types.Webhook) (*types.WebhookResponse, error) {
	hasContent := payload.Content != nil && strings.TrimSpace(*payload.Content) != ""
	hasEmbeds := len(payload.Embeds) > 0

	if !hasContent && !hasEmbeds {
		return nil, fmt.Errorf("webhook payload must contain either content or at least one embed")
	}
	if len(payload.Embeds) > 10 {
		return nil, fmt.Errorf("discord webhooks support a maximum of 10 embeds")
	}

	for i, embed := range payload.Embeds {
		if len(embed.Fields) > 25 {
			return nil, fmt.Errorf("embed %d exceeds the maximum of 25 fields", i)
		}
		for j, field := range embed.Fields {
			if strings.TrimSpace(field.Name) == "" {
				return nil, fmt.Errorf("embed %d, field %d: name is required", i, j)
			}
			if strings.TrimSpace(field.Value) == "" {
				return nil, fmt.Errorf("embed %d, field %d: value is required", i, j)
			}
		}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("webhook request failed with status: %d", resp.StatusCode)
	}

	var webhookResp types.WebhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&webhookResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &webhookResp, nil
}

func (c *Client) makeRequest(method, path string, payload interface{}, result interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client-ID", c.clientID)
	req.Header.Set("X-Access-Token", c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
