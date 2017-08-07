gosser
====
go server send events react

<!-- TOC -->

- [build app](#build-app)
    - [build backend](#build-backend)
    - [build frontend](#build-frontend)
- [run app](#run-app)
    - [run backend](#run-backend)
    - [run frontend](#run-frontend)
- [resources:](#resources)

<!-- /TOC -->

# build app

## build backend

```
>go build ./backend/*
```

## build frontend

````
>cd app
>yarn
````

# run app

## run backend

```
>./main
```

## run frontend

```
>cd app
>yarn start
```

# resources:

* [Stream Updates with Server-Sent Events](https://www.html5rocks.com/en/tutorials/eventsource/basics/)
* [Go sse example](https://github.com/silalahi/go-sse/blob/master/example/server.go)
* [Go sse example](https://github.com/kljensen/golang-html5-sse-example/blob/master/server.go)
* [Writing a Server Sent Events server in Go](https://robots.thoughtbot.com/writing-a-server-sent-events-server-in-go)
* [Using server-sent events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events)
* [addEventListener](https://developer.mozilla.org/en-US/docs/Web/API/EventTarget/addEventListener)
* [server-sent events sse](https://hpbn.co/server-sent-events-sse/)
