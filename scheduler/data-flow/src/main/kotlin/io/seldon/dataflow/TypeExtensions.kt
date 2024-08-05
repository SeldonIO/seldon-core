/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package io.seldon.dataflow

import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.FlowPreview
import kotlinx.coroutines.async
import kotlinx.coroutines.flow.DEFAULT_CONCURRENCY
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flatMapMerge
import kotlinx.coroutines.flow.flow

@OptIn(FlowPreview::class, ExperimentalCoroutinesApi::class)
suspend fun <T, R> Flow<T>.parallel(
    scope: CoroutineScope,
    concurrency: Int = DEFAULT_CONCURRENCY,
    transform: suspend (T) -> R,
): Flow<R> {
    return with(scope) {
        this@parallel
            .flatMapMerge(concurrency) { value ->
                flow {
                    emit(
                        async { transform(value) },
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
