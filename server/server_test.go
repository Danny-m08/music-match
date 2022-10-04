package server

import (
	"testing"
	"time"

	"github.com/danny-m08/music-match/types"
	"github.com/smartystreets/goconvey/convey"
)

func testNewUser(t *testing.T) {

}

func TestJWT(t *testing.T) {
	expiry = time.Second

	convey.Convey("If we generate a valid JWT token we should get no error and a token that isn't empty", t, func() {
		s := server{secret: []byte("testSecret")}

		token, err := s.createJWT(&types.User{Email: "test@gmail.com"})
		convey.So(err, convey.ShouldBeNil)
		convey.So(token, convey.ShouldNotEqual, expiry)

		convey.Convey("If we try to validate this same JWT Token we should get no errors and it should be valid", func() {
			valid, err := s.validateJWT(token)
			convey.So(err, convey.ShouldBeNil)
			convey.So(valid, convey.ShouldBeTrue)
		})

		time.Sleep(time.Second)

		convey.Convey("If we try to use token after expiration we should get an error and no token", func() {
			valid, err := s.validateJWT(token)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(valid, convey.ShouldBeFalse)
		})
	})
}

func TestCreateJWTNoLogin(t *testing.T) {
	convey.Convey("If we try to generate JWT without username or email we should get an error", t, func() {
		s := server{secret: []byte("testSecret")}
		token, err := s.createJWT(&types.User{})
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(token, convey.ShouldEqual, "")
	})
}
