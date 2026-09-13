package catalog

import (
	"context"

	catalogv1 "funtime.local/contracts/catalog/v1"
)

type Service struct {
	catalogv1.UnimplementedCatalogServiceServer
}

func (s *Service) ListGames(context.Context, *catalogv1.ListGamesRequest) (*catalogv1.ListGamesResponse, error) {
	// A fresh response prevents callers from mutating shared catalog data.
	return &catalogv1.ListGamesResponse{Games: []*catalogv1.Game{{
		Id:          "world-domination",
		Title:       "Мировое господство",
		Description: "Одна планета. Ваша компания. Большие амбиции. Первая игра FunTime — скоро здесь.",
		Category:    "Игра для компании",
		Status:      "coming_soon",
		Accent:      "lime",
	}}}, nil
}
