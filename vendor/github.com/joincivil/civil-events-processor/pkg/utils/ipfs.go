package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	chttp "github.com/joincivil/go-common/pkg/http"
)

const (
	ipfsNodeURI = "https://ipfs.infura.io"
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
	client := chttp.NewRestHelper(ipfsNodeURI, "")
	targetURI := fmt.Sprintf("/ipfs/%v", addr)

	return client.SendRequest(targetURI, http.MethodGet, nil, nil)
}
