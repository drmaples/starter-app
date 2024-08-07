package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/drmaples/starter-app/app/dto"
	"github.com/drmaples/starter-app/app/repo"
)

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) GetUserByID(_ context.Context, _ repo.Querier, _ string, userID int) (*repo.User, error) {
	args := m.Called(mock.Anything, mock.Anything, mock.Anything, userID)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repo.User), args.Error(1)
}

func (m *mockUserRepo) ListUsers(_ context.Context, _ repo.Querier, _ string) ([]repo.User, error) {
	args := m.Called()
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repo.User), args.Error(1)
}

func (m *mockUserRepo) CreateUser(_ context.Context, _ repo.Querier, _ string, _ repo.User) (*repo.User, error) {
	args := m.Called()
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	newUser := args.Get(0).(repo.User)
	newUser.ID = 9999
	return &newUser, args.Error(1)
}

type controllerTestSuite struct {
	suite.Suite
	FakeUser *repo.User
	Token    *jwt.Token
}

func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(controllerTestSuite))
}

func (s *controllerTestSuite) SetupTest() {
	s.FakeUser = &repo.User{
		ID:        111,
		Email:     "foo@example.com",
		FirstName: "foo",
		LastName:  "bar",
	}
	s.Token = newToken("logged-in@example.com", "first last", "example.com", 15*time.Minute)
}

func (s *controllerTestSuite) Test_handleGetUser_success() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()
	c := e.NewContext(req, recorder)
	c.Set(authContextKey, s.Token) // fake authentication
	c.SetPath("/v1/user/:id")
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprint(s.FakeUser.ID))

	m := new(mockUserRepo)
	m.On("GetUserByID", mock.Anything, mock.Anything, mock.Anything, s.FakeUser.ID).Return(s.FakeUser, nil)

	con := Controller{e: e, userRepo: m}
	assert.NoError(s.T(), con.handleGetUser(c))
	assert.Equal(s.T(), http.StatusOK, recorder.Code)

	var actual dto.User
	assert.NoError(s.T(), json.Unmarshal(recorder.Body.Bytes(), &actual))
	assert.Equal(s.T(), s.FakeUser.ID, actual.ID)
	assert.Equal(s.T(), s.FakeUser.Email, actual.Email)
}

func (s *controllerTestSuite) Test_handleGetUser_not_found() {
	bogusUserID := -999
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()
	c := e.NewContext(req, recorder)
	c.Set(authContextKey, s.Token) // fake authentication
	c.SetPath("/v1/user/:id")
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprint(bogusUserID))

	m := new(mockUserRepo)
	m.On("GetUserByID", mock.Anything, mock.Anything, mock.Anything, bogusUserID).Return(nil, repo.ErrNoRowsFound)

	con := Controller{e: e, userRepo: m}
	assert.NoError(s.T(), con.handleGetUser(c))
	assert.Equal(s.T(), http.StatusNotFound, recorder.Code)

	var actual dto.ErrorResponse
	assert.NoError(s.T(), json.Unmarshal(recorder.Body.Bytes(), &actual))
	assert.Equal(s.T(), actual.Message, "no user for given id")
}

func (s *controllerTestSuite) Test_unauthorized() {
	e := echo.New()
	con := Controller{e: e}
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	s.T().Run("missing_jwt", func(_ *testing.T) {
		recorder := httptest.NewRecorder()
		c := e.NewContext(req, recorder)
		c.SetPath("/v1/user")
		// c.Set(authContextKey, s.Token) // fake authentication

		assert.NoError(s.T(), con.handleListUsers(c))
		assert.Equal(s.T(), http.StatusUnauthorized, recorder.Code)

		var actual dto.ErrorResponse
		assert.NoError(s.T(), json.Unmarshal(recorder.Body.Bytes(), &actual))
		assert.Equal(s.T(), "jwt missing", actual.Message)
	})

	s.T().Run("invalid_jwt", func(_ *testing.T) {
		recorder := httptest.NewRecorder()
		c := e.NewContext(req, recorder)
		c.SetPath("/v1/user")
		c.Set(authContextKey, "bogus token")

		assert.NoError(s.T(), con.handleListUsers(c))
		assert.Equal(s.T(), http.StatusUnauthorized, recorder.Code)

		var actual dto.ErrorResponse
		assert.NoError(s.T(), json.Unmarshal(recorder.Body.Bytes(), &actual))
		assert.Equal(s.T(), "jwt is incorrect type", actual.Message)
	})
}

func (s *controllerTestSuite) Test_handleListUsers_success() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()
	c := e.NewContext(req, recorder)
	c.Set(authContextKey, s.Token) // fake authentication
	c.SetPath("/v1/user")

	m := new(mockUserRepo)
	m.On("ListUsers", mock.Anything, mock.Anything, mock.Anything).Return([]repo.User{*s.FakeUser}, nil)

	con := Controller{e: e, userRepo: m}
	assert.NoError(s.T(), con.handleListUsers(c))
	assert.Equal(s.T(), http.StatusOK, recorder.Code)

	var actual []dto.User
	assert.NoError(s.T(), json.Unmarshal(recorder.Body.Bytes(), &actual))
	assert.Len(s.T(), actual, 1)
	assert.Equal(s.T(), s.FakeUser.ID, actual[0].ID)
	assert.Equal(s.T(), s.FakeUser.Email, actual[0].Email)
}

func (s *controllerTestSuite) Test_handleCreateUser_bad_input() {
	e := echo.New()
	e.Validator = newValidator() // must register validator
	con := Controller{e: e}

	s.T().Run("missing_email", func(_ *testing.T) {
		input := `{"missing":"email", "first_name":"foo", "last_name":"bar"}`

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(input))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recorder := httptest.NewRecorder()
		c := e.NewContext(req, recorder)
		c.Set(authContextKey, s.Token) // fake authentication
		c.SetPath("/v1/user")

		assert.NoError(s.T(), con.handleCreateUser(c))
		assert.Equal(s.T(), http.StatusBadRequest, recorder.Code)

		var actual dto.ErrorResponse
		assert.NoError(s.T(), json.Unmarshal(recorder.Body.Bytes(), &actual))
		assert.Contains(s.T(), actual.Message, `'Email' failed on the 'required' tag`)
	})

	s.T().Run("invalid_json", func(_ *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("invalid_json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		recorder := httptest.NewRecorder()
		c := e.NewContext(req, recorder)
		c.Set(authContextKey, s.Token) // fake authentication
		c.SetPath("/v1/user")

		assert.NoError(s.T(), con.handleCreateUser(c))
		assert.Equal(s.T(), http.StatusBadRequest, recorder.Code)

		var actual dto.ErrorResponse
		assert.NoError(s.T(), json.Unmarshal(recorder.Body.Bytes(), &actual))
		assert.Contains(s.T(), actual.Message, `invalid character`)
	})
}

func (s *controllerTestSuite) Test_handleCreateUser_success() {
	// assert.Fail(s.T(), "implement me")
}
