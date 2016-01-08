### What is Octopus

Octopus is essentially a connection manager that keeps track of the connected devices. But more than just connecting devices, it is also capable of controlling them, collecting the metrics sent from them, indexing the data, and providing query interfaces that can be used for data visualizations. Octopus lets the team of embedded system developers to forget about reinventing the backend services and provides a commonly used protocol, websocket, to make real-time control easily achievable.

***
### Basic Concepts

#### Channel

Channel is a logical group of connections. They should share the same authentication keys, similar data schema,
and the same handlers of their messages.

Here is an example of defining a channel for monitoring hotel room sensors:

```
{
  "name": "hotel monitoring",
  "description": "This is a channel created for monitoring hotel.",
  "tags": ["building", "floor", "room", "room_type"],
  "fields": {
    "brightness": "int",
    "co2": "float",
    "cooler": "boolean",
    "humidity": "int",
    "noise": "float",
    "pm2_5": "int",
    "temperature": "float"
  },
  "message_handlers": ["indexer"],
  "access_tokens": ["abcdefg"]
}
```

A channel typically comes with a name and description. But more importantly, it needs to know what data is expected from the sensors, which sensor is it, how should uploaded data be treated, and also the access_tokens for authenticating the sensor connection.

In this channel definition, **fields** are the expected data their types. **Tags** are more like meta data of the actual field data. It tells which room does this data come from, what kinda of room type it is, etc. This tags are useful when you are interested in seeing data visualizations. They can be used to graph out the correlation between data and room types.

A channel can have multiple **access_tokens&&, and any one can be used at the time setting up connection.

Octopus provides middlewares as **message handlers**. So far only the `indexer` is provided. You can leave the message_handler as an empty array, which means, Octopus will solely just be your connection manager and won't try to index your data, and so that the data won't be queryable.

More message handlers will be supported soon, including **web hooks**, **alerters**, etc. Please check out [road map](https://github.com/vivowares/octopus/wiki/Road-Map) for more details.

#### Dashboard

A dashboard is a collection of data visualizations. You can create a dashboard and in there define the graphs that you want to see. Octopus provides a easy-to-use interface to make sense of your data. And tutorial of graphing out of them is also available! Check out our [online demo](#) and play with it!

***
### Components

#### Connection Manager

Currently Octopus is a single node connection manager, we've put clustering for scalability into our road map and it will be our milestone for the early 2016. Stay tuned for that!

#### Indexing Engine

Other than the backend services itself, Octopus uses [Elasticsearch](https://www.elastic.co/products/elasticsearch) as indexing engine. We've heard of requests of using other NoSql database such as [influxdb](https://influxdata.com/) or [mongodb](https://www.mongodb.org/). Supporting them will be discussed but Elasticsearch will be more than capable for most of the requirements.

#### Data Visualization

We provide a default front-end application for easily access to the data. It is called [overlook](#), and is a subproject under Octopus. Checkout the [demo site](#) and you will get a sense of it.

***
### Features

Here is a complete list of supported features. And more will be supported.

- [x] Connection Manager
- [x] Device Control
- [x] Basic Authentication
- [x] SSL protection
- [x] Data Indexing
- [x] Data Export
- [x] Data Retention
- [x] Data Visualization
- [x] Query Interface
- [ ] Clustering
- [ ] Custom Web hooks
- [ ] Custom Alerters

Please let us know if you want more features by creating issues. Pull requests are also very much welcome!

***
### Performance

How reliable is Octopus? Well we did a simple benchmark and the benchmark script is also available in the repo.
On a very basic setup: 1 CPU + 1GB mem on [Digital Ocean](https://www.digitalocean.com/). A single Octopus node can keep track of more than 15k devices easily. More detail please hit our [blog post](#).
