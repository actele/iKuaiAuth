// iKuai 认证 JavaScript

class IKuaiAuth {
    constructor() {
        // 动态获取API基础URL（支持不同的部署环境）
        this.apiBase = `${window.location.protocol}//${window.location.host}/api`;
        this.urlParams = this.parseURLParams();
        this.userIP = '';
        this.userMAC = '';
        this.gwid = '';
        this.routerInfo = {};
        this.debugMode = false; // 默认关闭，从配置读取
        this.config = null; // 配置信息（从后端获取）
        this.init();
    }

    // 添加调试日志到页面
    addDebugLog(message, type = 'info') {
        if (!this.debugMode) return;
        
        const debugInfo = document.getElementById('debugInfo');
        const debugContent = document.getElementById('debugContent');
        
        if (debugInfo && debugContent) {
            debugInfo.style.display = 'block';
            
            const timestamp = new Date().toLocaleTimeString();
            let color = '#0f0';
            let icon = 'ℹ️';
            
            switch(type) {
                case 'success':
                    color = '#0f0';
                    icon = '✅';
                    break;
                case 'error':
                    color = '#f00';
                    icon = '❌';
                    break;
                case 'warning':
                    color = '#ff0';
                    icon = '⚠️';
                    break;
                case 'request':
                    color = '#0af';
                    icon = '📤';
                    break;
                case 'response':
                    color = '#f0f';
                    icon = '📥';
                    break;
            }
            
            const logEntry = document.createElement('div');
            logEntry.style.marginBottom = '8px';
            logEntry.style.color = color;
            logEntry.innerHTML = `[${timestamp}] ${icon} ${message}`;
            
            debugContent.appendChild(logEntry);
            debugContent.scrollTop = debugContent.scrollHeight;
        }
        
        console.log(message);
    }

    // 解析URL参数（iKuai传递的参数）
    parseURLParams() {
        // 修复爱快路由器URL中的错误格式：将多余的 ? 替换为 &
        let search = window.location.search;
        
        // 替换第一个 ? 之后的所有 ? 为 &
        const firstQuestionMark = search.indexOf('?');
        if (firstQuestionMark !== -1) {
            const afterFirst = search.substring(firstQuestionMark + 1);
            const fixed = afterFirst.replace(/\?/g, '&');
            search = '?' + fixed;
        }
        
        console.log('原始URL参数:', window.location.search);
        console.log('修正后参数:', search);
        
        const params = new URLSearchParams(search);
        const result = {};
        
        // 获取所有参数
        for (const [key, value] of params.entries()) {
            result[key] = value;
        }
        
        // 解码user_ip（iKuai会将IP编码为 192%2E168%2E9%2E100 格式）
        if (result.user_ip) {
            result.user_ip = decodeURIComponent(result.user_ip);
        }
        
        // 处理mac地址（如果是"null"字符串，转为空）
        if (result.mac === 'null' || !result.mac) {
            result.mac = '';
        }
        
        console.log('解析URL参数:', result);
        return result;
    }

    // 初始化
    async init() {
        await this.loadConfig(); // 先加载配置
        this.parseRouterInfo();
        await this.getDeviceInfo();
        this.bindEvents();
        this.loadSavedUsername();
        this.displayRouterInfo();
    }

    // 加载配置
    async loadConfig() {
        try {
            const response = await fetch(`${this.apiBase}/config`);
            if (response.ok) {
                this.config = await response.json();
                this.debugMode = this.config.debug || false; // 从配置读取debug状态
                console.log('配置加载成功:', this.config);
            }
        } catch (error) {
            console.error('配置加载失败:', error);
        }
    }

    // 解析路由器信息
    parseRouterInfo() {
        this.routerInfo = {
            template: this.urlParams.template || 'custom',
            timestamp: this.urlParams.timestamp || '',
            router_ver: this.urlParams.router_ver || '',
            firmware: this.urlParams.firmware || '',
            gwid: this.urlParams.gwid || ''
        };
        
        this.gwid = this.routerInfo.gwid;
        
        console.log('路由器信息:', this.routerInfo);
    }

    // 显示路由器信息（仅控制台日志，不在界面显示）
    displayRouterInfo() {
        // 仅在控制台记录路由器信息，用于调试
        if (this.routerInfo.router_ver || this.routerInfo.firmware) {
            console.log(`路由器信息: ${this.routerInfo.firmware} ${this.routerInfo.router_ver}`);
        }
        if (this.gwid) {
            console.log(`网关ID: ${this.gwid}`);
        }
    }

    // 获取设备信息
    async getDeviceInfo() {
        try {
            // 从 URL 参数获取用户内网IP
            this.userIP = this.urlParams.user_ip || '';
            
            // 获取 MAC 地址（从 URL 参数）
            this.userMAC = this.urlParams.mac || '';
            
            // 更新页面显示
            document.getElementById('userIP').textContent = this.userIP || '未知';
            document.getElementById('userMAC').textContent = this.userMAC || '未获取';
            
            // 获取公网IP
            this.getPublicIP();
            
            // 添加调试日志
            this.addDebugLog(`设备信息获取 - IP: ${this.userIP || '空'}, MAC: ${this.userMAC || '空'}`, 'info');
            console.log('设备信息 - IP:', this.userIP, 'MAC:', this.userMAC);
            
        } catch (error) {
            console.error('获取设备信息失败:', error);
            this.showMessage('获取设备信息失败', 'error');
        }
    }

    // 获取公网IP地址
    async getPublicIP() {
        const publicIPElement = document.getElementById('publicIP');
        try {
            publicIPElement.textContent = '查询中...';
            
            // 尝试多个公网IP查询服务
            const apis = [
                'https://api.ipify.org?format=json',
                'https://api64.ipify.org?format=json',
                'https://api.ip.sb/jsonip'
            ];
            
            for (const api of apis) {
                try {
                    const response = await fetch(api, { timeout: 3000 });
                    const data = await response.json();
                    const ip = data.ip || data.query;
                    
                    if (ip) {
                        publicIPElement.textContent = ip;
                        publicIPElement.style.color = '#52c41a';
                        this.addDebugLog(`公网IP: ${ip}`, 'info');
                        return;
                    }
                } catch (err) {
                    console.log(`API ${api} 失败，尝试下一个...`);
                }
            }
            
            throw new Error('所有API都失败');
            
        } catch (error) {
            console.error('获取公网IP失败:', error);
            publicIPElement.textContent = '查询失败';
            publicIPElement.style.color = '#ff4d4f';
        }
    }



    // 绑定事件
    bindEvents() {
        const form = document.getElementById('authForm');
        form.addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleAuth();
        });

        // 添加键盘快捷键支持
        document.addEventListener('keydown', (e) => {
            // Enter 键提交表单
            if (e.key === 'Enter' && !document.getElementById('submitBtn').disabled) {
                e.preventDefault();
                this.handleAuth();
            }
            
            // Escape 键清空消息
            if (e.key === 'Escape') {
                document.getElementById('message').innerHTML = '';
            }
        });

        // 输入框聚焦效果
        const inputs = document.querySelectorAll('.ant-input');
        inputs.forEach(input => {
            input.addEventListener('focus', () => {
                input.parentElement.style.transform = 'scale(1.02)';
                input.parentElement.style.transition = 'transform 0.2s ease';
            });
            
            input.addEventListener('blur', () => {
                input.parentElement.style.transform = 'scale(1)';
            });
        });
    }

    // 处理认证（Ant Design 风格）
    async handleAuth() {
        const submitBtn = document.getElementById('submitBtn');
        const loading = document.getElementById('loading');
        const messageDiv = document.getElementById('message');

        try {
            this.addDebugLog('开始认证流程...', 'info');
            
            // 设置加载状态
            submitBtn.disabled = true;
            submitBtn.classList.add('ant-btn-loading');
            submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin" style="margin-right: 8px;"></i>认证中...';
            
            loading.style.display = 'block';
            messageDiv.innerHTML = '';

            // 收集表单数据
            const formData = this.collectFormData();
            this.addDebugLog(`表单数据: ${JSON.stringify(formData, null, 2)}`, 'info');
            
            // 验证必需字段
            if (!this.validateForm(formData)) {
                this.addDebugLog('表单验证失败', 'error');
                return;
            }

            // 调用同服务器的认证API（无需跨域）
            const apiUrl = `${this.apiBase}/auth/network`;
            this.addDebugLog(`请求URL: ${apiUrl}`, 'request');
            this.addDebugLog(`请求体: ${JSON.stringify(formData, null, 2)}`, 'request');
            
            const response = await fetch(apiUrl, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(formData)
            });

            this.addDebugLog(`HTTP状态码: ${response.status}`, 'response');
            
            const result = await response.json();
            this.addDebugLog(`后端响应: ${JSON.stringify(result, null, 2)}`, 'response');

            if (result.success) {
                this.addDebugLog('认证成功！', 'success');
                this.showMessage('认证成功！准备跳转...', 'success');

                // 保存用户名
                this.saveUsername();

                // 使用后端返回的放行URL
                const releaseURL = result.data.release_url;
                
                this.addDebugLog(`生成放行URL: ${releaseURL}`, 'success');
                
                this.addDebugLog(`生成放行URL: ${releaseURL}`, 'success');
                
                // 构建成功页面URL，传递用户信息、放行URL和调试模式标志
                const username = formData.user_account || formData.phone || '未知';
                const successUrl = `/success?username=${encodeURIComponent(username)}&ip=${encodeURIComponent(formData.ip_address)}&mac=${encodeURIComponent(formData.mac_address)}&release_url=${encodeURIComponent(releaseURL)}&debug=${this.debugMode ? '1' : '0'}`;
                
                this.addDebugLog(`准备跳转到成功页面...`, 'success');
                setTimeout(() => {
                    window.location.href = successUrl;
                }, 2000);
            } else {
                this.addDebugLog(`认证失败: ${result.message}`, 'error');
                this.showMessage(`认证失败: ${result.message}`, 'error');
            }

        } catch (error) {
            this.addDebugLog(`认证过程出错: ${error.message}`, 'error');
            this.addDebugLog(`错误堆栈: ${error.stack}`, 'error');
            this.showMessage('认证过程出错，请重试', 'error');
        } finally {
            // 恢复提交按钮状态
            submitBtn.disabled = false;
            submitBtn.classList.remove('ant-btn-loading');
            submitBtn.innerHTML = '<i class="fas fa-sign-in-alt" style="margin-right: 8px;"></i>连接网络';
            loading.style.display = 'none';
        }
    }

    // 收集表单数据（包含iKuai所需的所有参数）
    collectFormData() {
        // 获取表单输入的必填字段
        const deviceNumber = document.getElementById('deviceNumber')?.value.trim() || '';
        const vpnNumber = document.getElementById('vpnNumber')?.value.trim() || '';
        const phone = document.getElementById('phone')?.value.trim() || '';
        
        // 用户名密码字段（可能不存在，因为已隐藏）
        const username = document.getElementById('username')?.value.trim() || '';
        const password = document.getElementById('password')?.value || '';
        
        return {
            // API 认证必填字段（符合后端期望格式）
            device_number: deviceNumber,           // 设备编号
            user_account: phone || username,       // 使用人账号（优先使用手机号）
            vpn_number: vpnNumber,                 // VPN线路编号
            mac_address: this.userMAC,             // MAC地址
            ip_address: this.userIP,               // IP地址
            system_id: 1,                          // 路由系统ID（固定为1）
            
            // 额外信息（兼容字段）
            username: username,
            password: password,
            phone: phone,
            
            // 路由器信息（从URL参数获取）
            gwid: this.gwid,
            router_ver: this.routerInfo.router_ver,
            firmware: this.routerInfo.firmware,
            timestamp: this.routerInfo.timestamp,
            template: this.routerInfo.template
        };
    }

    // 验证表单
    validateForm(data) {
        // 验证设备信息（主要认证字段）
        if (!data.device_number) {
            this.addDebugLog('验证失败: 缺少设备编号', 'error');
            this.showMessage('请输入设备编号', 'error');
            return false;
        }

        if (!data.vpn_number) {
            this.addDebugLog('验证失败: 缺少VPN线路编号', 'error');
            this.showMessage('请输入VPN线路编号', 'error');
            return false;
        }

        if (!data.user_account) {
            this.addDebugLog('验证失败: 缺少用户账号', 'error');
            this.showMessage('请输入使用人手机号', 'error');
            return false;
        }

        if (!data.ip_address) {
            this.addDebugLog('验证失败: 缺少IP地址', 'error');
            this.showMessage('无法获取设备 IP 地址', 'error');
            return false;
        }

        if (!data.mac_address) {
            this.addDebugLog('验证失败: 缺少MAC地址', 'error');
            this.showMessage('无法获取设备 MAC 地址', 'error');
            return false;
        }

        this.addDebugLog('表单验证通过', 'success');
        // 用户名和密码是可选的，不需要验证
        return true;
    }

    // 在客户端发起放行请求（使用隐藏 iframe）
    releaseByClient(releaseURL) {
        console.log('� 客户端发起放行请求:', releaseURL);
        
        // 创建隐藏的 iframe
        const iframe = document.createElement('iframe');
        iframe.style.display = 'none';
        iframe.style.width = '0';
        iframe.style.height = '0';
        iframe.style.border = 'none';
        
        // 设置超时定时器
        const timeout = setTimeout(() => {
            console.log('✅ 放行请求超时，假设成功，准备跳转');
            this.navigateToSuccess();
        }, 3000); // 3秒后无论如何都跳转
        
        // 尝试监听 iframe 加载完成
        iframe.onload = () => {
            console.log('✅ 放行请求已发送');
            clearTimeout(timeout);
            // 延迟一下确保请求完成
            setTimeout(() => {
                this.navigateToSuccess();
            }, 1000);
        };
        
        iframe.onerror = () => {
            console.log('⚠️ iframe 加载出错，但放行请求可能已发送');
            clearTimeout(timeout);
            // 即使出错也跳转（因为跨域请求可能触发 onerror）
            setTimeout(() => {
                this.navigateToSuccess();
            }, 1000);
        };
        
        // 添加到页面并加载 URL
        document.body.appendChild(iframe);
        iframe.src = releaseURL;
        
        console.log('ℹ️ 已创建隐藏 iframe 发起放行请求');
    }
    
    // 跳转到成功页面
    navigateToSuccess() {
        const formData = this.collectFormData();
        const successUrl = `/success?username=${encodeURIComponent(formData.username)}&ip=${encodeURIComponent(formData.user_ip)}&mac=${encodeURIComponent(formData.mac)}`;
        console.log('🎉 跳转到成功页面:', successUrl);
        window.location.href = successUrl;
    }

    // 显示消息（Ant Design 风格）
    showMessage(message, type = 'info') {
        const messageDiv = document.getElementById('message');
        
        let iconClass = 'fas fa-info-circle';
        let alertClass = 'ant-alert';
        
        switch(type) {
            case 'success':
                iconClass = 'fas fa-check-circle';
                alertClass += ' ant-alert-success';
                break;
            case 'error':
                iconClass = 'fas fa-exclamation-circle';
                alertClass += ' ant-alert-error';
                break;
            case 'warning':
                iconClass = 'fas fa-exclamation-triangle';
                alertClass += ' ant-alert-warning';
                break;
            default:
                iconClass = 'fas fa-info-circle';
                alertClass += ' ant-alert-info';
        }
        
        messageDiv.innerHTML = `
            <div class="${alertClass}">
                <i class="${iconClass} ant-alert-icon"></i>
                ${message}
            </div>
        `;
        
        // 自动隐藏消息
        setTimeout(() => {
            messageDiv.innerHTML = '';
        }, type === 'success' ? 8088 : 5000);
    }

    // 格式化 MAC 地址
    formatMAC(mac) {
        if (!mac) return '';
        return mac.replace(/[:-]/g, '').match(/.{2}/g)?.join(':').toUpperCase() || mac;
    }

    // 保存用户名到本地存储
    saveUsername() {
        const usernameInput = document.getElementById('username');
        const rememberCheckbox = document.getElementById('remember');
        
        // 如果用户名输入框不存在（被隐藏），直接返回
        if (!usernameInput) {
            return;
        }
        
        const username = usernameInput.value;
        
        // 检查remember复选框是否存在
        const remember = rememberCheckbox ? rememberCheckbox.checked : true;
        
        if (remember && username) {
            localStorage.setItem('ikuai_remembered_username', username);
        } else {
            localStorage.removeItem('ikuai_remembered_username');
        }
    }

    // 加载保存的用户名
    loadSavedUsername() {
        const savedUsername = localStorage.getItem('ikuai_remembered_username');
        if (savedUsername) {
            const usernameInput = document.getElementById('username');
            const rememberCheckbox = document.getElementById('remember');
            
            if (usernameInput) {
                usernameInput.value = savedUsername;
            }
            
            if (rememberCheckbox) {
                rememberCheckbox.checked = true;
            }
            
            // 自动聚焦到密码框
            setTimeout(() => {
                const passwordInput = document.getElementById('password');
                if (passwordInput) {
                    passwordInput.focus();
                }
            }, 100);
        }
    }
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    new IKuaiAuth();
});

// 获取后端配置（缓存配置避免重复请求）
let cachedConfig = null;
async function getConfig() {
    if (cachedConfig) return cachedConfig;
    
    try {
        const response = await fetch(`${window.location.protocol}//${window.location.host}/api/config`);
        if (!response.ok) throw new Error('获取配置失败');
        cachedConfig = await response.json();
        return cachedConfig;
    } catch (error) {
        console.error('获取配置失败，使用默认值:', error);
        // 返回默认配置（不包含app_key，前端不需要）
        return {
            debug: false
        };
    }
}