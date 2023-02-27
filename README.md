# NubeDB

NubeDB is a simple distributed key-value database that uses the Raft consensus algorithm for data replication and consistency.


### Features
* Distributed and replicated data storage
* Simple to use
* Automatic fail-over and recovery if any of the nodes go down.
  * Limitations:
    * The system needs 1 leader and 2 nodes to operate correctly.
    * It can go down as far as 1 leader and 1 node, but there are no warranties that it will be stable.
* TODO:
  * True auto-scaling, in which a number of arbitrary nodes can be removed when the whole cluster is down, and the system can still reach a quorum on startup.


### Table of Contents  
* [Getting started](#getting-started)
  * [Starting a cluster](#starting-a-cluster)
  * [Using the API](#using-the-api)
    * [Consensus](#consensus)
    * [Database](#database)
      * [Store](#store)
      * [Get](#get)
      * [Delete](#delete)
      * [Backup](#backup)
      * [Restore](#restore)


## Getting started

#### Starting a cluster
You can edit the ``docker-compose.yml`` to enable performance mode at the cost of having larger DB files.

This is done by uncommenting :
```bash
environment:
  - FSM_PERFORMANCE=true
```
To start the cluster:
```bash
docker-compose up -d
```

#### Using the API
NubeDB provides a simple REST API for accessing its k/v database. You can interact with it using any HTTP client.

#### Consensus
##### State
To check the consensus state, you can send a `GET` request to `consensus`:
<img width="1920" src="https://user-images.githubusercontent.com/84069271/219970339-a24c7be6-474a-4837-9ab0-f96d8fec3d19.png">

##### Healthcheck
To check the consensus health, you can send a `GET` request to `health`. It will return only a status code for simplicity:
<img width="1920" src="https://user-images.githubusercontent.com/84069271/219970383-13b308ee-2c97-4850-bfdd-66793dfbd036.png">


#### Database
##### Store
To store a value for a key, you can send a `POST` request to `store`:
<img width="1920" src="https://user-images.githubusercontent.com/84069271/219970407-db100714-4304-4a9d-99fb-3b0cd9ec4f32.png">


##### Get
To retrieve a value for a key, you can send a `GET` request to `store`:
<img width="1920" src="https://user-images.githubusercontent.com/84069271/219970431-33cd0df3-fa48-442c-8946-e71f3b8ddab2.png">

##### GetKeys
To retrieve all keys in the DB, you can send a `GET` request to `store/keys`:
<img width="1920" src="https://user-images.githubusercontent.com/84069271/221429650-ce774f1d-c8d1-4525-88a1-6420c69c67e2.png">


##### Delete
To delete a key, you can send a `DELETE` request to `store`:
<img width="1920" src="https://user-images.githubusercontent.com/84069271/219970470-3928d7e6-be00-405e-b3a0-e8c1fd999a7d.png">

##### Backup
To get a full backup of the DB, you can visit or send a `GET` request to `store/backup`:
<img width="1920" src="https://user-images.githubusercontent.com/84069271/221430304-6f109e26-be8c-4870-ba59-061d99d4b632.png">


##### Restore
To restore a backup of the DB, you can send a `POST` request to `store/restore`:

First ensure you are sending a `form-data` with a key `backup` as a `file`:
<img width="1920" src="https://user-images.githubusercontent.com/84069271/221430396-baf0978f-b40a-4e8e-8e48-c096917ca9f7.png">


Example:
<img width="1920" src="https://user-images.githubusercontent.com/84069271/221430404-51369a6c-e99d-40a1-bbae-56514b1d4c1b.png">
