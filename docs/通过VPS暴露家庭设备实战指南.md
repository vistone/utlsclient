# 通过 VPS 暴露家庭设备实战指南

## 📋 目录
1. [需求分析](#需求分析)
2. [技术方案对比](#技术方案对比)
3. [方案一：SSH 反向隧道](#方案一ssh-反向隧道)
4. [方案二：frp 内网穿透](#方案二frp-内网穿透)
5. [方案三：WireGuard VPN](#方案三wireguard-vpn)
6. [方案四：Tailscale](#方案四tailscale)
7. [组合 uTLS/uQUIC 增强隐蔽性](#组合-utlsuquic-增强隐蔽性)
8. [完整实战案例](#完整实战案例)

---

## 需求分析

你想做的：
```
家庭设备（NAT后面）-> VPS（公网IP）-> 公网访问
         ↑                    ↑
      内网IP，无法         提供公网入口
      从外网访问
```

**应用场景**：
- 家庭 NAS 远程访问
- 内网 Web 服务暴露
- 远程桌面连接
- 内网应用访问
- 绕过 NAT 限制

**注意**：这**不是** TapDance 的功能！TapDance 是用于绕过审查的反向代理。

---

## 技术方案对比

| 方案 | 难度 | 性能 | 稳定性 | 推荐度 |
|------|------|------|--------|--------|
| SSH 反向隧道 | ⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| frp | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| WireGuard VPN | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| Tailscale | ⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| ngrok | ⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ |

---

## 方案一：SSH 反向隧道

### 优点
- ✅ 简单，SSH 自带
- ✅ 无需额外软件
- ✅ 安全可靠
- ✅ 跨平台

### 缺点
- ❌ 需要保持 SSH 连接
- ❌ 断开需重连
- ❌ 端口固定

### 实现步骤

#### 在家庭设备上运行（被访问的设备）

```bash
# 安装 SSH 客户端
sudo apt install openssh-client  # Debian/Ubuntu
brew install openssh            # macOS

# 建立反向隧道
ssh -R 8080:localhost:80 root@your-vps-ip -N -f

# 参数说明：
# -R 8080:localhost:80: VPS 的 8080 端口转发到本地的 80 端口
# -N: 不执行远程命令
# -f: 后台运行
```

#### 在 VPS 上配置 SSH 服务端

```bash
# 编辑 SSH 配置
sudo nano /etc/ssh/sshd_config

# 添加以下配置
GatewayPorts yes          # 允许外部访问
ClientAliveInterval 60    # 保持连接
ClientAliveCountMax 3

# 重启 SSH 服务
sudo systemctl restart sshd
```

#### 访问测试

```bash
# 从任何地方访问 VPS:8080 即可访问家庭设备
curl http://your-vps-ip:8080
```

### 自动重连脚本

```bash
#!/bin/bash
# auto_reverse_ssh.sh

VPS_IP="your-vps-ip"
VPS_USER="root"
LOCAL_PORT=80
REMOTE_PORT=8080

while true; do
    ssh -R ${REMOTE_PORT}:localhost:${LOCAL_PORT} ${VPS_USER}@${VPS_IP} \
        -o ServerAliveInterval=60 \
        -o ServerAliveCountMax=3 \
        -o StrictHostKeyChecking=no \
        -N
    
    echo "Connection lost, reconnecting in 5 seconds..."
    sleep 5
done
```

**运行**：
```bash
chmod +x auto_reverse_ssh.sh
./auto_reverse_ssh.sh
```

### 使用 systemd 服务

```ini
# /etc/systemd/system/reverse-ssh.service
[Unit]
Description=Reverse SSH Tunnel
After=network.target

[Service]
Type=simple
User=your-username
Restart=always
RestartSec=5
ExecStart=/usr/bin/ssh -R 8080:localhost:80 -o ServerAliveInterval=60 root@your-vps-ip -N

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable reverse-ssh
sudo systemctl start reverse-ssh
sudo systemctl status reverse-ssh
```

---

## 方案二：frp 内网穿透

**frp** 是推荐的内网穿透工具。

### 安装 frp

```bash
# 下载 frp
wget https://github.com/fatedier/frp/releases/download/v0.52.3/frp_0.52.3_linux_amd64.tar.gz
tar -xzf frp_0.52.3_linux_amd64.tar.gz
cd frp_0.52.3_linux_amd64
```

### VPS 配置（服务端）

```ini
# frps.ini
[common]
bind_port = 7000           # frp 控制端口
dashboard_port = 7500      # Web 面板端口
dashboard_user = admin
dashboard_pwd = your-password

token = your-secret-token  # 认证令牌

# 日志配置
log_file = /var/log/frp/frps.log
log_level = info
log_max_days = 3
```

**启动服务端**：
```bash
./frps -c frps.ini
```

**systemd 配置**：
```ini
# /etc/systemd/system/frps.service
[Unit]
Description=Frp Server
After=network.target

[Service]
Type=simple
User=nobody
Restart=on-failure
RestartSec=5s
ExecStart=/usr/local/bin/frps -c /etc/frp/frps.ini

[Install]
WantedBy=multi-user.target
```

### 家庭设备配置（客户端）

```ini
# frpc.ini
[common]
server_addr = your-vps-ip
server_port = 7000
token = your-secret-token

# 暴露本地 Web 服务
[web]
type = tcp
local_ip = 127.0.0.1
local_port = 80
remote_port = 8080

# 暴露 SSH
[ssh]
type = tcp
local_ip = 127.0.0.1
local_port = 22
remote_port = 6000

# 暴露 VNC
[vnc]
type = tcp
local_ip = 127.0.0.1
local_port = 5900
remote_port = 5900
```

**启动客户端**：
```bash
./frpc -c frpc.ini
```

### 高级配置

#### 域名访问

```ini
# frpc.ini
[web]
type = http
local_ip = 127.0.0.1
local_port = 80
custom_domains = home.example.com
subdomain = home  # 使用你的域名
```

#### SSL/TLS

```ini
# frpc.ini
[web]
type = https
local_ip = 127.0.0.1
local_port = 443
custom_domains = home.example.com

# 配置证书
plugin_cert_path = /path/to/cert.pem
plugin_key_path = /path/to/key.pem
```

---

## 方案三：WireGuard VPN

WireGuard 是新一代 VPN，性能极好。

### VPS 服务端配置

```bash
# 安装 WireGuard
sudo apt install wireguard wireguard-tools

# 生成密钥
wg genkey | tee /etc/wireguard/privatekey | wg pubkey > /etc/wireguard/publickey

# 配置服务端
sudo nano /etc/wireguard/wg0.conf
```

```ini
[Interface]
PrivateKey = <VPS_PRIVATE_KEY>
Address = 10.0.0.1/24
ListenPort = 51820
PostUp = iptables -A FORWARD -i wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE

[Peer]
PublicKey = <CLIENT_PUBLIC_KEY>
AllowedIPs = 10.0.0.2/32
```

**启用**：
```bash
sudo wg-quick up wg0
sudo systemctl enable wg-quick@wg0
```

### 家庭设备配置

```bash
# 生成客户端密钥
wg genkey | tee privatekey | wg pubkey > publickey

# 配置
sudo nano /etc/wireguard/wg0.conf
```

```ini
[Interface]
PrivateKey = <CLIENT_PRIVATE_KEY>
Address = 10.0.0.2/24

[Peer]
PublicKey = <VPS_PUBLIC_KEY>
Endpoint = your-vps-ip:51820
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = 25
```

**启用**：
```bash
sudo wg-quick up wg0
sudo systemctl enable wg-quick@wg0
```

---

## 方案四：Tailscale

简单、易用，开箱即用。

### 安装

```bash
# 安装 Tailscale
curl -fsSL https://tailscale.com/install.sh | sh

# 或使用包管理器
# macOS
brew install tailscale

# Ubuntu/Debian
curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/jammy.noarmor.gpg | sudo tee /usr/share/keyrings/tailscale-archive-keyring.gpg >/dev/null
```

### 使用

```bash
# 在两台设备上运行
sudo tailscale up

# 访问 https://login.tailscale.com/admin/machines
# 两台设备会自动组网
```

### 暴露服务

```bash
# 在家庭设备上
# 暴露 HTTP 服务
sudo tailscale serve http / http://localhost:8080

# 暴露目录
sudo tailscale serve file /home/user/share /var/www/html
```

---

## 组合 uTLS/uQUIC 增强隐蔽性

### 场景：在反审查网络中使用

如果需要从被审查的网络访问家庭设备，可以叠加工具：

```
被审查网络 -> uTLS/uQUIC/TapDance -> VPS -> frp -> 家庭设备
```

### 实现示例

#### 使用 uTLS 的 frp 客户端

```go
package main

import (
    "fmt"
    "net/http"
    utls "github.com/refraction-networking/utls"
    "github.com/fatedier/frp/client"
)

func createUTLSClient() *http.Client {
    transport := &http.Transport{
        DialTLS: func(network, addr string) (net.Conn, error) {
            conn, err := net.Dial(network, addr)
            if err != nil {
                return nil, err
            }
            
            config := &utls.Config{
                ServerName:         addr,
                InsecureSkipVerify: false,
            }
            
            // 使用 Firefox 指纹
            uconn := utls.UClient(conn, config, utls.HelloFirefox_Auto)
            if err := uconn.Handshake(); err != nil {
                conn.Close()
                return nil, err
            }
            
            return uconn, nil
        },
    }
    
    return &http.Client{Transport: transport}
}
```

---

## 完整实战案例

### 案例：暴露家庭 NAS Web 界面

#### 架构图

```
家庭 NAS (192.168.1.100:80)
    ↓
frp 客户端
    ↓
VPS (公网 IP)
    ↓
公网访问 -> NAS Web 界面
```

#### VPS 部署

```bash
# 1. 安装 frp
wget https://github.com/fatedier/frp/releases/download/v0.52.3/frp_0.52.3_linux_amd64.tar.gz
tar -xzf frp_0.52.3_linux_amd64.tar.gz
sudo cp frp_0.52.3_linux_amd64/frps /usr/local/bin/

# 2. 创建配置目录
sudo mkdir -p /etc/frp
sudo nano /etc/frp/frps.ini
```

```ini
[common]
bind_port = 7000
dashboard_port = 7500
dashboard_user = admin
dashboard_pwd = your-strong-password
token = your-secret-token
```

```bash
# 3. 创建 systemd 服务
sudo nano /etc/systemd/system/frps.service
```

```ini
[Unit]
Description=Frp Server
After=network.target

[Service]
Type=simple
Restart=always
RestartSec=5
ExecStart=/usr/local/bin/frps -c /etc/frp/frps.ini

[Install]
WantedBy=multi-user.target
```

```bash
# 4. 启动服务
sudo systemctl enable frps
sudo systemctl start frps
sudo systemctl status frps

# 5. 开放端口
sudo ufw allow 7000
sudo ufw allow 7500
sudo ufw allow 8080
```

#### NAS 部署

```bash
# 1. 下载 frp 客户端
wget https://github.com/fatedier/frp/releases/download/v0.52.3/frp_0.52.3_linux_amd64.tar.gz
tar -xzf frp_0.52.3_linux_amd64.tar.gz
sudo cp frp_0.52.3_linux_amd64/frpc /usr/local/bin/

# 2. 配置
sudo nano /etc/frp/frpc.ini
```

```ini
[common]
server_addr = your-vps-ip
server_port = 7000
token = your-secret-token

[nas-web]
type = tcp
local_ip = 127.0.0.1
local_port = 80
remote_port = 8080
```

```bash
# 3. 创建服务
sudo nano /etc/systemd/system/frpc.service
```

```ini
[Unit]
Description=Frp Client
After=network.target

[Service]
Type=simple
Restart=always
RestartSec=5
ExecStart=/usr/local/bin/frpc -c /etc/frp/frpc.ini

[Install]
WantedBy=multi-user.target
```

```bash
# 4. 启动
sudo systemctl enable frpc
sudo systemctl start frpc
```

#### 访问测试

```bash
# 从任何地方访问
curl http://your-vps-ip:8080

# 浏览器访问
http://your-vps-ip:8080
```

### 案例：远程桌面访问

```ini
# frpc.ini
[rdp]
type = tcp
local_ip = 127.0.0.1
local_port = 3389
remote_port = 13389
```

**Windows RDP**：
```
连接到：your-vps-ip:13389
```

### 案例：SSH 访问

```ini
# frpc.ini
[ssh]
type = tcp
local_ip = 127.0.0.1
local_port = 22
remote_port = 6000
```

```bash
ssh -p 6000 user@your-vps-ip
```

---

## 安全建议

### ⚠️ 重要安全措施

1. **使用强密码和密钥**
2. **配置防火墙**
3. **使用 HTTPS/WSS**
4. **启用认证**
5. **定期更新**
6. **最小权限**
7. **审计日志**
8. **不要暴露敏感服务**

### frp 安全配置示例

```ini
# frps.ini
[common]
bind_port = 7000
token = very-long-random-secret-token

# 速率限制
max_pool_count = 5
max_ports_per_client = 0
allow_ports = 2000-3000,3001,3003,4000-50000

# 日志
log_file = /var/log/frp/frps.log
log_level = warn
log_max_days = 7

# TLS
tls_cert_file = /path/to/cert.pem
tls_key_file = /path/to/key.pem
```

---

## 常见问题解答

### Q1: TapDance 可以用于内网穿透吗？

**A:** **不可以**。TapDance 用于绕过审查，不用于内网穿透。用 frp、SSH 反向隧道或 Tailscale。

### Q2: 哪种方案最好？

**A:** 
- 简单稳定：Tailscale
- 灵活可控：frp
- 临时使用：SSH 反向隧道
- 长期稳定：WireGuard

### Q3: 如何提高隐蔽性？

**A:** 
- 使用 uTLS 修改 TLS 指纹
- 使用 uQUIC 适配 QUIC 场景
- 通过 TapDance 连接 VPS
- 加密通信

### Q4: 服务断开怎么办？

**A:** 
- 使用 systemd 自动重启
- 用进程管理（supervisor/systemd）
- 监控和告警

### Q5: 多设备如何管理？

**A:** 
- Tailscale：自动组网
- frp：前端 + 配置管理
- WireGuard：多 Peer
- SSH：端口/配置文件映射

---

## 参考资料

- **frp GitHub**：https://github.com/fatedier/frp
- **Tailscale**：https://tailscale.com
- **WireGuard**：https://www.wireguard.com
- **SSH 隧道**：https://man.openbsd.org/ssh

---

**创建日期**：2025-01-10  
**说明**：通过 VPS 暴露家庭设备的实战指南
