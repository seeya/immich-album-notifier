package immich

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"
)

type ImmichClient struct {
	APIKey string
	Host   string
	Client *http.Client
}

type Album struct {
	AlbumName                  string        `json:"albumName"`
	Description                string        `json:"description"`
	AlbumThumbnailAssetID      string        `json:"albumThumbnailAssetId"`
	CreatedAt                  time.Time     `json:"createdAt"`
	UpdatedAt                  time.Time     `json:"updatedAt"`
	ID                         string        `json:"id"`
	OwnerID                    string        `json:"ownerId"`
	AlbumUsers                 []interface{} `json:"albumUsers"` // Replace interface{} with actual type if known
	Shared                     bool          `json:"shared"`
	HasSharedLink              bool          `json:"hasSharedLink"`
	StartDate                  time.Time     `json:"startDate"`
	EndDate                    time.Time     `json:"endDate"`
	Assets                     []interface{} `json:"assets"` // Replace interface{} with actual type if known
	AssetCount                 int           `json:"assetCount"`
	IsActivityEnabled          bool          `json:"isActivityEnabled"`
	Order                      string        `json:"order"`
	LastModifiedAssetTimestamp time.Time     `json:"lastModifiedAssetTimestamp"`
}

type CustomTransport struct {
	Base      http.RoundTripper
	APIKey    string
	HeaderKey string
}

func (t *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clonedReq := req.Clone(req.Context())
	clonedReq.Header.Set(t.HeaderKey, t.APIKey)
	return t.Base.RoundTrip(clonedReq)
}

var client *ImmichClient

func InitImmichClient(apiKey string, host string) *ImmichClient {
	baseTransport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	customTransport := &CustomTransport{
		Base:      baseTransport,
		APIKey:    apiKey,
		HeaderKey: "x-api-key",
	}

	client = &ImmichClient{
		APIKey: apiKey,
		Host:   host,
		Client: &http.Client{
			Timeout:   10 * time.Second, // Total timeout including connection + redirects + response read
			Transport: customTransport,
		},
	}

	return client
}

func (c *ImmichClient) GetAllAlbums() *[]Album {
	endpoint := fmt.Sprintf("%s/albums", c.Host)
	resp, err := c.Client.Get(endpoint)
	if err != nil {
		slog.Error(err.Error())
		return nil
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		slog.Error(err.Error())
		return nil
	}

	var results []Album

	err = json.Unmarshal(body, &results)
	if err != nil {
		slog.Error(err.Error())
		return nil
	}

	// for _, a := range results {
	// 	slog.Debug("albums", slog.String("name", a.AlbumName), slog.Int("total", a.AssetCount))
	// }

	// cache the results
	err = os.WriteFile("./cache.json", body, 0600)
	if err != nil {
		slog.Error(err.Error())
		return nil
	}

	return &results
}

func (c *ImmichClient) ReadCached() []Album {
	cachedBytes, err := os.ReadFile("./cache.json")

	if err != nil {
		slog.Error(err.Error())
		return nil
	}

	var cachedJson []Album

	err = json.Unmarshal(cachedBytes, &cachedJson)

	if err != nil {
		slog.Error(err.Error())
		return nil
	}

	return cachedJson
}
