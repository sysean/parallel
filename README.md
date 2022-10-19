# parallel

some commonly used go concurrency models

concurrent requests with timeout capability

## how to use

```go
tasks := make([]TaskFunc, 0, testSize)
tasks = append(tasks, func (ctx context.Context, i int) (Result, error) {
    // do your things, like rpc call
})
wo := New(tasks)
result, err := wo.ParallelingWithTimeout(context.Background(), 5*time.Second)
```

## explain

there are 3 version about `ParallelingWithTimeout` function

- `ParallelingWithTimeout` original concurrency models
- `ParallelingWithTimeoutV2` use WaitGroup secure synchronization about goroutines
- `ParallelingWithTimeoutV3` use errgroup streamline the code

v3 is recommended

## design principle
1. If you want implement a function with timeout, do not use sync.WaitGroup.Wait on the main goroutine, because the function you want to implement needs to exit during timeout, If the main goroutine is synchronized using wait, main goroutine can not quit midway.
2. When you use channel, try to make sure, who send the data, who close it. If you make sure there are no one to send data, you can close it anywhere.
3. In fact, the `close` function is a very useful method, it can be used to tell multiple coroutines to end blocking, like the `cancelCtx` do.
4. Please ensure that all new gorutine can be exited, unless you have a special purpose

## license
[MIT License][1]

[1]: http://opensource.org/licenses/MIT