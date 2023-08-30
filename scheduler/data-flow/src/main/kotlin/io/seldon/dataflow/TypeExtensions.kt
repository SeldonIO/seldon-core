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

import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.FlowPreview
import kotlinx.coroutines.async
import kotlinx.coroutines.flow.DEFAULT_CONCURRENCY
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flatMapMerge
import kotlinx.coroutines.flow.flow
import java.util.*

@OptIn(FlowPreview::class)
suspend fun <T, R> Flow<T>.parallel(
    scope: CoroutineScope,
    concurrency: Int = DEFAULT_CONCURRENCY,
    transform: suspend (T) -> R
): Flow<R> {
    return with(scope) {
        this@parallel
            .flatMapMerge(concurrency) { value ->
                flow {
                    emit(
                        async { transform(value) }
                    )
                }
            }
            .flatMapMerge {
                flow {
                    emit(it.await())
                }
            }
    }
}

fun ByteArray.decodeBase64() = Base64.getUrlDecoder().decode(this)
