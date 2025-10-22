/**
 * AliMPay 管理后台脚本
 * @version 1.0.0
 */

(function() {
    'use strict';

    // 配置
    const CONFIG = {
        PID: '1001003549245339',
        AUTO_REFRESH_INTERVAL: 30000, // 30秒自动刷新
        ALERT_DURATION: 3000
    };

    // 状态管理
    const state = {
        orders: [],
        loading: false,
        autoRefresh: true
    };

    // DOM元素
    const elements = {
        ordersBody: document.getElementById('ordersBody'),
        searchInput: document.getElementById('searchInput'),
        alert: document.getElementById('alert'),
        pendingCount: document.getElementById('pendingCount'),
        paidCount: document.getElementById('paidCount'),
        totalCount: document.getElementById('totalCount')
    };

    /**
     * 初始化应用
     */
    function init() {
        loadOrders();
        setupEventListeners();
        startAutoRefresh();
        console.log('AliMPay Admin Dashboard initialized');
    }

    /**
     * 设置事件监听
     */
    function setupEventListeners() {
        // 搜索功能
        if (elements.searchInput) {
            elements.searchInput.addEventListener('input', debounce(handleSearch, 300));
            elements.searchInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    handleSearch();
                }
            });
        }

        // 键盘快捷键
        document.addEventListener('keydown', (e) => {
            // Ctrl/Cmd + R: 刷新
            if ((e.ctrlKey || e.metaKey) && e.key === 'r') {
                e.preventDefault();
                loadOrders();
            }
        });
    }

    /**
     * 加载订单列表
     */
    async function loadOrders() {
        if (state.loading) return;

        state.loading = true;
        showLoading();

        try {
            const response = await fetch('/admin/orders');
            const data = await response.json();

            if (data.code === 1) {
                state.orders = data.orders || [];
                renderOrders(state.orders);
                updateStats(state.orders);
            } else {
                showAlert('加载失败：' + data.msg, 'error');
            }
        } catch (error) {
            console.error('Failed to load orders:', error);
            showAlert('网络错误：' + error.message, 'error');
        } finally {
            state.loading = false;
        }
    }

    /**
     * 渲染订单列表
     */
    function renderOrders(orders) {
        if (!elements.ordersBody) return;

        if (orders.length === 0) {
            elements.ordersBody.innerHTML = `
                <tr>
                    <td colspan="8" class="empty-state">
                        <svg viewBox="0 0 64 64" fill="currentColor">
                            <path d="M32 2C15.432 2 2 15.432 2 32s13.432 30 30 30 30-13.432 30-30S48.568 2 32 2zm0 54C17.641 56 6 44.359 6 32S17.641 8 32 8s26 11.641 26 24-11.641 24-26 24zm-4-34c0-2.209 1.791-4 4-4s4 1.791 4 4-1.791 4-4 4-4-1.791-4-4zm-2 30h12v-2H26v2zm0-6h12v-2H26v2zm0-6h12v-2H26v2z"/>
                        </svg>
                        <h3>暂无订单数据</h3>
                        <p>系统中还没有订单记录</p>
                    </td>
                </tr>
            `;
            return;
        }

        const html = orders.map(order => `
            <tr data-order-id="${order.trade_no}" data-status="${order.status}">
                <td><code>${escapeHtml(order.trade_no)}</code></td>
                <td><code>${escapeHtml(order.out_trade_no)}</code></td>
                <td title="${escapeHtml(order.name)}">${truncate(escapeHtml(order.name), 20)}</td>
                <td>¥${formatAmount(order.price)}</td>
                <td>¥${formatAmount(order.payment_amount || order.price)}</td>
                <td>${renderStatus(order.status)}</td>
                <td>${formatTime(order.add_time)}</td>
                <td>
                    <div class="actions">
                        ${renderActions(order)}
                    </div>
                </td>
            </tr>
        `).join('');

        elements.ordersBody.innerHTML = html;
    }

    /**
     * 渲染订单状态
     */
    function renderStatus(status) {
        const statusMap = {
            0: { class: 'pending', text: '待支付' },
            1: { class: 'paid', text: '已支付' },
            2: { class: 'closed', text: '已关闭' }
        };

        const statusInfo = statusMap[status] || { class: '', text: '未知' };
        return `<span class="status ${statusInfo.class}">${statusInfo.text}</span>`;
    }

    /**
     * 渲染操作按钮
     */
    function renderActions(order) {
        if (order.status === 0) {
            return `
                <button class="btn btn-success" onclick="window.adminActions.markPaid('${order.out_trade_no}')" title="标记为已支付">
                    ✓ 已支付
                </button>
                <button class="btn btn-danger" onclick="window.adminActions.cancelOrder('${order.out_trade_no}')" title="取消订单">
                    ✗ 取消
                </button>
            `;
        } else {
            return `
                <button class="btn btn-info" onclick="window.adminActions.viewDetails('${order.trade_no}')" title="查看详情">
                    📄 详情
                </button>
            `;
        }
    }

    /**
     * 更新统计数据
     */
    function updateStats(orders) {
        const pending = orders.filter(o => o.status === 0).length;
        const paid = orders.filter(o => o.status === 1).length;
        
        // 今日订单
        const today = new Date().toDateString();
        const todayOrders = orders.filter(o => {
            const orderDate = new Date(o.add_time).toDateString();
            return orderDate === today;
        }).length;

        // 动画更新数字
        animateNumber(elements.pendingCount, pending);
        animateNumber(elements.paidCount, paid);
        animateNumber(elements.totalCount, todayOrders);
    }

    /**
     * 标记为已支付
     */
    async function markPaid(outTradeNo) {
        if (!confirm('确认标记此订单为已支付？\n\n这将触发商户回调通知。')) {
            return;
        }

        try {
            const response = await fetch(
                `/admin?action=mark_paid&pid=${CONFIG.PID}&out_trade_no=${encodeURIComponent(outTradeNo)}`,
                { method: 'POST' }
            );
            const data = await response.json();

            if (data.success) {
                showAlert('订单已标记为已支付', 'success');
                setTimeout(() => loadOrders(), 1000);
            } else {
                showAlert('操作失败：' + data.error, 'error');
            }
        } catch (error) {
            console.error('Failed to mark paid:', error);
            showAlert('网络错误：' + error.message, 'error');
        }
    }

    /**
     * 取消订单
     */
    async function cancelOrder(outTradeNo) {
        if (!confirm('确认取消此订单？\n\n取消后订单将无法恢复。')) {
            return;
        }

        try {
            const response = await fetch(
                `/admin?action=cancel&pid=${CONFIG.PID}&out_trade_no=${encodeURIComponent(outTradeNo)}`,
                { method: 'POST' }
            );
            const data = await response.json();

            if (data.success) {
                showAlert('订单已取消', 'success');
                setTimeout(() => loadOrders(), 1000);
            } else {
                showAlert('操作失败：' + data.error, 'error');
            }
        } catch (error) {
            console.error('Failed to cancel order:', error);
            showAlert('网络错误：' + error.message, 'error');
        }
    }

    /**
     * 查看订单详情
     */
    function viewDetails(tradeNo) {
        // 查找订单
        const order = state.orders.find(o => o.trade_no === tradeNo);
        if (!order) {
            showAlert('订单不存在', 'error');
            return;
        }

        // 构建详情信息
        const details = `
订单详情
━━━━━━━━━━━━━━━━
订单号: ${order.trade_no}
商户订单号: ${order.out_trade_no}
商品名称: ${order.name}
订单金额: ¥${formatAmount(order.price)}
实付金额: ¥${formatAmount(order.payment_amount)}
订单状态: ${getStatusText(order.status)}
创建时间: ${formatTime(order.add_time)}
支付时间: ${order.pay_time ? formatTime(order.pay_time) : '未支付'}
━━━━━━━━━━━━━━━━
        `.trim();

        alert(details);
    }

    /**
     * 搜索订单
     */
    function handleSearch() {
        const keyword = elements.searchInput.value.trim().toLowerCase();
        
        if (!keyword) {
            renderOrders(state.orders);
            return;
        }

        const filtered = state.orders.filter(order => 
            order.trade_no.toLowerCase().includes(keyword) ||
            order.out_trade_no.toLowerCase().includes(keyword) ||
            order.name.toLowerCase().includes(keyword)
        );

        renderOrders(filtered);
        
        if (filtered.length === 0) {
            showAlert(`未找到包含 "${keyword}" 的订单`, 'info');
        }
    }

    /**
     * 显示提示消息
     */
    function showAlert(message, type = 'info') {
        if (!elements.alert) return;

        elements.alert.className = `alert ${type}`;
        elements.alert.textContent = message;
        elements.alert.style.display = 'block';

        setTimeout(() => {
            elements.alert.style.display = 'none';
        }, CONFIG.ALERT_DURATION);
    }

    /**
     * 显示加载状态
     */
    function showLoading() {
        if (!elements.ordersBody) return;

        elements.ordersBody.innerHTML = `
            <tr>
                <td colspan="8" class="empty-state">
                    <div class="loading"></div>
                    <p style="margin-top: 16px;">加载中...</p>
                </td>
            </tr>
        `;
    }

    /**
     * 启动自动刷新
     */
    function startAutoRefresh() {
        if (!state.autoRefresh) return;

        setInterval(() => {
            if (state.autoRefresh && !state.loading) {
                loadOrders();
                console.log('Auto refresh triggered');
            }
        }, CONFIG.AUTO_REFRESH_INTERVAL);
    }

    // 工具函数

    /**
     * 防抖函数
     */
    function debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    /**
     * 格式化金额
     */
    function formatAmount(amount) {
        return parseFloat(amount).toFixed(2);
    }

    /**
     * 格式化时间
     */
    function formatTime(timeStr) {
        const date = new Date(timeStr);
        return date.toLocaleString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit'
        });
    }

    /**
     * 获取状态文本
     */
    function getStatusText(status) {
        const statusMap = {
            0: '待支付',
            1: '已支付',
            2: '已关闭'
        };
        return statusMap[status] || '未知';
    }

    /**
     * 截断文本
     */
    function truncate(text, length) {
        return text.length > length ? text.substring(0, length) + '...' : text;
    }

    /**
     * 转义HTML
     */
    function escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    /**
     * 数字动画
     */
    function animateNumber(element, targetValue) {
        if (!element) return;

        const currentValue = parseInt(element.textContent) || 0;
        const duration = 500;
        const steps = 20;
        const stepValue = (targetValue - currentValue) / steps;
        const stepDuration = duration / steps;

        let currentStep = 0;

        const timer = setInterval(() => {
            currentStep++;
            const newValue = Math.round(currentValue + stepValue * currentStep);
            element.textContent = newValue;

            if (currentStep >= steps) {
                clearInterval(timer);
                element.textContent = targetValue;
            }
        }, stepDuration);
    }

    // 导出全局方法
    window.adminActions = {
        loadOrders,
        markPaid,
        cancelOrder,
        viewDetails,
        searchOrder: handleSearch
    };

    // 页面加载完成后初始化
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

})();

