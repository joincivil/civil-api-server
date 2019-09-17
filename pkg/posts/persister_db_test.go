package posts_test

import (
	"github.com/jinzhu/gorm"
	"github.com/joincivil/civil-api-server/pkg/posts"
	"github.com/joincivil/civil-api-server/pkg/testruntime"
	"github.com/joincivil/civil-api-server/pkg/testutils"
	"testing"
	"time"
)

var (
	aliceUserUUID     = "7497e091-fa1c-47c4-88ca-5680f25c5729"
	aliceNewsroomUUID = "e9e86977-9d45-416c-83f2-9b70253174f3"
	bobNewsroomUUID   = "8124bc09-94eb-4694-bb25-88c0af262577"
	bobUserUUID       = "e684d748-80ae-4e42-a096-9314cf4f605e"
)

func helperCreatePost(t *testing.T, persister posts.PostPersister, post posts.Post) posts.Post {
	createdPost, err := persister.CreatePost(aliceUserUUID, post)
	if err != nil {
		t.Errorf("error: %v", err)
	}

	return createdPost
}

func makeValidBoost() *posts.Boost {
	return &posts.Boost{
		PostModel: posts.PostModel{
			ChannelID: aliceNewsroomUUID,
		},
		CurrencyCode: "USD",
		GoalAmount:   100.10,
		Title:        "some title",
		About:        "_abouttest_",
		Items: []posts.BoostItem{
			{
				Item: "foo",
				Cost: 100.1,
			},
			{
				Item: "bar",
				Cost: 100.1,
			},
		},
		DateEnd: time.Now().AddDate(0, 1, 0),
	}
}

func makeBadEndDateBoost() *posts.Boost {
	return &posts.Boost{
		PostModel: posts.PostModel{
			ChannelID: aliceNewsroomUUID,
		},
		CurrencyCode: "USD",
		GoalAmount:   100.10,
		Title:        "some title",
		About:        "_abouttest_",
		Items: []posts.BoostItem{
			{
				Item: "foo",
				Cost: 100.1,
			},
			{
				Item: "bar",
				Cost: 100.1,
			},
		},
		DateEnd: time.Now().AddDate(0, -1, 0),
	}
}

var db *gorm.DB

func initPersister(t *testing.T) posts.PostPersister {
	if db == nil {
		testDB, err := testutils.GetTestDBConnection()
		if err != nil {
			t.Errorf("error: %v", err)
		}
		db = testDB
		err = testruntime.RunMigrations(db)
		if err != nil {
			t.Fatalf("error cleaning DB: %v", err)
		}
	}

	return posts.NewDBPostPersister(db)
}

func TestCreatePost(t *testing.T) {
	persister := initPersister(t)

	boost := makeValidBoost()
	link := &posts.ExternalLink{
		PostModel: posts.PostModel{
			ChannelID: bobNewsroomUUID,
			AuthorID:  bobUserUUID,
		},
		URL: "https://totallylegitnews.com",
	}

	boostPost := helperCreatePost(t, persister, boost)
	linkPost := helperCreatePost(t, persister, link)

	_, err := persister.GetPost(linkPost.GetPostModel().ID)
	if err != nil {
		t.Errorf("error: %v", err)
	}

	boostReceived, err := persister.GetPost(boostPost.GetPostModel().ID)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	if boostReceived.(*posts.Boost).About != "_abouttest_" {
		t.Fatal("expected Boost to have a value of `_abouttest_` for About field ")
	}
	items := boostReceived.(*posts.Boost).Items
	if len(items) != 2 {
		t.Fatal("expected Boost have 2 items")
	}
	if items[0].Item != "foo" || items[1].Item != "bar" {
		t.Fatal("expected Boost boost items to be `foo` and `bar`")
	}

	if boostReceived.GetPostModel().AuthorID != aliceUserUUID {
		t.Fatalf("expected AuthorID to be `%v`", aliceUserUUID)
	}
}

func TestCreateBadBoost(t *testing.T) {
	persister := initPersister(t)
	boost := makeBadEndDateBoost()
	_, err := persister.CreatePost(aliceUserUUID, boost)
	if err == nil {
		t.Fatalf("expected ErrorBadBoostEndDate")
	}
}

func TestEditPost(t *testing.T) {
	persister := initPersister(t)

	boost := makeValidBoost()
	boostPost := helperCreatePost(t, persister, boost)

	patch := &posts.Boost{
		Title: "changed title",
		Why:   "changed value",
	}

	_, err := persister.EditPost(aliceUserUUID, boostPost.GetID(), patch)
	if err != nil {
		t.Fatalf("was not expecting an error: %v", err)
	}

	returnedPost, err := persister.GetPost(boostPost.GetID())
	if err != nil {
		t.Fatalf("was not expecting an error: %v", err)
	}
	if returnedPost.GetID() != boostPost.GetID() {
		t.Fatalf("was not expecting the ID to change after the edit")
	}

	returnedBoost := returnedPost.(*posts.Boost)

	if returnedBoost.Title != "changed title" {
		t.Fatalf("expecting the returned boost to have the new title")
	}
	if returnedBoost.About != "_abouttest_" {
		t.Fatalf("fields that were not changed should still exist")
	}

	_, err = persister.EditPost("some dude", boostPost.GetID(), patch)
	if err != posts.ErrorNotAuthorized {
		t.Fatalf("was expecting ErrorNotAuthorized: %v", err)
	}

	// t.Fatal("need to implement: edit post by another author")

}

func TestGetPost(t *testing.T) {

	persister := initPersister(t)

	_, err := persister.GetPost("70f163b2-9c1e-11e9-a2a3-2a2ae2dbcce4")
	if err != posts.ErrorNotFound {
		t.Fatalf("expecting posts.ErrorNotFound but instead received: %v", err)
	}

	boost := makeValidBoost()
	boostPost := helperCreatePost(t, persister, boost)

	retrievedPost, err := persister.GetPost(boostPost.GetID())
	if err != nil {
		t.Fatalf("was not expecting an error: %v", err)
	}

	if retrievedPost.GetPostModel().ID != boostPost.GetID() {
		t.Fatalf("expecting retrieved post to have the same ad as created post")
	}
}

func TestDelete(t *testing.T) {
	persister := initPersister(t)

	boost := makeValidBoost()
	boostPost := helperCreatePost(t, persister, boost)

	err := persister.DeletePost(aliceUserUUID, boostPost.GetID())
	if err != nil {
		t.Fatalf("was not expecting an error: %v", err)
	}

	_, err = persister.GetPost(boostPost.GetID())
	if err != posts.ErrorNotFound {
		t.Fatalf("expecting ErrorNotFound: %v", err)
	}
}
