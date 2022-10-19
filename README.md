# parallel

some commonly used go concurrency models

concurrent requests with timeout capability

## how to use

```go
tasks := make([]TaskFunc, 0, testSize)
tasks = append(tasks, func (ctx context.Context, i int) (Result, error) {
// do your things
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

## License
[MIT License][1]

[1]: http://opensource.org/licenses/MIT