package channels_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/testutils"
	uuid "github.com/satori/go.uuid"
)

func buildService(t *testing.T) *channels.Service {
	db, err := testutils.GetTestDBConnection()
	db.AutoMigrate(&channels.Channel{}, &channels.ChannelMember{})
	db.Unscoped().Delete(&channels.Channel{})
	if err != nil {
		t.Fatalf("error getting DB")
	}
	persister := channels.NewDBPersister(db)
	return channels.NewService(persister)
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
	t.Run("newsroom type", func(t *testing.T) {
		userUUID, err := uuid.NewV4()
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}
		userID := userUUID.String()

		input := channels.CreateNewsroomChannelInput{
			ContractAddress: fmt.Sprintf("testaddress%v", rand.Intn(10000)),
		}
		_, err = svc.CreateNewsroomChannel(userID, input)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		// only allow a newsroom to have 1 channel
		_, err = svc.CreateNewsroomChannel(userID, input)
		if err != channels.ErrorNotUnique {
			t.Fatalf("was expecting ErrorNotUnique")
		}

		// TODO(dankins): make sure that user is on the newsroom smart contract

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
