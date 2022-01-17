package rclone

import (
	"github.com/seldonio/seldon-core/scheduler/pkg/agent/config"
	"github.com/seldonio/seldon-core/scheduler/pkg/agent/k8s"
)

func (r *RCloneClient) loadRcloneConfiguration(config *config.AgentConfiguration) error {
	logger := r.logger.WithField("func", "loadRcloneConfiguration")
	var rcloneNamesAdded []string
	var err error
	if config != nil {
		// Load any secrets that have Rclone config
		rcloneNamesAdded, err = r.loadRcloneSecretsConfiguration(config)
		if err != nil {
			return err
		}
		// Load any raw Rclone configs
		rcloneNamesAddedSecrets, err := r.loadRcloneRawConfiguration(config)
		if err != nil {
			return err
		}
		rcloneNamesAdded = append(rcloneNamesAdded, rcloneNamesAddedSecrets...)

		// Delete any existing remotes not in defaults
		err = r.deleteUnusedRcloneConfiguration(config, rcloneNamesAdded)
		if err != nil {
			logger.WithError(err).Errorf("Failed to delete unused Rclone configuration")
		}
		existingRemotes, err := r.ListRemotes()
		if err != nil {
			return err
		}
		logger.Infof("After update current set of remotes is %v", existingRemotes)
	} else {
		logger.Warn("nil config passed")
	}
	return nil
}

func (r *RCloneClient) loadRcloneRawConfiguration(config *config.AgentConfiguration) ([]string, error) {
	logger := r.logger.WithField("func", "loadRcloneRawConfiguration")
	var rcloneNamesAdded []string
	if len(config.Rclone.Config) > 0 {
		for _, config := range config.Rclone.Config {
			logger.Infof("Loading rclone config %s", config)
			name, err := r.Config([]byte(config))
			if err != nil {
				return nil, err
			}
			rcloneNamesAdded = append(rcloneNamesAdded, name)
		}
	}
	return rcloneNamesAdded, nil
}

func (r *RCloneClient) deleteUnusedRcloneConfiguration(config *config.AgentConfiguration, rcloneNamesAdded []string) error {
	logger := r.logger.WithField("func", "deleteUnusedRcloneConfiguration")
	existingRemotes, err := r.ListRemotes()
	if err != nil {
		return err
	}
	for _, existingName := range existingRemotes {
		found := false
		for _, addedName := range rcloneNamesAdded {
			if existingName == addedName {
				found = true
				break
			}
		}
		if !found {
			logger.Infof("Delete remote %s as not in new list of defaults", existingName)
			err := r.DeleteRemote(existingName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *RCloneClient) loadRcloneSecretsConfiguration(config *config.AgentConfiguration) ([]string, error) {
	logger := r.logger.WithField("func", "loadRcloneSecretsConfiguration")
	var rcloneNamesAdded []string
	// Load any secrets that have Rclone config
	if len(config.Rclone.ConfigSecrets) > 0 {
		secretClientSet, err := k8s.CreateClientset()
		if err != nil {
			return nil, err
		}
		secretsHandler := k8s.NewSecretsHandler(secretClientSet, r.namespace)
		for _, secret := range config.Rclone.ConfigSecrets {
			logger.Infof("Loading rclone secret %s", secret)
			config, err := secretsHandler.GetSecretConfig(secret)
			if err != nil {
				return nil, err
			}
			name, err := r.Config(config)
			if err != nil {
				return nil, err
			}
			rcloneNamesAdded = append(rcloneNamesAdded, name)
		}
	}
	return rcloneNamesAdded, nil
}
