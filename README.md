# Raft Example

A simple raft example based on [hashicorp/raft](https:///github.com/hashicorp/raft).

## How to run?

Then run 3 server with different program in different terminal tab:

```bash
$ SERVER_PORT=20001 RAFT_NODE_ID=node1 RAFT_PORT=10001 RAFT_VOL_DIR=/tmp/node_1_data go run main.go
$ SERVER_PORT=20002 RAFT_NODE_ID=node2 RAFT_PORT=10002 RAFT_VOL_DIR=/tmp/node_2_data go run main.go
$ SERVER_PORT=20003 RAFT_NODE_ID=node3 RAFT_PORT=10003 RAFT_VOL_DIR=/tmp/node_3_data go run main.go
```

## Creating clusters

After running the above command, we have 3 servers:

* http://localhost:20001 with raft server localhost:10001
* http://localhost:20002 with raft server localhost:10002
* http://localhost:20003 with raft server localhost:10003

We can check using `/raft/stats` for each server and see that all server initiated as Leader.

Now, manually pick one server as the real Leader, for example http://localhost:20001 with raft server localhost:10001.
Using Postman, we can register http://localhost:20002 as a Follower to http://localhost:20001 as a Leader.

```curl
curl --location --request POST 'localhost:20001/raft/join' \
--header 'Content-Type: application/json' \
--data-raw '{
	"node_id": "node_2", 
	"raft_address": "127.0.0.1:10002"
}'
```

And doing the same to register http://localhost:20003 as a Follower to http://localhost:20001 as a Leader:

```curl
curl --location --request POST 'localhost:20001/raft/join' \
--header 'Content-Type: application/json' \
--data-raw '{
	"node_id": "node_3", 
	"raft_address": "127.0.0.1:10003"
}'
```

> What happen when we do cURL?
>
> When we're running the cURL, we send the data of `node_id` and `raft_address` that being registered as a Voter.
> We say `Voter` because we don't know the real Leader yet.
>
>
> In server http://localhost:20001 it will add the configuration stating that http://localhost:20002 and http://localhost:20003
> now is a Voter.
> After add the Voter, raft will choose the server http://localhost:20001 as the Leader.
>
> Adding Voter must be done in Leader server, that's why we always send to the same server for adding server.
> You can see that we always call port 20001 both for adding port 20002 or 20003

Then, check each of this endpoint, it will return the status that the port 20001 is now the only leader and the other is just a follower:

* http://localhost:20001/raft/stats
* http://localhost:20002/raft/stats
* http://localhost:20003/raft/stats

Now, raft cluster already created!

## Store, Get and Delete Data

As already mentioned before, this cluster will create a simple distributed KV storage with eventual consistency in read.
This means, all writes command (Store and Delete) **must** be redirected to the Leader server, since the Leader server is the only one
that can do `Apply` in raft protocol. After doing Store and Delete, we can make sure that the Raft already committed the message to all Follower servers.

Then, in `Get` method in order to fetch data, we can use the internal database instead calling `raft.Apply`.
This makes all Get command can be targeted to any server, not only the Leader.

So, why we call it _eventual consistency in read_ while we can make sure that every after Store and Delete response returned it means that the raft already applied the logs to n quorum servers?

That is because while reading data directly in badgerDB we only use read transaction. From BadgerDB's Readme:

> You cannot perform any writes or deletes within this transaction. Badger ensures that you get a consistent view of the database within this closure. Any writes that happen elsewhere after the transaction has started, will not be seen by calls made within the closure.

To do store data, use this cURL:

```curl
curl --location --request POST 'localhost:20001/store' \
--header 'Content-Type: application/json' \
--data-raw '{
	"key": "hello",
	"value": "world"
}'
```

To get data (`localhost:20001` can be replaced with `localhost:20002` or `localhost:20003`):

```curl
curl --location --request GET 'localhost:20001/store/hello'
```

To delete data, use this:

```curl
curl --location --request DELETE 'localhost:20001/store/hello'
```
