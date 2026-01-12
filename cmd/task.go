package cmd

import (
	"fmt"
	"gin-api/internal/cron/tasks"
	"gin-api/internal/injector"
	"os"

	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "手动执行定时任务",
}

var runTaskCmd = &cobra.Command{
	Use:   "run [task_name]",
	Short: "执行指定任务",
	Run: func(cmd *cobra.Command, args []string) {
		taskName := args[0]

		// 初始化 DI 容器（与 API/Job 服务共享）
		container := injector.SetupInjector()
		defer container.Shutdown() // 自动关闭所有资源

		fmt.Printf("正在手动执行任务: %s\n", taskName)

		// 任务注册表
		taskRegistry := map[string]func() error{
			"test": func() error {
				task := tasks.NewExampleTask(container)
				task.Run()
				return nil
			},
		}
		// 执行任务
		runFunc, exists := taskRegistry[taskName]
		if !exists {
			_, _ = fmt.Fprintf(os.Stderr, "未知任务: %s\n可用任务: %v\n", taskName, getAvailableTasks(taskRegistry))
			os.Exit(1)
		}

		// 同步执行
		if err := runFunc(); err != nil {
			fmt.Fprintf(os.Stderr, "任务执行失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("任务 [%s] 执行完成\n", taskName)
	},
}

func getAvailableTasks(registry map[string]func() error) []string {
	var t []string
	for name := range registry {
		t = append(t, name)
	}
	return t
}
