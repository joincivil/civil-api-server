package channels_test

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/testruntime"
	"github.com/joincivil/civil-api-server/pkg/testutils"
	uuid "github.com/satori/go.uuid"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

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

type MockUserEthAddressGetter struct{}

func (g MockUserEthAddressGetter) GetETHAddresses(userID string) ([]common.Address, error) {
	if userID == user1ID {
		return []common.Address{user1Address}, nil
	} else if userID == user2ID {
		return []common.Address{user2Address}, nil
	}
	return nil, errors.New("not found")
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
	svc := channels.NewService(persister, MockGetNewsroomHelper{}, MockUserEthAddressGetter{})

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
		handle := fmt.Sprintf("test%v", r.Int31())

		_, err = svc.CreateGroupChannel(userID, handle)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		// don't allow if handle already exists
		_, err = svc.CreateGroupChannel(userID, handle)
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
			_, err := svc.CreateNewsroomChannel(user1ID, channels.CreateNewsroomChannelInput{
				ContractAddress: newsroomAddress,
			})
			if err != channels.ErrorInvalidHandle {
				t.Fatalf("was expecting error `channels.ErrorInvalidHandle`")
			}
		})

		t.Run("user doesn't have ethereum address", func(t *testing.T) {
			newsroomAddress := newsroom1Address.String()
			userID := randomUUID()
			_, err := svc.CreateNewsroomChannel(userID, channels.CreateNewsroomChannelInput{
				ContractAddress: newsroomAddress,
			})
			if err != channels.ErrorUnauthorized {
				t.Fatalf("was expecting error `channels.ErrorUnauthorized` but received: %v", err)
			}
		})
		t.Run("user not member of multisig", func(t *testing.T) {
			newsroomAddress := newsroom3Address.String()
			_, err := svc.CreateNewsroomChannel(user2ID, channels.CreateNewsroomChannelInput{
				ContractAddress: newsroomAddress,
			})
			if err != channels.ErrorUnauthorized {
				t.Fatalf("was expecting error `channels.ErrorUnauthorized` but received: %v", err)
			}
		})

		t.Run("success", func(t *testing.T) {
			channel, err := svc.CreateNewsroomChannel(user1ID, channels.CreateNewsroomChannelInput{
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
}

func TestChannelMembers(t *testing.T) {
	// admin can add members
	// admin can add admin
	// member cannot add
	// admin can remove members
	// members can't add
	t.Skip("not implemented")
}

var handletests = []struct {
	in  string
	out bool
}{
	{"foo", true},
	{"bar", true},
	{"foo_bar", true},
	{"foo%bar", false},
	{"F00", true},
	{"F-11", false},
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
