# SSL证书配置说明 | SSL Certificate Configuration Guide

## 证书文件要求 | Certificate File Requirements

请将您的SSL证书文件放置在此目录下，文件名必须匹配nginx.conf中的配置：

Please place your SSL certificate files in this directory with the following filenames as configured in nginx.conf:

### 必需文件 | Required Files

1. **fullchain.pem** - 完整证书链文件 | Full certificate chain file
   - 包含您的域名证书和中间证书 | Contains your domain certificate and intermediate certificates
   
2. **privkey.pem** - 私钥文件 | Private key file
   - 对应域名证书的私钥 | Private key corresponding to your domain certificate

### 文件权限 | File Permissions

确保证书文件具有正确的权限：
Make sure certificate files have correct permissions:

```bash
chmod 644 fullchain.pem
chmod 600 privkey.pem
```

### 获取SSL证书的方式 | Ways to Obtain SSL Certificates

#### 1. Let's Encrypt (免费) | Let's Encrypt (Free)

使用Certbot获取免费SSL证书：
Use Certbot to obtain free SSL certificates:

```bash
# 安装certbot | Install certbot
sudo apt-get update
sudo apt-get install certbot

# 获取证书 | Obtain certificate
sudo certbot certonly --standalone -d your-domain.com

# 证书通常位于 | Certificates are usually located at:
# /etc/letsencrypt/live/your-domain.com/fullchain.pem
# /etc/letsencrypt/live/your-domain.com/privkey.pem
```

#### 2. 商业证书 | Commercial Certificates

如果您有来自商业CA的证书，请确保：
If you have certificates from commercial CAs, make sure:

- 证书文件是PEM格式 | Certificate files are in PEM format
- 包含完整的证书链 | Include the complete certificate chain
- 私钥没有密码保护 | Private key is not password protected

#### 3. 自签名证书（仅用于测试）| Self-signed Certificates (Testing Only)

⚠️ **警告：自签名证书仅适用于开发和测试环境**
⚠️ **Warning: Self-signed certificates are only suitable for development and testing**

```bash
# 生成自签名证书 | Generate self-signed certificate
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout privkey.pem \
    -out fullchain.pem \
    -subj "/C=CN/ST=State/L=City/O=Organization/CN=localhost"
```

### 证书更新 | Certificate Renewal

#### Let's Encrypt自动更新 | Let's Encrypt Auto-renewal

```bash
# 设置自动更新cron任务 | Set up auto-renewal cron job
sudo crontab -e

# 添加以下行 | Add the following line:
0 3 * * * certbot renew --quiet --post-hook "docker-compose restart nginx"
```

### 验证证书 | Verify Certificates

部署前验证证书文件：
Verify certificate files before deployment:

```bash
# 检查证书信息 | Check certificate information
openssl x509 -in fullchain.pem -text -noout

# 检查私钥 | Check private key
openssl rsa -in privkey.pem -check

# 验证证书和私钥匹配 | Verify certificate and private key match
openssl x509 -noout -modulus -in fullchain.pem | openssl md5
openssl rsa -noout -modulus -in privkey.pem | openssl md5
```

### 故障排除 | Troubleshooting

#### 常见问题 | Common Issues

1. **证书文件权限错误 | Certificate file permission errors**
   ```bash
   sudo chown root:root fullchain.pem privkey.pem
   sudo chmod 644 fullchain.pem
   sudo chmod 600 privkey.pem
   ```

2. **nginx启动失败 | Nginx startup failure**
   - 检查证书文件是否存在 | Check if certificate files exist
   - 验证证书格式 | Verify certificate format
   - 查看nginx错误日志 | Check nginx error logs

3. **浏览器证书警告 | Browser certificate warnings**
   - 确保域名匹配 | Ensure domain name matches
   - 检查证书是否过期 | Check if certificate is expired
   - 验证证书链完整性 | Verify certificate chain integrity

### 安全建议 | Security Recommendations

1. **定期更新证书 | Regularly update certificates**
2. **使用强加密算法 | Use strong encryption algorithms**
3. **启用HSTS | Enable HSTS**
4. **监控证书过期时间 | Monitor certificate expiration**
5. **备份证书文件 | Backup certificate files**

---

配置完成后，使用以下命令启动服务：
After configuration, start the services with:

```bash
docker-compose up -d
```

检查服务状态：
Check service status:

```bash
docker-compose ps
docker-compose logs nginx