//go:build windows

package agents

import (
	"os/exec"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type procGroup struct {
	job  windows.Handle
	info *windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
}

func procNewGroup(cmd *exec.Cmd) (*procGroup, error) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, err
	}
	group := &procGroup{job: job, info: &windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}}
	group.info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	_, err = windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(group.info)), uint32(unsafe.Sizeof(*group.info)))
	if err != nil {
		_ = windows.CloseHandle(job)
		return nil, err
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
	return group, nil
}

func (g *procGroup) adopt(cmd *exec.Cmd) error {
	process, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(cmd.Process.Pid))
	if err != nil {
		return err
	}
	defer windows.CloseHandle(process)
	return windows.AssignProcessToJobObject(g.job, process)
}

func (g *procGroup) kill(*exec.Cmd) error {
	return windows.TerminateJobObject(g.job, 1)
}

func (g *procGroup) close() {
	_ = windows.CloseHandle(g.job)
}
