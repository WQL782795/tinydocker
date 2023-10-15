package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"tinyDocker/workspace"
)

func main() {
	switch os.Args[1] {
	case "run":
		initCmd, err := os.Readlink("/proc/self/exe")
		if err != nil {
			fmt.Println("get init process error ", err)
			return
		}
		os.Args[1] = "init"
		containerName := os.Args[2]
		cmd := exec.Command(initCmd, os.Args[1:]...)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
				syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
		}
		cmd.Env = os.Environ()
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Start()
		if err != nil {
			fmt.Println(err)
		}
		err = cmd.Wait()
		if err != nil {
			return
		}
		if erro := workspace.DelMnt(containerName); erro != nil {
			fmt.Printf("Del Mnt fail, %s", erro)
			return
		}
		fmt.Println("init proc end", initCmd)
		return
	case "init":
		var (
			containerName = os.Args[2]
			cmd           = os.Args[3]
		)

		//使用overlay挂载目录
		err := workspace.Mnt(containerName)
		if err != nil {
			fmt.Println("Mnt fail, ", err)
			return
		}
		err = syscall.Chdir("/")
		if err != nil {
			fmt.Println("Chdir fail, ", err)
			return
		}

		if err = syscall.Unmount("/.old", syscall.MNT_DETACH); err != nil {
			fmt.Println("Unmount .old fail, ", err)
			return
		}
		defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
		err = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
		if err != nil {
			fmt.Println("mount proc fail, ", err)
			return
		}
		fmt.Printf("cmd: %s  ", cmd)
		fmt.Printf("args: %s\n  ", os.Args[3:])
		err = syscall.Exec(cmd, os.Args[3:], os.Environ())
		if err != nil {
			fmt.Println("exec proc fail ", err)
			return
		}
		fmt.Println("forever exec it ")
		return
	default:
		fmt.Println("not valid cmd")
	}
}
