# streamer

Streamer is an internally developed Event-Driven Architecture (EDA) framework by Arcology. It offers a unified solution for seamless inter-thread and inter-process communication. 

## Key Features
- Unified Framework: Abstracts event broker logic into a unified framework, eliminating the need for multiple brokers.

- Built-in Filters: Manages complex event flows efficiently with built-in event filters.

- 3rd-party Integration: Seamlessly integrates with Apache Kafka for high-throughput inter-process communication.

- Deployment Flexibility: Separates implementation from deployment, emphasizing flexibility.

## Use Cases

- Scalable EDA: Suited for high-scalability event-driven architecture.

- Distributed Streaming: Ideal for distributed streaming with scalability, durability, and fault tolerance.

## Components
- Event Broker: Manages event flow to and from actors.

- Actors: Entities reacting to broker events, comprising a State Machine, Event Filter, and Worker Thread.

## Workflow

-Event Broker: Central hub managing and distributing events.

- Actors: State Machine controls state transitions, Event Filter governs event relationships, and Worker Thread processes custom logic.

Explore the technical power of Streamer for efficient event-driven architecture with scalability and flexibility. Happy coding!


