# logx

Custom logging solution, and associated server for remote consolidation. 

The basic premise is to allow logging not only to std output or local file, but also to a remote host which is queryable
by postgres. 

For examples, please see [HostHandlerTest](logxhost/handler-host_test.go) and [GenericHandlerTest](handler_test.go)

Initialization
---
A `LogHandler` is required and is used to handle logging for the system. 

It requires a list of handlers. For example: 

```go

hostHandler := &logx.HostHandler{
    ...
}

logHandler := &LogHandler{
    Handlers: []Handler{   
        hostHandler,
        logx.StdHandler,
    },   
}

errCh := make(chan error)
go func() {
    for err := range errCh {
        //handle error
    }
}()

go hostHandler.RunForever(errCh)

```

Use in Routes
---
To use in a route, you can create a logger using the `NewRouteLogger` method. 

For example: 
```
http.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request){
    customRouteHandler(w, r, NewRouteLogger(r, logHandler))
})

...

func customRouteHandler(w http.ResponseWriter, r *http.Request, log *log.Logger){
    ...
    // no changes to the way log is used by default. 
}
```
