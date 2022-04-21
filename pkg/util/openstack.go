package util

import "strings"

func OpenrcToCloudConfig(content string) CloudConfig {
	cloud := CloudConfigCloud{
		Verify:             false,
		Interface:          "public",
		IdentityAPIVersion: 3,
	}

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "export ") {
			continue
		}

		trimmed = strings.TrimPrefix(trimmed, "export ")
		split := strings.SplitN(trimmed, "=", 2)
		if len(split) != 2 {
			continue
		}

		key := split[0]
		value := strings.Trim(split[1], "\"")

		switch key {
		case "OS_AUTH_URL":
			cloud.Auth.AuthURL = value
		case "OS_PROJECT_ID":
			cloud.Auth.ProjectID = value
		case "OS_USER_DOMAIN_NAME":
			cloud.Auth.UserDomainName = value
		case "OS_USERNAME":
			cloud.Auth.Username = value
		case "OS_PASSWORD":
			cloud.Auth.Password = value
		case "OS_REGION_NAME":
			cloud.RegionName = value
		}
	}

	return CloudConfig{
		Clouds: map[string]CloudConfigCloud{
			"openstack": cloud,
		},
	}
}
