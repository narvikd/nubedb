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
##### State
Check consensus state:
<img width="1920" src="https://user-images.githubusercontent.com/84069271/219970339-a24c7be6-474a-4837-9ab0-f96d8fec3d19.png">

##### Healthcheck
Check consensus health:
<img width="1920" src="https://user-images.githubusercontent.com/84069271/219970383-13b308ee-2c97-4850-bfdd-66793dfbd036.png">


#### Database
##### Store
To store a value for a key, you can send a `POST` request to the API:
<img width="1920" src="https://user-images.githubusercontent.com/84069271/219970407-db100714-4304-4a9d-99fb-3b0cd9ec4f32.png">


##### Get
To retrieve a value for a key, you can send a `GET` to the API:
<img width="1920" src="https://user-images.githubusercontent.com/84069271/219970431-33cd0df3-fa48-442c-8946-e71f3b8ddab2.png">


If it doesn't exist:
<img width="1920" src="https://user-images.githubusercontent.com/84069271/219970451-db5c0173-9e61-4fbe-b139-1544a97afcfe.png">


##### Delete
To delete a key, you can send a `DELETE` request to the API:
<img width="1920" src="https://user-images.githubusercontent.com/84069271/219970470-3928d7e6-be00-405e-b3a0-e8c1fd999a7d.png">


If it doesn't exist:
<img width="1920" src="https://user-images.githubusercontent.com/84069271/219970490-d300b229-cef1-4a0b-9346-f9f900eddb24.png">
