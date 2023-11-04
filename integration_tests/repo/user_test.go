package test_repo

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/drmaples/starter-app/app/repo"
)

type userSuite struct {
	suite.Suite

	container IPostgresContainer
	ctx       context.Context
	db        *sql.DB
	userRepo  repo.IUserRepo
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(userSuite))
}

func (s *userSuite) SetupSuite() {
	s.container = NewPostgresContainer()
	assert.NoError(s.T(), s.container.Setup())

	s.db = s.container.GetDB()
	s.userRepo = repo.NewUserRepo()
}

func (s *userSuite) TearDownSuite() {
	assert.NoError(s.T(), s.container.TearDown())
}

func (s *userSuite) SetupTest() {
	s.ctx = context.TODO()
}

func (s *userSuite) TearDownTest() {
}

func (s *userSuite) TestCreateUser() {
	u := repo.User{
		Email:     fmt.Sprintf("%s@example.com", uuid.New().String()),
		FirstName: "foo",
		LastName:  "bar",
	}
	newUser, err := s.userRepo.CreateUser(s.ctx, s.db, repo.DefaultSchema, u)
	assert.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), newUser.ID, 1)
	assert.Equal(s.T(), u.Email, newUser.Email)
	assert.Equal(s.T(), u.FirstName, newUser.FirstName)
	assert.Equal(s.T(), u.LastName, newUser.LastName)
}

func (s *userSuite) TestGetUserByID() {
	u, err := s.userRepo.CreateUser(s.ctx, s.db, repo.DefaultSchema, repo.User{
		Email:     fmt.Sprintf("%s@example.com", uuid.New().String()),
		FirstName: "foo",
		LastName:  "bar",
	})
	assert.NoError(s.T(), err)

	fetchedUser, err := s.userRepo.GetUserByID(s.ctx, s.db, repo.DefaultSchema, u.ID)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), u.ID, fetchedUser.ID)
}

func (s *userSuite) TestListUsers() {
	u, err := s.userRepo.CreateUser(s.ctx, s.db, repo.DefaultSchema, repo.User{
		Email:     fmt.Sprintf("%s@example.com", uuid.New().String()),
		FirstName: "foo",
		LastName:  "bar",
	})
	assert.NoError(s.T(), err)

	users, err := s.userRepo.ListUsers(s.ctx, s.db, repo.DefaultSchema)
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), users)

	assert.True(s.T(), lo.ContainsBy(users, func(x repo.User) bool { return u.ID == x.ID }))
}
