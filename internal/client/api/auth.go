package api

import (
	"context"
	"encoding/base64"
	"net/http"
	"time"
)

type RequestChallengeOutput struct {
	Nonce []byte
}

type VerifyChallengeOutput struct {
	SessionToken string
	ExpiresAt    time.Time
}

// RequestChallenge запрашивает у сервера nonce для подписи.
func (c *Client) RequestChallenge(ctx context.Context, userID int) (RequestChallengeOutput, error) {
	var out challengeResponse
	err := c.doJSON(ctx, http.MethodPost, "/auth/challenge", challengeRequest{
		UserID: userID,
	}, &out)
	if err != nil {
		return RequestChallengeOutput{}, err
	}

	nonce, err := base64.StdEncoding.DecodeString(out.Nonce)
	if err != nil {
		return RequestChallengeOutput{}, err
	}

	return RequestChallengeOutput{Nonce: nonce}, nil
}

// VerifyChallenge отправляет подпись nonce и получает session token.
func (c *Client) VerifyChallenge(ctx context.Context, userID int, signature []byte) (VerifyChallengeOutput, error) {
	var out verifyResponse
	err := c.doJSON(ctx, http.MethodPost, "/auth/verify", verifyRequest{
		UserID:    userID,
		Signature: base64.StdEncoding.EncodeToString(signature),
	}, &out)
	if err != nil {
		return VerifyChallengeOutput{}, err
	}

	expiresAt, err := time.Parse(time.RFC3339, out.ExpiresAt)
	if err != nil {
		return VerifyChallengeOutput{}, err
	}

	return VerifyChallengeOutput{
		SessionToken: out.SessionToken,
		ExpiresAt:    expiresAt,
	}, nil
}
