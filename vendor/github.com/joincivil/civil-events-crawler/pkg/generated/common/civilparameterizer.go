// Code generated by 'gen/eventhandlergen.go'  DO NOT EDIT.
// IT SHOULD NOT BE EDITED BY HAND AS ANY CHANGES MAY BE OVERWRITTEN
// Please reference 'gen/filterergen_template.go' for more details
// File was generated at 2019-03-20 17:29:21.820534 +0000 UTC
package common

var eventTypesParameterizerContract = []string{
	"ChallengeFailed",
	"ChallengeSucceeded",
	"NewChallenge",
	"ProposalAccepted",
	"ProposalExpired",
	"ReparameterizationProposal",
	"RewardClaimed",
}

func EventTypesParameterizerContract() []string {
	tmp := make([]string, len(eventTypesParameterizerContract))
	copy(tmp, eventTypesParameterizerContract)
	return tmp
}
