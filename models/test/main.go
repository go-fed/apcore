// apcore is a server framework for implementing an ActivityPub application.
// Copyright (C) 2020 Cory Slep
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"time"

	"github.com/go-fed/activity/pub"
	"github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/streams/vocab"
	"github.com/go-fed/apcore/framework/db"
	"github.com/go-fed/apcore/models"
	"github.com/go-fed/apcore/util"
	"github.com/go-fed/oauth2"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var dburl = flag.String("db", "", "database url to connect to")
var schema = flag.String("schema", "modeltest", "schema to use in the sql dialect")

var users = &models.Users{}
var fedData = &models.FedData{}
var localData = &models.LocalData{}
var inboxes = &models.Inboxes{}
var outboxes = &models.Outboxes{}
var deliveryAttempts = &models.DeliveryAttempts{}
var privateKeys = &models.PrivateKeys{}
var clientInfos = &models.ClientInfos{}
var tokenInfos = &models.TokenInfos{}
var credentials = &models.Credentials{}
var following = &models.Following{}
var followers = &models.Followers{}
var liked = &models.Liked{}
var policies = &models.Policies{}
var resolutions = &models.Resolutions{}
var testModels []models.Model

func init() {
	testModels = []models.Model{
		users,
		fedData,
		localData,
		inboxes,
		outboxes,
		deliveryAttempts,
		privateKeys,
		clientInfos,
		tokenInfos,
		credentials,
		following,
		followers,
		liked,
		policies,
		resolutions,
	}
}

func main() {
	flag.Parse()

	ctx := util.Context{context.Background()}
	db, err := connectPostgres(*dburl)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err = db.PingContext(ctx); err != nil {
		panic(err)
	}

	d := dialectPostgres(*schema)
	fmt.Println("Creating tables...")
	if err = createTables(ctx, db, d); err != nil {
		panic(err)
	}
	fmt.Println("Preparing statements...")
	if err = prepareStatements(ctx, db, d); err != nil {
		panic(err)
	}
	fmt.Println("Running UserModel calls...")
	if err = runUserModelCalls(ctx, db); err != nil {
		panic(err)
	}
	fmt.Println("Running FedData calls...")
	if err = runFedDataCalls(ctx, db); err != nil {
		panic(err)
	}
	fmt.Println("Running LocalData calls...")
	if err = runLocalDataCalls(ctx, db); err != nil {
		panic(err)
	}
	fmt.Println("Running Inboxes calls...")
	if err = runInboxesCalls(ctx, db); err != nil {
		panic(err)
	}
	fmt.Println("Running Outboxes calls...")
	if err = runOutboxesCalls(ctx, db); err != nil {
		panic(err)
	}
	fmt.Println("Running DeliveryAttempts calls...")
	if err = runDeliveryAttemptsCalls(ctx, db); err != nil {
		panic(err)
	}
	fmt.Println("Running PrivateKeys calls...")
	if err = runPrivateKeysCalls(ctx, db); err != nil {
		panic(err)
	}
	fmt.Println("Running ClientInfos calls...")
	clientInfoID, err := runClientInfosCalls(ctx, db)
	if err != nil {
		panic(err)
	}
	fmt.Println("Running TokenInfos calls...")
	if err = runTokenInfosCalls(ctx, db, clientInfoID); err != nil {
		panic(err)
	}
	fmt.Println("Running Credentials calls...")
	if err = runCredentialsCalls(ctx, db, clientInfoID); err != nil {
		panic(err)
	}
	fmt.Println("Running Followers calls...")
	if err = runFollowersCalls(ctx, db); err != nil {
		panic(err)
	}
	fmt.Println("Running Following calls...")
	if err = runFollowingCalls(ctx, db); err != nil {
		panic(err)
	}
	fmt.Println("Running Liked calls...")
	if err = runLikedCalls(ctx, db); err != nil {
		panic(err)
	}
	fmt.Println("Running Policies calls...")
	policyID, err := runPoliciesCalls(ctx, db)
	if err != nil {
		panic(err)
	}
	fmt.Println("Running Resolutions calls...")
	if err = runResolutionsCalls(ctx, db, policyID); err != nil {
		panic(err)
	}
	fmt.Println("Close models...")
	if err = closeModels(); err != nil {
		panic(err)
	}
	fmt.Println("done")
}

/* Resolutions */

func runResolutionsCalls(ctx util.Context, db *sql.DB, policyID string) error {
	if err := runResolutionsCreate(ctx, db, policyID); err != nil {
		return err
	}
	return nil
}

func runResolutionsCreate(ctx util.Context, db *sql.DB, policyID string) error {
	cr := models.CreateResolution{
		PolicyID: policyID,
		IRI:      mustParse(testActivity1IRI),
		R: models.Resolution{
			Time:    time.Now(),
			Matched: true,
			MatchLog: []string{
				"first entry",
				"second",
				"final line",
			},
		},
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return resolutions.Create(ctx, tx, cr)
	})
}

/* Policies */

func runPoliciesCalls(ctx util.Context, db *sql.DB) (policyID string, err error) {
	policyID, err = runPoliciesCreate(ctx, db)
	if err != nil {
		return
	}
	var po []models.PolicyAndPurpose
	po, err = runPoliciesGetForActor(ctx, db)
	if err != nil {
		return
	}
	fmt.Printf("> GetForActor: %v\n", po)
	var pd []models.PolicyAndID
	pd, err = runPoliciesGetForActorAndPurpose(ctx, db)
	if err != nil {
		return
	}
	fmt.Printf("> GetForActorAndPurpose: %v\n", pd)
	return
}

func runPoliciesCreate(ctx util.Context, db *sql.DB) (policyID string, err error) {
	cp := models.CreatePolicy{
		ActorID: mustParse(testActor1IRI),
		Purpose: models.FederatedBlockPurpose,
		Policy: models.Policy{
			Name:        "Test Policy 1",
			Description: "A test policy.",
			Matchers: []*models.KVMatcher{
				{
					KeyPathQuery: "actor",
					ValueMatcher: &models.UnaryMatcher{
						Value: &models.Value{
							EqualsString: testActor3IRI,
						},
					},
				},
			},
		},
	}
	return policyID, doWithTx(ctx, db, func(tx *sql.Tx) error {
		policyID, err = policies.Create(ctx, tx, cp)
		return err
	})
}

func runPoliciesGetForActor(ctx util.Context, db *sql.DB) (p []models.PolicyAndPurpose, err error) {
	return p, doWithTx(ctx, db, func(tx *sql.Tx) error {
		p, err = policies.GetForActor(ctx, tx, mustParse(testActor1IRI))
		return err
	})
}

func runPoliciesGetForActorAndPurpose(ctx util.Context, db *sql.DB) (p []models.PolicyAndID, err error) {
	return p, doWithTx(ctx, db, func(tx *sql.Tx) error {
		p, err = policies.GetForActorAndPurpose(ctx, tx, mustParse(testActor1IRI), models.FederatedBlockPurpose)
		return err
	})
}

/* Liked */

func runLikedCalls(ctx util.Context, db *sql.DB) error {
	if err := runLikedCreate(ctx, db); err != nil {
		return err
	}
	has, err := runLikedContainsForActorTrue(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> ContainsForActorTrue: %v\n", has)
	has, err = runLikedContainsForActorFalse(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> ContainsForActorFalse: %v\n", has)
	has, err = runLikedContainsTrue(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> ContainsTrue: %v\n", has)
	has, err = runLikedContainsFalse(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> ContainsFalse: %v\n", has)
	p, isEnd, err := runLikedGetPage(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> GetPage(%d, %d): %s %v\n", 21, 27, p, isEnd)
	if pb, err := toJSON(p); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	p, startIdx, err := runLikedGetLastPage(ctx, db, 19)
	if err != nil {
		return err
	}
	fmt.Printf("> GetLastPage(%d): %s %v\n", 19, p, startIdx)
	if pb, err := toJSON(p); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	if err := runLikedPrependItem(ctx, db); err != nil {
		return err
	}
	if err := runLikedDeleteItem(ctx, db); err != nil {
		return err
	}
	c, err := runLikedGetAllForActor(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> GetAllForActor: %s\n", c)
	if pb, err := toJSON(c); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	return nil
}

func runLikedCreate(ctx util.Context, db *sql.DB) error {
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return liked.Create(ctx, tx, mustParse(testActor1IRI), testActor1Liked)
	})
}

func runLikedContainsForActorTrue(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = liked.ContainsForActor(ctx, tx, mustParse(testActor1IRI), mustParse(testActivity2IRI))
		return err
	})
}

func runLikedContainsForActorFalse(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = liked.ContainsForActor(ctx, tx, mustParse(testActor1IRI), mustParse(testActivity1IRI))
		return err
	})
}

func runLikedContainsTrue(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = liked.Contains(ctx, tx, mustParse(testActor1LikedIRI), mustParse(testActivity2IRI))
		return err
	})
}

func runLikedContainsFalse(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = liked.Contains(ctx, tx, mustParse(testActor1LikedIRI), mustParse(testActivity1IRI))
		return err
	})
}

func runLikedGetPage(ctx util.Context, db *sql.DB) (p models.ActivityStreamsCollectionPage, isEnd bool, err error) {
	if err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		return liked.Create(ctx, tx, mustParse(testActor2IRI), testActor2Liked)
	}); err != nil {
		return
	}
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		p, isEnd, err = liked.GetPage(ctx, tx, mustParse(testActor2LikedIRI), 21, 27)
		return err
	})
	return
}

func runLikedGetLastPage(ctx util.Context, db *sql.DB, n int) (p models.ActivityStreamsCollectionPage, idx int, err error) {
	return p, idx, doWithTx(ctx, db, func(tx *sql.Tx) error {
		p, idx, err = liked.GetLastPage(ctx, tx, mustParse(testActor2LikedIRI), n)
		return err
	})
}

func runLikedPrependItem(ctx util.Context, db *sql.DB) error {
	if err := doWithTx(ctx, db, func(tx *sql.Tx) error {
		return liked.Create(ctx, tx, mustParse(testActor3IRI), testActor3Liked)
	}); err != nil {
		return err
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return liked.PrependItem(ctx, tx, mustParse(testActor3LikedIRI), mustParse(testActivity2IRI))
	})
}

func runLikedDeleteItem(ctx util.Context, db *sql.DB) error {
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return liked.DeleteItem(ctx, tx, mustParse(testActor1LikedIRI), mustParse(testActivity2IRI))
	})
}

func runLikedGetAllForActor(ctx util.Context, db *sql.DB) (p models.ActivityStreamsCollection, err error) {
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		p, err = liked.GetAllForActor(ctx, tx, mustParse(testActor2IRI))
		return err
	})
	return
}

/* Following */

func runFollowingCalls(ctx util.Context, db *sql.DB) error {
	if err := runFollowingCreate(ctx, db); err != nil {
		return err
	}
	has, err := runFollowingContainsForActorTrue(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> ContainsForActorTrue: %v\n", has)
	has, err = runFollowingContainsForActorFalse(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> ContainsForActorFalse: %v\n", has)
	has, err = runFollowingContainsTrue(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> ContainsTrue: %v\n", has)
	has, err = runFollowingContainsFalse(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> ContainsFalse: %v\n", has)
	p, isEnd, err := runFollowingGetPage(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> GetPage(%d, %d): %s %v\n", 21, 27, p, isEnd)
	if pb, err := toJSON(p); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	p, startIdx, err := runFollowingGetLastPage(ctx, db, 19)
	if err != nil {
		return err
	}
	fmt.Printf("> GetLastPage(%d): %s %v\n", 19, p, startIdx)
	if pb, err := toJSON(p); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	if err := runFollowingPrependItem(ctx, db); err != nil {
		return err
	}
	if err := runFollowingDeleteItem(ctx, db); err != nil {
		return err
	}
	c, err := runFollowingGetAllForActor(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> GetAllForActor: %s\n", c)
	if pb, err := toJSON(c); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	return nil
}

func runFollowingCreate(ctx util.Context, db *sql.DB) error {
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return following.Create(ctx, tx, mustParse(testActor1IRI), testActor1Following)
	})
}

func runFollowingContainsForActorTrue(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = following.ContainsForActor(ctx, tx, mustParse(testActor1IRI), mustParse(testActor2IRI))
		return err
	})
}

func runFollowingContainsForActorFalse(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = following.ContainsForActor(ctx, tx, mustParse(testActor1IRI), mustParse(testActor1IRI))
		return err
	})
}

func runFollowingContainsTrue(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = following.Contains(ctx, tx, mustParse(testActor1FollowingIRI), mustParse(testActor2IRI))
		return err
	})
}

func runFollowingContainsFalse(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = following.Contains(ctx, tx, mustParse(testActor1FollowingIRI), mustParse(testActor1IRI))
		return err
	})
}

func runFollowingGetPage(ctx util.Context, db *sql.DB) (p models.ActivityStreamsCollectionPage, isEnd bool, err error) {
	if err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		return following.Create(ctx, tx, mustParse(testActor2IRI), testActor2Following)
	}); err != nil {
		return
	}
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		p, isEnd, err = following.GetPage(ctx, tx, mustParse(testActor2FollowingIRI), 21, 27)
		return err
	})
	return
}

func runFollowingGetLastPage(ctx util.Context, db *sql.DB, n int) (p models.ActivityStreamsCollectionPage, idx int, err error) {
	return p, idx, doWithTx(ctx, db, func(tx *sql.Tx) error {
		p, idx, err = following.GetLastPage(ctx, tx, mustParse(testActor2FollowingIRI), n)
		return err
	})
}

func runFollowingPrependItem(ctx util.Context, db *sql.DB) error {
	if err := doWithTx(ctx, db, func(tx *sql.Tx) error {
		return following.Create(ctx, tx, mustParse(testActor3IRI), testActor3Following)
	}); err != nil {
		return err
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return following.PrependItem(ctx, tx, mustParse(testActor3FollowingIRI), mustParse(testActor2IRI))
	})
}

func runFollowingDeleteItem(ctx util.Context, db *sql.DB) error {
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return following.DeleteItem(ctx, tx, mustParse(testActor1FollowingIRI), mustParse(testActor2IRI))
	})
}

func runFollowingGetAllForActor(ctx util.Context, db *sql.DB) (p models.ActivityStreamsCollection, err error) {
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		p, err = following.GetAllForActor(ctx, tx, mustParse(testActor2IRI))
		return err
	})
	return
}

/* Followers */

func runFollowersCalls(ctx util.Context, db *sql.DB) error {
	if err := runFollowersCreate(ctx, db); err != nil {
		return err
	}
	has, err := runFollowersContainsForActorTrue(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> ContainsForActorTrue: %v\n", has)
	has, err = runFollowersContainsForActorFalse(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> ContainsForActorFalse: %v\n", has)
	has, err = runFollowersContainsTrue(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> ContainsTrue: %v\n", has)
	has, err = runFollowersContainsFalse(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> ContainsFalse: %v\n", has)
	p, isEnd, err := runFollowersGetPage(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> GetPage(%d, %d): %s %v\n", 21, 27, p, isEnd)
	if pb, err := toJSON(p); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	p, startIdx, err := runFollowersGetLastPage(ctx, db, 19)
	if err != nil {
		return err
	}
	fmt.Printf("> GetLastPage(%d): %s %v\n", 19, p, startIdx)
	if pb, err := toJSON(p); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	if err := runFollowersPrependItem(ctx, db); err != nil {
		return err
	}
	if err := runFollowersDeleteItem(ctx, db); err != nil {
		return err
	}
	c, err := runFollowersGetAllForActor(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> GetAllForActor: %s\n", c)
	if pb, err := toJSON(c); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	return nil
}

func runFollowersCreate(ctx util.Context, db *sql.DB) error {
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return followers.Create(ctx, tx, mustParse(testActor1IRI), testActor1Followers)
	})
}

func runFollowersContainsForActorTrue(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = followers.ContainsForActor(ctx, tx, mustParse(testActor1IRI), mustParse(testActor2IRI))
		return err
	})
}

func runFollowersContainsForActorFalse(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = followers.ContainsForActor(ctx, tx, mustParse(testActor1IRI), mustParse(testActor1IRI))
		return err
	})
}

func runFollowersContainsTrue(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = followers.Contains(ctx, tx, mustParse(testActor1FollowersIRI), mustParse(testActor2IRI))
		return err
	})
}

func runFollowersContainsFalse(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = followers.Contains(ctx, tx, mustParse(testActor1FollowersIRI), mustParse(testActor1IRI))
		return err
	})
}

func runFollowersGetPage(ctx util.Context, db *sql.DB) (p models.ActivityStreamsCollectionPage, isEnd bool, err error) {
	if err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		return followers.Create(ctx, tx, mustParse(testActor2IRI), testActor2Followers)
	}); err != nil {
		return
	}
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		p, isEnd, err = followers.GetPage(ctx, tx, mustParse(testActor2FollowersIRI), 21, 27)
		return err
	})
	return
}

func runFollowersGetLastPage(ctx util.Context, db *sql.DB, n int) (p models.ActivityStreamsCollectionPage, idx int, err error) {
	return p, idx, doWithTx(ctx, db, func(tx *sql.Tx) error {
		p, idx, err = followers.GetLastPage(ctx, tx, mustParse(testActor2FollowersIRI), n)
		return err
	})
}

func runFollowersPrependItem(ctx util.Context, db *sql.DB) error {
	if err := doWithTx(ctx, db, func(tx *sql.Tx) error {
		return followers.Create(ctx, tx, mustParse(testActor3IRI), testActor3Followers)
	}); err != nil {
		return err
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return followers.PrependItem(ctx, tx, mustParse(testActor3FollowersIRI), mustParse(testActor2IRI))
	})
}

func runFollowersDeleteItem(ctx util.Context, db *sql.DB) error {
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return followers.DeleteItem(ctx, tx, mustParse(testActor1FollowersIRI), mustParse(testActor2IRI))
	})
}

func runFollowersGetAllForActor(ctx util.Context, db *sql.DB) (p models.ActivityStreamsCollection, err error) {
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		p, err = followers.GetAllForActor(ctx, tx, mustParse(testActor2IRI))
		return err
	})
	return
}

/* Credentials */

func runCredentialsCalls(ctx util.Context, db *sql.DB, clientID string) error {
	id, err := runCredentialsCreate(ctx, db, clientID)
	if err != nil {
		return err
	}
	fmt.Printf("> Create: %v\n", id)
	if err := runCredentialsUpdate(ctx, db, id, clientID); err != nil {
		return err
	}
	if err := runCredentialsUpdateExpires(ctx, db, id); err != nil {
		return err
	}
	if err := runCredentialsDelete(ctx, db, clientID); err != nil {
		return err
	}
	ti, err := runCredentialsGetTokenInfo(ctx, db, id)
	if err != nil {
		return err
	}
	fmt.Printf("> GetTokenInfo: %v\n", ti)
	if err := runCredentialsDeleteExpired(ctx, db, clientID); err != nil {
		return err
	}
	return nil
}

func runCredentialsCreate(ctx util.Context, db *sql.DB, clientID string) (id string, err error) {
	uid, err := getUserID(ctx, db)
	if err != nil {
		return "", err
	}
	tis := []*models.TokenInfo{
		{
			ClientID:    clientID,
			UserID:      uid,
			RedirectURI: "cred_redirect1",
			Scope:       "cred_scope1",
			Code:        sql.NullString{"cred_code1", true},
			Access:      sql.NullString{"cred_access1", true},
			Refresh:     sql.NullString{"cred_refresh1", true},
		},
		{
			ClientID:    clientID,
			UserID:      uid,
			RedirectURI: "cred_redirect2",
			Scope:       "cred_scope2",
			Code:        sql.NullString{"cred_code2", true},
			Access:      sql.NullString{"cred_access2", true},
			Refresh:     sql.NullString{"cred_refresh2", true},
		},
	}
	for i, ti := range tis {
		var tid string
		if err := doWithTx(ctx, db, func(tx *sql.Tx) error {
			tid, err = tokenInfos.Create(ctx, tx, ti)
			return err
		}); err != nil {
			return "", err
		}
		var cid string
		err := doWithTx(ctx, db, func(tx *sql.Tx) error {
			cid, err = credentials.Create(ctx, tx, uid, tid, time.Now().Add(time.Minute*10))
			return err
		})
		if err != nil {
			return "", err
		}
		if i == 0 {
			id = cid
		}
	}
	return
}

func runCredentialsUpdate(ctx util.Context, db *sql.DB, id, clientID string) error {
	uid, err := getUserID(ctx, db)
	if err != nil {
		return err
	}
	ti := &models.TokenInfo{
		ClientID:    clientID,
		UserID:      uid,
		RedirectURI: "cred_redirect1_updated",
		Scope:       "cred_scope1_updated",
		Code:        sql.NullString{"cred_code1_updated", true},
		Access:      sql.NullString{"cred_access1_updated", true},
		Refresh:     sql.NullString{"cred_refresh1_updated", true},
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return credentials.Update(ctx, tx, id, ti)
	})
}

func runCredentialsUpdateExpires(ctx util.Context, db *sql.DB, id string) error {
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return credentials.UpdateExpires(ctx, tx, id, time.Now().Add(time.Hour))
	})
}

func runCredentialsDelete(ctx util.Context, db *sql.DB, clientID string) error {
	uid, err := getUserID(ctx, db)
	if err != nil {
		return err
	}
	ti := &models.TokenInfo{
		ClientID:    clientID,
		UserID:      uid,
		RedirectURI: "cred_redir_should_delete",
		Scope:       "cred_scope_should_delete",
		Code:        sql.NullString{"cred_code_should_delete", true},
		Access:      sql.NullString{"cred_access_should_delete", true},
		Refresh:     sql.NullString{"cred_refresh_should_delete", true},
	}
	var tid string
	if err := doWithTx(ctx, db, func(tx *sql.Tx) error {
		tid, err = tokenInfos.Create(ctx, tx, ti)
		return err
	}); err != nil {
		return err
	}
	var id string
	if err := doWithTx(ctx, db, func(tx *sql.Tx) error {
		id, err = credentials.Create(ctx, tx, uid, tid, time.Now())
		return err
	}); err != nil {
		return err
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return credentials.Delete(ctx, tx, id)
	})
}

func runCredentialsGetTokenInfo(ctx util.Context, db *sql.DB, id string) (ti oauth2.TokenInfo, err error) {
	return ti, doWithTx(ctx, db, func(tx *sql.Tx) error {
		ti, err = credentials.GetTokenInfo(ctx, tx, id)
		return err
	})
}

func runCredentialsDeleteExpired(ctx util.Context, db *sql.DB, clientID string) error {
	uid, err := getUserID(ctx, db)
	if err != nil {
		return err
	}
	tis := []*models.TokenInfo{
		{
			ClientID:    clientID,
			UserID:      uid,
			RedirectURI: "cred_redir_should_delete1",
			Scope:       "cred_scope_should_delete1",
			Code:        sql.NullString{"cred_code_should_delete1", true},
			Access:      sql.NullString{"cred_access_should_delete1", true},
			Refresh:     sql.NullString{"cred_refresh_should_delete1", true},
		},
		{
			ClientID:    clientID,
			UserID:      uid,
			RedirectURI: "cred_redir_should_not_delete1",
			Scope:       "cred_scope_should_not_delete1",
			Code:        sql.NullString{"cred_code_should_not_delete1", true},
			Access:      sql.NullString{"cred_access_should_not_delete1", true},
			Refresh:     sql.NullString{"cred_refresh_should_not_delete1", true},
		},
		{
			ClientID:    clientID,
			UserID:      uid,
			RedirectURI: "cred_redir_should_delete2",
			Scope:       "cred_scope_should_delete2",
			Code:        sql.NullString{"cred_code_should_delete2", true},
			Access:      sql.NullString{"cred_access_should_delete2", true},
			Refresh:     sql.NullString{"cred_refresh_should_delete2", true},
		},
	}
	times := []time.Time{
		time.Now().Add(-time.Second),
		time.Now().Add(5 * time.Second),
		time.Now().Add(-5 * time.Minute),
	}
	for i, ti := range tis {
		var tid string
		if err := doWithTx(ctx, db, func(tx *sql.Tx) error {
			tid, err = tokenInfos.Create(ctx, tx, ti)
			return err
		}); err != nil {
			return err
		}
		if err := doWithTx(ctx, db, func(tx *sql.Tx) error {
			_, err = credentials.Create(ctx, tx, uid, tid, times[i])
			return err
		}); err != nil {
			return err
		}
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return credentials.DeleteExpired(ctx, tx)
	})
}

/* TokenInfos */

func runTokenInfosCalls(ctx util.Context, db *sql.DB, clientID string) error {
	if id, err := runTokenInfosCreate(ctx, db, clientID); err != nil {
		return err
	} else {
		fmt.Printf("> Create: %v\n", id)
	}
	if err := runTokenInfosRemoveByCode(ctx, db, clientID); err != nil {
		return err
	}
	if err := runTokenInfosRemoveByAccess(ctx, db, clientID); err != nil {
		return err
	}
	if err := runTokenInfosRemoveByRefresh(ctx, db, clientID); err != nil {
		return err
	}
	ti, err := runTokenInfosGetByCode(ctx, db, clientID)
	if err != nil {
		return err
	}
	fmt.Printf("> GetByCode: %v\n", ti)
	ti, err = runTokenInfosGetByAccess(ctx, db, clientID)
	if err != nil {
		return err
	}
	fmt.Printf("> GetByAccess: %v\n", ti)
	ti, err = runTokenInfosGetByRefresh(ctx, db, clientID)
	if err != nil {
		return err
	}
	fmt.Printf("> GetByRefresh: %v\n", ti)
	return nil
}

func runTokenInfosCreate(ctx util.Context, db *sql.DB, clientID string) (id string, err error) {
	uid, err := getUserID(ctx, db)
	if err != nil {
		return "", err
	}
	ti := &models.TokenInfo{
		ClientID:    clientID,
		UserID:      uid,
		RedirectURI: "redirect1",
		Scope:       "scope1",
		Code:        sql.NullString{"code1", true},
		Access:      sql.NullString{"access1", true},
		Refresh:     sql.NullString{"refresh1", true},
	}
	return id, doWithTx(ctx, db, func(tx *sql.Tx) error {
		id, err = tokenInfos.Create(ctx, tx, ti)
		return err
	})
}

func runTokenInfosRemoveByCode(ctx util.Context, db *sql.DB, clientID string) error {
	uid, err := getUserID(ctx, db)
	if err != nil {
		return err
	}
	code := "code_should_be_deleted"
	ti := &models.TokenInfo{
		ClientID:    clientID,
		UserID:      uid,
		RedirectURI: "redirect2",
		Scope:       "scope2",
		Code:        sql.NullString{code, true},
		CodeCreated: sql.NullTime{time.Now(), true},
		CodeExpires: models.NullDuration{5 * time.Second, true},
	}
	if err := doWithTx(ctx, db, func(tx *sql.Tx) error {
		_, err = tokenInfos.Create(ctx, tx, ti)
		return err
	}); err != nil {
		return err
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return tokenInfos.RemoveByCode(ctx, tx, code)
	})
}

func runTokenInfosRemoveByAccess(ctx util.Context, db *sql.DB, clientID string) error {
	uid, err := getUserID(ctx, db)
	if err != nil {
		return err
	}
	access := "access_should_be_deleted"
	ti := &models.TokenInfo{
		ClientID:      clientID,
		UserID:        uid,
		RedirectURI:   "redirect3",
		Scope:         "scope3",
		Access:        sql.NullString{access, true},
		AccessCreated: sql.NullTime{time.Now(), true},
		AccessExpires: models.NullDuration{6 * time.Second, true},
	}
	if err := doWithTx(ctx, db, func(tx *sql.Tx) error {
		_, err = tokenInfos.Create(ctx, tx, ti)
		return err
	}); err != nil {
		return err
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return tokenInfos.RemoveByAccess(ctx, tx, access)
	})
}

func runTokenInfosRemoveByRefresh(ctx util.Context, db *sql.DB, clientID string) error {
	uid, err := getUserID(ctx, db)
	if err != nil {
		return err
	}
	refresh := "refresh_should_be_deleted"
	ti := &models.TokenInfo{
		ClientID:       clientID,
		UserID:         uid,
		RedirectURI:    "redirect4",
		Scope:          "scope4",
		Refresh:        sql.NullString{refresh, true},
		RefreshCreated: sql.NullTime{time.Now(), true},
		RefreshExpires: models.NullDuration{7 * time.Second, true},
	}
	if err := doWithTx(ctx, db, func(tx *sql.Tx) error {
		_, err = tokenInfos.Create(ctx, tx, ti)
		return err
	}); err != nil {
		return err
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return tokenInfos.RemoveByRefresh(ctx, tx, refresh)
	})
}

func runTokenInfosGetByCode(ctx util.Context, db *sql.DB, clientID string) (ti oauth2.TokenInfo, err error) {
	var uid string
	uid, err = getUserID(ctx, db)
	if err != nil {
		return
	}
	code := "code5"
	pti := &models.TokenInfo{
		ClientID:    clientID,
		UserID:      uid,
		RedirectURI: "redirect5",
		Scope:       "scope5",
		Code:        sql.NullString{code, true},
		CodeCreated: sql.NullTime{time.Now(), true},
		CodeExpires: models.NullDuration{8 * time.Second, true},
	}
	if err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		_, err = tokenInfos.Create(ctx, tx, pti)
		return err
	}); err != nil {
		return
	}
	return ti, doWithTx(ctx, db, func(tx *sql.Tx) error {
		ti, err = tokenInfos.GetByCode(ctx, tx, code)
		return err
	})
}

func runTokenInfosGetByAccess(ctx util.Context, db *sql.DB, clientID string) (ti oauth2.TokenInfo, err error) {
	var uid string
	uid, err = getUserID(ctx, db)
	if err != nil {
		return
	}
	access := "access6"
	pti := &models.TokenInfo{
		ClientID:      clientID,
		UserID:        uid,
		RedirectURI:   "redirect6",
		Scope:         "scope6",
		Access:        sql.NullString{access, true},
		AccessCreated: sql.NullTime{time.Now(), true},
		AccessExpires: models.NullDuration{9 * time.Second, true},
	}
	if err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		_, err = tokenInfos.Create(ctx, tx, pti)
		return err
	}); err != nil {
		return
	}
	return ti, doWithTx(ctx, db, func(tx *sql.Tx) error {
		ti, err = tokenInfos.GetByAccess(ctx, tx, access)
		return err
	})
}

func runTokenInfosGetByRefresh(ctx util.Context, db *sql.DB, clientID string) (ti oauth2.TokenInfo, err error) {
	var uid string
	uid, err = getUserID(ctx, db)
	if err != nil {
		return
	}
	refresh := "refresh7"
	pti := &models.TokenInfo{
		ClientID:       clientID,
		UserID:         uid,
		RedirectURI:    "redirect7",
		Scope:          "scope7",
		Refresh:        sql.NullString{refresh, true},
		RefreshCreated: sql.NullTime{time.Now(), true},
		RefreshExpires: models.NullDuration{10 * time.Second, true},
	}
	if err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		_, err = tokenInfos.Create(ctx, tx, pti)
		return err
	}); err != nil {
		return
	}
	return ti, doWithTx(ctx, db, func(tx *sql.Tx) error {
		ti, err = tokenInfos.GetByRefresh(ctx, tx, refresh)
		return err
	})
}

/* ClientInfos */

func runClientInfosCalls(ctx util.Context, db *sql.DB) (string, error) {
	id, err := runClientInfosCreate(ctx, db)
	if err != nil {
		return id, err
	}
	ci, err := runClientInfosGetByID(ctx, db, id)
	if err != nil {
		return id, err
	}
	fmt.Printf("> GetByUserID: %v\n", ci)
	return id, nil
}

func runClientInfosCreate(ctx util.Context, db *sql.DB) (id string, err error) {
	var uid string
	uid, err = getUserID(ctx, db)
	if err != nil {
		return
	}
	ci := &models.ClientInfo{
		ID:     "ci_id",
		Secret: sql.NullString{"ci_secret", true},
		Domain: "ci_domain",
		UserID: uid,
	}
	return id, doWithTx(ctx, db, func(tx *sql.Tx) error {
		id, err = clientInfos.Create(ctx, tx, ci)
		return err
	})
}

func runClientInfosGetByID(ctx util.Context, db *sql.DB, id string) (ci oauth2.ClientInfo, err error) {
	return ci, doWithTx(ctx, db, func(tx *sql.Tx) error {
		ci, err = clientInfos.GetByID(ctx, tx, id)
		return err
	})
}

/* PrivateKeys */

func runPrivateKeysCalls(ctx util.Context, db *sql.DB) error {
	if err := runPrivateKeysCreate(ctx, db); err != nil {
		return err
	}
	b, err := runPrivateKeysGetByUserID(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> GetByUserID: %v\n", b)
	b, err = runPrivateKeysGetForInstanceActor(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> GetForInstanceActor: %v\n", b)
	return nil
}

func runPrivateKeysCreate(ctx util.Context, db *sql.DB) error {
	id, err := getUserID(ctx, db)
	if err != nil {
		return err
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return privateKeys.Create(ctx, tx, id, "test", []byte{9, 8, 7, 6, 5, 4, 3, 2, 1, 0})
	})
}

func runPrivateKeysGetByUserID(ctx util.Context, db *sql.DB) (b []byte, err error) {
	id, err := getUserID(ctx, db)
	if err != nil {
		return nil, err
	}
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = privateKeys.GetByUserID(ctx, tx, id, "test")
		return err
	})
}

func runPrivateKeysGetForInstanceActor(ctx util.Context, db *sql.DB) (b []byte, err error) {
	id, err := getInstanceActorUserID(ctx, db)
	if err != nil {
		return nil, err
	}
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		return privateKeys.Create(ctx, tx, id, "test", []byte{1, 1, 0, 0, 2, 2, 4, 4, 3, 3})
	})
	if err != nil {
		return
	}
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = privateKeys.GetInstanceActor(ctx, tx, "test")
		return err
	})
}

/* DeliveryAttempts */

func runDeliveryAttemptsCalls(ctx util.Context, db *sql.DB) error {
	id, err := runDeliveryAttemptsCreate(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> Create: %v\n", id)
	if err := runDeliveryAttemptsMarkSuccessful(ctx, db); err != nil {
		return err
	}
	if err := runDeliveryAttemptsMarkFailed(ctx, db); err != nil {
		return err
	}
	if err := runDeliveryAttemptsMarkAbandoned(ctx, db); err != nil {
		return err
	}
	rf, ft, err := runDeliveryAttemptsFirstPage(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> FirstPage: len=%d\n", len(rf))
	for i, r := range rf {
		fmt.Printf("> [%d]=%v\n", i, r)
	}
	for k := 0; k < 3; k++ {
		last := rf[len(rf)-1]
		rf, err = runDeliveryAttemptsNextPage(ctx, db, last.ID, ft)
		if err != nil {
			return err
		}
		fmt.Printf("> NextPage[%d, %s]: len=%d\n", k, last.ID, len(rf))
		for i, r := range rf {
			fmt.Printf("> [%d]=%v\n", i, r)
		}
	}
	return nil
}

func runDeliveryAttemptsCreate(ctx util.Context, db *sql.DB) (id string, err error) {
	id, err = getUserID(ctx, db)
	if err != nil {
		return "", err
	}
	return id, doWithTx(ctx, db, func(tx *sql.Tx) error {
		id, err = deliveryAttempts.Create(ctx, tx, id, mustParse(testPeerActor1InboxIRI), []byte("hello1"))
		return err
	})
}

func runDeliveryAttemptsMarkSuccessful(ctx util.Context, db *sql.DB) error {
	id, err := getUserID(ctx, db)
	if err != nil {
		return err
	}
	var daID string
	if err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		daID, err = deliveryAttempts.Create(ctx, tx, id, mustParse(testPeerActor1InboxIRI), []byte("hello2"))
		return err
	}); err != nil {
		return err
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return deliveryAttempts.MarkSuccessful(ctx, tx, daID)
	})
}

func runDeliveryAttemptsMarkFailed(ctx util.Context, db *sql.DB) error {
	id, err := getUserID(ctx, db)
	if err != nil {
		return err
	}
	var daID string
	if err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		daID, err = deliveryAttempts.Create(ctx, tx, id, mustParse(testPeerActor1InboxIRI), []byte("hello3"))
		return err
	}); err != nil {
		return err
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return deliveryAttempts.MarkFailed(ctx, tx, daID)
	})
}

func runDeliveryAttemptsMarkAbandoned(ctx util.Context, db *sql.DB) error {
	id, err := getUserID(ctx, db)
	if err != nil {
		return err
	}
	var daID string
	if err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		daID, err = deliveryAttempts.Create(ctx, tx, id, mustParse(testPeerActor1InboxIRI), []byte("hello4"))
		return err
	}); err != nil {
		return err
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return deliveryAttempts.MarkAbandoned(ctx, tx, daID)
	})
}

func runDeliveryAttemptsFirstPage(ctx util.Context, db *sql.DB) (rf []models.RetryableFailure, ft time.Time, err error) {
	var id string
	id, err = getUserID(ctx, db)
	if err != nil {
		return
	}
	if err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		// Make 24 additional failed, in addition to the existing one.
		for i := 0; i < 24; i++ {
			var daID string
			daID, err = deliveryAttempts.Create(ctx, tx, id, mustParse(testPeerActor1InboxIRI), []byte("hello_fetch_me"))
			if err != nil {
				return err
			}
			err = deliveryAttempts.MarkFailed(ctx, tx, daID)
			if err != nil {
				return err
			}
		}
		// Get the current "fetch" time
		ft = time.Now()
		return nil
	}); err != nil {
		return
	}
	if err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		// Make 10 more failed, which should be skipped
		for i := 0; i < 10; i++ {
			var daID string
			daID, err = deliveryAttempts.Create(ctx, tx, id, mustParse(testPeerActor2InboxIRI), []byte("hello_no_fetch"))
			if err != nil {
				return err
			}
			err = deliveryAttempts.MarkFailed(ctx, tx, daID)
			if err != nil {
				return err
			}
		}
		return err
	}); err != nil {
		return
	}
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		rf, err = deliveryAttempts.FirstPageFailures(ctx, tx, ft, 10)
		return err
	})
	return
}

func runDeliveryAttemptsNextPage(ctx util.Context, db *sql.DB, prev string, ft time.Time) (rf []models.RetryableFailure, err error) {
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		rf, err = deliveryAttempts.NextPageFailures(ctx, tx, prev, ft, 10)
		return err
	})
	return
}

/* Outboxes */

func runOutboxesCalls(ctx util.Context, db *sql.DB) error {
	if err := runOutboxesCreateOutbox(ctx, db); err != nil {
		return err
	}
	b, err := runOutboxesContainsForActorTrue(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> OutboxContainsForActor(%s): %v\n", testActivity1IRI, b)
	b, err = runOutboxesContainsForActorFalse(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> OutboxContainsForActor(%s): %v\n", testActivity2IRI, b)
	b, err = runOutboxesContainsTrue(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> OutboxContains(%s): %v\n", testActivity1IRI, b)
	b, err = runOutboxesContainsFalse(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> OutboxContains(%s): %v\n", testActivity2IRI, b)
	p, isEnd, err := runOutboxesGetOutbox(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> GetOutbox(%s, %d, %d): %s %v\n", testActor2OutboxIRI, 21, 27, p, isEnd)
	if pb, err := toJSON(p); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	p, isEnd, err = runOutboxesGetPublicOutbox(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> GetPublicOutbox(%s, %d, %d): %s %v\n", testActor3OutboxIRI, 1, 2, p, isEnd)
	if pb, err := toJSON(p); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	p, idx, err := runOutboxesGetLastPage(ctx, db, 19)
	if err != nil {
		return err
	}
	fmt.Printf("> GetLastPage(%d): idx=%d\n", 19, idx)
	if pb, err := toJSON(p); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	p, idx, err = runOutboxesGetPublicLastPage(ctx, db, 2)
	if err != nil {
		return err
	}
	fmt.Printf("> GetPublicLastPage(%d): idx=%d\n", 2, idx)
	if pb, err := toJSON(p); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	if err := runOutboxesPrependOutboxItem(ctx, db); err != nil {
		return err
	}
	if err := runOutboxesDeleteOutboxItem(ctx, db); err != nil {
		return err
	}
	obox, err := runOutboxesOutboxForInbox(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> OutboxForInbox: %s\n", obox.URL)
	return nil
}

func runOutboxesCreateOutbox(ctx util.Context, db *sql.DB) error {
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return outboxes.Create(ctx, tx, mustParse(testActor1IRI), testActor1Outbox)
	})
}

func runOutboxesContainsForActorTrue(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = outboxes.ContainsForActor(ctx, tx, mustParse(testActor1IRI), mustParse(testActivity1IRI))
		return err
	})
}

func runOutboxesContainsForActorFalse(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = outboxes.ContainsForActor(ctx, tx, mustParse(testActor1IRI), mustParse(testActivity2IRI))
		return err
	})
}

func runOutboxesContainsTrue(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = outboxes.Contains(ctx, tx, mustParse(testActor1OutboxIRI), mustParse(testActivity1IRI))
		return err
	})
}

func runOutboxesContainsFalse(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = outboxes.Contains(ctx, tx, mustParse(testActor1OutboxIRI), mustParse(testActivity2IRI))
		return err
	})
}

func runOutboxesGetOutbox(ctx util.Context, db *sql.DB) (p models.ActivityStreamsOrderedCollectionPage, isEnd bool, err error) {
	if err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		return outboxes.Create(ctx, tx, mustParse(testActor2IRI), testActor2Outbox)
	}); err != nil {
		return
	}
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		p, isEnd, err = outboxes.GetPage(ctx, tx, mustParse(testActor2OutboxIRI), 21, 27)
		return err
	})
	return
}

func runOutboxesGetPublicOutbox(ctx util.Context, db *sql.DB) (p models.ActivityStreamsOrderedCollectionPage, isEnd bool, err error) {
	if err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		return outboxes.Create(ctx, tx, mustParse(testActor3IRI), testActor3Outbox)
	}); err != nil {
		return
	}
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		p, isEnd, err = outboxes.GetPublicPage(ctx, tx, mustParse(testActor3OutboxIRI), 1, 2)
		return err
	})
	return
}

func runOutboxesGetLastPage(ctx util.Context, db *sql.DB, n int) (p models.ActivityStreamsOrderedCollectionPage, idx int, err error) {
	return p, idx, doWithTx(ctx, db, func(tx *sql.Tx) error {
		p, idx, err = outboxes.GetLastPage(ctx, tx, mustParse(testActor2OutboxIRI), n)
		return err
	})
}

func runOutboxesGetPublicLastPage(ctx util.Context, db *sql.DB, n int) (p models.ActivityStreamsOrderedCollectionPage, idx int, err error) {
	return p, idx, doWithTx(ctx, db, func(tx *sql.Tx) error {
		p, idx, err = outboxes.GetPublicLastPage(ctx, tx, mustParse(testActor3OutboxIRI), n)
		return err
	})
}

func runOutboxesPrependOutboxItem(ctx util.Context, db *sql.DB) error {
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return outboxes.PrependOutboxItem(ctx, tx, mustParse(testActor3OutboxIRI), mustParse(testActivity2IRI))
	})
}

func runOutboxesDeleteOutboxItem(ctx util.Context, db *sql.DB) error {
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return outboxes.DeleteOutboxItem(ctx, tx, mustParse(testActor3OutboxIRI), mustParse(testActivity4IRI))
	})
}

func runOutboxesOutboxForInbox(ctx util.Context, db *sql.DB) (obox models.URL, err error) {
	return obox, doWithTx(ctx, db, func(tx *sql.Tx) error {
		obox, err = outboxes.OutboxForInbox(ctx, tx, mustParse(testActor1InboxIRI))
		return err
	})
}

/* Inboxes */

func runInboxesCalls(ctx util.Context, db *sql.DB) error {
	if err := runInboxesCreateInbox(ctx, db); err != nil {
		return err
	}
	b, err := runInboxesContainsForActorTrue(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> InboxContainsForActor(%s): %v\n", testActivity1IRI, b)
	b, err = runInboxesContainsForActorFalse(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> InboxContainsForActor(%s): %v\n", testActivity2IRI, b)
	b, err = runInboxesContainsTrue(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> InboxContains(%s): %v\n", testActivity1IRI, b)
	b, err = runInboxesContainsFalse(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> InboxContains(%s): %v\n", testActivity2IRI, b)
	p, isEnd, err := runInboxesGetInbox(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> GetInbox(%s, %d, %d): %s %v\n", testActor2InboxIRI, 21, 27, p, isEnd)
	if pb, err := toJSON(p); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	p, isEnd, err = runInboxesGetPublicInbox(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> GetPublicInbox(%s, %d, %d): %s %v\n", testActor3InboxIRI, 1, 2, p, isEnd)
	if pb, err := toJSON(p); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	p, idx, err := runInboxesGetLastPage(ctx, db, 19)
	if err != nil {
		return err
	}
	fmt.Printf("> GetLastPage(%d): idx=%d\n", 19, idx)
	if pb, err := toJSON(p); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	p, idx, err = runInboxesGetPublicLastPage(ctx, db, 2)
	if err != nil {
		return err
	}
	fmt.Printf("> GetPublicLastPage(%d): idx=%d\n", 2, idx)
	if pb, err := toJSON(p); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	if err := runInboxesPrependInboxItem(ctx, db); err != nil {
		return err
	}
	if err := runInboxesDeleteInboxItem(ctx, db); err != nil {
		return err
	}
	return nil
}

func runInboxesCreateInbox(ctx util.Context, db *sql.DB) error {
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return inboxes.Create(ctx, tx, mustParse(testActor1IRI), testActor1Inbox)
	})
}

func runInboxesContainsForActorTrue(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = inboxes.ContainsForActor(ctx, tx, mustParse(testActor1IRI), mustParse(testActivity1IRI))
		return err
	})
}

func runInboxesContainsForActorFalse(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = inboxes.ContainsForActor(ctx, tx, mustParse(testActor1IRI), mustParse(testActivity2IRI))
		return err
	})
}

func runInboxesContainsTrue(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = inboxes.Contains(ctx, tx, mustParse(testActor1InboxIRI), mustParse(testActivity1IRI))
		return err
	})
}

func runInboxesContainsFalse(ctx util.Context, db *sql.DB) (b bool, err error) {
	return b, doWithTx(ctx, db, func(tx *sql.Tx) error {
		b, err = inboxes.Contains(ctx, tx, mustParse(testActor1InboxIRI), mustParse(testActivity2IRI))
		return err
	})
}

func runInboxesGetInbox(ctx util.Context, db *sql.DB) (p models.ActivityStreamsOrderedCollectionPage, isEnd bool, err error) {
	if err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		return inboxes.Create(ctx, tx, mustParse(testActor2IRI), testActor2Inbox)
	}); err != nil {
		return
	}
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		p, isEnd, err = inboxes.GetPage(ctx, tx, mustParse(testActor2InboxIRI), 21, 27)
		return err
	})
	return
}

func runInboxesGetPublicInbox(ctx util.Context, db *sql.DB) (p models.ActivityStreamsOrderedCollectionPage, isEnd bool, err error) {
	if err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		return fedData.Create(ctx, tx, models.ActivityStreams{testActivity7})
	}); err != nil {
		return
	}
	if err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		return localData.Create(ctx, tx, models.ActivityStreams{testActivity8})
	}); err != nil {
		return
	}
	if err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		return inboxes.Create(ctx, tx, mustParse(testActor3IRI), testActor3Inbox)
	}); err != nil {
		return
	}
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		p, isEnd, err = inboxes.GetPublicPage(ctx, tx, mustParse(testActor3InboxIRI), 1, 2)
		return err
	})
	return
}

func runInboxesGetLastPage(ctx util.Context, db *sql.DB, n int) (p models.ActivityStreamsOrderedCollectionPage, idx int, err error) {
	return p, idx, doWithTx(ctx, db, func(tx *sql.Tx) error {
		p, idx, err = inboxes.GetLastPage(ctx, tx, mustParse(testActor2InboxIRI), n)
		return err
	})
}

func runInboxesGetPublicLastPage(ctx util.Context, db *sql.DB, n int) (p models.ActivityStreamsOrderedCollectionPage, idx int, err error) {
	return p, idx, doWithTx(ctx, db, func(tx *sql.Tx) error {
		p, idx, err = inboxes.GetPublicLastPage(ctx, tx, mustParse(testActor3InboxIRI), n)
		return err
	})
}

func runInboxesPrependInboxItem(ctx util.Context, db *sql.DB) error {
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return inboxes.PrependInboxItem(ctx, tx, mustParse(testActor3InboxIRI), mustParse(testActivity2IRI))
	})
}

func runInboxesDeleteInboxItem(ctx util.Context, db *sql.DB) error {
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return inboxes.DeleteInboxItem(ctx, tx, mustParse(testActor3InboxIRI), mustParse(testActivity4IRI))
	})
}

/* LocalData */

func runLocalDataCalls(ctx util.Context, db *sql.DB) error {
	if err := runLocalDataCreate(ctx, db); err != nil {
		return err
	}
	if err := runLocalDataUpdate(ctx, db); err != nil {
		return err
	}
	if err := runLocalDataDelete(ctx, db); err != nil {
		return err
	}
	p, err := runLocalDataGet(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> Get: %v\n", p)
	if pb, err := toJSON(p); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	ex, err := runLocalDataExists(ctx, db, testActivity4IRI)
	if err != nil {
		return err
	}
	fmt.Printf("> Exists(%s): %v\n", testActivity4IRI, ex)
	ex, err = runLocalDataExists(ctx, db, testActivity5IRI)
	if err != nil {
		return err
	}
	fmt.Printf("> Exists(%s): %v\n", testActivity5IRI, ex)
	st, err := runLocalDataStats(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> Stats: %v\n", st)
	if pb, err := toJSON(st); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	return nil
}

func runLocalDataCreate(ctx util.Context, db *sql.DB) error {
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return localData.Create(ctx, tx, models.ActivityStreams{testActivity4})
	})
}

func runLocalDataUpdate(ctx util.Context, db *sql.DB) error {
	if err := doWithTx(ctx, db, func(tx *sql.Tx) error {
		return localData.Create(ctx, tx, models.ActivityStreams{testActivity5})
	}); err != nil {
		return err
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return localData.Update(ctx, tx, mustParse(testActivity5IRI), models.ActivityStreams{testActivity6})
	})
}

func runLocalDataDelete(ctx util.Context, db *sql.DB) error {
	if err := doWithTx(ctx, db, func(tx *sql.Tx) error {
		return localData.Create(ctx, tx, models.ActivityStreams{testActivity5})
	}); err != nil {
		return err
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return localData.Delete(ctx, tx, mustParse(testActivity5IRI))
	})
}

func runLocalDataGet(ctx util.Context, db *sql.DB) (v models.ActivityStreams, err error) {
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		v, err = localData.Get(ctx, tx, mustParse(testActivity4IRI))
		return err
	})
	return
}

func runLocalDataExists(ctx util.Context, db *sql.DB, id string) (v bool, err error) {
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		v, err = localData.Exists(ctx, tx, mustParse(id))
		return err
	})
	return
}

func runLocalDataStats(ctx util.Context, db *sql.DB) (st models.LocalDataActivity, err error) {
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		st, err = localData.Stats(ctx, tx)
		return err
	})
	return
}

/* FedData */

func runFedDataCalls(ctx util.Context, db *sql.DB) error {
	if err := runFedDataCreate(ctx, db); err != nil {
		return err
	}
	if err := runFedDataUpdate(ctx, db); err != nil {
		return err
	}
	if err := runFedDataDelete(ctx, db); err != nil {
		return err
	}
	p, err := runFedDataGet(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> Get: %v\n", p)
	if pb, err := toJSON(p); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	ex, err := runFedDataExists(ctx, db, testActivity1IRI)
	if err != nil {
		return err
	}
	fmt.Printf("> Exists(%s): %v\n", testActivity1IRI, ex)
	ex, err = runFedDataExists(ctx, db, testActivity2IRI)
	if err != nil {
		return err
	}
	fmt.Printf("> Exists(%s): %v\n", testActivity2IRI, ex)
	return nil
}

func runFedDataCreate(ctx util.Context, db *sql.DB) error {
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return fedData.Create(ctx, tx, models.ActivityStreams{testActivity1})
	})
}

func runFedDataUpdate(ctx util.Context, db *sql.DB) error {
	if err := doWithTx(ctx, db, func(tx *sql.Tx) error {
		return fedData.Create(ctx, tx, models.ActivityStreams{testActivity2})
	}); err != nil {
		return err
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return fedData.Update(ctx, tx, mustParse(testActivity2IRI), models.ActivityStreams{testActivity3})
	})
}

func runFedDataDelete(ctx util.Context, db *sql.DB) error {
	if err := doWithTx(ctx, db, func(tx *sql.Tx) error {
		return fedData.Create(ctx, tx, models.ActivityStreams{testActivity2})
	}); err != nil {
		return err
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return fedData.Delete(ctx, tx, mustParse(testActivity2IRI))
	})
}

func runFedDataGet(ctx util.Context, db *sql.DB) (v models.ActivityStreams, err error) {
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		v, err = fedData.Get(ctx, tx, mustParse(testActivity1IRI))
		return err
	})
	return
}

func runFedDataExists(ctx util.Context, db *sql.DB, id string) (v bool, err error) {
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		v, err = fedData.Exists(ctx, tx, mustParse(id))
		return err
	})
	return
}

/* UserModel */

func runUserModelCalls(ctx util.Context, db *sql.DB) error {
	userID, err := runUserModelCreateUser(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> Create(): %s\n", userID)
	if err := runUserModelUpdateActor(ctx, db, userID); err != nil {
		return err
	}
	s, err := runUserModelSensitiveUserByEmail(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> SensitiveUserByEmail(%s): %v\n", testEmail1, s)
	u, err := runUserModelUserByID(ctx, db, s.ID)
	if err != nil {
		return err
	}
	fmt.Printf("> UserByID(%s): %v\n", s.ID, u)
	if pb, err := toJSON(u.Actor); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	u, err = runUserModelUserByPreferredUsernameNone(ctx, db)
	fmt.Printf("> UserByPreferredUsername( N/A ): %v %v\n", u, err)
	u, err = runUserModelUserByPreferredUsername(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> UserByPreferredUsername(%s): %v\n", testActor1PreferredUsername, u)
	if pb, err := toJSON(u.Actor); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	id, err := runUserModelActorIDForOutbox(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> ActorIDForOutbox: %s\n", id.URL)
	id, err = runUserModelActorIDForInbox(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> ActorIDForInbox: %s\n", id.URL)
	if err := runUserModelUpdatePreferences(ctx, db, userID); err != nil {
		return err
	}
	if err := runUserModelUpdatePrivileges(ctx, db, userID); err != nil {
		return err
	}
	u, err = runUserModelInstanceActor(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> InstanceActorUser: %v\n", u)
	if pb, err := toJSON(u.Actor); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	if err := runUserModelSetInstanceActorPreferences(ctx, db); err != nil {
		return err
	}
	iap, err := runUserModelInstanceActorPreferences(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> InstanceActorPreferences: %v\n", iap)
	if pb, err := toJSON(iap); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	st, err := runUserModelUserActivityStats(ctx, db)
	if err != nil {
		return err
	}
	fmt.Printf("> UserActivityStats: %v\n", st)
	if pb, err := toJSON(st); err != nil {
		return err
	} else {
		fmt.Printf("> JSON:\n%s\n", pb)
	}
	return nil
}

func runUserModelCreateUser(ctx util.Context, db *sql.DB) (id string, err error) {
	cu := &models.CreateUser{
		Email:       testEmail1,
		Hashpass:    []byte{1, 2, 3},
		Salt:        []byte{4, 5, 6},
		Actor:       models.ActivityStreamsPerson{streams.NewActivityStreamsPerson()},
		Privileges:  models.Privileges{},
		Preferences: models.Preferences{},
	}
	return id, doWithTx(ctx, db, func(tx *sql.Tx) error {
		id, err = users.Create(ctx, tx, cu)
		return err
	})
}

func runUserModelUpdateActor(ctx util.Context, db *sql.DB, userID string) error {
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return users.UpdateActor(ctx, tx, userID, testActor1)
	})
}

func runUserModelSensitiveUserByEmail(ctx util.Context, db *sql.DB) (s *models.SensitiveUser, err error) {
	var tx *sql.Tx
	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer tx.Rollback()
	if s, err = users.SensitiveUserByEmail(ctx, tx, testEmail1); err != nil {
		return
	}
	return s, tx.Commit()
}

func runUserModelUserByPreferredUsername(ctx util.Context, db *sql.DB) (s *models.User, err error) {
	var tx *sql.Tx
	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer tx.Rollback()
	if s, err = users.UserByPreferredUsername(ctx, tx, testActor1PreferredUsername); err != nil {
		return
	}
	return s, tx.Commit()
}

func runUserModelUserByPreferredUsernameNone(ctx util.Context, db *sql.DB) (s *models.User, err error) {
	var tx *sql.Tx
	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer tx.Rollback()
	if s, err = users.UserByPreferredUsername(ctx, tx, "gibberish"); err != nil {
		return
	}
	return s, tx.Commit()
}

func runUserModelUserByID(ctx util.Context, db *sql.DB, id string) (s *models.User, err error) {
	var tx *sql.Tx
	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer tx.Rollback()
	if s, err = users.UserByID(ctx, tx, id); err != nil {
		return
	}
	return s, tx.Commit()
}

func runUserModelActorIDForOutbox(ctx util.Context, db *sql.DB) (id models.URL, err error) {
	return id, doWithTx(ctx, db, func(tx *sql.Tx) error {
		id, err = users.ActorIDForOutbox(ctx, tx, mustParse(testActor1OutboxIRI))
		return err
	})
}

func runUserModelActorIDForInbox(ctx util.Context, db *sql.DB) (id models.URL, err error) {
	return id, doWithTx(ctx, db, func(tx *sql.Tx) error {
		id, err = users.ActorIDForInbox(ctx, tx, mustParse(testActor1InboxIRI))
		return err
	})
}

func runUserModelUpdatePreferences(ctx util.Context, db *sql.DB, id string) error {
	pref := models.Preferences{
		OnFollow: models.OnFollowBehavior(pub.OnFollowAutomaticallyAccept),
		Payload:  []byte(`{"test":"pref"}`),
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return users.UpdatePreferences(ctx, tx, id, pref)
	})
}

func runUserModelUpdatePrivileges(ctx util.Context, db *sql.DB, id string) error {
	priv := models.Privileges{
		Admin:   true,
		Payload: []byte(`{"test":"priv"}`),
	}
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return users.UpdatePrivileges(ctx, tx, id, priv)
	})
}

func runUserModelInstanceActor(ctx util.Context, db *sql.DB) (s *models.User, err error) {
	var id string
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		cu := &models.CreateUser{
			Email:       "",
			Hashpass:    []byte{},
			Salt:        []byte{},
			Actor:       models.ActivityStreamsApplication{streams.NewActivityStreamsApplication()},
			Privileges:  models.Privileges{InstanceActor: true},
			Preferences: models.Preferences{},
		}
		id, err = users.Create(ctx, tx, cu)
		return err
	})
	if err != nil {
		return
	}
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		s, err = users.InstanceActorUser(ctx, tx)
		return err
	})
	return
}

func runUserModelSetInstanceActorPreferences(ctx util.Context, db *sql.DB) error {
	return doWithTx(ctx, db, func(tx *sql.Tx) error {
		return users.SetInstanceActorPreferences(ctx, tx, models.InstanceActorPreferences{
			OpenRegistrations: true,
			ServerBaseURL:     "https://example.com/test/base/url",
			ServerName:        "test server name",
			OrgName:           "test org",
			OrgContact:        "test org contact",
			OrgAccount:        "test org account",
		})
	})
}

func runUserModelInstanceActorPreferences(ctx util.Context, db *sql.DB) (iap models.InstanceActorPreferences, err error) {
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		iap, err = users.InstanceActorPreferences(ctx, tx)
		return err
	})
	return
}

func runUserModelUserActivityStats(ctx util.Context, db *sql.DB) (st models.UserActivityStats, err error) {
	err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		st, err = users.ActivityStats(ctx, tx)
		return err
	})
	return
}

/* Models */

func createTables(ctx util.Context, db *sql.DB, d models.SqlDialect) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, m := range testModels {
		if err := m.CreateTable(tx, d); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func prepareStatements(ctx util.Context, db *sql.DB, d models.SqlDialect) error {
	for _, m := range testModels {
		if err := m.Prepare(db, d); err != nil {
			return err
		}
	}
	return nil
}

func closeModels() error {
	for _, m := range testModels {
		m.Close()
	}
	return nil
}

/* Utils */

func connectPostgres(url string) (*sql.DB, error) {
	return sql.Open("pgx", url)
}

func dialectPostgres(schema string) models.SqlDialect {
	return db.NewPgV0(schema)
}

func mustParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

func doWithTx(ctx util.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func toJSON(i interface{}) ([]byte, error) {
	switch v := i.(type) {
	case vocab.Type:
		m, err := streams.Serialize(v)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(m, "", "  ")
	default:
		return json.MarshalIndent(i, "", "  ")
	}
}

func getUserID(ctx util.Context, db *sql.DB) (id string, err error) {
	var s *models.SensitiveUser
	if err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		s, err = users.SensitiveUserByEmail(ctx, tx, testEmail1)
		return err
	}); err != nil {
		return
	}
	id = s.ID
	return
}

func getInstanceActorUserID(ctx util.Context, db *sql.DB) (id string, err error) {
	var s *models.User
	if err = doWithTx(ctx, db, func(tx *sql.Tx) error {
		s, err = users.InstanceActorUser(ctx, tx)
		return err
	}); err != nil {
		return
	}
	id = s.ID
	return
}
