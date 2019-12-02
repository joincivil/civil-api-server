package testruntime

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/joincivil/go-common/pkg/newsroom"
	"io"
	"io/ioutil"
	"strings"
)

// NewMockIPFS returns a new IPFSHelper
func NewMockIPFS() newsroom.IPFSHelper {
	data := make(map[string]string)
	return &MockIPFS{data}
}

// MockIPFS stores files in memory and doesn't actually use IPFS
type MockIPFS struct {
	data map[string]string
}

// Cat returns a file by path
func (m MockIPFS) Cat(path string) (io.ReadCloser, error) {

	if data, ok := m.data[path]; ok {
		reader := ioutil.NopCloser(strings.NewReader(data))
		return reader, nil
	}
	return ioutil.NopCloser(strings.NewReader("")), fmt.Errorf("not found")
}

// Add puts a new file into the map
func (m MockIPFS) Add(r io.Reader, options ...shell.AddOpts) (string, error) {
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	hash := common.Bytes2Hex(crypto.Keccak256(buf.Bytes()))

	m.data[hash] = buf.String()

	return hash, nil
}
