/*
Package ucare provides a client for using the Uploadcare API.

Usage:

	import (
		"github.com/uploadcare/uploadcare-go/ucare"
		"github.com/uploadcare/uploadcare-go/file"
	)

Construct a new Uploadcare client, then use the various services to
access different parts of the Uploadcare API. For example:

	creds := ucare.APICreds{
		SecretKey: "your_secret_key",
		PublicKey: "your_public_key",
	}

	client, err := ucare.NewClient(creds)
	if err != nil {
		// handle error
	}

	list, err := file.NewService(client).ListFiles()
	...

Authentication

To authenticate your account, every request made MUST be signed.
There are two available auth functions for that:
	SimpleAuth (default)
	SignBasedAuth

NOTE: If you want to use SignBasedAuth you need to enable it in the Uploadcare
dashboard first.
*/
package ucare
