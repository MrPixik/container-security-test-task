# My Worker Pool
Тестовое задание для команды Kaspersky Container Security в Kaspersky Safeboard
## В рамках тестового задания было реализовано API для обработки задач с использованием worker pool
## 1. Worker Pool
Был реализован следующий интерфейс:
```go
// Pool represents a pool of workers.
type Pool interface {
    // Submit submits a Task to the pool.
    Submit(t Task, maxRetries int) (int, error)
    // Wait waits for all tasks in buffer to be completed.
    Wait()
    // Stop stops the pool and all its workers.
    Stop() error
    // GetTaskBufferSize returns the size of the Task buffer.
    GetTaskBufferSize() int
    // GetTasksState returns all states of the tasks.
    GetTasksState() []TaskState
    // GetTaskStateById returns states of the Task by it's id.
    GetTaskStateById(id int) (TaskState, error)
}
```
Так же была добавлена его реализация в лице структуры pool:
```go
// pool represents a pool of workers.
type pool struct {
    workers     []*worker
    workerStack chan int
    workersNum  int
    
    // taskItem are added to this channel first, then sends to workers.
    taskBuffer chan *taskItem
    tasksMu    sync.RWMutex
    // slice of pointers to all taskItems.
    tasks []*taskItem
    // Set by WithTaskBufferSize(), used to set the size of the taskBuffer. Default is 1e6.
    taskBufferSize int
    // Set by WithRetryCount(), used to retry a Task when it fails. Default is 0.
    maxRetries int
    // baseDelay sets base delay time for backoff. Default is 0
    baseDelay time.Duration
    // Set by WithResultCallback(), used to handle the result of a Task. Default is nil.
    resultCallback func(interface{})
    // Set by WithErrorCallback(), used to handle the error of a Task. Default is nil.
    errorCallback func(error)
    ctx           context.Context
    // cancel is used to cancel the context. It is called when Stop() is called.
    cancel   context.CancelFunc
    stopOnce sync.Once
    wg       *sync.WaitGroup
}
```
## Функционал pool
### 1. Задание размера очереди через Option
```go
p := pool.New(WorkersNum, pool.WithTaskBufferSize(queueSize))
```
### 2. Задание callback функции для обработки результата и ошибки по результатам выполнения task
```go
// Задаем Callback функцию через Option WithResultCallback()
pool := pool.New(workersNum, pool.WithResultCallback(func(i interface{}) {
    s, ok := i.(string)
    if ok {
        fmt.Println(s)
    }
}))
// Пример применения
var counter int64 = 0
task := func() (interface{}, error) {
    time.Sleep(time.Millisecond * 1)
    atomic.AddInt64(&counter, 1)
    return strconv.FormatInt(counter, 10), nil
}

for i := 0; i < taskNum; i++ {
    if i, err := pool.Submit(task, 0); err != nil {
        fmt.Printf("task index: %d\nerror: %w\n", i, err)
    }
}
pool.Wait()
```
пример вывода
```cmd
3
2
4
5
1
```
Аналогично и с callback для ошибок через Option WithErrorCallback()
### 3. Задание количества повторных попыток при возникновении ошибки в task
```go
p := pool.New(WorkersNum, pool.WithRetryCount(retryCount))
```
### 4. Задание начального delay для повторных попыток
```go
p := pool.New(WorkersNum, pool.WithBaseDelay(baseDelay))
```
при повторных попытках в воркерах применяется экспоненциальный backoff + jitter
```go
backoff := time.Duration(1<<i) * pool.baseDelay
jitter := time.Duration(rand.Int63n(int64(backoff)))
sleepTime := backoff + jitter

time.Sleep(sleepTime)
```
### 5. Реализована возможность отслеживания состояния выполняемых задач (queued, running, done, failed)
Осуществляется посредством применения методов GetTaskStateById() для отслеживания задания по айди,либо GetTasksState() для отслеживания состояния всех тасок
```go
pool := pool.New(workersNum)

task := func() (interface{}, error) {
    time.Sleep(time.Millisecond * 1)
    return nil, nil
}

start := time.Now()
for i := 0; i < taskNum; i++ {
    if i, err := pool.Submit(task); err != nil {
        fmt.Printf("task index: %d\nerror: %w\n", i, err)
    }
}
time.Sleep(time.Millisecond * 1)
if taskState, err := pool.GetTaskStateById(1); err == nil {
    fmt.Println(taskState)
} else {
    fmt.Println(err)
}

if taskState, err := pool.GetTaskStateById(1e7); err == nil {
    fmt.Println(taskState)
} else {
    fmt.Println(err)
}

pool.Wait()
```
пример вывода:
```cmd
running
no task with this index
```
### 6. Есть поддержка задания конфигурации через переменные окружения (но работает и без этого)
поддерживаемые переменные окружения:
```cmd
WORKERS_NUM - количество воркеров
QUEUE_SIZE - размер очереди
```

### 7. Весь основной функционал покрыт тестами, которые можно запустить, находясь в корневой директории, следующими командами
```bash
go test -v ./internal/pool
```
вывод должен быть примерно таким
```cmd
=== RUN   TestPoolSubmitAndExecute
--- PASS: TestPoolSubmitAndExecute (0.10s)
=== RUN   TestPoolQueueFull
--- PASS: TestPoolQueueFull (0.00s)
=== RUN   TestPoolResultCallback
--- PASS: TestPoolResultCallback (0.10s)
=== RUN   TestPoolErrorCallback
=== RUN   TestPoolTaskExecutionOrder
--- PASS: TestPoolTaskExecutionOrder (0.10s)
=== RUN   TestTaskStates
--- PASS: TestTaskStates (1.10s)
ok      sandbox-dev-test-task/internal/pool     2.915s
```
## 2. Server
###  POST /enqueue
Ручка для передачи заданий в worker pool.\
Тело: JSON 
```json
{"id":"<string>","payload":"<string>","max_retries":"<int>"}
```
Ручка симулирует обработку задач. Каждая задача выполняется 100-500мс и "падает" в 20% случаев.
###  GET /healthcheck
Ручка возвращает 200 OK в случае, если сервер жив.
### Мидлварь для логгирования через slog.
Логгирует входящие запросы в JSON\
Пример логов:
```json
{"time":"2025-09-13T19:51:35.7316915+03:00","level":"INFO","msg":"request completed","component":"middleware/logger","method":"POST","path":"/enqueue","remote_addr":"[::1]:12126","user_agent":"PostmanRuntime/7.43.3","status":201,"bytes":13,"time":0}
{"time":"2025-09-13T19:52:56.6550833+03:00","level":"INFO","msg":"request completed","component":"middleware/logger","method":"GET","path":"/healthcheck","remote_addr":"[::1]:12126","user_agent":"PostmanRuntime/7.43.3","status":200,"bytes":0,"time":0}

```
## 3. Поддержка graceful shutdown через SIGINT/SIGTERM

## 4. Тестирование
Ручки также покрыты тестами. Для запуска введите в корневой директории:
```bach
go test -v ./internal/http-server
```