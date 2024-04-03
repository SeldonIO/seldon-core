/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow

import kotlinx.coroutines.FlowPreview
import kotlinx.coroutines.async
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.Test
import java.time.LocalDateTime

@OptIn(FlowPreview::class)
internal class PipelineSubscriberTest {
    @Test
    fun `should run sequentially`() {
        suspend fun waitAndPrint(i: Int) {
            kotlinx.coroutines.delay(1000)
            println("${LocalDateTime.now()} - $i")
        }

        val xs = (1..10).asFlow()
        runBlocking {
            xs
                .onEach { waitAndPrint(it) }
                .collect()
        }
    }

    @Test
    fun `should run ops concurrently`() {
        val xs = (1..10).asFlow()
        runBlocking {
            xs
                .flatMapMerge {
                    flow {
                        emit(
                            async {
                                kotlinx.coroutines.delay(1000)
                                println("${LocalDateTime.now()} - $it")
                            },
                        )
                    }
                }
                .flatMapMerge { flow { emit(it.await()) } }
                .collect()
        }
    }

    @Test
    fun `should run ops in parallel`() {
        suspend fun waitAndPrint(i: Int) {
            kotlinx.coroutines.delay(1000)
            println("${LocalDateTime.now()} - $i")
        }

        val xs = (1..10).asFlow()
        runBlocking {
            xs
                .parallel(this) { waitAndPrint(it) }
                .collect()
        }
    }
}
