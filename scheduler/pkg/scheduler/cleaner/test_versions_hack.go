/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package cleaner

import (
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

type TestVersionCleaner struct {
	*VersionCleaner
}

func NewTestVersionCleaner(schedStore store.ModelServerAPI, logger log.FieldLogger) *TestVersionCleaner {
	return &TestVersionCleaner{
		VersionCleaner: NewVersionCleaner(schedStore, logger),
	}
}

func (v *TestVersionCleaner) CleanupOldVersions(modelName string) error {
	return v.cleanupOldVersions(modelName)
}
