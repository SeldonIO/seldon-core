package io.seldon.test.utils;

import org.apache.kafka.common.header.Headers;
import org.apache.kafka.common.header.internals.RecordHeaders;
import org.apache.kafka.common.metrics.MetricConfig;
import org.apache.kafka.common.metrics.Metrics;
import org.apache.kafka.common.metrics.Sensor;
import org.apache.kafka.common.serialization.Serde;
import org.apache.kafka.common.serialization.Serdes;
import org.apache.kafka.common.utils.Time;
import org.apache.kafka.streams.StreamsMetrics;
import org.apache.kafka.streams.processor.*;
import org.apache.kafka.streams.processor.api.FixedKeyProcessorContext;
import org.apache.kafka.streams.processor.api.FixedKeyRecord;
import org.apache.kafka.streams.processor.api.RecordMetadata;
import org.apache.kafka.streams.processor.internals.metrics.StreamsMetricsImpl;
import org.apache.kafka.streams.processor.internals.metrics.TaskMetrics;

import java.io.File;
import java.time.Duration;
import java.util.*;

/**
 * Mock implementation of FixedKeyProcessorContext for testing FixedKeyProcessor implementations.
 */
public class MockFixedKeyProcessorContext<KForward, VForward> implements FixedKeyProcessorContext<KForward, VForward> {

    private final Map<String, StateStore> stateStores = new HashMap<>();
    private final List<CapturedForward<? extends KForward, ? extends VForward>> capturedForwards = new LinkedList<>();
    private final List<CapturedPunctuator> punctuators = new ArrayList<>();
    private final Properties appConfigs = new Properties();
    private final StreamsMetricsImpl metrics;
    
    private final String applicationId;
    private final TaskId taskId;
    private final Serde<KForward> keySerde;
    private final Serde<VForward> valueSerde;
    private final File stateDir;
    private final Headers recordHeaders;
    
    private long currentTimestamp = System.currentTimeMillis();
    private String currentTopic = "mock-topic";
    private int currentPartition = 0;
    private long currentOffset = 0L;
    private boolean committed = false;

    /**
     * Creates a new MockFixedKeyProcessorContext with default String serdes.
     */
    @SuppressWarnings("unchecked")
    public MockFixedKeyProcessorContext() {
        this("mock-application", new TaskId(0, 0), 
             (Serde<KForward>) Serdes.String(), 
             (Serde<VForward>) Serdes.String());
    }

    /**
     * Creates a new MockFixedKeyProcessorContext with the specified parameters.
     * 
     * @param applicationId the application ID
     * @param taskId the task ID
     * @param keySerde the key serde
     * @param valueSerde the value serde
     */
    public MockFixedKeyProcessorContext(String applicationId, TaskId taskId, 
                                        Serde<KForward> keySerde, Serde<VForward> valueSerde) {
        this.applicationId = applicationId;
        this.taskId = taskId;
        this.keySerde = keySerde;
        this.valueSerde = valueSerde;
        this.stateDir = new File(System.getProperty("java.io.tmpdir"), "kafka-streams-test");
        this.recordHeaders = new RecordHeaders();
        
        // Set some default application configs
        this.appConfigs.setProperty("application.id", applicationId);
        this.appConfigs.setProperty("bootstrap.servers", "localhost:9092");

        final MetricConfig metricConfig = new MetricConfig();
        metricConfig.recordLevel(Sensor.RecordingLevel.DEBUG);
        final String threadId = Thread.currentThread().getName();
        metrics = new StreamsMetricsImpl(
                new Metrics(metricConfig),
                threadId,
                "processId",
                Time.SYSTEM
        );
        TaskMetrics.droppedRecordsSensor(threadId, taskId.toString(), metrics);
    }

    @Override
    public String applicationId() {
        return applicationId;
    }

    @Override
    public TaskId taskId() {
        return taskId;
    }

    @Override
    public Optional<RecordMetadata> recordMetadata() {
        return Optional.empty();
    }

    @Override
    public Serde<?> keySerde() {
        return keySerde;
    }

    @Override
    public Serde<?> valueSerde() {
        return valueSerde;
    }

    @Override
    public File stateDir() {
        return stateDir;
    }

    @Override
    public StreamsMetrics metrics() {
        return metrics;
    }

    @Override
    public <S extends StateStore> S getStateStore(String name) {
        return (S) stateStores.get(name);
    }

    @Override
    public Cancellable schedule(Duration interval, PunctuationType type, Punctuator callback) {
        CapturedPunctuator punctuator = new CapturedPunctuator(interval, type, callback);
        punctuators.add(punctuator);
        return punctuator;
    }

    @Override
    public void commit() {
        this.committed = true;
    }

    @Override
    public <K extends KForward, V extends VForward> void forward(FixedKeyRecord<K, V> record) {
        forward(record, null);

    }

    @Override
    public <K extends KForward, V extends VForward> void forward(FixedKeyRecord<K, V> record, String childName) {
        capturedForwards.add(new CapturedForward<>(record, Optional.ofNullable(childName)));
    }

    @Override
    public Map<String, Object> appConfigs() {
        Map<String, Object> configs = new HashMap<>();
        for (String key : appConfigs.stringPropertyNames()) {
            configs.put(key, appConfigs.getProperty(key));
        }
        return configs;
    }

    @Override
    public Map<String, Object> appConfigsWithPrefix(String prefix) {
        Map<String, Object> configsWithPrefix = new HashMap<>();
        for (String key : appConfigs.stringPropertyNames()) {
            if (key.startsWith(prefix)) {
                configsWithPrefix.put(key.substring(prefix.length()), appConfigs.getProperty(key));
            }
        }
        return configsWithPrefix;
    }

    @Override
    public long currentSystemTimeMs() {
        return 0;
    }

    @Override
    public long currentStreamTimeMs() {
        return 0;
    }

    public Headers headers() {
        return recordHeaders;
    }

    public long timestamp() {
        return currentTimestamp;
    }

    public String topic() {
        return currentTopic;
    }

    public void setTopic(String topicName) {
        currentTopic = topicName;
    }

    public int partition() {
        return currentPartition;
    }

    public void setPartition(int partitionNo) {
        currentPartition = partitionNo;
    }

    public long offset() {
        return currentOffset;
    }

    // Test utility methods

    /**
     * All forwarded records, so that the test can inspect the exact behaviour of the
     * system under test (SuT): a class implementing FixedKeyProcessor
     */
    public List<CapturedForward<? extends KForward, ? extends VForward>> forwarded() {
        return new LinkedList<>(capturedForwards);
    }

    public void resetForwards() {
        capturedForwards.clear();
    }

    public boolean committed() {
        return committed;
    }

    public void resetCommit() {
        committed = false;
    }

    public List<CapturedPunctuator> scheduledPunctuators() {
        return new ArrayList<>(punctuators);
    }

    public void setRecordContext(String topic, int partition, long offset, long timestamp) {
        this.currentTopic = topic;
        this.currentPartition = partition;
        this.currentOffset = offset;
        this.currentTimestamp = timestamp;
    }

    public void addStateStore(StateStore store) {
        stateStores.put(store.name(), store);
    }

    public void addConfig(String key, String value) {
        appConfigs.setProperty(key, value);
    }

    /**
     * Used to get a {@link StateStoreContext} for use with
     * {@link StateStore#init(StateStoreContext, StateStore)}
     * if you need to initialize a store for your tests.
     * @return a {@link StateStoreContext} that delegates to this FixedKeyProcessorContext.
     *
     * Same overall approach as in MockProcessorContext, but adapted to FixedKeyProcessor.
     */
    public StateStoreContext getStateStoreContext() {
        return new StateStoreContext() {
            @Override
            public String applicationId() {
                return MockFixedKeyProcessorContext.this.applicationId();
            }

            @Override
            public TaskId taskId() {
                return MockFixedKeyProcessorContext.this.taskId();
            }

            @Override
            public Optional<RecordMetadata> recordMetadata() {
                return MockFixedKeyProcessorContext.this.recordMetadata();
            }

            @Override
            public Serde<?> keySerde() {
                return MockFixedKeyProcessorContext.this.keySerde();
            }

            @Override
            public Serde<?> valueSerde() {
                return MockFixedKeyProcessorContext.this.valueSerde();
            }

            @Override
            public File stateDir() {
                return MockFixedKeyProcessorContext.this.stateDir();
            }

            @Override
            public StreamsMetrics metrics() {
                return MockFixedKeyProcessorContext.this.metrics();
            }

            @Override
            public void register(final StateStore store,
                                 final StateRestoreCallback stateRestoreCallback) {
                register(store, stateRestoreCallback, () -> { });
            }

            @Override
            public void register(final StateStore store,
                                 final StateRestoreCallback stateRestoreCallback,
                                 final CommitCallback checkpoint) {
                stateStores.put(store.name(), store);
            }

            @Override
            public Map<String, Object> appConfigs() {
                return MockFixedKeyProcessorContext.this.appConfigs();
            }

            @Override
            public Map<String, Object> appConfigsWithPrefix(final String prefix) {
                return MockFixedKeyProcessorContext.this.appConfigsWithPrefix(prefix);
            }
        };
    }

    /**
     * Captures information about scheduled punctuators for testing.
     */
    public static class CapturedPunctuator implements Cancellable {
        private final Duration interval;
        private final PunctuationType type;
        private final Punctuator punctuator;
        private boolean cancelled = false;

        CapturedPunctuator(Duration interval, PunctuationType type, Punctuator punctuator) {
            this.interval = interval;
            this.type = type;
            this.punctuator = punctuator;
        }

        public Duration getInterval() {
            return interval;
        }

        public PunctuationType getType() {
            return type;
        }

        public Punctuator getPunctuator() {
            return punctuator;
        }

        public boolean isCancelled() {
            return cancelled;
        }

        @Override
        public void cancel() {
            cancelled = true;
        }

        /**
         * Executes the punctuator for testing purposes.
         */
        public void punctuate(long timestamp) {
            if (!cancelled) {
                punctuator.punctuate(timestamp);
            }
        }
    }

    /**
     * Data internally captured on each Processor call to "forward"
     */
    public static final class CapturedForward<K, V> {

        private final FixedKeyRecord<K, V> record;
        private final Optional<String> childName;

        public CapturedForward(final FixedKeyRecord<K, V> record) {
            this(record, Optional.empty());
        }

        public CapturedForward(final FixedKeyRecord<K, V> record, final Optional<String> childName) {
            this.record = Objects.requireNonNull(record);
            this.childName = Objects.requireNonNull(childName);
        }

        public Optional<String> childName() {
            return childName;
        }

        public FixedKeyRecord<K, V> record() {
            return record;
        }

        @Override
        public String toString() {
            return "CapturedForward{" +
                    "record=" + record +
                    ", childName=" + childName +
                    '}';
        }

        @Override
        public boolean equals(final Object o) {
            if (this == o) return true;
            if (o == null || getClass() != o.getClass()) return false;
            final CapturedForward<?, ?> that = (CapturedForward<?, ?>) o;
            return Objects.equals(record, that.record) &&
                    Objects.equals(childName, that.childName);
        }

        @Override
        public int hashCode() {
            return Objects.hash(record, childName);
        }
    }
}

