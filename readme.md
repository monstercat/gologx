# logx

Provides a logging utility which provides context(s) to an error message. 

Suppose you have a function `a` that calls another function `b` with an error. 
```
func a() {
   ctx := logx.New("a")
   ctx.Wrap( b() )
   ctx.Warn("hello")

   fmt.Printf("%#v", ctx.Map() )
}

func b() error {
  ctx := logx.New("b")
  ctx.Warn("some warning")
  return ctx.Fatal("fatal error")
}
```

It would print a map:
```
{
    "Name": "a",
    "Context": null,
    "Logs": [
        {
            "Name": "b",
            "Context": null,
            "Logs": [
                "some warning",
                "fatal error"
            ]
        },
        "hello"
    ]
}

```