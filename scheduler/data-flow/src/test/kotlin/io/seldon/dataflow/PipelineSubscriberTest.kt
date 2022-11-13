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
                            }
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