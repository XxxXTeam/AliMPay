/**
 * AliMPay 支付页面脚本
 * @version 1.0.0
 */

(function() {
    'use strict';

    // 配置
    const CONFIG = {
        CHECK_INTERVAL: 3000, // 3秒检查一次
        MAX_CHECK_TIME: 300000, // 5分钟超时
        COUNTDOWN_UPDATE_INTERVAL: 1000
    };

    // 状态管理
    const state = {
        checking: false,
        paid: false,
        checkCount: 0,
        startTime: Date.now(),
        countdownTimer: null,
        checkTimer: null
    };

    // DOM元素
    const elements = {
        statusIndicator: null,
        statusText: null,
        countdown: null,
        countdownTime: null
    };

    /**
     * 初始化应用
     */
    function init() {
        console.log('AliMPay Payment Page initialized');
        
        // 获取订单信息
        const orderInfo = getOrderInfo();
        if (!orderInfo) {
            console.error('Failed to get order info');
            return;
        }

        console.log('Order info:', orderInfo);

        // 初始化DOM元素
        initializeElements();

        // 启动支付检测
        startPaymentCheck(orderInfo);

        // 启动倒计时
        startCountdown(orderInfo);

        // 设置事件监听
        setupEventListeners();

        // 显示二维码提示
        showQRCodeTips();
    }

    /**
     * 初始化DOM元素
     */
    function initializeElements() {
        elements.statusIndicator = document.querySelector('.status-indicator');
        elements.statusText = document.querySelector('.status-text');
        elements.countdown = document.querySelector('.countdown');
        elements.countdownTime = document.querySelector('.countdown-time');
    }

    /**
     * 获取订单信息
     */
    function getOrderInfo() {
        // 从URL参数获取
        const params = new URLSearchParams(window.location.search);
        const tradeNo = params.get('trade_no');
        const amount = params.get('amount');

        if (!tradeNo || !amount) {
            return null;
        }

        return {
            tradeNo,
            amount: parseFloat(amount),
            pid: document.querySelector('[data-pid]')?.dataset.pid || ''
        };
    }

    /**
     * 启动支付检测
     */
    function startPaymentCheck(orderInfo) {
        state.checking = true;
        updateStatus('checking', '正在检测支付状态...');

        state.checkTimer = setInterval(() => {
            checkPaymentStatus(orderInfo);
        }, CONFIG.CHECK_INTERVAL);

        // 立即检查一次
        checkPaymentStatus(orderInfo);
    }

    /**
     * 检查支付状态
     */
    async function checkPaymentStatus(orderInfo) {
        if (state.paid) {
            stopPaymentCheck();
            return;
        }

        // 检查超时
        const elapsed = Date.now() - state.startTime;
        if (elapsed > CONFIG.MAX_CHECK_TIME) {
            stopPaymentCheck();
            updateStatus('timeout', '支付超时，请重新发起支付');
            return;
        }

        state.checkCount++;
        console.log(`Check payment status #${state.checkCount}`);

        try {
            const response = await fetch(
                `/api?action=order&pid=${orderInfo.pid}&out_trade_no=${orderInfo.tradeNo}`
            );
            const data = await response.json();

            if (data.code === 1 && data.status === 1) {
                // 支付成功
                handlePaymentSuccess(orderInfo);
            } else {
                console.log('Payment not completed yet, status:', data.status);
            }
        } catch (error) {
            console.error('Failed to check payment status:', error);
        }
    }

    /**
     * 处理支付成功
     */
    function handlePaymentSuccess(orderInfo) {
        state.paid = true;
        stopPaymentCheck();
        stopCountdown();

        updateStatus('success', '支付成功！');

        // 显示成功动画
        showSuccessAnimation();

        // 3秒后跳转
        setTimeout(() => {
            const returnUrl = new URLSearchParams(window.location.search).get('return_url');
            if (returnUrl) {
                window.location.href = returnUrl;
            } else {
                showSuccessPage();
            }
        }, 3000);
    }

    /**
     * 停止支付检测
     */
    function stopPaymentCheck() {
        state.checking = false;
        if (state.checkTimer) {
            clearInterval(state.checkTimer);
            state.checkTimer = null;
        }
    }

    /**
     * 启动倒计时
     */
    function startCountdown(orderInfo) {
        const maxTime = CONFIG.MAX_CHECK_TIME / 1000; // 转换为秒

        state.countdownTimer = setInterval(() => {
            const elapsed = Math.floor((Date.now() - state.startTime) / 1000);
            const remaining = maxTime - elapsed;

            if (remaining <= 0) {
                stopCountdown();
                updateCountdown('已超时');
                return;
            }

            const minutes = Math.floor(remaining / 60);
            const seconds = remaining % 60;
            updateCountdown(`${minutes}:${seconds.toString().padStart(2, '0')}`);
        }, CONFIG.COUNTDOWN_UPDATE_INTERVAL);
    }

    /**
     * 停止倒计时
     */
    function stopCountdown() {
        if (state.countdownTimer) {
            clearInterval(state.countdownTimer);
            state.countdownTimer = null;
        }
    }

    /**
     * 更新状态显示
     */
    function updateStatus(status, text) {
        if (!elements.statusIndicator) return;

        elements.statusIndicator.className = `status-indicator ${status}`;
        if (elements.statusText) {
            elements.statusText.textContent = text;
        }
    }

    /**
     * 更新倒计时显示
     */
    function updateCountdown(text) {
        if (elements.countdownTime) {
            elements.countdownTime.textContent = text;
        }
    }

    /**
     * 显示成功动画
     */
    function showSuccessAnimation() {
        // 创建成功动画元素
        const overlay = document.createElement('div');
        overlay.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.5);
            display: flex;
            align-items: center;
            justify-content: center;
            z-index: 9999;
            animation: fadeIn 0.3s ease-out;
        `;

        const successIcon = document.createElement('div');
        successIcon.style.cssText = `
            width: 120px;
            height: 120px;
            background: white;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 64px;
            animation: scaleIn 0.5s ease-out;
        `;
        successIcon.textContent = '✓';

        overlay.appendChild(successIcon);
        document.body.appendChild(overlay);

        // 3秒后移除
        setTimeout(() => {
            overlay.style.animation = 'fadeOut 0.3s ease-out';
            setTimeout(() => overlay.remove(), 300);
        }, 2700);
    }

    /**
     * 显示成功页面
     */
    function showSuccessPage() {
        document.body.innerHTML = `
            <div class="payment-container">
                <div class="payment-header">
                    <div class="logo">💰</div>
                    <h1>支付成功</h1>
                </div>
                <div class="payment-body">
                    <div class="result-container">
                        <div class="result-icon success">✓</div>
                        <div class="result-title">支付成功</div>
                        <div class="result-message">您的支付已完成，感谢您的使用！</div>
                        <div class="btn-group">
                            <button class="btn btn-primary" onclick="window.close()">关闭页面</button>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }

    /**
     * 显示二维码提示
     */
    function showQRCodeTips() {
        // 检查是否为移动设备
        const isMobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);
        
        if (isMobile) {
            // 移动设备显示打开支付宝提示
            const tips = document.querySelector('.qrcode-tips');
            if (tips) {
                tips.innerHTML = `
                    <div class="tip-icon">📱</div>
                    <p><strong>提示：</strong>请使用支付宝APP扫描二维码</p>
                    <p>如已安装支付宝，可点击二维码直接唤起</p>
                `;
            }
        }
    }

    /**
     * 设置事件监听
     */
    function setupEventListeners() {
        // 页面可见性变化
        document.addEventListener('visibilitychange', () => {
            if (document.hidden) {
                console.log('Page hidden, pausing checks');
                // 页面隐藏时暂停检测
            } else {
                console.log('Page visible, resuming checks');
                // 页面显示时恢复检测并立即检查一次
                if (state.checking && !state.paid) {
                    const orderInfo = getOrderInfo();
                    if (orderInfo) {
                        checkPaymentStatus(orderInfo);
                    }
                }
            }
        });

        // 页面卸载前清理
        window.addEventListener('beforeunload', () => {
            stopPaymentCheck();
            stopCountdown();
        });

        // 二维码点击放大（移动端）
        const qrcode = document.querySelector('.qrcode-wrapper img');
        if (qrcode) {
            qrcode.addEventListener('click', () => {
                const isMobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);
                if (isMobile) {
                    // 尝试唤起支付宝
                    attemptAlipayLaunch();
                }
            });
        }
    }

    /**
     * 尝试唤起支付宝
     */
    function attemptAlipayLaunch() {
        // 这里可以添加唤起支付宝的逻辑
        console.log('Attempting to launch Alipay...');
        // 实际实现需要支付宝的scheme URL
    }

    // 工具函数

    /**
     * 格式化金额
     */
    function formatAmount(amount) {
        return parseFloat(amount).toFixed(2);
    }

    /**
     * 格式化时间
     */
    function formatTime(seconds) {
        const minutes = Math.floor(seconds / 60);
        const secs = seconds % 60;
        return `${minutes}:${secs.toString().padStart(2, '0')}`;
    }

    // 添加必要的CSS动画
    const style = document.createElement('style');
    style.textContent = `
        @keyframes fadeIn {
            from { opacity: 0; }
            to { opacity: 1; }
        }
        @keyframes fadeOut {
            from { opacity: 1; }
            to { opacity: 0; }
        }
        @keyframes scaleIn {
            from { transform: scale(0); }
            to { transform: scale(1); }
        }
    `;
    document.head.appendChild(style);

    // 导出全局方法
    window.paymentActions = {
        checkPaymentStatus,
        stopPaymentCheck,
        getOrderInfo
    };

    // 页面加载完成后初始化
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

})();

