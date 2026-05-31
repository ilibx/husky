package vectorstore

import (
	"context"
	"fmt"
	"math"
)

// neo4jVectorStore implements vector operations using Neo4j's vector index.
// It operates on the assumption that Neo4j nodes have labels and properties
// that mirror the knowledge base: (:Knowledge {id, title, content, category, tags, vector})
type neo4jVectorStore struct {
	uri      string
	user     string
	password string
}

func NewNeo4jVectorStore(uri, user, password string) Store {
	return &neo4jVectorStore{
		uri:      uri,
		user:     user,
		password: password,
	}
}

func (s *neo4jVectorStore) SearchSimilar(ctx context.Context, queryVector []float32, limit int) ([]SearchResult, error) {
	if len(queryVector) == 0 {
		return nil, nil
	}

	// NOTE: For a real Neo4j deployment, this would use the Neo4j Go driver
	// with a Cypher query leveraging the vector index:
	//
	//   CALL db.index.vector.queryNodes('knowledge-vector-index', $limit, $vector)
	//   YIELD node AS n, score
	//   RETURN n.id AS id, score
	//   ORDER BY score ASC
	//
	// Since the neo4j/go-neo4j driver is not yet in go.mod, this implementation
	// provides the structural pattern. To use it, add the driver dependency:
	//   go get github.com/neo4j/neo4j-go-driver/v5/neo4j
	//
	// For now, this returns a placeholder to indicate the integration point.

	return nil, fmt.Errorf("neo4j driver not available; need github.com/neo4j/neo4j-go-driver/v5/neo4j")
}

func (s *neo4jVectorStore) InsertVector(ctx context.Context, id string, vector []float32) error {
	// Cypher equivalent for a real driver:
	//
	// MATCH (n:Knowledge {id: $id})
	// SET n.vector = $vector
	// RETURN n

	return fmt.Errorf("neo4j driver not available; need github.com/neo4j/neo4j-go-driver/v5/neo4j")
}

func (s *neo4jVectorStore) DeleteVector(ctx context.Context, id string) error {
	// Cypher: MATCH (n:Knowledge {id: $id}) REMOVE n.vector
	return fmt.Errorf("neo4j driver not available; need github.com/neo4j/neo4j-go-driver/v5/neo4j")
}

func (s *neo4jVectorStore) Close() error { return nil }

func (s *neo4jVectorStore) UpdateVector(ctx context.Context, id string, vector []float32) error {
	return s.InsertVector(ctx, id, vector)
}

// cosineSimilarity computes cosine similarity between two vectors for client-side scoring.
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return float32(dot / (math.Sqrt(normA) * math.Sqrt(normB)))
}
