/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

// Generated by the protocol buffer compiler. DO NOT EDIT!
// NO CHECKED-IN PROTOBUF GENCODE
// source: v2_dataplane.proto

// Generated files should ignore deprecation warnings
@file:Suppress("DEPRECATION")
package io.seldon.mlops.inference.v2;

@kotlin.jvm.JvmName("-initializeserverMetadataResponse")
public inline fun serverMetadataResponse(block: io.seldon.mlops.inference.v2.ServerMetadataResponseKt.Dsl.() -> kotlin.Unit): io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataResponse =
  io.seldon.mlops.inference.v2.ServerMetadataResponseKt.Dsl._create(io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataResponse.newBuilder()).apply { block() }._build()
/**
 * Protobuf type `inference.ServerMetadataResponse`
 */
public object ServerMetadataResponseKt {
  @kotlin.OptIn(com.google.protobuf.kotlin.OnlyForUseByGeneratedProtoCode::class)
  @com.google.protobuf.kotlin.ProtoDslMarker
  public class Dsl private constructor(
    private val _builder: io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataResponse.Builder
  ) {
    public companion object {
      @kotlin.jvm.JvmSynthetic
      @kotlin.PublishedApi
      internal fun _create(builder: io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataResponse.Builder): Dsl = Dsl(builder)
    }

    @kotlin.jvm.JvmSynthetic
    @kotlin.PublishedApi
    internal fun _build(): io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataResponse = _builder.build()

    /**
     * ```
     * The server name.
     * ```
     *
     * `string name = 1;`
     */
    public var name: kotlin.String
      @JvmName("getName")
      get() = _builder.getName()
      @JvmName("setName")
      set(value) {
        _builder.setName(value)
      }
    /**
     * ```
     * The server name.
     * ```
     *
     * `string name = 1;`
     */
    public fun clearName() {
      _builder.clearName()
    }

    /**
     * ```
     * The server version.
     * ```
     *
     * `string version = 2;`
     */
    public var version: kotlin.String
      @JvmName("getVersion")
      get() = _builder.getVersion()
      @JvmName("setVersion")
      set(value) {
        _builder.setVersion(value)
      }
    /**
     * ```
     * The server version.
     * ```
     *
     * `string version = 2;`
     */
    public fun clearVersion() {
      _builder.clearVersion()
    }

    /**
     * An uninstantiable, behaviorless type to represent the field in
     * generics.
     */
    @kotlin.OptIn(com.google.protobuf.kotlin.OnlyForUseByGeneratedProtoCode::class)
    public class ExtensionsProxy private constructor() : com.google.protobuf.kotlin.DslProxy()
    /**
     * ```
     * The extensions supported by the server.
     * ```
     *
     * `repeated string extensions = 3;`
     * @return A list containing the extensions.
     */
    public val extensions: com.google.protobuf.kotlin.DslList<kotlin.String, ExtensionsProxy>
      @kotlin.jvm.JvmSynthetic
      get() = com.google.protobuf.kotlin.DslList(
        _builder.getExtensionsList()
      )
    /**
     * ```
     * The extensions supported by the server.
     * ```
     *
     * `repeated string extensions = 3;`
     * @param value The extensions to add.
     */
    @kotlin.jvm.JvmSynthetic
    @kotlin.jvm.JvmName("addExtensions")
    public fun com.google.protobuf.kotlin.DslList<kotlin.String, ExtensionsProxy>.add(value: kotlin.String) {
      _builder.addExtensions(value)
    }
    /**
     * ```
     * The extensions supported by the server.
     * ```
     *
     * `repeated string extensions = 3;`
     * @param value The extensions to add.
     */
    @kotlin.jvm.JvmSynthetic
    @kotlin.jvm.JvmName("plusAssignExtensions")
    @Suppress("NOTHING_TO_INLINE")
    public inline operator fun com.google.protobuf.kotlin.DslList<kotlin.String, ExtensionsProxy>.plusAssign(value: kotlin.String) {
      add(value)
    }
    /**
     * ```
     * The extensions supported by the server.
     * ```
     *
     * `repeated string extensions = 3;`
     * @param values The extensions to add.
     */
    @kotlin.jvm.JvmSynthetic
    @kotlin.jvm.JvmName("addAllExtensions")
    public fun com.google.protobuf.kotlin.DslList<kotlin.String, ExtensionsProxy>.addAll(values: kotlin.collections.Iterable<kotlin.String>) {
      _builder.addAllExtensions(values)
    }
    /**
     * ```
     * The extensions supported by the server.
     * ```
     *
     * `repeated string extensions = 3;`
     * @param values The extensions to add.
     */
    @kotlin.jvm.JvmSynthetic
    @kotlin.jvm.JvmName("plusAssignAllExtensions")
    @Suppress("NOTHING_TO_INLINE")
    public inline operator fun com.google.protobuf.kotlin.DslList<kotlin.String, ExtensionsProxy>.plusAssign(values: kotlin.collections.Iterable<kotlin.String>) {
      addAll(values)
    }
    /**
     * ```
     * The extensions supported by the server.
     * ```
     *
     * `repeated string extensions = 3;`
     * @param index The index to set the value at.
     * @param value The extensions to set.
     */
    @kotlin.jvm.JvmSynthetic
    @kotlin.jvm.JvmName("setExtensions")
    public operator fun com.google.protobuf.kotlin.DslList<kotlin.String, ExtensionsProxy>.set(index: kotlin.Int, value: kotlin.String) {
      _builder.setExtensions(index, value)
    }/**
     * ```
     * The extensions supported by the server.
     * ```
     *
     * `repeated string extensions = 3;`
     */
    @kotlin.jvm.JvmSynthetic
    @kotlin.jvm.JvmName("clearExtensions")
    public fun com.google.protobuf.kotlin.DslList<kotlin.String, ExtensionsProxy>.clear() {
      _builder.clearExtensions()
    }}
}
@kotlin.jvm.JvmSynthetic
public inline fun io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataResponse.copy(block: `io.seldon.mlops.inference.v2`.ServerMetadataResponseKt.Dsl.() -> kotlin.Unit): io.seldon.mlops.inference.v2.V2Dataplane.ServerMetadataResponse =
  `io.seldon.mlops.inference.v2`.ServerMetadataResponseKt.Dsl._create(this.toBuilder()).apply { block() }._build()

