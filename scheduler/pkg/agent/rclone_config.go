package agent

import "github.com/seldonio/seldon-core/scheduler/pkg/agent/k8s"

func (c *Client) loadRcloneConfiguration(config *AgentConfiguration) error {
	logger := c.logger.WithField("func", "loadRcloneDefaults")
	var rcloneNamesAdded []string
	var err error
	if config != nil {
		// Load any secrets that have Rclone config
		rcloneNamesAdded, err = c.loadRcloneSecretsConfiguration(config)
		if err != nil {
			return err
		}
		// Load any raw Rclone configs
		rcloneNamesAddedSecrets, err := c.loadRcloneRawConfiguration(config)
		if err != nil {
			return err
		}
		rcloneNamesAdded = append(rcloneNamesAdded, rcloneNamesAddedSecrets...)

		// Delete any existing remotes not in defaults
		err = c.deleteUnusedRcloneConfiguration(config, rcloneNamesAdded)
		if err != nil {
			logger.WithError(err).Errorf("Failed to delete unused Rclone configuration")
		}
		existingRemotes, err := c.RCloneClient.ListRemotes()
		if err != nil {
			return err
		}
		logger.Infof("After update current set of remotes is %v", existingRemotes)
	}
	return nil
}

func (c *Client) loadRcloneRawConfiguration(config *AgentConfiguration) ([]string, error) {
	logger := c.logger.WithField("func", "loadRcloneRawConfiguration")
	var rcloneNamesAdded []string
	if len(config.Rclone.Config) > 0 {
		for _, config := range config.Rclone.Config {
			logger.Infof("Loading rclone config %s", config)
			name, err := c.RCloneClient.Config([]byte(config))
			if err != nil {
				return nil, err
			}
			rcloneNamesAdded = append(rcloneNamesAdded, name)
		}
	}
	return rcloneNamesAdded, nil
}

func (c *Client) deleteUnusedRcloneConfiguration(config *AgentConfiguration, rcloneNamesAdded []string) error {
	logger := c.logger.WithField("func", "deleteUnsedRcloneConfiguration")
	existingRemotes, err := c.RCloneClient.ListRemotes()
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
			err := c.RCloneClient.DeleteRemote(existingName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) loadRcloneSecretsConfiguration(config *AgentConfiguration) ([]string, error) {
	logger := c.logger.WithField("func", "loadRcloneSecretsConfiguration")
	var rcloneNamesAdded []string
	// Load any secrets that have Rclone config
	if len(config.Rclone.ConfigSecrets) > 0 {
		secretClientSet, err := k8s.CreateClientset()
		if err != nil {
			return nil, err
		}
		secretsHandler := k8s.NewSecretsHandler(secretClientSet, c.namespace)
		for _, secret := range config.Rclone.ConfigSecrets {
			logger.Infof("Loading rclone secret %s", secret)
			config, err := secretsHandler.GetSecretConfig(secret)
			if err != nil {
				return nil, err
			}
			name, err := c.RCloneClient.Config(config)
			if err != nil {
				return nil, err
			}
			rcloneNamesAdded = append(rcloneNamesAdded, name)
		}
	}
	return rcloneNamesAdded, nil
}
