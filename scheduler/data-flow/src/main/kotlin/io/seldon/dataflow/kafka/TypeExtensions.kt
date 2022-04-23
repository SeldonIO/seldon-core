package io.seldon.dataflow.kafka

operator fun Set<TensorName>?.contains(tensor: TensorName): Boolean {
    return this?.let {
        tensor in this
    } ?: true
}

