/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package rclone

import (
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/config"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/k8s"
)

func (r *RCloneClient) loadRcloneConfiguration(config *config.AgentConfiguration) error {
	logger := r.logger.WithField("func", "loadRcloneConfiguration")

	if config == nil {
		logger.Warn("nil config passed")
		return nil
	}

	// Load any secrets that have Rclone config
	addedFromSecrets, err := r.loadRcloneSecretsConfiguration(config)
	if err != nil {
		return err
	}

	// Load any raw Rclone configs
	addedFromRawConfig, err := r.loadRcloneRawConfiguration(config)
	if err != nil {
		return err
	}

	addedFromSecrets = append(addedFromSecrets, addedFromRawConfig...)

	// Delete any existing remotes not in defaults
	err = r.deleteUnusedRcloneConfiguration(config, addedFromSecrets)
	if err != nil {
		logger.WithError(err).Errorf("Failed to delete unused Rclone configuration")
	}

	existingRemotes, err := r.ListRemotes()
	if err != nil {
		return err
	}

	logger.Infof("After update current set of remotes is %v", existingRemotes)

	return nil
}

func (r *RCloneClient) loadRcloneRawConfiguration(config *config.AgentConfiguration) ([]string, error) {
	logger := r.logger.WithField("func", "loadRcloneRawConfiguration")

	var rcloneNamesAdded []string
	if len(config.Rclone.Config) > 0 {
		logger.Infof("found %d Rclone configs", len(config.Rclone.Config))

		for _, config := range config.Rclone.Config {
			logger.Info("loading Rclone config")

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
			logger.WithField("secret_name", secret).Infof("retrieving Rclone secret")

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
