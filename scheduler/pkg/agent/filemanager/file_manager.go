package filemanager

import "context"

//go:generate go tool mockgen -source=file_manager.go -destination=./mocks/mock_file_manager.go -package=mocks FileManager
type FileManager interface {
	StartConfigListener() error
	Ready() error
	Config(config []byte) (string, error)
	Copy(ctx context.Context, modelName string, srcUri string, config []byte) (string, error)
	PurgeLocal(path string) error
	ListRemotes() ([]string, error)
	DeleteRemote(name string) error
}
