/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.hashutils

import java.math.BigInteger
import java.security.MessageDigest

object HashUtils {
    private const val ALGO_MD5 = "MD5"
    private const val MAX_OUTPUT_LENGTH = 16

    fun hashIfLong(input: String): String {
        if (input.length <= MAX_OUTPUT_LENGTH) {
            return input
        }

        val md = MessageDigest.getInstance(ALGO_MD5)
        val hashedBytes = md.digest(input.toByteArray())
        return BigInteger(1, hashedBytes)
            .toString(16)
            .padStart(MAX_OUTPUT_LENGTH, '0')
    }
}
