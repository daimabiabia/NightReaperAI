//go:build windows

package security

import (
	"context"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

const createNoWindow = 0x08000000 // syscall.CREATE_NO_WINDOW (not exported by syscall)

func prepareShellCmdSession(cmd *exec.Cmd) error {
	if cmd == nil {
		return nil
	}
	// 独立进程组，便于 taskkill /T 终止整棵子进程树。
	// CREATE_NO_WINDOW：子进程拿全新隐藏控制台，不继承服务器控制台——
	// 真实 conhost 下经代父进程（PROC_THREAD_ATTRIBUTE_PARENT_PROCESS）
	// 派生的子进程会因控制台初始化失败而以 0xc0000142 退出；任务输出
	// 本就经管道捕获，隐藏控制台同时避免任务输出刷进服务器窗口。
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CreationFlags = syscall.CREATE_NEW_PROCESS_GROUP | createNoWindow
	return nil
}

// terminateProcessGroup 使用 taskkill /F /T 终止进程及其子进程；rootPID 为 0 时回退到 cmd.Process.Pid。
func terminateProcessGroup(rootPID int, cmd *exec.Cmd) {
	pid := rootPID
	if pid <= 0 && cmd != nil && cmd.Process != nil {
		pid = cmd.Process.Pid
	}
	if pid <= 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	tk := exec.CommandContext(ctx, "taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
	if err := tk.Run(); err != nil {
		if cmd != nil && cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}
}

// terminateCmdTree 使用 taskkill /F /T 终止进程及其子进程（Windows 上 Process.Kill 无法保证杀掉 python 等孙进程）。
func terminateCmdTree(cmd *exec.Cmd) {
	terminateProcessGroup(0, cmd)
}

func stopProcessGroup(pid int, cmd *exec.Cmd) {
	// Windows has no portable SIGTERM equivalent for arbitrary console jobs.
	terminateProcessGroup(pid, cmd)
}

// Windows taskkill /T is best effort; unlike a Unix PGID it has no persistent
// group handle to query after the root exits. Job Objects are needed for that.
func processGroupExists(pid int) bool { return false }
