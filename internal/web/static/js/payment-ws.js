/*
支付页面WebSocket客户端
功能:
  - 实时订单状态更新
  - 自动重连机制
  - HTTP轮询降级
  - 倒计时管理
  - Toast通知

使用示例:
  <script src="/static/js/payment-ws.js"></script>
  页面需包含以下元素:
    - [data-trade-no]: 订单号
    - [data-pid]: 商户ID
    - #statusIndicator: 状态指示器
    - #countdownTime: 倒计时显示
*/

(function() {
    'use strict';

    // 配置
    const CONFIG = {
        WS_RECONNECT_ATTEMPTS: 5,
        WS_RECONNECT_INTERVAL: 1000,
        WS_MAX_RECONNECT_INTERVAL: 30000,
        HTTP_POLL_INTERVAL: 3000,
        COUNTDOWN_TOTAL: 300, // 5分钟
        REDIRECT_DELAY: 2000
    };

    // 状态管理
    const state = {
        orderId: null,
        pid: null,
        ws: null,
        reconnectAttempts: 0,
        polling: false,
        pollTimer: null,
        countdownTimer: null,
        timeLeft: CONFIG.COUNTDOWN_TOTAL,
        paid: false
    };

    // DOM元素
    const elements = {};

    /*
    初始化应用
    */
    function init() {
        console.log('[Payment WS] Initializing...');

        // 获取订单信息
        const orderEl = document.querySelector('[data-trade-no]');
        const pidEl = document.querySelector('[data-pid]');
        
        if (!orderEl || !pidEl) {
            console.error('[Payment WS] Required elements not found');
            return;
        }

        state.orderId = orderEl.getAttribute('data-trade-no');
        state.pid = pidEl.getAttribute('data-pid');

        // 获取DOM元素
        elements.statusIndicator = document.getElementById('statusIndicator');
        elements.statusText = elements.statusIndicator?.querySelector('.status-text');
        elements.countdownTime = document.getElementById('countdownTime');
        elements.qrCode = document.getElementById('paymentQRCode');

        console.log('[Payment WS] Order:', state.orderId, 'PID:', state.pid);

        // 启动WebSocket
        connectWebSocket();

        // 启动倒计时
        startCountdown();

        // 页面可见性检测
        document.addEventListener('visibilitychange', handleVisibilityChange);
    }

    /*
    连接WebSocket
    */
    function connectWebSocket() {
        if (state.ws && (state.ws.readyState === WebSocket.OPEN || state.ws.readyState === WebSocket.CONNECTING)) {
            console.log('[Payment WS] Already connected or connecting');
            return;
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsURL = `${protocol}//${window.location.host}/ws/order?order_id=${state.orderId}`;
        
        console.log('[Payment WS] Connecting to:', wsURL);
        state.ws = new WebSocket(wsURL);

        state.ws.onopen = handleWSOpen;
        state.ws.onmessage = handleWSMessage;
        state.ws.onclose = handleWSClose;
        state.ws.onerror = handleWSError;
    }

    /*
    WebSocket打开事件
    */
    function handleWSOpen() {
        console.log('[Payment WS] Connected successfully');
        state.reconnectAttempts = 0;
        updateStatus('checking', '正在等待支付...');
        
        // 停止HTTP轮询（如果有）
        if (state.polling) {
            stopPolling();
        }
        
        showToast('✅ 实时连接已建立', 'success', 2000);
    }

    /*
    WebSocket消息事件
    */
    function handleWSMessage(event) {
        try {
            const data = JSON.parse(event.data);
            console.log('[Payment WS] Received:', data);

            if (data.type === 'status_update' && data.order_id === state.orderId) {
                if (data.status === 1 && !state.paid) {
                    handlePaymentSuccess(data);
                }
            }
        } catch (e) {
            console.error('[Payment WS] Parse error:', e);
        }
    }

    /*
    WebSocket关闭事件
    */
    function handleWSClose(event) {
        console.warn('[Payment WS] Disconnected:', event.code, event.reason);

        if (state.paid) {
            console.log('[Payment WS] Already paid, not reconnecting');
            return;
        }

        if (state.reconnectAttempts < CONFIG.WS_RECONNECT_ATTEMPTS) {
            state.reconnectAttempts++;
            const delay = Math.min(
                CONFIG.WS_RECONNECT_INTERVAL * Math.pow(2, state.reconnectAttempts - 1),
                CONFIG.WS_MAX_RECONNECT_INTERVAL
            );
            
            console.log(`[Payment WS] Reconnecting in ${delay}ms... (${state.reconnectAttempts}/${CONFIG.WS_RECONNECT_ATTEMPTS})`);
            showToast(`🔄 连接断开，${delay / 1000}秒后重连...`, 'warning', 2000);
            
            setTimeout(connectWebSocket, delay);
        } else {
            console.warn('[Payment WS] Max reconnect attempts reached, falling back to HTTP polling');
            showToast('⚠️ 实时推送不可用，已切换为轮询模式', 'warning', 3000);
            fallbackToPolling();
        }
    }

    /*
    WebSocket错误事件
    */
    function handleWSError(error) {
        console.error('[Payment WS] Error:', error);
        // onclose会被触发，在那里处理重连
    }

    /*
    降级到HTTP轮询
    */
    function fallbackToPolling() {
        if (state.polling) {
            return;
        }

        console.log('[Payment WS] Starting HTTP polling');
        state.polling = true;
        updateStatus('checking', '正在轮询支付状态...');

        // 立即检查一次
        checkOrderStatus();

        // 定期轮询
        state.pollTimer = setInterval(checkOrderStatus, CONFIG.HTTP_POLL_INTERVAL);
    }

    /*
    停止HTTP轮询
    */
    function stopPolling() {
        if (!state.polling) {
            return;
        }

        console.log('[Payment WS] Stopping HTTP polling');
        state.polling = false;

        if (state.pollTimer) {
            clearInterval(state.pollTimer);
            state.pollTimer = null;
        }
    }

    /*
    HTTP方式检查订单状态
    */
    function checkOrderStatus() {
        if (state.paid) {
            stopPolling();
            return;
        }

        const url = `/api?act=order&pid=${state.pid}&trade_no=${state.orderId}`;
        console.log('[Payment HTTP] Checking:', url);

        fetch(url)
            .then(res => res.json())
            .then(data => {
                console.log('[Payment HTTP] Order status:', data);
                
                if (data.code === 1 && data.status === 1) {
                    handlePaymentSuccess(data);
                }
            })
            .catch(err => {
                console.error('[Payment HTTP] Check failed:', err);
            });
    }

    /*
    处理支付成功
    */
    function handlePaymentSuccess(data) {
        if (state.paid) {
            return;
        }

        state.paid = true;
        console.log('[Payment] 🎉 Payment successful!', data);

        // 停止所有定时器
        stopCountdown();
        stopPolling();
        
        // 关闭WebSocket
        if (state.ws) {
            state.ws.close();
        }

        // 更新UI
        updateStatus('success', '✅ 支付成功！页面即将跳转...');
        showToast('💰 支付成功！', 'success', 3000);

        // 状态指示器变绿
        if (elements.statusIndicator) {
            elements.statusIndicator.style.background = 'linear-gradient(135deg, #52c41a 0%, #73d13d 100%)';
            elements.statusIndicator.style.color = '#fff';
            elements.statusIndicator.style.transform = 'scale(1.05)';
        }

        // 延迟跳转
        setTimeout(() => {
            // 优先使用return_url，否则使用默认返回页面
            const returnUrl = getReturnURL();
            if (returnUrl) {
                window.location.href = returnUrl;
            } else {
                window.location.href = `/return?trade_no=${state.orderId}`;
            }
        }, CONFIG.REDIRECT_DELAY);
    }

    /*
    启动倒计时
    */
    function startCountdown() {
        if (state.countdownTimer) {
            clearInterval(state.countdownTimer);
        }

        state.timeLeft = CONFIG.COUNTDOWN_TOTAL;
        updateCountdownDisplay();

        state.countdownTimer = setInterval(() => {
            state.timeLeft--;
            updateCountdownDisplay();

            if (state.timeLeft <= 0) {
                handleCountdownExpired();
            }
        }, 1000);
    }

    /*
    停止倒计时
    */
    function stopCountdown() {
        if (state.countdownTimer) {
            clearInterval(state.countdownTimer);
            state.countdownTimer = null;
        }
    }

    /*
    更新倒计时显示
    */
    function updateCountdownDisplay() {
        if (!elements.countdownTime) {
            return;
        }

        const minutes = Math.floor(state.timeLeft / 60);
        const seconds = state.timeLeft % 60;
        elements.countdownTime.textContent = `${minutes}:${seconds < 10 ? '0' : ''}${seconds}`;

        // 最后30秒变红
        if (state.timeLeft <= 30 && state.timeLeft > 0) {
            elements.countdownTime.style.color = '#ff4d4f';
            elements.countdownTime.style.fontWeight = 'bold';
        }
    }

    /*
    倒计时到期
    */
    function handleCountdownExpired() {
        console.log('[Payment] ⏰ Countdown expired');
        
        stopCountdown();
        stopPolling();

        if (state.ws) {
            state.ws.close();
        }

        updateStatus('error', '⏰ 订单已超时，请重新下单');
        showToast('订单已超时', 'error', 5000);

        // 禁用二维码
        if (elements.qrCode) {
            elements.qrCode.style.opacity = '0.3';
            elements.qrCode.style.filter = 'grayscale(100%)';
        }
    }

    /*
    页面可见性变化
    */
    function handleVisibilityChange() {
        if (document.visibilityState === 'visible' && !state.paid) {
            console.log('[Payment WS] 📱 Page visible, checking connection...');
            
            // 如果WebSocket断开，尝试重连
            if (!state.ws || state.ws.readyState !== WebSocket.OPEN) {
                if (state.reconnectAttempts < CONFIG.WS_RECONNECT_ATTEMPTS) {
                    state.reconnectAttempts = 0; // 重置重连次数
                    connectWebSocket();
                } else if (!state.polling) {
                    // WebSocket已失败，确保轮询在运行
                    fallbackToPolling();
                }
            }
            
            // 无论如何都检查一次状态
            if (state.polling) {
                checkOrderStatus();
            }
        }
    }

    /*
    更新状态显示
    */
    function updateStatus(type, message) {
        if (!elements.statusIndicator || !elements.statusText) {
            return;
        }

        elements.statusIndicator.className = `status-indicator ${type}`;
        elements.statusText.textContent = message;
    }

    /*
    显示Toast通知
    */
    function showToast(message, type = 'info', duration = 3000) {
        // 检查是否有全局toast函数
        if (typeof window.showToast === 'function') {
            window.showToast(message, type, duration);
            return;
        }

        // 简单实现
        console.log(`[Toast ${type}]`, message);
        
        const toast = document.createElement('div');
        toast.className = `toast toast-${type} slide-in-right`;
        
        const bgColors = {
            success: '#52c41a',
            error: '#ff4d4f',
            warning: '#faad14',
            info: '#1677ff'
        };
        
        toast.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 12px 20px;
            background: ${bgColors[type] || bgColors.info};
            color: white;
            border-radius: 8px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
            z-index: 10000;
            font-size: 14px;
            animation: slideInRight 0.3s ease;
            max-width: 300px;
        `;

        toast.textContent = message;
        document.body.appendChild(toast);

        setTimeout(() => {
            toast.style.animation = 'fadeOut 0.3s ease';
            setTimeout(() => toast.remove(), 300);
        }, duration);
    }

    /*
    获取返回URL
    */
    function getReturnURL() {
        const urlParams = new URLSearchParams(window.location.search);
        return urlParams.get('return_url') || '';
    }

    // 页面加载完成后初始化
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

    // 导出到全局（供调试使用）
    window.PaymentWS = {
        state,
        reconnect: connectWebSocket,
        checkStatus: checkOrderStatus,
        getState: () => ({ ...state })
    };
})();
