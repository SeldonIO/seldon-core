/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

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
