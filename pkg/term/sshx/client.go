package sshx

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

func ConnectToNextJumpHost(ctx context.Context, currentClient *ssh.Client, user string, host string, port int, method ssh.AuthMethod) (*ssh.Client, error) {
	// 创建下一层跳板机的 SSH 客户端配置
	nextConfig := newSshConfig(user, method)

	address := fmt.Sprintf("%s:%d", host, port)
	// 通过当前跳板机连接到下一层跳板机
	nextConn, err := currentClient.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to next jump host: %v", err)
	}

	// 创建下一层跳板机的 SSH 会话
	nextClient, nextChan, nextReqs, err := ssh.NewClientConn(nextConn, address, nextConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH client connection to next jump host: %v", err)
	}

	// 创建下一层跳板机的 SSH 客户端
	return ssh.NewClient(nextClient, nextChan, nextReqs), nil
}

func (mgm *MultiGatewayManager) Connect(ctx context.Context) (*ssh.Client, error) {
	var client *ssh.Client
	var err error

	for i, gateway := range mgm.Gateways {
		// 获取认证方式
		var authMethod ssh.AuthMethod
		authMethod, err = Auth(gateway)
		if err != nil {
			return nil, err
		}

		// 配置 SSH 客户端
		config := newSshConfig(gateway.Username, authMethod)

		if i == 0 {
			// 连接第一个网关
			client, err = dialWithContext(ctx, gateway.Host, gateway.Port, config)
			if err != nil {
				return nil, err
			}
		} else {
			// 这里你应该确保 client 是继续在原连接上做的，而不是新建连接
			client, err = ConnectToNextJumpHost(ctx, client, gateway.Username, gateway.Host, gateway.Port, authMethod)
			if err != nil {
				return nil, fmt.Errorf("通过网关 %s 连接失败: %v", gateway.Host, err)
			}
		}
	}

	return client, nil
}

func newSshConfig(username string, authMethod ssh.AuthMethod) *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         20 * time.Second,
	}
}

func dialWithContext(ctx context.Context, host string, port int, config *ssh.ClientConfig) (*ssh.Client, error) {
	// 连接第一个网关
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %v", err)
	}

	var client *ssh.Client
	// 使用 goroutine 和 select 处理 SSH 握手超时
	errChan := make(chan error, 1)
	go func() {
		// 完成 SSH 握手
		c, channels, reqs, er := ssh.NewClientConn(conn, address, config)
		if er != nil {
			errChan <- fmt.Errorf("SSH 握手失败: %v", er)
			return
		}

		client = ssh.NewClient(c, channels, reqs)
		errChan <- nil
	}()

	// 等待握手完成或上下文超时
	select {
	case err = <-errChan:
		if err != nil {
			return nil, err
		}
		return client, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func Auth(gateway *GatewayConfig) (ssh.AuthMethod, error) {
	switch gateway.AuthType {
	case "passwd":
		return ssh.Password(gateway.Password), nil
	case "passphrase":
		privateKeyBytes := []byte(gateway.PrivateKey)
		passwordBytes := []byte(gateway.Password)
		signer, err := ssh.ParsePrivateKeyWithPassphrase(privateKeyBytes, passwordBytes)
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %v", err)
		}
		return ssh.PublicKeys(signer), nil
	case "publickey":
		privateKeyBytes := []byte(gateway.PrivateKey)
		signer, err := ssh.ParsePrivateKey(privateKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %v", err)
		}

		return ssh.PublicKeys(signer), nil
	}

	return nil, fmt.Errorf("无可匹配认证类型, 请进行绑定")
}
