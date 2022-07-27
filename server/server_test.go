package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/danny-m08/music-match/config"
	"github.com/danny-m08/music-match/neo4j"
	"github.com/danny-m08/music-match/types"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"
)

type jar struct {
	cookies []*http.Cookie
}

func (j *jar) SetCookies(_ *url.URL, cookies []*http.Cookie) {
	j.cookies = cookies
}

func (j *jar) Cookies(_ *url.URL) []*http.Cookie {
	return j.cookies
}

func RandomPort() string {
	rand.Seed(time.Now().UnixNano())
	return strconv.Itoa(8000 + rand.Intn(1000))
}

func Test_E2E(t *testing.T) {

	convey.Convey("Neo4j End to End testing...", t, func() {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		URL := "localhost:" + RandomPort()

		httpConfig := config.HTTPConfig{
			ListenAddr: URL,
			Secret:     "test_123",
		}

		URL = "http://" + URL
		neo4jMock := neo4j.NewMockNeo4jClient(ctrl)

		httpServer := &server{
			httpConfig:  &httpConfig,
			neo4jClient: neo4jMock,
			sessions:    make(map[string]string),
		}

		user := types.User{
			Name:     "Test User",
			Username: "testUser",
			Email:    "testUser@gmail.com",
			Password: "password",
		}

		client := http.Client{
			Jar: &jar{
				cookies: make([]*http.Cookie, 0),
			},
		}

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

		data, _ := json.Marshal(user)

		t.Run("StartServer", func(t *testing.T) {
			convey.Convey("If we start http server on a random port and a mock neo4j client, we should get no errors", t, func() {
				convey.So(httpServer.StartServer(), convey.ShouldBeNil)
			})
		})

		t.Run("NewUser", func(t *testing.T) {

			convey.Convey("If we try to create a new user with proper fields, we should get a new user and it should exist", t, func() {
				neo4jMock.EXPECT().InsertUser(gomock.Any()).Times(1).Return(nil)

				resp, err := client.Post(URL+"/signup", "application/json", bytes.NewReader(data))
				convey.So(err, convey.ShouldBeNil)
				convey.So(resp, convey.ShouldNotBeNil)
				convey.So(resp.StatusCode, convey.ShouldEqual, http.StatusOK)
				token, ok := httpServer.sessions[user.Username]
				convey.So(ok, convey.ShouldBeTrue)
				convey.So(token, convey.ShouldNotEqual, "")

				convey.So(len(resp.Cookies()), convey.ShouldEqual, len(setCookies))
				for _, cookie := range resp.Cookies() {
					convey.So(cookie.Name, convey.ShouldBeIn, setCookies)
				}
				client.Jar.SetCookies(&url.URL{}, resp.Cookies())
			})

			convey.Convey("If we try to create a new user and neo4j returns an errors, we should expect an error", t, func() {
				neo4jMock.EXPECT().InsertUser(gomock.Any()).Times(1).Return(errors.New("Neo4j error"))

				resp, err := http.Post(URL+"/signup", "application/json", bytes.NewReader(data))
				convey.So(err, convey.ShouldBeNil)
				convey.So(resp, convey.ShouldNotBeNil)
				convey.So(resp.StatusCode, convey.ShouldEqual, http.StatusInternalServerError)
			})
		})

		t.Run("Login", func(t *testing.T) {

			convey.Convey("If we try to login with email that already has a session, we should expect no errors", t, func() {
				loginData, _ := json.Marshal(types.LoginRequest{
					Login:    user.Username,
					Password: user.Password,
				})

				resp, err := client.Post(URL+"/login", "application/json", bytes.NewReader(loginData))
				convey.So(err, convey.ShouldBeNil)
				convey.So(resp, convey.ShouldNotBeNil)
				convey.So(resp.StatusCode, convey.ShouldEqual, http.StatusOK)
			})

			convey.Convey("If we try to login with username and neo4j returns a valid user, we should expect no errors", t, func() {
				u := types.User{
					Name:     user.Name,
					Username: user.Username,
					Email:    user.Email,
					Password: string(hashedPassword),
				}
				neo4jMock.EXPECT().GetUser(gomock.Any()).Times(1).Return(&u, nil)

				loginData, _ := json.Marshal(types.LoginRequest{
					Login:    user.Username,
					Password: user.Password,
				})

				resp, err := http.Post(URL+"/login", "application/json", bytes.NewReader(loginData))
				convey.So(err, convey.ShouldBeNil)
				convey.So(resp, convey.ShouldNotBeNil)
				convey.So(resp.StatusCode, convey.ShouldEqual, http.StatusOK)
			})

			convey.Convey("If we try to login with an invalid username, we should expect no errors", t, func() {
				neo4jMock.EXPECT().GetUser(gomock.Any()).Times(1).Return(nil, errors.New("Neo4j error"))

				loginData, _ := json.Marshal(types.LoginRequest{
					Login:    user.Username,
					Password: user.Password,
				})

				resp, err := http.Post(URL+"/login", "application/json", bytes.NewReader(loginData))
				convey.So(err, convey.ShouldBeNil)
				convey.So(resp, convey.ShouldNotBeNil)
				convey.So(resp.StatusCode, convey.ShouldEqual, http.StatusInternalServerError)
			})

			convey.Convey("If we try to login with username and neo4j returns no such user, we should expect an error", t, func() {
				neo4jMock.EXPECT().GetUser(gomock.Any()).Times(1).Return(nil, nil)

				loginData, _ := json.Marshal(types.LoginRequest{
					Login:    user.Username,
					Password: user.Password,
				})

				resp, err := http.Post(URL+"/login", "application/json", bytes.NewReader(loginData))
				convey.So(err, convey.ShouldBeNil)
				convey.So(resp, convey.ShouldNotBeNil)
				convey.So(resp.StatusCode, convey.ShouldEqual, http.StatusUnauthorized)
			})
		})

		t.Run("GetUserInfo", func(t *testing.T) {

			data, _ := json.Marshal(types.GetUserRequest{
				Login: "SomeUser",
			})

			convey.Convey("If we try to get user info without logging in we should get an error and nil user", t, func() {
				resp, err := http.Post(URL+"/getUser", "application/json", bytes.NewReader(data))
				convey.So(err, convey.ShouldBeNil)
				convey.So(resp, convey.ShouldNotBeNil)
				convey.So(resp.StatusCode, convey.ShouldEqual, http.StatusUnauthorized)

			})
		})
	})
}

func TestStartServerNoDB(t *testing.T) {
	convey.Convey("If we call StartServer with no DB we should get an error", t, func() {
		conf := &config.HTTPConfig{
			ListenAddr: "0.0.0.0:0",
		}

		dbConf := &config.Neo4jConfig{
			URI: "neo4j://localhost",
		}

		srver, err := NewServer(conf, dbConf)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(srver, convey.ShouldBeNil)
	})
}
