package tls

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	serverName := "test-server-name"
	errCert := ErrCertificateNotFound{ServerName: serverName}
	errMsg := fmt.Sprintf(`Certificate for "%s" not found`, serverName)
	assert.Equal(t, errCert.Error(), errMsg)
}

func TestGetWildCard(t *testing.T) {
	domain := "ironpack.vmprovider.hackerrank.link"
	wildCard := getWildCard(domain)
	assert.Equal(t, wildCard, "*.vmprovider.hackerrank.link")
}
