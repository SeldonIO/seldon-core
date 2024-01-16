/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/


package io.seldon.dataflow.kafka.security

import org.apache.kafka.common.security.auth.SecurityProtocol

val KafkaSecurityProtocols = arrayOf(
    SecurityProtocol.PLAINTEXT,
    SecurityProtocol.SSL,
    SecurityProtocol.SASL_SSL,
)