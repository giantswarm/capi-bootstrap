package util

import (
	"io"
	"net/http"
	"os"

	"github.com/giantswarm/microerror"
)

func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return microerror.Mask(err)
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return microerror.Mask(err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return microerror.Mask(err)
}

func Contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

func ProviderShort(provider string) string {
	switch provider {
	case "aws":
		return "CAPA"
	case "azure":
		return "CAPZ"
	case "gcp":
		return "CAPG"
	case "openstack":
		return "CAPO"
	case "vsphere":
		return "CAPV"
	default:
		return ""
	}
}
