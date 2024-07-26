package models

import (
	"snippetbox.glyphsmiths.com/internal/assert"
	"testing"
)

func TestUserModelExists(t *testing.T) {

	// Skip the rest if the "-short" flag is provided when runnign the test.
	if testing.Short() {
		t.Skip("models: skipping integration test")
	}
	// Setup a suite of table-driver tests and expected results.

	tests := []struct {
		name   string
		userId int
		want   bool
	}{
		{
			name:   "Valid ID",
			userId: 1,
			want:   true,
		},
		{
			name:   "Zero ID",
			userId: 0,
			want:   false,
		},
		{
			name:   "Non-existent ID",
			userId: 2,
			want:   false,
		},
	}

	for _, tt := range tests {
		// Call the newTestDB() helerp function to get a connection pool to our test database. Calling thishere --inside t.Run() --meansthat fresh database tables and data will be setup and torn down for reach sub-set

		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			// Createa new instance of the userModel.
			m := UserModel{db}

			exists, err := m.Exists(tt.userId)
			assert.Equal(t, exists, tt.want)
			assert.NilError(t, err)
		})
	}
}
