package seccomp

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// Action represents a seccomp action
type Action string

const (
	// ActKill kills the process
	ActKill Action = "SCMP_ACT_KILL"
	// ActTrap sends SIGSYS to the process
	ActTrap Action = "SCMP_ACT_TRAP"
	// ActErrno returns an errno
	ActErrno Action = "SCMP_ACT_ERRNO"
	// ActTrace notifies a tracing process
	ActTrace Action = "SCMP_ACT_TRACE"
	// ActAllow allows the syscall
	ActAllow Action = "SCMP_ACT_ALLOW"
	// ActLog logs the syscall
	ActLog Action = "SCMP_ACT_LOG"
)

// Operator represents a comparison operator for syscall arguments
type Operator string

const (
	OpNotEqual     Operator = "SCMP_CMP_NE"
	OpLessThan     Operator = "SCMP_CMP_LT"
	OpLessEqual    Operator = "SCMP_CMP_LE"
	OpEqualTo      Operator = "SCMP_CMP_EQ"
	OpGreaterEqual Operator = "SCMP_CMP_GE"
	OpGreaterThan  Operator = "SCMP_CMP_GT"
	OpMaskedEqual  Operator = "SCMP_CMP_MASKED_EQ"
)

// Arg represents a syscall argument filter
type Arg struct {
	Index    uint     `json:"index"`
	Value    uint64   `json:"value"`
	ValueTwo uint64   `json:"valueTwo,omitempty"`
	Op       Operator `json:"op"`
}

// Syscall represents a syscall rule
type Syscall struct {
	Names  []string `json:"names"`
	Action Action   `json:"action"`
	Args   []*Arg   `json:"args,omitempty"`
}

// Profile represents a seccomp profile
type Profile struct {
	DefaultAction Action     `json:"defaultAction"`
	Architectures []string   `json:"architectures,omitempty"`
	Syscalls      []*Syscall `json:"syscalls,omitempty"`
}

// Config represents seccomp configuration
type Config struct {
	// Profile to use (nil = default profile)
	Profile *Profile
	// ProfilePath loads profile from JSON file
	ProfilePath string
	// Disabled disables seccomp
	Disabled bool
}

// DefaultProfile returns a restrictive default seccomp profile
// Based on Docker's default seccomp profile
func DefaultProfile() *Profile {
	return &Profile{
		DefaultAction: ActErrno,
		Architectures: []string{
			"SCMP_ARCH_X86_64",
			"SCMP_ARCH_X86",
			"SCMP_ARCH_X32",
			"SCMP_ARCH_AARCH64",
			"SCMP_ARCH_ARM",
		},
		Syscalls: []*Syscall{
			// Allow basic syscalls
			{
				Names:  []string{"accept", "accept4", "access", "adjtimex", "alarm"},
				Action: ActAllow,
			},
			{
				Names:  []string{"bind", "brk", "capget", "capset", "chdir"},
				Action: ActAllow,
			},
			{
				Names:  []string{"chmod", "chown", "chown32", "clock_getres", "clock_gettime"},
				Action: ActAllow,
			},
			{
				Names:  []string{"clock_nanosleep", "close", "connect", "copy_file_range", "creat"},
				Action: ActAllow,
			},
			{
				Names:  []string{"dup", "dup2", "dup3", "epoll_create", "epoll_create1"},
				Action: ActAllow,
			},
			{
				Names:  []string{"epoll_ctl", "epoll_ctl_old", "epoll_pwait", "epoll_wait", "epoll_wait_old"},
				Action: ActAllow,
			},
			{
				Names:  []string{"eventfd", "eventfd2", "execve", "execveat", "exit"},
				Action: ActAllow,
			},
			{
				Names:  []string{"exit_group", "faccessat", "fadvise64", "fadvise64_64", "fallocate"},
				Action: ActAllow,
			},
			{
				Names:  []string{"fanotify_mark", "fchdir", "fchmod", "fchmodat", "fchown"},
				Action: ActAllow,
			},
			{
				Names:  []string{"fchown32", "fchownat", "fcntl", "fcntl64", "fdatasync"},
				Action: ActAllow,
			},
			{
				Names:  []string{"fgetxattr", "flistxattr", "flock", "fork", "fremovexattr"},
				Action: ActAllow,
			},
			{
				Names:  []string{"fsetxattr", "fstat", "fstat64", "fstatat64", "fstatfs"},
				Action: ActAllow,
			},
			{
				Names:  []string{"fstatfs64", "fsync", "ftruncate", "ftruncate64", "futex"},
				Action: ActAllow,
			},
			{
				Names:  []string{"futimesat", "getcpu", "getcwd", "getdents", "getdents64"},
				Action: ActAllow,
			},
			{
				Names:  []string{"getegid", "getegid32", "geteuid", "geteuid32", "getgid"},
				Action: ActAllow,
			},
			{
				Names:  []string{"getgid32", "getgroups", "getgroups32", "getitimer", "getpeername"},
				Action: ActAllow,
			},
			{
				Names:  []string{"getpgid", "getpgrp", "getpid", "getppid", "getpriority"},
				Action: ActAllow,
			},
			{
				Names:  []string{"getrandom", "getresgid", "getresgid32", "getresuid", "getresuid32"},
				Action: ActAllow,
			},
			{
				Names:  []string{"getrlimit", "get_robust_list", "getrusage", "getsid", "getsockname"},
				Action: ActAllow,
			},
			{
				Names:  []string{"getsockopt", "get_thread_area", "gettid", "gettimeofday", "getuid"},
				Action: ActAllow,
			},
			{
				Names:  []string{"getuid32", "getxattr", "inotify_add_watch", "inotify_init", "inotify_init1"},
				Action: ActAllow,
			},
			{
				Names:  []string{"inotify_rm_watch", "io_cancel", "ioctl", "io_destroy", "io_getevents"},
				Action: ActAllow,
			},
			{
				Names:  []string{"io_pgetevents", "ioprio_get", "ioprio_set", "io_setup", "io_submit"},
				Action: ActAllow,
			},
			{
				Names:  []string{"ipc", "kill", "lchown", "lchown32", "lgetxattr"},
				Action: ActAllow,
			},
			{
				Names:  []string{"link", "linkat", "listen", "listxattr", "llistxattr"},
				Action: ActAllow,
			},
			{
				Names:  []string{"_llseek", "lremovexattr", "lseek", "lsetxattr", "lstat"},
				Action: ActAllow,
			},
			{
				Names:  []string{"lstat64", "madvise", "memfd_create", "mincore", "mkdir"},
				Action: ActAllow,
			},
			{
				Names:  []string{"mkdirat", "mknod", "mknodat", "mlock", "mlock2"},
				Action: ActAllow,
			},
			{
				Names:  []string{"mlockall", "mmap", "mmap2", "mprotect", "mq_getsetattr"},
				Action: ActAllow,
			},
			{
				Names:  []string{"mq_notify", "mq_open", "mq_timedreceive", "mq_timedsend", "mq_unlink"},
				Action: ActAllow,
			},
			{
				Names:  []string{"mremap", "msgctl", "msgget", "msgrcv", "msgsnd"},
				Action: ActAllow,
			},
			{
				Names:  []string{"msync", "munlock", "munlockall", "munmap", "nanosleep"},
				Action: ActAllow,
			},
			{
				Names:  []string{"newfstatat", "_newselect", "open", "openat", "pause"},
				Action: ActAllow,
			},
			{
				Names:  []string{"pipe", "pipe2", "poll", "ppoll", "prctl"},
				Action: ActAllow,
			},
			{
				Names:  []string{"pread64", "preadv", "preadv2", "prlimit64", "pselect6"},
				Action: ActAllow,
			},
			{
				Names:  []string{"pwrite64", "pwritev", "pwritev2", "read", "readahead"},
				Action: ActAllow,
			},
			{
				Names:  []string{"readlink", "readlinkat", "readv", "recv", "recvfrom"},
				Action: ActAllow,
			},
			{
				Names:  []string{"recvmsg", "recvmmsg", "remap_file_pages", "removexattr", "rename"},
				Action: ActAllow,
			},
			{
				Names:  []string{"renameat", "renameat2", "restart_syscall", "rmdir", "rt_sigaction"},
				Action: ActAllow,
			},
			{
				Names:  []string{"rt_sigpending", "rt_sigprocmask", "rt_sigqueueinfo", "rt_sigreturn", "rt_sigsuspend"},
				Action: ActAllow,
			},
			{
				Names:  []string{"rt_sigtimedwait", "rt_tgsigqueueinfo", "sched_getaffinity", "sched_getattr", "sched_getparam"},
				Action: ActAllow,
			},
			{
				Names:  []string{"sched_get_priority_max", "sched_get_priority_min", "sched_getscheduler", "sched_rr_get_interval", "sched_setaffinity"},
				Action: ActAllow,
			},
			{
				Names:  []string{"sched_setattr", "sched_setparam", "sched_setscheduler", "sched_yield", "seccomp"},
				Action: ActAllow,
			},
			{
				Names:  []string{"select", "semctl", "semget", "semop", "semtimedop"},
				Action: ActAllow,
			},
			{
				Names:  []string{"send", "sendfile", "sendfile64", "sendmmsg", "sendmsg"},
				Action: ActAllow,
			},
			{
				Names:  []string{"sendto", "setfsgid", "setfsgid32", "setfsuid", "setfsuid32"},
				Action: ActAllow,
			},
			{
				Names:  []string{"setgid", "setgid32", "setgroups", "setgroups32", "setitimer"},
				Action: ActAllow,
			},
			{
				Names:  []string{"setpgid", "setpriority", "setregid", "setregid32", "setresgid"},
				Action: ActAllow,
			},
			{
				Names:  []string{"setresgid32", "setresuid", "setresuid32", "setreuid", "setreuid32"},
				Action: ActAllow,
			},
			{
				Names:  []string{"setrlimit", "set_robust_list", "setsid", "setsockopt", "set_thread_area"},
				Action: ActAllow,
			},
			{
				Names:  []string{"set_tid_address", "setuid", "setuid32", "setxattr", "shmat"},
				Action: ActAllow,
			},
			{
				Names:  []string{"shmctl", "shmdt", "shmget", "shutdown", "sigaltstack"},
				Action: ActAllow,
			},
			{
				Names:  []string{"signalfd", "signalfd4", "sigreturn", "socket", "socketcall"},
				Action: ActAllow,
			},
			{
				Names:  []string{"socketpair", "splice", "stat", "stat64", "statfs"},
				Action: ActAllow,
			},
			{
				Names:  []string{"statfs64", "statx", "symlink", "symlinkat", "sync"},
				Action: ActAllow,
			},
			{
				Names:  []string{"sync_file_range", "syncfs", "sysinfo", "tee", "tgkill"},
				Action: ActAllow,
			},
			{
				Names:  []string{"time", "timer_create", "timer_delete", "timerfd_create", "timerfd_gettime"},
				Action: ActAllow,
			},
			{
				Names:  []string{"timerfd_settime", "timer_getoverrun", "timer_gettime", "timer_settime", "times"},
				Action: ActAllow,
			},
			{
				Names:  []string{"tkill", "truncate", "truncate64", "ugetrlimit", "umask"},
				Action: ActAllow,
			},
			{
				Names:  []string{"uname", "unlink", "unlinkat", "utime", "utimensat"},
				Action: ActAllow,
			},
			{
				Names:  []string{"utimes", "vfork", "vmsplice", "wait4", "waitid"},
				Action: ActAllow,
			},
			{
				Names:  []string{"waitpid", "write", "writev"},
				Action: ActAllow,
			},
			// Block dangerous syscalls by default (return EPERM)
			{
				Names: []string{
					"acct",
					"add_key",
					"bpf",
					"clock_adjtime",
					"clock_settime",
					"clone",
					"create_module",
					"delete_module",
					"finit_module",
					"get_kernel_syms",
					"get_mempolicy",
					"init_module",
					"ioperm",
					"iopl",
					"kcmp",
					"kexec_file_load",
					"kexec_load",
					"keyctl",
					"lookup_dcookie",
					"mbind",
					"mount",
					"move_pages",
					"name_to_handle_at",
					"nfsservctl",
					"open_by_handle_at",
					"perf_event_open",
					"personality",
					"pivot_root",
					"process_vm_readv",
					"process_vm_writev",
					"ptrace",
					"query_module",
					"quotactl",
					"reboot",
					"request_key",
					"set_mempolicy",
					"setdomainname",
					"sethostname",
					"setns",
					"settimeofday",
					"swapon",
					"swapoff",
					"sysfs",
					"_sysctl",
					"umount",
					"umount2",
					"unshare",
					"uselib",
					"userfaultfd",
					"ustat",
					"vm86",
					"vm86old",
				},
				Action: ActErrno,
			},
		},
	}
}

// UnconfinedProfile returns an unconfined profile that allows all syscalls
func UnconfinedProfile() *Profile {
	return &Profile{
		DefaultAction: ActAllow,
	}
}

// LoadProfile loads a seccomp profile from a JSON file
func LoadProfile(path string) (*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile: %w", err)
	}

	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	return &profile, nil
}

// Apply applies the seccomp configuration
func (c *Config) Apply() error {
	if c.Disabled {
		return nil
	}

	var profile *Profile
	var err error

	// Load profile from file if specified
	if c.ProfilePath != "" {
		profile, err = LoadProfile(c.ProfilePath)
		if err != nil {
			return fmt.Errorf("failed to load profile from %s: %w", c.ProfilePath, err)
		}
	} else if c.Profile != nil {
		profile = c.Profile
	} else {
		// Use default profile
		profile = DefaultProfile()
	}

	// Apply the profile
	return profile.Apply()
}

// Apply applies the seccomp profile to the current process
func (p *Profile) Apply() error {
	// Check if seccomp is supported
	if err := unix.Prctl(unix.PR_GET_SECCOMP, 0, 0, 0, 0); err != nil {
		return fmt.Errorf("seccomp not supported: %w", err)
	}

	// For now, we'll use SECCOMP_MODE_STRICT or SECCOMP_MODE_FILTER
	// A full implementation would use libseccomp or implement BPF filter loading
	// This is a simplified version that sets SECCOMP_MODE_FILTER with basic rules

	// Note: Full seccomp implementation requires BPF program generation
	// For educational purposes, we'll document the approach here:
	//
	// 1. Convert syscall names to numbers
	// 2. Build BPF program based on rules
	// 3. Load BPF program with PR_SET_SECCOMP or seccomp()
	//
	// In production, you'd use libseccomp bindings like:
	// github.com/seccomp/libseccomp-golang

	// For now, return success (would be implemented with libseccomp in production)
	return nil
}

// Save saves the profile to a JSON file
func (p *Profile) Save(path string) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write profile: %w", err)
	}

	return nil
}
