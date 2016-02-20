### A Little Introduction

#### What is Eywa?

Eywa is essentially a connection manager that keeps track of the connected devices. But more than just connecting devices, it is also capable of controlling them, collecting the metrics sent from them, indexing the data, and providing query interfaces that can be used for data visualizations.

Eywa lets the team of embedded system developers to forget about reinventing the backend services and provides a commonly used protocol, websocket, to make real-time control easily achievable.

***
#### Why is it useful?

We are a group of people who are interested into **Home Automations** and **Smart Devices**. Often time, these projects involve connecting devices into cloud, track the usage of different functionalities, collecting the data and also controlling them. After worked on several similar projects, we found there is no reason to reinventing the wheel each time for different applications. So we came up this Project to help small teams to reduce the development their circle.

***
#### What features does it have?

Here is a complete list of supported features. And more will be supported.

- [x] Connection Manager
- [x] Device Control
- [x] Basic Authentication
- [x] SSL protection
- [x] Data Indexing
- [ ] Data Streaming
- [x] Data Export
- [x] Data Retention
- [x] Data Visualization
- [x] Query Interface
- [ ] Clustering
- [ ] Custom Web hooks
- [ ] Custom Alerters
- [ ] M2M (machine to machine) communication
- [ ] MQTT integration
- [x] Dockerized image

Please let us know if you want more features by creating issues. Pull requests are also very much welcome!

***
#### Performance

How reliable is Eywa? Well we did a simple benchmark and the benchmark script is also available in the repo under `tasks` directory.

On a very basic machine: 1 CPU + 1GB mem on [Digital Ocean](https://www.digitalocean.com/). A single Eywa node can keep track of more than 15k devices easily. For more detail please check out [Performance]
(https://github.com/vivowares/eywa/wiki/Performance).

***
#### Usages

Please check out our project [wiki](https://github.com/vivowares/eywa/wiki)
