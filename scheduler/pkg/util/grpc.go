/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package util

import "google.golang.org/grpc/keepalive"

func GetClientKeepAliveParameters() keepalive.ClientParameters {
	return keepalive.ClientParameters{
		Time:                gRPCKeepAliveTime,
		Timeout:             clientKeepAliveTimeout,
		PermitWithoutStream: gRPCKeepAlivePermit,
	}
}

func GetServerKeepAliveEnforcementPolicy() keepalive.EnforcementPolicy {
	return keepalive.EnforcementPolicy{
		MinTime:             gRPCKeepAliveTime,
		PermitWithoutStream: gRPCKeepAlivePermit,
	}
}
