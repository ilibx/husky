package vectorstore

import (
	"fmt"

	"gorm.io/gorm"
)

type Config struct {
	Provider string
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// NewFromConfig creates a Store based on the provider name in the config.
// Supported providers:
//
//	"postgresql" (default) — uses pgvector via existing GORM connection
//	"neo4j" — uses Neo4j vector index (requires driver dependency)
func NewFromConfig(cfg Config, db *gorm.DB) (Store, error) {
	switch cfg.Provider {
	case "", "postgresql":
		if db == nil {
			return nil, fmt.Errorf("pgvector store requires a non-nil *gorm.DB")
		}
		return NewPGVectorStore(db), nil

	case "neo4j":
		uri := fmt.Sprintf("bolt://%s:%s", cfg.Host, cfg.Port)
		return NewNeo4jVectorStore(uri, cfg.User, cfg.Password), nil

	default:
		return nil, fmt.Errorf("unsupported vector store provider: %s", cfg.Provider)
	}
}
