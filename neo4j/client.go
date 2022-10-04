package neo4j

import (
	"errors"
	"fmt"
	"time"

	"github.com/danny-m08/music-match/config"
	"github.com/danny-m08/music-match/logging"
	"github.com/danny-m08/music-match/types"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type Client interface {
	GetUser(*types.User) (*types.User, error)
	CreateFollowing(*types.User, *types.User) error
	Unfollow(*types.User, *types.User) error
	GetFollowers(*types.User) (*[]types.User, error)
	DeleteUser(*types.User) error
	InsertUser(*types.User) error
}

type client struct {
	session neo4j.Session
	driver  neo4j.Driver
}

const (
	username = "username"
	email    = "email"
	password = "password"
	//id       = "id"
	//date     = "date"

	defaultDatabase = "neo4j"
	dateFormat      = "2006-01-02 15:04:05.999999999 -0700 MST"
)

// NewClient creates a new neo4j client using the specified config
func NewClient(conf *config.Neo4jConfig) (Client, error) {
	if conf == nil {
		return nil, errors.New("Neo4j config cannot be nil")
	}

	var auth neo4j.AuthToken
	if conf.Plaintext {
		auth = neo4j.NoAuth()
	} else {
		if conf.Username == "" || conf.Password == "" {
			return nil, errors.New("Username or password cannot be empty")
		}

		auth = neo4j.BasicAuth(conf.Username, conf.Password, "")
	}

	if conf.Database == "" {
		conf.Database = defaultDatabase
	}

	driver, err := neo4j.NewDriver(conf.URI, auth)
	if err != nil {
		return nil, err
	}

	err = driver.VerifyConnectivity()
	if err != nil {
		return nil, err
	}

	logging.Info(fmt.Sprintf("Creating Neo4j session to config %s at %s", conf.Database, conf.URI))
	session := driver.NewSession(neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: conf.Database,
		FetchSize:    conf.BatchSize,
	})

	return client{
		session: session,
		driver:  driver,
	}, nil
}

// //GetUser queries the DB for the user with the given types object
func (c client) GetUser(user *types.User) (*types.User, error) {
	query := fmt.Sprintf("MATCH (user:User) WHERE user.username = '%s' OR user.email = '%s' return user", user.Username, user.Email)

	records, err := c.readTransaction(query)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, nil
	}

	entry := records[0].Values[0].(neo4j.Node)
	return &types.User{
		Email:    entry.Props["email"].(string),
		Username: entry.Props["username"].(string),
		Password: entry.Props["password"].(string),
	}, nil
}

// CreateFollowing creates a follower relationship from user -> follower in the Neo4j DB
func (c client) CreateFollowing(user, follower *types.User) error {
	logging.Info(fmt.Sprintf("Creating Follower relationship with follower %s -> user %s", user.String(), follower.String()))
	query := `MATCH (user:User), (follower:User) WHERE (user.email = '%s' AND follower.email = '%s') OR (user.username = '%s' AND follower.username = '%s') CREATE (follower)-[f:FOLLOWS]->(user) return type(f)`
	query = fmt.Sprintf(query, user.Email, follower.Email, user.Username, follower.Username)
	_, err := c.writeTransaction(query)
	return err
}

// Unfollow removes the FOLLOWS relationship between the 2 users starting from follower -> user
func (c client) Unfollow(user, follower *types.User) error {
	logging.Info(fmt.Sprintf("Unfollow request to unfollow %s from %s", follower.String(), user.String()))
	query := `MATCH (follower:User { username: '%s' })-[f:FOLLOWS]->(user:User { username: '%s' }) DELETE f`
	query = fmt.Sprintf(query, follower.Username, user.Username)
	_, err := c.writeTransaction(query)
	return err
}

// GetFollowers queries the database for all followers for the given user
func (c client) GetFollowers(user *types.User) ([]*types.User, error) {
	logging.Info("Retrieving followers for " + user.String())
	users := make([]*types.User, 0)

	query := `MATCH (follower:User)-[f:FOLLOWS]->(user:User { username: '%s' }) return follower`
	query = fmt.Sprintf(query, user.Username)
	records, err := c.readTransaction(query)
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		user := record.Values[0].(neo4j.Node)
		users = append(users, &types.User{
			Email:    user.Props["email"].(string),
			Username: user.Props["username"].(string),
		})
	}

	return users, nil
}

// DeleteUser deletes a user from the database
func (c client) DeleteUser(username, email string) error {
	query := `Match (u:User {email: '%s'}) DETACH DELETE u`
	query = fmt.Sprintf(query, email)
	_, err := c.writeTransaction(query)
	return err
}

// Insert User inserts the user into the database
func (c client) InsertUser(user *types.User) error {
	query := fmt.Sprintf("CREATE (u:User %s)", user.String())

	_, err := c.writeTransaction(query)
	return err
}

func (c client) CreateListing(listing *types.Listing) error {
	query := "CREATE (l:Listing { id : '%s', price: '%s', date: '%s'}) return l"
	logging.Info("Creating new listing: " + listing.String())
	query = fmt.Sprintf(query, listing.ID, listing.Price.String(), listing.Created.String())
	_, err := c.writeTransaction(query)
	return err
}

// NewListing creates a new listing and sets the given user as the seller
func (c client) CreateUserListing(user *types.User, l *types.Listing) error {
	err := c.CreateListing(l)
	if err != nil {
		return err
	}

	query := `MATCH (u:User { username: '%s' }),(listing:Listing { id : '%s'}) CREATE (u)-[s:SELLING]->(listing) return s`
	query = fmt.Sprintf(query, user.Username, l.ID)
	_, err = c.writeTransaction(query)
	return err
}

//func (c client) GetListingsForUser(user *types.User) ([]*types.Listing, error) {
//	query := `MATCH (u:User { username: '%s'}), (l:Listing)-(u)-[SELLING]->(l) return l`
//	query = fmt.Sprintf(query, user.Username)
//	records, err := c.readTransaction(query)
//	if err != nil {
//		return nil, err
//	}
//
//	for _, record := range records {
//		node := record.Values[0].(neo4j.Node)
//		created, _ := time.Parse(dateFormat, node.Props["created"].(string))
//		l := &types.Listing{
//			ID:      node.Props["id"].(string),
//			Created: &created,
//			Track:
//		}
//	}
//}

// Sold marks the listing as sold in the DB
func (c client) Sold(user *types.User, l *types.Listing) error {
	query := `MATCH (u:User { username: '%s' }) CREATE (u)-[:BOUGHT {}]->(:Listing { id : '%s', date: '%s'})`
	query = fmt.Sprintf(query, user.Username, l.ID, l.Created.String())
	_, err := c.writeTransaction(query)
	return err
}

// IsSold checks if the given listing is sold and returns transaction details if sold
func (c client) IsSold(l *types.Listing) (*types.Transaction, error) {
	query := `MATCH (u:User)-[BOUGHT]->(l:Listing { id: '%s' }) return u, l`
	query = fmt.Sprintf(query, l.ID)
	records, err := c.readTransaction(query)
	if err != nil {
		return nil, err
	}

	tx := &types.Transaction{}
	for _, record := range records {
		buyer := record.Values[0].(neo4j.Node)
		txDetails := record.Values[1].(neo4j.Node)
		tx.ID = txDetails.Props["id"].(string)
		tx.Date, _ = time.Parse(dateFormat, txDetails.Props["date"].(string))
		tx.Buyer, err = getUser(&buyer, map[string]bool{})
		if err != nil {
			return nil, err
		}
	}

	return tx, nil
}

//func getTransaction(record *neo4j.Record, buyer, tx int) (*types.Transaction, error) {
//	var err error
//	res := &types.Transaction{}
//	b := record.Values[buyer].(neo4j.Node)
//	t := record.Values[tx].(neo4j.Relationship)
//
//	res.Buyer, err = getUser(&b, map[string]bool{
//		username: true,
//		email:    true,
//	})
//
//	if err != nil {
//		return nil, err
//	}
//
//	err = getTransactionDetails(res, &t)
//	if err != nil {
//		return nil, err
//	}
//
//	return res, nil
//}

// writeTransaction is a generic write operation on the database
func (c client) writeTransaction(query string) ([]*neo4j.Record, error) {
	records, err := c.session.WriteTransaction(
		func(tx neo4j.Transaction) (interface{}, error) {

			results, err := tx.Run(query, map[string]interface{}{})
			if err != nil {
				return nil, err
			}

			return results.Collect()
		})

	if err != nil {
		return nil, err
	}

	return records.([]*neo4j.Record), nil
}

func (c client) readTransaction(query string) ([]*neo4j.Record, error) {
	records, err := c.session.ReadTransaction(
		func(tx neo4j.Transaction) (interface{}, error) {

			results, err := tx.Run(query, map[string]interface{}{})
			if err != nil {
				return nil, err
			}

			return results.Collect()
		})

	if err != nil {
		return nil, err
	}

	return records.([]*neo4j.Record), nil
}

func getUser(node *neo4j.Node, required map[string]bool) (*types.User, error) {
	user := &types.User{}

	if node.Props[username] == nil {
		return nil, errors.New("Unable to retrieve username for user")
	}

	if node.Props[email] == nil {
		return nil, errors.New("Unable to retrieve email for user")
	}

	if node.Props[password] == nil && required[password] {
		user.Password = node.Props[password].(string)
		return nil, errors.New("Unable to retrieve password for user")
	}

	user.Username = node.Props[username].(string)
	user.Password = node.Props[email].(string)
	return user, nil
}

//func getTransactionDetails(tx *types.Transaction, relationship *neo4j.Relationship) error {
//	var err error
//
//	if relationship.Props[id] == nil {
//		return errors.New("unable to retrieve transaction ID")
//	}
//
//	if relationship.Props[date] == nil {
//		return errors.New("unable to retrieve transaction execution date")
//	}
//
//	tx.ID = relationship.Props[id].(string)
//	tx.Date, err = time.Parse(dateFormat, relationship.Props[date].(string))
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

func (c client) Close() error {
	err := c.session.Close()
	if err != nil {
		return err
	}

	return c.driver.Close()
}
