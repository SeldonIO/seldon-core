/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow.kafka

import org.apache.kafka.streams.KafkaStreams
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.Arguments
import org.junit.jupiter.params.provider.Arguments.arguments
import org.junit.jupiter.params.provider.MethodSource
import strikt.api.expect
import strikt.api.expectThat
import strikt.assertions.isEqualTo
import java.util.stream.Stream

internal class PipelineStatusIsActiveIsErrorTest {
    @DisplayName("check PipelineStatus properties")
    @ParameterizedTest(name = "{0} expected isActive:{1} | isError:{2}")
    @MethodSource()
    fun checkState(
        testName: String,
        expectedIsActive: Boolean,
        expectedIsError: Boolean,
        state: PipelineStatus,
    ) {
        val statusIsActive = state.isActive()
        val statusIsError = state.isError()

        expect {
            that(statusIsActive).isEqualTo(expectedIsActive)
            that(statusIsError).isEqualTo(expectedIsError)
        }
    }

    companion object {
        private const val IS_ACTIVE = true
        private const val IS_ERROR = true

        @JvmStatic
        fun checkState(): Stream<Arguments> =
            Stream.of(
                arguments(
                    "StreamStopped(prevState=null)",
                    !IS_ACTIVE,
                    !IS_ERROR,
                    PipelineStatus.StreamStopped(null),
                ),
                arguments(
                    "StreamStopped(prevState=StreamStopping)",
                    !IS_ACTIVE,
                    !IS_ERROR,
                    PipelineStatus.StreamStopped(PipelineStatus.StreamStopping()),
                ),
                arguments(
                    "StreamStopped(prevState=nested StreamStopped without error)",
                    !IS_ACTIVE,
                    !IS_ERROR,
                    PipelineStatus.StreamStopped(
                        PipelineStatus.StreamStopped(
                            PipelineStatus.Started(),
                        ),
                    ),
                ),
                arguments(
                    "StreamStopping",
                    !IS_ACTIVE,
                    !IS_ERROR,
                    PipelineStatus.StreamStopping(),
                ),
                arguments(
                    "StreamStopped(prevState=Error)",
                    !IS_ACTIVE,
                    IS_ERROR,
                    PipelineStatus.StreamStopped(PipelineStatus.Error(null)),
                ),
                arguments(
                    "StreamStopped(prevState=nested StreamStopped with Error)",
                    !IS_ACTIVE,
                    IS_ERROR,
                    PipelineStatus.StreamStopped(
                        PipelineStatus.StreamStopped(
                            PipelineStatus.Error(KafkaStreams.State.ERROR),
                        ),
                    ),
                ),
                arguments(
                    "Error(errorState=null)",
                    !IS_ACTIVE,
                    IS_ERROR,
                    PipelineStatus.Error(null),
                ),
                arguments(
                    "Error(errorState=non-error,active state)",
                    IS_ACTIVE,
                    IS_ERROR,
                    PipelineStatus.Error(KafkaStreams.State.RUNNING),
                ),
                arguments(
                    "Error(errorState=non-error,non-active state)",
                    !IS_ACTIVE,
                    IS_ERROR,
                    PipelineStatus.Error(KafkaStreams.State.NOT_RUNNING),
                ),
                arguments(
                    "Error(errorState=error state)",
                    !IS_ACTIVE,
                    IS_ERROR,
                    PipelineStatus.Error(KafkaStreams.State.PENDING_ERROR),
                ),
                arguments(
                    "PipelineStatus(state=error state, hasError=false)",
                    !IS_ACTIVE,
                    IS_ERROR,
                    PipelineStatus(KafkaStreams.State.PENDING_ERROR, false),
                ),
                arguments(
                    "PipelineStatus(state=error state, hasError=true)",
                    !IS_ACTIVE,
                    IS_ERROR,
                    PipelineStatus(KafkaStreams.State.PENDING_ERROR, true),
                ),
                arguments(
                    "PipelineStatus(state=non-error,non-active state, hasError=false)",
                    !IS_ACTIVE,
                    !IS_ERROR,
                    PipelineStatus(KafkaStreams.State.NOT_RUNNING, false),
                ),
                arguments(
                    "PipelineStatus(state=non-error,non-active state, hasError=true)",
                    !IS_ACTIVE,
                    IS_ERROR,
                    PipelineStatus(KafkaStreams.State.NOT_RUNNING, true),
                ),
                arguments(
                    "PipelineStatus(state=non-error,active state, hasError=false)",
                    IS_ACTIVE,
                    !IS_ERROR,
                    PipelineStatus(KafkaStreams.State.CREATED, false),
                ),
                arguments(
                    "PipelineStatus(state=non-error,active state, hasError=true)",
                    IS_ACTIVE,
                    IS_ERROR,
                    PipelineStatus(KafkaStreams.State.CREATED, true),
                ),
                arguments(
                    "StreamStarting",
                    IS_ACTIVE,
                    !IS_ERROR,
                    PipelineStatus.StreamStarting(),
                ),
                arguments(
                    "Started",
                    IS_ACTIVE,
                    !IS_ERROR,
                    PipelineStatus.Started(),
                ),
            )
    }
}

internal class PipelineStatusTest {
    @Test
    fun `check StreamStopped(prevState) state nesting is bounded to 1 level`() {
        // In the following nested initialisation of StreamStopped(prevState) objects,
        // each StreamStopped should unwrap the existing StreamStopped prevState and
        // only store the original prevState inside it (Error in this case).
        // We test that this works correctly.
        val nest =
            PipelineStatus.StreamStopped(
                PipelineStatus.StreamStopped(
                    PipelineStatus.StreamStopped(
                        PipelineStatus.StreamStopped(
                            PipelineStatus.StreamStopped(
                                PipelineStatus.Error(null),
                            ),
                        ),
                    ),
                ),
            )

        expectThat(nest).isEqualTo(
            PipelineStatus.StreamStopped(
                PipelineStatus.Error(null),
            ),
        )
    }
}
