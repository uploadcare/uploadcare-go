// Package test contains an integration test for all of the uploadcare-go lib
package test

import (
	"testing"

	"github.com/segmentio/go-env"
	"github.com/uploadcare/uploadcare-go/test/testenv"
	"github.com/uploadcare/uploadcare-go/ucare"
)

var integrationTests = []struct {
	name string
	fn   func(t *testing.T, r *testenv.Runner)
}{
	{"upload file", uploadFile},
	{"upload file from url", uploadFromURL},
	{"upload file info", uploadFileInfo},
	{"create group from uploaded files", uploadCreateGroup},
	{"get uploaded group info", uploadGroupInfo},
	{"upload file through multipart upload API", uploadMultipart},
	{"list file groups", groupList},
	{"store file group", groupStore},
	{"file group info", groupInfo},
	{"convert document", conversionDocument},
	{"document conversion status", conversionDocumentStatus},
	{"list files", listFiles},
	{"get file info", fileInfo},
	{"store file", fileStore},
	{"batch file store", fileBatchStore},
	{"local file copy", fileLocalCopy},
	{"remote file copy", fileRemoteCopy},
	{"delete file", fileDelete},
	{"batch file delete", fileBatchDelete},
	{"create webhook", webhookCreate},
	{"update webhook", webhookUpdate},
	{"list webhooks", webhookList},
	{"delete webhook", webhookDelete},
	{"project info", projectInfo},
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	creds := ucare.APICreds{
		SecretKey: env.MustGet("SECRET_KEY"),
		PublicKey: env.MustGet("PUBLIC_KEY"),
	}

	// TODO: test with different config settings
	conf := ucare.Config{
		SignBasedAuthentication: true,
		APIVersion:              ucare.APIv06,
	}

	client, err := ucare.NewClient(creds, &conf)
	if err != nil {
		t.Fatal(err)
	}

	customStorage := env.MustGet("CUSTOM_STORAGE_BUCKET")

	r := testenv.NewRunner(client, customStorage)

	// The ordering here is important as each test adds state to artifacts
	for _, test := range integrationTests {
		t.Run(test.name, func(t *testing.T) { test.fn(t, r) })
	}
}
