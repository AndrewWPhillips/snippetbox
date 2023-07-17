package mysql

import (
	"database/sql"
	"log"
	"strings"

	"github.com/andrewwphillips/snippetbox/pkg/models"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// UserModel provides methods for login system
type UserModel struct {
	DB *sql.DB
}

// NewUserModel creates a UserModel for using the users table
func NewUserModel(dsn string) *UserModel {
	// Add parseTime to the DSN so that time.Time (Created) field is translated correctly
	db, err := sql.Open("mysql", dsn+"?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	return &UserModel{DB: db}
}

func (m *UserModel) Close() {
	m.DB.Close()
}

// Insert adds a new user
func (m *UserModel) Insert(name, email, password string) (int, error) {
	query := "INSERT " +
		"INTO users (name, email, hashed_password, created) " +
		"VALUES(?, ?, ?, UTC_TIMESTAMP()) "

	bCryptCost := 12 // TODO adjust cost dep. on current year?
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bCryptCost)
	if err != nil {
		return 0, err
	}

	result, err2 := m.DB.Exec(query, name, email, string(hashedPassword))
	if err2 != nil {
		// Check for the special case of an email address being the same as an existing one
		if mysqlErr, ok := err2.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "users_uc_email") {
				return 0, models.ErrDuplicateEmail
			}
		}
		return 0, err2
	}

	id, err3 := result.LastInsertId()
	if err3 != nil {
		return 0, err3
	}

	return int(id), nil
}

// Authenticate verifies a user exists with the specified password and returns their user ID and name
// If not found or the wrong password is given then it returns the error models.ErrInvalidCredentials
func (m *UserModel) Authenticate(email, password string) (int, string, error) {
	// Get the ID and encrypted password for the user (email)
	var id int
	var name string
	var hashedPassword []byte
	row := m.DB.QueryRow("SELECT id, name, hashed_password FROM users WHERE email = ?", email)
	err := row.Scan(&id, &name, &hashedPassword)
	if err == sql.ErrNoRows {
		return 0, "", models.ErrInvalidCredentials
	} else if err != nil {
		return 0, "", err
	}

	// Check that the password given is correct
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", models.ErrInvalidCredentials
	} else if err != nil {
		return 0, "", err
	}

	return id, name, nil
}

// Get retrieves user info based on ID
// It returns
//   - ptr to models.User with data on the user and nil (no error) on success
//   - nil and an error if something goes wrong
//   - nil and nil (no error) if the user was not found
func (m *UserModel) Get(id int) (*models.User, error) {
	s := &models.User{}

	stmt := `SELECT id, name, email, created FROM users WHERE id = ?`
	err := m.DB.QueryRow(stmt, id).Scan(&s.ID, &s.Name, &s.Email, &s.Created)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		} else {
			return nil, nil
		}
	}

	return s, nil
}
