/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka

import strikt.api.Assertion
import strikt.api.expect
import strikt.assertions.isEqualTo

fun Assertion.Builder<PipelineStep>.isSameTypeAs(other: PipelineStep) =
    assert("Same type") {
        when {
            it::class == other::class -> pass()
            else -> fail(actual = other::class.simpleName)
        }
    }

fun Assertion.Builder<PipelineStep>.matches(expected: PipelineStep) =
    assert("Type and values are the same") {
        when {
            it is Chainer && expected is Chainer ->
                expect {
                    that(it) {
                        get { inputTopic }.isEqualTo(expected.inputTopic)
                        get { outputTopic }.isEqualTo(expected.outputTopic)
                        get { tensors }.isEqualTo(expected.tensors)
                    }
                }
            it is Joiner && expected is Joiner ->
                expect {
                    that(it) {
                        get { inputTopics }.isEqualTo(expected.inputTopics)
                        get { outputTopic }.isEqualTo(expected.outputTopic)
                        get { tensorsByTopic }.isEqualTo(expected.tensorsByTopic)
                        get { tensorRenaming }.isEqualTo(expected.tensorRenaming)
                        get { kafkaDomainParams }.isEqualTo(expected.kafkaDomainParams)
                    }
                }
            else -> fail(actual = expected)
        }
    }
