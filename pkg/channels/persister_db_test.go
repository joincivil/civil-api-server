package channels_test

import (
	"github.com/joincivil/civil-api-server/pkg/channels"
	"github.com/joincivil/civil-api-server/pkg/testruntime"
	"github.com/joincivil/civil-api-server/pkg/testutils"
	"github.com/joincivil/civil-api-server/pkg/utils"
	"github.com/joincivil/go-common/pkg/email"
	"testing"
)

// This test uses a USER channel instead of a newsroom channel, to avoid having to set up a listing. The `SetNewsroomHandleOnAccepted` and
// `ClearNewsroomHandleOnRemoved` functions should never be called on a user channel or by anything but the governance event handler
func TestNewsroomHandle(t *testing.T) {
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

	channel, err = persister.SetNewsroomHandleOnAccepted(channel.ID, "test-handle")
	if err != nil {
		t.Fatalf("error setting channel handle via SetNewsroomHandleOnAccepted")
	}
	if *(channel.Handle) != "test-handle" {
		t.Fatalf("error setting test-handle")
	}

	channel, err = persister.ClearNewsroomHandleOnRemoved(channel.ID)
	if err != nil {
		t.Fatalf("error clearing channel handle via ClearNewsroomHandleOnRemoved")
	}
	if channel.Handle != nil {
		t.Fatalf("error clearing test-handle")
	}
}