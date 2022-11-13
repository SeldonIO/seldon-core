/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package io.seldon.dataflow.hashutils

import java.math.BigInteger
import java.security.MessageDigest

object HashUtils {
    private const val algoMD5 = "MD5"
    private const val maxOutputLength = 16

    fun hashIfLong(input: String): String {
        if (input.length <= maxOutputLength) {
            return input
        }

        val md = MessageDigest.getInstance(algoMD5)
        val hashedBytes = md.digest(input.toByteArray())
        return BigInteger(1, hashedBytes)
            .toString(16)
            .padStart(maxOutputLength, '0')
    }
}