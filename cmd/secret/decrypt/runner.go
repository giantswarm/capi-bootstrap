package encrypt

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/spf13/cobra"
	"go.mozilla.org/sops/v3"
	"go.mozilla.org/sops/v3/aes"
	"go.mozilla.org/sops/v3/cmd/sops/common"
	"go.mozilla.org/sops/v3/cmd/sops/formats"
	"go.mozilla.org/sops/v3/keyservice"
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

func ensureMetadata(branch sops.TreeBranch) error {
	for _, b := range branch {
		if b.Key == "sops" {
			return nil
		}
	}
	return microerror.Maskf(invalidConfigError, "input file not encrypted")
}

func (r *Runner) Do(ctx context.Context, _ *cobra.Command, _ []string) error {
	inputStore := common.StoreForFormat(formats.Yaml)
	outputStore := common.StoreForFormat(formats.Yaml)

	/*
		fileBytes, err := ioutil.ReadFile(r.flag.InputFile)
		if err != nil {
			return microerror.Mask(err)
		}

		branches, err := inputStore.LoadPlainFile(fileBytes)
		if err != nil {
			return microerror.Mask(err)
		}

		if err := ensureMetadata(branches[0]); err != nil {
			return microerror.Mask(err)
		}

		path, err := filepath.Abs(r.flag.InputFile)
		if err != nil {
			return microerror.Mask(err)
		}

		tree := sops.Tree{
			Branches: branches,
			Metadata: sops.Metadata{
				EncryptedRegex:  "^(data|stringData)$",
				Version:         project.Version(),
				ShamirThreshold: 0,
			},
			FilePath: path,
		}
	*/

	var keyServices []keyservice.KeyServiceClient
	keyServices = append(keyServices, keyservice.NewLocalClient())

	cipher := aes.NewCipher()
	tree, err := common.LoadEncryptedFileWithBugFixes(common.GenericDecryptOpts{
		Cipher:      cipher,
		InputStore:  inputStore,
		InputPath:   r.flag.InputFile,
		KeyServices: keyServices,
	})
	if err != nil {
		return err
	}

	_, err = common.DecryptTree(common.DecryptTreeOpts{
		KeyServices: keyServices,
		Tree:        tree,
		Cipher:      cipher,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	decryptedFile, err := outputStore.EmitPlainFile(tree.Branches)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = r.stdout.Write(decryptedFile)
	return microerror.Mask(err)
}
