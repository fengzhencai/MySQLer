#!/usr/bin/env node

/**
 * Playwright MCP Server for MySQLer Project
 * 提供浏览器自动化功能的MCP服务器
 */

const { chromium, firefox, webkit } = require('playwright');

class PlaywrightMCPServer {
    constructor() {
        this.browser = null;
        this.context = null;
        this.page = null;
        this.tools = {
            'playwright_navigate': this.navigate.bind(this),
            'playwright_click': this.click.bind(this),
            'playwright_type': this.type.bind(this),
            'playwright_screenshot': this.screenshot.bind(this),
            'playwright_get_text': this.getText.bind(this),
            'playwright_wait_for_selector': this.waitForSelector.bind(this),
            'playwright_close': this.closeBrowser.bind(this)
        };
    }

    async initialize() {
        console.log('初始化 Playwright MCP 服务器...');
        // 启动浏览器
        this.browser = await chromium.launch({ 
            headless: false,  // 设置为 true 可以无头模式运行
            args: ['--disable-web-security', '--disable-features=VizDisplayCompositor']
        });
        this.context = await this.browser.newContext();
        this.page = await this.context.newPage();
        console.log('Playwright 浏览器已启动');
    }

    async navigate(url) {
        if (!this.page) await this.initialize();
        console.log(`导航到: ${url}`);
        await this.page.goto(url);
        return { success: true, url: this.page.url() };
    }

    async click(selector) {
        if (!this.page) throw new Error('浏览器未初始化');
        console.log(`点击元素: ${selector}`);
        await this.page.click(selector);
        return { success: true, selector };
    }

    async type(selector, text) {
        if (!this.page) throw new Error('浏览器未初始化');
        console.log(`在 ${selector} 中输入: ${text}`);
        await this.page.fill(selector, text);
        return { success: true, selector, text };
    }

    async screenshot(path = 'screenshot.png') {
        if (!this.page) throw new Error('浏览器未初始化');
        console.log(`截图保存到: ${path}`);
        await this.page.screenshot({ path, fullPage: true });
        return { success: true, path };
    }

    async getText(selector) {
        if (!this.page) throw new Error('浏览器未初始化');
        const text = await this.page.textContent(selector);
        return { success: true, selector, text };
    }

    async waitForSelector(selector, timeout = 30000) {
        if (!this.page) throw new Error('浏览器未初始化');
        console.log(`等待元素: ${selector}`);
        await this.page.waitForSelector(selector, { timeout });
        return { success: true, selector };
    }

    async closeBrowser() {
        if (this.browser) {
            console.log('关闭浏览器...');
            await this.browser.close();
            this.browser = null;
            this.context = null;
            this.page = null;
        }
        return { success: true };
    }

    // MCP协议处理
    handleMCPRequest(request) {
        console.log('收到MCP请求:', JSON.stringify(request, null, 2));
        
        if (request.method === 'tools/list') {
            return {
                tools: Object.keys(this.tools).map(name => ({
                    name,
                    description: `Playwright ${name} tool`,
                    inputSchema: {
                        type: 'object',
                        properties: {},
                        required: []
                    }
                }))
            };
        }

        if (request.method === 'tools/call') {
            const toolName = request.params.name;
            const args = request.params.arguments || {};
            
            if (this.tools[toolName]) {
                return this.tools[toolName](args);
            } else {
                throw new Error(`Unknown tool: ${toolName}`);
            }
        }

        throw new Error(`Unknown method: ${request.method}`);
    }
}

// 启动MCP服务器
const server = new PlaywrightMCPServer();

// 处理标准输入/输出的MCP通信
process.stdin.on('data', async (data) => {
    try {
        const request = JSON.parse(data.toString());
        const response = await server.handleMCPRequest(request);
        console.log(JSON.stringify(response));
    } catch (error) {
        console.error('错误:', error.message);
        console.log(JSON.stringify({ error: error.message }));
    }
});

// 优雅关闭
process.on('SIGINT', async () => {
    console.log('收到关闭信号，正在清理...');
    await server.closeBrowser();
    process.exit(0);
});

process.on('SIGTERM', async () => {
    console.log('收到终止信号，正在清理...');
    await server.closeBrowser();
    process.exit(0);
});

console.log('Playwright MCP 服务器已启动，等待请求...');
console.log('可用工具:', Object.keys(server.tools).join(', '));
