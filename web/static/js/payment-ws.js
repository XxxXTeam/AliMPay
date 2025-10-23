/**
 * payment-ws.js - 支付页面WebSocket增强脚本
 * Author: AliMPay Team
 * Description: 使用WebSocket实时获取订单支付状态，替代HTTP轮询
 * 
 * 功能:
 *   - WebSocket连接管理（自动重连）
 *   - 实时接收订单状态更新
 *   - 倒计时管理
 *   - 设备检测和适配
 *   - Toast提示组件
 */

/*
全局状态管理
*/
const PaymentState = {
    ws: null,              // WebSocket连接
    orderId: null,         // 订单号
    reconnectTimer: null,  // 重连定时器
    reconnectAttempts: 0,  // 重连尝试次数
    maxReconnectAttempts: 5, // 最大重连次数
    isConnected: false,    // 连接状态
    countdownTimer: null,  // 倒计时定时器
    timeLeft: 300         // 剩余时间(秒)
};

/*
设备检测工具类
功能: 检测用户设备类型和浏览器环境
*/
const DeviceDetector = {
    /** 检测是否为移动设备 */
    isMobile: function() {
        return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);
    },
    
    /** 检测是否在微信中 */
    isWeChat: function() {
        return /micromessenger/i.test(navigator.userAgent);
    },
    
    /** 检测是否在支付宝中 */
    isAlipay: function() {
        return /AlipayClient/i.test(navigator.userAgent);
    },
    
    /** 获取设备类型描述 */
    getDeviceType: function() {
        if (this.isMobile()) {
            return this.isWeChat() ? 'WeChat' : 
                   this.isAlipay() ? 'Alipay' : 'Mobile';
        }
        return 'Desktop';
    }
};

/*
WebSocket管理器
功能: 管理WebSocket连接、自动重连、消息处理
*/
const WebSocketManager = {
    /**
     * 连接WebSocket
     * @param {string} orderId - 订单号
     */
    connect: function(orderId) {
        if (PaymentState.ws && PaymentState.ws.readyState === WebSocket.OPEN) {
            console.log('[WS] Already connected');
            return;
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws/order?order_id=${orderId}`;

        console.log('[WS] Connecting to:', wsUrl);
        showToast('正在连接实时推送服务...', 'info', 1500);

        try {
            PaymentState.ws = new WebSocket(wsUrl);
            
            PaymentState.ws.onopen = this.onOpen.bind(this);
            PaymentState.ws.onmessage = this.onMessage.bind(this);
            PaymentState.ws.onerror = this.onError.bind(this);
            PaymentState.ws.onclose = this.onClose.bind(this);
        } catch (error) {
            console.error('[WS] Connection error:', error);
            this.fallbackToPolling();
        }
    },

    /**
     * 连接打开回调
     */
    onOpen: function() {
        console.log('[WS] Connected successfully');
        PaymentState.isConnected = true;
        PaymentState.reconnectAttempts = 0;
        updateConnectionStatus(true);
        showToast('实时推送已连接', 'success', 1500);
    },

    /**
     * 接收消息回调
     * @param {MessageEvent} event - 消息事件
     */
    onMessage: function(event) {
        try {
            const data = JSON.parse(event.data);
            console.log('[WS] Message received:', data);

            if (data.type === 'status_update') {
                handleStatusUpdate(data);
            }
        } catch (error) {
            console.error('[WS] Message parse error:', error);
        }
    },

    /**
     * 错误回调
     * @param {Event} error - 错误事件
     */
    onError: function(error) {
        console.error('[WS] Error:', error);
        updateConnectionStatus(false);
    },

    /**
     * 连接关闭回调
     * @param {CloseEvent} event - 关闭事件
     */
    onClose: function(event) {
        console.log('[WS] Connection closed:', event.code, event.reason);
        PaymentState.isConnected = false;
        updateConnectionStatus(false);

        // 自动重连
        if (PaymentState.reconnectAttempts < PaymentState.maxReconnectAttempts) {
            const delay = Math.min(1000 * Math.pow(2, PaymentState.reconnectAttempts), 30000);
            console.log(`[WS] Reconnecting in ${delay}ms...`);
            
            PaymentState.reconnectTimer = setTimeout(() => {
                PaymentState.reconnectAttempts++;
                this.connect(PaymentState.orderId);
            }, delay);
        } else {
            console.log('[WS] Max reconnect attempts reached, falling back to polling');
            showToast('实时推送连接失败，切换到轮询模式', 'warning', 3000);
            this.fallbackToPolling();
        }
    },

    /**
     * 关闭连接
     */
    close: function() {
        if (PaymentState.ws) {
            PaymentState.ws.close();
            PaymentState.ws = null;
        }
        if (PaymentState.reconnectTimer) {
            clearTimeout(PaymentState.reconnectTimer);
            PaymentState.reconnectTimer = null;
        }
    },

    /**
     * 降级到HTTP轮询
     */
    fallbackToPolling: function() {
        console.log('[Polling] Starting HTTP polling fallback');
        startPolling();
    }
};

/*
处理订单状态更新
@param {Object} data - 状态更新数据
*/
function handleStatusUpdate(data) {
    console.log('[Status] Update received:', data);

    const statusIndicator = document.getElementById('statusIndicator');
    const statusText = document.querySelector('.status-text');

    if (data.status === 1) {
        // 支付成功
        statusIndicator.classList.remove('checking');
        statusIndicator.classList.add('success');
        statusText.textContent = '✓ 支付成功！';

        // 停止倒计时
        stopCountdown();

        // 显示成功提示
        showToast('支付成功！正在跳转...', 'success', 2000);

        // 关闭WebSocket
        WebSocketManager.close();

        // 延迟跳转
        setTimeout(() => {
            const returnUrl = document.querySelector('[data-return-url]')?.getAttribute('data-return-url');
            if (returnUrl) {
                window.location.href = returnUrl;
            } else {
                showToast('支付完成', 'success');
            }
        }, 2000);
    }
}

/*
更新连接状态显示
@param {boolean} connected - 是否已连接
*/
function updateConnectionStatus(connected) {
    const indicator = document.getElementById('statusIndicator');
    if (!indicator) return;

    if (connected) {
        indicator.style.borderColor = '#52c41a';
    } else {
        indicator.style.borderColor = '#faad14';
    }
}

/*
HTTP轮询 (降级方案)
*/
function startPolling() {
    const pid = document.querySelector('[data-pid]')?.getAttribute('data-pid');
    const outTradeNo = document.querySelector('[data-out-trade-no]')?.getAttribute('data-out-trade-no');

    if (!pid || !outTradeNo) return;

    const poll = setInterval(async () => {
        try {
            const response = await fetch(`/api/order?pid=${pid}&out_trade_no=${outTradeNo}`);
            const data = await response.json();

            if (data.status === 1) {
                clearInterval(poll);
                handleStatusUpdate({ status: 1, type: 'status_update' });
            }
        } catch (error) {
            console.error('[Polling] Error:', error);
        }
    }, 3000);

    // 保存定时器ID用于清理
    PaymentState.pollingTimer = poll;
}

/*
倒计时管理
*/
function startCountdown() {
    const countdownEl = document.getElementById('countdownTime');
    if (!countdownEl) return;

    const createTimeStr = document.querySelector('[data-create-time]')?.getAttribute('data-create-time');
    if (createTimeStr) {
        const createTime = new Date(createTimeStr.replace(' ', 'T'));
        const now = new Date();
        const elapsed = Math.floor((now - createTime) / 1000);
        PaymentState.timeLeft = Math.max(0, 300 - elapsed);
    }

    PaymentState.countdownTimer = setInterval(() => {
        if (PaymentState.timeLeft <= 0) {
            stopCountdown();
            countdownEl.textContent = '已过期';
            countdownEl.style.color = '#ff4d4f';
            showToast('订单已过期', 'error', 3000);
            return;
        }

        const minutes = Math.floor(PaymentState.timeLeft / 60);
        const seconds = PaymentState.timeLeft % 60;
        countdownEl.textContent = `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
        PaymentState.timeLeft--;
    }, 1000);
}

function stopCountdown() {
    if (PaymentState.countdownTimer) {
        clearInterval(PaymentState.countdownTimer);
        PaymentState.countdownTimer = null;
    }
}

/*
拉起支付宝APP
功能: 在移动端调用支付宝URL Scheme拉起APP
*/
function launchAlipay() {
    if (!DeviceDetector.isMobile()) {
        showToast('请使用手机扫描二维码支付', 'warning');
        return;
    }

    const qrCodeId = document.querySelector('[data-qrcode-id]')?.getAttribute('data-qrcode-id');
    const amount = document.querySelector('[data-amount]')?.getAttribute('data-amount');
    const tradeNo = document.querySelector('[data-trade-no]')?.getAttribute('data-trade-no');

    if (!qrCodeId) {
        showToast('系统配置错误：缺少收款码ID', 'error');
        return;
    }

    if (DeviceDetector.isWeChat()) {
        showToast('请点击右上角，选择"在浏览器中打开"', 'info', 3000);
        return;
    }

    const alipayUrl = encodeURIComponent(`https://qr.alipay.com/${qrCodeId}?amount=${amount}&remark=${tradeNo}`);
    const scheme = `alipays://platformapi/startapp?saId=10000007&url=${alipayUrl}`;

    console.log('[Alipay] Launching with scheme:', scheme);
    showToast('正在打开支付宝...', 'success');

    window.location.href = scheme;
}

/*
Toast提示组件
@param {string} message - 提示内容
@param {string} type - 类型: success/error/warning/info
@param {number} duration - 显示时长(ms)
*/
function showToast(message, type = 'info', duration = 2000) {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.textContent = message;
    toast.style.cssText = `
        position: fixed;
        top: 20px;
        left: 50%;
        transform: translateX(-50%);
        padding: 12px 24px;
        border-radius: 8px;
        color: white;
        font-size: 14px;
        z-index: 10000;
        animation: slideDown 0.3s ease;
        box-shadow: 0 4px 12px rgba(0,0,0,0.15);
    `;

    const colors = {
        success: '#52c41a',
        error: '#ff4d4f',
        warning: '#faad14',
        info: '#1677ff'
    };
    toast.style.backgroundColor = colors[type] || colors.info;

    document.body.appendChild(toast);

    setTimeout(() => {
        toast.style.animation = 'slideUp 0.3s ease';
        setTimeout(() => toast.remove(), 300);
    }, duration);
}

/*
页面初始化
功能: 页面加载时执行的初始化逻辑
*/
function initPaymentPage() {
    const deviceType = DeviceDetector.getDeviceType();
    const launchBtn = document.getElementById('alipayLaunchBtn');
    
    console.log('[Device] Type:', deviceType);

    // 处理拉起支付宝按钮
    if (launchBtn) {
        if (DeviceDetector.isMobile()) {
            launchBtn.parentElement.style.display = 'block';
            
            if (DeviceDetector.isWeChat()) {
                launchBtn.innerHTML = `
                    <svg class="alipay-icon" viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg" width="20" height="20">
                        <path d="M1024 701.9v202.8c0 66.6-53.9 120.4-120.4 120.4H120.4C53.9 1025.1 0 971.3 0 904.7V120.4C0 53.9 53.9 0 120.4 0h783.1c66.6 0 120.4 53.9 120.4 120.4V701.9z" fill="#00A0E9"/>
                        <path d="M928.9 735.7c-99.7-47.4-244.8-110.9-325.6-146.5 21.9-36.3 39.3-75.8 51.6-117.6H546v-64.3h199.4v-38.7H546v-96.8h-38.7c0 0 0 0 0 0H444.2v96.8H244.8v38.7h199.4v64.3H335.3c-32.3 116.5-103.9 217.4-203.5 289.2 51.6 39.3 122.5 72.6 171.1 90.6 90.6-77.4 154.8-184.5 184.5-315.5h258.1c-19.4 64.3-45.2 125.8-77.4 181.3 38.7 16.1 141.9 58.1 225.8 96.8V735.7z" fill="#FFFFFF"/>
                    </svg>
                    <span>在浏览器中打开</span>
                `;
            }
        } else {
            launchBtn.parentElement.innerHTML = `
                <div class="pc-scan-tip">
                    <div style="font-size: 16px; margin-bottom: 8px;">💻 电脑端访问</div>
                    <div style="font-size: 14px; color: #00000073;">
                        请使用手机扫描上方二维码完成支付
                    </div>
                </div>
            `;
        }
    }

    // 添加动画样式
    const style = document.createElement('style');
    style.textContent = `
        @keyframes slideDown {
            from { transform: translateX(-50%) translateY(-20px); opacity: 0; }
            to { transform: translateX(-50%) translateY(0); opacity: 1; }
        }
        @keyframes slideUp {
            from { transform: translateX(-50%) translateY(0); opacity: 1; }
            to { transform: translateX(-50%) translateY(-20px); opacity: 0; }
        }
        .pc-scan-tip {
            text-align: center;
            padding: 20px;
            background: #f5f5f5;
            border-radius: 12px;
            color: #000000d9;
        }
    `;
    document.head.appendChild(style);

    // 获取订单ID并连接WebSocket
    const orderId = document.querySelector('[data-order-id]')?.getAttribute('data-order-id');
    if (orderId) {
        PaymentState.orderId = orderId;
        WebSocketManager.connect(orderId);
    } else {
        console.error('[Init] No order ID found');
        startPolling(); // 降级到轮询
    }

    // 启动倒计时
    startCountdown();
}

// 页面加载完成后初始化
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initPaymentPage);
} else {
    initPaymentPage();
}

// 页面卸载时清理
window.addEventListener('beforeunload', () => {
    WebSocketManager.close();
    stopCountdown();
    if (PaymentState.pollingTimer) {
        clearInterval(PaymentState.pollingTimer);
    }
});

