package ssh

import (
	"log"

	"net"

	"fmt"

	"github.com/kvz/logstreamer"
	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
)

type Connection struct {
	logger       *log.Logger
	client       *ssh.Client
	hostname     string
	port         int
	clientConfig *ssh.ClientConfig
}

const Port = 22

func NewConnection(logger *log.Logger, hostname string, port int, username string, key []byte) (*Connection, error) {
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	cfg := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			logger.Printf("WARNING: ignoring SSH certificate issue for %s", hostname)
			return nil
		},
	}

	c := &Connection{logger: logger, hostname: hostname, port: port, clientConfig: cfg}

	return c, nil
}

func (c *Connection) EnsureConnected() error {
	if c.client == nil {
		var err error
		c.client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.hostname, c.port), c.clientConfig)
		if err != nil {
			return err
		}
	}

	ses, err := c.client.NewSession()
	if err != nil {
		return err
	}
	return ses.Run("exit 0")
}

func (c *Connection) UploadFile(path, destName string) error {
	ses, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("uploading file: %v", err)
	}
	defer ses.Close()

	c.logger.Printf("uploading file %s as %s", path, destName)
	return scp.CopyPath(path, fmt.Sprintf("~/%s", destName), ses)
}

func (c *Connection) RunCommand(cmd string) error {
	ses, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("uploading file: %v", err)
	}
	defer ses.Close()

	c.logger.Printf("running command %s", cmd)

	out := logstreamer.NewLogstreamer(c.logger, "", false)
	defer out.Close()

	ses.Stdout = out
	ses.Stderr = out
	if err := ses.Run(cmd); err != nil {
		return fmt.Errorf("running command: %v", err)
	}

	return nil
}

func (c *Connection) Close() {
	c.client.Close()
}