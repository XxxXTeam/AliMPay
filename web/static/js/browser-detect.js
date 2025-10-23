/*
浏览器环境检测模块
功能：
  - 检测浏览器类型（Chrome/Safari/Firefox/Edge/QQ浏览器/UC浏览器等）
  - 检测设备类型（Desktop/Mobile/Tablet）
  - 检测操作系统（iOS/Android/Windows/Mac/Linux）
  - 检测特殊环境（微信/支付宝/小程序内）
  - 提供浏览器能力检测（WebSocket/HTTP2/ServiceWorker等）

使用示例：
  const browser = BrowserDetector.detect();
  console.log('浏览器:', browser.name, browser.version);
  console.log('设备:', browser.device.type);
  console.log('操作系统:', browser.os.name);
*/

const BrowserDetector = (function() {
    const ua = navigator.userAgent.toLowerCase();
    const standalone = navigator.standalone || window.matchMedia('(display-mode: standalone)').matches;
    
    /*
    检测浏览器类型
    @return {Object} 浏览器信息
    */
    function detectBrowser() {
        let name = 'unknown';
        let version = 'unknown';
        let engine = 'unknown';
        
        // 特殊应用内浏览器
        if (ua.includes('micromessenger')) {
            name = 'WeChat';
            const match = ua.match(/micromessenger\/([\d.]+)/);
            version = match ? match[1] : 'unknown';
        } else if (ua.includes('alipayclient')) {
            name = 'Alipay';
            const match = ua.match(/alipayclient\/([\d.]+)/);
            version = match ? match[1] : 'unknown';
        } else if (ua.includes('qq/') || ua.includes('qqbrowser')) {
            name = 'QQBrowser';
            const match = ua.match(/(?:qq|qqbrowser)\/([\d.]+)/);
            version = match ? match[1] : 'unknown';
        } else if (ua.includes('ucbrowser')) {
            name = 'UCBrowser';
            const match = ua.match(/ucbrowser\/([\d.]+)/);
            version = match ? match[1] : 'unknown';
        } else if (ua.includes('mqqbrowser')) {
            name = 'MQQBrowser';
            const match = ua.match(/mqqbrowser\/([\d.]+)/);
            version = match ? match[1] : 'unknown';
        }
        // 主流浏览器
        else if (ua.includes('edg/')) {
            name = 'Edge';
            engine = 'Chromium';
            const match = ua.match(/edg\/([\d.]+)/);
            version = match ? match[1] : 'unknown';
        } else if (ua.includes('chrome') && !ua.includes('edg')) {
            name = 'Chrome';
            engine = 'Chromium';
            const match = ua.match(/chrome\/([\d.]+)/);
            version = match ? match[1] : 'unknown';
        } else if (ua.includes('safari') && !ua.includes('chrome')) {
            name = 'Safari';
            engine = 'WebKit';
            const match = ua.match(/version\/([\d.]+)/);
            version = match ? match[1] : 'unknown';
        } else if (ua.includes('firefox')) {
            name = 'Firefox';
            engine = 'Gecko';
            const match = ua.match(/firefox\/([\d.]+)/);
            version = match ? match[1] : 'unknown';
        } else if (ua.includes('opera') || ua.includes('opr/')) {
            name = 'Opera';
            engine = 'Chromium';
            const match = ua.match(/(?:opera|opr)\/([\d.]+)/);
            version = match ? match[1] : 'unknown';
        }
        
        return { name, version, engine };
    }
    
    /*
    检测设备类型
    @return {Object} 设备信息
    */
    function detectDevice() {
        let type = 'desktop';
        let vendor = 'unknown';
        
        if (/(tablet|ipad|playbook|silk)|(android(?!.*mobi))/i.test(ua)) {
            type = 'tablet';
        } else if (/mobile|android|iphone|ipod|blackberry|iemobile|opera mini/i.test(ua)) {
            type = 'mobile';
        }
        
        // 设备厂商
        if (ua.includes('iphone') || ua.includes('ipad')) {
            vendor = 'Apple';
        } else if (ua.includes('huawei')) {
            vendor = 'Huawei';
        } else if (ua.includes('xiaomi') || ua.includes('mi ')) {
            vendor = 'Xiaomi';
        } else if (ua.includes('oppo')) {
            vendor = 'Oppo';
        } else if (ua.includes('vivo')) {
            vendor = 'Vivo';
        } else if (ua.includes('samsung')) {
            vendor = 'Samsung';
        }
        
        return { type, vendor };
    }
    
    /*
    检测操作系统
    @return {Object} 操作系统信息
    */
    function detectOS() {
        let name = 'unknown';
        let version = 'unknown';
        
        if (ua.includes('windows')) {
            name = 'Windows';
            if (ua.includes('windows nt 10.0')) version = '10/11';
            else if (ua.includes('windows nt 6.3')) version = '8.1';
            else if (ua.includes('windows nt 6.2')) version = '8';
            else if (ua.includes('windows nt 6.1')) version = '7';
        } else if (ua.includes('mac os')) {
            name = 'macOS';
            const match = ua.match(/mac os x ([\d_]+)/);
            if (match) version = match[1].replace(/_/g, '.');
        } else if (ua.includes('iphone') || ua.includes('ipad')) {
            name = 'iOS';
            const match = ua.match(/os ([\d_]+)/);
            if (match) version = match[1].replace(/_/g, '.');
        } else if (ua.includes('android')) {
            name = 'Android';
            const match = ua.match(/android ([\d.]+)/);
            version = match ? match[1] : 'unknown';
        } else if (ua.includes('linux')) {
            name = 'Linux';
        }
        
        return { name, version };
    }
    
    /*
    检测特殊环境
    @return {Object} 特殊环境信息
    */
    function detectEnvironment() {
        return {
            isWeChat: ua.includes('micromessenger'),
            isAlipay: ua.includes('alipayclient'),
            isQQ: ua.includes('qq/'),
            isMiniProgram: ua.includes('miniprogram') || window.__wxjs_environment === 'miniprogram',
            isPWA: standalone,
            isWebView: !!(window.webkit?.messageHandlers || window.AndroidBridge),
        };
    }
    
    /*
    检测浏览器能力
    @return {Object} 浏览器能力信息
    */
    function detectCapabilities() {
        return {
            webSocket: 'WebSocket' in window,
            webRTC: 'RTCPeerConnection' in window,
            serviceWorker: 'serviceWorker' in navigator,
            pushNotification: 'PushManager' in window,
            geolocation: 'geolocation' in navigator,
            vibrate: 'vibrate' in navigator,
            bluetooth: 'bluetooth' in navigator,
            nfc: 'nfc' in navigator,
            webGL: !!document.createElement('canvas').getContext('webgl'),
            localStorage: (() => {
                try {
                    localStorage.setItem('test', 'test');
                    localStorage.removeItem('test');
                    return true;
                } catch(e) {
                    return false;
                }
            })(),
            cookies: navigator.cookieEnabled,
            online: navigator.onLine,
        };
    }
    
    /*
    检测屏幕信息
    @return {Object} 屏幕信息
    */
    function detectScreen() {
        return {
            width: window.screen.width,
            height: window.screen.height,
            availWidth: window.screen.availWidth,
            availHeight: window.screen.availHeight,
            colorDepth: window.screen.colorDepth,
            pixelDepth: window.screen.pixelDepth,
            orientation: window.screen.orientation?.type || 'unknown',
            devicePixelRatio: window.devicePixelRatio || 1,
            touchPoints: navigator.maxTouchPoints || 0,
        };
    }
    
    /*
    获取完整的浏览器环境信息
    @return {Object} 完整环境信息
    */
    function detect() {
        const browser = detectBrowser();
        const device = detectDevice();
        const os = detectOS();
        const env = detectEnvironment();
        const capabilities = detectCapabilities();
        const screen = detectScreen();
        
        return {
            browser,
            device,
            os,
            env,
            capabilities,
            screen,
            userAgent: ua,
            language: navigator.language || navigator.userLanguage,
            languages: navigator.languages || [navigator.language],
            platform: navigator.platform,
            vendor: navigator.vendor,
            timestamp: Date.now(),
        };
    }
    
    /*
    获取友好的浏览器描述
    @return {String} 浏览器描述
    */
    function getDescription() {
        const info = detect();
        return `${info.browser.name} ${info.browser.version} on ${info.os.name} ${info.os.version} (${info.device.type})`;
    }
    
    /*
    判断是否为移动设备
    @return {Boolean}
    */
    function isMobile() {
        return detectDevice().type === 'mobile';
    }
    
    /*
    判断是否为平板设备
    @return {Boolean}
    */
    function isTablet() {
        return detectDevice().type === 'tablet';
    }
    
    /*
    判断是否为桌面设备
    @return {Boolean}
    */
    function isDesktop() {
        return detectDevice().type === 'desktop';
    }
    
    /*
    判断是否支持触摸
    @return {Boolean}
    */
    function isTouchDevice() {
        return detectScreen().touchPoints > 0 || 'ontouchstart' in window;
    }
    
    // 公开API
    return {
        detect,
        getDescription,
        isMobile,
        isTablet,
        isDesktop,
        isTouchDevice,
        getBrowser: detectBrowser,
        getDevice: detectDevice,
        getOS: detectOS,
        getEnvironment: detectEnvironment,
        getCapabilities: detectCapabilities,
        getScreen: detectScreen,
    };
})();

// 自动打印检测结果（仅开发环境）
if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    console.group('🔍 浏览器环境检测');
    console.log('描述:', BrowserDetector.getDescription());
    console.log('完整信息:', BrowserDetector.detect());
    console.groupEnd();
}

