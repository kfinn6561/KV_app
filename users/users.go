package users

import (
	"encoding/csv"
	"errors"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"os"
	"store/logging"
	"time"
)

const UserDBFile = "../users/users.csv"

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var jwtKey = []byte("this is so secret you will never guess it")

var (
	ErrCannotBuildJWT = errors.New("cannot build JWT")
	ErrUnauthorised   = errors.New("unauthorised")
)

func readCsvFile(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.New("unable to read input file")
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		logging.ErrorLogger.Println("Unable to parse file as CSV for "+filePath, err)
		return nil, errors.New("invalid csv file")
	}

	return records, nil
}

type User struct {
	Username     string
	passwordHash string
}

func (u User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.passwordHash), []byte(password))
	return err == nil
}

func NewUser(username, password string) (*User, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return nil, err
	}
	user := User{
		Username:     username,
		passwordHash: string(bytes),
	}
	return &user, nil
}

var fakeUser, _ = NewUser("FaketasticMrFake", "fakeyfakeyfakefake")

var userDB = map[string]*User{} //username data will appear as both the key and in the user struct. Although this involves duplication it will greatly improve usability

func FillUserDB() error {
	users, err := readCsvFile(UserDBFile)
	if err != nil {
		return err
	}
	for _, user := range users {
		username := user[0]
		password := user[1]
		userStruct, err := NewUser(username, password)
		if err != nil {
			logging.ErrorLogger.Printf("error constructing the User struct for user with name %s and password %s\n", username, password)
			return err
		}
		userDB[username] = userStruct
	}
	logging.InfoLogger.Println("successfully built users database")
	return nil
}

func WipeUserDB() { //used for testing
	userDB = map[string]*User{}
}

func CheckUserPassword(username, password string) bool {
	user, present := userDB[username]
	if !present {
		fakeUser.CheckPassword("fakePassword") //perform a fake hashing check to make login constant time
		return false
	}
	return user.CheckPassword(password)
}

func GenerateJWT(username, password string) (string, error) {
	ok := CheckUserPassword(username, password)
	if !ok {
		return "", ErrUnauthorised
	}
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Issuer:    "KieranKVStore",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", ErrCannotBuildJWT
	}
	return tokenString, nil
}

func ValidateJWT(tokenString string) (string, bool) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
	if err != nil {
		return "", false
	}
	if !token.Valid {
		return "", false
	}
	username := token.Claims.(*Claims).Username
	return username, true
}
