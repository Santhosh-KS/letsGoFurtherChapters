package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"greelight.techkunstler.com/internal/validator"
)

// Define constants for the token scope. For now we just define the scope "activation"
// but we'll add additional scopes later in the book.

const (
	ScopeActivation = "activation"
)

type Token struct {
	Plaintext string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
	Scope     string
}

func generatedToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	// Create a Token instance containing the user ID, expiry, and scope information.
	// Notice that we add the provided ttl (time to live) duration parameter to the
	// current time to get the expiry time?

	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	// Initialize a zero-valued byte slice with a lenght of 16 bytes.

	randomBytes := make([]byte, 16)

	// Use the Read() function from teh crypto/rand pacakge to fill
	// the byte slice with the random bytes from your OS's CSPRNG.
	// This will return an error if CSPRNG fails to function correctly.

	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// Encode the byte slice to a base-32-encoded string and assign it to the
	// token Plaintext field. This will be the token string that we send to the user in their
	// welcome email. They will look smilar to this:
	// Y3QMDG6FJSDFT176FDNLKNXUOJ
	// Note that by default base-32 strings may be padded at the end with the = charecter.
	// we don't need this padding character for the purpose of our tokens, so
	// we use the WithPadding(base32.NoPadding) method in the line below to omit them.

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// Generate a SHA-256 hash of the plaintext token string. This will be the value
	// that we store in the `hash` field of our database table. Note that the
	// sha256.Sum256() function returns an *array* of length 32, so to make it easier to
	// work with we convert it to a slice using [:] operator before storing it.

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

// Check that the plaintet token has been provided and is exactly 26 bytes long.
func ValidateTokenPlainText(v *validator.Validator, tokenPlainText string) {
	v.Check(tokenPlainText != "", "token", "must be provided")
	v.Check(len(tokenPlainText) == 26, "token", "must be 26 bytes long")
}

type TokenModel struct {
	DB *sql.DB
}

// The New() method is a shortcut which creates a new Token struct and then inserts the
// data in the tokens table.

func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generatedToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

// Insert() adds the data for a specific token to the tokens table.

func (m TokenModel) Insert(token *Token) error {
	query := `
	INSERT INTO tokesn (hash, user_id, expiry, scope)
	VALUES ($1, $2, $3, $4)
	`

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)

	return err
}

// DeleteAllForUser() deletes all tokens for a specific user and scope.

func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
	DLETE FROM tokens
	WHERE scope = $1 AND user_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}
