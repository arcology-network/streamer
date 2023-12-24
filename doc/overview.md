# Streamer

Streamer is an advanced, flexible, loose coupling, highly scalable, and in-house developed event-driven architecture (EDA) development framework by Arcology. Streamer serves as a unified solution for inter-thread and inter-process communication, allowing seamless data exchange between modules regardless of deployment method or location. It is designed to ensure that modules always interact with a unified event broker, regardless of the underlying communication mechanisms, whether inter-process or inter-thread. Streamer is the foundation of the Arcology Network's client software.


## Why Streamer

Compare to other similar systems, Streamer 

1. Abstracts underlying event broker logic into a single, unified framework, eliminating the need for developers to work with different event brokers.

2. Comes with some built-in event filters to manage complex event flows effectily.

3. Seamlessly integrates with 3rd-party systems like Apache Kafka for high-throughput inter-process communication.

4. Separates implementation from deployment, allowing developers to focus on logic without worrying about how deployment may affect throughout.

5. Provides lockless communication.
 

### Streamer vs Kafka

|              | Streamer                                           | Kafka                                              |
|---------------------------------|---------------------------------------------------|----------------------------------------------------|
| **Abstraction of Event Broker Logic** | Abstracts complexities into a unified system, minimizing platform-specific code. | Specialized as a distributed streaming platform.    |
| **Unified Event Management**        | Provides a unified approach for inter-thread and inter-process communication. | Optimized for distributed streaming scenarios.      |
| **Built-in Event Filtering Mechanism** | Offers a built-in mechanism for managing complex event flows efficiently. | Primarily used for message queueing with streaming capabilities. |
| **Deployment Flexibility**           | Supports deployment transparency, allowing easy transitions between environments. | May require more effort in configuration changes during deployment transitions. |
| **Integration with Kafka**           | Seamlessly integrates with Kafka for enhanced interoperability. | Positioned as a distributed streaming platform.      |
| **Scalability and Durability**       | Emphasizes abstraction and flexibility for deployment scenarios. | Known for scalability, durability, and fault-tolerant data streaming. |
| **Community Adoption**               | Developed as part of the Arcology ecosystem.         | Widely adopted with a strong community presence.     |
| **Use Cases**                       | Suited for unified event-driven architecture with seamless deployment transitions. | Ideal for distributed streaming with scalability, durability, and fault tolerance. |

## Components

There are two major entities in the whole design, event Broker and Actors.  

- **Broker:** [The event broker]() is responsible for managing the event flow to and from the actors.

- **Actors:** [An actor]() is a computational entity that react to the events subscribed from the broker.

The event broker receives events from publishers (actors) and distributes them to subscribers (actors) based on their interests.
In Streamer, an actors usually communicates asynchronously with each others. Actors can publish and subscribe to events through the event broker. 
