package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"greelight.techkunstler.com/internal/validator"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	// Use the Runtime type instead of int32. Note that the omitempty directive will
	// still work on this: if the Runtime field has the underlying value 0, then
	// it will be considered empty and omited -- and the MarshalJSON() method we just mad
	// won't be called at all.
	Runtime Runtime  `json:"runttime,omitempty,string"`
	Genres  []string `json:"genres,omitempty"`
	Version int32    `json:"version"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {

	// Use the Check() method to execute our validation checks. This will add
	// the provided key and error message to the errors map if the check does not
	// evaluate to true. For example, in the first line here we "check that the title
	// is not equal to the empty string".
	// In the second, we "check that the lenght of the title is less than or equal to
	// 500 bytes" and so on.
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year > 1888, "year", "must be greatethan 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be positive integer")

	v.Check(movie.Genres != nil, "geners", "must be provided")
	v.Check(len(movie.Genres) >= 1, "geners", "must contain at least one genre")
	v.Check(len(movie.Genres) <= 5, "geners", "must not contain more than five genres")

	// Note that we're using the Unique helper in the line below to check that all
	// values in the input.Genres slice are unique.

	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")

}

// Define a MovieModel struct type which wraps a sql.DB connection pool.

type MovieModel struct {
	DB *sql.DB
}

// The Insert method accepts a pointer to a movie struct, which should contain the data
// for the new record.
func (m MovieModel) Insert(movie *Movie) error {
	// Define the SQL query for inserting a new record in the movies table and returning
	// the system-generated data..
	query := `
	INSERT INTO movies (title, year, runtime, genres)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version`

	// Create an args slice containing the values for the plaeholder parameters from
	// the movie struct. Declaring this slice immediately next to our SQL query helps to
	// make it nice and clear *what values are being used where* in the query.

	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// Use the QueryRow() method to execute the SQL query on our connection pool,
	// passing in the args slice as a variadic parameter and scanning the system-generated id, created_at and version values into the movie struct.

	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Add a placeholder method for fetching a specific record from the movies table.
func (m MovieModel) Get(id int64) (*Movie, error) {
	// The PostgresSQL bigserial type that we're using for the movie ID starts
	// auto-incrementing at 1 by default, so we know that no movies will have ID values
	// less than that. To avoid making an unnecessary database call, we take a shortcut
	// and return an ErrRecordNotFound error straight away.

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retriveing the movie data.
	query :=
		` SELECT id, created_at, title, year, runtime, genres, version
	FROM movies
	WHERE id = $1`

	// Declare a Movie struct to hold the data returned by the query.
	var movie Movie

	// Execute the query using the QueryRow() method, passing in the provided id value
	// as a placeholder parameter, and scan the response data into the fields of the
	// movie struct. Importantly, notice that we need to convert the scan target for the
	// geners column using the pq.Array() adapter function again.

	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	// Handle any errors. If there was no matching movie found, Scan() will return
	// a sql.ErrNoRows error. We check for this and return our custom ErrRecordNotFound
	// error instead.

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// Otherwise, return a pointer to the movie struct.
	return &movie, nil
}

// Add a placeholder method for updating a specific record in the movies table.
func (m MovieModel) Update(movie *Movie) error {
	// Declare the SQL query for updating the record and returning the new version number.
	// Add the 'AND version = $6' clause to the SQL query.

	query :=
		`UPDATE movies
	SET title = $1, year = $2, runtime= $3, genres = $4, version = version +1
	WHERE id = $5 AND version = $6
	RETURNING version`

	// Create an args slice containing the values for the placeholder parameters.

	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version, // Add the expected movie version.
	}

	/* // Use the QueryRow() method to execute the query, passing in the args slice as
	// a variadic parameter and scanning the new version value into the movie struct.

	// return m.DB.QueryRow(query, args...).Scan(&movie.Version) */
	// Execute the SQL query. If no matching row could be found, we know the movie version has changed (or the record has been deleted)
	// And we return our custom ErrEditConflict error

	err := m.DB.QueryRow(query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil

}

func (m MovieModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1.
	if id < 1 {
		return ErrRecordNotFound
	}

	// Construct the SQL Query to delete the record.
	query := `
	DELETE FROM movies
	WHERE id = $1
	`
	// Execute the SQL query using the Exec() method, passing in the id variable as the value for the placeholder parameter. The Exec() method returns a sql.Result object.

	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}

	// Call the RowsAffected() method on the sql.Result object to get the number of rows
	// affected by the query.

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows were affected, we know that the movies table didn't contain a record
	// with the provided ID at the moment we tried to delete it. In that case
	// we return an ErrRecordNotFound error.

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
