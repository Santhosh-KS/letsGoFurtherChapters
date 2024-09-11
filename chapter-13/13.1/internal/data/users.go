package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"greelight.techkunstler.com/internal/validator"
)

// Define a custom ErrDuplicateEmail error
var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

// Define a User struct to represent an individual user. Importantly,
// notice how we are uisng json:"-" struct tag to prevent the password
// and version fields appearing in any output when we encode it to JSON.
// Also notice that the pasword field uses the custom password type defined below.

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at`
	Name      string    `json:"name"`
	Email     string    `json:email`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `"json:"-"`
}

// Create a custom password type which is a struct containing the plaintext and
// hashed versions of the password for a user. The plaintext field is *pointer*
// to a string, so that we're able to distinguish between a plaintext password
// not being present in the struct at all, versus a plaintext password which is the
// empty string"".

type password struct {
	plaintext *string
	hash      []byte
}

// The Set() method calculates the bcrypt hash of a plaintext password, and stores both
// the hash and the plaintext versions in the struct.

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX),
		"email", "must be a valid email address")

}

func ValidatePasswordPlainText(v *validator.Validator, password string) {
	v.Check(password != "", "passwrod", "must be provided")
	v.Check(len(password) >= 8, "password", "must be atleast 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "mst be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	// Call the standalone ValidateEmail() helper.
	ValidateEmail(v, user.Email)

	// If the plaintext password is not nil, call the standalone
	// ValidatePasswordPlainText() helper

	if user.Password.plaintext != nil {
		ValidatePasswordPlainText(v, *user.Password.plaintext)
	}

	// If the password hash is ever nil, this will be due to a logic error in our
	// codebase (probably because we forgot to set password for the user). It's
	// a useful sanity check to include here, but it's not a problem with the data
	// provided by the client. So rather than adding an error to the validation map
	// we raise a panic instead.
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

// Create a UserModel struct which wraps the connection pool.

type UserModel struct {
	DB *sql.DB
}

// Insert a new record in the databasw for the user. Note that the id,
// created_at and version fields are all automatically generated by our database, so we use the
// RETURNING clause to read them into the User struct after the insert,
// in the same way that we did when creating a movie.

func (m UserModel) Insert(user *User) error {
	query := `
	INSERT INTO users (name, email, password_hash, activated)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version
	`

	args := []any{user.Name, user.Email, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// If the table already contains a record with this email address, then when
	// we try to perform the insert there will be violation of the UNIQUE "user_email_key"
	// constraint that we setup in the previous chapter. We check for this error
	// specifically, and return custom ErrDuplicateEmail error instead.

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID,
		&user.CreatedAt, &user.Version)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates the unique constraint "user_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

// Retrieve the User details from the database based on the user's email address.
// Because we have a UNIQUE constraint on the email column, this SQL query will only
// return one record(ornot at all, in which case we return a ErrRecordNotFound error)

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
	SELECT id, created_at, name, email, password_hash, activated, version
	FROM users
	WHERE email = $1
	`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// Update the details for a speicifc user. Notice that we check against the version
// field to help prevent any race conditions during the request cycle,
// just like we did when updating a movie. And we also check for violations of the
// "users_email_key" constraint when performing the update, just like we did when
// inserting the user record originally.

func (m UserModel) Update(user *User) error {
	query := `
	UPDATE users
	SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
	WHERE id = $5 AND version = $6
	RETURNING version
	`

	args := []any{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "user_email_key`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}
