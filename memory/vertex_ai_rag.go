// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	aiplatform "cloud.google.com/go/aiplatform/apiv1beta1"
	"cloud.google.com/go/aiplatform/apiv1beta1/aiplatformpb"
	"cloud.google.com/go/auth/credentials"
	"cloud.google.com/go/storage"
	"github.com/go-json-experiment/json"
	"google.golang.org/api/option"
	"google.golang.org/genai"

	"github.com/go-a2a/adk-go/types"
)

// VertexAIRagService implements Service with Google Cloud Vertex AI RAG.
type VertexAIRagService struct {
	vertexRagClient         *aiplatform.VertexRagClient
	vertexRagDataClient     *aiplatform.VertexRagDataClient
	ragCorpus               string
	similarityTopK          int
	vectorDistanceThreshold float64
	vertexRAGStore          *genai.VertexRAGStore
	logger                  *slog.Logger
}

var _ types.MemoryService = (*VertexAIRagService)(nil)

// VertexAIRagOption is a functional option for configuring [VertexAIRagService].
type VertexAIRagOption func(*VertexAIRagService)

// WithVertexAIRagLogger sets the logger for the [VertexAIRagService].
func WithVertexAIRagLogger(logger *slog.Logger) VertexAIRagOption {
	return func(s *VertexAIRagService) {
		s.logger = logger
	}
}

// WithSimilarityTopK sets the number of top results to return for the [VertexAIRagService].
func WithSimilarityTopK(topK int) VertexAIRagOption {
	return func(s *VertexAIRagService) {
		s.similarityTopK = topK
	}
}

// WithVectorDistanceThreshold sets the threshold for vector similarity for the [VertexAIRagService].
func WithVectorDistanceThreshold(threshold float64) VertexAIRagOption {
	return func(s *VertexAIRagService) {
		s.vectorDistanceThreshold = threshold
	}
}

// NewVertexAIRagService creates a new VertexAIRagService.
func NewVertexAIRagService(ctx context.Context, ragCorpus string, opts ...VertexAIRagOption) (*VertexAIRagService, error) {
	creds, err := credentials.DetectDefault(&credentials.DetectOptions{
		Scopes: []string{
			storage.ScopeFullControl,
			storage.ScopeReadWrite,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get credentials for storage: %w", err)
	}

	vertexRagClient, err := aiplatform.NewVertexRagClient(ctx, option.WithAuthCredentials(creds))
	if err != nil {
		return nil, err
	}

	vertexRagDataClient, err := aiplatform.NewVertexRagDataClient(ctx, option.WithAuthCredentials(creds))
	if err != nil {
		return nil, err
	}

	s := &VertexAIRagService{
		vertexRagClient:         vertexRagClient,
		vertexRagDataClient:     vertexRagDataClient,
		ragCorpus:               ragCorpus,
		similarityTopK:          5,   // Default value
		vectorDistanceThreshold: 0.7, // Default value
		logger:                  slog.Default(),
	}
	for _, opt := range opts {
		opt(s)
	}

	vertexGagStore := &genai.VertexRAGStore{
		RAGResources: []*genai.VertexRAGStoreRAGResource{
			{
				RAGCorpus: ragCorpus,
			},
		},
		SimilarityTopK:          genai.Ptr(int32(s.similarityTopK)),
		VectorDistanceThreshold: genai.Ptr(s.vectorDistanceThreshold),
	}
	s.vertexRAGStore = vertexGagStore

	return s, nil
}

// AddSessionToMemory implements [types.MemoryService].
//
// TODO(zchee): implements
func (s *VertexAIRagService) AddSessionToMemory(ctx context.Context, session types.Session) error {
	return errors.New("not implemented: Vertex AI RAG integration requires additional dependencies")

	if len(s.vertexRAGStore.RAGResources) == 0 {
		return errors.New("rag resources must be set")
	}

	s.logger.InfoContext(ctx, "Adding session to Vertex AI RAG memory",
		slog.String("app_name", session.AppName()),
		slog.String("user_id", session.UserID()),
		slog.String("session_id", session.ID()),
		slog.String("rag_corpus", s.ragCorpus),
	)

	tempfile, err := os.CreateTemp(os.TempDir(), "*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(tempfile.Name())

	outputLines := []string{}
	for _, event := range session.Events() {
		if event.Content == nil || len(event.Content.Parts) == 0 {
			continue
		}
		textParts := make([]string, 0, len(event.Content.Parts))
		for _, part := range event.Content.Parts {
			if part.Text != "" {
				text := strings.ReplaceAll(part.Text, "\n", " ")
				textParts = append(textParts, text)
			}
		}

		if len(textParts) > 0 {
			m := map[string]any{
				"author":    event.Author,
				"timestamp": event.Timestamp,
				"text":      strings.Join(textParts, "."),
			}
			data, err := json.Marshal(m)
			if err != nil {
				return err
			}
			outputLines = append(outputLines, string(data))
		}
	}

	outputString := strings.Join(outputLines, "\n")
	if _, err := tempfile.WriteString(outputString); err != nil {
		return err
	}

	for _, ragResources := range s.vertexRAGStore.RAGResources {
		// TODO(zchee): set path=temp_file_path,
		_ = ragResources
		req := &aiplatformpb.UploadRagFileRequest{
			RagFile: &aiplatformpb.RagFile{
				RagFileSource: &aiplatformpb.RagFile_DirectUploadSource{
					DirectUploadSource: &aiplatformpb.DirectUploadSource{},
				},
				DisplayName: fmt.Sprintf("%s.%s.%s", session.AppName(), session.UserID(), session.ID()),
			},
		}
		s.vertexRagDataClient.UploadRagFile(ctx, req)
	}

	return nil
}

// SearchMemory implements [types.MemoryService].
//
// TODO(zchee): implements
func (s *VertexAIRagService) SearchMemory(ctx context.Context, appName, userID, query string) (*types.SearchMemoryResponse, error) {
	return nil, errors.New("not implemented: Vertex AI RAG integration requires additional dependencies")

	s.logger.InfoContext(ctx, "Searching Vertex AI RAG memory",
		slog.String("app_name", appName),
		slog.String("user_id", userID),
		slog.String("query", query),
		slog.String("rag_corpus", s.ragCorpus),
	)

	// This would require integration with Google Cloud Vertex AI
	// Implementation would involve:
	// 1. Creating a search query for the RAG corpus
	// 2. Retrieving matching documents
	// 3. Converting documents to Result objects

	return nil, nil
}
