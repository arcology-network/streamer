### 2.1. State Machine



In Streamer, the subscribers by subscribing to the specific names on the event broker to get the expected events.
Because of the nature of asynchronous communication,  the event broker cannot guarantee the events that the actor subscribes 
to will arrive in the desired order. Some actors do rely on a specific temporal event sequence to work properly. 

For instance, an actor subscribe to 2 event A and B to perform its tasks. It becomes ready for event B only after receiving and processing 
event A. Similarly, after processing event B, it returns to a state ready for event A, and this cycle continues.

#### 2.1.3. Solutions

One approach is to buffer all the events until all the events are received, put them in order, and then send these events to the actor to process.
The approach is simple and straightforward, but there are some performance issues. The worker thread sits idle until all the events have been received. 
The time wasted on waiting for all the message to arrive could be spent on processing them instead when partial sequence events have become available. 

#### 2.1.4. What Does the FSM Do
The state machine is responsible for controlling the state transitions of the actor. At initialization, the actor registers with the state machine and inform it about the message info and order in which they need to be received. The state machine will only forward an event to the actor if it is received in right order or buffer the event otherwise. The actor will notify the state machine about the next event it is ready to receive.
