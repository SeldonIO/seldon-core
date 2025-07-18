/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

// Generated by the protocol buffer compiler. DO NOT EDIT!
// NO CHECKED-IN PROTOBUF GENCODE
// source: chainer.proto

// Generated files should ignore deprecation warnings
@file:Suppress("DEPRECATION")
package io.seldon.mlops.chainer;

@kotlin.jvm.JvmName("-initializebatch")
public inline fun batch(block: io.seldon.mlops.chainer.BatchKt.Dsl.() -> kotlin.Unit): io.seldon.mlops.chainer.ChainerOuterClass.Batch =
  io.seldon.mlops.chainer.BatchKt.Dsl._create(io.seldon.mlops.chainer.ChainerOuterClass.Batch.newBuilder()).apply { block() }._build()
/**
 * Protobuf type `seldon.mlops.chainer.Batch`
 */
public object BatchKt {
  @kotlin.OptIn(com.google.protobuf.kotlin.OnlyForUseByGeneratedProtoCode::class)
  @com.google.protobuf.kotlin.ProtoDslMarker
  public class Dsl private constructor(
    private val _builder: io.seldon.mlops.chainer.ChainerOuterClass.Batch.Builder
  ) {
    public companion object {
      @kotlin.jvm.JvmSynthetic
      @kotlin.PublishedApi
      internal fun _create(builder: io.seldon.mlops.chainer.ChainerOuterClass.Batch.Builder): Dsl = Dsl(builder)
    }

    @kotlin.jvm.JvmSynthetic
    @kotlin.PublishedApi
    internal fun _build(): io.seldon.mlops.chainer.ChainerOuterClass.Batch = _builder.build()

    /**
     * `optional uint32 size = 1;`
     */
    public var size: kotlin.Int
      @JvmName("getSize")
      get() = _builder.getSize()
      @JvmName("setSize")
      set(value) {
        _builder.setSize(value)
      }
    /**
     * `optional uint32 size = 1;`
     */
    public fun clearSize() {
      _builder.clearSize()
    }
    /**
     * `optional uint32 size = 1;`
     * @return Whether the size field is set.
     */
    public fun hasSize(): kotlin.Boolean {
      return _builder.hasSize()
    }

    /**
     * `optional uint32 windowMs = 2;`
     */
    public var windowMs: kotlin.Int
      @JvmName("getWindowMs")
      get() = _builder.getWindowMs()
      @JvmName("setWindowMs")
      set(value) {
        _builder.setWindowMs(value)
      }
    /**
     * `optional uint32 windowMs = 2;`
     */
    public fun clearWindowMs() {
      _builder.clearWindowMs()
    }
    /**
     * `optional uint32 windowMs = 2;`
     * @return Whether the windowMs field is set.
     */
    public fun hasWindowMs(): kotlin.Boolean {
      return _builder.hasWindowMs()
    }

    /**
     * `bool rolling = 3;`
     */
    public var rolling: kotlin.Boolean
      @JvmName("getRolling")
      get() = _builder.getRolling()
      @JvmName("setRolling")
      set(value) {
        _builder.setRolling(value)
      }
    /**
     * `bool rolling = 3;`
     */
    public fun clearRolling() {
      _builder.clearRolling()
    }
  }
}
@kotlin.jvm.JvmSynthetic
public inline fun io.seldon.mlops.chainer.ChainerOuterClass.Batch.copy(block: `io.seldon.mlops.chainer`.BatchKt.Dsl.() -> kotlin.Unit): io.seldon.mlops.chainer.ChainerOuterClass.Batch =
  `io.seldon.mlops.chainer`.BatchKt.Dsl._create(this.toBuilder()).apply { block() }._build()

