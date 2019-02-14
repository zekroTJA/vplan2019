package drivers

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/zekroTJA/vplan2019/internal/auth"
	"github.com/zekroTJA/vplan2019/internal/config"
)

// DebugAuthProvider is an auth provider, which
// is only purposed to use in debugging and testing
type DebugAuthProvider struct {
	cfg   config.Model
	creds map[string]string
}

// Connect _
func (d *DebugAuthProvider) Connect(options config.Model) error {
	d.cfg = options
	d.creds = map[string]string{
		"test": "passwd",
	}
	return nil
}

// Close _
func (d *DebugAuthProvider) Close() {}

// GetConfigModel _
func (d *DebugAuthProvider) GetConfigModel() config.Model {
	return make(map[string]string)
}

// Authenticate _
func (d *DebugAuthProvider) Authenticate(username, password string) (*auth.Response, error) {
	if pw, ok := d.creds[username]; ok && pw == password {
		ident := fmt.Sprintf("%x", sha256.Sum256([]byte(username+password)))
		return &auth.Response{
			Ident: ident,
		}, nil
	}
	return nil, errors.New("unauthorized")
}