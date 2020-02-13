# logx

Custom logging solution, and associated server for remote consolidation. 

Initialization
---
A `LogHandler` is required and is used to handle logging for the system. 

It requires a list of handlers. For example: 

```go
logHandler := &LogHandler{
    Handlers: []Handler{   
    	&logx.DbHostHandler{
            ....,
        }, 
        logx.StdHandler,
    },   
}
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

Host Logger
---
TODO 

Server 
---
TODO 