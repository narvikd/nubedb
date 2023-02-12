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

Please note that they will all start as a `Leader` and not know about each other, with the API we will connect them between each other.

<img width="1920" alt="Startup" src="https://user-images.githubusercontent.com/84069271/218325196-710e9074-ffe3-49e4-8aaf-61c3d2deeb19.png">


#### Using the API
NubeDB provides a simple REST API for accessing its k/v database. You can interact with it using any HTTP client.

#### Consensus
Join `node 2` to `node 1`:
<img width="1920" alt="Node2 joins" src="https://user-images.githubusercontent.com/84069271/218324925-e6648532-2aba-433f-a72a-a0fc28a285ee.png">


Join `node 3` to `node 1`:
<img width="1920" alt="Node3 joins" src="https://user-images.githubusercontent.com/84069271/218325412-82085884-c203-4d8b-bca2-4e13a8866656.png">


Consensus changes are reflected on the console, now the nodes are no longer `Leader` but `Follower`:
<img width="1920" alt="Nodes join console" src="https://user-images.githubusercontent.com/84069271/218324983-1cfe1fec-388b-4437-8d7b-d56566627a09.png">


Check consensus state:
<img width="1920" alt="Consensus state" src="https://user-images.githubusercontent.com/84069271/218325055-92af693a-fe39-48c8-a42a-dacdc10ad5f0.png">


#### Database
To set a value for a key, you can send a `POST` request to the API:
<img width="1920" alt="Set Value" src="https://user-images.githubusercontent.com/84069271/218323894-50175fd1-2fa7-4c24-9b0f-c6e0feedb759.png">


Please note that this can only be done in the `Leader` node. If you attempt to do it in a `Follower` you will get:
<img width="1920" alt="Non Leader trying to store" src="https://user-images.githubusercontent.com/84069271/218324097-8f8a3927-0a55-40db-b50a-16008d10aa69.png">


To retrieve a value for a key, you can send a `GET` request to any of the nodes, it doesn't need to be a `Leader`:
<img width="1920" alt="Get a key on leader" src="https://user-images.githubusercontent.com/84069271/218324205-d8d9cbd2-5248-499a-b95f-38ad5eba55c1.png">


<img width="1920" alt="Get a key on follower" src="https://user-images.githubusercontent.com/84069271/218324242-a4e66d74-19b4-4148-9322-6b3eaf4d83c7.png">



If it doesn't exist:
<img width="1920" alt="Key doesn't exist" src="https://user-images.githubusercontent.com/84069271/218324306-9bb85f07-01e9-49f6-b848-00b5458343fb.png">


## Testing the system

Store a value:
<img width="1920" alt="Store value" src="https://user-images.githubusercontent.com/84069271/218323894-50175fd1-2fa7-4c24-9b0f-c6e0feedb759.png">

Get from a node:
<img width="1920" alt="Get from node" src="https://user-images.githubusercontent.com/84069271/218324242-a4e66d74-19b4-4148-9322-6b3eaf4d83c7.png">

Ensuring the 3 nodes are working:
<img width="1920" alt="3 nodes working" src="https://user-images.githubusercontent.com/84069271/218326585-db754e4b-e17b-4712-8a1c-6737bd585dbe.png">


Shutdown `Leader`. We check that a `Follower` gets `Leader` status:
<img width="1920" alt="Leader shutdown" src="https://user-images.githubusercontent.com/84069271/218326617-1d2b0e29-c88f-4682-b013-ed5ad211c94b.png">


Mutate key `hello` on new `Leader` on `node2`:
<img width="1920" alt="Mutate hello" src="https://user-images.githubusercontent.com/84069271/218326636-34c7d07f-5d2d-44e2-b985-eff09fad39ef.png">


Check if changes are reflected on the `Follower` on `node3`:
<img width="1920" alt="Node3 check" src="https://user-images.githubusercontent.com/84069271/218326653-1eae3355-0108-4d10-a343-f984558aa24f.png">


Startup old `Leader`. Check if it's now a `Follower`. Snapshot of the new data is downloaded:
<img width="1920" alt="Startup old leader" src="https://user-images.githubusercontent.com/84069271/218326677-b7f62686-cbfe-429e-b78a-a821b35231a6.png">


Check if the data on `node1` is the new data, and not the old one. The cluster doesn't have a [split brain](https://en.wikipedia.org/wiki/Split-brain_(computing)):
<img width="1920" alt="Get data from node1" src="https://user-images.githubusercontent.com/84069271/218326709-6cc64f86-8172-4b5b-a7b2-dd669933a999.png">


Check if old `Leader` on `node1` is recognized as a `Follower` in the consensus status:
<img width="1920" alt="Recognized as follower" src="https://user-images.githubusercontent.com/84069271/218326750-3de1a5da-a39c-4836-a526-71d2a6898401.png">
