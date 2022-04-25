package encrypt

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
	"go.mozilla.org/sops/v3"
	"go.mozilla.org/sops/v3/aes"
	"go.mozilla.org/sops/v3/age"
	"go.mozilla.org/sops/v3/cmd/sops/common"
	"go.mozilla.org/sops/v3/cmd/sops/formats"
	"go.mozilla.org/sops/v3/keys"
	"go.mozilla.org/sops/v3/keyservice"

	"github.com/giantswarm/capi-bootstrap/pkg/project"
)

func (r *Runner) Run(cmd *cobra.Command, args []string) error {
	err := r.flag.Validate()
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.Do(cmd.Context(), cmd, args)
	return microerror.Mask(err)
}

func ensureNoMetadata(branch sops.TreeBranch) error {
	for _, b := range branch {
		if b.Key == "sops" {
			return microerror.Maskf(invalidConfigError, "input file already encrypted")
		}
	}
	return nil
}

func (r *Runner) Do(ctx context.Context, _ *cobra.Command, _ []string) error {
	var keyGroups []sops.KeyGroup
	{
		var masterKeys []keys.MasterKey
		ageKeys, err := age.MasterKeysFromRecipients(r.flag.PublicKey)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, ageKey := range ageKeys {
			masterKeys = append(masterKeys, ageKey)
		}

		keyGroups = append(keyGroups, masterKeys)
	}

	inputStore := common.StoreForFormat(formats.Yaml)
	outputStore := common.StoreForFormat(formats.Yaml)

	fileBytes, err := ioutil.ReadFile(r.flag.InputFile)
	if err != nil {
		return microerror.Mask(err)
	}

	branches, err := inputStore.LoadPlainFile(fileBytes)
	if err != nil {
		return microerror.Mask(err)
	}

	if err := ensureNoMetadata(branches[0]); err != nil {
		return microerror.Mask(err)
	}

	path, err := filepath.Abs(r.flag.InputFile)
	if err != nil {
		return microerror.Mask(err)
	}

	tree := sops.Tree{
		Branches: branches,
		Metadata: sops.Metadata{
			KeyGroups:       keyGroups,
			EncryptedRegex:  "^(data|stringData)$",
			Version:         project.Version(),
			ShamirThreshold: 0,
		},
		FilePath: path,
	}
	var keyServices []keyservice.KeyServiceClient
	keyServices = append(keyServices, keyservice.NewLocalClient())
	dataKey, errs := tree.GenerateDataKeyWithKeyServices(keyServices)
	if len(errs) > 0 {
		return microerror.Mask(fmt.Errorf("could not generate data key: %s", errs))
	}

	err = common.EncryptTree(common.EncryptTreeOpts{
		DataKey: dataKey,
		Tree:    &tree,
		Cipher:  aes.NewCipher(),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	encryptedFile, err := outputStore.EmitEncryptedFile(tree)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = r.stdout.Write(encryptedFile)
	return microerror.Mask(err)
}
