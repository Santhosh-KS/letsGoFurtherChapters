package data

import (
	"database/sql"
	"greelight.techkunstler.com/internal/validator"
	"time"
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

// Add a placeholder method for inserting a new record in the movies table
func (m MovieModel) Insert(movie *Movie) error {
	return nil
}

// Add a placeholder method for fetching a specific record from the movies table.
func (m MovieModel) Get(id int64) (*Movie, error) {
	return nil, nil
}

// Add a placeholder method for updating a specific record in the movies table.
func (m MovieModel) Update(movie *Movie) error {
	return nil
}

func (m MovieModel) Delete(id int64) error {
	return nil
}
