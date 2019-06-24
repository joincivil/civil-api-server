package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	chttp "github.com/joincivil/go-common/pkg/http"
)

const (
	ipfsNodeURI = "https://ipfs.infura.io"
	timeout     = 3 * time.Second
)

// RetrieveIPFSLink retrieves data from a given IPFS link via the given IPFS
// node
func RetrieveIPFSLink(uri string) ([]byte, error) {
	if !strings.HasPrefix(uri, "ipfs://") {
		return nil, fmt.Errorf("Invalid IPFS link: %v", uri)
	}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	addr := u.Host
	client := chttp.NewRestHelperWithTimeout(ipfsNodeURI, "", timeout)
	targetURI := fmt.Sprintf("ipfs/%v", addr)

	maxAtts := 3
	baseWaitMs := 500
	return client.SendRequestWithRetry(
		targetURI,
		http.MethodGet,
		nil,
		nil,
		maxAtts,
		baseWaitMs,
	)
}
