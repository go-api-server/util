package coroutine

import (
	"time"

	"wps.cn/lib/go/log"
)

type AsyncFunc func(arg interface{})

type AsyncTask struct {
	curCoroutineIndex uint
	coroutines        []*taskCoroutine
}

func NewAsyncTask(asyncTaskName string, coroutineSize int, bufferSize int) *AsyncTask {
	if coroutineSize <= 0 || bufferSize <= 0 {
		panic("NewAsyncTask error arg")
	}

	a := &AsyncTask{}
	a.curCoroutineIndex = 0
	a.coroutines = make([]*taskCoroutine, 0, coroutineSize)

	for i := 0; i < coroutineSize; i++ {
		a.coroutines = append(a.coroutines, newTaskCoroutine(asyncTaskName, bufferSize))
	}

	return a
}

func (self *AsyncTask) Start() {
	for i := 0; i < len(self.coroutines); i++ {
		go self.coroutines[i].Run()
	}
}

func (self *AsyncTask) AddTask(fun AsyncFunc, arg interface{}) {
	self.AddTaskEx(fun, arg, 100)
}

func (self *AsyncTask) Stop() {
	for _, c := range self.coroutines {
		c.Stop()
	}
	self.coroutines = nil
}

/**
*  timeoutms : 0 队列满立即返回, -1 永远等待  > 0 等待到超时
 */
func (self *AsyncTask) AddTaskEx(fun AsyncFunc, arg interface{}, timeoutMs int64) bool {
	index := self.curCoroutineIndex
	self.curCoroutineIndex++
	index = index % uint(len(self.coroutines))

	return self.coroutines[index].AddTask(fun, arg, timeoutMs)
}

type Task struct {
	Func AsyncFunc
	Arg  interface{}
}

type taskCoroutine struct {
	tasks         chan *Task
	asyncTaskName string
	stop          chan struct{}
}

func newTaskCoroutine(asyncTaskName string, chanBuffSize int) *taskCoroutine {
	t := &taskCoroutine{}
	t.asyncTaskName = asyncTaskName
	t.tasks = make(chan *Task, chanBuffSize)
	t.stop = make(chan struct{}, 1)
	return t
}

func (self *taskCoroutine) callFunc(task *Task) {
	defer func() {
		recover()
	}()

	task.Func(task.Arg)
}

func (self *taskCoroutine) Stop() {
	close(self.tasks)
	<-self.stop
}

func (self *taskCoroutine) Run() {

	for true {
		task := <-self.tasks
		if task == nil {
			close(self.stop)
			return
		}

		self.callFunc(task)

		/*sz := len(self.tasks)
		if sz >= 100 {
			log.Warn("%s channel size to much, current size=%d", self.asyncTaskName, sz)
		}*/
	}
}

func (self *taskCoroutine) AddTask(fun AsyncFunc, arg interface{}, timeoutMs int64) bool {
	task := &Task{Func: fun, Arg: arg}
	if timeoutMs < 0 {
		self.tasks <- task
		return true

	} else if timeoutMs == 0 {
		select {
		case self.tasks <- task:
			return true
		default:
			return false
		}

	} else {
		select {
		case self.tasks <- task:
			return true
		case <-time.After(time.Millisecond * time.Duration(timeoutMs)):
			log.Error("%s taskCoroutine addTask timeout", self.asyncTaskName)
			return false
		}
	}
}
