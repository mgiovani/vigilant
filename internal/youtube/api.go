package youtube

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type VideoResult struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type Service struct {
	apiKey string
	client *youtube.Service
}

// NewService creates a new YouTube service with the given API key
func NewService(apiKey string) (*Service, error) {
	ctx := context.Background()
	client, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	return &Service{
		apiKey: apiKey,
		client: client,
	}, nil
}

// SearchLofiStreams searches for embeddable lofi hip hop streams
func (s *Service) SearchLofiStreams(ctx context.Context) ([]VideoResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// Direct search for lofi hip hop videos
	call := s.client.Search.List([]string{"snippet"}).
		Q("lofi hip hop radio beats to relax").
		Type("video").
		Order("relevance").
		MaxResults(10).
		Context(ctx)

	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to search YouTube: %w", err)
	}

	if len(response.Items) == 0 {
		return nil, fmt.Errorf("no lofi streams found")
	}

	var results []VideoResult
	for _, item := range response.Items {
		if item.Id.VideoId != "" {
			results = append(results, VideoResult{
				ID:    item.Id.VideoId,
				Title: item.Snippet.Title,
			})
		}
	}

	return results, nil
}

// GetLofiVideoID returns a working lofi video ID, falling back to default if needed
func (s *Service) GetLofiVideoID(ctx context.Context) (string, error) {
	results, err := s.SearchLofiStreams(ctx)
	if err != nil || len(results) == 0 {
		// Fallback to the original hardcoded ID
		return "jfKfPfyJRdk", nil
	}
	return results[0].ID, nil
}

// FromEnv creates a YouTube service from environment variables
func FromEnv() (*Service, error) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("YOUTUBE_API_KEY environment variable not set")
	}
	return NewService(apiKey)
}
