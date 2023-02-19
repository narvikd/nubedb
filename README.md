# NubeDB

NubeDB is a simple distributed key-value database that uses the Raft consensus algorithm for data replication and consistency.


### Features
* Distributed and replicated data storage
* Simple to use
* Automatic failover and recovery if any of the nodes go down -> (nodes/2)+1 are the max number of nodes that can go down


### Table of Contents  
* [Getting started](#getting-started)
  * [Starting a cluster](#starting-a-cluster)
  * [Using the API](#using-the-api)
    * [Consensus](#consensus)
    * [Database](#database)
      * [Store](#store)
      * [Get](#get)
      * [Delete](#delete)


## Getting started

#### Starting a cluster
```bash
docker-compose up -d
```

#### Using the API
NubeDB provides a simple REST API for accessing its k/v database. You can interact with it using any HTTP client.

#### Consensus
Check consensus state:
<img width="1920" alt="Consensus state" src="https://user-images.githubusercontent.com/84069271/218325055-92af693a-fe39-48c8-a42a-dacdc10ad5f0.png">


#### Database
##### Store
To store a value for a key, you can send a `POST` request to the API:
<img width="1920" alt="Set Value" src="https://user-images.githubusercontent.com/84069271/218323894-50175fd1-2fa7-4c24-9b0f-c6e0feedb759.png">

##### Get
To retrieve a value for a key, you can send a `GET` to the API:
<img width="1920" alt="Get a key on leader" src="https://user-images.githubusercontent.com/84069271/218324205-d8d9cbd2-5248-499a-b95f-38ad5eba55c1.png">

If it doesn't exist:
<img width="1920" alt="Get key doesn't exist" src="https://user-images.githubusercontent.com/84069271/218324306-9bb85f07-01e9-49f6-b848-00b5458343fb.png">


##### Delete
To delete a key, you can send a `DELETE` request to the API:
<img width="1920" alt="Delete key" src="https://user-images.githubusercontent.com/84069271/218793049-5dc0669b-ba19-4415-8a53-4e409ca573cf.png">


If it doesn't exist:
<img width="1920" alt="Delete key doesn't exist" src="https://user-images.githubusercontent.com/84069271/218792895-8bf28fad-6dc0-45b9-9210-0aacfe460f4f.png">
