/*
管理后台WebSocket客户端
功能：
  - 实时订单状态更新
  - 新订单通知
  - 订单支付通知
  - 自动重连机制
  - 消息通知（浏览器通知API）
  - 音效提醒

使用示例：
  AdminWebSocket.init();
  AdminWebSocket.on('order_paid', (order) => {
      console.log('订单支付:', order);
  });
*/

const AdminWebSocket = (function() {
    let ws = null;
    let reconnectAttempts = 0;
    const maxReconnectAttempts = 10;
    let reconnectInterval = 1000;
    const maxReconnectInterval = 30000;
    const eventHandlers = {};
    
    /*
    播放通知音效
    @param type {String} 通知类型 (success/info/warning)
    */
    function playNotificationSound(type) {
        const audioContext = new (window.AudioContext || window.webkitAudioContext)();
        const oscillator = audioContext.createOscillator();
        const gainNode = audioContext.createGain();
        
        oscillator.connect(gainNode);
        gainNode.connect(audioContext.destination);
        
        // 根据类型设置不同频率
        const frequencies = {
            'success': [800, 1000, 1200],
            'info': [600, 800],
            'warning': [400, 400, 400]
        };
        
        const freq = frequencies[type] || frequencies.info;
        const duration = 0.1;
        
        let time = audioContext.currentTime;
        freq.forEach((f, index) => {
            oscillator.frequency.setValueAtTime(f, time + index * duration);
        });
        
        gainNode.gain.setValueAtTime(0.3, audioContext.currentTime);
        gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + duration * freq.length);
        
        oscillator.start(audioContext.currentTime);
        oscillator.stop(audioContext.currentTime + duration * freq.length);
    }
    
    /*
    显示浏览器通知
    @param title {String} 通知标题
    @param body {String} 通知内容
    @param icon {String} 通知图标
    */
    function showBrowserNotification(title, body, icon) {
        if (!('Notification' in window)) {
            return;
        }
        
        if (Notification.permission === 'granted') {
            new Notification(title, {
                body: body,
                icon: icon || '/static/img/logo.png',
                badge: '/static/img/badge.png',
                tag: 'alimpay-order',
                requireInteraction: false,
                silent: false
            });
        } else if (Notification.permission !== 'denied') {
            Notification.requestPermission().then(permission => {
                if (permission === 'granted') {
                    showBrowserNotification(title, body, icon);
                }
            });
        }
    }
    
    /*
    显示页面内通知
    @param message {String} 消息内容
    @param type {String} 消息类型 (success/info/warning/error)
    */
    function showToast(message, type = 'info') {
        const toast = document.createElement('div');
        toast.className = `toast toast-${type} slide-in-right`;
        
        const icons = {
            'success': '✅',
            'info': 'ℹ️',
            'warning': '⚠️',
            'error': '❌'
        };
        
        toast.innerHTML = `
            <span class="toast-icon">${icons[type]}</span>
            <span class="toast-message">${message}</span>
        `;
        
        document.body.appendChild(toast);
        
        setTimeout(() => {
            toast.classList.add('fade-out');
            setTimeout(() => toast.remove(), 300);
        }, 3000);
        
        // 添加音效
        playNotificationSound(type);
    }
    
    /*
    连接WebSocket
    */
    function connect() {
        if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) {
            return;
        }
        
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsURL = `${protocol}//${window.location.host}/ws/admin`;
        
        console.log('[Admin WS] Connecting to:', wsURL);
        ws = new WebSocket(wsURL);
        
        ws.onopen = () => {
            console.log('[Admin WS] Connected');
            reconnectAttempts = 0;
            reconnectInterval = 1000;
            showToast('实时连接已建立', 'success');
            triggerEvent('connected');
        };
        
        ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                console.log('[Admin WS] Message received:', data);
                handleMessage(data);
            } catch (e) {
                console.error('[Admin WS] Parse error:', e);
            }
        };
        
        ws.onclose = (event) => {
            console.warn('[Admin WS] Disconnected:', event.code, event.reason);
            triggerEvent('disconnected');
            
            if (reconnectAttempts < maxReconnectAttempts) {
                reconnectAttempts++;
                const delay = Math.min(reconnectInterval * Math.pow(2, reconnectAttempts - 1), maxReconnectInterval);
                console.log(`[Admin WS] Reconnecting in ${delay / 1000}s... (${reconnectAttempts}/${maxReconnectAttempts})`);
                showToast(`连接断开，${delay / 1000}秒后重连...`, 'warning');
                setTimeout(connect, delay);
            } else {
                console.error('[Admin WS] Max reconnect attempts reached');
                showToast('连接失败，请刷新页面', 'error');
            }
        };
        
        ws.onerror = (error) => {
            console.error('[Admin WS] Error:', error);
            ws.close();
        };
    }
    
    /*
    处理WebSocket消息
    @param data {Object} 消息数据
    */
    function handleMessage(data) {
        const { type, ...payload } = data;
        
        switch (type) {
            case 'order_created':
                handleOrderCreated(payload);
                break;
            case 'order_paid':
                handleOrderPaid(payload);
                break;
            case 'order_expired':
                handleOrderExpired(payload);
                break;
            case 'stats_update':
                handleStatsUpdate(payload);
                break;
            default:
                console.warn('[Admin WS] Unknown message type:', type);
        }
        
        // 触发自定义事件
        triggerEvent(type, payload);
    }
    
    /*
    处理订单创建
    @param order {Object} 订单信息
    */
    function handleOrderCreated(order) {
        console.log('[Admin WS] Order created:', order);
        showToast(`新订单: ${order.name} (¥${order.payment_amount})`, 'info');
        showBrowserNotification(
            '新订单通知',
            `${order.name} - ¥${order.payment_amount}`,
            null
        );
    }
    
    /*
    处理订单支付
    @param order {Object} 订单信息
    */
    function handleOrderPaid(order) {
        console.log('[Admin WS] Order paid:', order);
        showToast(`订单支付成功: ${order.name} (¥${order.payment_amount})`, 'success');
        showBrowserNotification(
            '💰 支付成功',
            `${order.name} - ¥${order.payment_amount}`,
            null
        );
    }
    
    /*
    处理订单过期
    @param order {Object} 订单信息
    */
    function handleOrderExpired(order) {
        console.log('[Admin WS] Order expired:', order);
        showToast(`订单已过期: ${order.name}`, 'warning');
    }
    
    /*
    处理统计更新
    @param stats {Object} 统计信息
    */
    function handleStatsUpdate(stats) {
        console.log('[Admin WS] Stats update:', stats);
        triggerEvent('stats', stats);
    }
    
    /*
    注册事件处理器
    @param event {String} 事件名称
    @param handler {Function} 处理函数
    */
    function on(event, handler) {
        if (!eventHandlers[event]) {
            eventHandlers[event] = [];
        }
        eventHandlers[event].push(handler);
    }
    
    /*
    移除事件处理器
    @param event {String} 事件名称
    @param handler {Function} 处理函数
    */
    function off(event, handler) {
        if (eventHandlers[event]) {
            eventHandlers[event] = eventHandlers[event].filter(h => h !== handler);
        }
    }
    
    /*
    触发事件
    @param event {String} 事件名称
    @param data {Object} 事件数据
    */
    function triggerEvent(event, data) {
        if (eventHandlers[event]) {
            eventHandlers[event].forEach(handler => {
                try {
                    handler(data);
                } catch (e) {
                    console.error('[Admin WS] Event handler error:', e);
                }
            });
        }
    }
    
    /*
    发送消息
    @param data {Object} 消息数据
    */
    function send(data) {
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify(data));
        } else {
            console.error('[Admin WS] Cannot send message, not connected');
        }
    }
    
    /*
    断开连接
    */
    function disconnect() {
        if (ws) {
            reconnectAttempts = maxReconnectAttempts; // 防止自动重连
            ws.close();
            ws = null;
        }
    }
    
    /*
    初始化
    */
    function init() {
        console.log('[Admin WS] Initializing...');
        
        // 请求通知权限
        if ('Notification' in window && Notification.permission === 'default') {
            Notification.requestPermission();
        }
        
        // 连接WebSocket
        connect();
        
        // 页面可见性变化时重连
        document.addEventListener('visibilitychange', () => {
            if (document.visibilityState === 'visible') {
                if (!ws || ws.readyState !== WebSocket.OPEN) {
                    console.log('[Admin WS] Page visible, reconnecting...');
                    connect();
                }
            }
        });
    }
    
    // 公开API
    return {
        init,
        connect,
        disconnect,
        send,
        on,
        off,
        showToast
    };
})();

// 自动初始化
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => AdminWebSocket.init());
} else {
    AdminWebSocket.init();
}

