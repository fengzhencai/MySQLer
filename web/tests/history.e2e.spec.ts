import { test, expect } from '@playwright/test'

test.describe('执行历史筛选/停止/重试 E2E', () => {
  test.beforeEach(async ({ page }) => {
    // 登录
    await page.goto('/login')
    await page.getByPlaceholder('用户名').fill('admin')
    await page.getByPlaceholder('密码').fill('admin123')
    await page.getByRole('button', { name: '登录' }).click()
    await page.waitForURL('**/')
  })

  test('历史筛选 -> 停止 -> 重试', async ({ page }) => {
    // 进入历史列表
    await page.goto('/history')
    await expect(page.getByText('执行历史')).toBeVisible()

    // 触发搜索
    await page.getByRole('button', { name: '搜索' }).click()
    await expect(page.getByRole('table')).toBeVisible()

    // 若存在运行中任务，尝试停止第一条
    const runningRow = page.locator('tr:has-text("执行中")').first()
    if (await runningRow.count()) {
      await runningRow.getByRole('button', { name: '停止' }).click()
      // 二次确认
      await page.getByRole('button', { name: '确定停止' }).click()
      // 成功提示
      await expect(page.getByText('已提交停止请求')).toBeVisible()
    }

    // 再次搜索，尝试查找失败任务
    await page.getByRole('button', { name: '搜索' }).click()

    // 若存在失败任务，点击重试
    const failedRow = page.locator('tr:has-text("失败")').first()
    if (await failedRow.count()) {
      await failedRow.getByRole('button', { name: '重试' }).click()
      await page.getByRole('button', { name: '确定重新执行' }).click()
      await expect(page.getByText('已提交重试请求')).toBeVisible()
    }

    // 关键词搜索（表名 / 执行ID片段）
    await page.getByPlaceholder('搜索表名、数据库名或执行ID').fill('test_')
    await page.getByRole('button', { name: '搜索' }).click()
    // 能正常返回空或若干条，断言表格可见不报错
    await expect(page.getByRole('table')).toBeVisible()
  })
})


