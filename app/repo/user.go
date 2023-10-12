package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

// ErrNoRowsFound is a custom error used when no db rows are found
var ErrNoRowsFound = errors.New("no rows found")

// User represents a user in db
type User struct {
	ID        int    `db:"id"`
	Email     string `db:"email"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
}

// IUserRepo is repo interface for accessing users in db
type IUserRepo interface {
	GetUserByID(ctx context.Context, tx Querier, schema string, userID int) (*User, error)
	ListUsers(ctx context.Context, tx Querier, schema string) ([]User, error)
	CreateUser(ctx context.Context, tx Querier, schema string, u User) (*User, error)
}

// UserRepo is implementation of IUserRepo
type UserRepo struct{}

// NewUserRepo creates a new user repo
func NewUserRepo() IUserRepo {
	return &UserRepo{}
}

// GetUserByID fetches a user from the db by ID
func (r *UserRepo) GetUserByID(ctx context.Context, tx Querier, schema string, userID int) (*User, error) {
	sqlStatement := fmt.Sprintf(
		`SELECT id, email, first_name, last_name
		FROM %[1]s.users
		WHERE id = $1`,
		schema)
	row := tx.QueryRowContext(ctx, sqlStatement, userID)

	var u User
	if err := row.Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRowsFound
		}
		return nil, errors.Wrap(err, "problem fetching user by id")
	}
	return &u, nil
}

// ListUsers gets all users from db
func (r *UserRepo) ListUsers(ctx context.Context, tx Querier, schema string) ([]User, error) {
	sqlStatement := fmt.Sprintf(
		`SELECT id, email, first_name, last_name
		FROM %[1]s.users`,
		schema)
	rows, err := tx.QueryContext(ctx, sqlStatement)
	if err != nil {
		return nil, errors.Wrap(err, "problem getting all users")
	}
	defer rows.Close()

	var result []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName); err != nil {
			return nil, errors.Wrap(err, "problem scanning user row")
		}
		result = append(result, u)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating rows")
	}

	return result, nil
}

// CreateUser creates a new user in db
func (r *UserRepo) CreateUser(ctx context.Context, tx Querier, schema string, u User) (*User, error) {
	sqlStatement := fmt.Sprintf(
		`INSERT INTO %[1]s.users
		(email, first_name, last_name)
		VALUES
		($1, $2, $3)
		RETURNING id`,
		schema)
	row := tx.QueryRowContext(ctx, sqlStatement, u.Email, u.FirstName, u.LastName)

	if err := row.Scan(&u.ID); err != nil {
		return nil, errors.Wrap(err, "problem inserting user")
	}

	return &u, nil
}
