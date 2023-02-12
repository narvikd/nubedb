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

For example, to set a value for a key, you can send a `POST` request to the API:
<img width="1920" alt="Set Value" src="https://user-images.githubusercontent.com/84069271/218323894-50175fd1-2fa7-4c24-9b0f-c6e0feedb759.png">


Please note that this can only be done in the `Leader` node. If you attempt to do it in a `Follower` you will get:
<img width="1920" alt="Non Leader trying to store" src="https://user-images.githubusercontent.com/84069271/218324097-8f8a3927-0a55-40db-b50a-16008d10aa69.png">


To retrieve a value for a key, you can send a `GET` request to any of the nodes, it doesn't need to be a `Leader`:
<img width="1920" alt="Get a key on leader" src="https://user-images.githubusercontent.com/84069271/218324205-d8d9cbd2-5248-499a-b95f-38ad5eba55c1.png">


<img width="1920" alt="Get a key on follower" src="https://user-images.githubusercontent.com/84069271/218324242-a4e66d74-19b4-4148-9322-6b3eaf4d83c7.png">



If it doesn't exist:
<img width="1920" alt="Key doesn't exist" src="https://user-images.githubusercontent.com/84069271/218324306-9bb85f07-01e9-49f6-b848-00b5458343fb.png">

