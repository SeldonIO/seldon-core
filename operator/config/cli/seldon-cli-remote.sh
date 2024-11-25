#!/bin/sh

#
# Copyright (c) 2024 Seldon Technologies Ltd.

# Use of this software is governed by
# (1) the license included in the LICENSE file or
# (2) if the license included in the LICENSE file is the Business Source License 1.1,
# the Change License after the Change Date as each is defined in accordance with the LICENSE file.
#


# This script is used to run commands on the seldon-cli pod in the seldon-mesh namespace
# Usage: ./seldon-cli-remote.sh <command>
# command: the command to run on the seldon-cli pod in the format: `seldon` <args> 

N="${NAMESPACE:-seldon-mesh}"

kubectl exec seldon-cli -n ${N} -- $*