package mongodb

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"opengachacodes/internal/domain"
)

type Repository struct {
	client *mongo.Client
	games  *mongo.Collection
	codes  *mongo.Collection
}

func Connect(ctx context.Context, uri, database string) (*Repository, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect MongoDB: %w", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("ping MongoDB: %w", err)
	}
	db := client.Database(database)
	return &Repository{client: client, games: db.Collection("games"), codes: db.Collection("codes")}, nil
}

func (r *Repository) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}

func (r *Repository) EnsureIndexes(ctx context.Context) error {
	if _, err := r.games.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "slug", Value: 1}}, Options: options.Index().SetUnique(true),
	}); err != nil {
		return fmt.Errorf("create games index: %w", err)
	}
	if _, err := r.codes.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "gameSlug", Value: 1}, {Key: "canonicalCode", Value: 1}, {Key: "region", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "gameSlug", Value: 1}, {Key: "status", Value: 1}, {Key: "expiresAt", Value: 1}}},
	}); err != nil {
		return fmt.Errorf("create codes indexes: %w", err)
	}
	return nil
}

func (r *Repository) RenameGameSlug(ctx context.Context, oldSlug, newSlug string) error {
	if oldSlug == newSlug {
		return nil
	}

	cursor, err := r.codes.Find(ctx, bson.M{"gameSlug": oldSlug})
	if err != nil {
		return fmt.Errorf("list codes for game slug %s: %w", oldSlug, err)
	}
	defer cursor.Close(ctx)

	var oldCodes []domain.Code
	if err := cursor.All(ctx, &oldCodes); err != nil {
		return fmt.Errorf("decode codes for game slug %s: %w", oldSlug, err)
	}
	for _, code := range oldCodes {
		oldFilter := bson.M{"gameSlug": oldSlug, "canonicalCode": code.CanonicalCode, "region": code.Region}
		newFilter := bson.M{"gameSlug": newSlug, "canonicalCode": code.CanonicalCode, "region": code.Region}

		var existing domain.Code
		err := r.codes.FindOne(ctx, newFilter).Decode(&existing)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("find migrated code %s: %w", code.CanonicalCode, err)
		}
		code.GameSlug = newSlug
		if err == nil {
			code = mergeCode(existing, code)
		}
		if _, err := r.codes.ReplaceOne(ctx, newFilter, code, options.Replace().SetUpsert(true)); err != nil {
			return fmt.Errorf("migrate code %s from %s to %s: %w", code.CanonicalCode, oldSlug, newSlug, err)
		}
		if _, err := r.codes.DeleteOne(ctx, oldFilter); err != nil {
			return fmt.Errorf("delete old code %s for %s: %w", code.CanonicalCode, oldSlug, err)
		}
	}

	targetExists, err := r.GameExists(ctx, newSlug)
	if err != nil {
		return err
	}
	if targetExists {
		if _, err := r.games.DeleteMany(ctx, bson.M{"slug": oldSlug}); err != nil {
			return fmt.Errorf("delete old game slug %s: %w", oldSlug, err)
		}
		return nil
	}
	if _, err := r.games.UpdateMany(ctx, bson.M{"slug": oldSlug}, bson.M{"$set": bson.M{"slug": newSlug}}); err != nil {
		return fmt.Errorf("rename game slug %s to %s: %w", oldSlug, newSlug, err)
	}
	return nil
}

func (r *Repository) EnsureGames(ctx context.Context, games []domain.Game) error {
	for _, game := range games {
		_, err := r.games.UpdateOne(ctx, bson.M{"slug": game.Slug}, bson.M{"$set": game}, options.UpdateOne().SetUpsert(true))
		if err != nil {
			return fmt.Errorf("upsert game %s: %w", game.Slug, err)
		}
	}
	return nil
}

func (r *Repository) ListGames(ctx context.Context) ([]domain.Game, error) {
	cursor, err := r.games.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "slug", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("list games: %w", err)
	}
	defer cursor.Close(ctx)
	games := make([]domain.Game, 0)
	if err := cursor.All(ctx, &games); err != nil {
		return nil, fmt.Errorf("decode games: %w", err)
	}
	return games, nil
}

func (r *Repository) GameExists(ctx context.Context, slug string) (bool, error) {
	err := r.games.FindOne(ctx, bson.M{"slug": slug}).Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("find game: %w", err)
	}
	return true, nil
}

func (r *Repository) UpsertCodes(ctx context.Context, codes []domain.Code) error {
	for _, incoming := range codes {
		filter := bson.M{"gameSlug": incoming.GameSlug, "canonicalCode": incoming.CanonicalCode, "region": incoming.Region}
		var existing domain.Code
		err := r.codes.FindOne(ctx, filter).Decode(&existing)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("find code %s: %w", incoming.CanonicalCode, err)
		}
		if err == nil {
			incoming = mergeCode(existing, incoming)
		}
		_, err = r.codes.ReplaceOne(ctx, filter, incoming, options.Replace().SetUpsert(true))
		if err != nil {
			return fmt.Errorf("upsert code %s: %w", incoming.CanonicalCode, err)
		}
	}
	return nil
}

func (r *Repository) DeleteCodes(ctx context.Context, gameSlug string, canonicalCodes []string) error {
	if len(canonicalCodes) == 0 {
		return nil
	}
	_, err := r.codes.DeleteMany(ctx, bson.M{
		"gameSlug":      gameSlug,
		"canonicalCode": bson.M{"$in": canonicalCodes},
	})
	if err != nil {
		return fmt.Errorf("delete excluded codes: %w", err)
	}
	return nil
}

func (r *Repository) ListActiveCodes(ctx context.Context, gameSlug string, now time.Time) ([]domain.Code, error) {
	filter := bson.M{
		"gameSlug": gameSlug,
		"status":   domain.StatusActive,
		"$or": bson.A{
			bson.M{"expiresAt": nil},
			bson.M{"expiresAt": bson.M{"$exists": false}},
			bson.M{"expiresAt": bson.M{"$gt": now}},
		},
	}
	cursor, err := r.codes.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "canonicalCode", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("list active codes: %w", err)
	}
	defer cursor.Close(ctx)
	codes := make([]domain.Code, 0)
	if err := cursor.All(ctx, &codes); err != nil {
		return nil, fmt.Errorf("decode active codes: %w", err)
	}
	return codes, nil
}

func mergeCode(existing, incoming domain.Code) domain.Code {
	incoming.Sources = mergeSources(existing.Sources, incoming.Sources)
	incoming.Rewards = mergeRewards(existing.Rewards, incoming.Rewards)

	selected := domain.SourceAttribution{}
	selectedRank := 0
	for _, source := range incoming.Sources {
		if source.Status == "" || source.Status == domain.StatusUnknown {
			continue
		}
		rank := 1
		if source.Authority == domain.AuthorityOfficial {
			rank = 2
		}
		if rank > selectedRank || rank == selectedRank && source.LastSeen.After(selected.LastSeen) {
			selected = source
			selectedRank = rank
		}
	}
	if selectedRank > 0 {
		incoming.Status = selected.Status
		incoming.ExpiresAt = selected.ExpiresAt
	} else if existing.Status != "" {
		incoming.Status = existing.Status
		incoming.ExpiresAt = existing.ExpiresAt
	}
	return incoming
}

func mergeRewards(existing, incoming []string) []string {
	seen := make(map[string]bool, len(existing)+len(incoming))
	result := make([]string, 0, len(existing)+len(incoming))
	for _, reward := range incoming {
		if reward != "" && !seen[reward] {
			result = append(result, reward)
			seen[reward] = true
		}
	}
	for _, reward := range existing {
		if reward == "" || seen[reward] || containsMultipleRewards(reward, incoming) {
			continue
		}
		result = append(result, reward)
		seen[reward] = true
	}
	sort.Strings(result)
	return result
}

func containsMultipleRewards(combined string, rewards []string) bool {
	matches := 0
	for _, reward := range rewards {
		if reward != "" && strings.Contains(combined, reward) {
			matches++
		}
	}
	return matches >= 2
}

func mergeSources(existing, incoming []domain.SourceAttribution) []domain.SourceAttribution {
	merged := make(map[string]domain.SourceAttribution, len(existing)+len(incoming))
	for _, source := range existing {
		merged[source.SourceID] = source
	}
	for _, source := range incoming {
		if current, ok := merged[source.SourceID]; ok {
			if current.FirstSeen.Before(source.FirstSeen) {
				source.FirstSeen = current.FirstSeen
			}
			if source.Status == "" || source.Status == domain.StatusUnknown {
				source.Status = current.Status
			}
			if source.ExpiresAt == nil {
				source.ExpiresAt = current.ExpiresAt
			}
			if current.LastSeen.After(source.LastSeen) {
				source.LastSeen = current.LastSeen
				source.RevisionID = current.RevisionID
				source.Status = current.Status
				source.ExpiresAt = current.ExpiresAt
			}
		}
		merged[source.SourceID] = source
	}
	result := make([]domain.SourceAttribution, 0, len(merged))
	for _, source := range merged {
		result = append(result, source)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].SourceID < result[j].SourceID })
	return result
}
