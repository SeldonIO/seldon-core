/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package hodometer

var (
	BuildVersion string = "0.0.0"
	BuildTime    string
	GitBranch    string
	GitCommit    string
	ReleaseType  string
)

func GetBuildDetails() map[string]interface{} {
	return map[string]interface{}{
		"BuildVersion": BuildVersion,
		"BuildTime":    BuildTime,
		"GitBranch":    GitBranch,
		"GitCommit":    GitCommit,
		"ReleaseType":  ReleaseType,
	}
}
