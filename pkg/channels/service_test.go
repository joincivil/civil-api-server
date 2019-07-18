package channels_test

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/testruntime"
	"github.com/joincivil/civil-api-server/pkg/testutils"
	uuid "github.com/satori/go.uuid"
)

var (
	user1ID           = "ceaf1993-2912-4b02-b8c4-3809aa9a12fb"
	user1Address      = common.HexToAddress("0xa6210D155a96dd11604626a86D7498a3d0959932")
	user2ID           = "9d87acbb-d876-48b2-90ac-78820bd1d88c"
	user2Address      = common.HexToAddress("0xD55f67fFEE825e98c9ed2223505755eFF22Ea297")
	newsroom1Address  = common.HexToAddress("0xc187a99320195EF6c83894e5920577dad306e3Dd")
	newsroom1Multisig = common.HexToAddress("0x030a7DDB19D3cEc37031838aa77830CCFC495372")
	newsroom2Address  = common.HexToAddress("0x259439bC023932a75e41D5d1e3Aa7D60Bb20A6C8")
	newsroom2Multisig = common.HexToAddress("0x6626F2c20dc807892c20d94507947aeec0aeDCB2")
)

type MockGetNewsroomHelper struct{}

func (g MockGetNewsroomHelper) GetMultisigMembers(newsroomAddress common.Address) ([]common.Address, error) {
	if newsroomAddress == newsroom1Address {
		return []common.Address{user1Address}, nil
	} else if newsroomAddress == newsroom2Address {
		return []common.Address{user2Address}, nil
	}
	return nil, errors.New("found found")
}
func (g MockGetNewsroomHelper) GetOwner(newsroomAddress common.Address) (common.Address, error) {
	if newsroomAddress == newsroom1Address {
		return newsroom1Multisig, nil
	} else if newsroomAddress == newsroom2Address {
		return newsroom2Multisig, nil
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

func buildService(t *testing.T) *channels.Service {
	db, err := testutils.GetTestDBConnection()
	if err != nil {
		t.Fatalf("error getting DB: %v", err)
	}
	err = testruntime.CleanDatabase(db)
	if err != nil {
		t.Fatalf("error cleaning DB: %v", err)
	}

	persister := channels.NewDBPersister(db)
	return channels.NewService(persister, MockGetNewsroomHelper{}, MockUserEthAddressGetter{})
}

func TestCreateChannel(t *testing.T) {
	svc := buildService(t)
	userUUID, err := uuid.NewV4()
	if err != nil {
		t.Fatalf("not expecting error: %v", err)
	}
	userID := userUUID.String()

	channel, err := svc.CreateUserChannel(userID)
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

		if member.UserID != userID {
			t.Fatal("initial member should be the creator")
		}
		if member.Role != channels.RoleAdmin {
			t.Fatal("initial role should be `admin`")
		}
	})

	t.Run("user type", func(t *testing.T) {
		userUUID, err := uuid.NewV4()
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}
		userID := userUUID.String()

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
		userUUID, err := uuid.NewV4()
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}
		userID := userUUID.String()
		handle := fmt.Sprintf("test%v", rand.Intn(10000))

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
			userID := userUUID.String()
			_, err := svc.CreateNewsroomChannel(userID, channels.CreateNewsroomChannelInput{
				ContractAddress: newsroomAddress,
			})
			if err != channels.ErrorInvalidHandle {
				t.Fatalf("was expecting error `channels.ErrorInvalidHandle`")
			}
		})

		t.Run("user doesn't have ethereum address", func(t *testing.T) {
			newsroomAddress := newsroom1Address.String()
			userID := userUUID.String()
			_, err := svc.CreateNewsroomChannel(userID, channels.CreateNewsroomChannelInput{
				ContractAddress: newsroomAddress,
			})
			if err != channels.ErrorUnauthorized {
				t.Fatalf("was expecting error `channels.ErrorUnauthorized` but received: %v", err)
			}
		})
		t.Run("user not member of multisig", func(t *testing.T) {
			newsroomAddress := newsroom1Address.String()
			_, err := svc.CreateNewsroomChannel(user2ID, channels.CreateNewsroomChannelInput{
				ContractAddress: newsroomAddress,
			})
			if err != channels.ErrorUnauthorized {
				t.Fatalf("was expecting error `channels.ErrorUnauthorized`")
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
