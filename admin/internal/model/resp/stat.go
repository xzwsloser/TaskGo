package resp

type (
	RspSystemStatisics struct {
		NormalNodeCount			int64	`json:"normal_node_count"`
		FailNodeCount			int64	`json:"fail_node_count"`
		TaskExcSuccessCount		int64	`json:"task_exc_success_count"`
		TaskRunningCount		int64	`json:"task_running_count"`
		TaskExcFailCount		int64	`json:"task_exc_fail_count"`
	}

	RspDateCount struct {
		Date	string	`json:"date"`
		Count	string	`json:"count"`
	}

	RspDateCountSet struct {
		SuccessDateCount	[]RspDateCount	`json:"success_date_count"`
		FailDateCount		[]RspDateCount	`json:"fail_date_count"`
	}
)




