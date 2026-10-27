//go:build reviewisolated

package node

// Only for the explicit-file review test command. Certificate code is not
// exercised by these HTTP/identity tests and never contacts external issuers.
func (c *Controller) requestCert() error   { panic("unexpected certificate request in identity test") }
func (c *Controller) renewCertTask() error { panic("unexpected certificate renewal in identity test") }
