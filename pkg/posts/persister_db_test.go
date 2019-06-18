package posts_test

import (
	"testing"

	"github.com/joincivil/civil-api-server/pkg/posts"
	"github.com/joincivil/civil-api-server/pkg/testutils"
)

func TestPersistPost(t *testing.T) {
	db, err := testutils.GetTestDBConnection()
	if err != nil {
		t.Errorf("error: %v", err)
	}

	persister := posts.NewDBPostPersister(db)

	boost := &posts.Boost{
		PostModel: posts.PostModel{
			ChannelID: "alice_newsrooom",
			AuthorID:  "alice",
		},
		CurrencyCode: "USD",
		GoalAmount:   100.10,
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
	}
	link := &posts.ExternalLink{
		PostModel: posts.PostModel{
			ChannelID: "bob_newsroom",
			AuthorID:  "bob",
		},
		URL: "https://totallylegitnews.com",
	}

	boostPost, err := persister.CreatePost(boost)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	t.Logf("created boost: %v", boostPost)

	linkPost, err := persister.CreatePost(link)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	t.Logf("created external link: %v", linkPost.GetPostModel().ID)

	linkReceived, err := persister.GetPost(linkPost.GetPostModel().ID)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	t.Logf("got external link: %v", linkReceived)

	boostReceived, err := persister.GetPost(boostPost.GetPostModel().ID)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	t.Logf("got boost: %v", boostReceived)
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
}
