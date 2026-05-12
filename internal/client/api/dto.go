package api

type createUserRequest struct {
	Username            string `json:"username"`
	EncryptionPublicKey string `json:"encryption_public_key"`
	SigningPublicKey    string `json:"signing_public_key"`
}

type createUserResponse struct {
	ID int `json:"id"`
}

type getUserPublicKeysResponse struct {
	ID                  int    `json:"id"`
	Username            string `json:"username"`
	EncryptionPublicKey string `json:"encryption_public_key"`
	SigningPublicKey    string `json:"signing_public_key"`
}

type getUserByUsernameResponse struct {
	ID                  int    `json:"id"`
	EncryptionPublicKey string `json:"encryption_public_key"`
	SigningPublicKey    string `json:"signing_public_key"`
}

type storeContainerRequest struct {
	RecipientID int    `json:"recipient_id"`
	Container   string `json:"container"`
	FileName    string `json:"file_name"`
	MimeType    string `json:"mime_type"`
	Size        int64  `json:"size"`
}

type storeContainerResponse struct {
	ID int `json:"id"`
}

type loadContainerResponse struct {
	ID          int    `json:"id"`
	SenderID    int    `json:"sender_id"`
	RecipientID int    `json:"recipient_id"`
	Container   string `json:"container"`
	FileName    string `json:"file_name"`
	MimeType    string `json:"mime_type"`
	Size        int64  `json:"size"`
	CreatedAt   string `json:"created_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type challengeRequest struct {
	UserID int `json:"user_id"`
}

type challengeResponse struct {
	Nonce string `json:"nonce"`
}

type verifyRequest struct {
	UserID    int    `json:"user_id"`
	Signature string `json:"signature"`
}

type verifyResponse struct {
	SessionToken string `json:"session_token"`
	ExpiresAt    string `json:"expires_at"`
}
