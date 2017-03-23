Eywa
====


What is Eywa?
-------------

**"Eywa is the guiding force and deity of Pandora and the Na'vi. All living things on Pandora connect to Eywa."** -- [Avatar Wiki](http://james-camerons-avatar.wikia.com/wiki/Eywa)

**Project Eywa** is essentially a connection manager that keeps track of connected devices. But more than just connecting devices, it is also capable of designing flexible APIs, sending control commands to them, collecting data from them, real-time monitoring and debugging, in the end, providing query interfaces that can be used for data visualization.

Eywa lets the teams of embedded system developers forget about reinventing the backend services and provides a commonly used protocol, websocket, to make real-time communication easily achievable.

Why is it useful?
-----------------

We are a group of people who are interested in **Home Automations** and **Smart Devices**. Often time, these projects involve connecting devices into cloud, tracking the usage of different functionalities, collecting the data and also controlling them. After worked on several similar projects, we found there is no reason to reinvent the wheel each time for different applications. So we came up with **Project Eywa** to help small teams like us reduce their development circles.

What features does it have?
---------------------------

Here is a growing list of features we want to support:

- [x] Connection Manager
- [x] Device Control
- [x] Command Line Tools
- [x] Connection Attach Mode
- [ ] Admin Panel
- [x] Basic Authentication
- [x] SSL protection
- [x] Data Indexing
- [ ] Data Streaming
- [x] Data Export
- [x] Data Retention
- [ ] Data Visualization
- [x] Query Interface
- [ ] Clustering
- [ ] Custom Web hooks
- [ ] Custom Monitors
- [ ] M2M (machine to machine) communication
- [x] HTTP Long-Polling
- [x] Websocket
- [ ] MQTT integration
- [x] Dockerized image

Our Admin Panel and Data Visualization dashboard will be releasing soon.

Please let us know if you want more features by creating issues. Pull requests are also very much welcome!

Performance
-----------

How performant is Eywa? Well, we did a simple benchmark and the [benchmark script](https://github.com/xcodersun/eywa/blob/master/benchmark/benchmark.go) is also available in the repo under `benchmark` directory.

The latest benchmark shows, on a machine of 12 CPUs + 32GB mem from [Digital Ocean](https://www.digitalocean.com/). A single Eywa node can keep track of more than 1.5 million devices, with a lot of potential. CPU is merely used, but memory is the limiting factor.

For more details please check out wiki on [Performance](https://github.com/xcodersun/eywa/wiki/Performance).

How to use?
-----------

You can get started with our [wiki](https://github.com/xcodersun/eywa/wiki).


Community / Contributing
------------------------

Eywa maintains a forum [goeywa](https://groups.google.com/forum/#!forum/goeywa), where you should feel free to ask questions, request features, or to announce projects that are built with Eywa. You should also see updates and road maps on Eywa in this forum.

Contributions to Eywa are very much welcomed. Fork us if you would like to.
