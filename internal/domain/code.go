package domain

import "time"

type Authority string

const (
	AuthorityCommunity Authority = "community"
	AuthorityOfficial  Authority = "official"
)

type Status string

const (
	StatusActive  Status = "active"
	StatusExpired Status = "expired"
	StatusUnknown Status = "unknown"
)

type SourceAttribution struct {
	SourceID   string     `json:"sourceId" bson:"sourceId"`
	SourceURL  string     `json:"sourceUrl" bson:"sourceUrl"`
	Authority  Authority  `json:"authority" bson:"authority"`
	Status     Status     `json:"status,omitempty" bson:"status,omitempty"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty" bson:"expiresAt,omitempty"`
	FirstSeen  time.Time  `json:"firstSeenAt" bson:"firstSeenAt"`
	LastSeen   time.Time  `json:"lastSeenAt" bson:"lastSeenAt"`
	RevisionID int64      `json:"revisionId,omitempty" bson:"revisionId,omitempty"`
}

type CodeCandidate struct {
	GameSlug   string
	Code       string
	Rewards    []string
	Region     string
	Status     Status
	ExpiresAt  *time.Time
	SourceID   string
	SourceURL  string
	Authority  Authority
	ObservedAt time.Time
	RevisionID int64
}

type Code struct {
	GameSlug      string              `json:"gameSlug" bson:"gameSlug"`
	Code          string              `json:"code" bson:"code"`
	CanonicalCode string              `json:"canonicalCode" bson:"canonicalCode"`
	Rewards       []string            `json:"rewards" bson:"rewards"`
	Region        string              `json:"region" bson:"region"`
	Status        Status              `json:"status" bson:"status"`
	Sources       []SourceAttribution `json:"sources" bson:"sources"`
	ExpiresAt     *time.Time          `json:"expiresAt" bson:"expiresAt"`
}
