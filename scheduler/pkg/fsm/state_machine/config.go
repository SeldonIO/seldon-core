/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package state_machine

import (
	"time"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/server/filter"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/server/sorters"
)

// Config todo: this is very raw and was just a rough guide a lot of this options will not exist
type Config struct {
	// ========================================
	// Scheduling Policies
	// ========================================

	// Server selection strategy for model placement
	ServerSelectionStrategy ServerSelectionStrategy // "LeastLoaded", "RoundRobin", "Affinity"
	ServerSortingStrategy   ServerSortingStrategy

	// Whether to allow overcommit of server resources
	AllowOvercommit bool
	OvercommitRatio float64 // e.g., 1.5 = allow 150% of capacity

	serverFilters  []filter.ServerFilter
	replicaFilters []filter.ReplicaFilter
	serverSorts    []sorters.ServerSorter

	// ========================================
	// Model Replica Policies
	// ========================================

	ModelDeploymentStrategy ModelDeploymentStrategy

	// Default replica counts if not specified in model
	DefaultMinReplicas int
	DefaultMaxReplicas int

	// Auto-scaling policy
	AutoscalingEnabled bool
	AutoscalingPolicy  AutoscalingPolicy // "TargetUtilization", "TargetLatency", "None"

	// Replica health thresholds
	ReplicaFailureThreshold   int           // Failures before marking replica as failed
	ReplicaStartupTimeout     time.Duration // How long to wait for replica to start
	ReplicaUnhealthyThreshold int           // Consecutive unhealthy checks before removal

	// Model version management
	MaxModelVersionsRetained int  // How many old versions to keep
	AllowInPlaceUpdates      bool // Allow updating existing version vs always creating new

	// ========================================
	// Pipeline Policies
	// ========================================

	// How pipeline status is calculated from model statuses
	PipelineReadinessStrategy PipelineReadinessStrategy // "AllActive", "AtLeastOne", "Majority"

	// Whether pipeline failure should cascade to dependent pipelines
	PipelineFailurePropagation bool

	// ========================================
	// Experiment (A/B Testing) Policies
	// ========================================

	// Traffic splitting strategy
	ExperimentTrafficSplitStrategy TrafficSplitStrategy // "Weighted", "Canary", "ShadowTraffic"

	// Minimum traffic percentage for experiment variants
	MinExperimentTrafficPercent float64 // e.g., 0.01 = 1%

	// Auto-promotion thresholds
	ExperimentAutoPromoteEnabled    bool
	ExperimentAutoPromoteThreshold  float64       // e.g., error rate threshold
	ExperimentAutoRollbackThreshold float64       // Auto-rollback if errors exceed this
	ExperimentMinDuration           time.Duration // Min time before auto-promote

	// Whether to allow multiple experiments on same model
	AllowMultipleConcurrentExperiments bool

	// ========================================
	// Resource & Capacity Policies
	// ========================================

	// Server capacity management
	ServerCapacityStrategy CapacityStrategy // "Conservative", "Aggressive", "Balanced"

	// Whether to preempt lower-priority models for higher-priority ones
	EnableModelPreemption bool

	// Model priority weights (for scheduling decisions)
	DefaultModelPriority int

	// ========================================
	// Failure & Recovery Policies
	// ========================================

	// Model failure handling
	ModelFailureStrategy FailureStrategy // "Retry", "FailFast", "GracefulDegradation"
	MaxModelRetries      int

	// Server failure handling
	ServerEvictionTimeout time.Duration // How long to wait before evicting models from failed server

	// Whether to automatically reschedule models from failed servers
	AutoRescheduleOnServerFailure bool

	// ========================================
	// Dependency Resolution
	// ========================================

	// Pipeline model dependency ordering
	PipelineDependencyResolution DependencyResolution // "Parallel", "Sequential", "Optimistic"

	// Whether to allow partial pipeline activation
	AllowPartialPipelineActivation bool

	// ========================================
	// State Transition Rules
	// ========================================

	// Whether model can skip states (e.g., Pending -> Active without Loading)
	AllowStateSkipping bool

	// Grace period before marking model as failed
	ModelFailureGracePeriod time.Duration

	// Whether to allow model recreation while old version is unloading
	AllowModelRecreationDuringUnload bool
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		ServerSelectionStrategy: ServerSelectionLeastLoaded,
		ServerSortingStrategy:   ServerSelectionModelsLoaded,
		AllowOvercommit:         false,
		OvercommitRatio:         1.0,
		serverFilters: []filter.ServerFilter{
			filter.ServerReplicaFilter{}, filter.DeletedServerFilter{},
			filter.ServerRequirementFilter{}, filter.SharingServerFilter{},
		},
		ModelDeploymentStrategy:          ModelDeploymentAllAtOnce,
		DefaultMinReplicas:               1,
		DefaultMaxReplicas:               5,
		ReplicaFailureThreshold:          3,
		ReplicaStartupTimeout:            5 * time.Minute,
		MaxModelVersionsRetained:         3,
		AllowInPlaceUpdates:              false,
		PipelineReadinessStrategy:        PipelineAllActive,
		PipelineFailurePropagation:       false,
		ServerCapacityStrategy:           CapacityBalanced,
		EnableModelPreemption:            false,
		DefaultModelPriority:             100,
		ModelFailureStrategy:             FailureRetry,
		MaxModelRetries:                  3,
		ServerEvictionTimeout:            10 * time.Minute,
		AutoRescheduleOnServerFailure:    true,
		AllowStateSkipping:               false,
		ModelFailureGracePeriod:          2 * time.Minute,
		AllowModelRecreationDuringUnload: false,
	}
}

// ========================================
// Enum Types for Policies
// ========================================

type ServerSelectionStrategy string

const (
	ServerSelectionLeastLoaded ServerSelectionStrategy = "LeastLoaded"

	ServerSelectionRoundRobin   ServerSelectionStrategy = "RoundRobin"
	ServerSelectionAffinity     ServerSelectionStrategy = "Affinity"
	ServerSelectionAntiAffinity ServerSelectionStrategy = "AntiAffinity"
)

type ServerSortingStrategy string

const (
	ServerSelectionModelsLoaded ServerSortingStrategy = "ModelsLoaded"
)

type ModelDeploymentStrategy string

const (
	ModelDeploymentRollingUpdate = "RollingUpdate"
	ModelDeploymentAllAtOnce     = "AllAtOnce"
)

type AutoscalingPolicy string

const (
	AutoscalingNone              AutoscalingPolicy = "None"
	AutoscalingTargetUtilization AutoscalingPolicy = "TargetUtilization"
	AutoscalingTargetLatency     AutoscalingPolicy = "TargetLatency"
	AutoscalingTargetThroughput  AutoscalingPolicy = "TargetThroughput"
)

type PipelineReadinessStrategy string

const (
	PipelineAllActive  PipelineReadinessStrategy = "AllActive"  // All models must be active
	PipelineAtLeastOne PipelineReadinessStrategy = "AtLeastOne" // At least one model active
	PipelineMajority   PipelineReadinessStrategy = "Majority"   // Majority of models active
)

type TrafficSplitStrategy string

const (
	TrafficSplitWeighted      TrafficSplitStrategy = "Weighted"      // Static weights
	TrafficSplitCanary        TrafficSplitStrategy = "Canary"        // Gradual rollout
	TrafficSplitShadowTraffic TrafficSplitStrategy = "ShadowTraffic" // Shadow without serving
	TrafficSplitMirror        TrafficSplitStrategy = "Mirror"        // Duplicate traffic to both
)

type CapacityStrategy string

const (
	CapacityConservative CapacityStrategy = "Conservative" // Leave buffer
	CapacityBalanced     CapacityStrategy = "Balanced"     // Default
	CapacityAggressive   CapacityStrategy = "Aggressive"   // Pack tightly
)

type FailureStrategy string

const (
	FailureRetry               FailureStrategy = "Retry"
	FailureFailFast            FailureStrategy = "FailFast"
	FailureGracefulDegradation FailureStrategy = "GracefulDegradation"
)

type DependencyResolution string

const (
	DependencyParallel   DependencyResolution = "Parallel"   // Start all models concurrently
	DependencySequential DependencyResolution = "Sequential" // Start in dependency order
	DependencyOptimistic DependencyResolution = "Optimistic" // Start immediately, resolve later
)
