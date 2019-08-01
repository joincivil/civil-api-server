package payments_test

import (
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jinzhu/gorm"
	"github.com/joincivil/civil-api-server/pkg/payments"
	"github.com/joincivil/civil-api-server/pkg/testruntime"
	"github.com/joincivil/go-common/pkg/email"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

const (
	sendGridKeyEnvVar = "SENDGRID_TEST_KEY"
	useSandbox        = true
)

func getSendGridKeyFromEnvVar() string {
	return os.Getenv(sendGridKeyEnvVar)
}

func TestService(t *testing.T) {
	sendGridKey := getSendGridKeyFromEnvVar()
	if sendGridKey == "" {
		t.Log("No SENDGRID_TEST_KEY set, skipping signup email test")
		return
	}
	var paymentService *payments.Service
	var transactions *testruntime.MockTransactionReader
	var paymentHelper *testruntime.MockPaymentHelper
	emailer := email.NewEmailerWithSandbox(sendGridKey, useSandbox)
	var db *gorm.DB
	app := fxtest.New(t,
		testruntime.TestModule,
		fx.Populate(&paymentService),
		fx.Populate(&transactions),
		fx.Populate(&paymentHelper),
		fx.Populate(&db),
		fx.Populate(&emailer),
	)
	app.RequireStart().RequireStop()
	err := testruntime.RunMigrations(db)
	if err != nil {
		t.Fatalf("error running migrations: %v", err)
	}
	err = db.Unscoped().Delete(&payments.PaymentModel{}).Error
	if err != nil {
		t.Fatalf("error cleaning database: %v", err)
	}

	channelID := "test"
	channelAddress, _ := paymentHelper.GetEthereumPaymentAddress("foo")
	ownerType := "post"
	ownerID, err := uuid.NewV4()
	if err != nil {
		t.Fatalf("error creating uuid, err: %v", err)
	}

	var nonce uint64
	makeTx := func(status string, to common.Address) *types.Transaction {
		nonce++

		transaction := types.NewTransaction(nonce, to, big.NewInt(1*1e18), 10, big.NewInt(1), []byte{})
		transactions.AddTransaction(transaction.Hash(), transaction)

		if status == "complete" {
			receipt := &types.Receipt{
				Status: 1,
			}
			transactions.AddReceipt(transaction.Hash(), receipt)
		} else if status == "failed" {
			receipt := &types.Receipt{
				Status: 0,
			}
			transactions.AddReceipt(transaction.Hash(), receipt)
		}
		return transaction
	}

	t.Run("UpdateEtherPayment complete", func(t *testing.T) {
		tx := makeTx("complete", channelAddress)

		// create the payment
		p, err := paymentService.CreateEtherPayment(channelID, ownerType, ownerID.String(), tx.Hash().String(), "", nil)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		if p.Status != "pending" {
			t.Fatalf("expecting payment status to be pending but is: %v", p.Status)
		}

		err = paymentService.UpdateEtherPayment(&p.PaymentModel)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}
	})

	t.Run("UpdateEtherPayment complete with email", func(t *testing.T) {
		tx := makeTx("complete", channelAddress)

		// create the payment
		p, err := paymentService.CreateEtherPayment(channelID, ownerType, ownerID.String(), tx.Hash().String(), "nick@joincivil.com", nil)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		if p.Status != "pending" {
			t.Fatalf("expecting payment status to be pending but is: %v", p.Status)
		}

		err = paymentService.UpdateEtherPayment(&p.PaymentModel)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}
	})

	t.Run("UpdateEtherPayment complete with bad email", func(t *testing.T) {
		tx := makeTx("complete", channelAddress)

		// create the payment
		p, err := paymentService.CreateEtherPayment(channelID, ownerType, ownerID.String(), tx.Hash().String(), "234-2", nil)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		if p.Status != "pending" {
			t.Fatalf("expecting payment status to be pending but is: %v", p.Status)
		}

		err = paymentService.UpdateEtherPayment(&p.PaymentModel)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}
	})

	t.Run("UpdateEtherPayment failed", func(t *testing.T) {

		tx := makeTx("failed", channelAddress)
		// create the payment
		p, err := paymentService.CreateEtherPayment(channelID, ownerType, ownerID.String(), tx.Hash().String(), "", nil)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		err = paymentService.UpdateEtherPayment(&p.PaymentModel)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		retrieved, err := paymentService.GetPayment(p.ID)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		if retrieved.(*payments.EtherPayment).PaymentModel.Status != "failed" {
			t.Fatalf("status should be failed")
		}
	})

	//
	t.Run("UpdateEtherPayment pending", func(t *testing.T) {

		tx := makeTx("pending", channelAddress)
		// create the payment
		p, err := paymentService.CreateEtherPayment(channelID, ownerType, ownerID.String(), tx.Hash().String(), "", nil)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		err = paymentService.UpdateEtherPayment(&p.PaymentModel)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		retrieved, err := paymentService.GetPayment(p.ID)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		if retrieved.(*payments.EtherPayment).PaymentModel.Status != "pending" {
			t.Fatalf("status should be pending")
		}
	})

	t.Run("UpdateEtherPayment invalid", func(t *testing.T) {

		tx := makeTx("complete", common.HexToAddress("deadbeef"))
		// create the payment
		p, err := paymentService.CreateEtherPayment(channelID, ownerType, ownerID.String(), tx.Hash().String(), "", nil)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		err = paymentService.UpdateEtherPayment(&p.PaymentModel)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		retrieved, err := paymentService.GetPayment(p.ID)
		if err != nil {
			t.Fatalf("not expecting error: %v", err)
		}

		if retrieved.(*payments.EtherPayment).PaymentModel.Status != "invalid" {
			t.Fatalf("status should be invalid")
		}
	})
}
