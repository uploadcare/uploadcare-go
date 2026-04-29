// Package test contains an integration test for all of the uploadcare-go lib
package test

import (
	"os"
	"testing"

	"github.com/uploadcare/uploadcare-go/v2/test/testenv"
	"github.com/uploadcare/uploadcare-go/v2/ucare"
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
	{"file group info", groupInfo},
	{"delete file group", groupDelete},
	{"convert document", conversionDocument},
	{"document conversion status", conversionDocumentStatus},
	{"list files", listFiles},
	{"get file info", fileInfo},
	{"store file", fileStore},
	{"set file metadata", metadataSet},
	{"get file metadata", metadataGet},
	{"list file metadata", metadataList},
	{"delete file metadata", metadataDelete},
	{"execute clamav addon", addonClamAVExecute},
	{"check clamav addon status", addonClamAVStatus},
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

	secretKey := os.Getenv("SECRET_KEY")
	publicKey := os.Getenv("PUBLIC_KEY")
	if secretKey == "" || publicKey == "" {
		t.Fatal("SECRET_KEY and PUBLIC_KEY env vars are required")
	}

	creds := ucare.APICreds{
		SecretKey: secretKey,
		PublicKey: publicKey,
	}

	// TODO: test with different config settings
	conf, err := ucare.NewConfig(creds,
		ucare.WithSignBasedAuthentication(),
		ucare.WithAPIVersion(ucare.APIv07),
	)
	if err != nil {
		t.Fatal(err)
	}

	client, err := ucare.NewClient(creds, conf)
	if err != nil {
		t.Fatal(err)
	}

	customStorage := os.Getenv("CUSTOM_STORAGE_BUCKET")

	r := testenv.NewRunner(client, customStorage)

	// The ordering here is important as each test adds state to artifacts
	for _, test := range integrationTests {
		t.Run(test.name, func(t *testing.T) { test.fn(t, r) })
	}
}
