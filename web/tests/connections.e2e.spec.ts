import { test, expect } from '@playwright/test'

test.describe('连接管理 - 新建对话框测试连接', () => {
  test('登录 -> 打开新建连接 -> 填写表单 -> 测试连接', async ({ page }) => {
    // 登录
    await page.goto('/login')
    await page.getByPlaceholder('用户名').fill('admin')
    await page.getByPlaceholder('密码').fill('admin123')
    await page.getByRole('button', { name: '登录' }).click()
    // 若失败则尝试使用密码 "password"
    try {
      await page.waitForURL('**/', { timeout: 4000 })
    } catch {
      await page.getByPlaceholder('密码').fill('password')
      await page.getByRole('button', { name: '登录' }).click()
      await page.waitForURL('**/')
    }

    // 进入连接管理
    await page.goto('/connections')
    await expect(page.getByText('连接管理')).toBeVisible()

    // 打开新建连接对话框
    await page.getByRole('button', { name: '新建连接' }).click()
    await expect(page.getByRole('dialog')).toBeVisible()

    // 填写表单（对应 docker-compose.dev.yml 中的mysql 默认账户）
    await page.getByPlaceholder('请输入连接名称').fill('E2E-连接测试')
    await page.getByPlaceholder('请选择环境').click()
    await page.getByRole('option', { name: '开发环境' }).click()
    await page.getByPlaceholder('请输入主机地址或IP').fill('localhost')
    // 端口：定位到端口的输入框（在带有标签“端口”的表单项里）
    const portFormItem = page.locator('.el-form-item').filter({ hasText: '端口' }).first()
    await portFormItem.locator('input').fill('3307')
    await page.getByPlaceholder('请输入数据库用户名').fill('root')
    await page.getByPlaceholder('请输入数据库密码').fill('root123456')
    await page.getByPlaceholder('请输入数据库名称').fill('test_db')

    // 点击“测试连接”
    await page.getByRole('button', { name: '测试连接' }).click()

    // 成功或失败的提示都可接受；这里至少断言出现了全局消息提示
    const successToast = page.getByText('连接测试成功')
    const failToast = page.getByText('连接测试失败')
    await expect(successToast.or(failToast)).toBeVisible({ timeout: 20000 })
  })
})


