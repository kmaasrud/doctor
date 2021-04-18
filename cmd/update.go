package cmd

import (
    "errors"

    "github.com/kmaasrud/doctor/msg"
    "github.com/equinox-io/equinox"
)

// assigned when creating a new application in the dashboard
const appID = "app_gvefXKeSXD5"

// public portion of signing key generated by `equinox genkey`
var publicKey = []byte(`
-----BEGIN ECDSA PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEBIokTYcFzVBGV68Vs+32HiIFdIyRfUeZ
ggZtn72eXWLSzARQCtDtC05lAWu/7DZj1kpkC5aX1iiZ0Luw4135nHNXGcTch0/f
EnlrZMZSJhNdxu2/9VhgG/UEISHrp0iX
-----END ECDSA PUBLIC KEY-----
`)

func Update() error {
	done := make(chan struct{})
	go msg.Do("Looking for new version...", done)
	var opts equinox.Options
	if err := opts.SetPublicKeyPEM(publicKey); err != nil {
		msg.CloseDo(done)
		return errors.New("Could not set public key. " + err.Error())
	}

	// check for the update
	resp, err := equinox.Check(appID, opts)
	msg.CloseDo(done)
	switch {
	case err == equinox.NotAvailableErr:
		msg.Info("No update available, already at the latest version!")
		return nil
	case err != nil:
		return errors.New("Update failed: " + err.Error())
	}

	// fetch the update and apply it
	done = make(chan struct{})
	go msg.Do("Found update! Applying it...", done)
	err = resp.Apply()
	msg.CloseDo(done)
	if err != nil {
		return errors.New("Could not apply update. " + err.Error())
	}

	msg.Success("Updated to new version: %s" + resp.ReleaseVersion + "!")
	return nil
}
