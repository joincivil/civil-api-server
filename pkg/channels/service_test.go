package channels_test

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/testruntime"
	"github.com/joincivil/civil-api-server/pkg/testutils"
	"github.com/joincivil/civil-api-server/pkg/utils"
	"github.com/joincivil/go-common/pkg/email"
	uuid "github.com/satori/go.uuid"
	"math/rand"
	"os"
	"strconv"
	"strings"
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
	u, err := uuid.NewV4()
	if err != nil {
		panic("couldnt generate uuid")
	}
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

func GetETHAddresses(userID string) []common.Address {
	if userID == user1ID {
		return []common.Address{user1Address}
	} else if userID == user2ID {
		return []common.Address{user2Address}
	}
	return []common.Address{}
}

type MockStripeConnector struct{}

func (s MockStripeConnector) ConnectAccount(code string) (string, error) {
	if code == "fail" {
		return "", fmt.Errorf("error setting stripe account")
	}
	return stripeAccountID, nil
}

func getSendGridKeyFromEnvVar() string {
	return os.Getenv(sendGridKeyEnvVar)
}

func TestCreateChannel(t *testing.T) {
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
	svc := channels.NewService(persister, MockGetNewsroomHelper{}, MockStripeConnector{}, generator, emailer, testSignupLoginProtoHost)

	channel, err := svc.CreateUserChannel(user1ID)
	if err != nil {
		t.Fatalf("not expecting error: %v", err)
	}

	if channel.ID == "" {
		t.Fatal("channel does not have an ID")
	}

	retrieved, err := svc.GetChannel(channel.ID)
	if err != nil {
		t.Fatalf("not expecting error: %v", err)
	}

	if retrieved.ID != channel.ID {
		t.Fatal("retrieved channel does not match initial ID")
	}

	t.Run("roles", func(t *testing.T) {
		// channel should set the creator as the owner
		if len(channel.Members) != 1 {
			t.Fatalf("should have 1 retrieved channel member but received %v", len(retrieved.Members))
		}
		member := channel.Members[0]

		if member.UserID != user1ID {
			t.Fatal("initial member should be the creator")
		}
		if member.Role != channels.RoleAdmin {
			t.Fatal("initial role should be `admin`")
		}
	})

	t.Run("user type", func(t *testing.T) {
		userID := randomUUID()

		_, err = svc.CreateUserChannel(userID)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		// only allow a user to have 1 channel
		_, err = svc.CreateUserChannel(userID)
		if err != channels.ErrorNotUnique {
			t.Fatalf("was expecting ErrorNotUnique")
		}
	})

	t.Run("group type", func(t *testing.T) {
		userID := randomUUID()
		randomInt := r.Int31()
		handle := fmt.Sprintf("tEst%v", randomInt)
		nonUniqueHandle := fmt.Sprintf("test%v", randomInt)

		_, err = svc.CreateGroupChannel(userID, handle)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		// don't allow if handle already exists
		_, err = svc.CreateGroupChannel(userID, handle)
		if err != channels.ErrorNotUnique {
			t.Fatalf("was expecting ErrorNotUnique")
		}

		// don't allow if handle already exists
		_, err = svc.CreateGroupChannel(userID, nonUniqueHandle)
		if err != channels.ErrorNotUnique {
			t.Fatalf("was expecting ErrorNotUnique")
		}

		// only allow valid handles
		_, err = svc.CreateGroupChannel(userID, "Civil Found")
		if err != channels.ErrorInvalidHandle {
			t.Fatalf("was expecting ErrorInvalidHandle")
		}

	})

	t.Run("newsroom type", func(t *testing.T) {
		t.Run("invalid address", func(t *testing.T) {
			newsroomAddress := "hello"
			ethAddresses := GetETHAddresses(user1ID)
			_, err := svc.CreateNewsroomChannel(user1ID, ethAddresses, channels.CreateNewsroomChannelInput{
				ContractAddress: newsroomAddress,
			})
			if err != channels.ErrorInvalidHandle {
				t.Fatalf("was expecting error `channels.ErrorInvalidHandle`")
			}
		})

		t.Run("user not member of multisig", func(t *testing.T) {
			newsroomAddress := newsroom3Address.String()
			ethAddresses := GetETHAddresses(user2ID)
			_, err := svc.CreateNewsroomChannel(user2ID, ethAddresses, channels.CreateNewsroomChannelInput{
				ContractAddress: newsroomAddress,
			})
			if err != channels.ErrorUnauthorized {
				t.Fatalf("was expecting error `channels.ErrorUnauthorized` but received: %v", err)
			}
		})

		t.Run("success", func(t *testing.T) {
			ethAddresses := GetETHAddresses(user1ID)
			channel, err := svc.CreateNewsroomChannel(user1ID, ethAddresses, channels.CreateNewsroomChannelInput{
				ContractAddress: newsroom1Address.String(),
			})
			if err != nil {
				t.Fatalf("not expecting error: %v", err)
			}
			if channel.Reference != newsroom1Address.String() {
				t.Fatalf("expecting newsroom `reference` field to be " + newsroom1Address.String())
			}
		})
	})

	t.Run("ChannelMembers", func(t *testing.T) {
		// admin can add members
		// admin can add admin
		// member cannot add
		// admin can remove members
		// members can't add
		// newsroom multisig members can add themselves?
		t.Skip("not implemented")
	})

	t.Run("IsChannelAdmin", func(t *testing.T) {
		u1 := randomUUID()
		u2 := randomUUID()
		channel, err := svc.CreateUserChannel(u1)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		isAdmin, err := svc.IsChannelAdmin(u1, channel.ID)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}
		if !isAdmin {
			t.Fatalf("isAdmin should be true")
		}

		isAdmin, err = svc.IsChannelAdmin(u2, channel.ID)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}
		if isAdmin {
			t.Fatalf("isAdmin should be false")
		}
	})

	t.Run("ConnectStripe", func(t *testing.T) {
		u1 := randomUUID()
		u2 := randomUUID()
		channel, err := svc.CreateUserChannel(u1)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		// invalid input, no OAuthCode
		_, err = svc.ConnectStripe(u1, channels.ConnectStripeInput{ChannelID: channel.ID})
		if err != channels.ErrorsInvalidInput {
			t.Fatalf("was expecting ErrorsInvalidInput: %v", err)
		}

		// stripe connect function returns an error
		_, err = svc.ConnectStripe(u1, channels.ConnectStripeInput{ChannelID: channel.ID, OAuthCode: "fail"})
		if err != channels.ErrorStripeIssue {
			t.Fatalf("was expecting ErrorStripeIssue: %v", err)
		}

		// make sure stripe account is not set
		result, err := svc.GetStripePaymentAccount(channel.ID)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}
		if result != "" {
			t.Fatalf("expected stripeAccountID to be %v but is: %v", "", result)
		}

		// this should succeed
		_, err = svc.ConnectStripe(u1, channels.ConnectStripeInput{ChannelID: channel.ID, OAuthCode: "test"})
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		// u2 is not an admin, so this should error
		_, err = svc.ConnectStripe(u2, channels.ConnectStripeInput{ChannelID: channel.ID, OAuthCode: "test"})
		if err != channels.ErrorUnauthorized {
			t.Fatalf("was expecting ErrorUnauthorized: %v", err)
		}

		result, err = svc.GetStripePaymentAccount(channel.ID)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}
		if result != stripeAccountID {
			t.Fatalf("expected stripeAccountID to be %v but is: %v", stripeAccountID, result)
		}

	})

	t.Run("SendEmailConfirmation", func(t *testing.T) {
		u1 := randomUUID()
		randomInt := r.Int31()
		invalidEmail := fmt.Sprintf("tEst%v", randomInt)
		validEmail := "tEst@civil.co"

		// create channelf for u1
		channel1, err := svc.CreateUserChannel(u1)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		// don't allow invalid email
		_, err = svc.SendEmailConfirmation(u1, channel1.ID, invalidEmail, channels.SetEmailEnumUser)
		if err != channels.ErrorInvalidEmail {
			t.Fatalf("was expecting ErrorInvalidEmail: %v", err)
		}

		// allow valid handle
		_, err = svc.SendEmailConfirmation(u1, channel1.ID, validEmail, channels.SetEmailEnumUser)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}
	})

	t.Run("SetEmailConfirm", func(t *testing.T) {
		u1 := randomUUID()
		u2 := randomUUID()
		randomInt := r.Int31()
		invalidEmail := fmt.Sprintf("tEst%v", randomInt)
		validEmail := "tEst@civil.co"

		// create channelf for u1
		channel1, err := svc.CreateUserChannel(u1)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		// create channelf for u2
		channel2, err := svc.CreateUserChannel(u2)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		parts := []string{validEmail, string(channels.SetEmailEnumUser), u1, channel1.ID}
		sub := strings.Join(parts, "||")
		token, err := generator.GenerateToken(sub, 360)

		if err != nil {
			t.Fatalf("not expecting error generating token: %v", err)
		}
		// set email for user channel correctly
		_, err = svc.SetEmailConfirm(token)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		updatedChannel1, err := svc.GetChannel(channel1.ID)
		if err != nil {
			t.Fatalf("error getting updated channel: %v", err)
		}
		channelEmail := updatedChannel1.EmailAddress
		if channelEmail != validEmail {
			t.Fatalf("channel email not set to correct email. channelEmail: %v", channelEmail)
		}

		parts = []string{validEmail, string(channels.SetEmailEnumUser), u1, channel2.ID}
		sub = strings.Join(parts, "||")
		token, err = generator.GenerateToken(sub, 360)
		if err != nil {
			t.Fatalf("was not expecting error. err: %v", err)
		}
		// don't set email for wrong channel
		_, err = svc.SetEmailConfirm(token)
		if err != channels.ErrorUnauthorized {
			t.Fatalf("was expecting ErrorUnauthorized. Error: %v", err)
		}

		parts = []string{invalidEmail, string(channels.SetEmailEnumUser), u1, channel1.ID}
		sub = strings.Join(parts, "||")
		token, err = generator.GenerateToken(sub, 360)
		if err != nil {
			t.Fatalf("was not expecting error. err: %v", err)
		}
		// don't set invalid email
		_, err = svc.SetEmailConfirm(token)
		if err != channels.ErrorInvalidEmail {
			t.Fatalf("was expecting ErrorInvalidEmail. Error: %v", err)
		}
	})
}

var handletests = []struct {
	in  string
	out bool
}{
	{"foo", false},
	{"food", true},
	{"foofoo", true},
	{"barfoo", true},
	{"foo_bar", true},
	{"foo%bar", false},
	{"F00_f", true},
	{"F00-f", false},
	{"F_1", false},
	{"F/AA", false},
	{"hello world", false},
}

func TestHandles(t *testing.T) {
	for _, tt := range handletests {
		t.Run(tt.in, func(t *testing.T) {
			s := channels.IsValidHandle(tt.in)
			if s != tt.out {
				t.Errorf("got %v, want %v", s, tt.out)
			}
		})
	}
}

var emailtests = []struct {
	in  string
	out bool
}{
	{"foo", false},
	{"bar@bar.com", true},
	{"foo_bar@bar.asdfew", true},
	{"foo-bar@bar.asdfew", true},
	{"foo_bar+1@bar.asdfew", true},
	{"foo%bar", false},
	{"foo%bar@bar.asdf", true},
	{"foo.bar@bar.asdf", true},
	{"foo?bar@bar.asdf", true},
	{"F00@a.ae", true},
	{"F-11", false},
	{"F-11@a", false},
	{"F/AA", false},
	{"hello world", false},
}

func TestEmails(t *testing.T) {
	for _, tt := range emailtests {
		t.Run(tt.in, func(t *testing.T) {
			s := channels.IsValidEmail(tt.in)
			if s != tt.out {
				t.Errorf("got %v, want %v", s, tt.out)
			}
		})
	}
}
