package gumble

import (
	"crypto/x509"
	"net"
	"time"
)

// UserStats contains additional information about a user.
type UserStats struct {
	// The owner of the stats.
	User *User

	// The user's version.
	Version Version
	// When the user connected to the server.
	Connected time.Time
	// How long the user has been idle.
	Idle time.Duration
	// How much bandwidth the user is current using.
	Bandwidth int
	// The user's certificate chain.
	Certificates []*x509.Certificate
	// Does the user's client supports the Opus audio codec?
	Opus bool

	// The user's IP address.
	IP net.IP
}
