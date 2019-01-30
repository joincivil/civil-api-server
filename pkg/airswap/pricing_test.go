package airswap_test

import (
	"testing"

	"github.com/joincivil/civil-api-server/pkg/airswap"
)

func TestGetQuote(t *testing.T) {
	totalOffering := 34000000.0
	totalRaiseUSD := 20000000.0
	startingPrice := 0.2
	manager := airswap.NewPricingManager(totalOffering, totalRaiseUSD, startingPrice)

	// buying 0 tokens should return 0
	quote := manager.GetQuote(0.0)
	if quote != 0.0 {
		t.Fatalf("expected quote to be 0 but was %v", quote)
	}

	// buying 1 token should return 0.2
	quote = manager.GetQuote(1.0)
	if quote != 0.2 {
		t.Fatalf("expected quote to be 0.2 but was %v", quote)
	}

	// buying 1 token should return 0.2 (plus a super small fraction of a penny)
	quote = manager.GetQuote(totalOffering)
	if quote != totalRaiseUSD {
		t.Fatalf("expected quote to be %v but was %v", totalRaiseUSD, quote)
	}

	// getting a quote for (X+Y) tokens should be the same as the sum of X tokens and Y tokens
	qT := manager.GetQuote(6000)
	q1 := manager.GetQuote(1000)
	manager.IncreaseTokensSold(1000)
	q2 := manager.GetQuote(5000)

	if (q1 + q2) != qT {
		t.Fatalf("expected the sum of q1 and q2 to be the same as qT %v | %v | %v | %v", qT, q1, q2, q1+q2)
	}

}

func TestGetTokensToBuy(t *testing.T) {
	totalOffering := 34000000.0
	totalRaiseUSD := 20000000.0
	startingPrice := 0.2
	manager := airswap.NewPricingManager(totalOffering, totalRaiseUSD, startingPrice)

	// spending $0 should result in 0 tokens
	tokens := manager.GetTokensToBuy(0.0)
	if tokens != 0.0 {
		t.Fatalf("expected quote to be 0.0 but was %v", tokens)
	}

	// buying 1 token should return 0.2
	tokens = manager.GetTokensToBuy(0.2)
	if tokens != 1 {
		t.Fatalf("expected tokens to be 1.0 but was %v", tokens)
	}

	// buying 1 token should return 0.2 (plus a super small fraction of a penny)
	tokens = manager.GetTokensToBuy(totalRaiseUSD)
	if tokens != totalOffering {
		t.Fatalf("expected tokens to be %v but was %v", totalOffering, tokens)
	}

	// getting a quote for (X+Y) tokens should be the same as the sum of X tokens and Y tokens
	qT := manager.GetTokensToBuy(6000)
	q1 := manager.GetTokensToBuy(1000)
	manager.IncreaseTokensSold(q1)
	q2 := manager.GetTokensToBuy(5000)

	if (q1 + q2) != qT {
		t.Fatalf("expected the sum of q1 and q2 to be the same as qT %v | %v | %v | %v", qT, q1, q2, q1+q2)
	}

}
