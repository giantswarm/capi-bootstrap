package bootstrap

import "context"

func (b *Bootstrapper) Apply(ctx context.Context) error {
	defer b.cleanup()
	return b.apply(ctx)
}

func (b *Bootstrapper) apply(ctx context.Context) error {
	// 0. setup helm
	err := b.configureHelmCatalogRepo()
	if err != nil {
		return err
	}

	// 1. create/find kind cluster and get kubeconfig/client
	err = b.ensureBootstrapCluster(ctx)
	if err != nil {
		return err
	}

	// 2. setup app platform and capi on bootstrap cluster
	err = b.setupMC(ctx, false)
	if err != nil {
		return err
	}

	// 3. create permanent cluster as apps in bootstrap cluster and wait for ready
	err = b.createCluster(ctx)
	if err != nil {
		return err
	}

	// 4. setup app platform and capi on permanent cluster
	err = b.setupMC(ctx, true)
	if err != nil {
		return err
	}

	// 5. move cluster resources from bootstrap into permanent cluster
	err = b.moveCluster(ctx, true)
	if err != nil {
		return err
	}

	return nil
}
