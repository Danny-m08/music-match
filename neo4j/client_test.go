package neo4j_test

import (
	"fmt"
	"github.com/bojanz/currency"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/danny-m08/music-match/config"
	"github.com/danny-m08/music-match/neo4j"
	"github.com/danny-m08/music-match/types"
	"github.com/smartystreets/goconvey/convey"
)

func TestMain(m *testing.M) {

	output, err := exec.Command("docker-compose", "up", "-d", "neo4j", "init").CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("Unable to start neo4j docker container: %s: %s", string(output), err.Error()))
	}

	time.Sleep(15 * time.Second)

	res := m.Run()

	output, err = exec.Command("docker-compose", "down").CombinedOutput()
	if err != nil {
		fmt.Printf("Error removing neo4j docker container: %s: %s", output, err.Error())
	}

	os.Exit(res)
}

func TestE2E(t *testing.T) {

	convey.Convey("Neo4j End to End testing...", t, func() {
		var err error
		client := &neo4j.Client{}
		user := types.User{
			Username: "danielson",
			Password: "test1234",
			Email:    "danny@gmail.com",
		}

		follower := types.User{
			Username: "follower",
			Password: "follower123",
			Email:    "follower123@gmail.com",
		}

		price, _ := currency.NewAmount("25", "USD")
		now := time.Now()
		forSale := types.Listing{
			ID:    types.GenerateID(),
			Price: price,
			Track: &types.Track{
				Name: "testTrack",
				Path: "./testTrack.jpg",
			},
			Created: &now,
		}

		t.Run("NewClient", func(t *testing.T) {
			convey.Convey("If we create a new client and try to connect to the DB we should get no errors and a valid connection\n", t, func() {
				conf := &config.Neo4jConfig{
					URI:       "neo4j://localhost",
					Plaintext: true,
				}

				client, err = neo4j.NewClient(conf)
				convey.So(err, convey.ShouldBeNil)
				convey.So(client, convey.ShouldNotBeNil)
			})
		})

		t.Run("InsertUser", func(t *testing.T) {
			convey.Convey("If we try to insert a user into the database it should be successful\n", t, func() {
				convey.So(client.InsertUser(&user), convey.ShouldBeNil)
				convey.So(client.InsertUser(&follower), convey.ShouldBeNil)
			})
		})

		t.Run("InsertUserError", func(t *testing.T) {
			convey.Convey("If we try to create users whose email or username already exist we should get an error\n", t, func() {
				sameEmail := types.User{
					Username: "sameEmail",
					Email:    user.Email,
					Password: "sameEmail",
				}
				sameUsername := types.User{
					Username: user.Username,
					Email:    "sameusername@gmail.com",
					Password: "sameusername",
				}

				convey.So(client.InsertUser(&sameEmail), convey.ShouldNotBeNil)
				convey.So(client.InsertUser(&sameUsername), convey.ShouldNotBeNil)
			})
		})

		t.Run("RetrieveUser", func(t *testing.T) {
			convey.Convey("If we try to retrieve user data from the database then we should get no errors\n", t, func() {
				usr, err := client.GetUser(&user)
				convey.So(err, convey.ShouldBeNil)
				convey.So(*usr, convey.ShouldResemble, user)

				fllw, err := client.GetUser(&follower)
				convey.So(err, convey.ShouldBeNil)
				convey.So(*fllw, convey.ShouldResemble, follower)
			})
		})

		t.Run("CreateFollowing", func(t *testing.T) {
			convey.Convey(fmt.Sprintf("If we create a following from %s -> %s we should get no errors and proper structs\n", follower.String(), user.String()), t, func() {
				convey.So(client.CreateFollowing(&user, &follower), convey.ShouldBeNil)
			})
		})

		t.Run("GetFollowers", func(t *testing.T) {
			convey.Convey("If we try to retrieve followers for the users, we should get proper values and no error\n", t, func() {
				user.Password = ""
				follower.Password = ""

				followers, err := client.GetFollowers(&user)
				convey.So(err, convey.ShouldBeNil)
				convey.So(len(followers), convey.ShouldEqual, 1)
				convey.So(*followers[0], convey.ShouldResemble, follower)

				followers, err = client.GetFollowers(&follower)
				convey.So(err, convey.ShouldBeNil)
				convey.So(len(followers), convey.ShouldEqual, 0)
			})
		})

		t.Run("Unfollow", func(t *testing.T) {
			convey.Convey("If follower unfollows user, we should get no error and the user should no longer have any followers\n", t, func() {
				convey.So(client.Unfollow(&user, &follower), convey.ShouldBeNil)

				followers, err := client.GetFollowers(&user)
				convey.So(err, convey.ShouldBeNil)
				convey.So(len(followers), convey.ShouldEqual, 0)
			})
		})

		t.Run("CreateListing", func(t *testing.T) {
			convey.Convey("If a user creates a listing we should get no errors\n", t, func() {
				convey.So(client.CreateUserListing(&user, &forSale), convey.ShouldBeNil)

				isSold, err := client.IsSold(&forSale)
				convey.So(err, convey.ShouldBeNil)
				convey.So(isSold, convey.ShouldNotBeNil)
			})
		})

		t.Run("BuyListing", func(t *testing.T) {
			convey.Convey("If a user buys a listing then we should get no error\n", t, func() {
				convey.So(client.Sold(&follower, &forSale), convey.ShouldBeNil)

				isSold, err := client.IsSold(&forSale)
				convey.So(err, convey.ShouldBeNil)
				convey.So(isSold, convey.ShouldNotBeNil)
			})
		})

		t.Run("DeleteUser", func(t *testing.T) {
			convey.Convey("If we try to delete a user we should get no error and the user should no longer exist in the DB\n", t, func() {
				convey.So(client.DeleteUser(follower.Username, follower.Email), convey.ShouldBeNil)

				usr, err := client.GetUser(&follower)
				convey.So(err, convey.ShouldBeNil)
				convey.So(usr, convey.ShouldBeNil)

				followers, err := client.GetFollowers(&user)
				convey.So(err, convey.ShouldBeNil)
				convey.So(len(followers), convey.ShouldEqual, 0)
			})
		})

		t.Run("CloseSession", func(t *testing.T) {
			convey.Convey("If we close the client session we should get no errors\n", t, func() {
				convey.So(client.Close(), convey.ShouldBeNil)
			})
		})
	})

}
