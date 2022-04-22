package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"

	application "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c Client) WaitForClusterReady(ctx context.Context, clusterNamespace, clusterName string) error {
	var appNames []string
	for _, appName := range []string{
		"app-operator",
		"chart-operator",
		"cert-exporter",
		"cilium",
		"cloud-provider-openstack",
		"kube-state-metrics",
		"metrics-server",
		"net-exporter",
		"node-exporter",
	} {
		appNames = append(appNames, fmt.Sprintf("%s-%s", clusterName, appName))
	}

	err := c.WaitForAppsDeployed(ctx, clusterNamespace, appNames)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (c Client) WaitForClusterDeleted(ctx context.Context, clusterNamespace, clusterName string) error {
	for {
		var clusterApps application.AppList
		err := c.ctrlClient.List(ctx, &clusterApps, client.InNamespace(clusterNamespace), client.MatchingLabels{
			"giantswarm.io/cluster": clusterName,
		})
		if err != nil {
			return microerror.Mask(err)
		}

		if len(clusterApps.Items) == 0 {
			c.logger.Debugf(ctx, "no apps fond")
			break
		}

		appsByName := map[string]application.App{}
		for _, app := range clusterApps.Items {
			appName := strings.TrimPrefix(app.ObjectMeta.Name, fmt.Sprintf("%s-", clusterName))
			if appName != "" {
				appsByName[appName] = app
			}
		}

		c.logger.Debugf(ctx, "found %d cluster apps", len(appsByName))

		allAppsDeleted := true
		for _, appName := range []string{
			"app-operator",
			"chart-operator",
			"cert-exporter",
			"cilium",
			"cloud-provider-openstack",
			"kube-state-metrics",
			"metrics-server",
			"net-exporter",
			"node-exporter",
		} {
			if _, ok := appsByName[appName]; ok {
				c.logger.Debugf(ctx, "waiting for app %s to be deleted", appName)
				allAppsDeleted = false
				break
			}
		}

		if allAppsDeleted {
			c.logger.Debugf(ctx, "all apps deleted")
			break
		}

		time.Sleep(time.Second * 10)
	}

	c.logger.Debugf(ctx, "waiting for cluster to be deleted")

	for {
		// using PartialObjectMetadata so we don't have to import types from external CAPI module
		cluster := meta.PartialObjectMetadata{
			TypeMeta: meta.TypeMeta{
				Kind:       "Cluster",
				APIVersion: "cluster.x-k8s.io/v1beta1",
			},
		}
		err := c.ctrlClient.Get(ctx, client.ObjectKey{
			Name:      clusterName,
			Namespace: clusterNamespace,
		}, &cluster)
		if apierrors.IsNotFound(err) {
			c.logger.Debugf(ctx, "cluster deleted")
			break
		} else if err != nil {
			return microerror.Mask(err)
		}

		time.Sleep(time.Second * 10)
	}

	return nil
}
