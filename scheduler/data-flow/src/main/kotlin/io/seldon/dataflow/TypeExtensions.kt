package io.seldon.dataflow

import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.FlowPreview
import kotlinx.coroutines.async
import kotlinx.coroutines.flow.DEFAULT_CONCURRENCY
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flatMapMerge
import kotlinx.coroutines.flow.flow

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