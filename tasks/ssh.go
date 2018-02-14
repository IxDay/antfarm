package tasks

import (
	"context"
	"fmt"
	"github.com/ixday/antfarm"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"io"
	"net"
	"os"
	"os/user"
)

type SSHConfig struct {
	User string
	Host string
	Port int
	Auth []ssh.AuthMethod
	PTY  SSHConfigPTY
}

type SSHConfigPTY struct {
	Modes         ssh.TerminalModes
	Term          string
	Height, Width int
}

type SSHConfigRun struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type SSH struct {
	client *ssh.Client
	pty    SSHConfigPTY
}

func (ssh *SSH) Run(cmd string, options ...func(*SSHConfigRun)) antfarm.Task {
	sshConfigRun := &SSHConfigRun{os.Stdin, os.Stdout, os.Stderr}
	for _, option := range options {
		option(sshConfigRun)
	}

	return antfarm.TaskFunc(func(ctx context.Context) error {
		session, err := ssh.client.NewSession()
		if err != nil {
			return fmt.Errorf("Failed to create session: %s", err)
		}

		defer session.Close()

		if err := session.RequestPty(ssh.pty.Term, ssh.pty.Height, ssh.pty.Width, ssh.pty.Modes); err != nil {
			return fmt.Errorf("request for pseudo terminal failed: %s", err)
		}

		stdin, err := session.StdinPipe()
		if err != nil {
			return fmt.Errorf("Unable to setup stdin for session: %v", err)
		}
		go io.Copy(stdin, sshConfigRun.Stdin)

		stdout, err := session.StdoutPipe()
		if err != nil {
			return fmt.Errorf("Unable to setup stdout for session: %v", err)
		}
		go io.Copy(sshConfigRun.Stdout, stdout)

		stderr, err := session.StderrPipe()
		if err != nil {
			return fmt.Errorf("Unable to setup stderr for session: %v", err)
		}
		go io.Copy(sshConfigRun.Stderr, stderr)

		return session.Run(cmd)

	})
}

func NewSSH(options ...func(*SSHConfig)) (*SSH, error) {
	sshConfig := &SSHConfig{Host: "localhost", Port: 22, PTY: SSHConfigPTY{
		ssh.TerminalModes{
			ssh.ECHO:          0,     // disable echoing
			ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
			ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
		},
		"xterm", 40, 80,
	}}
	authMethods := []ssh.AuthMethod{}

	login, err := user.Current()
	if err != nil {
		login = &user.User{Username: "root"}
	}
	sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err == nil {
		authMethods = append(authMethods, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
	}

	sshConfig.User = login.Username
	sshConfig.Auth = authMethods

	for _, option := range options {
		option(sshConfig)
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshConfig.Host, sshConfig.Port), &ssh.ClientConfig{
		User: sshConfig.User,
		Auth: sshConfig.Auth,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to dial: %s", err)
	}

	return &SSH{client, sshConfig.PTY}, nil
}
