/*
Package uploadcare provides a client for using the Uploadcare API.

Usage:

	import (
		"github.com/uploadcare/uploadcare-go/uploadcare"
		"github.com/uploadcare/uploadcare-go/file"
	)

Construct a new Uploadcare client, then use the various services to
access different parts of the Uploadcare API. For example:

	creds := uploadcare.APICreds{
		SecretKey: "your_secret_key",
		PublicKey: "your_public_key",
	}

	client, err := uploadcare.NewClient(creds)
	if err != nil {
		// handle error
	}

	list, err := file.NewService(client).ListOfFiles()

Authentication

To authenticate your account, every request made MUST be signed.
There are two available auth functions for that:
- SimpleAuth
- SignBasedAuth

NOTE: by default SimpleAuth is used, if you want to use SignBasedAuth
you need to enable it in the Uploadcare dashboard first.
*/
package uploadcare
