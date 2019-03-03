trees
=====

Trees is an experimental project management web application. It allows users to model
projects in trees of tasks, each parent task node is an summary of all it's descendants
showing total remaining time to completion, total logged time on the tasks and minimum
remaining time to completion. total and minimum remaining time on a parent task node differ
when the parent task node is set to a parallel task, meaning all of it's child tasks can be
worked on in parallel.

The client side is written as a Vuejs based progressive web app and the server side is written
in Go.

## Prerequisites

* Go must be installed and configured to compile the server
* node/npm must be installed to build and run the client side dev mode
* docker and docker-compose are used to run 3rd party services, specifically:
  * redis - for caching
  * mariadb - for data storage
  
## Setup

redis and mariadb are setup in their default configs to use ports 6379 and 3306, the api server and
client dev server are setup to use port 80, if you have binding conflicts feel free to re config them.

update hosts file to include the line:
```shell
127.0.0.2 lcl.project-trees.com
```

pull repo:
```shell
go get github.com/0xor1/trees
cd $GOPATH/src/github.com/0xor1/trees
```

run third party services:
```shell
docker-compose -f meta/docker-compose.yml up
```

init mariadb schemas:
```shell
go run meta/init_db.go
```

run server:
```shell
cd server
go run main.go
```

run client dev server:
```shell
cd client
npm i
npm run dev
```

your default web browser should automatically open the page `http://lcl.project-trees.com`,
click the register button and fill out the form, after submitting the form, look in the
terminal where the api server is running for the activation link to go to to get access
to the system.

The dev environment setup is here https://dev.project-trees.com
staging and production environments to come.

# System Features

* Auto api documentation - each endpoint automatically generates its own docs and they are published
at [/api/docs](https://dev.project-trees.com/api/docs)

* Multi endpoint calls - due to the strict format of endpoints it is possible to make a generic means of
calling multiple endpoints in a single request, this is done via the `/api/mdo` endpoint. It can be seen in
use on the client side when loading a task node:

```ecmascript 6
let params = router.currentRoute.params
let mapi = api.newMDoApi(params.region)
mapi.v1.project.get(params.region, params.shard, params.account, params.project).then((project) => {
  this.project = project
  if (params.project === params.task) {
    this.task = project
    this.task.isAbstract = true
  }
})
if (params.project !== params.task) {
  mapi.v1.task.get(params.region, params.shard, params.account, params.project, params.task).then((task) => {
    this.task = task
  })
}
mapi.v1.task.getAncestors(params.region, params.shard, params.account, params.project, params.task, 100).then((res) => {
  this.ancestors = res.ancestors
  this.moreAncestors = res.more
})
mapi.v1.task.getChildren(params.region, params.shard, params.account, params.project, params.task, null, 100).then((res) => {
  this.children = res.children
  this.moreChildren = res.more
})
mapi.v1.timeLog.get(params.region, params.shard, params.account, params.project, params.task, null, null, true, null, 100).then((res) => {
  this.timeLogs = res.timeLogs
  this.moreTimeLogs = res.more
})
mapi.sendMDo().then(() => {
  this.loading = false
})
```
each endpoint call looks and feels just like a typical request with a promise for handling the response or error,
but all the calls are wrapped between:
```ecmascript 6
let mapi = api.newMDoApi(params.region)
// use mapi instance to bundle multiple requests into a single request 
// mapi.v1.section.endpoint1 ...
// mapi.v1.section.endpoint2 ...
// mapi.v1.section.endpoint3 ...
mapi.sendMDo().then(() => {
    //this promise code is executed after all the individual requests promises have been resolved
})
```

* Multi region support - users can choose to host their project data close to them for faster access, the
system is setup to be able to easily add new regions in different data centers around the world.

* Local, development, staging, production environments - the system was designed to be flexible and easy for
developers to setup on their local development machines without having to configure and run many different services,
all third party services are run in docker containers with a docker-compose file for easy setup and tear down, and the
api server can easily be configured to run as a single executable with all the endpoints available on it, or be split
for production mode where the central directory server and regional servers have different sets of endpoints available.

## Tests

the server side tests are system tests, they require that the redis and mariadb containers are running,
simply run:
```shell
go test -coverprofile cover.out ./...
```
