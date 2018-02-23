package lib

type Task struct {
	name  string
	size  float64
	flops float64
	data  interface{}
}

func NewTask(name string, flops float64, size float64, data interface{}) *Task {
	t := &Task{
		name:  name,
		size:  size,
		flops: flops,
		data:  data,
	}
	return t
}

func (task *Task) GetName() string {
	return task.name
}

func (task *Task) GetSize() float64 {
	return task.size
}

func (task *Task) GetFlops() float64 {
	return task.flops
}

func (task *Task) GetData() interface{} {
	return task.data
}
