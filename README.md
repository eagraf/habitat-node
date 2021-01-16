# Habitat Node
Server code for the Habitat network.

## Architecture

### Software Architecture
The software architecture roughly follows the [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html).
* Entities: The `entities` package contains common structs for basic concepts
* Use Cases: Appication logic is included in the various packages
* Interface Adapters: Major interfaces include that between the backnets and `fs`, as and between `fs` and the app layer.

### Process Architecture
* `state`: Each community has critical state that it needs to keep track of, including membership, backnet addresses, and running applications.
   This state is so important that all peers need to reliably agree on it. To do this, the `state` module communicates with peers in the various
   communities hosted on the node to reach consensus, using the [Paxos algorithm](https://lamport.azurewebsites.net/pubs/lamport-paxos.pdf).
* `orchestrator`: This process manages all other processes on the node, dynamically starting and stopping backnets, apps, etc as they are created
   and destroyed. The `orchestrator` module acts as a state machine on the critical state declared by the `state` module.
* `fs`: The filesystem module interfaces between applications requesting files, and the backnets hosting them. It manages permissions, file encryption,
   and handling the contingency of an unavailable backnet, according to [local-first principles](https://www.inkandswitch.com/local-first.html).
* `client` This module allows for the owners of the host machine to configure the node. Users are logged into the node through this module.
* `backnets`: Backnets, such as IPFS and DAT host peer-2-peer filesystems, which are accessed by the `fs` module through a backnet interface to provide
   data to apps.
* `apps`: Apps serve data to clients (on browsers for example), through standard methods like REST, GraphQL, etc. Web frontends for these apps are served
   via the gateways provided by backnets. That code then makes calls to an address for the app. Apps can be load balanced between nodes in a community.

