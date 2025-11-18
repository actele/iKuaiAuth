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
        this.init();
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
        this.parseRouterInfo();
        await this.getDeviceInfo();
        this.bindEvents();
        this.loadSavedUsername();
        this.displayRouterInfo();
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
            // 优先从 URL 参数获取用户 IP
            this.userIP = this.urlParams.user_ip || '';
            
            // 如果URL参数没有IP，尝试其他方式
            if (!this.userIP) {
                this.userIP = await this.getUserIPFromAPI();
            }
            
            // 获取 MAC 地址（从 URL 参数）
            this.userMAC = this.urlParams.mac || '';
            
            // 更新页面显示
            document.getElementById('userIP').textContent = this.userIP || '未知';
            document.getElementById('userMAC').textContent = this.userMAC || '未获取';
            
            console.log('设备信息 - IP:', this.userIP, 'MAC:', this.userMAC);
            
        } catch (error) {
            console.error('获取设备信息失败:', error);
            this.showMessage('获取设备信息失败', 'error');
        }
    }

    // 从API获取用户 IP 地址（备用方案）
    async getUserIPFromAPI() {
        try {
            const response = await fetch('https://api.ipify.org?format=json');
            const data = await response.json();
            return data.ip;
        } catch (error) {
            console.error('获取 IP 失败:', error);
            return '';
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
            // 设置加载状态
            submitBtn.disabled = true;
            submitBtn.classList.add('ant-btn-loading');
            submitBtn.innerHTML = '<i class="fas fa-spinner fa-spin" style="margin-right: 8px;"></i>认证中...';
            
            loading.style.display = 'block';
            messageDiv.innerHTML = '';

            // 收集表单数据
            const formData = this.collectFormData();
            console.log('📋 发送认证请求:', formData);
            
            // 验证必需字段
            if (!this.validateForm(formData)) {
                return;
            }

            // 发送认证请求
            console.log(`🌐 请求URL: ${this.apiBase}/auth`);
            const response = await fetch(`${this.apiBase}/auth`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(formData)
            });

            console.log(`📡 HTTP状态码: ${response.status}`);
            const result = await response.json();
            console.log('📥 服务器响应:', result);

            if (result.success) {
                console.log('✅ 认证成功');
                this.showMessage('认证成功！准备跳转...', 'success');

                // 保存用户名
                this.saveUsername();

                // 获取放行 URL 并跳转到成功页面（不自动发起请求）
                const releaseURL = result.data.release_url;
                const debugMode = result.debug || false; // 获取调试模式标志
                const formData = this.collectFormData();
                
                // 构建成功页面URL，传递用户信息、放行URL和调试模式标志
                const successUrl = `/success?username=${encodeURIComponent(formData.username)}&ip=${encodeURIComponent(formData.user_ip)}&mac=${encodeURIComponent(formData.mac)}&release_url=${encodeURIComponent(releaseURL)}&debug=${debugMode ? '1' : '0'}`;
                
                console.log('🎉 跳转到成功页面，调试模式:', debugMode);
                setTimeout(() => {
                    window.location.href = successUrl;
                }, 1000);
            } else {
                console.log('❌ 认证失败:', result.message);
                this.showMessage(`认证失败: ${result.message}`, 'error');
            }

        } catch (error) {
            console.error('❌ 认证过程出错:', error);
            console.error('错误详情:', error.message);
            console.error('错误堆栈:', error.stack);
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
        return {
            // 用户输入
            username: document.getElementById('username').value.trim(),
            password: document.getElementById('password').value,
            
            // 设备信息（从URL参数获取）
            user_ip: this.userIP,
            mac: this.userMAC,
            
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
        if (!data.username) {
            this.showMessage('请输入用户名', 'error');
            return false;
        }

        if (!data.password) {
            this.showMessage('请输入密码', 'error');
            return false;
        }

        if (!data.user_ip) {
            this.showMessage('无法获取设备 IP 地址', 'error');
            return false;
        }

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
        const username = document.getElementById('username').value;
        const rememberCheckbox = document.getElementById('remember');
        
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

// 处理页面可见性变化（用户可能切换到其他应用）
document.addEventListener('visibilitychange', () => {
    if (!document.hidden) {
        // 页面重新可见时，可以检查网络状态
        console.log('页面重新可见');
    }
});

// 工具函数
const Utils = {
    // URL 参数解析
    getURLParams() {
        const params = {};
        const urlParams = new URLSearchParams(window.location.search);
        for (const [key, value] of urlParams) {
            params[key] = value;
        }
        return params;
    },

    // 设备检测
    isMobile() {
        return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);
    },

    // 时间格式化
    formatTime(timestamp) {
        return new Date(timestamp * 1000).toLocaleString('zh-CN');
    }
};

// 调试模式
if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    console.log('调试模式已启用');
    
    // 显示更多调试信息
    window.debugInfo = () => {
        console.log('URL 参数:', Utils.getURLParams());
        console.log('用户代理:', navigator.userAgent);
        console.log('是否移动设备:', Utils.isMobile());
    };
    
    // 自动填充测试数据
    setTimeout(() => {
        if (document.getElementById('username').value === '') {
            document.getElementById('username').value = 'test';
            document.getElementById('password').value = 'test123';
        }
    }, 1000);
}