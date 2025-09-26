import { test, expect } from '@playwright/test'

const BASE_URL = 'http://localhost:8090'

function buildPreviewPayload(noCheckAlter: boolean) {
  return {
    connection_id: '00000000-0000-0000-0000-000000000000',
    table_name: 't_demo',
    database_name: 'test_db',
    ddl_type: 'custom',
    original_ddl: "ADD COLUMN c1 INT NULL, DROP PRIMARY KEY, ADD PRIMARY KEY (`id`,`c1`)",
    execution_params: {
      chunk_size: 1000,
      max_load: 'Threads_running=25',
      critical_load: 'Threads_running=50',
      charset: 'utf8mb4',
      lock_wait_timeout: 60,
      no_check_alter: noCheckAlter,
    },
  }
}

test.describe('Preview --no-check-alter toggle', () => {
  test('default false: preview should NOT contain --no-check-alter', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/executions/preview`, {
      data: buildPreviewPayload(false),
    })
    expect(res.status()).toBe(200)
    const body = await res.json()
    expect(body.code).toBe(200)
    expect(body.data.command).toContain('pt-online-schema-change')
    expect(body.data.command).not.toContain('--no-check-alter')
  })

  test('enabled true: preview should contain --no-check-alter', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/executions/preview`, {
      data: buildPreviewPayload(true),
    })
    expect(res.status()).toBe(200)
    const body = await res.json()
    expect(body.code).toBe(200)
    expect(body.data.command).toContain('pt-online-schema-change')
    expect(body.data.command).toContain('--no-check-alter')
  })
})


