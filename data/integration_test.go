//go:build integration

// run tests with this command:
// go test . --tags integration --count=1
// go test -coverprofile=coverage.out . --tags integration
// go tool cover -html=coverage.out
package data

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	host     = "localhost"
	user     = "postgres"
	password = "secret"
	dbName   = "celeritas_test"
	port     = "5435"
	dsn      = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5"
)

var dummyUser = User{
	FirstName: "Some",
	LastName:  "Guy",
	Email:     "me@here.com",
	Active:    1,
	Password:  "password",
}

var models Models
var testDB *sql.DB
var resource *dockertest.Resource
var pool *dockertest.Pool

func TestMain(m *testing.M) {
	os.Setenv("DATABASE_TYPE", "postgres")

	// get docker image and run it
	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker: %s", err)
	}

	pool = p

	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13.4", // postgres version
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + dbName,
		},
		// ports to expose in docker image
		ExposedPorts: []string{"5432"},
		// bind ports to local ports
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432": {
				{HostIP: "0.0.0.0", HostPort: port},
			},
		},
	}

	// run resource
	resource, err = pool.RunWithOptions(&opts)
	if err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not start resource: %s", err)
	}

	// wait for docker to build the db.
	// ie continue to retry until ping is successful
	if err = pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("pgx", fmt.Sprintf(dsn, host, port, user, password, dbName))
		if err != nil {
			log.Println("error:", err)
			return err
		}
		return testDB.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not connect to docker: %s", err)
	}

	// populate db with tables
	err = createTables(testDB)
	if err != nil {
		log.Fatalf("error creating tables: %s", err)
	}

	// assign models variable
	models = New(testDB)

	// run tests
	code := m.Run()

	// clean up docker image
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %s", err)
	}

	os.Exit(code)
}

func createTables(db *sql.DB) error {
	stmt := `
	CREATE OR REPLACE FUNCTION trigger_set_timestamp()
	RETURNS TRIGGER AS $$
	BEGIN
	  NEW.updated_at = NOW();
	RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;
	
	drop table if exists users cascade;
	
	CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		first_name character varying(255) NOT NULL,
		last_name character varying(255) NOT NULL,
		user_active integer NOT NULL DEFAULT 0,
		email character varying(255) NOT NULL UNIQUE,
		password character varying(60) NOT NULL,
		created_at timestamp without time zone NOT NULL DEFAULT now(),
		updated_at timestamp without time zone NOT NULL DEFAULT now()
	);
	
	CREATE TRIGGER set_timestamp
		BEFORE UPDATE ON users
		FOR EACH ROW
		EXECUTE PROCEDURE trigger_set_timestamp();
	
	drop table if exists remember_tokens;
	
	CREATE TABLE remember_tokens (
		id SERIAL PRIMARY KEY,
		user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
		remember_token character varying(100) NOT NULL,
		created_at timestamp without time zone NOT NULL DEFAULT now(),
		updated_at timestamp without time zone NOT NULL DEFAULT now()
	);
	
	CREATE TRIGGER set_timestamp
		BEFORE UPDATE ON remember_tokens
		FOR EACH ROW
		EXECUTE PROCEDURE trigger_set_timestamp();
	
	drop table if exists tokens;
	
	CREATE TABLE tokens (
		id SERIAL PRIMARY KEY,
		user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
		first_name character varying(255) NOT NULL,
		email character varying(255) NOT NULL,
		token character varying(255) NOT NULL,
		token_hash bytea NOT NULL,
		created_at timestamp without time zone NOT NULL DEFAULT now(),
		updated_at timestamp without time zone NOT NULL DEFAULT now(),
		expiry timestamp without time zone NOT NULL
	);
	
	CREATE TRIGGER set_timestamp
		BEFORE UPDATE ON tokens
		FOR EACH ROW
		EXECUTE PROCEDURE trigger_set_timestamp();
		
	`

	_, err := db.Exec(stmt)
	if err != nil {
		return err
	}

	return nil
}

func TestUser_Table(t *testing.T) {
	s := models.Users.Table()
	if s != "users" {
		t.Error("wrong table named returned:", s)
	}
}

func TestUser_Insert(t *testing.T) {
	id, err := models.Users.Insert(dummyUser)
	if err != nil {
		t.Error("failed to insert user:", err)
	}

	if id == 0 {
		t.Error("0 returned as id after insert")
	}
}

func TestUser_Get(t *testing.T) {
	u, err := models.Users.Get(1)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	if u.ID == 0 {
		t.Error("id of returned user is 0:", err)
	}
}

func TestUser_GetAll(t *testing.T) {
	_, err := models.Users.GetAll()
	if err != nil {
		t.Error("failed to get all users:", err)
	}
}

func TestUser_GetByEmail(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get users by email:", err)
	}

	if u.ID == 0 {
		t.Error("id of returned user is 0:", err)
	}
}

func TestUser_Update(t *testing.T) {
	// get user
	u, err := models.Users.Get(1)
	if err != nil {
		t.Error("failed to get users:", err)
	}

	// change last name
	u.LastName = "Smith"
	err = u.Update(*u)
	if err != nil {
		t.Error("failed to update users:", err)
	}

	// get user again
	u, err = models.Users.Get(1)
	if err != nil {
		t.Error("failed to get users:", err)
	}

	if u.LastName != "Smith" {
		t.Error("last name not updated in database:", err)
	}
}

func TestUser_PasswordMatches(t *testing.T) {
	u, err := models.Users.Get(1)
	if err != nil {
		t.Error("failed to get users:", err)
	}

	// check valid password
	matches, err := u.PasswordMatches("password")
	if err != nil {
		t.Error("error checking match:", err)
	}

	if !matches {
		t.Error("password does not match when it should")
	}

	// check invalid password
	matches, err = u.PasswordMatches("123")
	if err != nil {
		t.Error("error checking match:", err)
	}

	if matches {
		t.Error("password matches when it should not")
	}
}

func TestUser_ResetPassword(t *testing.T) {
	// reset password for existing user
	err := models.Users.ResetPassword(1, "new_password")
	if err != nil {
		t.Error("error resetting password for existing user:", err)
	}

	// reset password for non-existing user
	err = models.Users.ResetPassword(100, "new_password")
	if err == nil {
		t.Error("did not get an error when resetting a non-existing user's password:", err)
	}
}

func TestUser_Delete(t *testing.T) {
	err := models.Users.Delete(1)
	if err != nil {
		t.Error("failed to delete user:", err)
	}

	// try and get deleted user
	_, err = models.Users.Get(1)
	if err == nil {
		t.Error("retreived a deleted user:", err)
	}
}

func TestToken_Table(t *testing.T) {
	s := models.Tokens.Table()
	if s != "tokens" {
		t.Error("wrong table name returned for tokens")
	}
}

func TestToken_GenerateToken(t *testing.T) {
	id, err := models.Users.Insert(dummyUser)
	if err != nil {
		t.Error("error inserting user:", err)
	}

	_, err = models.Tokens.GenerateToken(id, time.Hour*24*365)
	if err != nil {
		t.Error("error generating token:", err)
	}
}

func TestToken_Insert(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	token, err := models.Tokens.GenerateToken(u.ID, time.Hour*24*365)
	if err != nil {
		t.Error("error generating token:", err)
	}
	err = models.Tokens.Insert(*token, *u)
	if err != nil {
		t.Error("error inserting token:", err)
	}
}

func TestToken_GetUserForToken(t *testing.T) {
	// first check non-existing token
	fakeToken := "abc"
	_, err := models.Tokens.GetUserForToken(fakeToken)
	if err == nil {
		t.Error("error expected but not received when getting a user from bad token:", err)
	}

	// check existing token by first getting the token from an existing user
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	_, err = u.Token.GetUserForToken(u.Token.PlainText)
	if err != nil {
		t.Error("failed to get user with valid token:", err)
	}
}

func TestToken_GetTokensForUser(t *testing.T) {
	tokens, err := models.Tokens.GetTokensForUser(1)
	if err != nil {
		t.Error(err)
	}

	if len(tokens) > 0 {
		t.Error("tokens returned for non-existing user")
	}
}

func TestToken_Get(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	_, err = models.Tokens.Get(u.Token.ID)
	if err != nil {
		t.Error("error getting token by id:", err)
	}
}

func TestToken_GetByToken(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	// try getting existing token
	_, err = models.Tokens.GetByToken(u.Token.PlainText)
	if err != nil {
		t.Error("error getting token by token:", err)
	}

	// try getting non-existing token
	_, err = models.Tokens.GetByToken("123")
	if err == nil {
		t.Error("error getting non-existing token by token:", err)
	}
}

var authData = []struct {
	name        string
	token       string
	email       string
	errExpected bool
	message     string
}{
	{"invalid", "abcdefghijklmnopqrstuvwxyz", "invalid-email@here.com", true, "invalid token accepted as valid"},
	{"invalid_length", "abcdefghijklmnopqrstuvwxyz", "invalid-email@here.com", true, "token of wronglength accepted as valid"},
	{"no_user", "abcdefghijklmnopqrstuvwxyz", "invalid-email@here.com", true, "no user but token accepted as valid"},
	{"valid", "abcdefghijklmnopqrstuvwxyz", "me@here.com", false, "valid token reported as invalid"},
}

// table function
func TestToken_AuthenticateToken(t *testing.T) {
	for _, tt := range authData {
		// token := ""
		var token Token
		if tt.email == dummyUser.Email {
			// user exists so get the user
			user, err := models.Users.GetByEmail(tt.email)
			if err != nil {
				t.Error("failed to get user:", err)
			}
			token.PlainText = user.Token.PlainText
		} else {
			token.PlainText = tt.token
		}

		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add("Authorization", "Bearer "+token.PlainText)

		_, err := token.AuthenticateToken(req)
		if tt.errExpected && err == nil {
			// ie if expected error but did not get one
			t.Errorf("%s: %s", tt.name, tt.message)
		} else if !tt.errExpected && err != nil {
			// ie if not expected error but got one
			t.Errorf("%s: %s - %s", tt.name, tt.message, err)
		} else {
			t.Logf("passed %s", tt.name)
		}
	}
}

func TestToken_Delete(t *testing.T) {

}
