package taskreg

import (
	"pegasus/mergesort"
	"pegasus/task"
)

var taskGens = make(map[string]task.TaskGenerator)

// TODO
func registerTask(kind string, gen task.TaskGenerator) {
	taskGens[kind] = gen
}

func GetTaskGenerator(kind string) task.TaskGenerator {
	gen, ok := taskGens[kind]
	if !ok {
		return nil
	}
	return gen
}

func init() {
	registerTask(mergesort.RANDINTS_TASK_KIND, mergesort.TaskGenRandInts)
}
