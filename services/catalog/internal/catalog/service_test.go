package catalog

import (
	"context"
	catalogv1 "funtime.local/contracts/catalog/v1"
	"testing"
)

func TestCatalogHasUniqueIDsAndIsNotMutableAcrossCalls(t *testing.T) {
	service := &Service{}
	result, err := service.ListGames(context.Background(), &catalogv1.ListGamesRequest{})
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, game := range result.Games {
		if game.Id == "" || seen[game.Id] || game.Title == "" {
			t.Fatalf("invalid catalog entry: %v", game)
		}
		seen[game.Id] = true
	}
	if len(result.Games) != 1 || result.Games[0].Status != "available" {
		t.Fatal("world domination must be available")
	}
	result.Games[0].Title = "changed"
	fresh, _ := service.ListGames(context.Background(), &catalogv1.ListGamesRequest{})
	if fresh.Games[0].Title == "changed" {
		t.Fatal("catalog mutation leaked between requests")
	}
}
