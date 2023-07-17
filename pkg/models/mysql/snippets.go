package mysql

import (
	"database/sql"
	"log"

	"github.com/andrewwphillips/snippetbox/pkg/models"
)

// SnippetModel wraps a sql.DB connection pool and provides methods to operate on the snippets table
type SnippetModel struct {
	DB *sql.DB
}

// NewSnippetModel creates a SnippetModel for manipulating the snippets database table
func NewSnippetModel(dsn string) *SnippetModel {
	// Add parseTime to the DSN so that time.Time (Created and Expires) fields are translated correctly
	db, err := sql.Open("mysql", dsn+"?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	return &SnippetModel{DB: db}
}

func (m *SnippetModel) Close() {
	m.DB.Close()
}

// Insert adds a new snippet to the database.
func (m *SnippetModel) Insert(title, content, expires string) (int, error) {
	query := "INSERT " +
		"INTO snippets (title, content, created, expires) " +
		"VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY)) "

	result, err := m.DB.Exec(query, title, content, expires)
	if err != nil {
		return 0, err
	}

	id, err2 := result.LastInsertId()
	if err != nil {
		return 0, err2
	}

	return int(id), nil
}

// Get returns a snippet as long as it has not expired
// If the snippet is found it is returned (and error return is nil)
// If the snippet is NOT found it returns nil for the snippet AND the error.
// It returns an error (and nil snippet) if there was some real error.
func (m *SnippetModel) Get(id int) (*models.Snippet, error) {
	query := "SELECT id, title, content, created, expires " +
		"FROM snippets " +
		"WHERE expires > UTC_TIMESTAMP() AND id = ? "

	// Query for the record by ID and get its fields.  Note that the parameters passed to Scan must
	// correspond to the fields requested (number and rough type) in the query.
	s := &models.Snippet{}
	if err := m.DB.QueryRow(query, id).Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires); err == sql.ErrNoRows {
		return nil, nil // not an error - just snippet not found
	} else if err != nil {
		return nil, err
	}

	return s, nil // return the found snippet
}

// Latest returns the latest snippets (up to 10) as long as not expired
func (m *SnippetModel) Latest() ([]*models.Snippet, error) {
	const limit = 10
	query := "SELECT id, title, content, created, expires " +
		"FROM snippets " +
		"WHERE expires > UTC_TIMESTAMP() " +
		"ORDER BY created DESC " +
		"LIMIT ? "

	// Query for the top "limit" number of records when ordered by creation date
	rows, err := m.DB.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan the resulting rows and add them to the returned slice
	snippets := make([]*models.Snippet, 0, limit)
	for rows.Next() {
		// Get the  fields.  Note that the parameters passed to Scan must
		// correspond to the fields requested (number and rough type) in the query.
		s := &models.Snippet{}
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, s)
	}

	// Next may return false at the end or if there was an error
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
