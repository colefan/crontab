package crontab

// Job 定时任务接口
type Job interface {
	// Run 任务执行函数
	Run()
}
