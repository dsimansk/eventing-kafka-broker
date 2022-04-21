/*
 * Copyright © 2018 Knative Authors (knative-dev@googlegroups.com)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package dev.knative.eventing.kafka.broker.dispatcher.impl;

import dev.knative.eventing.kafka.broker.core.AsyncCloseable;
import dev.knative.eventing.kafka.broker.core.metrics.Metrics;
import dev.knative.eventing.kafka.broker.core.tracing.TracingSpan;
import dev.knative.eventing.kafka.broker.dispatcher.ResponseHandler;
import io.cloudevents.CloudEvent;
import io.cloudevents.http.vertx.VertxMessageFactory;
import io.vertx.core.CompositeFuture;
import io.vertx.core.Future;
import io.vertx.core.buffer.Buffer;
import io.vertx.ext.web.client.HttpResponse;
import io.vertx.kafka.client.producer.KafkaProducer;
import io.vertx.kafka.client.producer.KafkaProducerRecord;
import java.util.Objects;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import static dev.knative.eventing.kafka.broker.core.utils.Logging.keyValue;

/**
 * This class implements a {@link ResponseHandler} that will convert the sink response into a {@link CloudEvent} and push it to Kafka.
 */
public final class KafkaResponseHandler implements ResponseHandler {

  private static final Logger logger = LoggerFactory.getLogger(KafkaResponseHandler.class);

  private final String topic;
  private final KafkaProducer<String, CloudEvent> producer;
  private final AutoCloseable producerMeterBinder;

  /**
   * All args constructor.
   *
   * @param producer Kafka producer.
   * @param topic    topic to produce records.
   */
  public KafkaResponseHandler(final KafkaProducer<String, CloudEvent> producer, final String topic) {

    Objects.requireNonNull(topic, "provide topic");
    Objects.requireNonNull(producer, "provide producer");

    this.topic = topic;
    this.producer = producer;
    this.producerMeterBinder = Metrics.register(this.producer.unwrap());
  }

  /**
   * Handle the given response.
   *
   * @param response response to handle
   * @return a succeeded or failed future.
   */
  @Override
  public Future<Void> handle(final HttpResponse<Buffer> response) {
    CloudEvent event;

    try {
      event = VertxMessageFactory.createReader(response).toEvent();
    } catch (final Exception ex) {
      if (maybeIsNotEvent(response)) {
        if (logger.isDebugEnabled()) {
          logger.debug(
            "Response is not recognized as event, discarding it {} {} {}",
            keyValue("response", response),
            keyValue("response.body", response == null || response.body() == null ? "null" : response.body()),
            keyValue("response.body.len",
              response == null || response.body() == null ? "null" : response.body().length())
          );
        }
        return Future.succeededFuture();
      }

      // When the sink returns a malformed event we return a failed future to avoid committing the message to Kafka.
      return Future.failedFuture(
        new IllegalStateException("Unable to decode response: unknown encoding and non empty response", ex)
      );
    }

    if (event == null) {
      return Future.failedFuture(new IllegalArgumentException("event cannot be null"));
    }

    TracingSpan.decorateCurrentWithEvent(event);

    return producer
      .send(KafkaProducerRecord.create(topic, event))
      .mapEmpty();
  }

  private static boolean maybeIsNotEvent(final HttpResponse<Buffer> response) {
    // This checks whether there is something in the body or not, though binary events can contain only headers and they
    // are valid Cloud Events.
    return response == null || response.body() == null || response.body().length() <= 0;
  }

  @Override
  public Future<Void> close() {
    return CompositeFuture.all(
      this.producer.close(),
      AsyncCloseable.wrapAutoCloseable(producerMeterBinder).close()
    ).mapEmpty();
  }
}
