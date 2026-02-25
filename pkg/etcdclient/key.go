package etcdclient

const (
	keyEtcdPrefix = "/taskGo/"

	// Register Node
	// key /taskGo/node/<node_uuid>
	KeyEtcdNodePrefix = keyEtcdPrefix + "node/"
	KeyEtcdNodeFormat = KeyEtcdNodePrefix + "%s"

	// Register Process
	// key /taskGo/proc/<node_uuid>/<task_id>/<pid>
	KeyEtcdProcPrefix = keyEtcdPrefix + "proc/"
	KeyEtcdNodeProcPrefix = KeyEtcdProcPrefix + "%s/"
	KeyEtcdTaskProcPrefix = KeyEtcdNodeProcPrefix + "%d/"
	KeyEtcdProcFormat	  = KeyEtcdTaskProcPrefix + "%d"

	// Register Task On Node
	// key /taskGo/task/<node_uuid>/<task_id>
	KeyEtcdTaskPrefix 	= keyEtcdPrefix + "task/%s/"
	KeyEtcdTaskFormat	= KeyEtcdTaskPrefix + "%d"

	// Register Task
	KeyEtcdOncePrefix = keyEtcdPrefix + "once/"
	KeyEtcdOnceFormat = KeyEtcdOncePrefix + "%d"

	// Distributed Lock
	KeyEtcdLockPrefix = keyEtcdPrefix + "lock/"
	KeyEtcdLockFormat = KeyEtcdLockPrefix + "%s"

	// Register Node On System
	KeyEtcdSystemPrefix = keyEtcdPrefix + "system/"
	KeyEtcdSystemSwitch = KeyEtcdSystemPrefix + "switch/" + "%s"
	KeyEtcdSystemGet	= KeyEtcdSystemPrefix + "get/" + "%s"
)





