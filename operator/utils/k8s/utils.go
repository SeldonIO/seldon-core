package k8s

import (
	"fmt"
	"github.com/go-logr/logr"
	"k8s.io/client-go/discovery"
	"strconv"
	"strings"
)

// This will cause CRD V1 to be installed
const DefaultMinorVersion = 18

func GetServerVersion(client discovery.DiscoveryInterface, logger logr.Logger) (string, error) {
	serverVersion, err := client.ServerVersion()
	if err != nil {
		logger.Error(err, "Failed to get server version")
		return "", err
	}
	logger.Info("Server version", "Major", serverVersion.Major, "Minor", serverVersion.Minor)
	majorVersion, err := strconv.Atoi(serverVersion.Major)
	if err != nil {
		logger.Error(err, "Failed to parse majorVersion defaulting to 1")
		majorVersion = 1
	}
	minorVersion, err := strconv.Atoi(serverVersion.Minor)
	if err != nil {
		if strings.HasSuffix(serverVersion.Minor, "+") {
			minorVersion, err = strconv.Atoi(serverVersion.Minor[0 : len(serverVersion.Minor)-1])
			if err != nil {
				logger.Error(err, "Failed to parse minorVersion defaulting to 12")
				minorVersion = DefaultMinorVersion
			}
		} else {
			logger.Error(err, "Failed to parse minorVersion defaulting to 12")
			minorVersion = DefaultMinorVersion
		}
	}
	return fmt.Sprintf("%d.%d", majorVersion, minorVersion), nil
}
