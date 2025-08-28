package firebase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/theHinneh/budgeting/internal/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RefreshTokenRepository struct {
	Firestore *firestore.Client
}

const refreshTokensCollection = "refresh_tokens"

func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	// Hash the token for security
	token.TokenHash = hashToken(token.TokenHash)

	// Set creation time
	if token.CreatedAt.IsZero() {
		token.CreatedAt = time.Now()
	}

	// Generate a unique ID if not provided
	if token.ID == "" {
		token.ID = generateTokenID()
	}

	_, err := r.Firestore.Collection(refreshTokensCollection).Doc(token.ID).Set(ctx, token)
	return err
}

func (r *RefreshTokenRepository) GetByID(ctx context.Context, id string) (*domain.RefreshToken, error) {
	doc, err := r.Firestore.Collection(refreshTokensCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("refresh token not found")
		}
		return nil, err
	}

	var token domain.RefreshToken
	if err := doc.DataTo(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *RefreshTokenRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.RefreshToken, error) {
	iter := r.Firestore.Collection(refreshTokensCollection).
		Where("user_id", "==", userID).
		Documents(ctx)

	var tokens []*domain.RefreshToken
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var token domain.RefreshToken
		if err := doc.DataTo(&token); err != nil {
			return nil, err
		}
		tokens = append(tokens, &token)
	}

	return tokens, nil
}

func (r *RefreshTokenRepository) GetValidToken(ctx context.Context, userID, tokenHash string) (*domain.RefreshToken, error) {
	hashedToken := hashToken(tokenHash)

	iter := r.Firestore.Collection(refreshTokensCollection).
		Where("user_id", "==", userID).
		Where("token_hash", "==", hashedToken).
		Where("is_revoked", "==", false).
		Where("expires_at", ">", time.Now()).
		Limit(1).
		Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("valid refresh token not found")
	}
	if err != nil {
		return nil, err
	}

	var token domain.RefreshToken
	if err := doc.DataTo(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *RefreshTokenRepository) RevokeToken(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.Firestore.Collection(refreshTokensCollection).Doc(id).Update(ctx, []firestore.Update{
		{Path: "is_revoked", Value: true},
		{Path: "revoked_at", Value: now},
	})
	return err
}

func (r *RefreshTokenRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	iter := r.Firestore.Collection(refreshTokensCollection).
		Where("user_id", "==", userID).
		Where("is_revoked", "==", false).
		Documents(ctx)

	batch := r.Firestore.Batch()
	now := time.Now()
	count := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		batch.Update(doc.Ref, []firestore.Update{
			{Path: "is_revoked", Value: true},
			{Path: "revoked_at", Value: now},
		})
		count++
	}

	if count > 0 {
		_, err := batch.Commit(ctx)
		return err
	}

	return nil
}

func (r *RefreshTokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	iter := r.Firestore.Collection(refreshTokensCollection).
		Where("expires_at", "<", time.Now()).
		Documents(ctx)

	batch := r.Firestore.Batch()
	count := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		batch.Delete(doc.Ref)
		count++
	}

	if count > 0 {
		_, err := batch.Commit(ctx)
		return err
	}

	return nil
}

func (r *RefreshTokenRepository) DeleteToken(ctx context.Context, id string) error {
	_, err := r.Firestore.Collection(refreshTokensCollection).Doc(id).Delete(ctx)
	return err
}

// Helper functions
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func generateTokenID() string {
	return fmt.Sprintf("rt_%d", time.Now().UnixNano())
}
