## 1. Event Broker

In Arcology, the event broker is a pub/sub based middleware that mediates the communication of event messages between producers and consumers. The role is to manage the publication and subscription of events, allowing different components or services to communicate with each other without direct coupling.

>>They are the ones doing the most of heavy lifting. Each Streamer instance may own an arbitrary number of event handlers, but each event >>handler can only belong to one Streamer instance at a time.


### Event Distribution:

The Event Broker acts as a centralized hub that receives events from producers (components generating events) and distributes them to consumers (components interested in specific events).

In this model, producers publish events to specific channels or topics, and consumers subscribe to these channels to receive relevant events.

###  Loose Coupling:

The use of an Event Broker promotes loose coupling between components. Producers and consumers are not directly aware of each other. They interact through the Event Broker, allowing for greater flexibility and independence.

### Event filtering
allowing consumers to express interest in specific types of events. This ensures that consumers receive only the events relevant to their functionality.

Each Streamer instance may own an arbitrary number of event handlers, but each event handler can only belong to one Streamer instance at a time.

### Decoupling in Time and Space:

Event Brokers enable decoupling in both time and space. Producers and consumers can operate independently and may not be required to be active simultaneously or reside on the same physical or logical system.


### Scalability

EDA with an Event Broker supports scalability. New components can be added without direct changes to existing components. Each component can scale independently.
Reliability and Resilience:

Event Brokers can contribute to system reliability and resilience. Even if a component is temporarily unavailable, events can be buffered or retried, ensuring that they are eventually delivered.
Event Routing:

Some Event Brokers provide advanced features like event routing, allowing events to be directed to specific consumers based on predefined rules.

### Event Transformation:

Event Brokers may support event transformation, allowing events to be modified or enriched as they pass through the system

### 3.6. Router
If the recipients are under the Streamer instance, the instance will route the events directly to the target worker Threads based on their registration information. Otherwise, Streamer pass the events on to the inter-process broker to do the delivery.


### Example

These producers are configured with unique names, output events, and initial buffer lengths. The code snippet below demonstrates the interaction between the broker and a consumer. The example initiate a Streamer broker and a consumer. It then sets up and registers the event consumer with the broker.

#### MockActor

The MockActor in a simple event consumer with a single attribute c, which is a channel of the empty interface (chan interface{}). The fmt.Printf statements in the start method are used to print the received data in a formatted way for debugging or verification purposes. MockActor. There are two methods:

- **mockActor Type:**
Defines a type mockActor with a channel c for receiving data.

- **Consume Method:** Implements a method Consume that places data into the c channel.

- **Start Method:** Implements a method start that continuously receives data from c and prints formatted information about two states.

#### The Function

The "TestConsumer()" Function does the following:

1. Creates an event broker.

2. Instantiates a mockActor.

3. Creates a consumer named 'Consumer' with output events 'output-event-a' and 'output-event-b.' 

4. Registers the consumer with the broker.

5. Starts the broker with broker.Serve().

6. Launches a goroutine to execute actorC.start() concurrently (starts receiving data from the broker).

7. Sends two sets of events to the broker with delays for processing.

8. Introduces a pause using time.Sleep to allow for event processing.


The conjunction filter used in the step 8 basically inform the broker that the actor only triggers when both events are available; otherwise, the actor will cache the event first.

#### Code
``` go

type mockActor struct {
    c chan interface{}
}

func (mw1 *mockActor) Consume(data interface{}) {
    mw1.c <- data
}

func (mw1 *mockActor) start() {
    for v := range mw1.c {
        fmt.Printf("State1: %v, State2: %v\n", v.([]interface{})[0].(*mockState), v.([]interface{})[1].(*mockState))
    }
}

func TestStreamAggregator() {
	broker := NewStatefulBroker()

	actorC := mockActor{
		c: make(chan interface{}),
	}
	
    // Initialize a consumer with a conjunction filter
    consumer := NewDefaultConsumer("consumer", []string{"output-event-a", "output-event-b"}, NewConjunctions(&actorC))
	broker.RegisterConsumer(consumer) // Register the consumer with the broker

	broker.Serve() // Broker starts
	go actorC.start() // Start receiving data from the broker

	broker.Send("output-event-a", &mockState{"1"})
	broker.Send("output-event-b", &mockState{"2"})
	time.Sleep(time.Second)

	broker.Send("output-event-a", &mockState{"1"})
	broker.Send("output-event-b", &mockState{"2"})
	time.Sleep(time.Second)
}

```