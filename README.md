# NubeDB

NubeDB is a simple distributed key-value database that uses the Raft consensus algorithm for data replication and consistency.

### Features
* Distributed and replicated data storage
* Simple to use
* Automatic failover and recovery if any of the nodes go down


## Getting started

#### Installation
You can install NubeDB by downloading the binary release for your platform from the releases page.

Alternatively, you can build NubeDB from its source, by cloning the repository.

#### Starting a cluster
To start a NubeDB cluster, you need to run a NubeDB node on each machine that you want it to be part of the cluster.

1. On the first machine, start a NubeDB node with the following command:
```bash
./nubedb --id node1 --host localhost --api-port 3001 --consensus-port 4001
```

2. On the second machine, start a NubeDB node with the following command:
```bash
./nubedb --id node2 --host localhost --api-port 3002 --consensus-port 4002
```

3. On the third machine, start a NubeDB node with the following command:
```bash
./nubedb --id node3 --host localhost --api-port 3003 --consensus-port 4003
```

This will start a 3 node NudeDB cluster, where each node is able to communicate with the others and participate in the consensus.

#### Using the API
NubeDB provides a simple REST API for accessing its k/v database. You can interact with it using any HTTP client.

For example, to set a value for a key, you can send a POST request to the API:
```bash
example here
```
Please note that this can only be done in the `Leader` server. If you attempt to do it in a `participant` you will get:
```bash
example here
```

And to retrieve a value for a key, you can send a GET request:
```bash
example here
```
If it doesn't exist:
```bash
example here
```
