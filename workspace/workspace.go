package workspace

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// 该文件的作用是：模拟overlay联合文件系统在docker中的使用
//需要lowerlayer,upperlayer,mnt,work
const (
	lowerlayer = "/projects/tinydocker/ubuntu-net-rootfs"
	upperlayer = "/projects/tinydocker/upper"
	mnt        = "/projects/tinydocker/merged"
	work       = "/projects/tinydocker/work"
	backup     = ".old"
)

func LowerLayer() string {
	return fmt.Sprintf("%s", lowerlayer)
}

func UpperLayer(containerName string) string {
	return fmt.Sprintf("%s/%s", upperlayer, containerName)
}

func WorkDir(containerName string) string {
	return fmt.Sprintf("%s/%s", work, containerName)
}

func BackUpDir(containerName string) string {
	return fmt.Sprintf("%s/%s", MntLayer(containerName), backup)
}

func MntLayer(containerName string) string {
	return fmt.Sprintf("%s/%s", mnt, containerName)
}

func DelMnt(containerName string) (err error) {
	if err = delMnt(MntLayer(containerName)); err != nil {
		return
	}
	if err = delMnt(WorkDir(containerName)); err != nil {
		return
	}
	if err = delMnt(UpperLayer(containerName)); err != nil {
		return
	}
	return
}

func delMnt(mntPath string) (err error) {
	if _, err = os.Stat(mntPath); err != nil {
		fmt.Printf("%s not found", mntPath)
		return
	}
	var stat syscall.Statfs_t
	if err = syscall.Statfs(mntPath, &stat); err != nil {
		return
	}
	if stat.Type != 61267 {
		_, err = exec.Command("umount", mntPath).CombinedOutput()
		if err != nil {
			return
		}
	}
	//del dir
	err = os.RemoveAll(mntPath)
	if err != nil {
		return
	}
	return
}

func Mnt(containerName string) (err error) {
	if err = os.MkdirAll(MntLayer(containerName), 0700); err != nil {
		return fmt.Errorf("mkdir mntlayer fail err=%s", err)
	}
	if err = os.MkdirAll(UpperLayer(containerName), 0700); err != nil {
		return fmt.Errorf("mkdir UpperLayer fail err=%s", err)
	}
	if err = os.MkdirAll(WorkDir(containerName), 0700); err != nil {
		return fmt.Errorf("mkdir WorkDir fail err=%s", err)
	}
	if _, err = os.Stat(MntLayer(containerName)); err != nil {
		fmt.Printf("Mnt Dir not found ,%s\n", err)
		return
	}
	if _, err = os.Stat(LowerLayer()); err != nil {
		fmt.Printf("Lower Dir not found ,%s\n", err)
		return
	}
	if _, err = os.Stat(UpperLayer(containerName)); err != nil {
		fmt.Printf("up Dir not found ,%s\n", err)
		return
	}
	if _, err = os.Stat(WorkDir(containerName)); err != nil {
		fmt.Printf("work Dir not found ,%s\n", err)
		return
	}
	if err := syscall.Mount("overlay", MntLayer(containerName), "overlay", 0,
		fmt.Sprintf("upperdir=%s,lowerdir=%s,workdir=%s",
			UpperLayer(containerName), LowerLayer(), WorkDir(containerName))); err != nil {
		return fmt.Errorf("mount overlay fail err=%s", err)
	}
	if err = syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("reclare rootfs private fail err=%s", err)
	}
	if err := syscall.Mount(MntLayer(containerName), MntLayer(containerName), "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs in new mnt space fail err=%s", err)
	}
	if err = os.MkdirAll(BackUpDir(containerName), 0700); err != nil {
		return fmt.Errorf("mkdir BackUpDir fail err=%s", err)
	}
	if err = syscall.PivotRoot(MntLayer(containerName), BackUpDir(containerName)); err != nil {
		return fmt.Errorf("pivot root  fail err=%s", err)
	}
	return
}
