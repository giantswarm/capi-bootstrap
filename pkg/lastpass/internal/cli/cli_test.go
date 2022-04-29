package cli

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/capi-bootstrap/pkg/key"
)

func Test_Create(t *testing.T) {
	var cli Client
	created, err := cli.Create(context.Background(), key.EncryptionKeySecretShare, "Encryption Keys", "Thomas Test", "notes")
	if err != nil {
		t.Fatal(err)
	}

	retrieved, err := cli.Get(context.Background(), key.EncryptionKeySecretShare, "Encryption Keys", "Thomas Test")
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(created, retrieved); diff != "" {
		t.Fatal(diff)
	}

	err = cli.Delete(context.Background(), retrieved.ID)
	if err != nil {
		t.Fatal(err)
	}
}
