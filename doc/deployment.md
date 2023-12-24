## Deployment Options

Stream is location transparent, meaning all the worker threads can communicate seamlessly regardless of their physical locations, whether in different thread of a single process, in different processes on the same machine or distributed across a network.


### Deployment Options

when Deployed on a single machine, the system should be able to avoid all the steps and overhead associated with inter process communication whenever possible. Creating multiple processing on a single machine is still perfectly possible, but if the users opt to a more performance focus, they could deploy all the service as different threads in the same process and all these should be done only with the changes of a configuration file.

 or sacraficing performancec advantage of share memory can offer. 

The Streamer also comes with some extra functionalities like event sequence controlling and message wrapping for an worker thread to work more effectively.

Kafka is a general purpose 
Apache Kafka is an event streaming platform. What does that mean?
