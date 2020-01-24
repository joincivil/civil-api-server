package posts_test

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/posts"
	"github.com/joincivil/civil-api-server/pkg/testruntime"
	"github.com/joincivil/civil-api-server/pkg/testutils"
	"github.com/joincivil/civil-api-server/pkg/utils"
	"github.com/joincivil/go-common/pkg/email"
	"github.com/joincivil/go-common/pkg/newsroom"
	uuid "github.com/satori/go.uuid"
	"math/rand"
	"os"
	"strconv"
	// "strings"
	"testing"
	"time"
)

const (
	testSignupLoginProtoHost = "http://localhost:8080"
	sendGridKeyEnvVar        = "SENDGRID_TEST_KEY"
	useSandbox               = true
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

var stripeAccountID = "testaccountid"

func randomAddress() common.Address {
	str := strconv.FormatInt(r.Int63(), 10)
	return common.HexToAddress(str)
}

func randomUUID() string {
	u := uuid.NewV4()
	return u.String()
}

var (
	user1ID           = randomUUID()
	user1Address      = randomAddress()
	user2ID           = randomUUID()
	user2Address      = randomAddress()
	newsroom1Address  = randomAddress()
	newsroom1Multisig = randomAddress()
	newsroom2Address  = randomAddress()
	newsroom2Multisig = randomAddress()
	newsroom3Address  = randomAddress()
	newsroom3Multisig = randomAddress()
)

func makeValidChannelBoost(channelID string) posts.Boost {
	return posts.Boost{
		PostModel: posts.PostModel{
			ChannelID: channelID,
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

func makeValidPostComment(postID string, commentType string) posts.Comment {
	return posts.Comment{
		PostModel: posts.PostModel{
			ParentID: &postID,
		},
		CommentType: commentType,
		Text:        "whatever I feel like",
	}
}

type MockGetNewsroomHelper struct{}

func (g MockGetNewsroomHelper) GetMultisigMembers(newsroomAddress common.Address) ([]common.Address, error) {
	if newsroomAddress == newsroom1Address {
		return []common.Address{user1Address}, nil
	} else if newsroomAddress == newsroom2Address {
		return []common.Address{user2Address}, nil
	} else if newsroomAddress == newsroom3Address {
		return []common.Address{user1Address}, nil
	}
	return nil, errors.New("not found")
}
func (g MockGetNewsroomHelper) GetOwner(newsroomAddress common.Address) (common.Address, error) {
	if newsroomAddress == newsroom1Address {
		return newsroom1Multisig, nil
	} else if newsroomAddress == newsroom2Address {
		return newsroom2Multisig, nil
	} else if newsroomAddress == newsroom3Address {
		return newsroom3Multisig, nil
	}
	return common.Address{}, errors.New("found found")
}

type MockNewsroomService struct{}

func (s MockNewsroomService) GetNewsroomByAddress(newsroomAddress string) (*newsroom.Newsroom, error) {
	return nil, nil
}

type MockStripeConnector struct{}

func (s MockStripeConnector) ConnectAccount(code string) (string, error) {
	if code == "fail" {
		return "", fmt.Errorf("error setting stripe account")
	}
	return stripeAccountID, nil
}

func (s MockStripeConnector) EnableApplePay(stripeAccountID string) ([]string, error) {
	return []string{}, nil
}

func (s MockStripeConnector) GetApplyPayDomains(stripeAccountID string) ([]string, error) {
	return []string{}, nil
}

func (s MockStripeConnector) IsApplePayEnabled(stripeAccountID string) (bool, error) {
	return false, nil
}

func getSendGridKeyFromEnvVar() string {
	return os.Getenv(sendGridKeyEnvVar)
}

func TestPosts(t *testing.T) {
	db, err := testutils.GetTestDBConnection()
	if err != nil {
		t.Fatalf("error getting DB: %v", err)
	}
	err = testruntime.RunMigrations(db)
	if err != nil {
		t.Fatalf("error cleaning DB: %v", err)
	}

	persister := channels.NewDBPersister(db)
	generator := utils.NewJwtTokenGenerator([]byte("secret"))

	sendGridKey := getSendGridKeyFromEnvVar()
	emailer := email.NewEmailerWithSandbox(sendGridKey, useSandbox)
	channelService := channels.NewService(persister, MockGetNewsroomHelper{}, MockStripeConnector{}, generator, emailer, testSignupLoginProtoHost)

	randomInt := r.Int31()
	groupHandle := fmt.Sprintf("gh%v", randomInt)

	_, err = channelService.CreateUserChannel(user1ID)
	if err != nil {
		t.Fatalf("not expecting error creating user channel: %v", err)
	}
	_, err = channelService.CreateUserChannel(user2ID)
	if err != nil {
		t.Fatalf("not expecting error creating user channel: %v", err)
	}
	channel, err := channelService.CreateGroupChannel(user1ID, groupHandle)
	if err != nil {
		t.Fatalf("not expecting error creating group channel: %v", err)
	}

	postPersister := posts.NewDBPostPersister(db)
	postService := posts.NewService(postPersister, channelService, MockNewsroomService{})

	boost := makeValidChannelBoost(channel.ID)
	post, err := postService.CreatePost(user1ID, boost)
	if err != nil {
		t.Fatalf("not expecting error creating post: %v", err)
	}

	commentAnnouncementValid := makeValidPostComment(post.GetID(), posts.TypeCommentAnnouncement)
	commentPost, err := postService.CreatePost(user1ID, commentAnnouncementValid)
	if err != nil {
		t.Fatalf("not expecting error creating valid announcement comment: %v", err)
	}

	t.Logf("postID: %s", post.GetID())
	t.Logf("commentPost: %s", commentPost.GetID())

	children, err := postService.GetChildrenOfPost(post.GetID())
	if err != nil {
		t.Fatalf("not expecting error getting children: %v", err)
	}

	if children == nil {
		t.Fatalf("not expecting children to be nil")
	}

	t.Logf("children: %v", children)
	for _, p := range children {
		if p == nil {
			t.Fatalf("not expecting child to be nil")
		}
		childID := p.GetID()
		t.Logf("p.ID: %s", p.GetID())
		if childID != commentPost.GetID() {
			t.Fatalf("childID not equal to created comment")
		}
	}

	_, err = postService.CreatePost(user2ID, commentAnnouncementValid)
	if err == nil {
		t.Fatalf("was expecting error creating valid announcement comment from non-admin")
	}

	commentDefaultValid := makeValidPostComment(post.GetID(), posts.TypeCommentDefault)
	_, err = postService.CreatePost(user2ID, commentDefaultValid)
	if err != nil {
		t.Fatalf("was not expecting error creating valid default comment from non-admin: %v", err)
	}
}
