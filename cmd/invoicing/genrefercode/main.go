package main

// Simple script to generate a referral code.  Used to generate them
// for newsrooms or for one-offs.  Does not store this value anywhere, so need to
// record it elsewhere.

import (
	"fmt"

	"github.com/joincivil/civil-api-server/pkg/invoicing"
)

func main() {
	// Generate a new code for a user
	referralCode, err := invoicing.GenerateReferralCode()
	if err != nil {
		fmt.Printf("error generating referral code: err: %v\n", err)
		return
	}
	fmt.Printf("Referral Code: %v\n", referralCode)
}
