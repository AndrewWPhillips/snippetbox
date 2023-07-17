package mock

import (
	"time"

	"github.com/andrewwphillips/snippetbox/pkg/models"
)

var mockUser = &models.User{
	ID:      1,
	Name:    "Alice",
	Email:   "alice@example.com",
	Created: time.Now(),
}

type (
	UserModel struct{ users []*models.User }
)

func NewUserModel(dsn string) *UserModel {
	return &UserModel{[]*models.User{mockUser}}
}

func (m *UserModel) Close() {
}

func (m *UserModel) Insert(name, email, password string) (int, error) {
	// Refactor to make the mock more real by keeping a slice of users
	//switch email {
	//case "dupe@example.com":
	//	return 0, models.ErrDuplicateEmail
	//default:
	//	return 2, nil
	//}
	// Check if existing user has the email address
	for _, user := range m.users {
		if user.Email == email {
			return 0, models.ErrDuplicateEmail
		}
	}
	ID := len(m.users)
	m.users = append(m.users, &models.User{ID, name, email, []byte(password), time.Now()})
	return ID, nil
}

func (m *UserModel) Authenticate(email, password string) (int, string, error) {
	// Refactor to use the slice of users
	//switch email {
	//case "alice@example.com":
	//	return mockUser.ID, mockUser.Name, nil
	//default:
	//	return 0, "", models.ErrInvalidCredentials
	//}
	for i, user := range m.users {
		if user.Email == email {
			if string(user.HashedPassword) != password {
				return 0, "", models.ErrInvalidCredentials
			}
			return i, user.Name, nil // found
		}
	}
	return 0, "", models.ErrInvalidCredentials // not found
}

func (m *UserModel) Get(id int) (*models.User, error) {
	//switch id {
	//case 1:
	//	return mockUser, nil
	//default:
	//	return nil, nil
	//}
	if id < 0 || id >= len(m.users) {
		return nil, nil
	}
	return m.users[id], nil
}
