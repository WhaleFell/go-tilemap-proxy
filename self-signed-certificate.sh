#!/bin/bash
set -e # Exit immediately if a command exits with a non-zero status.
set -u # Treat unset variables as an error when substituting.

# --- 配置信息 (请修改为您的实际信息) ---
ORG_NAME="YourCompany"        # 您的组织名称
STATE_NAME="YourState"        # 您的省份/州
CITY_NAME="YourCity"          # 您的城市
COUNTRY_CODE="CN"             # 您的国家代码 (例如: CN, US, GB)
CA_COMMON_NAME="My Private CDN Root CA" # Root CA 的通用名称
SERVER_UNIT_NAME="Origin Server" # 服务器证书的组织单位

# --- 文件名 ---
CA_KEY="ca.key"
CA_CRT="ca.crt"
SERVER_KEY="server.key"
SERVER_CSR="server.csr"
SERVER_CRT="server.crt"
OPENSSL_CNF="openssl_ip.cnf"
CA_SERIAL="ca.srl" # OpenSSL 内部使用的序列号文件

# --- 提示用户输入信息 ---
echo "--- IP SAN 证书生成脚本 ---"
read -p "请输入您的源站 IP 地址 (例如: 192.168.1.100): " ORIGIN_IP
if [[ -z "$ORIGIN_IP" ]]; then
    echo "错误: IP 地址不能为空。"
    exit 1
fi

read -s -p "请为您的 Root CA 私钥设置一个强密码 (请牢记!): " CA_PASSWORD
echo # 换行
if [[ -z "$CA_PASSWORD" ]]; then
    echo "错误: CA 密码不能为空。"
    exit 1
fi
echo "CA 密码已设置。"

echo ""
echo "正在生成证书文件到当前目录..."

# --- 1. 创建用于 IP SAN 的 openssl 配置文件 ---
echo "正在创建 '${OPENSSL_CNF}' 配置文件..."
cat <<EOF > "${OPENSSL_CNF}"
[ req ]
default_bits        = 2048
prompt              = no
default_md          = sha256
req_extensions      = v3_req
distinguished_name  = req_distinguished_name

[ req_distinguished_name ]
countryName                 = ${COUNTRY_CODE}
stateOrProvinceName         = ${STATE_NAME}
localityName                = ${CITY_NAME}
organizationName            = ${ORG_NAME}
organizationalUnitName      = ${SERVER_UNIT_NAME}
commonName                  = ${ORIGIN_IP} # 可以是IP，但SAN更重要

[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = IP:${ORIGIN_IP} # <<<<<<<<<<<<<<< 关键的 IP SAN 配置
EOF
echo "'${OPENSSL_CNF}' 已创建。"

# --- 2. 生成 Root CA 私钥和自签名证书 ---
echo "正在生成 Root CA 私钥和证书..."
openssl genrsa -aes256 -passout pass:"${CA_PASSWORD}" -out "${CA_KEY}" 4096
openssl req -x509 -new -key "${CA_KEY}" -passin pass:"${CA_PASSWORD}" -sha256 -days 3650 -out "${CA_CRT}" -subj "/C=${COUNTRY_CODE}/ST=${STATE_NAME}/L=${CITY_NAME}/O=${ORG_NAME}/CN=${CA_COMMON_NAME}"
echo "Root CA 证书 ('${CA_CRT}') 和私钥 ('${CA_KEY}') 已生成。"

# --- 3. 生成源站服务器私钥和证书 ---
echo "正在生成源站服务器私钥和证书..."
openssl genrsa -out "${SERVER_KEY}" 2048
openssl req -new -key "${SERVER_KEY}" -out "${SERVER_CSR}" -config "${OPENSSL_CNF}"
openssl x509 -req -in "${SERVER_CSR}" -CA "${CA_CRT}" -CAkey "${CA_KEY}" -CAcreateserial -passin pass:"${CA_PASSWORD}" -out "${SERVER_CRT}" -days 365 -sha256 -extensions v3_req -extfile "${OPENSSL_CNF}"
echo "源站服务器证书 ('${SERVER_CRT}') 和私钥 ('${SERVER_KEY}') 已生成。"

# --- 4. 清理中间文件 ---
echo "正在清理中间文件..."
rm -f "${SERVER_CSR}" "${OPENSSL_CNF}" "${CA_SERIAL}"
echo "清理完成。"

# --- 5. 验证证书 (可选) ---
echo ""
echo "--- 证书验证 (可选) ---"
echo "验证服务器证书是否包含 IP SAN:"
openssl x509 -in "${SERVER_CRT}" -text -noout | grep -A1 "Subject Alternative Name"
echo "验证证书链 ('${SERVER_CRT}' 由 '${CA_CRT}' 签发):"
openssl verify -CAfile "${CA_CRT}" "${SERVER_CRT}"

# --- 6. 最终指示 ---
echo ""
echo "--- 证书生成完成！请按照以下说明部署 ---"
echo "您已生成以下重要文件："
echo "1. Root CA 证书: ${CA_CRT}"
echo "2. 源站服务器证书: ${SERVER_CRT}"
echo "3. 源站服务器私钥: ${SERVER_KEY}"
echo ""
echo "--- 部署到您的源站 Web 服务器 (例如 Nginx, Apache) ---"
echo "   - 将 '${SERVER_CRT}' 和 '${SERVER_KEY}' 配置到您的 Web 服务器中。"
echo "   - 如果您的 Web 服务器需要证书链 (例如 'ssl_certificate_chain' 配置)，请将 '${CA_CRT}' 也配置进去。"
echo ""
echo "--- 部署到您的 CDN 平台 ---"
echo "   - 登录到您的 CDN 控制台，找到您的回源配置。"
echo "   - 将回源协议设置为 'HTTPS'。"
echo "   - 将回源地址设置为您的源站 IP 地址: ${ORIGIN_IP}"
echo "   - 查找类似 '自定义回源 CA 证书' 或 '信任证书链' 的选项。"
echo "   - 将 **'${CA_CRT}' 文件中的内容** 粘贴或上传到此配置项中。"
echo ""
echo "--- 重要提示 ---"
echo "   - 请妥善保管您的 Root CA 私钥 ('${CA_KEY}') 和密码。不要将其上传到任何地方！"
echo "   - 如果您需要为其他源站生成 IP 证书，可以使用相同的 '${CA_KEY}' 和 '${CA_CRT}' 来签署。"
echo ""
echo "脚本执行完毕。"

