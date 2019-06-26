package processor

import (
	"math/big"
	"strings"

	log "github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	commongen "github.com/joincivil/civil-events-crawler/pkg/generated/common"
	crawlermodel "github.com/joincivil/civil-events-crawler/pkg/model"

	cbytes "github.com/joincivil/go-common/pkg/bytes"
	cerrors "github.com/joincivil/go-common/pkg/errors"
	"github.com/joincivil/go-common/pkg/generated/contract"
	cpersist "github.com/joincivil/go-common/pkg/persistence"
	ctime "github.com/joincivil/go-common/pkg/time"

	"github.com/joincivil/civil-events-processor/pkg/model"
	"github.com/joincivil/civil-events-processor/pkg/scraper"
)

const (
	listingNameFieldName    = "Name"
	ownerAddressesFieldName = "OwnerAddresses"
	urlFieldName            = "URL"

	defaultCharterContentID = 0
	// approvalDateNoUpdate    = int64(-1)
	approvalDateEmptyValue = int64(0)
)

// NewNewsroomEventProcessor is a convenience function to init an EventProcessor
func NewNewsroomEventProcessor(client bind.ContractBackend, listingPersister model.ListingPersister,
	revisionPersister model.ContentRevisionPersister, errRep cerrors.ErrorReporter) *NewsroomEventProcessor {
	return &NewsroomEventProcessor{
		client:            client,
		listingPersister:  listingPersister,
		revisionPersister: revisionPersister,
		errRep:            errRep,
	}
}

// NewsroomEventProcessor handles the processing of raw events into aggregated data
// for use via the API.
type NewsroomEventProcessor struct {
	client            bind.ContractBackend
	listingPersister  model.ListingPersister
	revisionPersister model.ContentRevisionPersister
	errRep            cerrors.ErrorReporter
}

// Process processes Newsroom Events into aggregated data
func (n *NewsroomEventProcessor) Process(event *crawlermodel.Event) (bool, error) {
	if !n.isValidNewsroomContractEventName(event.EventType()) {
		return false, nil
	}

	var err error
	ran := true
	eventName := strings.Trim(event.EventType(), " _")

	// Handling all the actionable events from Newsroom Addressses
	switch eventName {
	// When a listing's name has changed
	case "NameChanged":
		log.Infof("Handling NameChanged for %v\n", event.ContractAddress().Hex())
		err = n.processNewsroomNameChanged(event)

	// When there is a new revision on content
	case "RevisionUpdated":
		log.Infof("Handling RevisionUpdated for %v\n", event.ContractAddress().Hex())
		err = n.processNewsroomRevisionUpdated(event)

	// When there is a new owner
	case "OwnershipTransferred":
		log.Infof("Handling OwnershipTransferred for %v\n", event.ContractAddress().Hex())
		err = n.processNewsroomOwnershipTransferred(event)

	default:
		ran = false
	}
	return ran, err
}

func (n *NewsroomEventProcessor) isValidNewsroomContractEventName(name string) bool {
	name = strings.Trim(name, " _")
	eventNames := commongen.EventTypesNewsroomContract()
	return isStringInSlice(eventNames, name)
}

func (n *NewsroomEventProcessor) processNewsroomNameChanged(event *crawlermodel.Event) error {
	var updatedFields []string
	payload := event.EventPayload()
	listing, err := n.retrieveOrCreateListingForNewsroomEvent(event)
	if err != nil {
		return errors.Wrap(err, "error retrieving listing or creating by address")
	}
	name, ok := payload["NewName"]
	if !ok {
		return errors.New("No NewName field found")
	}
	listing.SetName(name.(string))
	updatedFields = append(updatedFields, listingNameFieldName)
	err = n.listingPersister.UpdateListing(listing, updatedFields)
	return err
}

func (n *NewsroomEventProcessor) processNewsroomRevisionUpdated(event *crawlermodel.Event) error {
	// Create a new listing if none exists for the address in the event
	_, err := n.retrieveOrCreateListingForNewsroomEvent(event)
	if err != nil {
		return errors.WithMessage(err, "error retrieving listing or creating by address")
	}

	payload := event.EventPayload()
	listingAddress := event.ContractAddress()

	editorAddress, ok := payload["Editor"]
	if !ok {
		return errors.New("No editor address found")
	}
	contentID, ok := payload["ContentId"]
	if !ok {
		return errors.New("No content id found")
	}
	revisionID, ok := payload["RevisionId"]
	if !ok {
		return errors.New("No revision id found")
	}
	// Metadata URI
	revisionURI, ok := payload["Uri"]
	if !ok {
		return errors.New("No revision uri found")
	}

	// Pull data from the newsroom contract
	newsroom, err := contract.NewNewsroomContract(listingAddress, n.client)
	if err != nil {
		return errors.WithMessage(err, "error creating newsroom contract")
	}

	content, err := newsroom.GetContent(&bind.CallOpts{}, contentID.(*big.Int))
	if err != nil {
		return errors.WithMessage(err, "error retrieving newsroom content")
	}
	contentHash := cbytes.Byte32ToHexString(content.ContentHash)

	// Scrape the metadata or content for the revision
	metadata, scraperContent, err := n.scrapeData(contentID.(*big.Int), revisionURI.(string))
	if err != nil {
		log.Errorf("Error scraping data: err: %v", err)
	}

	articlePayload := model.ArticlePayload{}
	if metadata != nil || scraperContent != nil {
		articlePayload = n.scraperDataToPayload(metadata, scraperContent)
	}

	// Store the new revision
	revision := model.NewContentRevision(
		listingAddress,
		articlePayload,
		contentHash,
		editorAddress.(common.Address),
		contentID.(*big.Int),
		revisionID.(*big.Int),
		revisionURI.(string),
		event.Timestamp(),
	)

	err = n.revisionPersister.CreateContentRevision(revision)
	if err != nil {
		return err
	}

	// If the revision is for the charter, need to update the data in the listing.
	if contentID.(*big.Int).Int64() == defaultCharterContentID {
		err = n.updateListingCharterRevision(revision)
	}
	return err
}

func (n *NewsroomEventProcessor) processNewsroomOwnershipTransferred(event *crawlermodel.Event) error {
	var updatedFields []string
	payload := event.EventPayload()
	listing, err := n.retrieveOrCreateListingForNewsroomEvent(event)
	if err != nil {
		return err
	}
	previousOwner, ok := payload["PreviousOwner"]
	if !ok {
		return errors.New("No previous owner found")
	}
	newOwner, ok := payload["NewOwner"]
	if !ok {
		return errors.New("No new owner found")
	}
	listing.RemoveOwnerAddress(previousOwner.(common.Address))
	listing.AddOwnerAddress(newOwner.(common.Address))
	updatedFields = append(updatedFields, ownerAddressesFieldName)
	return n.listingPersister.UpdateListing(listing, updatedFields)
}

func (n *NewsroomEventProcessor) updateListingCharterRevision(revision *model.ContentRevision) error {
	if revision.ContractContentID().Int64() != int64(defaultCharterContentID) {
		return errors.New("incorrect content revision ID for charter")
	}

	listing, err := n.listingPersister.ListingByAddress(revision.ListingAddress())
	if err != nil {
		return err
	}

	newsroom, newsErr := contract.NewNewsroomContract(revision.ListingAddress(), n.client)
	if newsErr != nil {
		return errors.WithMessage(err, "error reading from Newsroom contract")
	}

	charterContent, contErr := newsroom.GetRevision(
		&bind.CallOpts{},
		big.NewInt(defaultCharterContentID),
		revision.ContractRevisionID(),
	)
	if contErr != nil {
		return errors.WithMessage(contErr, "Error getting charter revision from Newsroom contract")
	}

	updatedFields := []string{"Charter"}
	updatedCharter := model.NewCharter(&model.CharterParams{
		URI:         revision.RevisionURI(),
		ContentID:   big.NewInt(defaultCharterContentID),
		RevisionID:  revision.ContractRevisionID(),
		Signature:   charterContent.Signature,
		Author:      charterContent.Author,
		ContentHash: charterContent.ContentHash,
		Timestamp:   charterContent.Timestamp,
	})
	listing.SetCharter(updatedCharter)

	// Try to get the charter data to get newsroom URL
	// TODO(PN): Perhaps use the content revision here? Might not be in DB at this point.
	_, charterData, chartErr := n.scrapeData(big.NewInt(defaultCharterContentID), revision.RevisionURI())
	if chartErr != nil {
		log.Errorf("Error retrieving charter data from %v: err: %v", revision.RevisionURI(), chartErr)
		n.errRep.Error(errors.Wrapf(chartErr, "error retrieving charter data from %v", revision.RevisionURI()), nil)
	}
	if charterData != nil {
		nrURL, ok := charterData.Data()["newsroomUrl"]
		if !ok {
			log.Errorf("Could not find newsroomUrl in the charter data: %v", charterData.URI())
			n.errRep.Error(errors.Errorf("could not find newsroomUrl in charter data: %v", charterData.URI()), nil)
		} else {
			listing.SetURL(nrURL.(string))
			updatedFields = append(updatedFields, urlFieldName)
		}
	}

	return n.listingPersister.UpdateListing(listing, updatedFields)
}

func (n *NewsroomEventProcessor) retrieveOrCreateListingForNewsroomEvent(event *crawlermodel.Event) (*model.Listing, error) {
	listingAddress := event.ContractAddress()
	listing, err := n.listingPersister.ListingByAddress(listingAddress)
	if err != nil && err != cpersist.ErrPersisterNoResults {
		return nil, errors.WithMessage(err, "error retrieving or creating listing")
	}
	if listing != nil {
		return listing, nil
	}
	// If a listing doesn't exist, create a new one from contract. This shouldn't happen if events are ordered
	log.Infof("Listing not found in persistence for %v, events may be processed out of order\n", listingAddress.Hex())
	listing, err = n.persistNewListing(listingAddress)
	return listing, err
}

func (n *NewsroomEventProcessor) persistNewListing(listingAddress common.Address) (*model.Listing, error) {
	// NOTE: This is the function that is called to get data from newsroom contract
	// in the case events are out of order and persists a listing

	// We retrieve the URL from the charter data in IPFS/content revision
	url := ""

	// charter is the first content item in the newsroom contract
	charterContentID := big.NewInt(defaultCharterContentID)
	newsroom, newsErr := contract.NewNewsroomContract(listingAddress, n.client)
	if newsErr != nil {
		return nil, errors.Errorf("error reading from Newsroom contract: %v ", newsErr)
	}
	name, nameErr := newsroom.Name(&bind.CallOpts{})
	if nameErr != nil {
		return nil, errors.Errorf("error getting Name from Newsroom contract: %v ", nameErr)
	}

	revisionCount, countErr := newsroom.RevisionCount(&bind.CallOpts{}, charterContentID)
	if countErr != nil {
		return nil, errors.Errorf("error getting RevisionCount from Newsroom contract: %v ", countErr)
	}
	if revisionCount.Int64() <= 0 {
		return nil, errors.Errorf("error there are no revisions for the charter: addr: %v", listingAddress)
	}

	// latest revision should be total revisions - 1 for index
	latestRevisionID := big.NewInt(revisionCount.Int64() - 1)
	charterContent, contErr := newsroom.GetRevision(&bind.CallOpts{}, charterContentID, latestRevisionID)
	if contErr != nil {
		return nil, errors.Errorf("Error getting charter revision from Newsroom contract: %v ", contErr)
	}

	// Try to get the charter data to get newsroom URL
	// TODO(PN): Perhaps use the content revision here? Might not be in DB at this point.
	_, charterData, chartErr := n.scrapeData(big.NewInt(defaultCharterContentID), charterContent.Uri)
	if chartErr != nil {
		log.Errorf("Error retrieving charter data from %v: err: %v", charterContent.Uri, chartErr)
		n.errRep.Error(errors.Wrapf(chartErr, "error retrieving charter data from %v", charterContent.Uri), nil)
	}
	if charterData != nil {
		nrURL, ok := charterData.Data()["newsroomUrl"]
		if !ok {
			log.Errorf("Could not find newsroomUrl in the charter data: %v", charterData.URI())
			n.errRep.Error(errors.Errorf("could not find newsroomUrl in charter data: %v", charterData.URI()), nil)
		} else {
			url = nrURL.(string)
		}
	}

	charter := model.NewCharter(&model.CharterParams{
		URI:         charterContent.Uri,
		ContentID:   charterContentID,
		RevisionID:  latestRevisionID,
		Signature:   charterContent.Signature,
		Author:      charterContent.Author,
		ContentHash: charterContent.ContentHash,
		Timestamp:   charterContent.Timestamp,
	})

	charterAuthorAddr := charterContent.Author
	ownerAddr, err := newsroom.Owner(&bind.CallOpts{})
	if err != nil {
		return nil, err
	}
	ownerAddresses := []common.Address{ownerAddr}
	contributorAddresses := []common.Address{charterAuthorAddr}
	// NOTE(IS): The values in newlistingparams which aren't initialized will initialize to nil
	// values. We can't get this data from the tcr contract since we don't have a TCR address.
	listing := model.NewListing(&model.NewListingParams{
		Name:            name,
		ContractAddress: listingAddress,
		// Whitelisted:          whitelisted,
		// LastState:            lastGovernanceState,
		URL:     url,
		Charter: charter,
		// Owner:                ownerAddr,
		OwnerAddresses:       ownerAddresses,
		ContributorAddresses: contributorAddresses,
		// CreatedDateTs:        creationDate,
		// ApplicationDateTs:    applicationDate,
		// ApprovalDateTs:       approvalDate,
		LastUpdatedDateTs: ctime.CurrentEpochSecsInInt64(),
	})
	err = n.listingPersister.CreateListing(listing)
	return listing, err
}

func (n *NewsroomEventProcessor) scrapeData(contentID *big.Int, revisionURI string) (
	*model.ScraperCivilMetadata, *model.ScraperContent, error) {
	if revisionURI == "" {
		return nil, nil, nil
	}

	// Ignore self-tx links for now
	// This is context embedded in the transaction input data

	// Basic IPFS charter support
	// Charter is content 0
	if strings.Contains(revisionURI, "ipfs://") && contentID.Int64() == 0 {
		charterScraper := &scraper.CharterIPFSScraper{}
		charterContent, err := charterScraper.ScrapeContent(revisionURI)
		if err != nil {
			return nil, nil, err
		}
		return nil, charterContent, nil

		// If it looks like a wordpress metadata URI
	} else if strings.Contains(revisionURI, "/wp-json/") {
		metadataScraper := &scraper.CivilMetadataScraper{}
		civilMetadata, err := metadataScraper.ScrapeCivilMetadata(revisionURI)
		if err != nil {
			return nil, nil, err
		}
		// TODO(PN): Hack to fix bad URLs received for metadata
		// Remove this later after testing
		if civilMetadata.Title() == "" && civilMetadata.RevisionContentHash() == "" {
			revisionURI = strings.Replace(revisionURI, "/wp-json", "/crawler-pod/wp-json", -1)
			civilMetadata, err = metadataScraper.ScrapeCivilMetadata(revisionURI)
			if err != nil {
				return nil, nil, err
			}
		}
		return civilMetadata, nil, nil
	}

	return nil, nil, nil
}

// TODO(PN): This isn't great, rework is needed later.
func (n *NewsroomEventProcessor) scraperDataToPayload(metadata *model.ScraperCivilMetadata,
	content *model.ScraperContent) model.ArticlePayload {
	payload := model.ArticlePayload{}
	if metadata != nil {
		payload["title"] = metadata.Title()
		payload["revisionContentHash"] = metadata.RevisionContentHash()
		payload["revisionContentURL"] = metadata.RevisionContentURL()
		payload["canonicalURL"] = metadata.CanonicalURL()
		payload["slug"] = metadata.Slug()
		payload["description"] = metadata.Description()
		payload["primaryTag"] = metadata.PrimaryTag()
		payload["revisionDate"] = metadata.RevisionDate()
		payload["originalPublishDate"] = metadata.OriginalPublishDate()
		payload["opinion"] = metadata.Opinion()
		payload["schemaVersion"] = metadata.SchemaVersion()
		payload["authors"] = n.buildContributors(metadata)
		payload["images"] = n.buildImages(metadata)
	}
	if content != nil {
		payload["contentAuthor"] = content.Author()
		payload["contentText"] = content.Text()
		payload["contentURI"] = content.URI()
		payload["contentHTML"] = content.HTML()
		payload["contentData"] = content.Data()
	}
	return payload
}

func (n *NewsroomEventProcessor) buildContributors(metadata *model.ScraperCivilMetadata) []map[string]interface{} {
	contributors := []map[string]interface{}{}
	for _, contributor := range metadata.Contributors() {
		entry := map[string]interface{}{
			"role":      contributor.Role(),
			"name":      contributor.Name(),
			"address":   contributor.Address(),
			"signature": contributor.Signature(),
		}
		contributors = append(contributors, entry)
	}
	return contributors
}

func (n *NewsroomEventProcessor) buildImages(metadata *model.ScraperCivilMetadata) []map[string]interface{} {
	images := []map[string]interface{}{}
	for _, image := range metadata.Images() {
		entry := map[string]interface{}{
			"url":  image.URL(),
			"hash": image.Hash(),
			"h":    image.Height(),
			"w":    image.Width(),
		}
		images = append(images, entry)
	}
	return images
}
