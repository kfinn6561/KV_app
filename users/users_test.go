package users_test

import (
	"fmt"
	"math"
	"store/logging"
	"store/users"
	"testing"
	"time"
)

type TestUser struct {
	username string
	password string
}

var testUsers = [4]TestUser{
	{username: "user_a", password: "passwordA"},
	{username: "user_b", password: "passwordB"},
	{username: "user_c", password: "passwordC"},
	{username: "admin", password: "Password1"},
}

func TestMain(m *testing.M) {
	logging.SetupLoggers("../info.log", "../htaccess.log") //pass in the log files so they can be closed at the end of the main function
	defer logging.Shutdown()
	users.FillUserDB()
	m.Run()
}

func TestFillUserDB(t *testing.T) {
	users.WipeUserDB()
	err := users.FillUserDB()
	if err != nil {
		t.Error("cannot fill database", err)
	}
}

func TestGoodPasswords(t *testing.T) {
	for _, user := range testUsers {
		ok := users.CheckUserPassword(user.username, user.password)
		if !ok {
			t.Errorf("unable to log in as user %s with password %s", user.username, user.password)
		}
	}
}

func TestBadPasswords(t *testing.T) {
	badPassword := "wrong"
	for _, user := range testUsers {
		ok := users.CheckUserPassword(user.username, badPassword)
		if ok {
			t.Errorf("able to log in as user %s with bad password %s", user.username, badPassword)
		}
	}
}

func TestBadUser(t *testing.T) {
	badUser := "wrong"
	badPassword := "wrong"
	ok := users.CheckUserPassword(badUser, badPassword)
	if ok {
		t.Errorf("able to log in as user %s with bad password %s\n", badUser, badPassword)
	}
}

func TestGoodJWTs(t *testing.T) {
	for _, user := range testUsers {
		token, err := users.GenerateJWT(user.username, user.password)
		if err != nil {
			t.Errorf("unable to generate a token for user %s with password %s\n", user.username, user.password)
		}
		username, ok := users.ValidateJWT(token)
		if username != user.username || !ok {
			t.Errorf("token generated for user %s with password %s was invalid and returned username%s\n", user.username, user.password, username)
		}
	}
}

func TestBadJWTs(t *testing.T) {
	badPassword := "wrong"
	for _, user := range testUsers {
		token, err := users.GenerateJWT(user.username, badPassword)
		if err == nil {
			t.Errorf("able to generate a token for user %s with password %s\n", user.username, badPassword)
		}
		username, ok := users.ValidateJWT(token)
		if username != "" || ok {
			t.Errorf("token generated for user %s with wrong password %s was valid. returned username %s\n", user.username, badPassword, username)
		}
	}
}

func TestRandomJWT(t *testing.T) {
	randomJWT := "sohubi233yhdweoufyecu7y31bceoriuwehr"
	username, ok := users.ValidateJWT(randomJWT)
	if username != "" || ok {
		t.Errorf("random JWT %s was marked as valid and returned %s\n", randomJWT, username)
	}
}

func TestConstantTime(t *testing.T) {
	//want to check if giving an invalid user is quicker than giving a valid one
	howStrict := 3.0 //test will fail if percentage variation between good and bad is worse than howStrict time the natural variation
	NRuns := 10
	badUser := "wrong"
	badPassword := "wrong"
	var start time.Time

	start = time.Now()
	for i := 0; i < NRuns; i++ {
		users.GenerateJWT(testUsers[0].username, badPassword)
	}
	goodTime1 := time.Since(start)

	start = time.Now()
	for i := 0; i < NRuns; i++ {
		users.GenerateJWT(testUsers[1].username, badPassword)
	}
	goodTime2 := time.Since(start)

	naturalVariation := math.Abs(float64(goodTime2-goodTime1)) / float64(goodTime1)
	fmt.Printf("Natural variation is %0.2f%%.\n", 100*naturalVariation)
	start = time.Now()
	for i := 0; i < NRuns; i++ {
		users.GenerateJWT(badUser, badPassword)
	}
	badTime := time.Since(start)

	badVariation := math.Abs(float64(badTime-goodTime1)) / float64(goodTime1)

	if naturalVariation*howStrict < badVariation {
		t.Errorf("Bad logins happen quicker than good logins. Bad login took %v, good logins took %v and %v. Natural variation is %0.2f%%, bad variation is %0.2f%%\n", badTime, goodTime1, goodTime2, 100*naturalVariation, 100*badVariation)
	}

}
