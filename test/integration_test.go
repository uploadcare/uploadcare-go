// Package test contains an integration test for all of the uploadcare-go lib
package test

import (
	"context"
	"os"
	"testing"

	"github.com/uploadcare/uploadcare-go/v2/projectapi"
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
	conf := ucare.Config{
		SignBasedAuthentication: true,
		APIVersion:             ucare.APIv07,
	}

	client, err := ucare.NewClient(creds, &conf)
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

func TestProjectAPIIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	token := os.Getenv("PROJECT_API_TOKEN")
	if token == "" {
		t.Skip("PROJECT_API_TOKEN env var is required for Project API tests")
	}

	client, err := ucare.NewBearerClient(token, &ucare.Config{
		Retry: &ucare.RetryConfig{MaxRetries: 3},
	})
	if err != nil {
		t.Fatal(err)
	}

	svc := projectapi.NewService(client)

	// List projects to get a pub_key for subsequent tests
	t.Run("list projects", func(t *testing.T) { projectAPIList(t, svc) })

	// Get the first project's pub_key for further tests
	list, err := svc.List(context.Background(), nil)
	if err != nil {
		t.Fatal("cannot list projects: ", err)
	}
	if !list.Next() {
		t.Fatal("no projects found")
	}
	first, err := list.ReadResult()
	if err != nil {
		t.Fatal("cannot read first project: ", err)
	}
	pubKey := first.PubKey

	t.Run("get project", func(t *testing.T) { projectAPIGet(t, svc, pubKey) })
	t.Run("list secrets", func(t *testing.T) { projectAPIListSecrets(t, svc, pubKey) })
}
