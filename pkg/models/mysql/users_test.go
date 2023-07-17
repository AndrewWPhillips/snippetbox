package mysql

import (
	"reflect"
	"testing"
	"time"

	"github.com/andrewwphillips/snippetbox/pkg/models"
)

// TestUserModelGet tests that getting a user from the users DB table works as expected
// The "Let's Go" book calls this "integration testing" but it's really just testing the
// Users interface (of which Get is a method)
func TestUserModelGet(t *testing.T) {
	if testing.Short() {
		t.Skip("mysql: skipping integration test due to use of -short")
	}

	// Set up a suite of table-driven tests and expected results.
	tests := []struct {
		name     string
		userID   int
		wantUser *models.User
		notFound bool
	}{
		{
			name:   "Valid ID",
			userID: 1,
			wantUser: &models.User{
				ID:      1,
				Name:    "Alice Jones",
				Email:   "alice@example.com",
				Created: time.Date(2023, 03, 23, 17, 25, 22, 0, time.UTC),
			},
		},
		{
			name:     "Zero ID",
			userID:   0,
			wantUser: nil,
			notFound: true,
		},
		{
			name:     "Non-existent ID",
			userID:   2,
			wantUser: nil,
			notFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize a connection pool to our test database, and defer a
			// call to the teardown function, so it is always run immediately
			// before this sub-test returns.
			db, teardown := newTestDB(t)
			defer teardown()

			// Create a new instance of the UserModel.
			m := UserModel{db}

			// Call the UserModel.Get() method and check that the return value
			// and error match the expected values for the sub-test.
			user, err := m.Get(tt.userID)

			if err != nil {
				t.Fatal(err)
			}
			if tt.notFound {
				if user != nil {
					t.Errorf("expected user not found but got %q", user.Name)
				}
			} else {
				if !reflect.DeepEqual(user, tt.wantUser) {
					t.Errorf("want %v; got %v", tt.wantUser, user)
				}
			}
		})
	}
}
