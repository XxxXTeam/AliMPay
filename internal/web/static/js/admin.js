/*
 * AliMPay 管理后台前端脚本
 * @version 2.0.0
 * @description 订单管理、实时更新、操作功能
 */

(function() {
    'use strict';

    // 全局状态
    const state = {
        orders: [],
        ws: null,
        stats: {
            pending: 0,
            paid: 0,
            total: 0
        }
    };

    // API配置
    const API = {
        orders: '/admin/orders',
        action: '/admin/action',
        wsAdmin: '/admin/ws', // 管理后台WebSocket（需要认证）
        logout: '/admin/logout'
    };

    // 工具函数
    const utils = {
        // 格式化时间
        formatTime(timestamp) {
            if (!timestamp) return '-';
            const date = new Date(timestamp);
            return date.toLocaleString('zh-CN', {
                year: 'numeric',
                month: '2-digit',
                day: '2-digit',
                hour: '2-digit',
                minute: '2-digit',
                second: '2-digit'
            });
        },

        // 格式化金额
        formatAmount(amount) {
            return `¥${parseFloat(amount).toFixed(2)}`;
        },

        // 获取状态文本和类
        getStatusInfo(status) {
            const statusMap = {
                0: { text: '待支付', class: 'status-pending' },
                1: { text: '已支付', class: 'status-paid' },
                2: { text: '已关闭', class: 'status-closed' },
                3: { text: '已过期', class: 'status-expired' }
            };
            return statusMap[status] || { text: '未知', class: '' };
        },

        // 显示消息
        showAlert(message, type = 'success') {
            const alert = document.getElementById('alert');
            alert.className = `alert alert-${type}`;
            alert.textContent = message;
            alert.style.display = 'block';

            setTimeout(() => {
                alert.style.display = 'none';
            }, 3000);
        },

        // 确认对话框
        confirm(message) {
            return window.confirm(message);
        }
    };

    // 订单管理
    const orderManager = {
        // 加载订单列表
        async loadOrders() {
            try {
                const response = await fetch(API.orders, {
                    credentials: 'include'
                });

                if (!response.ok) {
                    if (response.status === 401) {
                        window.location.href = '/admin/login';
                        return;
                    }
                    throw new Error('Failed to load orders');
                }

                const data = await response.json();

                if (data.code === 1) {
                    state.orders = data.orders || [];
                    this.renderOrders(state.orders);
                } else {
                    utils.showAlert(data.msg || '加载订单失败', 'error');
                }
            } catch (error) {
                console.error('Load orders error:', error);
                utils.showAlert('加载订单失败: ' + error.message, 'error');
            }
        },

        // 渲染订单列表
        renderOrders(orders) {
            const tbody = document.getElementById('ordersBody');

            if (!orders || orders.length === 0) {
                tbody.innerHTML = `
                    <tr>
                        <td colspan="8" class="empty-state">
                            <p>📭 暂无订单</p>
                        </td>
                    </tr>
                `;
                return;
            }

            tbody.innerHTML = orders.map(order => {
                const statusInfo = utils.getStatusInfo(order.status);
                return `
                    <tr data-order-id="${order.trade_no}">
                        <td><code>${order.trade_no}</code></td>
                        <td>${order.out_trade_no || '-'}</td>
                        <td>${order.name || '-'}</td>
                        <td>${utils.formatAmount(order.price)}</td>
                        <td class="amount">${utils.formatAmount(order.payment_amount || order.price)}</td>
                        <td><span class="status ${statusInfo.class}">${statusInfo.text}</span></td>
                        <td>${utils.formatTime(order.add_time)}</td>
                        <td>${this.renderActions(order)}</td>
                    </tr>
                `;
            }).join('');
        },

        // 渲染操作按钮
        renderActions(order) {
            const actions = [];

            if (order.status === 0) {
                actions.push(`
                    <button class="btn btn-sm btn-success" onclick="window.adminActions.markPaid('${order.trade_no}')">
                        ✅ 标记已支付
                    </button>
                `);
                actions.push(`
                    <button class="btn btn-sm btn-danger" onclick="window.adminActions.cancelOrder('${order.trade_no}')">
                        ❌ 取消
                    </button>
                `);
            }

            return actions.length > 0 ? actions.join('') : '<span style="color: #999;">-</span>';
        },

        // 搜索订单
        searchOrder() {
            const input = document.getElementById('searchInput');
            const keyword = input.value.trim().toLowerCase();

            if (!keyword) {
                this.renderOrders(state.orders);
                return;
            }

            const filtered = state.orders.filter(order => 
                (order.trade_no && order.trade_no.toLowerCase().includes(keyword)) ||
                (order.out_trade_no && order.out_trade_no.toLowerCase().includes(keyword)) ||
                (order.name && order.name.toLowerCase().includes(keyword))
            );

            this.renderOrders(filtered);
        }
    };

    // 订单操作
    const orderActions = {
        // 标记订单为已支付
        async markPaid(tradeNo) {
            if (!utils.confirm(`确定要标记订单 ${tradeNo} 为已支付吗？\n\n此操作将：\n1. 更新订单状态为已支付\n2. 发送通知给商户`)) {
                return;
            }

            try {
                const response = await fetch(API.action, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    credentials: 'include',
                    body: JSON.stringify({
                        action: 'pay',
                        trade_no: tradeNo
                    })
                });

                if (!response.ok) {
                    throw new Error('操作失败');
                }

                const data = await response.json();

                if (data.success) {
                    utils.showAlert('订单已标记为已支付', 'success');
                    // 重新加载订单列表
                    orderManager.loadOrders();
                } else {
                    utils.showAlert(data.error || '操作失败', 'error');
                }
            } catch (error) {
                console.error('Mark paid error:', error);
                utils.showAlert('操作失败: ' + error.message, 'error');
            }
        },

        // 取消订单
        async cancelOrder(tradeNo) {
            if (!utils.confirm(`确定要取消订单 ${tradeNo} 吗？\n\n此操作不可撤销！`)) {
                return;
            }

            try {
                const response = await fetch(API.action, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    credentials: 'include',
                    body: JSON.stringify({
                        action: 'cancel',
                        trade_no: tradeNo
                    })
                });

                if (!response.ok) {
                    throw new Error('操作失败');
                }

                const data = await response.json();

                if (data.success) {
                    utils.showAlert('订单已取消', 'success');
                    // 重新加载订单列表
                    orderManager.loadOrders();
                } else {
                    utils.showAlert(data.error || '操作失败', 'error');
                }
            } catch (error) {
                console.error('Cancel order error:', error);
                utils.showAlert('操作失败: ' + error.message, 'error');
            }
        },

        // 刷新订单列表
        loadOrders() {
            orderManager.loadOrders();
        },

        // 搜索订单
        searchOrder() {
            orderManager.searchOrder();
        }
    };

    // WebSocket管理
    const wsManager = {
        connect() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsURL = `${protocol}//${window.location.host}${API.wsAdmin}`;

            console.log('[Admin WS] Connecting to:', wsURL);
            state.ws = new WebSocket(wsURL);

            state.ws.onopen = () => {
                console.log('[Admin WS] Connected');
            };

            state.ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    this.handleMessage(data);
                } catch (error) {
                    console.error('[Admin WS] Message parse error:', error);
                }
            };

            state.ws.onclose = () => {
                console.log('[Admin WS] Disconnected, reconnecting in 3s...');
                setTimeout(() => this.connect(), 3000);
            };

            state.ws.onerror = (error) => {
                console.error('[Admin WS] Error:', error);
            };
        },

        handleMessage(data) {
            console.log('[Admin WS] Message:', data);

            switch (data.type) {
                case 'stats_update':
                    this.updateStats(data);
                    break;
                case 'order_created':
                    this.handleOrderCreated(data);
                    break;
                case 'order_paid':
                    this.handleOrderPaid(data);
                    break;
                case 'order_expired':
                    this.handleOrderExpired(data);
                    break;
            }
        },

        updateStats(data) {
            state.stats.pending = data.pending_count || 0;
            state.stats.paid = data.paid_count || 0;
            state.stats.total = data.total_count || 0;

            document.getElementById('pendingCount').textContent = state.stats.pending;
            document.getElementById('paidCount').textContent = state.stats.paid;
            document.getElementById('totalCount').textContent = state.stats.total;
        },

        handleOrderCreated(data) {
            utils.showAlert(`新订单：${data.name} (${utils.formatAmount(data.payment_amount)})`, 'info');
            // 重新加载订单列表
            orderManager.loadOrders();
        },

        handleOrderPaid(data) {
            utils.showAlert(`订单已支付：${data.name} (${utils.formatAmount(data.payment_amount)})`, 'success');
            // 重新加载订单列表
            orderManager.loadOrders();
        },

        handleOrderExpired(data) {
            utils.showAlert(`订单已过期：${data.order_id}`, 'warning');
            // 重新加载订单列表
            orderManager.loadOrders();
        }
    };

    // 初始化
    function init() {
        console.log('[Admin] Initializing...');

        // 加载订单
        orderManager.loadOrders();

        // 连接WebSocket
        wsManager.connect();

        // 绑定搜索框回车事件
        const searchInput = document.getElementById('searchInput');
        if (searchInput) {
            searchInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    orderManager.searchOrder();
                }
            });
        }

        console.log('[Admin] Initialized successfully');
    }

    // 暴露全局API
    window.adminActions = orderActions;

    // 页面加载完成后初始化
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
